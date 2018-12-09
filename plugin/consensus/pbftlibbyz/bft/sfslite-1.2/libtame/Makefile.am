
$(PROGRAMS): $(LDEPS)

sfslib_LTLIBRARIES = libtame.la

SUFFIXES = .C .T .h .Th
.T.C:
	$(TAME) -o $@ $< || (rm -f $@ && false)
.Th.h:
	$(TAME) -o $@ $< || (rm -f $@ && false)






pipeline.o: pipeline.C
pipeline.lo: pipeline.C
pipeline.C: pipeline.T $(TAME)

lock.o: lock.C
lock.lo: lock.C
lock.C: lock.T $(TAME)

io.o: io.C
io.lo: io.C
io.C: io.T $(TAME)

aio.o: aio.C
aio.lo: aio.C
aio.C: aio.T $(TAME)

rpcserver.o: rpcserver.C
rpcserver.lo: rpcserver.C
rpcserver.C: rpcserver.T $(TAME)

tame_connectors.h: tame_connectors.Th $(TAME)

tame_nlock.h: tame_nlock.Th $(TAME)



recycle.o:  tame_event_ag.h recycle.C
recycle.lo: tame_event_ag.h recycle.C

mkevent.o:  tame_event_ag.h mkevent.C
mkevent.lo: tame_event_ag.h mkevent.C

tfork.o:    tame_event_ag.h tame_tfork_ag.h tfork.C
tfork.lo:   tame_event_ag.h tame_tfork_ag.h tfork.C

io.o: tame_connectors.h tame_nlock.h
io.lo: tame_connectors.h tame_nlock.h

tame_event_ag.h: $(srcdir)/mkevent.pl
	$(PERL) $(srcdir)/mkevent.pl > $@ || (rm -f $@ && false)

tame_tfork_ag.h: $(srcdir)/mktfork_ag.pl
	$(PERL) $(srcdir)/mktfork_ag.pl > $@ || (rm -f $@ && false)


libtame_la_SOURCES = \
	recycle.C \
	closure.C \
	leak.C \
	init.C \
	run.C \
	mkevent.C \
	tfork.C \
	thread.C \
	trigger.C \
	event.C \
	event_green.C \
	 pipeline.C lock.C io.C aio.C rpcserver.C 

sfsinclude_HEADERS = \
	tame_event.h \
	tame_run.h \
	tame_recycle.h \
	tame_weakref.h \
	tame_closure.h \
	tame_rendezvous.h \
	tame_event_ag.h \
	tame_tfork.h \
	tame_tfork_ag.h \
	tame_thread.h \
	tame_typedefs.h \
	tame_slotset.h \
	tame_event_green.h \
	tame.h \
	tame_pipeline.h \
	tame_lock.h \
	tame_autocb.h \
	tame_trigger.h \
	tame_pc.h \
	tame_io.h \
	tame_aio.h \
	tame_rpcserver.h \
	tame_rpc.h \
	 tame_connectors.h tame_nlock.h

.PHONY: tameclean

tameclean:
	rm -f  pipeline.C lock.C io.C aio.C rpcserver.C  tame_connectors.h tame_nlock.h

clean:
	rm -rf  pipeline.C lock.C io.C aio.C rpcserver.C  tame_connectors.h tame_nlock.h *.o *.lo .libs _libs core *.core *~ *.rpo \
		tame_event_ag.h tame_tfork_ag.h *.la

dist-hook:
	cd $(distdir) && rm -f  pipeline.C lock.C io.C aio.C rpcserver.C

EXTRA_DIST = .svnignore  pipeline.T lock.T io.T aio.T rpcserver.T Makefile.am.m4 mkevent.pl mktfork_ag.pl  \
	 tame_connectors.Th tame_nlock.Th
CLEANFILES = core *.core *~ *.rpo tame_event_ag.h tame_tfork_ag.h  pipeline.C lock.C io.C aio.C rpcserver.C \
	 tame_connectors.h tame_nlock.h
MAINTAINERCLEANFILES = Makefile.in Makefile.am

$(srcdir)/Makefile.am: $(srcdir)/Makefile.am.m4
	@rm -f $(srcdir)/Makefile.am~
	$(M4) $(srcdir)/Makefile.am.m4 > $(srcdir)/Makefile.am~
	mv -f $(srcdir)/Makefile.am~ $(srcdir)/Makefile.am
