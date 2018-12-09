/* $Id: xdr_suio.c 435 2004-06-02 15:46:36Z max $ */

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


#include "xdr_suio.h"

struct cxx_xdr_ops {
  bool_t (*x_getlong) (XDR *, long *);
  bool_t (*x_putlong) (XDR *, long *);
  bool_t (*x_getbytes) (XDR *, caddr_t, u_int);
  bool_t (*x_putbytes) (XDR *, caddr_t, u_int);
  u_int (*x_getpostn) (XDR *);
  bool_t (*x_setpostn) (XDR *, u_int);
  int32_t *(*x_inline) (XDR *, u_int);
  void (*x_destroy) (XDR *);
};
typedef struct cxx_xdr_ops cxx_xdr_ops;

static cxx_xdr_ops xsops = {
    NULL,
    xdrsuio_putlong,
    NULL,
    xdrsuio_putbytes,
    xdrsuio_getpostn,
    NULL,
    xdrsuio_inline,
    xdrsuio_destroy,
};

static const XDR xsproto = {
  XDR_ENCODE,
#ifdef __cplusplus
  (XDR::xdr_ops *) &xsops,
#else /* !__cplusplus */
  (struct xdr_ops *) &xsops,
#endif /* !__cplusplus */
  NULL, NULL, NULL, 0
};

void
xdrsuio_create (XDR *xdrs, enum xdr_op op)
{
  assert (op == XDR_ENCODE);
  *xdrs = xsproto;
  xsuio (xdrs) = suio_alloc ();
}

void
xdrsuio_destroy (XDR *xdrs)
{
  suio_free (xsuio (xdrs));
}

#ifdef __cplusplus
suio *
xdrsuio::uio ()
{
  return xsuio (&x);
}

iovec *
xdrsuio::iov ()
{
  return xsuio (&x)->uio_iov;
}

u_int
xdrsuio::iovcnt ()
{
  return xsuio (&x)->uio_iovcnt;
}
#endif /* __cplusplus */

bool_t
xdr_flatten (xdrproc_t proc, void *objp, char **rdat, u_int *rlen)
{
  XDR x;
  xdrsuio_create (&x, XDR_ENCODE);
  if (!proc (&x, objp)) {
#ifdef DMALLOC
    *rdat = (char *) 0xdeadf1a5;
    *rlen = 0xdeadf1a5;
#endif
    xdrsuio_destroy (&x);
    return FALSE;
  }
  else {
    *rlen = xsuio(&x)->uio_resid;
    *rdat = suio_flatten (xsuio (&x));
    xdrsuio_destroy (&x);
    return TRUE;
  }
}
