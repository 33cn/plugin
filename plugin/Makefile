all:
	go get -v -u gitlab.33.cn/chain33/chain33
	go build -i -o tool github.com/33cn/chain33/cmd/tools
	./tool import --path "." --packname "github.com/33cn/plugin" --conf "" --out "plugin.toml"
