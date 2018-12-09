/* $Id: pad.C 1117 2005-11-01 16:20:39Z max $ */

/*
 *
 * Copyright (C) 1998 David Mazieres (dm@uun.org)
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

/*
 * pre_encrypt turns a message into a bigint suitable for public-key
 * encryption.  On input, nbits is the number of bits the resulting
 * bigint should contain, while msg is the message.  Msg consists of
 * an integral number of bytes, but the number of bits in the output
 * need not be a multiple of 8.
 *
 * Let || deonte concatenation with the left operand more significant
 * Let 0^{k} denote k 0 bits
 * Let |val| be the number of bits in a value
 *
 * Let G be the sha1oracle with index enc_gidx
 * Let H be the sha1oracle with index enc_hidx
 *
 * All numbers are written with the most significant byte to the left.
 * Thus, for instance 0x10 || 0x20 means 0x1020.
 *
 * When numbers are composed of C data structures, the bytes composing
 * the C data structures are interpreted little-endian order.  In
 * other words, if char a[2] = { 0x1, 0x2 } and char b[2] = { 0x40,
 * 0x80 }, then b || a means 0x80400201; the bytes in b are more
 * significant than those in a, but each of a and b is stored in
 * little-endian order in memory.  Likewise, if a = 0x12345678, then
 * running H (m) means running H on the input bytes { 0x78, 0x56,
 * 0x34, 0x12 }.
 *
 * Given as input a message "msg", and an output size "nbits",
 * pre_encrypt computes the following:
 *
 *        Mz = 0^{8*enc_zbytes} || msg
 *   padsize = nbits - |Mz|
 *         r = padsize random bits
 *        r' = 0^{-padsize % 8} || r
 *        Mg = Mz ^ first |Mz| bits of G(r')
 *        Mh = r ^ first |r| bits of H(Mg)
 *         R = Mh || Mg
 *
 * pre_encrypt returns R.
 *
 * The idea is to do something like the simpler:
 *    R = (r ^ H (Mz ^ G(r))) || (Mz ^ G(r))
 *
 * However, since the sha1oracle functions are defined in terms of
 * bytes and not bits, we prepend some 0 bits to the front of r before
 * calculating G.
 */

#include "crypt.h"

const int enc_zbytes = 16;
const int enc_minpad = 16;
const int enc_gidx = 1;
const int enc_hidx = 2;

const int sig_rndbytes = 16;
const int sig_minpad = 16;
const int sig_gidx = 3;
const int sigr_gidx = 4;

inline void
sha1oracle_lookup (int idx, char *dst, size_t dstlen,
		   const char *src, size_t srclen)
{
  sha1oracle ora (dstlen, idx);
  ora.update (src, srclen);
  ora.final (reinterpret_cast<u_char *> (dst));
}

bigint
pre_encrypt (str msg, size_t nbits)
{
  if (msg.len () + enc_zbytes + enc_minpad > nbits / 8) {
    warn ("pre_encrypt: message too large\n");
    return 0;
  }

  const char msbmask = 0xff >> (-nbits & 7);
  const size_t msgzlen = msg.len () + enc_zbytes;
  const size_t padsize = (nbits + 7) / 8 - msgzlen;
  zcbuf res (padsize + msgzlen);

  char *mp = res;
  char *hp = mp + msgzlen;

  rnd.getbytes (hp, padsize);
  hp[padsize-1] &= msbmask;
  sha1oracle_lookup (enc_gidx, mp, msgzlen, hp, padsize);
  for (size_t i = 0; i < msg.len (); i++)
    mp[i] ^= msg[i];

  zcbuf h (padsize);
  sha1oracle_lookup (enc_hidx, h, h.size, mp, msgzlen);
  for (size_t i = 0; i < padsize; i++)
    hp[i] ^= h[i];
  hp[padsize-1] &= msbmask;

  bigint r;
  mpz_set_rawmag_le (&r, res, res.size);
  return r;
}

str
post_decrypt (const bigint &m, size_t msglen, size_t nbits)
{
  const u_int msgzlen = msglen + enc_zbytes;
  const size_t padsize = (nbits + 7) / 8 - msgzlen;
  const char msbmask = 0xff >> (-nbits & 7);

  if (msglen + enc_zbytes + enc_minpad > nbits / 8) {
    warn ("post_decrypt: message too large\n");
    return NULL;
  }

  zcbuf msg ((nbits + 7) / 8);
  mpz_get_rawmag_le (msg, msg.size, &m);

  char *mp = msg;
  char *zp = msg + msglen;
  char *hp = zp + enc_zbytes;

  zcbuf h (padsize);
  sha1oracle_lookup (enc_hidx, h, h.size, mp, msgzlen);
  for (size_t i = 0; i < padsize; i++)
    hp[i] ^= h[i];
  hp[padsize-1] &= msbmask;

  zcbuf g (msgzlen);
  sha1oracle_lookup (enc_gidx, g, msgzlen, hp, padsize);
  for (size_t i = 0; i < msgzlen; i++)
    mp[i] ^= g[i];
  for (size_t i = 0; i < size_t (enc_zbytes); i++)
    if (zp[i])
      return NULL;		// Failure

  wmstr r (msglen);
  memcpy (r, mp, msglen);
  return r;
}

