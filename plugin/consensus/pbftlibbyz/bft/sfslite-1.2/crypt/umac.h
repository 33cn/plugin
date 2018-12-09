// -*-c++-*-
/* $Id: umac.h 1117 2005-11-01 16:20:39Z max $ */

/*
 *
 * Copyright (C) 2003 David Mazieres (dm@uun.org)
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU General Public License as
 * published by the Free Software Foundation; either version 2, or (at
 * your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 59 Temple Place, Suite 330, Boston, MA 02111-1307
 * USA
 *
 */

#ifndef _CRYPT_UMACS_H_
#define _CRYPT_UMACS_H_ 1


#include "aes.h"
#include "bigint.h"
#include "serial.h"

template<int n> struct umac_word;
template<> struct umac_word<1> {
  static u_int8_t ntohw (u_int8_t v) { return v; }
  static u_int8_t htonw (u_int8_t v) { return v; }
  static u_int8_t getwordbe (const void *dp)
    { return *static_cast<const u_int8_t *> (dp); }
  static void putwordbe (void *dp, u_int8_t v)
    { *static_cast<u_int8_t *> (dp) = v; }
};
template<> struct umac_word<2> {
  static u_int16_t ntohw (u_int16_t v) { return ntohs (v); }
  static u_int16_t htonw (u_int16_t v) { return htons (v); }
  static u_int16_t getwordbe (const void *_dp) {
    const u_char *dp = static_cast<const u_char *> (_dp);
    return dp[0] << 8 | dp[1];
  }
  static void putwordbe (void *_dp, u_int16_t v) {
    u_char *dp = static_cast<u_char *> (_dp);
    dp[0] = v >> 8;
    dp[1] = v;
  }
};
template<> struct umac_word<4> {
  static u_int32_t ntohw (u_int32_t v) { return ntohl (v); }
  static u_int32_t htonw (u_int32_t v) { return htonl (v); }
  static u_int32_t getwordbe (const void *dp) { return getint (dp); }
  static void putwordbe (void *dp, u_int32_t v) { putint (dp, v); }
};
template<> struct umac_word<8> {
  static u_int64_t ntohw (u_int64_t v) { return ntohq (v); }
  static u_int64_t htonw (u_int64_t v) { return htonq (v); }
  static u_int64_t getwordbe (const void *dp) { return gethyper (dp); }
  static void putwordbe (void *dp, u_int64_t v) { puthyper (dp, v); }
};
template<> struct umac_word<16> {
  static mpdelayed<const char *, size_t> getwordbe (const void *dp) {
    return mpdelayed<const char *, size_t> (mpz_set_rawmag_be,
					    static_cast<const char *> (dp),
					    16);
  }
};

template<int n> struct umac_prime;
template<> struct umac_prime<19> {
  typedef u_int32_t prime_t;
  enum { prime = 0x0007ffffU };
  enum { offset = 1 };
  enum { marker = prime - 1 };
};
template<> struct umac_prime<32> {
  typedef u_int32_t prime_t;
  enum { prime = 0xfffffffbU };
  enum { offset = 5 };
  enum { marker = prime - 1 };
  enum { maxword = 0xfffffffaU };
};
template<> struct umac_prime<36> {
  typedef u_int64_t prime_t;
  enum { prime = INT64 (0x0000000ffffffffbU) };
  enum { offset = 5 };
  enum { marker = prime - 1 };
};
template<> struct umac_prime<64> {
  typedef u_int64_t prime_t;
  enum { prime = INT64 (0xffffffffffffffc5U) };
  enum { offset = 59 };
  enum { marker = prime - 1 };
  enum { maxword = INT64 (0xffffffff00000000U) };
};
template<> struct umac_prime<128> {
  typedef bigint prime_t;
  static const prime_t prime;
  enum { offset = 159 };
  static const prime_t marker;
  static const prime_t maxword;
};

