/* $Id: authuint.C 1749 2006-05-19 17:50:42Z max $ */

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

#include "arpc.h"

static int authuint_marshal (AUTH *, XDR *);
static void authuint_destroy (AUTH *);

static cxx_auth_ops auth_uint_ops = {
  0, authuint_marshal, 0, 0, authuint_destroy
};

u_int32_t
authuint_getval (AUTH *auth)
{
  assert (auth->ah_ops == TYPE_PUN_CAST(AUTH::auth_ops, &auth_uint_ops));
  return auth->ah_key.key.low;
}

AUTH *
authuint_create (u_int32_t val)
{
  AUTH *auth = New AUTH;
  bzero (auth, sizeof (*auth));
  auth->ah_key.key.low = val;
  auth->ah_ops = TYPE_PUN_CAST(AUTH::auth_ops, &auth_uint_ops);
  return auth;
}

int
authuint_marshal (AUTH *auth, XDR *x)
{
  if (u_int32_t *dp = (u_int32_t *) XDR_INLINE (x, 5*4)) {
    dp[0] = htonl (AUTH_UINT);
    dp[1] = htonl (4);
    dp[2] = htonl (auth->ah_key.key.low);
    dp[3] = htonl (AUTH_NONE);
    dp[4] = htonl (0);
    return TRUE;
  }
  else
    return xdr_putint (x, AUTH_UINT) && xdr_putint (x, 4)
      && xdr_putint (x, auth->ah_key.key.low)
      && xdr_putint (x, AUTH_NONE) && xdr_putint (x, 0);

}

void
authuint_destroy (AUTH *auth)
{
  delete auth;
}
