# ff
This is file manager on terminal written in Go.

# Features
- Preview file/directory
- copy/paste file
- can edit file using `$EDITOR`

# Support OS
- Linux
- Mac

# Installtion
```sh
$ git clone https://github.com/skanehira/ff
$ cd ff
$ go install
```

# Keybinding
| panel | key   | operation                      |
|-------|-------|--------------------------------|
| path  | `tab` | focus to files                 |
| files | `tab` | focus to path                  |
| files | `h`   | cd to specified path           |
| files | `l`   | cd to parent path              |
| files | `y`   | copy selected file             |
| files | `p`   | paste copy file to current dir |
