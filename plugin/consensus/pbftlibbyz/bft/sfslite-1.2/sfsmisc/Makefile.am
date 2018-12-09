## Process this file with automake to produce Makefile.in
# $Id: Makefile.am 2345 2006-11-30 23:49:01Z max $

sfslib_LTLIBRARIES = libsfsmisc.la
LIBSFSMISC = libsfsmisc.la

if USE_SFSMISC
libsfsmisc_la_SOURCES = \
 afsdir.C afsnode.C agentconn.C agentmisc.C attrcache.C closesim.C	\
 findfs.C getfh3.C nfs3_err.C nfsserv.C nfstrans.C nfs3attr.C		\
 nfsxattr.C pathexpand.C sfs_err.C sfsaid.C sfsauthorizer.C sfsclient.C	\
 sfsclientauth.C sfsconnect.C sfsconst.C sfskeyfetch.C sfskeymisc.C	\
 sfshostalias.C \
 sfspath.C sfsserv.C sfssesskey.C sfssrpconnect.C sfstty.C suidgetfd.C	\
 unixserv.C uvfstrans.C sfscrypt.C sfsschnorr.C validshell.C sfskeymgr.C \
 sfsgroupmgr.C 
else
libsfsmisc_la_SOURCES =	sfsconst_stub.C
endif

sfsinclude_HEADERS = \
afsnode.h agentconn.h agentmisc.h getfh3.h nfs3_nonnul.h	\
nfsserv.h nfstrans.h rex.h sfsclient.h sfsconnect.h sfskeymisc.h\
sfsmisc.h sfsserv.h sfstty.h uvfstrans.h sfscrypt.h sfsschnorr.h\
sfskeymgr.h sfsgroupmgr.h rexcommon.h sfssesscrypt.h

noinst_LIBRARIES = libmallock.a	# hack for automake to build mallock.o
libmallock_a_SOURCES = mallock.C

sfslib_DATA = mallock.o

DEPEND_ON_MAKEFILE = sfsconst.o sfsconst.lo
$(DEPEND_ON_MAKEFILE): Makefile

if REPO
sfsmisc_repo_OBJECTS = $(libsfsmisc_la_OBJECTS) \
	$(LIBSVC) $(LIBSFSCRYPT) $(LIBARPC) $(LIBASYNC)
stamp-sfsmisc-repo: $(sfsmisc_repo_OBJECTS)
	-$(CXXLINK) $(sfsmisc_repo_OBJECTS)
	@rm -f a.out
	touch $@
libsfsmisc_la_DEPENDENCIES = stamp-sfsmisc-repo
endif

EXTRA_DIST = .cvsignore
CLEANFILES = core *.core *~ *.rpo stamp-sfsmisc-repo
MAINTAINERCLEANFILES = Makefile.in
