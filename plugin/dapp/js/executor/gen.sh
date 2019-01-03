#!/bin/sh

echo "package executor\n\nvar callcode = \`" >const.go
cat runtime.js >>const.go
echo '`' >>const.go
echo 'var jscode = `' >>const.go
cat test.js >>const.go
echo '`' >>const.go