/*
 * pre_sign returns R from this calculation:
 *
 *   padsize = nbits - sha1::hashsize
 *         r = sig_rndbytes random bytes
 *        r' = 0^{padsize - 8*sig_rndbytes} || r
 *        M1 = SHA1 (M, r)
 *        Mg = r' ^ first padsize bytes of G(M1)
 *         R = Mg || M1
 */

bigint
pre_sign (sha1ctx *sc, size_t nbits)
{
  if (nbits/8 < sig_minpad + sig_rndbytes + sc->hashsize) {
    warn ("pre_sign: nbits too small\n");
    return 0;
  }

  zcbuf r (sig_rndbytes);
  rnd.getbytes (r, sig_rndbytes);

  zcbuf m ((nbits + 7) / 8);
  char *mp = m.base;
  sc->update (r, sig_rndbytes);
  sc->final (reinterpret_cast<u_char *> (mp));

  char *hp = mp + sc->hashsize;
  const size_t padsize = m.size - sc->hashsize;
  sha1oracle_lookup (sig_gidx, hp, padsize, mp, sc->hashsize);
  hp[padsize-1] &= 0xff >> (-nbits & 7);

  for (int i = 0; i < sig_rndbytes; i++)
    hp[i] ^= r[i];

  bigint res;
  mpz_set_rawmag_le (&res, m, m.size);
  return res;
}

bool
post_verify (sha1ctx *sc, const bigint &s, size_t nbits)
{
  if (nbits/8 < sig_minpad + sig_rndbytes + sc->hashsize) {
    warn ("post_verify: nbits too small\n");
    return false;
  }

  zcbuf m ((nbits + 7) / 8);
  mpz_get_rawmag_le (m.base, m.size, &s);

  char *mp = m;
  char *hp = mp + sc->hashsize;
  const size_t padsize = m.size - sc->hashsize;
  zcbuf g (padsize);
  sha1oracle_lookup (sig_gidx, g, g.size, mp, sc->hashsize);
  g[padsize-1] &= 0xff >> (-nbits & 7);

  if (memcmp (hp + sig_rndbytes, g + sig_rndbytes, padsize - sig_rndbytes))
    return false;

  for (int i = 0; i < sig_rndbytes; i++)
    hp[i] ^= g[i];
  sc->update (hp, sig_rndbytes);

  u_char mrh[sha1::hashsize];
  sc->final (mrh);
  return !memcmp (mrh, mp, sizeof (mrh));
}

bigint
pre_sign_r (str msg, size_t nbits)
{
  if (nbits/8 < sig_rndbytes + sha1::hashsize
      + max ((size_t) sig_minpad, msg.len ())) {
    warn ("pre_sign_r: nbits too small\n");
    return 0;
  }

  zcbuf r (sig_rndbytes);
  rnd.getbytes (r, sig_rndbytes);

  zcbuf m ((nbits + 7) / 8);
  char *mp = m.base;
  sha1ctx sc;
  sc.update (msg.cstr (), msg.len ());
  sc.update (r, sig_rndbytes);
  sc.final (reinterpret_cast<u_char *> (mp));

  char *hp = mp + sc.hashsize;
  const size_t padsize = m.size - sc.hashsize;
  sha1oracle_lookup (sigr_gidx, hp, padsize, mp, sc.hashsize);
  hp[padsize-1] &= 0xff >> (-nbits & 7);

  for (int i = 0; i < sig_rndbytes; i++)
    hp[i] ^= r[i];
  for (int i = sig_rndbytes, e = i + msg.len (); i < e; i++)
    hp[i] ^= msg[i - sig_rndbytes];

  bigint res;
  mpz_set_rawmag_le (&res, m, m.size);
  return res;
}

str
post_verify_r (const bigint &s, size_t msglen, size_t nbits)
{
  if (nbits/8 < sig_rndbytes + sha1::hashsize
      + max ((size_t) sig_minpad, msglen)) {
    warn ("post_verify_r: nbits too small\n");
    return NULL;
  }

  zcbuf m ((nbits + 7) / 8);
  mpz_get_rawmag_le (m.base, m.size, &s);

  char *mp = m;
  char *hp = mp + sha1::hashsize;
  const size_t padsize = m.size - sha1::hashsize;
  zcbuf g (padsize);
  sha1oracle_lookup (sigr_gidx, g, g.size, mp, sha1::hashsize);
  g[padsize-1] &= 0xff >> (-nbits & 7);

  for (u_int i = 0; i < padsize; i++)
    hp[i] ^= g[i];

  for (u_int i = sig_rndbytes + msglen; i < padsize; i++)
    if (hp[i])
      return NULL;

  sha1ctx sc;
  sc.update (hp + sig_rndbytes, msglen);
  sc.update (hp, sig_rndbytes);

  u_char mrh[sha1::hashsize];
  sc.final (mrh);
  if (memcmp (mrh, mp, sizeof (mrh)))
    return NULL;

  return str2wstr (str (hp + sig_rndbytes, msglen));
}
