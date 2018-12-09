/* $Id: rsa.C 1117 2005-11-01 16:20:39Z max $ */

/*
 *
 * Copyright (C) 1998 Kevin Fu (fubob@mit.edu)
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

/* RSA public key algorithm.  From:
 *
 */

#include "crypt.h"
#include "rsa.h"
#include "prime.h"

INITFN (scrubinit);

static void
scrubinit ()
{
  mp_setscrub ();
}

void
rsa_priv::init ()
{
  assert (p < q);
}

rsa_priv::rsa_priv (const bigint &n1, const bigint &n2)
  : rsa_pub (n1 * n2), p (n1), q (n2)
{
  bigint p1 (n1-1);
  bigint q1 (n2-1);
  phin = p1 * q1;
  d = mod (invert (e, phin), phin);

  /* Precompute for CRT */
  dp = mod (d, p1);
  dq = mod (d, q1);
  pinvq = mod (invert (p, q), q);

  // warn << "p " << p.getstr () << "\n";
  // warn << "q " << q.getstr () << "\n";
  // warn << "e " << e.getstr << "\n";
  // warn << "d " << d.getstr << "\n";

  init ();
}


ptr<rsa_priv>
rsa_priv::make (const bigint &n1, const bigint &n2)
{
  if (n1 == n2 || n1 <= 1 || n2 <= 1
      || !n1.probab_prime (5) || !n2.probab_prime (5))
    return NULL;
  return n1 < n2 ? New refcounted<rsa_priv> (n1, n2)
    : New refcounted<rsa_priv> (n2, n1);
}

rsa_priv
rsa_keygen (size_t nbits)
{
  random_init ();
  bigint p1 = random_srpprime (nbits/2 + (nbits & 1));
  bigint p2 = random_srpprime (nbits/2 + (nbits & 1));
  if (p1 > p2)
    swap (p1, p2);
  return rsa_priv (p1, p2);
}
