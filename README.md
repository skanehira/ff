# ff
This is file manager on terminal written in Go.


# Installtion
```sh
$ git clone https://github.com/skanehira/filemanager
$ cd filemanager
$ GOMODULE=on go install
```

# Keybinding
| panel      | key    | operation                  |
|------------|--------|----------------------------|
| path input | tab    | change to file list panel  |
| file list  | tab    | change to path input panel |
| file list  | h      | move to specified path     |
| file list  | l      | move to parent path        |
| file list  | ctrl+h | move to previous path      |
| file list  | ctrl+l | move to next path          |
