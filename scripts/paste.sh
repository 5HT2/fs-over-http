#!/bin/bash

# Usage:
# paste

# shellcheck disable=SC1091
source "$HOME/.profile"

TOKEN="$FOH_SERVER_AUTH"
URL="https://i.l1v.in"
CDN_URL="https://cdn.l1v.in" # This is a reverse proxy to $URL/media/
APP_NAME="cdn.l1v.in"

# Set default filename and path

filename="$(date +"paste-%s.txt")"
filepath="$HOME/.cache/$filename"

printf 'Type your paste and press \u001b[31mCtrl C\u001b[0m when finished\n'

# Loop until keyboard interrupt
trap printout SIGINT
printout() {
    printf '\nUploading...\n'

    # We want newlines to work here
    # shellcheck disable=SC2059
    printf "$lines" > "$filepath"

    # Upload the screenshot
    RESPONSE=$(curl -s -X POST -H "Auth: $TOKEN" "$URL/public/media/$filename" -F "file=@$filepath")
    FULL_URL="$CDN_URL/$(echo "$RESPONSE" | sed "s/^filesystem\/public\/media\///g")"

    # Copy the screenshot URL to clipboard
    printf '%s' "$FULL_URL" | xclip -sel clip
    echo "Uploaded $FULL_URL"

    # Send notification after copying to clipboard
    notify-send "Uploaded paste" "$filename" --icon=clipboard --app-name="$APP_NAME"
    exit
}
while read -r line
do
    ((count++))
    lines+="$line\n"
done < "${1:-/dev/stdin}"
