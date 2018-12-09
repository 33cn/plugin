## Process this file with automake to produce Makefile.in
## $Id: Makefile.am 2 2003-09-24 14:35:33Z max $

info_TEXINFOS = sfs.texi
man_MANS = sfs.7 $(TEXIMANS)

TEXIMANS = ssu.1 sfs_config.5 sfs_srp_params.5 sfs_users.5 sfsagent.1	\
sfsauthd.8 sfsauthd_config.5 sfscd.8 sfscd_config.5 sfskey.1 sfsrwsd.8	\
sfsrwsd_config.5 sfssd.8 sfssd_config.5 rex.1 funmount.8 dirsearch.1	\
vidb.8 newaid.1


INFODEPS = dirs-stamp version.texi
DVIDEPS = components.eps epsf.tex $(INFODEPS)

components.eps: $(srcdir)/components.fig
	fig2dev -L ps -e $(srcdir)/components.fig > components.eps

epsf.tex:
	@rm -f epsf.tex
	epsftex=`find $(prefix)/share/texmf \
		/usr/share/texmf \
		/usr/local/share/texmf \
		/usr/athena/share/texmf \
		-name epsf.tex | head -1`; \
		test "$$epsftex" && ln -s $$epsftex .

PATHFILT = sed -e 's!\([^ ]\)/!\1@dslash{}!g'
dirs-stamp: Makefile
	-@rm -f $(srcdir)/dirs.texi~
	-(echo @set PREFIX $(prefix); \
		echo @set ETCDIR $(etcdir); \
		echo @set SFSDIR $(sfsdir); \
		echo @set PKGDATADIR $(pkgdatadir); \
		echo @set PKGLIBDIR $(pkglibdir); \
		echo @set SFSLIBDIR $(sfslibdir)) \
		| $(PATHFILT) >> $(srcdir)/dirs.texi~ \
	&& mv -f $(srcdir)/dirs.texi~ $(srcdir)/dirs.texi
	touch dirs-stamp

sfs.info: $(INFODEPS)
sfs.dvi: $(DVIDEPS)

sfs.ps: sfs.dvi
	$(DVIPS) -Ppdf sfs.dvi -o sfs.ps

sfs.pdf: sfs.ps
	ps2pdf sfs.ps sfs.pdf

sfs.html: $(srcdir)/sfs.texi dirs-stamp
	output=`pwd`/$@; (cd $(srcdir) \
		&& makeinfo --html --no-headers --no-split sfs.texi \
			-o $$output)

stamp-man: $(srcdir)/sfs.texi $(srcdir)/manpager.pl
	-cd $(srcdir) && \
	  SFSDIR=$(sfsdir) VERSION=$(VERSION) ETCDIR=$(etcdir) \
	  PKGDATADIR=$(pkgdatadir) PKGLIBDIR=$(pkglibdir) \
	  SFSLIBDIR=$(sfslibdir) POD_MAN=$(PERL_POD_MAN) POD2MAN=$(POD2MAN) \
		$(PERL) manpager.pl sfs.texi .
	touch $(srcdir)/stamp-man

.PHONY: manclean
manclean: clean
	cd $(srcdir) && rm -f stamp-man $(man_MANS) sfs.info sfs.info-*
	rm -f epsf.tex sfs.html sfs.dvi sfs.ps sfs.pdf components.eps

$(TEXIMANS): stamp-man
	@:

$(srcdir)/sfs.7: $(srcdir)/sfs.pod
	@rm -f $@~
	pod2man $(srcdir)/sfs.pod > $@~
	mv -f $@~ $@

am_maintainer_clean = maintainer-clean
$(am_maintainer_clean): manclean

EXTRA_DIST = .cvsignore epsf.tex components.eps components.fig \
	README manpager.pl sfs.pod stamp-man dirs.texi $(man_MANS)
CLEANFILES = core *.core *~ sfs.tmp sfs.dvi sfs.ps $(srcdir)/stamp-vti
DISTCLEANFILES = sfs.html sfs.pdf dirs-stamp
MAINTAINERCLEANFILES = Makefile.in texinfo.tex epsf.tex components.eps	\
$(man_MANS) stamp-man $(srcdir)/dirs.texi $(srcdir)/sfs.7
