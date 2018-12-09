// -*-c++-*-
/* $Id: str.h 2885 2007-05-18 23:10:26Z max $ */

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


#ifndef _ASYNC_STR_H_
#define _ASYNC_STR_H_ 1

#include "suio++.h"
#include "keyfunc.h"
#include "stllike.h"
#include "refcnt.h"

extern void (*const strobj_opdel) (void *);

class mstr;
class strbuf;

class strobj {
  friend class strobjptr;

  u_int refcnt;

  strobj () : refcnt (0), delfn (strobj_opdel) {}
  void refcount_inc () { refcnt++; }
  void refcount_dec () { if (!--refcnt) delfn (this); }

public:
  size_t len;
  void (*delfn) (void *);
#if __GNUC__ >= 2
  char contents[0];
#endif /* gcc2 */

  char *dat () { return (char *) this + sizeof (*this); }
  static strobj *alloc (size_t n)
    { return new (opnew (n + sizeof (strobj))) strobj; }
};

class strobjptr {
  strobj *p;

public:
  strobjptr () : p (NULL) {}
  strobjptr (strobj *pp) : p (pp) { if (p) p->refcount_inc (); }
  strobjptr (const strobjptr &b) : p (b.p) { if (p) p->refcount_inc (); }
  ~strobjptr () { if (p) p->refcount_dec (); }

  strobjptr &operator= (strobj *pp) {
    if (pp)
      pp->refcount_inc ();
    if (p)
      p->refcount_dec ();
    p = pp;
    return *this;
  }
  strobjptr &operator= (const strobjptr &b) { return *this = b.p; }

  operator bool () const { return p; }
  strobj *operator-> () const { return p; }
  void *Xleak () const { if (p) p->refcount_inc (); return p; }
  static void Xplug (void *p)
    { if (p) static_cast<strobj *> (p)->refcount_dec (); }
};

class str {
  // friend const strbuf &strbuf_cat (const strbuf &, const str &);
  friend void suio_print (suio *, const str &);
  friend class str_init;
  friend class mstr;
  friend class bssstr;

  strobjptr b;

  static strobj *buf2strobj (const char *buf, size_t len) {
    strobj *b = strobj::alloc (1 + len);
    b->len = len;
    memcpy (b->dat (), buf, len);
    b->dat ()[len] = '\0';
    return b;
  }
  strobj *iov2strobj (const iovec *iov, int cnt);
  explicit str (__bss_init) {}
public:
  str () {}
  str (const str &s) : b (s.b) {}
  str (const char *p) { b = p ? buf2strobj (p, strlen (p)) : NULL; }
  str (const strbuf &b);
  explicit str (const suio &u) : b (iov2strobj (u.iov (), u.iovcnt ())) {}
  str (const char *buf, size_t len) : b (buf2strobj (buf, len)) {}
  str (const iovec *iov, int cnt) { setiov (iov, cnt); }
  str (mstr &);

  str &operator= (const str &s) { b = s.b; return *this; }
  str &operator= (const char *p) {
    if (p)
      setbuf (p, strlen (p));
    else
      b = NULL;
    return *this;
  }
  str &operator= (const strbuf &b);
  str &operator= (mstr &m);

  str &setbuf (const char *buf, size_t len) {
    b = buf2strobj (buf, len);
    return *this;
  }
  str &setiov (const iovec *iov, int cnt) {
    b = iov2strobj (iov, cnt);
    return *this;
  }

  size_t len () const { return b->len; }
  const char *cstr () const { return b ? b->dat () : NULL; }
  operator const char *() const { return cstr (); }
  char operator[] (ptrdiff_t n) const {
#ifdef CHECK_BOUNDS
    assert (size_t (n) <= b->len);
#endif /* CHECK_BOUNDS */
    return b->dat ()[n];
  }

  int cmp (const str &s) const {
    if (int r = memcmp (cstr (), s.cstr (), min (len (), s.len ())))
      return r;
    return len () - s.len ();
  }
  int cmp (const char *p) const {
    const char *s = cstr ();
    const char *e = s + len ();
    while (*s == *p)
      if (!*p++)
	return e - s;
      else if (s++ == e)
	return -1;
    return (u_char) *s - (u_char) *p;
  }

