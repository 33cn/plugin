/* $Id: authopaque.C 3758 2008-11-13 00:36:00Z max $ */

/*
 *
 * Copyright (C) 2001 David Mazieres (dm@uun.org)
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


#include "arpc.h"

extern "C" {
  /* for SunOS 5.7 */
  bool_t xdr_opaque_auth(XDR *, struct opaque_auth *);
}

static int authopaque_marshal (AUTH *, XDR *);
static void authopaque_destroy (AUTH *);

static cxx_auth_ops auth_opaque_ops = {
  0, authopaque_marshal, 0, 0, authopaque_destroy
};

AUTH *
authopaque_create ()
{
  AUTH *auth = New AUTH;
  bzero (auth, sizeof (*auth));
  auth->ah_ops = TYPE_PUN_CAST(AUTH::auth_ops, &auth_opaque_ops);
  auth->ah_cred.oa_base = static_cast<caddr_t> (xmalloc (MAX_AUTH_BYTES));
  auth->ah_verf.oa_base = static_cast<caddr_t> (xmalloc (MAX_AUTH_BYTES));
  authopaque_set (auth, NULL, NULL);
  return auth;
}

inline void
authcopy (opaque_auth *dst, const opaque_auth *src)
{
  if (src) {
    dst->oa_flavor = src->oa_flavor;
    dst->oa_length = src->oa_length;
    assert (dst->oa_length <= MAX_AUTH_BYTES);
    memcpy (dst->oa_base, src->oa_base, dst->oa_length);
  }
  else {
    dst->oa_flavor = AUTH_NONE;
    dst->oa_length = 0;
  }
}

void
authopaque_set (AUTH *auth, const opaque_auth *cred, const opaque_auth *verf)
{
  assert (auth->ah_ops == TYPE_PUN_CAST(AUTH::auth_ops, &auth_opaque_ops));
  authcopy (&auth->ah_cred, cred);
  authcopy (&auth->ah_verf, verf);
}

void
authopaque_set (AUTH *auth, const authunix_parms *aup)
{
  assert (auth->ah_ops == TYPE_PUN_CAST (AUTH::auth_ops, &auth_opaque_ops));

  auth->ah_cred.oa_flavor = AUTH_UNIX;
  xdrmem xdr (auth->ah_cred.oa_base, MAX_AUTH_BYTES);
  u_int ng = min<u_int> (aup->aup_len, 16);
  u_int mnl = strlen (aup->aup_machname);
  auth->ah_cred.oa_length = 20 + 4*ng + ((mnl + 3) & ~3);
  xdr_putint (&xdr, aup->aup_time);
  xdr_putint (&xdr, mnl);
  xdr_putpadbytes (&xdr, aup->aup_machname, mnl);
  xdr_putint (&xdr, aup->aup_uid);
  xdr_putint (&xdr, aup->aup_gid);
  xdr_putint (&xdr, ng);
  for (u_int i = 0; i < ng; i++)
    xdr_putint (&xdr, aup->aup_gids[i]);
  assert (XDR_GETPOS (&xdr) == auth->ah_cred.oa_length);

  authcopy (&auth->ah_verf, NULL);
}

int
authopaque_marshal (AUTH *auth, XDR *x)
{
  return xdr_opaque_auth (x, &auth->ah_cred)
    && xdr_opaque_auth (x, &auth->ah_verf);
}

void
authopaque_destroy (AUTH *auth)
{
  delete auth->ah_cred.oa_base;
  delete auth->ah_verf.oa_base;
  delete auth;
}
