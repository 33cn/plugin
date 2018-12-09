// -*-c++-*-
/* $Id: blowfish.h 1117 2005-11-01 16:20:39Z max $ */

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


#ifndef _BLOWFISH_H_
#define _BLOWFISH_H_

#include "sysconf.h"

class block64cipher {
public:
  virtual ~block64cipher () {}
  virtual void setkey (const void *, size_t) = 0;
  void setkey_sha1 (const void *, size_t);
  virtual void encipher (u_int32_t *, u_int32_t *) const = 0;
  virtual void decipher (u_int32_t *, u_int32_t *) const = 0;
  void encipher_bytes (u_char block[8]);
  void decipher_bytes (u_char block[8]);
};

class blowfish : public block64cipher {
protected:
  /* Note: BF_N must be divisible by 2! */
  enum { BF_N = 16 };
  static void enforce_bf_n_even ()
    { switch (1) case BF_N & 1: case 1:; }

  u_int32_t P[BF_N + 2];
  u_int32_t S[4][256];

  u_int32_t F (u_int32_t x) const;

  void initstate ();
  void keysched (const void *, size_t);

public:
  blowfish () {}
  blowfish (const void *key, size_t len) { setkey (key, len); }
  virtual ~blowfish ();

  void setkey (const void *, size_t);
  void encipher (u_int32_t *, u_int32_t *) const;
  void decipher (u_int32_t *, u_int32_t *) const;
};

/* Blowfish, but with a more expensive key schedule. */
class eksblowfish : public blowfish {
public:
  static const u_char cryptmsg[24];

  eksblowfish () {}
  void eksched (u_int cost, const void *key, size_t keybytes,
		const void *salt, const size_t saltlen);
  void eksetkey (u_int cost, const void *key, size_t keybytes,
		 const void *salt, const size_t saltlen)
    { initstate (); eksched (cost, key, keybytes, salt, saltlen); }

  /* The following function is deprecated. */
  str hashpwd (str pwd, str saltstr = NULL);
};

class cbc64iv {
  const block64cipher &c;
  u_int32_t ivl;
  u_int32_t ivr;

public:
  cbc64iv (const block64cipher &bc) : c (bc), ivl (0), ivr (0) {}

  void encipher_words (u_int32_t *data, size_t bytes);
  void decipher_words (u_int32_t *data, size_t bytes);
  void encipher_bytes (void *data, size_t bytes);
  void decipher_bytes (void *data, size_t bytes);

  void setiv (u_int32_t l, u_int32_t r) { ivl = l; ivr = r; }
  u_int32_t getivl () const { return ivl; }
  u_int32_t getivr () const { return ivr; }
};

#endif /* _BLOWFISH_H_ */
