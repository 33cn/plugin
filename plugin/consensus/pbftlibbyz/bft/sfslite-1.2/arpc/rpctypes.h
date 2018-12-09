// -*-c++-*-
/* $Id: rpctypes.h 3713 2008-10-10 06:31:41Z max $ */

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

#ifndef _RPCTYPES_H_
#define _RPCTYPES_H_ 1

#include "str.h"
#include "vec.h"
#include "array.h"
#include "union.h"
#include "keyfunc.h"
#include "err.h"
#include "qhash.h"

struct rpcgen_table {
  const char *name;

  const std::type_info *type_arg;
  void *(*alloc_arg) ();
  sfs::xdrproc_t xdr_arg;
  void (*print_arg) (const void *, const strbuf *, int,
		     const char *, const char *);

  const std::type_info *type_res;
  void *(*alloc_res) ();
  sfs::xdrproc_t xdr_res;
  void (*print_res) (const void *, const strbuf *, int,
		     const char *, const char *);
};

struct rpc_program {
  u_int32_t progno;
  u_int32_t versno;
  const struct rpcgen_table *tbl;
  size_t nproc;
  const char *name;
  bool lookup (const char *rpc, u_int32_t *out) const;
};

enum { RPC_INFINITY = 0x7fffffff };

template<class T> class rpc_ptr {
  T *p;

public:
  rpc_ptr () { p = NULL; }
  rpc_ptr (const rpc_ptr &rp) { p = rp ? New T (*rp) : NULL; }
  ~rpc_ptr () { delete p; }

  void clear () { delete p; p = NULL; }
  rpc_ptr &alloc () { if (!p) p = New T; return *this; }
  rpc_ptr &assign (T *tp) { clear (); p = tp; return *this; }
  T *release () { T *r = p; p = NULL; return r; }

  operator T *() const { return p; }
  T *operator-> () const { return p; }
  T &operator* () const { return *p; }

  rpc_ptr &operator= (const rpc_ptr &rp) {
    if (!rp.p)
      clear ();
    else if (p)
      *p = *rp.p;
    else
      p = New T (*rp.p);
    return *this;
  }
  void swap (rpc_ptr &a) { T *ap = a.p; a.p = p; p = ap; }
};

template<class T> inline void
swap (rpc_ptr<T> &a, rpc_ptr<T> &b)
{
  a.swap (b);
}


template<class T, size_t max> class rpc_vec : private vec<T> {
  typedef vec<T> super;
public:
  typedef typename super::elm_t elm_t;
  typedef typename super::base_t base_t;
  using super::basep;
  using super::firstp;
  using super::lastp;
  using super::limp;
  enum { maxsize = max };

protected:
  bool nofree;

  void init () { nofree = false; super::init (); }
  void del () { if (!nofree) super::del (); }

  void copy (const elm_t *p, size_t n) {
    clear ();
    reserve (n);
    const elm_t *e = p + n;
    while (p < e)
      push_back (*p++);
  }
  template<size_t m> void copy (const rpc_vec<T, m> &v) {
    assert (v.size () <= maxsize);
#if 0
    if (v.nofree) {
      del ();
      nofree = true;
      basep = limp = NULL;
      firstp = v.firstp;
      lastp = v.lastp;
      return;
    }
#endif
    if (nofree)
      init ();
    super::operator= (v);
  }
  void ensure (size_t m) { assert (!nofree); assert (size () + m <= maxsize); }

public:
  rpc_vec () { init (); }
  rpc_vec (const rpc_vec &v) { init (); copy (v); }
  template<size_t m> rpc_vec (const rpc_vec<T, m> &v) { init (); copy (v); }
  ~rpc_vec () { if (nofree) super::init (); }
  void clear () { del (); init (); }

  rpc_vec &operator= (const rpc_vec &v) { copy (v); return *this; }
  template<size_t m> rpc_vec &operator= (const rpc_vec<T, m> &v)
    { copy (v); return *this; }
  template<size_t m> rpc_vec &operator= (const ::vec<T, m> &v)
    { copy (v.base (), v.size ()); return *this; }
  template<size_t m> rpc_vec &operator= (const array<T, m> &v)
    { switch (0) case 0: case m <= max:; copy (v.base (), m); return *this; }

  void swap (rpc_vec &v) {
    bool nf = v.nofree;
    v.nofree = nofree;
    nofree = nf;
    base_t::swap (v);
  }

  rpc_vec &set (elm_t *base, size_t len) {
    assert (len <= maxsize);
    del ();
    nofree = true;
    basep = limp = NULL;
    firstp = base;
    lastp = base + len;
    return *this;
  }
  template<size_t m> rpc_vec &set (const ::vec<T, m> &v)
    { set (v.base (), v.size ()); return *this; }

  void reserve (size_t m) { ensure (m); super::reserve (m); }

  void setsize (size_t n) {
    assert (!nofree);
    assert (n <= max);
    super::setsize (n);
  }

  size_t size () const { return super::size (); }
  bool empty () const { return super::empty (); }

  elm_t *base () { return super::base (); }
  const elm_t *base () const { return super::base (); }

  elm_t *lim () { return super::lim (); }
  const elm_t *lim () const { return super::lim (); }

  elm_t &operator[] (size_t i) { return super::operator[] (i); }
  const elm_t &operator[] (size_t i) const { return super::operator[] (i); }
  elm_t &at (size_t i) { return (*this)[i]; }
  const elm_t &at (size_t i) const { return (*this)[i]; }

  elm_t &push_back () {
    ensure (1);
    return super::push_back ();
  }
  elm_t &push_back (const elm_t &e) {
    ensure (1);
    return super::push_back (e);
  }
  elm_t pop_back () {
    if (nofree) {
      assert (lastp > firstp);
      return *--lastp;
    }
    else
      return super::pop_back ();
  }
  elm_t pop_front () {
    if (nofree) {
      assert (lastp > firstp);
      return *firstp++;
    }
    else
      return super::pop_front ();
  }

  elm_t &front () { return super::front (); }
  const elm_t &front () const { return super::front (); }
  elm_t &back () { return super::back (); }
  const elm_t &back () const { return super::back (); }
};

