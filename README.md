# ff
This is file manager on terminal written in Go.

# Features
- preview file/directory
- copy/paste file
- make a new file/directory
- rename a file/directory
- edit file with `$EDITOR`

# Support OS
- Linux
- Mac

# Installtion
```sh
$ git clone https://github.com/skanehira/ff
$ cd ff
$ go install
```

NOTE: Installation with `go get` is not recommended because libraries is not version locked.

# Usage
## Options
```sh
$ ff -h
Usage of ff:
  -log
        enable log
  -preview
        enable preview panel
```

If you use `-log` that will print log. If log file not exists then will be create in `$HONE/ff.log`.

`-preview` is enable preview panel that you can preview file or directories.

## Keybinding
| panel | key   | operation                      |
|-------|-------|--------------------------------|
| path  | `tab` | focus to files                 |
| files | `tab` | focus to path                  |
| files | `h`   | cd to specified path           |
| files | `l`   | cd to parent path              |
| files | `y`   | copy selected file             |
| files | `p`   | paste copy file to current dir |
| files | `d`   | delete selected file or dir    |
| files | `m`   | make a new dir                 |
| files | `n`   | make a new file                |
| files | `r`   | rename a dir or file           |
| files | `e`   | edit file with `$EDITOR`       |

# Author
skanehira
