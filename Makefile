NAME   := l1ving/fs-over-http
TAG    := $(shell git log -1 --pretty=%h)
IMG    := ${NAME}:${TAG}
LATEST := ${NAME}:latest

fs-over-http: clean
	go get -u github.com/valyala/fasthttp
	go build

clean:
	rm -f fs-over-http

build:
	@docker build -t ${IMG} .
	@docker tag ${IMG} ${LATEST}

push:
	@docker push ${NAME}

login:
	@docker login -u ${DOCKER_USER} -p ${DOCKER_PASS}
