# WorkHelper-WinGui

使用([walk](https://github.com/lxn/walk))实现的工作助手windows图形化版本,方便使用.

## Build

```shell
set GOARCH=386
go build -ldflags "-s -w -X 'main.VERSION=0.1.0' -X 'main.NAME=WorkHelper' -H=windowsgui" -o ../bin/WorkHelper_win_386.exe -v
# set GOARCH=amd64
# go build -ldflags "-s -w -X 'main.VERSION=0.1.0' -X 'main.NAME=WorkHelper' -H=windowsgui" -o ../bin/WorkHelper_win_amd64.exe -v
```
