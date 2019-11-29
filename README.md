# ff
`ff` is file manager written in Go.

![](https://github.com/skanehira/ff/blob/image/screenshots/ff-demo.gif?raw=true)

# Features
- preview file/directory
- copy/paste file
- make a new file/directory
- rename a file/directory
- edit file with `$EDITOR`
- open file/directory
- bookmark directory

# Go version
- 1.13~

# Support OS
- Linux/Unix
- Mac

# Installtion
```sh
$ git clone https://github.com/skanehira/ff
$ cd ff
$ go install
```

NOTE: Installation with `go get` is not recommended because libraries's version  is not locked.

# Usage
## Settings
If your terminal `LC_CTYPE` is not `en_US.UTF-8`, please set as following.

```sh
export LC_CTYPE=en_US.UTF-8
```

## Options
```sh
$ ff -h
Usage of ff:
  -ignorecase
        ignore case when searching
  -log
        enable log
  -preview
        enable preview panel
```

If you use `-log` that will print log.
If log file not exists, then will be create in `$XDG_CONFIG_HOME/ff/ff.log`.

`-preview` is enable preview panel that you can preview file or directories.

## Config
You can using `config.yaml` to config log, preview, etc...

```yaml
# print log to file
log:
  enable: true
  file: $XDG_CONFIG_HOME/ff/ff.log

# preview the contents of file or directory
preview:
  enable: true
  # preview colorscheme. you can use colorscheme following
  # https://xyproto.github.io/splash/docs/all.html
  colorscheme: monokai

# if ignore_case is true, ignore case when searching
ignore_case: true

# if enable is true, can use bookmark
bookmark:
  enable: true
  file: $XDG_CONFIG_HOME/ff/bookmark.db
```

The `config.yaml` should be placed in the following path.

|OS        |path                                              |
|----------|--------------------------------------------------|
|MacOS     |`$HOME/Library/Application Support/ff/config.yaml`|
|Linux/Unix|`$XDG_CONFIG_HOME/ff/config.yaml`                 |

## About bookmark
`ff` can use `b` to bookmark directory. bookmark will be stored sqlite3 database.
If you want enable bookmark, you have to specify database file.

Database file will auto create when `ff` starting. If `ff` can't create database file,
then will use inmemory mode.

The inmemory mode will save bookmark to memory, so if `ff` quit bookmarks will lost.

## About Edit file
If you runing `ff` in Vim's terminal and `$EDITOR` is `vim`,
`ff` will use running Vim to edit file.

## About open action
If you use `o` to open file or directory, ff will using `open` in MacOS, `xdg-open` in Linux.
You can set environment variable `FF_OPEN_CMD` to change open command.

## About executing command
`ff` can executing command, but it can't use stdin, stdout, stderr.
Example, if you run `vim` , `ff` will freeze.
So, you only can executing command that doesn't use stdin, stdout, stderr.

## Keybinding
### path
| key     | operation        |
|---------|------------------|
| `Enter` | change directory |

### files
| key      | operation                            |
|----------|--------------------------------------|
| `tab`    | focus to files                       |
| `tab`    | focus to path                        |
| `j`      | move to next                         |
| `k`      | move to previous                     |
| `g`      | move to top                          |
| `G`      | move to bottom                       |
| `ctrl-b` | move previous page                   |
| `ctrl-f` | move netxt page                      |
| `h`      | cd to specified path                 |
| `l`      | cd to parent path                    |
| `y`      | copy selected file                   |
| `p`      | paste copy file to current directory |
| `d`      | delete selected file or directory    |
| `m`      | make a new dir                       |
| `n`      | make a new file                      |
| `r`      | rename a dir or file                 |
| `e`      | edit file with `$EDITOR`             |
| `o`      | open file or directory               |
| `f`/`/`  | search files or directories          |
| `ctrl-j` | scroll preview panel down            |
| `ctrl-k` | scroll preview panel up              |
| `c`/`:`  | focus cmdline panel                  |
| `.`      | edit config.yaml                     |
| `b`      | bookmark dirctory                    |
| `B`      | open bookmarks panel                 |

### bookmark
| key      | operation             |
|----------|-----------------------|
| `a`      | add bookmark          |
| `d`      | delete bookmark       |
| `q`      | close bookmarks panel |
| `ctrl-g` | go to bookmark        |
| `f`/`/`  | search bookmarks      |

### cmd
| key         | operation         |
|-------------|-------------------|
| `enter`     | executing command |
| `tab`/`esc` | focus files       |

# Author
skanehira
