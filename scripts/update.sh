#!/bin/bash

docker pull l1ving/fs-over-http:latest
CONTAINER_ID="$(docker ps | grep foh | head -n 1 | cut -c -12)"

echo "Stopping container $CONTAINER_ID"
docker stop "$CONTAINER_ID"
docker rm "$CONTAINER_ID"

docker run --name foh \
  -e MAXBODYSIZE="1048576000" \
  -e ADDRESS="localhost:6010" \
  --mount type=bind,source=/home/liv/fs-over-http,target=/foh-files \
  --network host -d \
  l1ving/fs-over-http
