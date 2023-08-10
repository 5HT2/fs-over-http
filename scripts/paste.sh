#!/bin/bash

# Usage:
# paste
# paste < tmp.log
# cat tmp.log | paste

# shellcheck disable=SC1091
source "$HOME/.env"

EXT="txt"
URL="https://i.l1v.in"
CDN_URL="https://cdn.frogg.ie" # This is a reverse proxy to $URL/media/
APP_NAME="cdn.frogg.ie"

if [[ -n "$1" ]]; then
    EXT="$1"
fi

# Set default filename and path
filename="$(date +"paste-%s.$EXT")"
filepath="$HOME/.cache/$filename"

printf 'Type your paste and press \u001b[31mCtrl D\u001b[0m when finished\n'

cat - > "$filepath"
printf 'Uploading...\n'

# Upload the screenshot
curl -s -X POST -H "Auth: $FOH_SERVER_AUTH" "$URL/public/media/$filename" -F "file=@$filepath"

# Copy the screenshot URL to clipboard
printf '%s/%s' "$CDN_URL" "$filename" | xclip -sel clip
echo "Uploaded $CDN_URL/$filename"

# Send notification after copying to clipboard
notify-send "Uploaded paste" "$filename" --icon=clipboard --app-name="$APP_NAME"

# Remove temporary file
rm "$filepath"