template<int n> struct umac_poly : umac_prime<n> {
  typedef typename umac_prime<n>::prime_t prime_t;
  prime_t yp;

  umac_poly () { poly_reset (); }
  void poly_reset () { yp = 1; }
  void poly_inner (prime_t k, prime_t m) {
    if (m >= this->maxword) {
      yp = (yp * k + this->marker) % this->prime;
      yp = (yp * k + (m - this->offset)) % this->prime;
    }
    else
      yp = (yp * k + m) % this->prime;
  }
};
template<> struct umac_poly<64> {
  typedef u_int64_t prime_t;
  static const bigint prime;
  static const bigint marker;
  static const bigint maxword;
  u_int64_t yp;

  umac_poly () { poly_reset (); }
  void poly_reset () { yp = 1; }
  void poly_inner (prime_t _k, prime_t _m) {
    bigint res (yp), k (_k), m (_m);
    if (m >= maxword) {
      res *= k;
      res += marker;
      res = mod (res, prime);
      res *= k;
      res += m;
      res = mod (res, prime);
    }
    else {
      res *= k;
      res += m;
      res = mod (res, prime);
    }
    yp = res.getu64 ();
  }
};
template<> struct umac_poly<128> : umac_prime<128> {
  prime_t yp;

  umac_poly () { poly_reset (); }
  void poly_reset () { yp = 1; }
  void poly_inner (const prime_t &k, const prime_t &m) {
    if (m >= maxword) {
      yp *= k;
      yp += marker;
      yp = mod (yp, prime);
      yp *= k;
      yp += m;
      yp = mod (yp, prime);
    }
    else {
      yp *= k;
      yp += m;
      yp = mod (yp, prime);
    }
  }
};

struct umac_u32_le : umac_word<4> {
  typedef u_int32_t word_t;
  typedef u_int64_t dword_t;

  enum { output_bits = 96 };
  enum { output_words = output_bits / (8 * sizeof (word_t)) };

  umac_poly<64>::prime_t k21[output_words];
  umac_poly<128>::prime_t k22[output_words];
  dword_t k31[output_words][8];
  word_t k32[output_words];

  umac_poly<64> y1[output_words];
  umac_poly<128> y2[output_words];

  static word_t getword (const void *_dp) {
    const u_int8_t *dp = static_cast<const u_char *> (_dp);
    return dp[0] | dp[1] << 8 | dp[2] << 16 | dp[3] << 24;
  }
  static dword_t nh_inner (const word_t *k, const word_t *m) {
    return (dword_t ((k[0] + m[0])) * (k[4] + m[4])
	    + dword_t ((k[1] + m[1])) * (k[5] + m[5])
	    + dword_t ((k[2] + m[2])) * (k[6] + m[6])
	    + dword_t ((k[3] + m[3])) * (k[7] + m[7]));
  }

  word_t l3hash (int polyno, u_int64_t val);
  word_t l3hash (int polyno, bigint val);

  void setkey2 (const aes_e &ek);
  void poly_reset ();
  void poly_set (int polyno, dword_t val) { y1[polyno].yp = val; }
  void poly_update (int polyno, dword_t val);
  void poly_final (void *dp);
};

struct umac : umac_u32_le {
  enum {  l1_key_len = 1024 };

  enum { word_size = sizeof (word_t), word_size_mask = sizeof (word_t) - 1 };
  enum { l1_key_words = l1_key_len / word_size };
  enum { l1_key_shift = 16 / word_size };

  enum { l1_block_size = 32 };

  static void size_sanity ();

  aes_e kpad;
  word_t k1[l1_key_len + output_words * l1_key_shift];

  union {
    char cbuf[l1_key_len];
    word_t wbuf[l1_key_len/sizeof (word_t)];
  };
  u_int l1len;
  size_t totlen;

  static void kdf (void *out, u_int nbytes, const aes_e &ek, u_int8_t index);
  static void kdfw (word_t *out, u_int nbytes, const aes_e &, u_int8_t);

  static dword_t nh (const word_t *k, const word_t *m);
  static dword_t nh (const word_t *k, const word_t *m, u_int nbytes);

  enum { mask32 = 0x1fffffffU };
  enum { mask64 = INT64 (0x01ffffff01ffffffU) };
  static const bigint mask128;

  void consume ();

  void setkey (const void *key, u_int keylen);
  void reset ();
  void update (const void *dp, size_t len);
  void final (void *mac);
};

#endif /* !_CRYPT_UMACS_H_ */
