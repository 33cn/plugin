/* $Id: str.C 1117 2005-11-01 16:20:39Z max $ */

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

#include "str.h"

void (*const strobj_opdel) (void *) (operator delete);

void
suio_print (suio *uio, const str &s)
{
  if (s.len () <= suio::smallbufsize)
    uio->copy (s.cstr (), s.len ());
  else {
    uio->print (s.cstr (), s.len ());
    uio->iovcb (wrap (&s.b.Xplug, s.b.Xleak ()));
  }
}

const strbuf &
strbuf_cat (const strbuf &b, const char *p, bool copy)
{
  suio *uio = b.tosuio ();
  if (copy)
    uio->copy (p, strlen (p));
  else
    suio_printcheck (uio, p, strlen (p));
  return b;
}

const strbuf &
operator<< (const strbuf &sb, const char *a)
{
  return strbuf_cat (sb, a);
}

const strbuf
operator<< (const str &s, const char *a)
{
  return strbuf (s).cat (a);
}

str
suio_getline (suio *uio)
{
  if (size_t n = uio->linelen ()) {
    mstr m (n - 1);
    uio->copyout (m, n - 1);
    uio->rembytes (n);
    if (m.len () && m.cstr ()[m.len () - 1] == '\r')
      m.setlen (m.len () -1 );
    return m;
  }
  return NULL;
}

strobj *
str::iov2strobj (const iovec *iov, int cnt)
{
  size_t l = iovsize (iov, cnt);
  strobj *b = strobj::alloc (1 + l);
  b->len = l;
  char *p = b->dat ();
  for (const iovec *end = iov + cnt; iov < end; iov++) {
    memcpy (p, iov->iov_base, iov->iov_len);
    p += iov->iov_len;
  }
  *p = '\0';
  assert (p == b->dat () + l);
  return b;
}

strbuf::strbuf (const char *fmt, ...)
  : uio (New refcounted<suio>)
{
  va_list ap;
  va_start (ap, fmt);
  suio_vuprintf (uio, fmt, ap);
  va_end (ap);
}

const strbuf &
strbuf::fmt (const char *fmt, ...) const
{
  va_list ap;
  va_start (ap, fmt);
  suio_vuprintf (uio, fmt, ap);
  va_end (ap);
  return *this;
}

/* For use from the debugger */
const char *
cstr (const str s)
{
  return s;
}

const char *
cstrp (const str *s)
{
  return *s;
}
