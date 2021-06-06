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
    echo "Uploading..."

    # We want newlines to work here
    # shellcheck disable=SC2059
    printf "$lines" > "$filepath"

    # Upload the screenshot
    RESPONSE=$(curl -s -X POST -H "Auth: $TOKEN" "$URL/public/media/$filename" -F "file=@$filepath")

    # Copy the screenshot URL to clipboard
    printf '%s/%s' \
        "$CDN_URL" \
        "$(echo "$RESPONSE" | sed "s/^filesystem\/public\/media\///g")" \
        | xclip -sel clip

    notify-send "Uploaded paste" "$filename" --icon=clipboard --app-name="$APP_NAME"
    rm "$filepath"
    exit
}
while read -r line
do
    ((count++))
    lines+="$line\n"
done < "${1:-/dev/stdin}"
