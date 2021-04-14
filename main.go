package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/valyala/fasthttp"
)

var (
	addr         = flag.String("addr", ":6060", "TCP address to listen to")
	compress     = flag.Bool("compress", true, "Whether to enable transparent response compression")
	useTls       = flag.Bool("tls", false, "Whether to enable TLS")
	tlsCert      = flag.String("cert", "", "Full certificate file path")
	tlsKey       = flag.String("key", "", "Full key file path")
	maxBodySize  = flag.Int("maxbodysize", 100*1024*1024, "MaxRequestBodySize, defaults to 100MiB")
	authToken    = []byte(ReadFileUnsafe("token", true))
	fsFolder     = "filesystem/"
	publicFolder = "filesystem/public/"
	disallowed   = ReadNonEmptyLines("private_folders")
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

	// TODO: Switch over to using something similar to https://github.com/alessiosavi/GoDiffBinary/blob/7a8d35a20e38b14268b9840a4f9631f537a4dfea/api/api.go#L15
	// Instead of the manual RequestHandler that we have going

	// The gzipHandler will serve a compress request only if the client request it with headers (Content-Type: gzip, deflate)
	// Compress data before sending (if requested by the client)
	gzipHandler := fasthttp.CompressHandlerLevel(h, fasthttp.CompressBestCompression)

	s := &fasthttp.Server{
		Handler:            gzipHandler,
		Name:               "fs-over-http",
		MaxRequestBodySize: *maxBodySize,
	}

	if *useTls && len(*tlsCert) > 0 && len(*tlsKey) > 0 {
		if err := s.ListenAndServeTLS(*addr, *tlsCert, *tlsKey); err != nil {
			log.Fatalf("- Error in ListenAndServeTLS: %s", err)
		}
	} else {
		if err := s.ListenAndServe(*addr); err != nil {
			log.Fatalf("- Error in ListenAndServe: %s", err)
		}
	}
}

func RequestHandler(ctx *fasthttp.RequestCtx) {
	// The authentication key provided with said Auth header
	auth := ctx.Request.Header.Peek("Auth")

	// requestPath is prefixed with a /
	path := TrimFirstRune(string(ctx.Path()))
	filePath := JoinStr(fsFolder, path)

	if len(auth) == 0 && ctx.IsGet() {
		if Contains(disallowed, RemoveLastRune(path, '/')) {
			HandleForbidden(ctx)
			return
		}

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
		HandleError(ctx, err)
		return
	}

	fmt.Fprint(ctx, RemoveLastRune(file, '/'))
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
		HandleError(ctx, err)
		return
	}

	fmt.Fprint(ctx, RemoveLastRune(file, '/'))
}

func HandleServeFile(ctx *fasthttp.RequestCtx, file string) {
	isDir, err := IsDirectory(file)

	if err != nil {
		HandleError(ctx, err)
		return
	}

	if isDir {
		file = AddLastRune(file, '/')

		files, err := ioutil.ReadDir(file)
		if err != nil {
			HandleError(ctx, err)
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
				fn = AddLastRune(fn, '/')
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
		HandleError(ctx, err)
		return
	}

	// Open the file and handle errors
	f, err := os.Open(file)
	if err != nil {
		HandleError(ctx, err)
		return
	}
	defer f.Close()

	// Get the contentType
	contentType, err := GetFileContentTypeExt(f, file)
	if err != nil {
		HandleError(ctx, err)
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
		HandleError(ctx, err)
		return
	}

	fmt.Fprint(ctx, RemoveLastRune(file, '/'))
}

func HandleDeleteFile(ctx *fasthttp.RequestCtx, file string) {
	if file == fsFolder {
		HandleForbidden(ctx)
		return
	}

	if _, err := os.Stat(file); err == nil {
		err = os.Remove(file)

		if err != nil {
			HandleError(ctx, err)
		} else {
			fmt.Fprint(ctx, RemoveLastRune(file, '/'))
		}
	} else {
		HandleError(ctx, err)
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
		HandleError(ctx, err)
		return
	}

	fmt.Fprint(ctx, AddLastRune(file, '/'))
}

func HandleForbidden(ctx *fasthttp.RequestCtx) {
	ctx.Response.SetStatusCode(fasthttp.StatusForbidden)
	fmt.Fprint(ctx, "403 Forbidden\n")
	log.Printf(
		"- Returned 403 to %s - tried to connect with '%s' to '%s'",
		ctx.RemoteIP(), ctx.Request.Header.Peek("Auth"), ctx.Path())
}

func HandleError(ctx *fasthttp.RequestCtx, err error) {
	ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
	if strings.Contains(err.Error(), "no such file or directory") {
		ctx.Error("File not found", fasthttp.StatusNotFound)
		log.Printf("- Returned 404 to %s with error %v", ctx.RemoteIP(), err)
	} else {
		ctx.Error("Internal Server Error", fasthttp.StatusInternalServerError)
		log.Printf("- Returned 500 to %s with error %v", ctx.RemoteIP(), err)
	}
}
