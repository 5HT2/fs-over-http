# [fs-over-http](https://i.l1v.in)
[![time tracker](https://wakatime.com/badge/github/5HT2/fs-over-http.svg)](https://wakatime.com/badge/github/5HT2/fs-over-http)
[![Docker Pulls](https://img.shields.io/docker/pulls/l1ving/fs-over-http?logo=docker&logoColor=white)](https://hub.docker.com/r/l1ving/fs-over-http)
[![Docker Build](https://img.shields.io/github/actions/workflow/status/5HT2/fs-over-http/docker-build.yml?logo=docker&logoColor=white&branch=master)](https://github.com/5HT2/fs-over-http/actions/workflows/docker-build.yml)
[![CodeFactor](https://img.shields.io/codefactor/grade/github/5HT2/fs-over-http?logo=codefactor&logoColor=white)](https://www.codefactor.io/repository/github/5HT2/fs-over-http)

A filesystem interface over http.

**NOTE:** I wrote this when I was still learning Go, and as such many improvements can be made. 
I have detailed what I would like to improve in the [TODO](#todo) section, with *Partial Content*,
better *error handling* and *response syntax* being the main focus.

## Contributing

Contributions to fix my code are welcome, as well as any improvements.

To build:
```bash
git clone git@github.com:5HT2/fs-over-http.git
cd fs-over-http
make
```

To run:
```bash
# I recommend using genpasswd https://gist.github.com/5HT2/30f98284e9f92e1b47b4df6e05a063fc
AUTH='some secure token'
echo "$AUTH" > token

# Change the port to whatever you'd like. 
# Change localhost to your public IP if you'd like.
./fs-over-http -addr=localhost:6060
```

## Usage

Please see [`USAGE.md`](https://github.com/5HT2/fs-over-http/blob/master/USAGE.md) for examples of interacting with
a fs-over-http server.

### IPv6

IPv6 is supported, do note that you need to format the `addr` flag differently.

```bash
# Example IPv4
./fs-over-http -addr "10.0.1.1:6060"
# Example IPv6
./fs-over-http -addr "[2fb1:e540:13a7:3fa1:37bc:80b4:0b96:dbb8]:6060"
```

#### Production

I recommend using Caddy for automatic renewal + as a reverse proxy.
```
# Caddyfile example
i.l1v.in {
  header Server Caddy "Nintendo Wii"
  reverse_proxy localhost:6060
}
```

There is also a docker image available with the following command, or checkout the 
[`update.sh`](https://github.com/5HT2/fs-over-http/blob/master/scripts/update.sh) script for automatically
updating a live docker image.
```bash
docker pull l1ving/fs-over-http:latest
```

## TODO

- [x] Binary file support
- [x] Allow marking a folder as public
- [ ] Custom shell for interacting
- [ ] Partial Content support [(docs)](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/206)
- [x] Switch `X-File-Content` to using forms
  - [x] eg: `curl -X POST -H "Auth: $TOKEN" -d 'content=File content' localhost:6060/file.txt`
  - [x] Switch folder creation to same syntax with empty `content`
  - [ ] Read 512 bytes at a time like [so](https://pkg.go.dev/github.com/valyala/fasthttp#RequestCtx.SetBodyStream).
- [x] Move error handling to ListenAndServe instead of individually sending the error
  - [x] Switch to using `X-Error-Message` instead of printing it out, add a newline end of normal responses
- [x] Refactor use of JoinStr to `fmt.Sprintf/Sprintln` and `+`
- [x] Set `ReadTimeout` and `WriteTimeout` to prevent abuse
- [x] Add Docker image
  - [x] Add CI service
- [X] Add Caddyfile example
  - [ ] Maybe with rate limit options and the such
  - [X] Refactor docs about TLS
- [x] Encoding of uploading text-based files (eg the ● character)
- [x] Cleanup README
- [x] Fix scripts to use new format
