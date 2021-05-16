#!/bin/bash

# Usage:
# screenshot                ## take a selection screenshot
# screenshot -a             ## take a screenshot of the active window
# screenshot -m             ## take a screenshot of the active monitor
# screenshot "" --fancyurl  ## An optional second arg to use a shorter fancy URL type

# shellcheck disable=SC1091
source "$HOME/.profile"

TOKEN="$FOH_SERVER_AUTH"
URL="https://i.l1v.in"
PIC_URL="https://p.l1v.in" # This is a reverse proxy to $URL/i/
APP_NAME="i.l1v.in"

# Set default filename and path
filename="$(date +"%Y-%m-%d-%T.png")"
filename_date="$filename"
filepath="$HOME/pictures/screenshots/$filename"

# Set the fancy url filename and filepath
if [ "$2" == "--fancyurl" ]; then
    array=("." "¨" "·" "˙" "•" "‥" "…" "∴" "∵" "∶" "∷" "⋮" "⋯" "⋰" "⋱" "⋅" "⋆" "∘")
    size=${#array[@]}
    filename=""

    # Generate 7 chars, 612,220,032 possible combinations
    for _ in {0..6}; do
        index=$((RANDOM % size))
        filename+="${array[$index]}"
    done

    filename+=".png"
fi

# Default argument is a selection screenshot
format="-region"

# Allow -a / -m / custom args
if [ -n "$1" ]; then
    format="$1"
fi

spectacle "$format" -p -b -n -o="$filepath" >/dev/null 2>&1 

# Wait for spectacle to finish saving the file
while [ ! -f "$filepath" ]; do
    sleep 0.2
done

# Upload the screenshot
RESPONSE=$(curl -s -X POST -H "Auth: $TOKEN" "$URL/public/i/$filename" -F "file=@$filepath")

# Copy the screenshot URL to clipboard
printf '%s/%s' \
    "$PIC_URL" \
    "$(echo "$RESPONSE" | sed "s/^filesystem\/public\/i\///g")" \
    | xclip -sel clip

notify-send "Saved screenshot" "$filename_date" --icon=spectacle --app-name="$APP_NAME"
