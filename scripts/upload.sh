#!/bin/bash

# Upload a file via 5HT2/fs-over-http
upload() {
    if [[ -z "$1" ]] || [[ -z "$2" ]]; then
        echo "Missing args 1 or 2!"
        return
    fi

    # shellcheck disable=SC1091
    source "$HOME/.env"
    curl -X POST \
        -H "Auth: $FOH_SERVER_AUTH" \
        -F "file=@${2//\~/\$HOME}" \
        "https://i.l1v.in/public/$1"

    BASE="https://i.l1v.in"
    FILE_NAME="$1"

    if [[ "$1" == media/* ]]; then
        FILE_NAME="$(echo "$1" | sed -E "s/[A-z0-9]+\///g")"
        BASE="https://cdn.l1v.in"
    fi

    printf '%s/%s' "$BASE" "$FILE_NAME" | xclip -sel clip
    notify-send "Uploaded file" "$(basename "$2")" --icon=clipboard --app-name="cdn.l1v.in"
}
