// -*-c++-*-
/* $Id: blowfish.C 3758 2008-11-13 00:36:00Z max $ */

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

/* This code is derived from a public domain implementation of
 * blowfish found on the net.  I don't remember the original author.
 * By this point, the code has probably been hacked beyond
 * recognition, anyway.  */

#include "crypt.h"
#include "blowfish_data.h"

/* Prepend a sha-1 hash of the key to the actual key.  This ensures
 * that every key affects the output, even when the key length is
 * longer than is supported by the cipher. */
void
block64cipher::setkey_sha1 (const void *key, size_t len)
{
  sha1ctx sc;
  zucbuf k (sc.hashsize + len);
  sc.update (key, len);
  sc.final (k);
  memcpy (k + sc.hashsize, key, len);
  setkey (k, sc.hashsize + len);
}

void
block64cipher::encipher_bytes (u_char block[8])
{
  u_int32_t l, r;
  l = getint (block);
  r = getint (block + 4);
  encipher (&l, &r);
  putint (block, l);
  putint (block + 4, r);
}

void
block64cipher::decipher_bytes (u_char block[8])
{
  u_int32_t l, r;
  l = getint (block);
  r = getint (block + 4);
  decipher (&l, &r);
  putint (block, l);
  putint (block + 4, r);
}

inline u_int32_t
blowfish::F (u_int32_t x) const
{
  return ((S[0][x >> 24] + S[1][x >> 16 & 0xff])
	  ^ S[2][x >> 8 & 0xff]) + S[3][x & 0xff];
}

void
blowfish::encipher (u_int32_t *xl, u_int32_t *xr) const
{
  u_int32_t Xl = *xl;
  u_int32_t Xr = *xr;

  for (int i = 0; i < BF_N;) {
    Xl ^= P[i++];
    Xr ^= F (Xl);

    Xr ^= P[i++];
    Xl ^= F (Xr);
  }

  *xr = Xl ^ P[BF_N];
  *xl = Xr ^ P[BF_N + 1];
}

void
blowfish::decipher (u_int32_t *xl, u_int32_t *xr) const
{
  u_int32_t Xl = *xl;
  u_int32_t Xr = *xr;

  for (int i = BF_N + 1; i > 1;) {
    Xl ^= P[i--];
    Xr ^= F (Xl);

    Xr ^= P[i--];
    Xl ^= F (Xr);
  }

  *xr = Xl ^ P[1];
  *xl = Xr ^ P[0];
}

void
blowfish::initstate ()
{
  const u_int32_t *idp = pihex;

  /* Load up the initialization data */
  for (int i = 0; i < BF_N + 2; ++i)
    P[i] = *idp++;
  for (int i = 0; i < 4; ++i)
    for (int j = 0; j < 256; ++j)
      S[i][j] = *idp++;
}

void
blowfish::keysched (const void *_key, size_t keybytes)
{
  const u_char *key = static_cast <const u_char *> (_key);

  if (keybytes > 0)
    for (u_int i = 0, keypos = 0; i < BF_N + 2; ++i) {
      u_int32_t data = 0;
      for (int k = 0; k < 4; ++k) {
	data = (data << 8) | key[keypos++];
	if (keypos >= keybytes)
	  keypos = 0;
      }
      P[i] ^= data;
    }

  u_int32_t datal = 0, datar = 0;

  for (int i = 0; i < BF_N + 2; i += 2) {
    encipher (&datal, &datar);
    P[i] = datal;
    P[i + 1] = datar;
  }

  for (int i = 0; i < 4; ++i) {
    for (int j = 0; j < 256; j += 2) {
      encipher (&datal, &datar);
      S[i][j] = datal;
      S[i][j + 1] = datar;
    }
  }
}

void
blowfish::setkey (const void *key, size_t keybytes)
{
  initstate ();
  keysched (key, keybytes);
}

blowfish::~blowfish ()
{
  bzero (P, sizeof (P));
  bzero (S, sizeof (S));
}

/* OrpheanBeholderScryDoubt */
const u_char eksblowfish::cryptmsg[24] = {
  'O', 'r', 'p', 'h', 'e', 'a', 'n',
  'B', 'e', 'h', 'o', 'l', 'd', 'e', 'r',
  'S', 'c', 'r', 'y', 'D', 'o', 'u', 'b', 't'
};

class salter {
  const u_char *salt;
  size_t len;
  size_t pos;

  u_char getbyte () { if (pos >= len) pos = 0; return salt[pos++]; }
public:
  salter (const void *s, size_t l)
    : salt (static_cast<const u_char *> (s)), len (l), pos (0)
    { assert (len > 0); }
  u_int32_t getword () {
    return getbyte () << 24 | getbyte () << 16
      | getbyte () << 8 | getbyte ();
  }
};

