// -*-c++-*-
/* $Id: xdrmisc.h 3758 2008-11-13 00:36:00Z max $ */

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


#ifndef _ARPC_XDRMISC_H_
#define _ARPC_XDRMISC_H_ 1

#include "sysconf.h"

extern "C" {
#define xdrproc_t sun_xdrproc_t
#define xdr_free foil_xdr_free
#define xdr_void foil_xdr_void
#define xdr_bool foil_xdr_bool
#define xdr_int foil_xdr_int
#define xdr_int32_t foil_xdr_int32_t
#define xdr_u_int32_t foil_xdr_u_int32_t
#define xdr_hyper foil_xdr_hyper
#define xdr_int64_t foil_xdr_int64_t
#define xdr_u_int64_t foil_xdr_u_int64_t
#define xdr_string foil_xdr_string
#define xdr_pointer foil_xdr_pointer
#define xdr_array foil_xdr_array
#define xdr_vector foil_xdr_vector

#define pmaplist foil_pmaplist
#define xdr_pmaplist foil_xdr_pmaplist

#include <rpc/rpc.h>

#undef xdrproc_t
#undef xdr_free
#undef xdr_void
#undef xdr_bool
#undef xdr_int
#undef xdr_int32_t
#undef xdr_u_int32_t
#undef xdr_hyper
#undef xdr_int64_t
#undef xdr_u_int64_t
#undef xdr_string
#undef xdr_pointer
#undef xdr_array
#undef xdr_vector

#undef pmaplist
#undef xdr_pmaplist
#undef PMAPPROC_NULL
#undef PMAPPROC_SET
#undef PMAPPROC_UNSET
#undef PMAPPROC_GETPORT
#undef PMAPPROC_DUMP
#undef PMAPPROC_CALLIT
}

#define BOOL bool_t

namespace sfs {
  typedef BOOL (*xdrproc_t) (XDR *, void *);
}
#include "rpctypes.h"

#ifdef __APPLE__
# define XDROPS_KNRPROTO 1
#endif /* __APPLE__ */

#ifdef XDROPS_KNRPROTO
#undef xdrlong_t
#define xdrlong_t long

#undef XDR_GETLONG
#define XDR_GETLONG(xdrs, longp)					\
	(*(bool_t (*) (...)) (xdrs)->x_ops->x_getlong) (xdrs, longp)

#undef XDR_PUTLONG
#define XDR_PUTLONG(xdrs, longp)					\
	(*(bool_t (*) (...)) (xdrs)->x_ops->x_putlong) (xdrs, longp)

#undef XDR_GETBYTES
#define XDR_GETBYTES(xdrs, addr, len)					  \
	(*(bool_t (*) (...)) (xdrs)->x_ops->x_getbytes) (xdrs, addr, len)

#undef XDR_PUTBYTES
#define XDR_PUTBYTES(xdrs, addr, len)					  \
	(*(bool_t (*) (...)) (xdrs)->x_ops->x_putbytes) (xdrs, addr, len)

#undef XDR_GETPOS
#define XDR_GETPOS(xdrs)					\
	(*(u_int (*) (...)) (xdrs)->x_ops->x_getpostn) (xdrs)

#undef XDR_INLINE
#define XDR_INLINE(xdrs, len)						\
	(*(xdrlong_t *(*) (...)) (xdrs)->x_ops->x_inline) (xdrs, len)

#undef XDR_DESTROY
#define XDR_DESTROY(xdrs)					\
	if ((xdrs)->x_ops->x_destroy)				\
	  (*(void (*) (...)) (xdrs)->x_ops->x_destroy) (xdrs)

#undef AUTH_MARSHALL
#define AUTH_MARSHALL(auth, xdrs)					\
	(*(bool_t (*) (...)) (auth)->ah_ops->ah_marshal) (auth, xdrs)

#undef AUTH_DESTROY
#define AUTH_DESTROY(auth)					\
	(*(void (*) (...)) (auth)->ah_ops->ah_destroy) (auth)

#endif /* XDROPS_KNRPROTO */

/*
 * rpc_traverse instantiation for sunrpc XDRs
 */

inline bool
xdr_putint (XDR *xdrs, u_int32_t val)
{
  xdrlong_t l = val;
  return XDR_PUTLONG (xdrs, &l);
}
inline bool
xdr_getint (XDR *xdrs, u_int32_t &val)
{
  xdrlong_t l;
  if (!XDR_GETLONG (xdrs, &l))
    return false;
  val = l;
  return true;
}

inline bool
xdr_puthyper (XDR *xdrs, u_int64_t val)
{
  xdrlong_t h = val >> 32, l = val & 0xffffffff;
  return XDR_PUTLONG (xdrs, &h) && XDR_PUTLONG (xdrs, &l);
}
inline bool
xdr_gethyper (XDR *xdrs, u_int64_t &val)
{
  xdrlong_t h, l;
  if (!XDR_GETLONG (xdrs, &h) || !XDR_GETLONG (xdrs, &l))
    return false;
  val = (u_int64_t) h << 32 | l;
  return true;
}

