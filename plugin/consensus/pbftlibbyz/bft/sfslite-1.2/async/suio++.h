// -*-c++-*-
/* $Id: suio++.h 1117 2005-11-01 16:20:39Z max $ */

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

#ifndef _ASYNC_SUIOXX_H_
#define _ASYNC_SUIOXX_H_ 1

#include "opnew.h"
#include "vec.h"
#include "callback.h"

class iovmgr {
  const iovec *iov;
  const iovec *lim;
  iovec cur;

public:
  iovmgr (const iovec *iov, int iovcnt);
  void skip (size_t);
  size_t copyout (char *, size_t);
  size_t size () const;
};

size_t iovsize (const iovec *, int);

class suio {
public:
  enum { smallbufsize = 0x80 };
  enum { blocksize = 0x2000 };

private:
  typedef callback<void>::ref cb_t;

  class uiocb {
    uiocb &operator= (const uiocb &);
  public:
    cb_t cb;
    u_int64_t nbytes;
    uiocb (u_int64_t nb, cb_t cb) : cb (cb), nbytes (nb) {}
  };

  vec<iovec, 4> iovs;
  vec<uiocb, 2> uiocbs;

  size_t uiobytes;
  u_int64_t nrembytes;
  u_int64_t nremiov;

  char *lastiovend;

  char *scratch_buf;
  char *scratch_pos;
  char *scratch_lim;

  suio (const suio &);
  suio &operator= (const suio &);

protected:
  void *(*allocator) (size_t);
  void (*deallocator) (void *, size_t);

  char defbuf[smallbufsize];

  static void *default_allocator (size_t n) { return txmalloc (n); }
  static void default_deallocator (void *p, size_t) { xfree (p); }

  void makeuiocbs ();
  char *morescratch (size_t);
  void pushiov (const void *base, size_t len);
  void slowcopy (const void *, size_t);
  void slowfill (char c, size_t n);

public:
  suio ();
  ~suio ();
  void clear ();

  char *getspace (size_t n);
  char *getspace_aligned (size_t n);
  void fill (char c, ssize_t n);
  void copy (const void *, size_t);
  void copyv (const iovec *iov, size_t cnt, size_t skip = 0);
  void copyu (const suio *uio) { copyv (uio->iov (), uio->iovcnt ()); }
  void print (const void *, size_t);
  void take (suio *src);
  void rembytes (size_t n);

  void iovcb (cb_t cb) {
    if (uiobytes)
      uiocbs.push_back (uiocb (nrembytes + uiobytes, cb));
    else
      (*cb) ();
  }
  void breakiov () { lastiovend = NULL; }

  const iovec *iov () const { return iovs.base (); }
  const iovec *iovlim () const { return iovs.lim (); }
  size_t iovcnt () const { return iovs.size (); }
  size_t resid () const { return uiobytes; }
  u_int64_t iovno () const { return nremiov; }
  u_int64_t byteno () const { return nrembytes; }

  int (output) (int fd, int cnt = -1);
  size_t copyout (void *buf, size_t len) const;
  size_t copyout (void *buf) const { return copyout (buf, (size_t) -1); }

  size_t fastspace () const { return scratch_lim - scratch_pos; }
  int (input) (int fd, size_t len = blocksize);
  size_t linelen () const;
};

inline void
suio::pushiov (const void *_base, size_t len)
{
  void *base = const_cast<void *> (_base);
  if (base == lastiovend) {
    lastiovend += len;
    iovs.back ().iov_len += len;
  }
  else if (len) {
    iovec *iov = &iovs.push_back ();
    iov->iov_base = static_cast<iovbase_t> (base);
    iov->iov_len = len;
    lastiovend = static_cast<char *> (base) + len;
  }
  uiobytes += len;
  if (base == scratch_pos) {
    scratch_pos += len;
#ifdef CHECK_BOUNDS
    assert (scratch_pos <= scratch_lim);
#endif /* CHECK_BOUNDS */
  }
}

