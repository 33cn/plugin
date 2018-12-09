## Process this file with automake to produce Makefile.in
# $Id: Makefile.am 3713 2008-10-10 06:31:41Z max $

sfslib_LTLIBRARIES = libarpc.la

libarpc_la_SOURCES = \
authunixint.c pmap_prot.C \
acallrpc.C aclnt.C asrv.C authopaque.C authuint.C axprt_dgram.C axprt_pipe.C axprt_stream.C axprt_unix.C clone.C xdr_suio.C xdrmisc.C xhinfo.C \
rpc_stats.C rpc_lookup.C

sfsinclude_HEADERS = pmap_prot.x \
aclnt.h arpc.h asrv.h axprt.h pmap_prot.h rpctypes.h xdr_suio.h xdrmisc.h \
xhinfo.h rpc_stats.h

pmap_prot.h: $(srcdir)/pmap_prot.x
	@rm -f $@
	-$(RPCC) -h $(srcdir)/pmap_prot.x || rm -f $@

pmap_prot.C: $(srcdir)/pmap_prot.x
	@rm -f $@ $@~
	(echo '#define proc XXX_egcs_bug_proc'; \
		echo '#include "sysconf.h"'; \
		echo '#undef proc') > $@~
	-$(RPCC) -c $(srcdir)/pmap_prot.x -o- >> $@~ \
		&& mv -f $@~ $@

dist-hook:
	cd $(distdir) && rm -f pmap_prot.h pmap_prot.C

acallrpc.o pmap_prot.o: pmap_prot.h
acallrpc.lo pmap_prot.lo: pmap_prot.h

if REPO
arpc_repo_OBJECTS = $(libarpc_la_OBJECTS) $(LIBASYNC)
stamp-arpc-repo: $(arpc_repo_OBJECTS)
	-$(CXXLINK) $(arpc_repo_OBJECTS)
	@rm -f a.out
	touch $@
libarpc_la_DEPENDENCIES = stamp-arpc-repo
endif

.PHONY: rpcclean
rpcclean:
	rm -f pmap_prot.h pmap_prot.C

EXTRA_DIST = pmap_prot.x .cvsignore
CLEANFILES = core *.core *~ *.rpo pmap_prot.h pmap_prot.C stamp-arpc-repo
MAINTAINERCLEANFILES = Makefile.in
