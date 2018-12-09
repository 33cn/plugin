// -*-c++-*-
/* $Id: aios.h 1117 2005-11-01 16:20:39Z max $ */

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

#ifndef _ASYNC_AIOS_H_
#define _ASYNC_AIOS_H_ 1

#include "str.h"
#include "union.h"
#include "init.h"
#include "cbuf.h"

struct timecb_t;

INIT(aiosinit);

class aios : virtual public refcount {
  friend class aiosout;

  typedef callback<void, str, int>::ptr rcb_t;
  typedef callback<void, int>::ptr wcb_t;

  bool rlock;
  bool (aios::*infn) ();
  rcb_t rcb;

  bool wblock;

  time_t timeoutval;
  time_t timeoutnext;
  timecb_t *timeoutcb;

  ssize_t debugiov;

  vec<int> fdsendq;

  void mkrcb (const str &s)
    { infn = &aios::rnone; rcb_t::ref cb = rcb; rcb = NULL; (*cb) (s, err); }
  void timeoutcatch ();
  void timeoutbump ();

  bool rnone () { return false; }
  bool rline ();
  bool rany ();
  void setreadcb (bool (aios::*fn) (), rcb_t cb);

  virtual void mkwcb (wcb_t cb) { if (fd >= 0) (*cb) (err); }
  void schedwrite ();

  void dumpdebug ();
  void outstart () {
    assert (!weof);
    if (debugname) {
      outb.tosuio ()->breakiov ();
      debugiov = outb.tosuio ()->iovcnt ();
    }
  }
  void outstop () {
    if (debugname)
      dumpdebug ();
    debugiov = -1;
    schedwrite ();
  }

protected:
  int fd;
  cbuf inb;
  strbuf outb;

  int err;
  bool eof;
  bool weof;

  str debugname;
  const char *wrpref;
  const char *rdpref;
  const char *errpref;

  aios (int, size_t);
  ~aios ();
  void finalize ();

  bool reading () { return rcb; }
  virtual bool writing () { return outb.tosuio ()->resid (); }
  void fail (int e);
  void input ();
  void (output) ();

  virtual int doinput ();
  virtual int dooutput ();
  virtual void setincb ();
  virtual void setoutcb ();

public:
  enum { defrbufsize = 0x2000 };
  static ref<aios> alloc (int fd, size_t rbsz = defrbufsize)
    { return New refcounted<aios> (fd, rbsz); }
  int fdno () { return fd; }
  void setdebug (str name) { debugname = name; }
  void settimeout (time_t secs) { timeoutval = secs; timeoutbump (); }
  time_t gettimeout () { return timeoutval; }
  void abort ();

  void setrbufsize (size_t n) { inb.resize (n); }
  size_t getrbufbytes () { return inb.bytes (); }
  void readline (rcb_t cb) { setreadcb (&aios::rline, cb); }
  void readany (rcb_t cb) { setreadcb (&aios::rany, cb); }
  void readcancel () { infn = &aios::rnone; rcb = NULL; }
  void unread (size_t n) { inb.unrembytes (n); }

  void writev (const iovec *iov, int iovcnt);
  void write (void *buf, size_t len)
    { iovec iov = { iovbase_t (buf), len }; writev (&iov, 1); }
  void sendeof ();
  virtual void setwcb (wcb_t cb)
    { suio_callback (outb.tosuio (), wrap (this, &aios::mkwcb, cb)); }
  int flush ();
  void sendfd (int sfd) { fdsendq.push_back (sfd); }
};
typedef ref<aios> aios_t;

class aiosout : public strbuf {
  aios *s;

  // aiosout (aiosout &o) : strbuf (o), s (o.s) { o.s = NULL; }
  aiosout &operator= (const aiosout &);
public:
  /* XXX - We intentionally make the copy constructor public and
   * undefined, because aiosout objects cannot be copied.  No
   * reasonable compiler should copy an aiosout during reference copy
   * initialization (e.g., ``aout << "hello world\n";'').  However,
   * section 8.5.3/5 of the C++ standard implies the compiler is
   * allowed to make copies of an aios object during reference copy
   * initialization, because it can make and initialize a temporary
   * aiosout using copy (not direct) initialization from the const
   * aios_t &.  Section 12.1/1 therefore requires that the copy
   * constructor be accessible, so we can't make it private either. */
  aiosout (const aiosout &o);

  aiosout (const aios_t &a) : strbuf (a->outb), s(a) { s->outstart (); }
  aiosout (const aios_t::ptr &a) : strbuf (a->outb), s(a) { s->outstart (); }
  ~aiosout () { s->outstop (); }
};

template<class T> inline const strbuf &
operator<< (const aiosout &o, const T &a)
{
  return strbuf_cat (o, a);
}
inline const strbuf &
operator<< (const aiosout &o, const str &s)
{
  suio_print (o.tosuio (), s);
  return o;
}
// XXX - gcc bug requires this:
inline const strbuf &
operator<< (const aiosout &o, const char *a)
{
  return strbuf_cat (o, a);
}

// extern const str endl;
#ifndef __AIOS_IMPLEMENTATION
extern aios_t ain, aout;
#endif /* !__AIOS_IMPLEMENTATION */

#endif /* !_ASYNC_AIOS_H_ */
