// -*-c++-*-
/* $Id: cbuf.h 1117 2005-11-01 16:20:39Z max $ */

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

#ifndef _ASYNC_CBUF_H_
#define _ASYNC_CBUF_H_ 1

#include "sysconf.h"

class cbuf {
  char *buf;
  size_t buflen;

  bool empty;
  size_t start;
  size_t end;

  iovec inv[2];
  iovec outv[2];

public:
  cbuf (size_t n)
    : buf ((char *) xmalloc (n)), buflen (n), empty (true),
      start (0), end (0) {}
  ~cbuf () { xfree (buf); }
  void resize (size_t n);
  void clear () { empty = false; start = end = 0; }

  size_t space () const
    { return (empty || start < end ? buflen : 0) + start - end; }
  size_t bytes () const { return buflen - space (); }
  const iovec *iniov ();
  int iniovcnt ()
    { return empty || start < end ? 1 + !!start : start != end; }
  void addbytes (size_t n);

  size_t size () const
    { return (empty || start < end ? 0 : buflen) + end - start; }
  char &at (size_t n) const
    { assert (n < size ()); return buf[(n + start) % buflen]; }
  char &operator[] (ptrdiff_t n) const { return at (n); }
  int find (char c);

  const iovec *outiov ();
  int outiovcnt () {
    if (empty)
      return 0;
    if (start < end)
      return 1;
    else
      return 1 + !!end;
  }
  void rembytes (size_t n) {
    if (n) {
      assert (n <= size ());
      start = (start + n) % buflen;
      empty = start == end;
    }
  }
  void unrembytes (size_t n) {
    if (n) {
      assert (n <= space ());
      start = (start >= n ? 0 : buflen) + start - n;
      empty = false;
    }
  }
  void copyout (void *_dst, size_t len);

private:
  cbuf (const cbuf &c);
  cbuf &operator= (const cbuf &c);
};

#endif /* !_ASYNC_CBUF_H_ */
