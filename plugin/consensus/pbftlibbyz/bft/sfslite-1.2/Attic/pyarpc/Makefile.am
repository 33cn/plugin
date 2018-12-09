## Process this file with automake to produce Makefile.in
# $Id: Makefile.am 863 2005-05-16 05:40:02Z max $

sfslib_LTLIBRARIES = libpyarpc.la

libpyarpc_la_SOURCES = py_rpctypes.C py_gen.C py_util.C py_debug.C

sfsinclude_HEADERS = py_rpctypes.h py_gen.h py_util.h  py_debug.h

if REPO
pyarpc_repo_OBJECTS = $(libpyarpc_la_OBJECTS) $(LIBASYNC)
stamp-pyarpc-repo: $(pyarpc_repo_OBJECTS)
	-$(CXXLINK) $(pyarpc_repo_OBJECTS)
	@rm -f a.out
	touch $@
libpyarpc_la_DEPENDENCIES = stamp-pyarpc-repo
endif

EXTRA_DIST = .cvsignore
CLEANFILES = core *.core *~ *.rpo stamp-pyarpc-repo
MAINTAINERCLEANFILES = Makefile.in