inline char *
suio::getspace (size_t n)
{
  if (n <= (size_t) (scratch_lim - scratch_pos))
    return scratch_pos;
  return morescratch (n);
}

inline char *
suio::getspace_aligned (size_t n)
{
  scratch_pos += - reinterpret_cast<size_t> (scratch_pos) & 3;
  return getspace (n);
}

inline void
suio::fill (char c, ssize_t n)
{
  if (n <= 0)
    return;
  if (n <= scratch_lim - scratch_pos) {
    memset (scratch_pos, c, n);
    pushiov (scratch_pos, n);
  }
  else
    slowfill (c, n);
}

inline void
suio::copy (const void *buf, size_t len)
{
  if (len <= (size_t) (scratch_lim - scratch_pos)) {
    memcpy (scratch_pos, buf, len);
    pushiov (scratch_pos, len);
  }
  else
    slowcopy (buf, len);
}

inline size_t
suio::linelen () const
{
  size_t n = 0;
  for (const iovec *v = iov (), *e = iovlim ();
       v < e; n += v++->iov_len)
    if (void *p = memchr ((char *) v->iov_base, '\n', v->iov_len))
      return n + (static_cast<char *> (p)
		  - static_cast<char *> (v->iov_base)) + 1;
  return 0;
}

size_t iovsize (const iovec *, int);
inline void
iovscrub (const iovec *iov, int cnt)
{
  const iovec *end = iov + cnt;
  while (iov < end)
    bzero (iov->iov_base, iov->iov_len);
}

/* Suio_printcheck for debugging */

#ifdef DMALLOC
void __suio_printcheck (const char *, suio *, const void *, size_t);
void __suio_check (const char *, suio *, const void *, size_t);
#else /* !DMALLOC */
#define __suio_printcheck(line, uio, buf, len) (uio)->print (buf, len)
#endif /* !DMALLOC */

#define suio_printcheck(uio, buf, len) \
  __suio_printcheck (__FL__, uio, buf, len)


inline void
suio::print (const void *buf, size_t len)
{
#ifdef DMALLOC
  if (buf != scratch_pos)
    __suio_check ("suio::print", this, buf, len);
#endif /* DMALLOC */
  if (len <= smallbufsize && buf != scratch_pos)
    copy (buf, len);
  else
    pushiov (buf, len);
}

/* Printf */

#ifndef DSPRINTF_DEBUG
/* The uprintf functions build up uio structures based on format
 * strings.  String ("%s") arguments are NOT copied, so you must not
 * modify any strings passed in.  Also, a format string of '%m'
 * doesn't convert any arguments but is equivalent to '%s' with an
 * argument of strerror (errno).  (This is like syslog.)  */
extern void suio_vuprintf (struct suio *, const char *, va_list);
extern void suio_uprintf (struct suio *, const char *, ...)
     __attribute__ ((format (printf, 2, 3)));
#else /* DSPRINTF_DEBUG */
extern void __suio_vuprintf (const char *, struct suio *,
			     const char *, va_list);
#define suio_vuprintf(uio, fmt, ap) __suio_vuprintf (__FL__, uio, fmt, ap)
extern void __suio_uprintf (const char *, struct suio *, const char *, ...)
  __attribute__ ((format (printf, 3, 4)));
#define suio_uprintf(uio, args...) __suio_uprintf (__FL__, uio, args)
#endif /* DSPRINTF_DEBUG */

/* Compatibility */

#ifdef DMALLOC
char *__suio_flatten (const struct suio *, const char *, int);
#define suio_flatten(uio) __suio_flatten (uio, __FILE__, __LINE__)
#else /* !DMALLOC */
char *suio_flatten (const struct suio *);
#endif /* !DMALLOC */

#define suio_fill(uio, c, len) (uio)->fill (c, len)
#define suio_copy(uio, buf, len) (uio)->copy (buf, len)
#define suio_callback(uio, cb) (uio)->iovcb (cb)

inline void
suio_print (suio *uio, const void *buf, size_t len)
{
  uio->print (buf, len);
}

#endif /* !_ASYNC_SUIOXX_H_ */