  bool operator== (const str &s) const
    { return len () == s.len () && !memcmp (cstr (), s.cstr (), len ()); }
  bool operator!= (const str &s) const
    { return len () != s.len () || memcmp (cstr (), s.cstr (), len ()); }
  bool operator< (const str &s) const { return cmp (s) < 0; }
  bool operator<= (const str &s) const { return cmp (s) <= 0; }
  bool operator> (const str &s) const { return cmp (s) > 0; }
  bool operator>= (const str &s) const { return cmp (s) >= 0; }

  bool operator== (const char *p) const 
  {
    if (!p && !b) return true;
    else if (!p || !b) return false;
    else return !cmp (p);
  }

  bool operator!= (const char *p) const { return !( *this == p); }

  operator hash_t () const { return hash_bytes (cstr (), len ()); }

  static void Xvstomp (const str *sp, void (*dfn) (void *))
    { assert (sp->b->delfn == strobj_opdel); sp->b->delfn = dfn; }
};

struct bssstr : public str {
public:
  bssstr () : str (__bss_init ()) {}
  ~bssstr () { assert (globaldestruction); b.Xleak (); } 
  str &operator= (const str &s) { return str::operator= (s); }
  str &operator= (const bssstr &s) { return str::operator= (s); }
};

inline bool
operator== (const char *p, const str &s)
{
  return s == p;
}
inline bool
operator!= (const char *p, const str &s)
{
  return s != p;
}

template<>
struct compare<str> {
  compare () {}
  int operator () (const str &a, const str &b) const { return a.cmp (b); }
  int operator () (const str &a, const char *b) const { return a.cmp (b); }
  int operator () (const char *a, const str &b) const { return -b.cmp (a); }
  int operator () (const char *a, const char *b) const
    { return strcmp (a, b); }
};

template<>
struct equals<str> {
  equals () {}
  int operator () (const str &a, const str &b) const { return a == b; }
  int operator () (const str &a, const char *b) const { return a == b; }
  int operator () (const char *a, const str &b) const { return a == b; }
  int operator () (const char *a, const char *b) const
    { return !strcmp (a, b); }
};

template<>
struct hashfn<str> {
  hashfn () {}
  hash_t operator () (const str &s) const { return s; }
  hash_t operator () (const char *p) const { return hash_string (p); }
};

class mstr {
  friend class str;

  mstr (const mstr &);
  mstr &operator= (const mstr &);

protected:
  strobjptr b;

  mstr () {}

public:
  explicit mstr (size_t n) : b (strobj::alloc (n + 1)) { b->len = n; }

  void setlen (size_t n) { assert (n <= b->len); b->len = n; }

  size_t len () const { return b->len; }
  const char *cstr () const { return b->dat (); }
  operator const char *() const { return cstr (); }
  char *cstr () { return b->dat (); }
  operator char *() { return cstr (); }
};

inline
str::str (mstr &m)
  : b (m.b)
{
  b->dat ()[b->len] = '\0';
  m.b = NULL;			// Destroy mutable string
}

inline str &
str::operator= (mstr &m)
{
  b = m.b;
  b->dat ()[b->len] = '\0';
  m.b = NULL;			// Destroy mutable string
  return *this;
}

void suio_print (suio *uio, const str &s);
str suio_getline (suio *uio);

class strbuf {
  friend inline const strbuf &strbuf_cat (const strbuf &, const strbuf &);

  strbuf &operator= (const strbuf &);

protected:
  const ref<suio> uio;

public:
  strbuf () : uio (New refcounted<suio>) {}
  strbuf (const str &s) : uio (New refcounted<suio>) { suio_print (uio, s); }
  strbuf (const strbuf &b) : uio (b.uio) {}
  explicit strbuf (const ref<suio> &u) : uio (u) {}
  explicit strbuf (const char *fmt, ...)
    __attribute__ ((format (printf, 2, 3)));

  const strbuf &fmt (const char *, ...) const
    __attribute__ ((format (printf, 2, 3)));
  const strbuf &vfmt (const char *fmt, va_list ap) const {
    suio_vuprintf (uio, fmt, ap);
    return *this;
  }
  const strbuf &buf (const char *buf, size_t len) const {
    suio_printcheck (uio, buf, len);
    return *this;
  }

  /* gcc bug requires explicit declarations for const char * */
  const strbuf &cat (const char *p, bool copy = true) const {
    if (copy)
      uio->copy (p, strlen (p));
    else
      suio_printcheck (uio, p, strlen (p));
    return *this;
  }
  // const strbuf &operator<< (const char *p) const { return cat (p); }

