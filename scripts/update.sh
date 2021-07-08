#!/bin/bash

docker pull l1ving/fs-over-http:latest

docker stop foh
docker rm foh

docker run --name foh \
  -e MAXBODYSIZE="1048576000" \
  -e ADDRESS="localhost:6010" \
  --mount type=bind,source=/home/liv/fs-over-http,target=/foh-files \
  --network host -d \
  l1ving/fs-over-http
