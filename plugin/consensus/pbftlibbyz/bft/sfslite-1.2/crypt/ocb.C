/* $Id: ocb.C 1117 2005-11-01 16:20:39Z max $ */

/*
 *
 * Copyright (C) 2002 David Mazieres (dm@uun.org)
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

#include "ocb.h"
#include "msb.h"
#include "serial.h"

/*
 * These functions treat a block as polynomials over Z_2 mod
 * x^128 + x^7 + x^2 + x + 1.
 */
void
ocb::lshift (ocb::blk *d, const ocb::blk &s)
{
  int carryin = 0;
  for (int i = blk::nc - 1; i >= 0; i--) {
    int carryout = s.c[i] < 0;
    d->c[i] = (s.c[i] << 1) | carryin;
    carryin = carryout;
  }
  if (carryin)
    d->c[blk::nc-1] ^= 0x87;
}
void
ocb::rshift (ocb::blk *d, const ocb::blk &s)
{
  int carryin = 0;
  for (int i = 0; i < blk::nc; i++) {
    int carryout = (s.c[i] & 1) << 7;
    d->c[i] = ((u_char) s.c[i] >> 1) | carryin;
    carryin = carryout;
  }
  if (carryin) {
    d->c[0] ^= 0x80;
    d->c[blk::nc-1] ^= 0x43;
  }
}

inline u_int
calc_l_size (size_t mms)
{
  u_int r = log2c (mms);
  if (r > 4)
    return r - 4;
  else
    return 1;
}

ocb::ocb (size_t mms)
  : maxmsg_size (mms), l_size (calc_l_size (mms)),
    l ((New blk[l_size + 2]) + 1)
{
}

ocb::~ocb ()
{
  bzero (l-1, (l_size + 2) * sizeof (l[0]));
  delete[] (l-1);
}

void
ocb::setkey (const void *key, u_int keylen)
{
  k.setkey (key, keylen);
  blkclear (&l[0]);
  k.encipher_bytes (l[0].c);
  rshift (&l[-1], l[0]);
  for (u_int i = 0; i < l_size; i++)
    lshift (&l[i+1], l[i]);
}

void
ocb::encrypt (void *_ctext, blk *tag, u_int64_t nonce,
	      const void *_ptext, size_t len) const
{
  char *ctext = static_cast <char *> (_ctext);
  const char *ptext = static_cast <const char *> (_ptext);

  blk r;
  blkclear (&r);
  puthyper (r.c + (r.nc - 8), nonce);
  blkxor (&r, l[0]);
  k.encipher_bytes (r.c);

  blk s;
  blkclear (&s);

  size_t i = 1;
  blk tmp;
  while (len > blk::nc) {
    tmp.get (ptext);
    blkxor (&s, tmp);
    blkxor (&r, l[ffs (i) - 1]);

    blkxor (&tmp, r);
    k.encipher_bytes (tmp.c);
    blkxor (&tmp, r);
    tmp.put (ctext);

    ptext += blk::nc;
    ctext += blk::nc;
    len -= blk::nc;
    i++;
  };

  blkxor (&r, l[ffs (i) - 1]);
  blkxor (&tmp, l[-1], r);
  tmp.c[tmp.nc - 1] ^= len << 3;
  k.encipher_bytes (tmp.c);

  blkxor (&s, tmp);
  for (u_int b = 0; b < len; b++)
    s.c[b] ^= (ctext[b] = tmp.c[b] ^ ptext[b]);
  blkxor (&tmp, s, r);
  k.encipher_bytes (tag->c, tmp.c);
}

bool
ocb::decrypt (void *_ptext, u_int64_t nonce, const void *_ctext,
	      const blk *tag, size_t len) const
{
  char *ptext = static_cast <char *> (_ptext);
  const char *ctext = static_cast <const char *> (_ctext);

  blk r;
  blkclear (&r);
  puthyper (r.c + (r.nc - 8), nonce);
  blkxor (&r, l[0]);
  k.encipher_bytes (r.c);

  blk s;
  blkclear (&s);

  size_t i = 1;
  blk tmp;
  while (len > blk::nc) {
    blkxor (&r, l[ffs (i) - 1]);

    tmp.get (ctext);
    blkxor (&tmp, r);
    k.decipher_bytes (tmp.c);
    blkxor (&tmp, r);
    tmp.put (ptext);

    blkxor (&s, tmp);

    ptext += blk::nc;
    ctext += blk::nc;
    len -= blk::nc;
    i++;
  };

  blkxor (&r, l[ffs (i) - 1]);
  blkxor (&tmp, l[-1], r);
  tmp.c[tmp.nc - 1] ^= len << 3;
  k.encipher_bytes (tmp.c);
  
  blkxor (&s, tmp);
  for (u_int b = 0; b < len; b++) {
    s.c[b] ^= ctext[b];
    ptext[b] = tmp.c[b] ^ ctext[b];
  }
  blkxor (&tmp, s, r);
  k.encipher_bytes (tmp.c);
  return !memcmp (tag->c, tmp.c, tag->nc);
}

#if 0
#include "str.h"
str
blk2str (const ocb::blk &s)
{
  strbuf sb;
  for (int i = 0; i < s.nc; i++)
    for (int b = 7; b >= 0; b--)
      sb << (s.c[i] & (1 << b) ? "1" : ".");
  return sb;
}
#endif