template<class T, size_t max> void
swap (rpc_vec<T, max> &a, rpc_vec<T, max> &b)
{
  a.swap (b);
}

extern const str rpc_emptystr;
template<size_t max = RPC_INFINITY> struct rpc_str : str
{
  enum { maxsize = max };

private:
  void check () {
    assert (len () == strlen (cstr ()));
    assert (len () <= maxsize);
  }

public:
  rpc_str () : str (rpc_emptystr) {}
  rpc_str (const rpc_str &s) : str (s) {}
  rpc_str (const str &s) : str (s) { check (); }
  rpc_str (const char *p) : str (p) { assert (len () <= maxsize); }
  rpc_str (const strbuf &b) : str (b) { check (); }
  rpc_str (const char *buf, size_t len) : str (buf, len) { check (); }
  rpc_str (const iovec *iov, int cnt) : str (iov, cnt) { check (); }
  rpc_str (mstr &m) : str (m) { check (); }

  rpc_str &operator= (const rpc_str &s)
    { str::operator= (s); return *this; }
  rpc_str &operator= (const char *p)
    { str::operator= (p); if (p) assert (len () <= maxsize); return *this; }
  template<class T> rpc_str &operator= (const T &t)
    { str::operator= (t); check (); return *this; }
  rpc_str &operator= (mstr &m)
    { str::operator= (m); check (); return *this; }
  rpc_str &setbuf (const char *buf, size_t len)
    { str::setbuf (buf, len); check (); return *this; }
  rpc_str &setiov (const iovec *iov, int cnt)
    { str::setiov (iov, cnt); check (); return *this; }
};

template<size_t n = RPC_INFINITY> struct rpc_opaque : array<char, n> {
  rpc_opaque () { bzero (this->base (), this->size ()); }
};
template<size_t n = RPC_INFINITY> struct rpc_bytes : rpc_vec<char, n> {
  using rpc_vec<char, n>::base;
  using rpc_vec<char, n>::size;

  void setstrmem (const str &s) { rpc_vec<char,n>::set (s.cstr (), s.len ()); }
  rpc_bytes &operator= (const str &s) 
  { 
    rpc_vec<char,n>::setsize (s.len ()); 
    memcpy (base (), s.cstr (), size ()); 
    return *this; 
  }
  template<size_t m> rpc_bytes &operator= (const rpc_vec<char, m> &v)
    { rpc_vec<char, n>::operator= (v); return *this; }
  template<size_t m> rpc_bytes &operator= (const array<char, m> &v)
    { rpc_vec<char, n>::operator= (v); return *this; }
};

#if 0
template<size_t n> struct equals<rpc_opaque<n> > {
  equals () {}
  bool operator() (const rpc_opaque<n> &a, const rpc_opaque<n> &b) const
    { return !memcmp (a.base (), b.base (), n); }
};

