## Usage

All files are stored inside the `filesystem` folder, which is automatically created inside the working directory of fs-over-http.

### Query Args

Supported query args and their values:

- [ ] `sort`
    - [ ] `name` (supported & default)
    - [x] `date` (supported)
    - [ ] `reverse` (planned)

- [ ] `format`
    - [x] `plain` (supported & default)
    - [ ] `json` (planned)
    - [ ] `visual` (planned)

### Response examples

Read the `X-Server-Message` header to get the response message.

Read the `X-Modified-Path` header to get the successfully modified path (eg POST / PUT / DELETE).
If the path is a folder it will end with `/`, otherwise it will not end in a `/`.

The only time you will get an error message as output instead of GET contents is on a non-200 response.

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
curl -X POST -H "Auth: $TOKEN" localhost:6060 -F "dir=my_folder"
```

#### Write to a file

```bash
# Note that this will overwrite an existing file
curl -X POST -H "Auth: $TOKEN" localhost:6060/myfile.txt -F "content=I created this file with http!"
```

#### Append to a file

```bash
# Note that this append to an existing file, and create a new file if one does not exist
curl -X PUT -H "Auth: $TOKEN" localhost:6060/myfile.txt -F "content=I appended content to this file with http!"
```

#### Delete a file or folder

```bash
curl -X DELETE -H "Auth: $TOKEN" localhost:6060/myfile.txt
```

#### Screenshot uploader

There is a screenshot uploader example in the `scripts` folder.

You will have to add the token in your `~/.env` and edit the arguments that you want.

```bash
# .env
FOH_SERVER_AUTH="secure token"
```

I have the keybinds assigned in my KDE custom commands, it allows you to run anything you want with a keyboard shortcut. For non-KDE you'll have to find your own way.

#### Quick aliases

Alternatively, if you'd like, here's a bunch of bash aliases you can use with examples

```bash
# get owo.txt
get() { curl -X GET -H "Auth: $TOKEN" "localhost:6060/$1"; }

# upload someimage.png ~/Pictures/someimage.png
upload() { curl -X POST -H "Auth: $TOKEN" "localhost:6060/$1" -F "file=@$(echo "$2" | sed "s/~/\$HOME/g")"; }

# mkdir my_folder
mkdir() { curl -X POST -H "Auth: $TOKEN" "localhost:6060" -F "dir=$1"; }

# mkfile myfile.txt "I created this file with http!"
mkfile() { curl -X POST -H "Auth: $TOKEN" "localhost:6060/$1" -F "content=$2"; }

# appendfile myfile.txt "I appended content to this file with http!"
appendfile() { curl -X PUT -H "Auth: $TOKEN" "localhost:6060/$1" -F "content=$2"; }

# rm myfile.txt
rm() { curl -X DELETE -H "Auth: $TOKEN" "localhost:6060/$1"; }
```
