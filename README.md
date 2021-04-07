# fs-over-http [![CodeFactor](https://www.codefactor.io/repository/github/l1ving/fs-over-http/badge)](https://www.codefactor.io/repository/github/l1ving/fs-over-http) [![time tracker](https://wakatime.com/badge/github/l1ving/fs-over-http.svg)](https://wakatime.com/badge/github/l1ving/fs-over-http)

A filesystem interface over http.

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
curl -X POST -H "Auth: $TOKEN" localhost:6060/myfolder -H "X-Create-Folder: true"
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

## TODO:

- [x] Binary file support
- [ ] Allow marking a folder as public
- [ ] Custom shell for interacting