  template<class A1> const strbuf &cat (const A1 &a1) const
    { return strbuf_cat (*this, a1); }
  template<class A1, class A2>
  const strbuf &cat (const A1 &a1, const A2 &a2) const
    { return strbuf_cat (*this, a1, a2); }

  const iovec *iov () const { return uio->iov (); }
  size_t iovcnt () const { return uio->iovcnt (); }
  suio *tosuio () const { return uio; }
};

inline
str::str (const strbuf &sb)
  : b (iov2strobj (sb.iov (), sb.iovcnt ()))
{
}

inline str &
str::operator= (const strbuf &sb)
{
  setiov (sb.iov (), sb.iovcnt ());
  return *this;
}

inline const strbuf &
strbuf_cat (const strbuf &b, const strbuf &b2)
{
  b.uio->copyu (b2.uio);
  return b;
}

inline const strbuf &
strbuf_cat (const strbuf &b, const str &s)
{
  suio *uio = b.tosuio ();
  suio_print (uio, s);
  return b;
}

const strbuf &strbuf_cat (const strbuf &b, const char *p, bool copy = true);

#define STRBUFOP(arg, op)			\
inline const strbuf &				\
strbuf_cat (const strbuf &b, arg)		\
{						\
  op;						\
  return b;					\
}

STRBUFOP (int n, b.fmt ("%d", n))
STRBUFOP (u_int n, b.fmt ("%u", n))
STRBUFOP (long n, b.fmt ("%ld", n))
STRBUFOP (u_long n, b.fmt ("%lu", n))
#if SIZEOF_LONG_LONG > 0
STRBUFOP (long long n, b.fmt ("%qd", n))
STRBUFOP (unsigned long long n, b.fmt ("%qu", n))
#endif /* SIZEOF_LONG_LONG > 0 */

#undef STRBUFOP

template<class A, class B = void> class strbufcatobj {
  const A &a;
  const B &b;
public:
  const strbuf &cat (const strbuf &sb) const { return sb.cat (a, b); }
  strbufcatobj (const A &aa, const B &bb) : a (aa), b (bb) {}
};
template<class A> class strbufcatobj<A> {
  const A &a;
public:
  const strbuf &cat (const strbuf &sb) const { return sb.cat (a); }
  strbufcatobj (const A &aa) : a (aa) {}
};

template<class A, class B> const strbuf &
strbuf_cat (const strbuf &sb, const strbufcatobj<A, B> &o)
{
  return o.cat (sb);
}

template<class A, class B> inline strbufcatobj<A, B>
cat (const A &a, const B &b)
{
  return strbufcatobj<A, B> (a, b);
}
template<class A> inline strbufcatobj<A>
cat (const A &a)
{
  return strbufcatobj<A> (a);
}

class hexdump {
  friend const strbuf &strbuf_cat (const strbuf &, const hexdump &);
  const void *const buf;
  const size_t len;
public:
  hexdump (const void *b, size_t l) : buf (b), len (l) {}
};

inline const strbuf &
strbuf_cat (const strbuf &sb, const hexdump &hd)
{
  const u_char *p = static_cast<const u_char *> (hd.buf);
  const u_char *end = p + hd.len;
  while (p < end)
    sb.fmt ("%02x", *p++);
  return sb;
}

template<class T> inline const strbuf &
operator<< (const strbuf &sb, const T &a)
{
  return strbuf_cat (sb, a);
}
inline const strbuf &
operator<< (const strbuf &sb, const str &s)
{
  suio_print (sb.tosuio (), s);
  return sb;
}
#if 0
/* XXX - work around g++ bug */
inline const strbuf &
operator<< (const strbuf &sb, const char *a)
{
  return strbuf_cat (sb, a);
}
#endif
/* XXX - compilation time goes through the roof when this is inlined */
/* XXX - work around g++ bug */
const strbuf &operator<< (const strbuf &sb, const char *a);

#if 0
template <class T> inline const strbuf
operator<< (const str &s, const T &t)
{
  return strbuf (s).cat (t);
}
#endif
/* XXX - work around g++ bug */
const strbuf operator<< (const str &s, const char *a);
inline const strbuf
operator<< (const str &s1, const str &s2)
{
  return strbuf (s1).cat (s2);
}

inline str
substr (const str &s, size_t pos, size_t len = (size_t) -1)
{
  if (pos >= s.len ())
    return "";
  if (len > s.len () - pos)
    len = s.len () - pos;
  return str (s.cstr () + pos, len);
}

#endif /* !_ASYNC_STR_H_ */
