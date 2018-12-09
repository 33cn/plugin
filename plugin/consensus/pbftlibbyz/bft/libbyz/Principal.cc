#include <stdlib.h>
#include <strings.h>
#include "Principal.h"
#include "Node.h"
#include "Reply.h"

#include "crypt.h"
#include "rabin.h"

Principal::Principal(int i, Addr a, char *p) {
  id = i;
  addr = a;
  last_fetch = 0;

  if (p == 0) { 
    pkey = 0;
    ssize = 0;
  } else {
    bigint b(p,16);
    ssize = (mpz_sizeinbase2(&b) >> 3) + 1 + sizeof(unsigned);
    pkey = new rabin_pub(b);
  }
  
  for (int j=0; j < 4; j++) {
    kin[j] = 0;
    kout[j] = 0; 
  }

#ifndef USE_SECRET_SUFFIX_MD5
  ctx_in = 0;
  ctx_out = umac_new((char*)kout);
#endif

  tstamp = 0;
  my_tstamp = zeroTime();
}


Principal::~Principal() { 
  delete pkey;
}

void Principal::set_in_key(const unsigned *k) { 
  memcpy(kin, k, Key_size);

#ifndef USE_SECRET_SUFFIX_MD5
  if (ctx_in)
    umac_delete(ctx_in);
  ctx_in = umac_new((char*)kin);
#endif
 
}

#ifdef USE_SECRET_SUFFIX_MD5
bool Principal::verify_mac(const char *src, unsigned src_len, 
			   const char *mac, unsigned *k) {
  // Do not accept MACs sent with uninitialized keys.
  if (k[0] == 0) return false;

  MD5_CTX context;
  unsigned int digest[4];

  MD5Init(&context);
  MD5Update(&context, src, src_len);
  MD5Update(&context, (char*)k, 16);
  MD5Final(digest, &context);
  return !memcmp(digest, mac, MAC_size);
}

void Principal::gen_mac(const char *src, unsigned src_len, 
			    char *dst, unsigned *k) {
  MD5_CTX context;
  unsigned int digest[4];

  MD5Init(&context);
  MD5Update(&context, src, src_len);
  MD5Update(&context, (char*)k, 16);
  MD5Final(digest, &context);

  // Copy to destination and truncate output to MAC_size
  memcpy(dst, (char*)digest, MAC_size);
}
#else
bool Principal::verify_mac(const char *src, unsigned src_len, 
			   const char *mac, const char *unonce, umac_ctx_t ctx) {
  // Do not accept MACs sent with uninitialized keys.
  if (ctx == 0) return false;

  char tag[20];
  umac(ctx, (char *)src, src_len, tag, (char *)unonce);
  umac_reset(ctx);
  return !memcmp(tag, mac, UMAC_size);
}

long long Principal::umac_nonce = 0;

void Principal::gen_mac(const char *src, unsigned src_len, 
			    char *dst, const char *unonce, umac_ctx_t ctx) {
  umac(ctx, (char *)src, src_len, dst, (char *)unonce);
  umac_reset(ctx);
}

#endif
 

void Principal::set_out_key(unsigned *k, ULong t) {
  if (t > tstamp) {
    memcpy(kout, k, Key_size);

#ifndef USE_SECRET_SUFFIX_MD5
    if (ctx_out)
      umac_delete(ctx_out);
    ctx_out = umac_new((char*)kout);
#endif

    tstamp = t;
    my_tstamp = currentTime();
  }
}


bool Principal::verify_signature(const char *src, unsigned src_len, 
				 const char *sig, bool allow_self) {
  // Principal never verifies its own authenticator.
  if ((id == node->id()) && !allow_self) return false;

  INCR_OP(num_sig_ver);
  START_CC(sig_ver_cycles);

  bigint bsig;
  int s_size;
  memcpy((char*)&s_size, sig, sizeof(int));
  sig += sizeof(int);
  if (s_size+(int)sizeof(int) > sig_size()) {
    STOP_CC(sig_ver_cycles);
    return false;
  }

  mpz_set_raw(&bsig, sig, s_size);  
  bool ret = pkey->verify(str(src, src_len), bsig);

  STOP_CC(sig_ver_cycles);
  return ret;
}


unsigned Principal::encrypt(const char *src, unsigned src_len, char *dst, 
			    unsigned dst_len) {
  // This is rather inefficient if message is big but messages will
  // be small.
  bigint ctext = pkey->encrypt(str(src, src_len));
  unsigned size = mpz_rawsize(&ctext);
  if (dst_len < size+2*sizeof(unsigned))
    return 0;

  memcpy(dst, (char*)&src_len, sizeof(unsigned));
  dst += sizeof(unsigned);
  memcpy(dst, (char*)&size, sizeof(unsigned));
  dst += sizeof(unsigned);

  mpz_get_raw(dst, size, &ctext);
  return size+2*sizeof(unsigned);
}

void random_nonce(unsigned *n) {
  bigint n1 = random_bigint(Nonce_size*8);
  mpz_get_raw((char*)n, Nonce_size, &n1);
}

int random_int() {
  bigint n1 = random_bigint(sizeof(int)*8);
  int i;
  mpz_get_raw((char*)&i, sizeof(int), &n1);
  return i;
}





