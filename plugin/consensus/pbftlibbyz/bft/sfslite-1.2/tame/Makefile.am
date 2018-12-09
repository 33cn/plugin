## Process this file with automake to produce Makefile.in
# $Id: Makefile.am 3056 2007-08-30 20:07:15Z max $

LDADD = $(LIBASYNC) 

sfsbin_PROGRAMS = tame

noinst_HEADERS = tame.h parse.h

tame_SOURCES = tame.C parse.yy scan.ll processor.C const.C output.C
tame_LDADD = $(LDADD) $(LDADD_THR)
tame_DEPENDENCIES = $(LIBASYNC)

YYFLAGS = -d

parse.h:
	@:

scan.cc: scan.ll
	$(LEX) $(AM_LFLAGS) $(LFLAGS) $(srcdir)/scan.ll \
		&& mv $(LEX_OUTPUT_ROOT).c $(srcdir)/scan.cc

parse.cc: parse.yy
	rm -f y.tab.c
	$(YACC) $(AM_YFLAGS) $(YYFLAGS) $(srcdir)/parse.yy
	case "$(YACC)" in					\
	*bison*)						\
		sed -e 's/union yyalloc/struct yyalloc/g'	\
			< y.tab.c > $(srcdir)/parse.cc ;	\
		rm -f y.tab.c					\
		;;						\
	*)							\
		mv -f y.tab.c $(srcdir)/parse.cc		\
		;;						\
	esac
	if test -f y.tab.h; then				\
		if cmp -s y.tab.h $(srcdir)/parse.h; then	\
			rm -f y.tab.h;				\
		else						\
			mv y.tab.h $(srcdir)/parse.h;		\
		fi;						\
	fi

parse.o: parse.cc 
# Note: certain versions of bison put __attribute__((unused)) after
# unused goto labels, which doesn't agree well with gcc 3.2.1.
# If you have problems, try the next line instead of the following one.
#	$(CXXCOMPILE) $(CXXNOERR) -D'__attribute__(x)=' -c $(srcdir)/parse.cc
	$(CXXCOMPILE) $(CXXNOERR) -c $(srcdir)/parse.cc

scan.o: parse.o scan.cc 
	$(CXXCOMPILE) $(CXXNOERR) -c $(srcdir)/scan.cc

.PHONY: parseclean
parseclean:
	rm -f $(srcdir)/parse.cc $(srcdir)/parse.h $(srcdir)/scan.cc

BUILT_SOURCES = parse.cc parse.h scan.cc

CLEANFILES = core *.core *~ *.rpo

EXTRA_DIST = .cvsignore
MAINTAINERCLEANFILES = Makefile.in parse.h parse.cc scan.cc
