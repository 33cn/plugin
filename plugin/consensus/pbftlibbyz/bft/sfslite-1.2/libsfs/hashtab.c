/* $Id: hashtab.c 435 2004-06-02 15:46:36Z max $ */

/*
 *
 * Copyright (C) 1998, 1999 David Mazieres (dm@uun.org)
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

#include "sysconf.h"
#include "hashtab.h"

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

/* Highest bit set in a byte */
const char bytemsb[0x100] = {
  0, 1, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4, 4, 4, 4, 4, 5, 5, 5, 5, 5, 5, 5, 5,
  5, 5, 5, 5, 5, 5, 5, 5, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6,
  6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 7, 7, 7, 7, 7, 7, 7, 7,
  7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
  7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
  7, 7, 7, 7, 7, 7, 7, 7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
  8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
  8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
  8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
  8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
  8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
};

/* Find last set (most significant bit) */
static inline u_int
fls32 (u_int32_t v)
{
  if (v & 0xffff0000) {
    if (v & 0xff000000)
      return 24 + bytemsb[v>>24];
    else
      return 16 + bytemsb[v>>16];
  }
  if (v & 0x0000ff00)
    return 8 + bytemsb[v>>8];
  else
    return bytemsb[v];
}

static inline int
log2c (u_int32_t v)
{
  return v ? (int) fls32 (v - 1) : -1;
}

#define hteof(p) ((struct hashtab_entry *) ((char *) (p) + eos))

void
_hashtab_grow (struct _hashtab *htp, u_int eos)
{
  u_int nbuckets;
  void **ntab;
  void *p, *np;
  u_int i;

  /* warn ("_hashtab_grow\n"); */

  nbuckets = exp2primes[log2c(htp->ht_buckets)+1];
  if (nbuckets < 3)
    nbuckets = 3;
  ntab = malloc (nbuckets * sizeof (*ntab));
  if (!ntab)
    return;
  bzero (ntab, nbuckets * sizeof (*ntab));

  for (i = 0; i < htp->ht_buckets; i++)
    for (p = htp->ht_tab[i]; p; p = np) {
      struct hashtab_entry *htep = hteof (p);
      u_int ni = htep->hte_hval % nbuckets;
      np = htep->hte_next;

      htep->hte_next = ntab[ni];
      htep->hte_prev = &ntab[ni];
      if (ntab[ni])
	hteof(ntab[ni])->hte_prev = &htep->hte_next;
      ntab[ni] = p;
    }

  xfree (htp->ht_tab);
  htp->ht_tab = ntab;
  htp->ht_buckets = nbuckets;
}
