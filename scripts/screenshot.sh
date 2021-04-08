#!/bin/bash

# Usage:
# screenshot    ## take a selection screenshot
# screenshot -a ## take a screenshot of the active window
# screenshot -m ## take a screenshot of the active monitor

source ~/.profile

TOKEN="$FOH_SERVER_AUTH"
URL="https://i.l1v.in"

filename="$(date +"%Y-%m-%d-%T.png")"
filepath="$HOME/pictures/screenshots/$filename"

# Default argument is a selection screenshot
format="-region"

# Allow -a / -m / custom args
if [ ! -z "$1" ]; then
    format="$1"
fi

spectacle "$format" -p -b -n -o="$filepath" >/dev/null 2>&1 

# Wait for spectacle to finish saving the file
while [ ! -f "$filepath" ]; do
    sleep 0.1
done

# Upload the screenshot
RESPONSE=$(curl -s -X POST -H "Auth: $TOKEN" "$URL/public/i/$filename" -F "file=@$filepath")

# Copy the screenshot URL to clipboard
printf "$URL/$(echo "$RESPONSE" | sed "s/^filesystem\/public\///g")" | xclip -sel clip

notify-send "Saved screenshot" "$filename" --icon=spectacle --app-name="i.l1v.in"
