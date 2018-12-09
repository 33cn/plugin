// -*-c++-*-
/* $Id: arena.h 3078 2007-09-20 20:21:55Z max $ */

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

#ifndef _ASYNC_ARENA_H_
#define _ASYNC_ARENA_H_ 1

#include "async.h"

class arena {
protected:
  enum { resv = sizeof (void *) };

  u_int size;
  u_int avail;
  char *chunk;
  char *cur;

  void newchunk (size_t);

 public:
  arena () {
    size = avail = 0;
    chunk = cur = 0;
  }

  void *alloc (size_t bytes, size_t align = sizeof (double)) {
    int pad = (align - (chunk - (char *) 0)) % align;
    if (avail < pad + bytes) {
      newchunk (bytes + align);
      pad = (align - (chunk - (char *) 0)) % align;
    }
    void *ret = cur + pad;
    cur += bytes + pad;
    avail -= bytes + pad;
    return ret;
  }

  char *(strdup) (const char *str)
    { return strcpy ((char *) alloc (1 + strlen (str), 1), str); }
#ifdef DMALLOC
  char *_strdup_leap (const char *, int, const char *str)
    { return strcpy ((char *) alloc (1 + strlen (str), 1), str); }
  char *dmalloc_strdup (const char *, int, const char *str, int)
    { return strcpy ((char *) alloc (1 + strlen (str), 1), str); }
  char *dmalloc_strndup (const char *, int, const char *str, int, int)
    { return strcpy ((char *) alloc (1 + strlen (str), 1), str); }
#endif /* DMALLOC */

  ~arena ();
};

inline void *
operator new (size_t n, arena &a)
{
  return a.alloc (n);
}

#endif /* !_ASYNC_ARENA_H_ */
