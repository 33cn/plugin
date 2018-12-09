/* $Id: mpz_misc.C 3769 2008-11-13 20:21:34Z max $ */

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

#include "amisc.h"
#include "bigint.h"
#include "msb.h"

#undef min
#define min(a, b) (((a) < (b)) ? (a) : (b))

size_t
(mpz_sizeinbase2) (const MP_INT *mp)
{
  size_t i;
  for (i = ABS (mp->_mp_size); i-- > 0;) {
    mp_limb_t v;
    if ((v = mp->_mp_d[i]))
      return 8 * sizeof (mp_limb_t) * i + fls (v);
  }
  return 0;
}

void
_mpz_fixsize (MP_INT *r)
{
  mp_limb_t *sp = r->_mp_d;
  mp_limb_t *ep = sp + ABS (r->_mp_size);
  while (ep > sp && !ep[-1])
    ep--;
  r->_mp_size = r->_mp_size < 0 ? sp - ep : ep - sp;
  _mpz_assert (r);
}

void
mpz_umod_2exp (MP_INT *r, const MP_INT *a, u_long b)
{
  size_t nlimbs;
  const mp_limb_t *ap, *ae;
  mp_limb_t *rp, *re;

  if (a->_mp_size >= 0) {
    mpz_tdiv_r_2exp (r, a, b);
    return;
  }

  nlimbs = ((b + (8 * sizeof (mp_limb_t) - 1))
	    / (8 * sizeof (mp_limb_t)));
  if ((size_t) r->_mp_alloc < nlimbs)
    _mpz_realloc (r, nlimbs);
  ap = a->_mp_d;
  ae = ap + min ((size_t) ABS (a->_mp_size), nlimbs);
  rp = r->_mp_d;

  while (ap < ae)
    if ((*rp++ = -*ap++))
      goto nocarry;
  r->_mp_size = 0;
  return;

 nocarry:
  while (ap < ae)
    *rp++ = ~*ap++;
  re = r->_mp_d + nlimbs;
  while (rp < re)
    *rp++ = ~(mp_limb_t) 0;
  re[-1] &= ~(mp_limb_t) 0 >> ((8*sizeof (mp_limb_t) - b)
			       % (8*sizeof (mp_limb_t)));
  r->_mp_size = nlimbs;
  _mpz_fixsize (r);
}

int
mpz_getbit (const MP_INT *mp, size_t bit)
{
  long limb = bit / mpz_bitsperlimb;
  long nlimbs = mp->_mp_size;
  if (mp->_mp_size >= 0) {
    if (limb >= nlimbs)
      return 0;
    return mp->_mp_d[limb] >> bit % mpz_bitsperlimb & 1;
  }
  else {
    int carry;
    mp_limb_t *p, *e;

    nlimbs = -nlimbs;
    if (limb >= nlimbs)
      return 1;

    carry = 1;
    p = mp->_mp_d;
    e = p + limb;
    for (; p < e; p++)
      if (*p) {
	carry = 0;
	break;
      }
    return (~*e + carry) >> bit % mpz_bitsperlimb & 1;
  }
}

#if SIZEOF_LONG < 8
void
mpz_set_u64 (MP_INT *mp, u_int64_t val)
{
  if (mp->_mp_alloc * sizeof (mp_limb_t) < 8)
    _mpz_realloc (mp, (8 + sizeof (mp_limb_t) - 1) / sizeof (mp_limb_t));
  u_int i = 0;
  while (val) {
    mp->_mp_d[i++] = val;
    val >>= 8 * sizeof (mp_limb_t);
  }
  mp->_mp_size = i;
}

void
mpz_set_s64 (MP_INT *mp, int64_t val)
{
  if (val < 0) {
    mpz_set_u64 (mp, -val);
    mp->_mp_size = -mp->_mp_size;
  }
  else
    mpz_set_u64 (mp, val);
}

u_int64_t
mpz_get_u64 (const MP_INT *mp)
{
  int i = ABS (mp->_mp_size);
  if (!i)
    return 0;
#if GMP_LIMB_SIZE >= 8
  u_int64_t ret = mp->_mp_d[0];
#else /* GMP_LIMB_SIZE < 8 */
  u_int64_t ret = mp->_mp_d[--i];
  while (i > 0)
    ret = (ret << 8 * sizeof (mp_limb_t)) | mp->_mp_d[--i];
#endif /* GMP_LIMB_SIZE < 8 */
  if (mp->_mp_size > 0)
    return ret;
  return ~(ret - 1);
}

int64_t
mpz_get_s64 (const MP_INT *mp)
{
  int i = ABS (mp->_mp_size);
  if (!i)
    return 0;
#if GMP_LIMB_SIZE >= 8
  int64_t ret = mp->_mp_d[0];
#else /* GMP_LIMB_SIZE < 8 */
  int64_t ret = mp->_mp_d[--i];
  while (i > 0)
    ret = (ret << 8 * sizeof (mp_limb_t)) | mp->_mp_d[--i];
  if (mp->_mp_size > 0)
    return ret;
#endif /* GMP_LIMB_SIZE < 8 */
  return ~(ret - 1);
}
#endif /* SIZEOF_LONG < 8 */

/* To be called from the debugger */
extern "C" void mpz_dump (const MP_INT *);
void
mpz_dump (const MP_INT *mp)
{
  char *str = (char *) xmalloc (mpz_sizeinbase (mp, 16) + 3);
  mpz_get_str (str, 16, mp);
  strcat (str, "\n");
  v_write (2, str, strlen (str));
  xfree (str);
}
