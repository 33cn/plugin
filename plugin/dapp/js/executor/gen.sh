#!/bin/sh

{
    printf 'package executor\n\nvar callcode = `\n'
    cat "runtime.js"
    printf '`\n'
    printf 'var jscode = `\n'
    cat "test.js"
    printf '`\n'
    printf 'var _ = jscode\n'
    printf 'var gamecode = `\n'
    cat "game.js"
    printf '`\n'
    printf 'var _ = gamecode\n'
} >const.go

{
    printf 'package executor_test\n\nvar jscode = `\n'
    cat "test.js"
    printf '`\n'
    printf 'var _ = jscode\n'
    printf 'var gamecode = `\n'
    cat "game.js"
    printf '`\n'
    printf 'var _ = gamecode\n'
} >const_test.go
