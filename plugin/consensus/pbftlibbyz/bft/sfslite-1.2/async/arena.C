/* $Id: arena.C 1117 2005-11-01 16:20:39Z max $ */

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

#include "arena.h"
#include "msb.h"

void
arena::newchunk (size_t bytes)
{
  char *c;
#ifndef DMALLOC
  if (bytes < size)
    bytes = size;
  size = (1 << (log2c (bytes + MALLOCRESV) + 1)) - MALLOCRESV;
#else /* DMALLOC */
  /* Malloc each chunk seperately so dmalloc catches overrun bugs */
  size = bytes + resv;
#endif /* DMALLOC */
  avail = size - resv;
  c = (char *) xmalloc (size);
  *(void **) c = chunk;
  chunk = c;
  cur = c + resv;
  assert (bytes <= avail);
}

arena::~arena ()
{
  void *p, *np;
  for (p = chunk; p; p = np) {
    np = *(void **) p;
    xfree (p);
  }
}
