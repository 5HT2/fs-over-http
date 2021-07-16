FROM golang:1.16.5

RUN mkdir /fs-over-http \
 && mkdir /foh-files
ADD . /fs-over-http
WORKDIR /fs-over-http

RUN go build -o foh-bin .

ENV ADDRESS "localhost:6060"
ENV MAXBODYSIZE "104857600"
CMD /fs-over-http/foh-bin -maxbodysize $MAXBODYSIZE -addr $ADDRESS

WORKDIR /foh-files
