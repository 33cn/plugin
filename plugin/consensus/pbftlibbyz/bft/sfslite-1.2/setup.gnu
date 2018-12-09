#!/bin/sh

M4=gm4
$M4 --version < /dev/null 2>&1 | grep GNU >/dev/null 2>&1 || M4=gnum4
$M4 --version < /dev/null 2>&1 | grep GNU >/dev/null 2>&1 || M4=m4
$M4 --version < /dev/null 2>&1 | grep GNU >/dev/null 2>&1 \
    || (echo Cannot locate GNU m4 >&2; exit 1)

find . -name Makefile.am.m4 | while read file
do
    if test -f $file; then
	out=`echo $file | sed -e 's/\.m4$//'`
	echo "+ $M4 $file > $out"
	rm -f $out~
	$M4 $file > $out~
	mv -f $out~ $out
    fi
done

autoreconf $*
