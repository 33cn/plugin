/* $Id: mpz_square.C 1117 2005-11-01 16:20:39Z max $ */

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

#include "sysconf.h"
#include "gmp.h"

#undef ABS
#define ABS(a) ((a) < 0 ? -(a) : (a))

#ifdef __cplusplus
extern "C" void mpz_square (MP_INT *, const MP_INT *);
#endif /* __cplusplus */
void
mpz_square (MP_INT *res, const MP_INT *arg)
{
  MP_INT tmp, *r;

  if (!arg->_mp_size) {
    res->_mp_size = 0;
    return;
  }
  if (res == arg) {
    r = &tmp;
    mpz_init (r);
  }
  else
    r = res;

  int asize = ABS (arg->_mp_size);
  const mp_limb_t *ap = arg->_mp_d;

  int rsize = 2 * asize;
  if (r->_mp_alloc < rsize)
    _mpz_realloc (r, rsize);
  mp_limb_t *rp = r->_mp_d;

  if (asize < 24) {
    /* This seems faster to be faster for small sizes. */
    mpn_mul_n (rp, ap, ap, asize);
  }
  else {
    bzero (rp, rsize * sizeof (mp_limb_t));
    for (int i = 1; i < asize; i++) {
      mp_limb_t *mrp = rp + (i<<1) - 1;
      mrp[asize - i] = mpn_addmul_1 (mrp, ap + i, asize - i, ap[i - 1]);
    }
    mpn_lshift (rp, rp, rsize, 1);
    for (int i = 0; i < asize; i++) {
      mp_size_t rpos = i << 1;
      mp_limb_t c = mpn_addmul_1 (rp + rpos, &ap[i], 1, ap[i]);
      mpn_add_1 (rp + rpos + 1, rp + rpos + 1, rsize - rpos - 1, c);
    }
  }

  while (rsize && !rp[rsize - 1])
    rsize--;
  r->_mp_size = rsize;
  if (res == arg) {
    mpz_clear (res);
    *(MP_INT *) res = tmp;
  }
}
