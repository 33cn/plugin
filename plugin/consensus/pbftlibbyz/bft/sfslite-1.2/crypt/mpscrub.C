/* $Id: mpscrub.C 1117 2005-11-01 16:20:39Z max $ */

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

#include "crypt.h"
#include "msb.h"

bool mp_scrubbing;

#ifndef DMALLOC

static inline size_t
size (size_t n)
{
  return n ? (1 << log2c (n + MALLOCRESV)) - MALLOCRESV : 0;
}

static void *
scrub_alloc (size_t n)
{
  return xmalloc (size (n));
}

static void
scrub_free (void *p, size_t n)
{
  bzero (p, n);
  xfree (p);
}

static void *
scrub_realloc (void *p, size_t o, size_t n)
{
  void *p2;
  size_t o2 = size (o);

  if (n <= o2) {
    if (n < o)
      bzero ((char *) p + n, o - n);
    return p;
  }

  p2 = xmalloc (size (n));
  memcpy (p2, p, o);
  bzero (p, o);
  xfree (p);
  return p2;
}

#else /* DMALLOC */

static void *
scrub_alloc (size_t n)
{
  return xmalloc (n);
}

static void
scrub_free (void *p, size_t n)
{
  bzero (p, n);
  xfree (p);
}

static void *
scrub_realloc (void *p, size_t o, size_t n)
{
  void *p2 = xmalloc (n);
  memcpy (p2, p, o);
  bzero (p, o);
  xfree (p);
  return p2;
}

#endif /* DMALLOC */

void
mp_setscrub ()
{
  mp_scrubbing = true;
  mp_set_memory_functions (scrub_alloc, scrub_realloc, scrub_free);
}

void
mp_clearscrub ()
{
  mp_scrubbing = false;
  mp_set_memory_functions (NULL, NULL, NULL);
}
