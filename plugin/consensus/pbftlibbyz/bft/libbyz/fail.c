#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include "fail.h"

void fail(char const* format, ...) {
    va_list ap;
    va_start(ap, format);
    vfprintf(stderr, format, ap);
    va_end(ap);
    putc('\n', stderr);

    exit(1);
}

void sysfail(char const* msg) {
    perror(msg);
    exit(1);
}

void syswarn(char const* msg) {
    perror(msg);
}
