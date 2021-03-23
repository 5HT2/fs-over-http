fs-over-http: clean
	go get -u github.com/valyala/fasthttp
	go build

clean:
	rm -f fs-over-http
