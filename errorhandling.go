package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/ferluci/fast-realip"
	"github.com/valyala/fasthttp"
)

func HandleModifyFsFolder(ctx *fasthttp.RequestCtx) {
	HandleGeneric(ctx, fasthttp.StatusMethodNotAllowed,
		"Cannot "+string(ctx.Request.Header.Method())+" on path \""+fsFolder+"\"")
}

func HandleGeneric(ctx *fasthttp.RequestCtx, status int, message string) {
	ctx.Response.SetStatusCode(status)
	ctx.Response.Header.Set("X-Server-Message", message)
	fmt.Fprintf(ctx, "%v %s\n", status, message)
}

func HandleForbidden(ctx *fasthttp.RequestCtx) {
	ctx.Response.SetStatusCode(fasthttp.StatusForbidden)
	ctx.Response.Header.Set("X-Server-Message", "403 Forbidden")
	fmt.Fprint(ctx, "403 Forbidden\n")
	log.Printf("- Returned 403 to %s - tried to connect with '%s' to '%s'",
		realip.FromRequest(ctx), ctx.Request.Header.Peek("Auth"), ctx.Path())
}

func HandleInternalServerError(ctx *fasthttp.RequestCtx, err error) {
	if strings.HasSuffix(err.Error(), "no such file or directory") {
		HandleGeneric(ctx, fasthttp.StatusNotFound, "Not Found")
		log.Printf("- Returned 404 to %s with error %v", realip.FromRequest(ctx), err)
		return
	}

	ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
	ctx.Response.Header.Set("X-Server-Message", "500 "+err.Error())
	fmt.Fprintf(ctx, "500 %v\n", err)
	log.Printf("- Returned 500 to %s with error %v", realip.FromRequest(ctx), err)
}
