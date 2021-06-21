package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/valyala/fasthttp"
	"io/fs"
	"log"
	"os"
	"sort"
	"strings"
)

var (
	addr         = flag.String("addr", "localhost:6060", "TCP address to listen to")
	compress     = flag.Bool("compress", true, "Whether to enable transparent response compression")
	useTls       = flag.Bool("tls", false, "Whether to enable TLS")
	tlsCert      = flag.String("cert", "", "Full certificate file path")
	tlsKey       = flag.String("key", "", "Full key file path")
	maxBodySize  = flag.Int("maxbodysize", 100*1024*1024, "MaxRequestBodySize, defaults to 100MiB")
	authToken    = []byte(ReadFileUnsafe("token", true))
	fsFolder     = "filesystem/"
	publicFolder = "filesystem/public/"
	privateDirs  = ReadNonEmptyLines("private_folders", publicFolder)
	ownerPerm    = os.FileMode(0700)
)

func main() {
	flag.Parse()

	protocol := "http"
	if *useTls {
		protocol += "s"
	}

	log.Printf("- Running fs-over-http on " + protocol + "://" + *addr)

	// If fsFolder or publicFolder don't exist
	SafeMkdir(fsFolder)
	SafeMkdir(publicFolder)

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
	filePath := fsFolder + path

	if len(auth) == 0 && ctx.IsGet() {
		filePath = publicFolder + path

		if Contains(privateDirs, RemoveLastRune(filePath, '/')) {
			HandleGeneric(ctx, fasthttp.StatusNotFound, "Not Found")
			return
		}

		HandleServeFile(ctx, filePath, true)
		return
	}

	// Make sure Auth key is correct
	if !bytes.Equal(auth, authToken) {
		HandleForbidden(ctx)
		return
	}

	switch string(ctx.Request.Header.Method()) {
	case fasthttp.MethodGet:
		HandleServeFile(ctx, filePath, false)
	case fasthttp.MethodPost:
		HandlePostRequest(ctx, filePath)
	case fasthttp.MethodPut:
		HandleAppendFile(ctx, filePath)
	case fasthttp.MethodDelete:
		HandleDeleteFile(ctx, filePath)
	default:
		HandleForbidden(ctx)
	}
}

func HandlePostRequest(ctx *fasthttp.RequestCtx, file string) {
	// If the dir key was provided, create that directory inside fsFolder
	dir := ctx.FormValue("dir")
	if len(dir) > 0 {
		dirStr := string(dir)
		// Remove all "/" and "." before the path
		for {
			dirStr = RemoveFirstRune(dirStr, '/')
			dirStr = RemoveFirstRune(dirStr, '.')
			if !strings.HasPrefix(dirStr, "/") && !strings.HasPrefix(dirStr, ".") {
				break
			}
		}

		// Add fsFolder as a prefix, we don't want the user to be able to make folders outside of it
		dirStr = fsFolder + dirStr
		err := os.MkdirAll(dirStr, ownerPerm)

		if err != nil {
			HandleInternalServerError(ctx, err)
			return
		}

		fmt.Fprint(ctx, AddLastRune(dirStr, '/'))
		return
	}

	// If not making a directory, don't allow writing directly to fsFolder
	if file == fsFolder {
		HandleNotAllowed(ctx, "Cannot POST on path \""+fsFolder+"\"")
		return
	}

	// If a file was provided, save it and return
	fh, err := ctx.FormFile("file")
	if err == nil {
		err = fasthttp.SaveMultipartFile(fh, file)

		if err != nil {
			HandleInternalServerError(ctx, err)
			return
		}

		fmt.Fprint(ctx, RemoveLastRune(file, '/'))
		return
	}

	// If the content key was provided, write to said file
	content := ctx.FormValue("content")
	if len(content) > 0 {
		err = WriteToFile(file, strings.ReplaceAll(string(content), "\\n", "\n"))

		if err != nil {
			HandleInternalServerError(ctx, err)
			return
		}

		fmt.Fprint(ctx, RemoveLastRune(file, '/'))
		return
	}

	// If none of the if statements passed, send a 400
	HandleGeneric(ctx, fasthttp.StatusBadRequest, "Missing 'file' or 'dir' or 'content' form")
}

func HandleServeFile(ctx *fasthttp.RequestCtx, file string, public bool) {
	isDir, err := IsDirectory(file)

	if err != nil {
		HandleInternalServerError(ctx, err)
		return
	}

	if isDir {
		file = AddLastRune(file, '/')

		files, err := fs.ReadDir(nil, file)

		// Don't list private folders
		if public {
			filter := func(s fs.DirEntry) bool { return !Contains(privateDirs, RemoveLastRune(file+s.Name(), '/')) }
			files = Filter(files, filter)
		}

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

		// TODO: Make file sorting customizable
		// Sort files by date
		sort.Slice(files, func(i, j int) bool {
			info1, err1 := files[i].Info()
			info2, err2 := files[j].Info()

			// TODO: better error handling (ie panic here)
			// If an error occurred while reading the file info
			if err1 != nil || err2 != nil {
				return true
			}

			return info1.ModTime().Before(info2.ModTime())
		})

		// Sort to put folders first
		sort.SliceStable(files, func(i, j int) bool { return files[i].IsDir() && !files[j].IsDir() })

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
	if file == fsFolder {
		HandleNotAllowed(ctx, "Cannot PUT on path \""+fsFolder+"\"")
		return
	}

	content := ctx.FormValue("content")
	// If the content key was not provided, return an error
	if len(content) == 0 {
		HandleGeneric(ctx, fasthttp.StatusBadRequest, "Missing 'content' form")
		return
	}

	contentStr := string(content) + "\n"
	oldContent, err := ReadFile(file)

	if err == nil {
		contentStr = oldContent + contentStr
	}

	err = WriteToFile(file, contentStr)

	if err != nil {
		HandleInternalServerError(ctx, err)
		return
	}

	fmt.Fprint(ctx, RemoveLastRune(file, '/'))
}

func HandleDeleteFile(ctx *fasthttp.RequestCtx, file string) {
	if file == fsFolder {
		HandleNotAllowed(ctx, "Cannot DELETE on path \""+fsFolder+"\"")
		return
	}

	if _, err := os.Stat(file); err == nil {
		err = os.Remove(file)

		if err != nil {
			HandleInternalServerError(ctx, err)
		} else {
			fmt.Fprint(ctx, RemoveLastRune(file, '/'))
		}
	} else {
		HandleInternalServerError(ctx, err)
	}
}
