#!/bin/bash

# Usage:
# paste
# paste < tmp.log
# cat tmp.log | paste

# shellcheck disable=SC1091
source "$HOME/.profile"

TOKEN="$FOH_SERVER_AUTH"
URL="https://i.l1v.in"
CDN_URL="https://cdn.l1v.in" # This is a reverse proxy to $URL/media/
APP_NAME="cdn.l1v.in"

# Set default filename
filename="$(date +"paste-%s.txt")"

printf 'Type your paste and press \u001b[31mCtrl D\u001b[0m when finished\n'

content="$(cat -)"
printf 'Uploading...\n'

# Upload the paste
RESPONSE=$(curl -s -X POST -H "Auth: $TOKEN" "$URL/public/media/$filename" -F "content=$content")
FULL_URL="$CDN_URL/$(echo "$RESPONSE" | sed "s/^filesystem\/public\/media\///g")"

# Copy the paste URL to clipboard
printf '%s' "$FULL_URL" | xclip -sel clip
echo "Uploaded $FULL_URL"

# Send notification after copying to clipboard
notify-send "Uploaded paste" "$filename" --icon=clipboard --app-name="$APP_NAME"
