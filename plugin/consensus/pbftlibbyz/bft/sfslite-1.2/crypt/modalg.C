// -*-c++-*-
/* $Id: modalg.C 1117 2005-11-01 16:20:39Z max $ */

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

#include "modalg.h"
#include "amisc.h"

#undef setbit

void
barrett::set (const bigint &m)
{
  assert (sgn (m) > 0);
  assert (m.getbit (0));
  mp = &m;
  k = ((*mp).nbits () + mpz_bitsperlimb - 1) / mpz_bitsperlimb;
  bk1 = 0;
  bk1.setbit ((k + 1) * mpz_bitsperlimb, 1);
  u = 0;
  u.setbit (2 * mpz_bitsperlimb * k, 1);
  u = u / *mp;
}

void
barrett::mpz_reduce (MP_INT *r, const MP_INT *a) const
{
  assert (a->_mp_size >= 0 && a->_mp_size <= 2 * k);

  mpz_tdiv_q_2exp (&q, a, (k-1) * mpz_bitsperlimb);
  q *= u;
  q >>= (k+1) * mpz_bitsperlimb;

  mpz_tdiv_r_2exp (&r1, a, (k+1) * mpz_bitsperlimb);
  r2 = q * *mp;
  r2.trunc ((k+1) * mpz_bitsperlimb);
  mpz_sub (r, &r1, &r2);

  if (mpz_sgn (r) < 0)
    mpz_add (r, r, &bk1);
  while (mpz_cmp (r, mp) > 0)
    mpz_sub (r, r, mp);
}

const bigint montgom::b (bigint (1) << ((int)(mpz_bitsperlimb)));

void
montgom::set (const bigint &m)
{
  mp = &m;
  assert (sgn (*mp) > 0 && mp->getbit (0));

  /* Yuck, but this isn't on the ciritcal path */
  bigint mitmp;
  mpz_invert (&mitmp, mp, &b);
  mitmp = b - mitmp;
  mi = mitmp.getui ();

  n = mp->_mp_size;
  r = 0;
  r.setbit (n * mpz_bitsperlimb, 1);
  rm = mod (r, *mp);
  ri = invert (rm, *mp);
  r2 = 0;
  r2.setbit (n * 2 * mpz_bitsperlimb, 1);
  r2 = mod (r2, *mp);
  mr = *mp * r;
}

void
montgom::mpz_mreduce (MP_INT *a, const MP_INT *t) const
{
  assert (t->_mp_size >= 0 && t->_mp_size <= 2 * n);
  assert (mpz_cmp (t, &mr) < 0);

  int sa = 2 * n + 1;
  if (a->_mp_alloc < sa)
    _mpz_realloc (a, sa);
  mpz_set (a, t);
  mp_limb_t *ap = a->_mp_d;
  bzero (ap + a->_mp_size, (sa - a->_mp_size) * sizeof (mp_limb_t));

  const mp_limb_t *mpp = mp->_mp_d;

  for (int i = 0; i < n; i++) {
    mp_limb_t u = ap[i] * mi;
    u = mpn_addmul_1 (ap + i, mpp, n, u);
    mpn_add_1 (ap + n + i, ap + n + i, n - i + 1, u);
  }

  while (sa && !ap[sa - 1])
    sa--;
  a->_mp_size = sa;
  mpz_tdiv_q_2exp (a, a, n * mpz_bitsperlimb);

  if (mpz_cmp (a, mp) >= 0)
    mpz_sub (a, a, mp);
}

void
montgom::mpz_mmul (MP_INT *a, const MP_INT *x, const MP_INT *y) const
{
  assert (x->_mp_size >= 0 && x->_mp_size <= n);
  assert (y->_mp_size >= 0 && y->_mp_size <= n);

  if (!x->_mp_size || !y->_mp_size) {
    a->_mp_size = 0;
    return;
  }

  MP_INT *rp = a;
  if (rp == x || rp == y)
    rp = &mmr;

  int sa = 2 * n + 1;
  if (rp->_mp_alloc < sa)
    _mpz_realloc (rp, sa);
  mp_limb_t *ap = rp->_mp_d;
  bzero (ap, sa * sizeof (mp_limb_t));

  const mp_limb_t *mpp = mp->_mp_d;
  const mp_limb_t *xp = x->_mp_d;
  const mp_limb_t *yp = y->_mp_d;
  int sx = x->_mp_size, sy = y->_mp_size;

  for (int i = 0; i < n; i++) {
    mp_limb_t xi = i < sx ? xp[i] : 0;
    mp_limb_t u = (ap[i] + xi * yp[0]) * mi;
    u = mpn_addmul_1 (ap + i, mpp, n, u);
    mpn_add_1 (ap + n + i, ap + n + i, n - i + 1, u);
    u = mpn_addmul_1 (ap + i, yp, sy, xi);
    mpn_add_1 (ap + sy + i, ap + sy + i, sa - sy - i, u);
  }

  while (sa && !ap[sa - 1])
    sa--;
  rp->_mp_size = sa;
  mpz_tdiv_q_2exp (rp, rp, n * mpz_bitsperlimb);

  if (mpz_cmp (rp, mp) >= 0)
    mpz_sub (rp, rp, mp);

  if (a == x || a == y)
    mmr.swap (a);
}

void
montgom::mpz_powm (MP_INT *a, const MP_INT *g, const MP_INT *e) const
{
  mpz_mmul (&gm, g, &r2);
  mpz_set (a, &rm);
  for (int i = mpz_sizeinbase2 (e); i-- > 0;) {
    mpz_mmul (a, a, a);
    if (mpz_getbit (e, i))
      mpz_mmul (a, a, &gm);
  }
  mpz_mreduce (a, a);
}
