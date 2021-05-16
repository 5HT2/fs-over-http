# fs-over-http [![CodeFactor](https://www.codefactor.io/repository/github/l1ving/fs-over-http/badge)](https://www.codefactor.io/repository/github/l1ving/fs-over-http) [![time tracker](https://wakatime.com/badge/github/l1ving/fs-over-http.svg)](https://wakatime.com/badge/github/l1ving/fs-over-http)

A filesystem interface over http.

**NOTE:** I wrote this when I was still learning Go, and as such many improvements can be made. 
I have detailed what I would like to improve in the [TODO](#todo) section, with *Partial Content*,
better *error handling* and *response syntax* being the main focus.

## Contributing

Contributions to fix my code are welcome, as well as any improvements.

To build:
```bash
git clone git@github.com:l1ving/fs-over-http.git
cd fs-over-http
make
```

To run:
```bash
# I recommend using genpasswd https://gist.github.com/l1ving/30f98284e9f92e1b47b4df6e05a063fc
AUTH='some secure token'
echo "$AUTH" > token

# Change the port to whatever you'd like. 
# Change localhost to your public IP if you'd like.
# Compression is optional, but enabled if not explicitly set.
./fs-over-http -addr=localhost:6060 -compress=true
```

## Usage

All files are stored inside the `filesystem` folder, which is automatically created inside the working directory of fs-over-http.

### Response examples

The printed response will always either be
- The file path
- The file contents, or a tree output
- The error message thrown on a non-200 response

The only time you will get a file contents, or a tree output, as a response is on a GET request.

The only time you will get an error message is on a non-200 response.

#### Read a file or directory

```bash
# Read the root directory
curl -X GET -H "Auth: $TOKEN" localhost:6060

# Example output:
# filesystem/
# ├── asd/
# ├── myfile.txt
# ├── openjdk.png
# └── uwu
#
# 1 directory, 3 files

# Read a file
curl -X GET -H "Auth: $TOKEN" localhost:6060/myfile.txt

# Example output:
# I created this file with http!
```

#### Upload a file

```bash
curl -X POST -H "Auth: $TOKEN" localhost:6060/someimage.png -F "file=@$HOME/Downloads/myimage.jpg"
```

#### Create a folder

```bash
curl -X POST -H "Auth: $TOKEN" localhost:6060/my_folder -H "X-Create-Folder: true"
```

#### Write to a file

```bash
# Note that this will overwrite an existing file
curl -X POST -H "Auth: $TOKEN" localhost:6060/myfile.txt -H "X-File-Content: I created this file with http!"
```

#### Append to a file

```bash
# Note that this append to an existing file, and create a new file if one does not exist
curl -X PUT -H "Auth: $TOKEN" localhost:6060/myfile.txt -H "X-File-Content: I appended content to this file with http!"
```

#### Delete a file

```bash
curl -X DELETE -H "Auth: $TOKEN" localhost:6060/myfile.txt
```

#### Quick aliases

Alternatively, if you'd like, here's a bunch of bash aliases you can use with examples

```bash
# get owo.txt
get() { curl -X GET -H "Auth: $TOKEN" "localhost:6060/$1"; }

# upload someimage.png ~/Pictures/someimage.png
upload() { curl -X POST -H "Auth: $TOKEN" "localhost:6060/$1" -F "file=@$(echo "$2" | sed "s/~/\$HOME/g")"; }

# mkdir my_folder
mkdir() { curl -X POST -H "Auth: $TOKEN" "localhost:6060/$1" -H "X-Create-Folder: true"; }

# mkfile myfile.txt "I created this file with http!"
mkfile() { curl -X POST -H "Auth: $TOKEN" "localhost:6060/$1" -H "X-File-Content: $2"; }

# appendfile myfile.txt "I appended content to this file with http!"
appendfile() { curl -X PUT -H "Auth: $TOKEN" "localhost:6060/$1" -H "X-File-Content: $2"; }

# rm myfile.txt
rm() { curl -X DELETE -H "Auth: $TOKEN" "localhost:6060/$1"; }
```

#### Screenshot uploader

There is a screenshot uploader example in the `scripts` folder.

You will have to add the token in your `~/.profile` and edit the arguments that you want.

I have the keybinds assigned in my KDE custom commands, it allows you to run anything you want with a keyboard shortcut. For non-KDE you'll have to find your own way.

## TODO

- [x] Binary file support
- [x] Allow marking a folder as public
- [ ] Custom shell for interacting
- [ ] Ratelimit for public folders
  - [ ] Should this be done on the reverse proxy side instead? (eg: Caddy)
- [ ] Make `ls` sorting customizable
- [ ] Partial Content support [(docs)](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/206)
- [ ] Switch `X-File-Content` to using forms
  - [ ] eg: `curl -X POST -H "Auth: $TOKEN" -d 'content=File content' localhost:6060/file.txt`
  - [ ] Switch folder creation to same syntax with empty `content`
  - [ ] Read 512 bytes at a time like [so](https://pkg.go.dev/github.com/valyala/fasthttp#RequestCtx.SetBodyStream).
- [ ] Move error handling to ListenAndServe instead of individually sending the error
  - [ ] Switch to using `X-Error-Message` instead of printing it out, add a newline end of normal responses
- [ ] Refactor use of JoinStr to `fmt.Sprintf/Sprintln` and `+`
- [ ] Set `ReadTimeout` and `WriteTimeout` to prevent abuse
- [ ] JSON based API readouts
- [ ] Add Docker image
  - [ ] Add CI service
- [ ] Add Caddyfile example
  - [ ] Maybe with rate limit options and the such
  - [ ] Refactor docs about TLS
