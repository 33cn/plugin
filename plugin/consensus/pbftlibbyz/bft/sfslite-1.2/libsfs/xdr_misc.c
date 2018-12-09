/* $Id: xdr_misc.c 435 2004-06-02 15:46:36Z max $ */

/*
 *
 * Copyright (C) 1999 David Mazieres (dm@uun.org)
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

#include "sfs-internal.h"

#ifndef HAVE_XDR_U_INT64_T
bool_t
xdr_u_int64_t (XDR *xdrs, u_int64_t *qp)
{
  long l, h;

  switch (xdrs->x_op) {
  case XDR_ENCODE:
    l = (long) (*qp & 0xffffffff);
    h = (long) ((*qp>>32) & 0xffffffff);
    return XDR_PUTLONG (xdrs, &h) && XDR_PUTLONG (xdrs, &l);
  case XDR_DECODE:
    if (!XDR_GETLONG (xdrs, &h) || !XDR_GETLONG (xdrs, &l))
      return FALSE;
    *qp = (u_int32_t) l | ((u_int64_t) h << 32);
    return TRUE;
  case XDR_FREE:
    return TRUE;
  }
  return FALSE;
}
#endif /* !HAVE_XDR_U_INT64_T */

#ifndef HAVE_XDR_INT64_T
bool_t
xdr_int64_t (XDR *xdrs, int64_t *qp)
{
  return xdr_u_int64_t (xdrs, (u_int64_t *) qp);
}
#endif /* !HAVE_XDR_INT64_T */

bool_t
xdr_bigint (XDR *xdrs, bigint *bp)
{
  return xdr_bytes (xdrs, &bp->val, &bp->len, ~0);
}
