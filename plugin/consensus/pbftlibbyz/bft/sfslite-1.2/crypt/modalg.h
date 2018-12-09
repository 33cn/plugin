// -*-c++-*-
/* $Id: modalg.h 1117 2005-11-01 16:20:39Z max $ */

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


#ifndef _MODALG_H_
#define _MODALG_H_ 1

#include "bigint.h"

/* On the Pentium II, this sucks; it's slower than mpz_mod. */
class barrett {
  const bigint *mp;		// Modulus, M
  mp_size_t k;			// mpz_bitsperlimb * # of limbs in M
  bigint bk1;			// 2^((k+1) * mpz_bitsperlimb)
  bigint u;			// floor (2^mpz_bitsperlimb, M)

  /* Workspace */
  mutable bigint q, r1, r2;

  static void sred (MP_INT *r, const barrett *t, const MP_INT *a) 
    { t->mpz_reduce (r, a); }

public:
  barrett () : mp (NULL) {}
  explicit barrett (const bigint &m) { set (m); }
  void set (const bigint &m);

  /* Barrett reduction, returns a % M. */
  void mpz_reduce (MP_INT *r, const MP_INT *a) const;
  mpdelayed<const barrett *, const MP_INT *> reduce (const bigint &a) const
    { return mpdelayed<const barrett *, const MP_INT *> (sred, this, &a); }
};

class montgom {
  static const bigint b;	// 2^mpz_bitsperlimb

  const bigint *mp;		// Modulus, M
  mp_limb_t mi;			// -M^(-1) mod b
  int n;			// Number of limbs in M
  bigint r;			// 2^(n*mpz_bitsperlimb)
  bigint rm;			// r mod M
  bigint ri;			// r^(-1) mod M
  bigint r2;			// r^2 mod M
  bigint mr;			// m * r

  /* Workspace */
  mutable bigint mmr;
  mutable bigint gm;

  static void smred (MP_INT *r, const montgom *t, const MP_INT *a) 
    { t->mpz_mreduce (r, a); }
  static void smmul (MP_INT *r, const montgom *t,
		     const MP_INT *a, const MP_INT *b) 
    { t->mpz_mmul (r, a, b); }
  static void smexp (MP_INT *r, const montgom *t,
		     const MP_INT *a, const MP_INT *b) 
    { t->mpz_powm (r, a, b); }

public:
  montgom () : mp (NULL) {}
  explicit montgom (const bigint &m) { set (m); }
  void set (const bigint &m);

  const bigint &getr () const { return r; }
  const bigint &getri () const { return ri; }
  const bigint &getr2 () const { return r2; }

  /* Montgomery reduction, returns (t * ri) % M */
  void mpz_mreduce (MP_INT *a, const MP_INT *t) const;
  mpdelayed<const montgom *, const MP_INT *> mreduce (const bigint &a) const
    { return mpdelayed<const montgom *, const MP_INT *> (smred, this, &a); }

  /* Montgomery multiplication, returns (x * y * ri) % M */
  void mpz_mmul (MP_INT *a, const MP_INT *x, const MP_INT *y) const;
  mpdelayed<const montgom *, const MP_INT *, const MP_INT *>
  mreduce (const bigint &a, const bigint &b) const {
    return mpdelayed<const montgom *, const MP_INT *,
      const MP_INT *> (smmul, this, &a, &b);
  }

  /* Montgomery exponentiation, returs (g^e) % M */
  void mpz_powm (MP_INT *r, const MP_INT *g, const MP_INT *e) const;
  mpdelayed<const montgom *, const MP_INT *, const MP_INT *>
  powm (const bigint &g, const bigint &e) const {
    return mpdelayed<const montgom *, const MP_INT *,
      const MP_INT *> (smexp, this, &g, &e);
  }
};

#endif /* _MODALG_H_ */
