## Process this file with automake to produce Makefile.in
# $Id: Makefile.am 3462 2008-06-25 20:54:38Z max $

LDADD = $(LIBTAME) $(LIBSFSCRYPT) $(LIBARPC) $(LIBSAFEPTR) $(LIBASYNC) $(LIBGMP) 

TESTS = test_aes \
	test_aiod \
	test_armor \
	test_axprt \
	test_backoff \
	test_barrett \
	test_bbuddy \
	test_bitvec \
	test_blowfish \
	test_esign \
	test_itree \
	test_montgom \
	test_mpz_raw \
	test_mpz_square \
	test_mpz_xor \
	test_rabin \
	test_sha1 \
	test_srp \
	test_passfd \
	test_tiger \
	test_timecb \
	test_hashcash \
	test_schnorr \
	test_rctree \
	test_vec \
	test_sp1 \
	test_sp2 \
	test_sp3

check_PROGRAMS = $(TESTS)

test_aes_SOURCES = test_aes.C
test_aiod_SOURCES = test_aiod.C
test_armor_SOURCES = test_armor.C
test_axprt_SOURCES = test_axprt.C
test_backoff_SOURCES = test_backoff.C
test_barrett_SOURCES = test_barrett.C
test_bbuddy_SOURCES = test_bbuddy.C
test_bitvec_SOURCES = test_bitvec.C
test_blowfish_SOURCES = test_blowfish.C
test_esign_SOURCES = test_esign.C
test_hashcash_SOURCES = test_hashcash.C
test_itree_SOURCES = test_itree.C
test_montgom_SOURCES = test_montgom.C
test_mpz_raw_SOURCES = test_mpz_raw.C
test_mpz_square_SOURCES = test_mpz_square.C
test_mpz_xor_SOURCES = test_mpz_xor.C
test_passfd_SOURCES = test_passfd.C
test_rabin_SOURCES = test_rabin.C
test_sha1_SOURCES = test_sha1.C
test_srp_SOURCES = test_srp.C
test_tiger_SOURCES = test_tiger.C
test_timecb_SOURCES = test_timecb.C
test_schnorr_SOURCES = test_schnorr.C
test_rctree_SOURCES = test_rctree.C
test_vec_SOURCES = test_vec.C
test_sp1_SOURCES = test_sp1.C
test_sp2_SOURCES = test_sp2.C
test_sp3_SOURCES = test_sp3.C

$(check_PROGRAMS): $(LDEPS)

CLEANFILES = core *.core *~ *.rpo
MAINTAINERCLEANFILES = Makefile.in

EXTRA_DIST = .cvsignore
