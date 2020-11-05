go env -w CGO_ENABLED=0
go build -o chain33.exe
go build -o chain33-cli.exe github.com/33cn/plugin/cli
