// -*-c++-*-
/* $Id: dsa.C,v 1.2 2006/02/20 05:34:26 kaminsky Exp $ */

/*
 *
 * Copyright (C) 2006 Michael Kaminsky (kaminsky at csail.mit.edu)
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

#include "dsa.h"

bigint
dsa_pub::msghash_to_bigint (const str &msg)
{
  sha1ctx sc;
  sc.update (msg.cstr (), msg.len ());

  char buf[sha1::hashsize];
  sc.final (buf);

  assert (sizeof (buf) <= q.nbits());
  
  bigint r;
  mpz_set_rawmag_le (&r, buf, sizeof (buf));

  return r;
}

static int
build_index (bigint exparray[], int k, int i, int t)
{
  int idx = 0;

  int bitno = t - i;
  for (int j = k - 1; j >= 0; j--) {
    idx <<= 1;
    if (exparray[j].getbit(bitno))
      idx |= 1;
  }
  return idx;
}

/*
 * Compute (b1^{e1} * b2^{e2}) mod m 
 *
 * Should generalize to (b1^{e1} * ... * bn^{en}) mod m if
 * the function definition and some initializations are fixed up
 */
bigint
mulpowm (const bigint &b1, const bigint &e1,
         const bigint &b2, const bigint &e2, const bigint &m)
{
  int t = e1.nbits() > e2.nbits() ? e1.nbits() : e2.nbits();
  int index, k = 2;
  bigint basearray[2] = { b1, b2 };
  bigint exparray[2] = { e1, e2 };
  bigint G[1 << k];

  bigint tmp, res (1);
  for (int i = 1; i <= t; i++) {
    mpz_square (&tmp, &res);
    tmp %= m;
    index = build_index (exparray, k, i, t);
    assert (index >= 0 && index < (1 << k));
    if (!G[index]) {
      if (!index)
        G[0] = 1;
      else {
        for (int j = 0; j < k; j++) {
          if ((index & (1 << j))) {
            if (!G[index])
              G[index] = basearray[j];
            else {
              G[index] = G[index] * basearray[j];
              G[index] %= m;
            }
          }
        }
        if (!G[index])
          G[index] = New bigint (0);
      }
    }
    res = tmp * G[index];
    res %= m;
  }
  return res;
}

/*
 * w  = s^{-1} mod q
 * u1 = (SHA-1(M) * w) mod q
 * u2 = (r * w) mod q
 * v  = (g^{u1} * y^{u2} mod p) mod q.
 */
bool
dsa_pub::verify (const str &msg, const bigint &r, const bigint &s)
{
  if (r <= 0 || r >= q || s <= 0 || s >= q)
    return false;

  bigint w, u1, u2, v, t;

  w = invert (s, q);
  u1 = msghash_to_bigint (msg);
  u1 *= w;
  u1 %= q;
  u2 = r * w;
  u2 %= q;

  //v = mulpowm (g, u1, y, u2, p);
  //v %= q;

  v = powm (g, u1, p);
  t = powm (y, u2, p);
  v = v * t;
  v %= p;
  v %= q;

  return v == r;
}

/*
 * r = (g^k mod p) mod q
 * s = (k-1 * (SHA-1(M) + x * r)) mod q.
 */
void
dsa_priv::sign (bigint *r, bigint *s, const str &msg)
{
  assert (r && s);

  bigint k, kinv, m;

  do k = random_zn (q);
  while (k == 0);
  kinv = invert (k, q);

  *r = powm (g, k, p);
  *r %= q;

  m = msghash_to_bigint (msg);

  *s = x * r;
  *s += m;
  *s *= kinv;
  *s %= q;

  assert (*r != 0);
  assert (*s != 0);
}

void
dsa_gen::gen (u_int iter)
{
  bigint q, p, g, y, x;
  do {
    gen_q (&q);
  } while (!gen_p (&p, q, iter) || !q.probab_prime (iter));
  gen_g (&g, p, q);

  do x = random_zn (q);
  while (x == 0);
  y = powm (g, x, p);

  sk = New refcounted<dsa_priv> (p, q, g, y, x);
}
