// -*-c++-*-
/* $Id: arpc.h 2508 2007-01-12 23:39:52Z yipal $ */

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


#ifndef _ARPC_H_
#define _ARPC_H_

#include "async.h"
#include "xdrmisc.h"

/* Solaris 2.4 specific fixes */
#ifdef NEED_XDR_CALLMSG_DECL
extern bool_t xdr_callmsg (XDR *, struct rpc_msg *);
#endif /* NEED_XDR_CALLMSG_DECL */
#ifdef xdr_authunix_parms
# undef xdr_authunix_parms
# define xdr_authunix_parms xdr_authsys_parms
#endif

#ifdef HAVE___SETERR_REPLY
#define seterr_reply __seterr_reply
#else /* !HAVE___SETERR_REPLY */
#define seterr_reply _seterr_reply
#endif /* !HAVE___SETERR_REPLY */

#define AUTH_UINT 10
AUTH *authuint_create (u_int32_t val);
u_int32_t authuint_getval (AUTH *auth);
AUTH *authopaque_create ();
void authopaque_set (AUTH *auth, const opaque_auth *cred,
		     const opaque_auth *verf = NULL);
void authopaque_set (AUTH *auth, const authunix_parms *aup);
extern "C" {
AUTH *authunixint_create (const char *host, u_int32_t uid, u_int32_t gid,
			  u_int32_t ngroups, const u_int32_t *groups);
AUTH *authunix_create_realids ();
}

/* For platforms that define AUTH::auth_ops with K and R prototypes */
struct cxx_auth_ops {
  void (*ah_nextverf) (AUTH *);
  int (*ah_marshal) (AUTH *, XDR *);
  int (*ah_validate) (AUTH *, struct opaque_auth *);
  int (*ah_refresh) (AUTH *);
  void (*ah_destroy) (AUTH *);
};

class auto_auth {
  AUTH *auth;

  auto_auth (const auto_auth &);
  auto_auth &operator= (const auto_auth &);
  void destroy () { if (auth) AUTH_DESTROY (auth); }
public:
  auto_auth (AUTH *a = NULL) : auth (a) {}
  auto_auth (auto_auth &a) : auth (a.auth) { a.auth = NULL; }
  ~auto_auth () { destroy (); }
  operator AUTH *() const { return auth; }
  auto_auth &operator= (AUTH *a) { destroy (); auth = a; return *this; }
  auto_auth &operator= (auto_auth &a)
    { destroy (); auth = a.auth; a.auth = NULL; return *this; }
};

#include "ihash.h"
#include "list.h"
#include "refcnt.h"

class xhinfo;

#include "axprt.h"
#include "aclnt.h"
#include "asrv.h"
#include "xhinfo.h"

ptr<axprt_dgram> udpxprt ();
ptr<aclnt> udpclnt ();

void __acallrpc (const char *host, u_int port,
		 u_int32_t prog, u_int32_t vers, u_int32_t proc,
		 sfs::xdrproc_t inxdr, void *inmem,
		 sfs::xdrproc_t outxdr, void *outmem,
		 aclnt_cb cb, AUTH *auth);
void __acallrpc (in_addr host, u_int port,
		 u_int32_t prog, u_int32_t vers, u_int32_t proc,
		 sfs::xdrproc_t inxdr, void *inmem,
		 sfs::xdrproc_t outxdr, void *outmem,
		 aclnt_cb cb, AUTH *auth);

inline void
acallrpc (const char *host, const rpc_program &rp, u_int32_t proc,
	  void *in, void *out, aclnt_cb cb, u_int port = 0,
	  AUTH *auth = NULL)
{
  assert (proc < rp.nproc);
  __acallrpc (host, port, rp.progno, rp.versno, proc,
	      rp.tbl[proc].xdr_arg, in, rp.tbl[proc].xdr_res, out,
	      cb, auth);
}
inline void
acallrpc (in_addr host, const rpc_program &rp, u_int32_t proc,
	  void *in, void *out, aclnt_cb cb, u_int port = 0,
	  AUTH *auth = NULL)
{
  assert (proc < rp.nproc);
  __acallrpc (host, port, rp.progno, rp.versno, proc,
	      rp.tbl[proc].xdr_arg, in, rp.tbl[proc].xdr_res, out,
	      cb, auth);
}
void acallrpc (const sockaddr_in *sinp, const rpc_program &rp,
	       u_int32_t proc, void *in, void *out, aclnt_cb cb,
	       AUTH *auth = NULL);

void pmap_map (int fd, const rpc_program &rp,
	       callback<void, bool>::ptr cb = NULL);

// Cast pointer p to be of type Y*, in a way that doesn't
// trigger GCC 4.1's mysterious 'type-punned pointer' warning.
#if __GNUC__ >= 4
template<class Y, class X>
inline Y *gcc41_cast (X *x)
{
  return reinterpret_cast<Y *> (reinterpret_cast<void *> (x));
}
# define TYPE_PUN_CAST(Y,p) (gcc41_cast<Y, typeof(*(p))> (p))
#else  /* __GNUC__ < 4 */
# define TYPE_PUN_CAST(Y,p) (reinterpret_cast<Y *> (p))
#endif /* __GNUC__ < 4 */

#endif /* ! _ARPC_H_ */
