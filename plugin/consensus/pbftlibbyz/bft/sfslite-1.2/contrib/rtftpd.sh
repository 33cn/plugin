#!/bin/sh
#
# $FreeBSD: ports/devel/rtftp/files/rtftpd.sh.in,v 1.1 2005/11/06 00:07:20 mnag Exp $
#

# PROVIDE: rtftpd
# REQUIRE: NETWORKING
# KEYWORD: FreeBSD shutdown

#
# Add the following lines to /etc/rc.conf to enable rtftpd:
#
# rtftpd_enable="YES"
#

rtftpd_enable=${rtftpd_enable-"NO"}
rtftpd_flags=${rtftpd_flags-"-rv -u rtftp -g okc -l local4.info /disk/rtftp/ccache"}

. /etc/rc.subr

name=rtftpd
rcvar=`set_rcvar`

command=/usr/local/lib/sfslite-1.2/shopt/${name}
pidfile=${pidfile:-/var/run/rtftpd.pid}

load_rc_config ${name}
run_rc_command "$1"
