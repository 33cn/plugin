## Process this file with automake to produce Makefile.in
# $Id: Makefile.am 3430 2008-06-17 18:13:38Z max $

# Note:  The files
#   internal.h pcre.h dftables.c maketables.c pcre.c study.c
# are from pcre-4.5, the perl compatible regular expression library.
# Pcre was written by Philip Hazel, and is distributed from
#   http://www.pcre.org/
#   ftp://ftp.csx.cam.ac.uk/pub/software/programming/pcre/

sfslib_LTLIBRARIES = libasync.la
LIBASYNC = libasync.la
noinst_PROGRAMS = dftables
sfsexec_PROGRAMS = aiod mmcd

libasync_la_SOURCES = \
arandom.c clock_gettime.c flock.c rwfd.c stktrace.c suidprotect.c \
maketables.c pcre.c study.c \
aerr.C aio.C aios.C arena.C armor.C bbuddy.C cbuf.C convertint.C	\
core.C daemonize.C dns.C dnsparse.C err.C fdwait.C ident.C ifchg.C	\
ihash.C itree.C lockfile.C malloc.C msb.C myaddrs.C myname.C		\
parseopt.C pipe2str.C refcnt.C rxx.C sigio.C socket.C spawn.C str.C	\
str2file.C straux.C suio++.C suio_vuprintf.C tcpconnect.C litetime.C \
select.C select_std.C select_epoll.C select_kqueue.C dynenum.C \
vec.C

aiod_SOURCES = aiod.C
aiod_LDADD = $(LIBASYNC) $(LIBPY) $(LDADD_THR)
aiod_DEPENDENCIES = $(LIBASYNC)

sfsinclude_HEADERS = \
aiod.h aiod_prot.h aios.h amisc.h arena.h array.h async.h backoff.h	\
bbuddy.h bitvec.h callback.h cbuf.h dns.h dnsimpl.h dnsparse.h err.h	\
fdlim.h ihash.h init.h itree.h keyfunc.h list.h msb.h opnew.h		\
parseopt.h pcre.h qhash.h refcnt.h rxx.h serial.h stllike.h str.h	\
suio++.h sysconf.h union.h vatmpl.h vec.h rwfd.h litetime.h       	\
corebench.h callback.h qtailq.h sfs_select.h rclist.h dynenum.h         \
rctailq.h rctree.h

#
# begin sfslite changes
#
mmcd_SOURCES = mmcd.c
mmcd_LDADD = 
mmcd_DEPENDENCIES = 


#
# end sfslite changes
#

noinst_HEADERS = internal.h env.mk

dftables_SOURCES = dftables.c
dftables_LDADD =

pcre.o pcre.lo: chartables.c
chartables.c: dftables
	./dftables $@

DEPEND_ON_MAKEFILE = daemonize.o daemonize.lo spawn.o spawn.lo
$(DEPEND_ON_MAKEFILE): Makefile

if REPO
async_repo_OBJECTS = $(libasync_la_OBJECTS)
stamp-async-repo: $(async_repo_OBJECTS)
	-$(CXXLINK) $(async_repo_OBJECTS)
	@rm -f a.out
	touch $@
libasync_la_DEPENDENCIES = stamp-async-repo
endif

stktrace.lo: $(srcdir)/stktrace.c $(srcdir)/sysconf.h
	$(LTCOMPILE) $(NW) -c $(srcdir)/stktrace.c

#
# include Makefile flags and options for pysfs1 
#
env.mk: Makefile
	@rm -f env.mk~
	echo "# Makefile fragment generated " `date` > env.mk~
	echo "CC = $(CC)" >> env.mk~
	echo "CXX = $(CXX)" >> env.mk~
	echo "CFLAGS = $(CFLAGS)" >> env.mk~
	echo "CXXFLAGS = $(CXXFLAGS)" >> env.mk~
	echo "LIBS = $(LIBS)" >> env.mk~
	echo "LIBGMP = $(LIBGMP)" >> env.mk~
	echo "LDFLAGS = $(LDFLAGS)" >> env.mk~
	echo "LDADD_THR = $(LDADD_THR)" >> env.mk~
	mv -f env.mk~ env.mk


EXTRA_DIST = .cvsignore env.mk
CLEANFILES = core *.core *~ *.rpo dftables chartables.c stamp-async-repo env.mk
MAINTAINERCLEANFILES = Makefile.in env.mk

.PHONY: cbh
cbh:
	cd $(srcdir) && sh ./callback.h > callback.h~ \
		     && mv -f callback.h~ callback.h 

.PHONY: vah
vah:
	cd $(srcdir) && sh ./vatmpl.h > vatmpl.h~ \
		&& mv -f vatmpl.h~ vatmpl.h
