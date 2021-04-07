package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"log"
	"os"
	"unicode/utf8"
)

var (
	addr         = flag.String("addr", ":6060", "TCP address to listen to")
	compress     = flag.Bool("compress", true, "Whether to enable transparent response compression")
	useTls       = flag.Bool("tls", false, "Whether to enable TLS")
	tlsCert      = flag.String("cert", "", "Full certificate file path")
	tlsKey       = flag.String("key", "", "Full key file path")
	authToken    = []byte(ReadFileUnsafe("token", true))
	fsFolder     = "filesystem/"
	publicFolder = "filesystem/public/"
	ownerPerm    = os.FileMode(0700)
)

func main() {
	flag.Parse()

	// If fsFolder does not exist
	if _, err := os.Stat(fsFolder); os.IsNotExist(err) {
		err := os.Mkdir(fsFolder, ownerPerm)

		if err != nil {
			log.Fatalf("- Error making fsFolder - %v", err)
		}
	}

	// If publicFolder does not exist
	if _, err := os.Stat(publicFolder); os.IsNotExist(err) {
		err := os.Mkdir(publicFolder, ownerPerm)

		if err != nil {
			log.Fatalf("- Error making publicFolder - %v", err)
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
	auth := ctx.Request.Header.Peek("Auth")

	// requestPath is prefixed with a /
	path := TrimFirstRune(string(ctx.Path()))
	filePath := JoinStr(fsFolder, path)

	if len(auth) == 0 {
		filePath = JoinStr(publicFolder, path)
		HandleServeFile(ctx, filePath)
		return
	}

	// Make sure Auth key is correct
	if !bytes.Equal(auth, authToken) {
		HandleForbidden(ctx)
		return
	}

	switch string(ctx.Request.Header.Method()) {
	case fasthttp.MethodPost:
		HandlePostRequest(ctx, filePath)
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

func HandlePostRequest(ctx *fasthttp.RequestCtx, file string) {
	fh, err := ctx.FormFile("file")

	if err != nil {
		HandleWriteFile(ctx, file)
		return
	}

	err = fasthttp.SaveMultipartFile(fh, file)

	if err != nil {
		HandleInternalServerError(ctx, err)
		return
	}

	fmt.Fprint(ctx, "200 Success\n")
}

func HandleWriteFile(ctx *fasthttp.RequestCtx, file string) {
	content := ctx.Request.Header.Peek("X-File-Content")
	cf := ctx.Request.Header.Peek("X-Create-Folder")

	if len(cf) != 0 {
		HandleCreateFolder(ctx, file, cf)
		return
	}

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
			fmt.Fprintf(ctx, "%s\n\n", file)
			fmt.Fprintf(ctx, "0 directories, 0 files\n")
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

	if err != nil {
		HandleInternalServerError(ctx, err)
		return
	}

	// Open the file and handle errors
	f, err := os.Open(file)
	if err != nil {
		HandleInternalServerError(ctx, err)
		return
	}
	defer f.Close()

	// Get the contentType
	contentType, err := GetFileContentTypeExt(f, file)
	if err != nil {
		HandleInternalServerError(ctx, err)
		return
	}

	ctx.Response.Header.Set(fasthttp.HeaderContentType, contentType)
	fmt.Fprint(ctx, content)
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

func HandleCreateFolder(ctx *fasthttp.RequestCtx, file string, cf []byte) {
	if string(cf) != "true" {
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		fmt.Fprint(ctx, "400 Invalid X-Create-Folder\n")
		return
	}

	err := os.Mkdir(file, ownerPerm)

	if err != nil {
		HandleInternalServerError(ctx, err)
		return
	}

	// TODO: Print folder path
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
