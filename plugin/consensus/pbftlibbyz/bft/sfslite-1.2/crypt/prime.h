// -*-c++-*-
/* $Id: prime.h 1117 2005-11-01 16:20:39Z max $ */

/*
 *
 * Copyright (C) 1999 David Mazieres (dm@uun.org)
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

#ifndef _SFSCRYPT_PRIME_H_
#define _SFSCRYPT_PRIME_H_ 1

#include "bigint.h"

extern const u_int small_primes[];
extern const int num_small_primes;
extern const u_int odd_sieve[2];
extern const u_int srpprime_sieve[];

class prime_finder {
public:
  enum { nprimes = 2048 };

private:
  bigint p;
  const u_int *const sieve;
  const u_int sievesize;
  u_int sievepos;
  int inc;
  int maxinc;
  bigint tmp;
  int mods[nprimes];

public:
  prime_finder (const bigint &p, const u_int *sv = odd_sieve, u_int svsz = 2);
  ~prime_finder ();

  void calcmods ();
  const bigint &getp () const { return p; }
  u_int getinc () const { return inc; }

  void setmax (int m) { assert (maxinc == -1 && m > 0); maxinc = m; }
  bigint &next_weak ();
  bigint &next_fermat ();
  bigint &next_strong (u_int iter = 32);
};

bigint random_zn (const bigint &n);
bigint random_bigint (size_t bits);
bool prime_test (const bigint &n, u_int iter = 32);
bigint prime_search (const bigint &base, u_int range,
		     const u_int *sieve = odd_sieve,
		     const u_int sievesize = 2, u_int iter = 32);
bool srpprime_test (const bigint &n, u_int iter = 32);
bigint srpprime_search (const bigint &start, u_int iter = 32);

inline bigint
random_prime (u_int nbits, const u_int *sieve = odd_sieve,
	      const u_int sievesize = 2, u_int iter = 32)
{
  bigint p;
  while (!(p = prime_search (random_bigint (nbits), 4 * sievesize * nbits,
			     sieve, sievesize, iter)))
    ;
  return p;
}

inline bigint
random_srpprime (u_int nbits)
{
  return srpprime_search (random_bigint (nbits - 1));
}

#endif /* !_SFSCRYPT_PRIME_H_ */