template<size_t n> struct equals<rpc_bytes<n> > {
  equals () {}
  bool operator() (const rpc_bytes<n> &a, const rpc_bytes<n> &b) const
    { return a.size () == b.size ()
	&& !memcmp (a.base (), b.base (), a.size ()); }
};
#endif

template<size_t n> inline bool
operator== (const rpc_opaque<n> &a, const rpc_opaque<n> &b)
{
  return !memcmp (a.base (), b.base (), n);
}
template<size_t n> inline bool
operator== (const rpc_bytes<n> &a, const rpc_bytes<n> &b)
{
  return a.size () == b.size () && !memcmp (a.base (), b.base (), a.size ());
}

template<size_t n> inline bool
operator!= (const rpc_opaque<n> &a, const rpc_opaque<n> &b)
{
  return memcmp (a.base (), b.base (), n);
}
template<size_t n> inline bool
operator!= (const rpc_bytes<n> &a, const rpc_bytes<n> &b)
{
  return a.size () != b.size () || memcmp (a.base (), b.base (), a.size ());
}

#if 0
template<size_t n, size_t m> inline bool
operator== (const rpc_bytes<n> &a, const rpc_bytes<m> &b)
{
  return a.size () == b.size () && !memcmp (a.base (), b.base (), a.size ());
}
template<size_t n, size_t m> inline bool
operator== (const rpc_bytes<n> &a, const rpc_opaque<m> &b)
{
  return a.size () == b.size () && !memcmp (a.base (), b.base (), a.size ());
}
template<size_t n, size_t m> inline bool
operator== (const rpc_opaque<n> &a, const rpc_bytes<m> &b)
{
  return a.size () == b.size () && !memcmp (a.base (), b.base (), a.size ());
}

template<size_t n, size_t m> inline bool
operator!= (const rpc_bytes<n> &a, const rpc_bytes<m> &b)
{
  return a.size () != b.size () || memcmp (a.base (), b.base (), a.size ());
}
template<size_t n, size_t m> inline bool
operator!= (const rpc_bytes<n> &a, const rpc_opaque<m> &b)
{
  return a.size () != b.size () || memcmp (a.base (), b.base (), a.size ());
}
template<size_t n, size_t m> inline bool
operator!= (const rpc_opaque<n> &a, const rpc_bytes<m> &b)
{
  return a.size () != b.size () || memcmp (a.base (), b.base (), a.size ());
}
#endif

template<size_t n> struct hashfn<rpc_opaque<n> > {
  hashfn () {}
  bool operator () (const rpc_opaque<n> &a) const
    { return hash_bytes (a.base (), n); }
};
template<size_t n> struct hashfn<rpc_bytes<n> > {
  hashfn () {}
  bool operator () (const rpc_bytes<n> &a) const
    { return hash_bytes (a.base (), a.size ()); }
};


/*
 * Default traversal functions
 */

template<class T, class R, size_t n> inline bool
rpc_traverse (T &t, array<R, n> &obj)
{
  typedef typename array<R, n>::elm_t elm_t;

  elm_t *p = obj.base ();
  elm_t *e = obj.lim ();
  while (p < e)
    if (!rpc_traverse (t, *p++))
      return false;
  return true;
}

template<class T, class R, size_t n> inline bool
rpc_traverse (T &t, rpc_vec<R, n> &obj)
{
  typedef typename rpc_vec<R, n>::elm_t elm_t;

  u_int32_t size = obj.size ();
  if (!rpc_traverse (t, size) || size > obj.maxsize)
    return false;

  if (size < obj.size ())
    obj.setsize (size);
  else if (size > obj.size ()) {
    size_t maxreserve = 0x10000 / sizeof (elm_t);
    maxreserve = min<size_t> (maxreserve, size);
    if (obj.size () < maxreserve)
      obj.reserve (maxreserve - obj.size ());
  }

  elm_t *p = obj.base ();
  elm_t *e = obj.lim ();
  while (p < e)
    if (!rpc_traverse (t, *p++))
      return false;
  for (size_t i = size - obj.size (); i > 0; i--)
    if (!rpc_traverse (t, obj.push_back ()))
      return false;
  return true;
}

template<class T, class R> inline bool
rpc_traverse (T &t, rpc_ptr<R> &obj)
{
  bool nonnil = obj;
  if (!rpc_traverse (t, nonnil))
    return false;
  if (nonnil)
    return rpc_traverse (t, *obj.alloc ());
  obj.clear ();
  return true;
}

