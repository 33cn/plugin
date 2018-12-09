## Process this file with automake to produce Makefile.in
# $Id: Makefile.am 2434 2006-12-19 22:24:45Z max $

LDADD = $(LIBSFSCRYPT) $(LIBARPC) $(LIBASYNC) $(LIBGMP) $(LDADD_THR)

sfslib_LTLIBRARIES = libsfscrypt.la
LIBSFSCRYPT = libsfscrypt.la

libsfscrypt_la_SOURCES = \
aes.C arc4.C axprt_crypt.C blowfish.C crypt_prot.C dsa.C elgamal.C     \
esign.C fips186.C getkbdnoise.C getsysnoise.C hashcash.C mdblock.C     \
modalg.C mpscrub.C mpz_misc.C mpz_raw.C mpz_square.C mpz_xor.C pad.C   \
paillier.C password.C pm.C poly.C prng.C rabin.C random_prime.C        \
rndseed.C rsa.C seqno.C serial.C sha1.C sha1oracle.C srp.C tiger.C     \
tiger_sboxes.C wmstr.C xdr_mpz_t.C schnorr.C ocb.C umac.C rabinpoly.C  \
rabin_fprint.C

sfsinclude_HEADERS = crypt_prot.x \
aes.h arc4.h axprt_crypt.h bench.h bigint.h blowfish.h crypt.h         \
crypthash.h crypt_prot.h dsa.h elgamal.h esign.h fips186.h hashcash.h  \
homoenc.h modalg.h paillier.h password.h pm.h poly.h prime.h prng.h    \
rabin.h rsa.h seqno.h sha1.h srp.h tiger.h wmstr.h schnorr.h ocb.h     \
umac.h rabinpoly.h rabin_fprint.h fprint.h


noinst_HEADERS = blowfish_data.h

crypt_prot.h: $(srcdir)/crypt_prot.x
	@rm -f $@
	-$(RPCC) -h $(srcdir)/crypt_prot.x || rm -f $@

crypt_prot.C: $(srcdir)/crypt_prot.x
	@rm -f $@
	-$(RPCC) -c $(srcdir)/crypt_prot.x || rm -f $@

dist-hook:
	cd $(distdir) && rm -f crypt_prot.h crypt_prot.C

srp.o srp.lo: crypt_prot.h
crypt_prot.o crypt_prot.lo: crypt_prot.h

noinst_PROGRAMS = tst timing
tst_SOURCES = tst.C
timing_SOURCES =  timing.C

#if STATIC
# MK: see comments in async/Makefile.am
tst_LDADD = $(LDADD)
tst_DEPENDENCIES = $(LIBSFSCRYPT) $(LIBARPC) $(LIBASYNC) 
#else
#tst_LDADD = $(LDADD)
#tst_DEPENDENCIES = 
#endif

if REPO
sfscrypt_repo_OBJECTS = $(libsfscrypt_la_OBJECTS) \
	$(LIBARPC) $(LIBASYNC)
stamp-sfscrypt-repo: $(sfscrypt_repo_OBJECTS)
	-$(CXXLINK) $(sfscrypt_repo_OBJECTS)
	@rm -f a.out
	touch $@
libsfscrypt_la_DEPENDENCIES = stamp-sfscrypt-repo
endif

.PHONY: rpcclean
rpcclean:
	rm -f crypt_prot.h crypt_prot.C

EXTRA_DIST = crypt_prot.x .cvsignore
CLEANFILES = core *.core *~ *.rpo stamp-sfscrypt-repo \
		crypt_prot.h crypt_prot.C
MAINTAINERCLEANFILES = Makefile.in
