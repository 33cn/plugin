all:
	go get -v -u gitlab.33.cn/chain33/chain33
	go build -i -o tool gitlab.33.cn/chain33/chain33/cmd/tools
	./tool import --path "." --packname "gitlab.33.cn/chain33/plugin" --conf "" --out "plugin.toml"
