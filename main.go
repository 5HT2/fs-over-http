package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	addr      = flag.String("addr", ":6060", "TCP address to listen to")
	compress  = flag.Bool("compress", true, "Whether to enable transparent response compression")
	useTls    = flag.Bool("tls", false, "Whether to enable TLS")
	tlsCert   = flag.String("cert", "", "Full certificate file path")
	tlsKey    = flag.String("key", "", "Full key file path")
	authToken = []byte(ReadFileUnsafe("token", true))
	fsFolder  = "filesystem/"
)

func main() {
	flag.Parse()

	// If folder does not exist
	if _, err := os.Stat(fsFolder); os.IsNotExist(err) {
		err := os.Mkdir(fsFolder, 0700)

		if err != nil {
			log.Fatalf("- Error making fsFolder - %v", err)
		}
	}

	h := RequestHandler
	if *compress {
		h = fasthttp.CompressHandler(h)
	}

	if *useTls && len(*tlsCert) > 0 && len(*tlsKey) > 0 {
		if err := fasthttp.ListenAndServeTLS(*addr, *tlsCert, *tlsKey, h); err != nil {
			log.Fatalf("- Error in ListenAndServeTLS: %s", err)
		}
	} else {
		if err := fasthttp.ListenAndServe(*addr, h); err != nil {
			log.Fatalf("- Error in ListenAndServe: %s", err)
		}
	}
}

func RequestHandler(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set(fasthttp.HeaderServer, "fs-over-http")

	// The authentication key provided with said Auth header
	header := ctx.Request.Header.Peek("Auth")

	// Make sure Auth key is correct
	if !bytes.Equal(header, authToken) {
		HandleForbidden(ctx)
		return
	}

	// requestPath is prefixed with a /
	path := TrimFirstRune(string(ctx.Path()))
	filePath := JoinStr(fsFolder, path)

	switch string(ctx.Request.Header.Method()) {
	case fasthttp.MethodPost:
		HandleWriteFile(ctx, filePath)
	case fasthttp.MethodGet:
		HandleServeFile(ctx, filePath)
	case fasthttp.MethodPut:
		HandleAppendFile(ctx, filePath)
	case fasthttp.MethodDelete:
		HandleDeleteFile(ctx, filePath)
	default:
		HandleForbidden(ctx)
	}
}

func HandleServeFile(ctx *fasthttp.RequestCtx, file string) {
	isDir, err := IsDirectory(file)

	if err != nil {
		HandleInternalServerError(ctx, err)
		return
	}

	if isDir {
		// Append / as last rune, if last rune is not / and is dir
		r, _ := utf8.DecodeLastRuneInString(file)
		if r != '/' {
			file = JoinStr(file, "/")
		}

		files, err := ioutil.ReadDir(file)
		if err != nil {
			HandleInternalServerError(ctx, err)
			return
		}

		filesAmt := len(files)
		// No files in dir
		if filesAmt == 0 {
			ctx.Response.SetStatusCode(fasthttp.StatusNoContent)
			return
		}

		// Print path name
		fmt.Fprintf(ctx, "%s\n", file)

		tFiles, tFolders, cFile := 0, 0, 0
		lineRune := "├── "

		for _, f := range files {
			cFile++
			fn := f.Name()

			if f.IsDir() {
				tFolders++
				fn = JoinStr(fn, "/")
			} else {
				tFiles++
			}

			// Fix line rune for last file
			if cFile == filesAmt {
				lineRune = "└── "
			}

			fmt.Fprintf(ctx, "%s%s\n", lineRune, fn)
		}

		// Print total dirs and files
		fmt.Fprintf(
			ctx, "\n%s, %s\n",
			Grammar(tFolders, "directory", "directories"),
			Grammar(tFiles, "file", "files"))

		return
	}

	content, err := ReadFile(file)

	// File is empty
	if len(content) == 0 {
		ctx.Response.SetStatusCode(fasthttp.StatusNoContent)
		return
	}

	if err == nil {
		switch {
		case strings.HasSuffix(file, ".css"):
			ctx.Response.Header.Set(fasthttp.HeaderContentType, "text/css; charset=utf-8")
		case strings.HasSuffix(file, ".html"):
			ctx.Response.Header.Set(fasthttp.HeaderContentType, "text/html; charset=utf-8")
		}

		fmt.Fprint(ctx, content)
	} else {
		HandleForbidden(ctx)
	}
}

func HandleWriteFile(ctx *fasthttp.RequestCtx, file string) {
	content := ctx.Request.Header.Peek("X-File-Content")

	if len(content) == 0 {
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		fmt.Fprint(ctx, "400 Missing X-File-Content\n")
		return
	}

	err := WriteToFile(file, JoinStr(string(content), "\n"))

	if err != nil {
		HandleInternalServerError(ctx, err)
		return
	}

	fmt.Fprint(ctx, "200 Success\n")
}

func HandleAppendFile(ctx *fasthttp.RequestCtx, file string) {
	content := ctx.Request.Header.Peek("X-File-Content")

	if len(content) == 0 {
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		fmt.Fprint(ctx, "400 Missing X-File-Content\n")
		return
	}

	if file == fsFolder {
		HandleForbidden(ctx)
		return
	}

	contentStr := JoinStr(string(content), "\n")
	oldContent, err := ReadFile(file)

	if err == nil {
		contentStr = JoinStr(oldContent, contentStr)
	}

	err = WriteToFile(file, contentStr)

	if err != nil {
		HandleInternalServerError(ctx, err)
		return
	}

	fmt.Fprint(ctx, "200 Success\n")
}

func HandleDeleteFile(ctx *fasthttp.RequestCtx, file string) {
	if file == fsFolder {
		HandleForbidden(ctx)
		return
	}

	if _, err := os.Stat(file); err == nil {
		err = os.Remove(file)

		if err != nil {
			HandleInternalServerError(ctx, err)
		} else {
			fmt.Fprint(ctx, "200 Success\n")
		}
	} else {
		HandleInternalServerError(ctx, err)
	}
}

func HandleForbidden(ctx *fasthttp.RequestCtx) {
	ctx.Response.SetStatusCode(fasthttp.StatusForbidden)
	fmt.Fprint(ctx, "403 Forbidden\n")
	log.Printf(
		"- Returned 403 to %s - tried to connect with '%s' to '%s'",
		ctx.RemoteIP(), ctx.Request.Header.Peek("Auth"), ctx.Path())
}

func HandleInternalServerError(ctx *fasthttp.RequestCtx, err error) {
	ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
	fmt.Fprintf(ctx, "500 %v\n", err)
	log.Printf("- Returned 500 to %s with error %v", ctx.RemoteIP(), err)
}

func JoinStr(str string, suffix string) string {
	strArr := []string{str, suffix}
	return strings.Join(strArr, "")
}

func TrimFirstRune(s string) string {
	_, i := utf8.DecodeRuneInString(s)
	return s[i:]
}

func Grammar(amount int, singular string, multiple string) string {
	if amount != 1 {
		return strings.Join([]string{strconv.Itoa(amount), multiple}, " ")
	} else {
		return strings.Join([]string{strconv.Itoa(amount), singular}, " ")
	}
}