template<class T> inline bool
rpc_traverse (T &t, bool &obj)
{
  u_int32_t val = obj;
  if (!rpc_traverse (t, val))
    return false;
  obj = val;
  return true;
}

template<class T> inline bool
rpc_traverse (T &t, u_int64_t &obj)
{
  u_int32_t hi = obj >> 32;
  u_int32_t lo = obj;
  if (!rpc_traverse (t, hi) || !rpc_traverse (t, lo))
    return false;
  obj = u_int64_t (hi) << 32 | lo;
  return true;
}

template<class T> inline bool
rpc_traverse (T &t, int32_t &obj)
{
  return rpc_traverse (t, reinterpret_cast<u_int32_t &> (obj));
}

template<class T> inline bool
rpc_traverse (T &t, int64_t &obj)
{
  return rpc_traverse (t, reinterpret_cast<u_int64_t &> (obj));
}

#define DUMBTRANS(T, type)			\
inline bool					\
rpc_traverse (T &, type &)			\
{						\
  return true;					\
}

#define DUMBTRAVERSE(T)				\
DUMBTRANS(T, char)				\
DUMBTRANS(T, bool)				\
DUMBTRANS(T, u_int32_t)				\
DUMBTRANS(T, u_int64_t)				\
template<size_t n> DUMBTRANS(T, rpc_str<n>)	\
template<size_t n> DUMBTRANS(T, rpc_opaque<n>)	\
template<size_t n> DUMBTRANS(T, rpc_bytes<n>)


/*
 * Stompcast support
 */

struct stompcast_t {};
extern const stompcast_t _stompcast;

DUMBTRAVERSE(const stompcast_t)

template<class T> inline bool
stompcast (T &t)
{
  return rpc_traverse (_stompcast, t);
}

/*
 * Clearing support
 */

struct rpc_clear_t {};
extern struct rpc_clear_t _rpcclear;
struct rpc_wipe_t : public rpc_clear_t {};
extern struct rpc_wipe_t _rpcwipe;

inline bool
rpc_traverse (rpc_clear_t &, u_int32_t &obj)
{
  obj = 0;
  return true;
}
template<size_t n> inline bool
rpc_traverse (rpc_clear_t &, rpc_opaque<n> &obj)
{
  bzero (obj.base (), obj.size ());
  return true;
}
template<size_t n> inline bool
rpc_traverse (rpc_wipe_t &, rpc_opaque<n> &obj)
{
  bzero (obj.base (), obj.size ());
  return true;
}
template<size_t n> inline bool
rpc_traverse (rpc_clear_t &, rpc_bytes<n> &obj)
{
  obj.setsize (0);
  return true;
}
template<size_t n> inline bool
rpc_traverse (rpc_wipe_t &, rpc_bytes<n> &obj)
{
  bzero (obj.base (), obj.size ());
  obj.setsize (0);
  return true;
}
template<size_t n> inline bool
rpc_traverse (rpc_clear_t &, rpc_str<n> &obj)
{
  obj = rpc_emptystr;
  return true;
}
template<class T> inline bool
rpc_traverse (rpc_clear_t &, rpc_ptr<T> &obj)
{
  obj.clear ();
  return true;
}
template<class T> inline bool
rpc_traverse (rpc_wipe_t &t, rpc_ptr<T> &obj)
{
  if (obj)
    rpc_traverse (t, *obj);
  obj.clear ();
  return true;
}
template<class T, size_t n> inline bool
rpc_traverse (rpc_clear_t &, rpc_vec<T, n> &obj)
{
  obj.setsize (0);
  return true;
}
template<class T, size_t n> inline bool
rpc_traverse (rpc_wipe_t &t, rpc_vec<T, n> &obj)
{
  for (typename rpc_vec<T, n>::elm_t *p = obj.base (); p < obj.lim (); p++)
    rpc_traverse (t, *p);
  obj.setsize (0);
  return true;
}


template<class T> inline void
rpc_clear (T &obj)
{
  rpc_traverse (_rpcclear, obj);
}

template<class T> inline void
rpc_wipe (T &obj)
{
  rpc_traverse (_rpcwipe, obj);
}

/*
 *  Pretty-printing functions
 */

#define RPC_PRINT_TYPE_DECL(type)					\
void print_##type (const void *objp, const strbuf *,			\
                   int recdepth = RPC_INFINITY, const char *name = "",	\
		   const char *prefix = "");

