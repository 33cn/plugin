#include "crypt.h"
#include "sha1.h"
#include "hashcash.h"
#include "serial.h"

/*
 *
 * Copyright (C) 1999 Frans Kaashoek and David Mazieres (kaashoek@lcs.mit.edu)
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

/* This is a simple implementation of hashcash for SFS.  It is inspired by
 * http://www.cypherspace.org/~adam/hashcash/
 */

static void 
addone (u_char *p, u_int len)
{
  for (int i = len - 1; i >= 0; i--) {
    p[i]++;
    if (p[i] != 0)
      break;
  }
}

static bool
check (const u_int32_t *l, const u_int32_t *r, unsigned int n)
{
  int nword = n / (sizeof(u_int32_t) * 8);
  int nbits = n % (sizeof(u_int32_t) * 8);
  int i;

  for (i = 0; i < nword; i++)
    if (l[i] != r[i])
      return false;

  if (nbits > 0) {
    u_int32_t mask = ~((1 << (sizeof(u_int32_t) * 8 - nbits)) - 1);
    return (l[i] & mask) == (r[i] & mask);
  }
  else
    return true;
}


/* Compute a payment such that the first "bitcost" bits of
 * SHA1-transfer(payment) match the first "bitcost" bits of the target.  The
 * SHA1-state is initialized to the value of inithash (e.g the hostid).  */

u_long 
hashcash_pay (char payment[sha1::blocksize],
	      const char inithash[sha1::hashsize], 
	      const char target[sha1::hashsize], unsigned int bitcost)
{
  u_int32_t state[sha1::hashwords];
  u_int32_t s[sha1::hashwords];
  u_int32_t t[sha1::hashwords];
  u_char *pay = reinterpret_cast<u_char *> (payment);

  rnd.getbytes (pay, sha1::blocksize);
  for (int i = 0; i < sha1::hashwords; i++) {
    s[i] = getint (inithash + 4 * i);
    t[i] = getint (target + 4 * i);
  }

  for (unsigned long j = 0; 1; j++) {
    for (int i = 0; i < sha1::hashwords; i++)
      state[i] = s[i];

    sha1::transform (state, pay);
    if (check (state, t, bitcost))
      return j;
    else
      addone (pay, sha1::blocksize);
  }
}


bool
hashcash_check (const char payment[sha1::blocksize],
		const char inithash[sha1::hashsize], 
		const char target[sha1::hashsize], unsigned int bitcost)
{
  u_int32_t s[sha1::hashwords];
  u_int32_t t[sha1::hashwords];
  const u_char *pay = reinterpret_cast<const u_char *> (payment);

  for (int i = 0; i < sha1::hashwords; i++) {
    s[i] = getint (inithash + 4 * i);
    t[i] = getint (target + 4 * i);
  }
  sha1::transform (s, pay);
  return check (s, t, bitcost);
}
