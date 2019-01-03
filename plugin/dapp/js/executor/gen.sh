#!/bin/sh

{
    printf 'package executor\n\nvar callcode = `\n' 
    cat "runtime.js"
    printf '`\n'
    printf 'var jscode = `\n'
    cat "test.js"
    printf '`\n'
    printf 'var _ = jscode\n'
} > const.go