extern const char __xdr_zero_bytes[4];
inline bool
xdr_putpadbytes (XDR *xdrs, const void *p, size_t n)
{
  if (n) {
    if (!XDR_PUTBYTES (xdrs, static_cast<char *> (const_cast<void *> (p)), n))
      return false;
    if (size_t nn = -n & 3)
      return XDR_PUTBYTES (xdrs, const_cast<char *> (__xdr_zero_bytes), nn);
  }
  return true;
}
inline bool
xdr_getpadbytes (XDR *xdrs, void *p, size_t n)
{
  char garbage[3];
  if (!XDR_GETBYTES (xdrs, static_cast<caddr_t> (p), n))
      return false;
  if (size_t nn = -n & 3)
    return XDR_GETBYTES (xdrs, garbage, nn);
  return true;
}

inline bool
rpc_traverse (XDR *xdrs, u_int32_t &obj)
{
  switch (xdrs->x_op) {
  case XDR_ENCODE:
    return xdr_putint (xdrs, obj);
  case XDR_DECODE:
    return xdr_getint (xdrs, obj);
  default:
    return true;
  }
}

template<size_t n> inline bool
rpc_traverse (XDR *xdrs, rpc_opaque<n> &obj)
{
  switch (xdrs->x_op) {
  case XDR_ENCODE:
    return xdr_putpadbytes (xdrs, obj.base (), obj.size ());
  case XDR_DECODE:
    return xdr_getpadbytes (xdrs, obj.base (), obj.size ());
  default:
    return true;
  }
}

template<size_t max> inline bool
rpc_traverse (XDR *xdrs, rpc_bytes<max> &obj)
{
  switch (xdrs->x_op) {
  case XDR_ENCODE:
    return xdr_putint (xdrs, obj.size ())
      && xdr_putpadbytes (xdrs, obj.base (), obj.size ());
  case XDR_DECODE:
    {
      u_int32_t size;
      if (!xdr_getint (xdrs, size) || size > obj.maxsize)
	return false;
      /* Assume XDR_INLINE works -- even though it could hypothetically
       * fail for some types of XDR.  The alternative would be to use:
       *    obj.setsize (size);
       *    return xdr_getpadbytes (xdrs, obj.base (), size);
       * however, then if size is garbage the program will die in
       * obj.setsize (size) from trying to allocate too much memory.
       */
      char *dp = (char *) XDR_INLINE (xdrs, (size + 3) & ~3);
      if (!dp)
	return false;
      obj.setsize (size);
      memcpy (obj.base (), dp, size);
      return true;
    }
  default:
    return true;
  }
}

template<size_t max> inline bool
rpc_traverse (XDR *xdrs, rpc_str<max> &obj)
{
  switch (xdrs->x_op) {
  case XDR_ENCODE:
    return obj && xdr_putint (xdrs, obj.len ())
      && xdr_putpadbytes (xdrs, obj.cstr (), obj.len ());
  case XDR_DECODE:
    {
      u_int32_t size;
      if (!xdr_getint (xdrs, size) || size > max)
	return false;
      /* See comment for rpc_bytes */
      char *dp = (char *) XDR_INLINE (xdrs, (size + 3) & ~3);
      if (!dp || memchr (dp, '\0', size))
	return false;
      obj.setbuf (dp, size);
      return true;
    }
  default:
    return true;
  }
}

inline bool
rpc_traverse (XDR *xdrs, str &obj)
{
  switch (xdrs->x_op) {
  case XDR_ENCODE:
    return obj && xdr_putint (xdrs, obj.len ())
      && xdr_putpadbytes (xdrs, obj.cstr (), obj.len ());
  case XDR_DECODE:
    {
      u_int32_t size;
      if (!xdr_getint (xdrs, size))
	return false;
      mstr m (size);
      if (!xdr_getpadbytes (xdrs, m, size) || memchr (m.cstr (), '\0', size))
	return false;
      obj = m;
    }
    return true;
  default:
    return true;
  }
}

template<class T> inline void
rpc_destruct (T *objp)
{
  objp->~T ();
}

inline void
xdr_free (sfs::xdrproc_t proc, void *objp)
{
  XDR x;
  x.x_op = XDR_FREE;
  proc (&x, objp);
}

inline void
xdr_delete (sfs::xdrproc_t proc, void *objp)
{
  xdr_free (proc, objp);
  operator delete (objp);
}

class auto_xdr_delete {
  const sfs::xdrproc_t proc;
  void *const objp;

  auto_xdr_delete (const auto_xdr_delete &);
  auto_xdr_delete &operator= (const auto_xdr_delete &);

public:
  auto_xdr_delete (sfs::xdrproc_t p, void *o) : proc (p), objp (o) {}
  ~auto_xdr_delete () { xdr_delete (proc, objp); }
};

