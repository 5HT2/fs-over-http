#!/bin/bash

# shellcheck disable=SC1091
source "$HOME/.env"
if [[ -z "$FOH_PATH" ]]; then
  echo "FOH_PATH not set!"
  exit 1
fi

docker pull l1ving/fs-over-http:latest

docker stop foh || echo "Could not stop missing container foh"
docker rm foh || echo "Could not remove missing container foh"

docker run --name foh \
  -e MAXBODYSIZE="1048576000" \
  -e ADDRESS="localhost:6010" \
  --mount type=bind,source="$FOH_PATH",target=/foh-files \
  --network host -d \
  l1ving/fs-over-http
