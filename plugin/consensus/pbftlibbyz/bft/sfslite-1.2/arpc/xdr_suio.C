/* $Id: xdr_suio.C 1749 2006-05-19 17:50:42Z max $ */

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

void
scrubbed_suio::scrub_and_free (void *buf, size_t len)
{
  bzero (buf, len);
  default_deallocator (buf, len);
}

scrubbed_suio::~scrubbed_suio ()
{
  /* XXX - GCC BUG:  We have to work around a serious, serious
   * compiler bug here.  Gcc does not call our supertype destructor
   * here for some incomprehensible reason.  Thsu, we must clear ()
   * and free everything ourselves.  (Though taking care that if the
   * compiler gets fixed, it won't break our code.)  */
  clear ();
  bzero (defbuf, sizeof (defbuf));
}

struct cxx_xdr_ops {
  bool_t (*x_getlong) (XDR *, xdrlong_t *);
  bool_t (*x_putlong) (XDR *, xdrlong_t *);
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
  TYPE_PUN_CAST (xdr_ops_t, &xsops),
  NULL, NULL, NULL, 0
};

static cxx_xdr_ops xsops_scrub = {
  NULL,
  xdrsuio_putlong,
  NULL,
  xdrsuio_putbytes,
  xdrsuio_getpostn,
  NULL,
  xdrsuio_inline,
  xdrsuio_scrub_destroy,
};
static const XDR xsproto_scrub = {
  XDR_ENCODE,
  TYPE_PUN_CAST (xdr_ops_t,  &xsops_scrub),
  NULL, NULL, NULL, 0
};

void
xdrsuio_create (XDR *xdrs, enum xdr_op op)
{
  assert (op == XDR_ENCODE);
  *xdrs = xsproto;
  xsuio (xdrs) = New suio;
}

void
xdrsuio_destroy (XDR *xdrs)
{
  delete xsuio (xdrs);
}

void
xdrsuio_scrub_create (XDR *xdrs, enum xdr_op op)
{
  assert (op == XDR_ENCODE);
  *xdrs = xsproto_scrub;
  xsuio (xdrs) = implicit_cast<suio *> (New scrubbed_suio);
}

void
xdrsuio_scrub_destroy (XDR *xdrs)
{
  scrubbed_suio *xsp = static_cast<scrubbed_suio *> (xsuio (xdrs));
  // XXX - gcc bug; this must read ::delete even though it should be delete
  ::delete xsp;
}

suio *
xdrsuio::uio ()
{
  return xsuio (this);
}

const iovec *
xdrsuio::iov ()
{
  return xsuio (this)->iov ();
}

u_int
xdrsuio::iovcnt ()
{
  return xsuio (this)->iovcnt ();
}