#define DECLXDR(type)				\
extern BOOL xdr_##type (XDR *, void *);		\
extern void *type##_alloc ();
DECLXDR(void)
DECLXDR(false)
DECLXDR(string)
DECLXDR(bool)
DECLXDR(int)
DECLXDR(int32_t)
DECLXDR(u_int32_t)
DECLXDR(int64_t)
DECLXDR(u_int64_t)
#undef DECLXDR

#ifdef MAINTAINER

# define RPC_TYPE_DECL(type)			\
RPC_TYPE2STR_DECL (type)			\
RPC_PRINT_DECL (type)				\
RPC_PRINT_TYPE_DECL (type)

# define XDRTBL_DECL(proc, arg, res)			\
{							\
  #proc,						\
  &typeid (arg), arg##_alloc, xdr_##arg, print_##arg,	\
  &typeid (res), res##_alloc, xdr_##res, print_##res	\
},

#else /* !MAINTAINER */

# define RPC_TYPE_DECL(type)
# define XDRTBL_DECL(proc, arg, res)		\
{						\
  #proc,					\
  &typeid (arg), arg##_alloc, xdr_##arg, NULL,	\
  &typeid (res), res##_alloc, xdr_##res, NULL	\
},

#endif /* !MAINTAINER */

#define RPC_STRUCT_DECL(type) RPC_TYPE_DECL (type)
#define RPC_UNION_DECL(type) RPC_TYPE_DECL (type)
#define RPC_ENUM_DECL(type) RPC_TYPE_DECL (type)
#ifdef MAINTAINER
#define RPC_TYPEDEF_DECL(type) RPC_PRINT_TYPE_DECL (type)
#endif /* MAINTAINER */

#define RPCUNION_SET(type, field) field.select ()
#define RPCUNION_TRAVERSE(type, field) return rpc_traverse (t, *obj.field)
#define RPCUNION_STOMPCAST(type, field) field.Xstompcast ()
#define RPCUNION_REC_STOMPCAST(type, field) \
  obj.field.Xstompcast (); return rpc_traverse (s, *obj.field)
#if __GNUC__ >= 4
#define RPCUNION_XXX_GCC40 return false;
#else /* not gcc 4 */
#define RPCUNION_XXX_GCC40
#endif /* not gcc 4 */

class xdrbase : public XDR {
protected:
  xdrbase () {}
  ~xdrbase () { XDR_DESTROY (implicit_cast<XDR *> (this)); }
  xdrbase (const xdrbase &);	// No copying
  const xdrbase &operator= (const xdrbase &);
public:
  XDR *xdrp () { return this; }
};

extern "C" void xdrsuio_create (XDR *, enum xdr_op);
extern "C" void xdrsuio_scrub_create (XDR *, enum xdr_op);
struct xdrsuio : xdrbase {
  explicit xdrsuio (xdr_op op = XDR_ENCODE, bool scrub = false) {
    if (scrub)
      xdrsuio_scrub_create (this, op);
    else
      xdrsuio_create (this, op);
  }
  suio *uio ();
  const iovec *iov ();
  u_int iovcnt ();
};

struct xdrmem : xdrbase {
  explicit xdrmem (char *base, size_t len, xdr_op op = XDR_DECODE)
    { xdrmem_create (this, base, len, op); }
  explicit xdrmem (const char *base, size_t len, xdr_op op = XDR_DECODE) {
    assert (op == XDR_DECODE);
    xdrmem_create (this, const_cast <char *> (base), len, op);
  }
};

inline const str &str2wstr (const str &s);
template<class T> str
xdr2str (const T &t, bool scrub = false)
{
  xdrsuio x (XDR_ENCODE, scrub);
  XDR *xp = &x;
  if (!rpc_traverse (xp, const_cast<T &> (t)))
    return NULL;
  mstr m (x.uio ()->resid ());
  x.uio ()->copyout (m);
  if (scrub)
    return str2wstr (m);
  return m;
}

template<class T> bool
str2xdr (T &t, const str &s)
{
  xdrmem x (s, s.len ());
  XDR *xp = &x;
  return rpc_traverse (xp, t);
}

template<class T, size_t n> bool
xdr2bytes (rpc_bytes<n> &out, const T &t, bool scrub = false)
{
  xdrsuio x (XDR_ENCODE, scrub);
  XDR *xp = &x;
  if (!rpc_traverse (xp, const_cast<T &> (t)) || x.uio ()->resid () > n)
    return false;
  if (scrub)
    bzero (out.base (), out.size ());
  out.setsize (x.uio ()->resid ());
  x.uio ()->copyout (out.base ());
  return true;
}

template<class T, size_t n> bool
bytes2xdr (T &t, const rpc_bytes<n> &in)
{
  xdrmem x (in.base (), in.size ());
  XDR *xp = &x;
  return rpc_traverse (xp, t);
}

template<class T> bool
buf2xdr (T &t, const void *buf, size_t len)
{
  xdrmem x (reinterpret_cast<const char *> (buf), len);
  XDR *xp = &x;
  return rpc_traverse (xp, t);
}

#endif /* !_ARPC_XDRMISC_H_ */
