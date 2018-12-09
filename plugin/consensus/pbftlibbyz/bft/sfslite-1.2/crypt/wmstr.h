// -*-c++-*-
/* $Id: wmstr.h 1117 2005-11-01 16:20:39Z max $ */

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

#ifndef _WMSTR_H_
#define _WMSTR_H_ 1

#include "str.h"

void wstrobj_delfn (void *);

inline const str &
str2wstr (const str &s)
{
  if (s)
    str::Xvstomp (&s, wstrobj_delfn);
  return s;
}

inline strobj *
wstrobj_alloc (size_t n)
{
  strobj *b = strobj::alloc (n);
  b->delfn = wstrobj_delfn;
  return b;
}

class wmstr : public mstr {
  void setiov (const iovec *iov, u_int cnt) {
    size_t n = iovsize (iov, cnt);
    b = wstrobj_alloc (n + 1);
    b->len = n;
    char *p = b->dat ();
    for (u_int i = 0; i < cnt; i++) {
      memcpy (p, iov[i].iov_base, iov[i].iov_len);
      p += iov[i].iov_len;
    }
  }

public:
  explicit wmstr (size_t n) {
    b = wstrobj_alloc (n + 1);
    b->len = n;
  }
  wmstr (const iovec *iov, u_int cnt) { setiov (iov, cnt); }
  wmstr (const suio *uio) { setiov (uio->iov (), uio->iovcnt ()); }
  wmstr (const strbuf &sb) { setiov (sb.iov (), sb.iovcnt ()); }
};

inline str
wstr (const void *buf, size_t len)
{
  wmstr m (len);
  memcpy (m, buf, len);
  return m;
}


template<class T> class zeroed_tmp_buf {
  zeroed_tmp_buf (const zeroed_tmp_buf &);
  zeroed_tmp_buf &operator= (const zeroed_tmp_buf &);

public:
  T *const base;
  const size_t size;

  explicit zeroed_tmp_buf (size_t n) : base (New T[n]), size (n) {}
  ~zeroed_tmp_buf () { bzero (base, size * sizeof (T)); delete[] base; }

  operator T *() const { return base; }
  T &operator[] (ptrdiff_t n) const {
#ifdef CHECK_BOUNDS
    assert (size_t (n) < size);
#endif /* CHECK_BOUNDS */
    return base[size_t (n)];
  }
};

typedef zeroed_tmp_buf<char> zcbuf;
typedef zeroed_tmp_buf<u_char> zucbuf;

#endif /* !_WMSTR_H_ */