#define RPC_PRINT_DECL(type)						     \
const strbuf &rpc_print (const strbuf &sb, const type &obj,		     \
			 int recdepth = RPC_INFINITY, const char *name = "", \
			 const char *prefix = "");

#define RPC_PRINT_DEFINE(T)						\
void									\
print_##T (const void *objp, const strbuf *sbp, int recdepth,		\
	   const char *name, const char *prefix)			\
{									\
  rpc_print (sbp ? *sbp : warnx, *static_cast<const T *> (objp),	\
             recdepth, name, prefix);					\
}
#define print_void NULL
#define print_false NULL

template<class T> struct rpc_type2str {
  static const char *type () { return typeid (T).name (); }
};
template<> struct rpc_type2str<char> {
  static const char *type () { return "opaque"; }
};
#define RPC_TYPE2STR_DECL(T)			\
template<> struct rpc_type2str<T> {		\
  static const char *type () { return #T; }	\
};

#define RPC_PRINT_GEN(T, expr)					\
const strbuf &							\
rpc_print (const strbuf &sb, const T &obj, int recdepth,	\
	   const char *name, const char *prefix)		\
{								\
  if (name) {							\
    if (prefix)							\
      sb << prefix;						\
    sb << rpc_namedecl<T >::decl (name) << " = ";		\
  }								\
  expr;								\
  if (prefix)							\
    sb << ";\n";						\
  return sb;							\
}

RPC_TYPE2STR_DECL (bool)
RPC_TYPE2STR_DECL (int32_t)
RPC_TYPE2STR_DECL (u_int32_t)
RPC_TYPE2STR_DECL (int64_t)
RPC_TYPE2STR_DECL (u_int64_t)

RPC_PRINT_TYPE_DECL (bool)
RPC_PRINT_TYPE_DECL (int32_t)
RPC_PRINT_TYPE_DECL (u_int32_t)
RPC_PRINT_TYPE_DECL (int64_t)
RPC_PRINT_TYPE_DECL (u_int64_t)

RPC_PRINT_DECL (char);
RPC_PRINT_DECL (int32_t);
RPC_PRINT_DECL (u_int32_t);
RPC_PRINT_DECL (int64_t);
RPC_PRINT_DECL (u_int64_t);
RPC_PRINT_DECL (bool);

#ifdef MAINTAINER

static inline str
rpc_dynsize (size_t n)
{
  if (n == (size_t) RPC_INFINITY)
    return "<>";
  return strbuf () << "<" << n << ">";
}
static inline str
rpc_parenptr (const str &name)
{
  if (name[0] == '*')
    return strbuf () << "(" << name << ")";
  return name;
}

template<class T> struct rpc_namedecl {
  static str decl (const char *name) {
    return strbuf () << rpc_type2str<T>::type () << " " << name;
  }
};

template<size_t n> struct rpc_namedecl<rpc_str<n> > {
  static str decl (const char *name) {
    return strbuf () << "string " << rpc_parenptr (name) << rpc_dynsize (n);
  }
};
template<class T> struct rpc_namedecl<rpc_ptr<T> > {
  static str decl (const char *name) {
    return rpc_namedecl<T>::decl (str (strbuf () << "*" << name));
  }
};
template<class T, size_t n> struct rpc_namedecl<rpc_vec<T, n> > {
  static str decl (const char *name) {
    return strbuf () << rpc_namedecl<T>::decl (rpc_parenptr (name))
		     << rpc_dynsize (n);
  }
};
template<class T, size_t n> struct rpc_namedecl<array<T, n> > {
  static str decl (const char *name) {
    return rpc_namedecl<T>::decl (rpc_parenptr (name)) << "[" << n << "]";
  }
};
template<size_t n> struct rpc_namedecl<rpc_bytes<n> > {
  static str decl (const char *name) {
    return rpc_namedecl<rpc_vec<char,n> >::decl (name);
  }
};
template<size_t n> struct rpc_namedecl<rpc_opaque<n> > {
  static str decl (const char *name) {
    return rpc_namedecl<array<char,n> >::decl (name);
  }
};

template<size_t n> const strbuf &
rpc_print (const strbuf &sb, const rpc_str<n> &obj,
	   int recdepth = RPC_INFINITY,
	   const char *name = NULL, const char *prefix = NULL)
{
  if (prefix)
    sb << prefix;
  if (name)
    sb << rpc_namedecl<rpc_str<n> >::decl (name) << " = ";
  if (obj)
    sb << "\"" << obj << "\"";	// XXX should map " to \" in string
  else
    sb << "NULL";
  if (prefix)
    sb << ";\n";
  return sb;
}

template<class T> const strbuf &
rpc_print (const strbuf &sb, const rpc_ptr<T> &obj,
	   int recdepth = RPC_INFINITY,
	   const char *name = NULL, const char *prefix = NULL)
{
  if (name) {
    if (prefix)
      sb << prefix;
    sb << rpc_namedecl<rpc_ptr<T> >::decl (name) << " = ";
  }
  if (!obj)
    sb << "NULL;\n";
  else if (!recdepth)
    sb << "...\n";
  else {
    sb << "&";
    rpc_print (sb, *obj, recdepth - 1, NULL, prefix);
  }
  return sb;
}

struct made_by_user_conversion {
  template<class T> made_by_user_conversion (const T &s) {}
};
inline bool
rpc_isstruct (const made_by_user_conversion &)
{
  return true;
}
inline bool
rpc_isstruct (u_int64_t)
{
  return false;
}

template<class T> const strbuf &
rpc_print_array_vec (const strbuf &sb, const T &obj,
		     int recdepth = RPC_INFINITY,
		     const char *name = NULL, const char *prefix = NULL)
{
  if (name) {
    if (prefix)
      sb << prefix;
    sb << rpc_namedecl<T >::decl (name) << " = ";
  }
  if (obj.size ()) {
    const char *sep;
    str npref;
    if (prefix) {
      npref = strbuf ("%s  ", prefix);
      sep = "";
      sb << "[" << obj.size () << "] {\n";
    }
    else {
      sep = ", ";
      sb << "[" << obj.size () << "] { ";
    }

    if (rpc_isstruct (obj[0])) {
      size_t i;
      size_t n = min<size_t> (obj.size (), recdepth);
      for (i = 0; i < n; i++) {
	if (i)
	  sb << sep;
	if (npref)
	  sb << npref;
	sb << "[" << i << "] = ";
	rpc_print (sb, obj[i], recdepth, NULL, npref);
      }
      if (i < obj.size ())
	sb << (i ? sep : "") << npref << "..." << (npref ? "\n" : " ");
    }
    else {
      size_t i;
      size_t n = recdepth == RPC_INFINITY ? obj.size ()
	: min  ((size_t) recdepth * 8, obj.size ());;
      if (npref)
	sb << npref;
      for (i = 0; i < n; i++) {
	if (i & 7)
	  sb << ", ";
	else if (i) {
	  sb << ",\n";
	  if (npref)
	    sb << npref;
	}
	rpc_print (sb, obj[i], recdepth, NULL, NULL);
      }
      if (i < obj.size ()) {
	if (i) {
	  sb << ",\n";
	  if (npref)
	    sb << npref;
	}
	sb << "...";
      }
      sb << (npref ? "\n" : " ");
    }

    if (prefix)
      sb << prefix << "};\n";
    else
      sb << " }";
  }
  else if (prefix)
    sb << "[0] {};\n";
  else
    sb << "[0] {}";
  return sb;
}

#define RPC_ARRAYVEC_DECL(TEMP)					\
template<class T, size_t n> const strbuf &			\
rpc_print (const strbuf &sb, const TEMP<T, n> &obj,		\
	   int recdepth = RPC_INFINITY,				\
	   const char *name = NULL, const char *prefix = NULL)	\
{								\
  return rpc_print_array_vec (sb, obj, recdepth, name, prefix);	\
}

RPC_ARRAYVEC_DECL (array)
RPC_ARRAYVEC_DECL (rpc_vec)

#undef RPC_ARRAYVEC_DECL
#define RPC_ARRAYVEC_DECL(TEMP)					\
template<size_t n> const strbuf &				\
rpc_print (const strbuf &sb, const TEMP<n> &obj,		\
	   int recdepth = RPC_INFINITY,				\
	   const char *name = NULL, const char *prefix = NULL)	\
{								\
  return rpc_print_array_vec (sb, obj, recdepth, name, prefix);	\
}

RPC_ARRAYVEC_DECL (rpc_opaque)
RPC_ARRAYVEC_DECL (rpc_bytes)

#undef RPC_ARRAYVEC_DECL

template<class T> RPC_PRINT_DECL (T);
template<class T> RPC_PRINT_GEN (T, sb << "???");

#endif /* MAINTAINER */

#endif /* !_RPCTYPES_H_ */

