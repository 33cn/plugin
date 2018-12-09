// -*-c++-*-
/* $Id: ocb.h 1117 2005-11-01 16:20:39Z max $ */

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

#ifndef _CRYPT_OCB_H_
#define _CRYPT_OCB_H_ 1

#include "aes.h"

class ocb {
public:
  struct blk {
    enum { nc = 16, nl = nc / sizeof (long) };
    union {
      signed char c[nc];
      long l[nl];
    };
    void get (const void *s) { memcpy (c, s, nc); }
    void put (void *d) { memcpy (d, c, nc); }
  };

  static void blkclear (blk *d)
    { for (int i = 0; i < blk::nl; i++) d->l[i] = 0; }
  static void blkxor (blk *d, const blk &s1, const blk &s2)
    { for (int i = 0; i < blk::nl; i++) d->l[i] = s1.l[i] ^ s2.l[i]; }
  static void blkxor (blk *d, const blk &s)
    { for (int i = 0; i < blk::nl; i++) d->l[i] ^= s.l[i]; }
  static void lshift (blk *d, const blk &s);
  static void lshift (blk *d) { lshift (d, *d); }
  static void rshift (blk *d, const blk &s);
  static void rshift (blk *d) { rshift (d, *d); }

private:
  const size_t maxmsg_size;
  const u_int l_size;
  aes k;
  blk *l;

public:
  ocb (size_t max_msg_size);
  ~ocb ();
  void setkey (const void *key, u_int keylen);
  void encrypt (void *ctext, blk *tag, u_int64_t nonce,
		const void *ptext, size_t len) const;
  bool decrypt (void *ptext, u_int64_t nonce, const void *ctext,
		const blk *tag, size_t len) const;
};

#endif /* !_CRYPT_OCB_H_ */
