// -*-c++-*-
/* $Id: rxx.h 3467 2008-07-01 03:49:10Z max $ */

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

#ifndef _RXX_H_
#define _RXX_H_ 1

#include "sysconf.h"
extern "C" {
#include "pcre.h"
}
#include "str.h"
#include "init.h"

INIT (rxxinit);

extern "C" void *rcmalloc (size_t n);
extern "C" void rcfree (void *p);
extern "C" void *rccopy (void *p);

extern bool sfs_rxx_panic;

class rxx {
protected:
  pcre *re;
  pcre_extra *extra;

  int nsubpat;
  int *ovector;
  int ovecsize;
  str subj;
  int _errcode;

  str init (const char *pat, const char *opt);
  void copy (const rxx &r) {
    re = static_cast<pcre *> (rccopy (r.re));
    extra = static_cast<pcre_extra *> (rccopy (r.extra));
    nsubpat = 0;
    ovector = NULL;
    ovecsize = r.ovecsize;
  }
  rxx &operator= (const rxx &);
  void mknull () {
    re = NULL;
    extra = NULL;
    nsubpat = 0;
    ovector = NULL;
    ovecsize = NULL;
    subj = NULL;
  }
  rxx () {}
  void freemem () { rcfree (re); rcfree (extra); delete[] ovector; }


public:
  bool _exec (const char *p, size_t len, int options);
  bool exec (str s, int options);

  class matchresult {
    const rxx &r;
  public:
    matchresult (const rxx &r) : r (r) {}
    operator bool () const { return r.success (); }
    operator str () const { return r.at (0); }
    str operator[] (ptrdiff_t n) const { return r.at (n); }
  };

  rxx (const char *pat, const char *opt = "")
    { if (str s = init (pat, opt)) panic ("%s", s.cstr ()); }
  rxx (const rxx &r) { assert (r.re); copy (r); }
  ~rxx () { freemem (); }

  str study ()
    { const char *err; extra = pcre_study (re, 0, &err); return err; }
  void clear () { nsubpat = 0; subj = NULL; }

  matchresult search (str s, int opt = 0) { exec (s, opt); return *this; }
  matchresult match (str s, int opt = 0) {
    if (exec (s, opt | PCRE_ANCHORED)) {
      if (nsubpat > 0 && implicit_cast<size_t> (ovector[1]) != s.len ())
	nsubpat = 0;
    }
    return *this;
  }

  bool success () const { return nsubpat > 0; }
  int errcode () const { return _errcode; }

  str at (ptrdiff_t n) const;
  int start (int n) const
    { assert (n >= 0); return n < nsubpat ? ovector[2*n] : -1; }
  int end (int n) const
    { assert (n >= 0); return n < nsubpat ? ovector[2*n+1] : -1; }
  int len (int n) const {
    assert (n >= 0);
    size_t i = 2 * n;
    return n < nsubpat && ovector[i] >= 0
      ? ovector[i+1] - ovector[i] : -1;
  }

  str operator[] (ptrdiff_t n) const { return at (n); }
};

class rrxx : public rxx {
  str err;
  void mknull () { err = "uninitialized"; rxx::mknull (); }
public:
  rrxx () { mknull (); }
  explicit rrxx (const char *pat, const char *opt = "")
    { err = init (pat, opt); }
  rrxx (const rxx &r) { err = NULL; copy (r); }
  rrxx (const rrxx &r) { err = r.err; copy (r); }
  bool compile (const char *pat, const char *opt = "")
    { freemem (); mknull (); err = init (pat, opt); return !err; }
  const str &geterr () const { return err; }
};

inline rxx::matchresult
operator/ (const str &s, rxx &r)
{
  return r.search (s);
}

class strstrmatch {
  const str &s;
  const char *const p;
protected:
  mutable const char *o;
public:
  strstrmatch (const str &s, const char *p)
    : s (s), p (p), o ("") {}
  operator bool () const
    { rxx r (p, o); return r.search (s); }
  operator str () const
    { rxx r (p, o); return r.search (s)[0]; }
  str operator[] (ptrdiff_t n) const
    { rxx r (p, o); return r.search (s)[n]; }
};
class strstroptmatch : public strstrmatch {
public:
  strstroptmatch (const str &s, const char *p)
    : strstrmatch (s, p) {}
  const strstrmatch &operator/ (const char *opt) const
    { o = opt; return *this; }
};

inline strstroptmatch
operator/ (const str &s, const char *p)
{
  return strstroptmatch (s, p);
}

int split (vec<str> *out, rxx pat, str expr,
	   size_t lim = (size_t) -1, bool emptylast = false);
str join (str sep, const vec<str> &v);

#endif /* !_RXX_H_ */
