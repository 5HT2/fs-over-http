FROM golang:1.16.5

RUN mkdir /fs-over-http
ADD . /fs-over-http
WORKDIR /fs-over-http

RUN go build -o foh-bin .

ENV MAXBODYSIZE 104857600
CMD /fs-over-http/foh-bin -maxbodysize $MAXBODYSIZE
