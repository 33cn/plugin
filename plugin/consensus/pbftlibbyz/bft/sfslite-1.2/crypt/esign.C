/* $Id: esign.C 3758 2008-11-13 00:36:00Z max $ */

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

#include "esign.h"
#include "crypt.h"
#include "prime.h"

#undef setbit

void
esign_pub::msg2bigint (bigint *resp, const str &msg, int bits)
{
  assert (bits);
  bits--;
  const size_t bytes = (bits + 7) >> 3;
  zcbuf buf (bytes);
  sha1oracle ora (bytes, 1);
  ora.update (msg.cstr (), msg.len ());
  ora.final (reinterpret_cast<u_char *> (buf.base));
  buf[bytes-1] &= 0xff >> (-bits & 7);
  mpz_set_rawmag_le (resp, buf, bytes);
}

int
esign_pub::calc_log2k (u_long k)
{
  assert (k > 4);
  int l = log2c (k);
  return k == (u_long) 1 << l ? l : -1;
}

esign_pub::esign_pub (const bigint &nn, u_long kk)
  : n (nn), k (kk), log2k (calc_log2k (k))
{
  size_t nb = mpz_sizeinbase2 (&n);
  nb = ((2 * nb) + 2) / 3;
  t.setbit (nb, 1);
}

bool
esign_pub::raw_verify (const bigint &z, const bigint &sig) const
{
  bigint u;
  kpow (&u, sig);
  if (u < z)
    return false;
  u -= t;
  return u <= z;
}

esign_priv::esign_priv (const bigint &p, const bigint &q, u_long kk)
  : esign_pub (p * p * q, kk), p (p), q (q), pq (p * q)
{
  assert (p > q);
}

void
esign_priv::precompute () const
{
  precomp &prc = prevec.push_back ();
  prc.x = random_zn (p);
  kpow (&prc.xk, prc.x);
  prc.x_over_kxk = prc.xk * k;
  prc.x_over_kxk = invert (prc.x_over_kxk, p);
  prc.x_over_kxk *= prc.x;
}

bigint
esign_priv::raw_sign (const bigint &v) const
{
  if (prevec.empty ()) {
    bigint x = random_zn (p);
    bigint xk;
    kpow (&xk, x);
    bigint w = v - xk;
    if (mpz_sgn (&w) < 0)
      w += n;
    mpz_cdiv_q (&w, &w, &pq);
    assert (mpz_sgn (&w) > 0);
#if 1
    xk *= k;
#else /* Don't notice a speedup */
    if (log2k < 0)
      xk *= k;
    else
      xk <<= log2k;
#endif
    xk = invert (xk, p);
    xk *= x;
    xk *= w;
    xk = mod (xk, p);
    return mod (x + xk * pq, n);
  }
  else {
    precomp &prc = prevec.front ();
    bigint w (v - prc.xk);
    if (mpz_sgn (&w) < 0)
      w += n;
    mpz_cdiv_q (&w, &w, &pq);
    assert (mpz_sgn (&w) > 0);
    w *= prc.x_over_kxk;
    w = mod (w, p);
    w *= pq;
    w += prc.x;
    w = mod (w, n);
    prevec.pop_front ();
    return w;
  }
}

esign_priv
esign_keygen (size_t nbits, u_long k)
{
  nbits = (nbits + 1) / 3;
  bigint p = random_prime (nbits);
  bigint q = random_prime (nbits);
  if (p < q)
    swap (p, q);
  return esign_priv (p, q, k);
}
