/* -*-c++-*- */
/* $Id: mpz_xor.C 1117 2005-11-01 16:20:39Z max $ */

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

#ifdef HAVE_CONFIG_H
#include <config.h>
#endif /* HAVE_CONFIG_H */

#ifndef HAVE_MPZ_XOR

#include "gmp.h"

#undef ABS
#define ABS(a) ((a) < 0 ? -(a) : (a))

static inline void
setlen1 (MP_INT *r, int max)
{
  mp_limb_t *lp = r->_mp_d;
  mp_limb_t *nz = lp;
  mp_limb_t *end = lp + max;
  for (;;) {
    mp_limb_t lv;
    if (lp == end)
      goto carryout;
    lv = *lp + 1;
    *lp++ = lv;
    if (lv) {
      nz = lp;
      break;
    }
  }
  while (lp < end)
    if (*lp++)
      nz = lp;
  r->_mp_size = r->_mp_d - nz;
  return;
 carryout:
  r->_mp_size = - ++max;
  if (max > r->_mp_alloc)
    _mpz_realloc(r, max);
  r->_mp_d[max-1] = 1;
}

#ifdef __cplusplus
extern "C" void mpz_xor (MP_INT *, const MP_INT *, const MP_INT *);
#endif /* __cplusplus */
void
mpz_xor (MP_INT *r, const MP_INT *a, const MP_INT *b)
{
  int sza = ABS (a->_mp_size);
  int szb = ABS (b->_mp_size);
  int i;

  if (sza < szb) {
    int szt;
    const MP_INT *t = a;
    a = b;
    b = t;
    szt = sza;
    sza = szb;
    szb = szt;
  }

  if (r->_mp_alloc < sza)
    _mpz_realloc(r, sza);

  if (a->_mp_size >= 0 && b->_mp_size >= 0) {
    int szr = 0;
    for (i = 0; i < szb; i++)
      if ((r->_mp_d[i] = a->_mp_d[i] ^ b->_mp_d[i]))
	szr = i + 1;
    for (; i < sza; i++)
      if ((r->_mp_d[i] = a->_mp_d[i]))
	szr = i + 1;
    r->_mp_size = szr;
  }
  else if (a->_mp_size >= 0) {
    for (i = 0; i < szb;) {
      mp_limb_t bl = b->_mp_d[i];
      r->_mp_d[i] = a->_mp_d[i] ^ (bl - 1);
      i++;
      if (bl)
	break;
    }
    for (; i < szb; i++)
      r->_mp_d[i] = a->_mp_d[i] ^ b->_mp_d[i];
    for (; i < sza; i++)
      r->_mp_d[i] = a->_mp_d[i];
    setlen1 (r, sza);
  }
  else if (b->_mp_size >= 0) {
    int borrow = 1;
    for (i = 0; i < szb; i++) {
      mp_limb_t al = a->_mp_d[i];
      r->_mp_d[i] = (al - borrow) ^ b->_mp_d[i];
      if (al)
	borrow = 0;
    }
    for (; i < sza; i++) {
      mp_limb_t al = a->_mp_d[i];
      r->_mp_d[i] = al - borrow;
      if (al)
	borrow = 0;
    }
    setlen1 (r, sza);
  }
  else {
    int szr = 0;
    int borrow = 1;
    for (i = 0; i < szb; ) {
      mp_limb_t al = a->_mp_d[i];
      mp_limb_t bl = b->_mp_d[i];
      if ((r->_mp_d[i] = (al - borrow) ^ (bl - 1)))
	szr = i + 1;
      if (al)
	borrow = 0;
      ++i;
      if (bl)
	break;
    }
    for (; i < szb; i++) {
      mp_limb_t al = a->_mp_d[i];
      if ((r->_mp_d[i] = (al - borrow) ^ b->_mp_d[i]))
	szr = i + 1;
      if (al)
	borrow = 0;
    }
    for (; i < sza; i++) {
      mp_limb_t al = a->_mp_d[i];
      if ((r->_mp_d[i] = al - borrow))
	szr = i + 1;
      if (al)
	borrow = 0;
    }
    r->_mp_size = szr;
  }
}

#endif /* !HAVE_MPZ_XOR */
