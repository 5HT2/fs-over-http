#!/bin/bash

docker pull l1ving/fs-over-http:latest
CONTAINER_ID="$(docker ps -f name=foh --format "{{.ID}}" | head -n 1)"

echo "Stopping container $CONTAINER_ID"
docker stop "$CONTAINER_ID"
docker rm "$CONTAINER_ID"

docker run --name foh \
  -e MAXBODYSIZE="1048576000" \
  -e ADDRESS="localhost:6010" \
  --mount type=bind,source=/home/liv/fs-over-http,target=/foh-files \
  --network host -d \
  l1ving/fs-over-http
