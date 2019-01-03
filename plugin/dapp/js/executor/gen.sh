#!/bin/sh

printf "package executor\n\nvar callcode = \`\n" >const.go
cat "runtime.js" >>const.go
printf '`\n' >>const.go
printf 'var jscode = `\n' >>const.go
cat "test.js" >>const.go
printf '`\n' >>const.go
printf "var _ = jscode\n" >>const.go