void
eksblowfish::eksched (u_int cost, const void *_key, size_t keybytes,
		      const void *salt, size_t saltlen)
{
  assert (cost <= 32);
  u_int32_t nrounds = cost ? 1 << (cost-1) : 0;
#if 0
  u_int32_t saltw[4] = {
    getint (salt), getint (salt + 0x4),
    getint (salt + 0x8), getint (salt + 0xc)
  };
#endif

  const u_char *key = static_cast <const u_char *> (_key);

  if (keybytes > 0)
    for (u_int i = 0, keypos = 0; i < BF_N + 2; ++i) {
      u_int32_t data = 0;
      for (int k = 0; k < 4; ++k) {
	data = (data << 8) | key[keypos++];
	if (keypos >= keybytes)
	  keypos = 0;
      }
      P[i] ^= data;
    }

  salter sr (salt, saltlen);
  u_int32_t datal = 0, datar = 0;

  for (int i = 0; i < BF_N + 2; i += 2) {
    datal ^= sr.getword ();
    datar ^= sr.getword ();
    encipher (&datal, &datar);
    P[i] = datal;
    P[i + 1] = datar;
  }

  for (int i = 0; i < 4; ++i) {
    for (int j = 0; j < 256; j += 2) {
      datal ^= sr.getword ();
      datar ^= sr.getword ();
      encipher (&datal, &datar);
      S[i][j] = datal;
      S[i][j + 1] = datar;
    }
  }

  for (u_int32_t i = 0; i < nrounds; i++) {
    keysched (key, keybytes);
    keysched (salt, saltlen);
  }
}

str
eksblowfish::hashpwd (str pwd, str saltstr)
{
  enum { saltsize = 16 };
  u_int cost = 5;
  str salt;

  if (saltstr) {
    char *p;
    u_int c = strtol (saltstr, &p, 10);
    if (p != saltstr && c <= 16) {
      cost = c;
      if (*p++ == '$') {
	salt = dearmor64 (p);
	if (salt.len () != saltsize)
	  salt = NULL;
      }
    }
  }

  if (!salt) {
    mstr m (saltsize);
    rnd.getbytes (m, m.len ());
    salt = m;
  }

  initstate ();
  eksched (cost, pwd.cstr (), pwd.len (), salt.cstr (), salt.len ());
  u_char hpw[sizeof (eksblowfish::cryptmsg)];
  memcpy (hpw, eksblowfish::cryptmsg, sizeof (hpw));
  for (int i = 0; i < 64; i++) {
    encipher_bytes (hpw);
    encipher_bytes (hpw + 8);
    encipher_bytes (hpw + 16);
  }

  return strbuf () << cost << "$" << armor64 (salt)
		   << "$" << armor64 (hpw, sizeof (hpw));
}

void
cbc64iv::encipher_words (u_int32_t *dp, size_t len)
{
  assert (!(len & 7));
  u_int32_t *ep = dp + len / 4;
  u_int32_t Ivl = ivl, Ivr = ivr;

  while (dp < ep) {
    Ivl ^= dp[0];
    Ivr ^= dp[1];
    c.encipher (&Ivl, &Ivr);
    dp[0] = Ivl;
    dp[1] = Ivr;
    dp += 2;
  }
  ivl = Ivl;
  ivr = Ivr;
}

void
cbc64iv::decipher_words (u_int32_t *dp, size_t len)
{
  assert (!(len & 7));
  u_int32_t *ep = dp + len / 4;
  u_int32_t nivl = ivl, nivr = ivr;
  u_int32_t Ivl, Ivr;

  while (dp < ep) {
    Ivl = nivl;
    Ivr = nivr;
    nivl = dp[0];
    nivr = dp[1];
    c.decipher (&dp[0], &dp[1]);
    dp[0] ^= Ivl;
    dp[1] ^= Ivr;
    dp += 2;
  }
  ivl = nivl;
  ivr = nivr;
}

void
cbc64iv::encipher_bytes (void *_dp, size_t len)
{
  assert (!(len & 7));
  u_char *dp = static_cast<u_char *> (_dp);
  u_char *ep = dp + len;
  u_int32_t Ivl = ivl, Ivr = ivr;

  while (dp < ep) {
    Ivl ^= getint (dp);
    Ivr ^= getint (dp + 4);
    c.encipher (&Ivl, &Ivr);
    putint (dp, Ivl);
    putint (dp + 4, Ivr);
    dp += 8;
  }
  ivl = Ivl;
  ivr = Ivr;
}

void
cbc64iv::decipher_bytes (void *_dp, size_t len)
{
  assert (!(len & 7));
  u_char *dp = static_cast<u_char *> (_dp);
  u_char *ep = dp + len;
  u_int32_t nivl = ivl, nivr = ivr;
  u_int32_t Ivl, Ivr;

  while (dp < ep) {
    u_int32_t l = getint (dp);
    u_int32_t r = getint (dp + 4);
    Ivl = nivl;
    Ivr = nivr;
    nivl = l;
    nivr = r;
    c.decipher (&l, &r);
    l ^= Ivl;
    r ^= Ivr;
    putint (dp, l);
    putint (dp + 4, r);
    dp += 8;
  }
  ivl = nivl;
  ivr = nivr;
}
