/* $Id: cbuf.C 1117 2005-11-01 16:20:39Z max $ */

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

#include "cbuf.h"
#include "stllike.h"

void
cbuf::resize (size_t n)
{
  int nend = size ();
  char *nbuf = (char *) xmalloc (n);
  copyout (nbuf, nend);
  xfree (buf);
  buf = nbuf;
  buflen = n;
  empty = nend;
  start = 0;
  end = nend;
}

const iovec *
cbuf::iniov ()
{
  inv[0].iov_base = buf + end;
  if (empty || start < end) {
    inv[0].iov_len = buflen - end;
    inv[1].iov_base = buf;
    inv[1].iov_len = start;
  }
  else {
    inv[0].iov_len = start - end;
    inv[1].iov_base = NULL;
    inv[1].iov_len = 0;
  }
  return inv;
}

void
cbuf::addbytes (size_t n)
{
  if (n) {
    assert (n <= space ());
    empty = false;
    end += n;
    if (end >= buflen)
      end -= buflen;
  }
}

int
cbuf::find (char c)
{
  if (empty)
    return -1;
  if (start < end) {
    if (char *p = (char *) memchr (buf + start, c, end - start))
      return p - (buf + start);
    return -1;
  }
  if (char *p = (char *) memchr (buf + start, c, buflen - start))
    return p - (buf + start);
  if (char *p = (char *) memchr (buf, c, end))
    return p - buf + buflen - start;
  return -1;
};

const iovec *
cbuf::outiov ()
{
  outv[0].iov_base = buf + start;
  if (start >= end && !empty) {
    outv[0].iov_len = buflen - start;
    outv[1].iov_base = buf;
    outv[1].iov_len = end;
  }
  else {
    outv[0].iov_len = end - start;
    outv[1].iov_base = NULL;
    outv[1].iov_len = 0;
  }
  return outv;
}

void
cbuf::copyout (void *_dst, size_t len)
{
  char *dst = static_cast<char *> (_dst);
  assert (len <= size ());
  if (empty || start < end)
    memcpy (dst, buf + start, min (len, end - start));
  else {
    size_t n = min (len, buflen - start);
    memcpy (dst, buf + start, n);
    if (len > n)
      memcpy (dst + n, buf, len - n);
  }
  rembytes (len);
}
