/* $Id: rabin.C 1117 2005-11-01 16:20:39Z max $ */

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

/* Rabin-Williams public key algorithm.  From:
 *    A Modification of the RSA Public-Key Encryption Procedure.
 *    H. C. Williams, IEEE Transactions on Information Theory,
 *    Vol. IT-26, No. 6, November, 1980.
 *
 * The cipher is based on Rabin's original signature scheme.  This is
 * NOT RSA, and is NOT covered by the RSA patent.  (That's not to say
 * they couldn't try to sue you anyway for using this, but it's
 * unlikely.  However, if you use this code, you must agree to assume
 * all responsibility for patent infringement.)
 *
 * Two prime numbers p, q, chosen such that:
 *    p = 3 mod 8
 *    q = 7 mod 8
 *
 * [Note:  This means pq = 5 mod 8, so J(2, pq) = -1.]
 *
 * Define N and k as:
 *    N = pq
 *    k = 1/2 (1/4 * (p-1) * (q-1) + 1);
 *    [k = ((p-1) * (q-1) + 4) / 8;]
 *
 * The public key is (N).
 * The private key is (N, k).
 *
 * Define the following four operations.  (Note D2 can only be
 * performed with the secret key.)
 *
 *          { 4(2M+1)       if J(2M+1, N) = 1
 * E1(M) =  { 2(2M+1)       if J(2M+1, N) = -1
 *          { key cracked!  if J(2M+1, N) = 0 (completely improbable)
 *
 * E2(M) =  M^2 % N
 *
 * D2(M) =  M^k % N
 *
 *          { (M/4-1)/2      when M = 0 mod 4
 * D1(M) =  { ((N-M)/4-1)/2  when M = 1 mod 4
 *          { (M/2-1)/2      when M = 2 mod 4
 *          { ((N-M)/2-1)/2  when M = 3 mod 4
 *
 * Public key operations on messages M < (N-4)/8:
 *   Encrypt:  E2(E1(M))
 *   Decrypt:  D1(D2(M))
 *   Sign:     D2(E1(M))
 *   Verify:   D1(E2(M))
 */

#include "crypt.h"
#include "rabin.h"
#include "prime.h"

INITFN (scrubinit);

static void
scrubinit ()
{
  mp_setscrub ();
}

bool
rabin_pub::E1 (bigint &m, const bigint &in) const
{
  m = in << 1;
  m += 1;
  switch (jacobi (m, n)) {
  case 1:
    m <<= 2;
    break;
  case -1:
    m <<= 1;
    break;
  case 0:
    warn << "Key factored! jacobi (" << m << ", " << n << ") = 0\n";
    return false;
  }
  if (m >= n) {
    warn ("rabin_pub::E1: input too large\n");
    return false;
  }
  return true;
}

void
rabin_pub::D1 (bigint &m, const bigint &in) const
{
  switch (in.getui () & 3) {
  case 0:
    m = in - 4;
    m >>= 3;
    break;
  case 1:
    m = n - in;
    m -= 4;
    m >>= 3;
    break;
  case 2:
    m = in - 2;
    m >>= 2;
    break;
  case 3:
    m = n - in;
    m -= 2;
    m >>= 2;
    break;
  }
}

void
rabin_pub::E2 (bigint &m, const bigint &in) const
{
  //m = in * in;
  mpz_square (&m, &in);
  m %= n;
}

/* Calculate m = {in}^k % n.  Use Chinese remainder theorem for speed. */
void
rabin_priv::D2 (bigint &m, const bigint &in, int rsel) const
{
  /* Multiply input by random r = (ri)^{-2} mod n, to randomize the
   * timing of the modular reductions. */
  bigint r, ri;
  r = random_bigint (n.nbits () - 1);
  mpz_square (&ri, &r);
  ri %= n;
  mpz_square (&r, &ri);
  r = invert (r, n);
  r *= in;
  r %= n;

  /* find op, oq such that out % p = op, out % q = oq */
  bigint op (powm (r, kp, p));
  bigint oq (powm (r, kq, q));

  /* rsel selects which of 4 square roots */
  if (rsel & 1)
    op = p - op;
  if (rsel & 2)
    oq = q - oq;

  /* m = (((op - oq) * u) % p) * q + oq; */
  m = op - oq;
  m *= u;
  m = mod (m, p);
  m *= q;
  m += oq;

  /* Divide r back out */
  m *= ri;
  m %= n;
}

void
rabin_priv::init ()
{
  assert (p < q);

  u = mod (invert (q, p), p);

  bigint p1 = p - 1;
  bigint q1 = q - 1;

  kp = (p1 * q1 + 4) >> 3;
  kq = kp % q1;
  kp %= p1;
}

rabin_priv::rabin_priv (const bigint &n1, const bigint &n2)
  : rabin_pub (n1 * n2), p (n1), q (n2)
{
  init ();
}


ptr<rabin_priv>
rabin_priv::make (const bigint &n1, const bigint &n2)
{
  if (n1 == n2 || n1 <= 1 || n2 <= 1
      || !n1.probab_prime (5) || !n2.probab_prime (5))
    return NULL;
  return n1 < n2 ? New refcounted<rabin_priv> (n1, n2)
    : New refcounted<rabin_priv> (n2, n1);
}

static const u_int sieve_3_mod_4[4] = { 3, 2, 1, 4 };
static const u_int sieve_3_mod_8[8] = { 3, 2, 1, 8, 7, 6, 5, 4 };
static const u_int sieve_7_mod_8[8] = { 7, 6, 5, 4, 3, 2, 1, 8 };

rabin_priv
rabin_keygen (size_t bits, u_int iter)
{
  random_init ();
  bigint p1 = random_prime (bits/2 + (bits & 1), sieve_3_mod_4, 4, iter);
  bigint p2 = random_prime (bits/2 + 1,
			    p1.getbit (2) ? sieve_3_mod_8 : sieve_7_mod_8,
			    8, iter);
  if (p1 > p2)
    swap (p1, p2);
  return rabin_priv (p1, p2);
}
