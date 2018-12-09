/* $Id: xdr_suio.h 1749 2006-05-19 17:50:42Z max $ */

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

static inline suio *&
xsuio (XDR *x)
{
  return * TYPE_PUN_CAST(suio *, &x->x_private);
}

extern "C" {

extern void xdrsuio_create (XDR *, enum xdr_op);
extern void xdrsuio_destroy (XDR *);

/* xdrsuio_scrub is just like xdrsuio, except it takes extra time to
 * scrub (bzero) any temporary memmory it has allocated.  This may be
 * desirable for debugging or in cases where one is marshaling
 * sensitive data. */
class scrubbed_suio : public suio {
  static void scrub_and_free (void *buf, size_t len);
public:
  scrubbed_suio () { deallocator = scrub_and_free; }
  ~scrubbed_suio ();
};
extern void xdrsuio_scrub_create (XDR *, enum xdr_op);
extern void xdrsuio_scrub_destroy (XDR *);

static inline bool_t
isxdrsuio (XDR *x)
{
  return ((void (*) (...)) x->x_ops->x_destroy
	  == (void (*) (...)) xdrsuio_destroy);
}

static inline int32_t *
xdrsuio_inline (XDR *xdrs, u_int count)
{
  register suio *const uio = xsuio (xdrs);
  char *space;
  assert (!(count & 3));
  space = uio->getspace_aligned (count);
  uio->print (space, count);
  assert (!((u_long) space & 0x3));
  return (int32_t *) space;
}

static inline bool_t
xdrsuio_putlong (XDR *xdrs, xdrlong_t *lp)
{
  *xdrsuio_inline (xdrs, 4) = htonl (*lp);
  return TRUE;
}

static inline bool_t
xdrsuio_putbytes (XDR *xdrs, caddr_t addr, u_int len)
{
  /* assert (! (len & 3)); */
  suio_printcheck (xsuio (xdrs), addr, len);
  return TRUE;
}

static inline u_int
xdrsuio_getpostn (XDR *xdrs)
{
  return (xsuio(xdrs)->resid ());
}

}
