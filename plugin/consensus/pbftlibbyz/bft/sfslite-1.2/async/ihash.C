/* $Id: ihash.C 2414 2006-12-16 00:58:37Z max $ */

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
#include "ihash.h"
#include "msb.h"

/* Some prime numbers <= powers of 2 */
const u_int exp2primes[33] = {
  0x1, /* place holder */
  0x2, 0x3, 0x7, 0xd,
  0x1f, 0x3d, 0x7f, 0xfb,
  0x1fd, 0x3fd, 0x7f7, 0xffd,
  0x1fff, 0x3ffd, 0x7fed, 0xfff1,
  0x1ffff, 0x3fffb, 0x7ffff, 0xffffd,
  0x1ffff7, 0x3ffffd, 0x7ffff1, 0xfffffd,
  0x1ffffd9, 0x3fffffb, 0x7ffffd9, 0xfffffc7,
  0x1ffffffd, 0x3fffffdd, 0x7fffffff, 0xfffffffb,
};

#define hteof(p) ((_ihash_entry *) ((char *) (p) + eos))

void
_ihash_grow (_ihash_table *htp, const size_t eos)
{
  u_int nbuckets;
  void **ntab;
  void *p, *np;
  size_t i;

  nbuckets = exp2primes[log2c(htp->buckets)+1];
  if (nbuckets < 3)
    nbuckets = 3;

  ntab = New (void * [nbuckets]);
  bzero (ntab, nbuckets * sizeof (*ntab));

  for (i = 0; i < htp->buckets; i++)
    for (p = htp->tab[i]; p; p = np) {
      _ihash_entry *htep = hteof (p);
      size_t ni = htep->val % nbuckets;
      np = htep->next;

      htep->next = ntab[ni];
      htep->pprev = &ntab[ni];
      if (ntab[ni])
	hteof(ntab[ni])->pprev = &htep->next;
      ntab[ni] = p;
    }

  delete[] htp->tab;
  htp->tab = ntab;
  htp->buckets = nbuckets;
}
