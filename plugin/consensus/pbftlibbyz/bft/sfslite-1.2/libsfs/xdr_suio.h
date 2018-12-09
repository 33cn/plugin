/* $Id: xdr_suio.h 435 2004-06-02 15:46:36Z max $ */

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

#include "suio.h"

#ifdef __cplusplus
static inline suio *&
xsuio (XDR *x)
{
  return (suio *&) x->x_private;
}
#else /* !__cplusplus */
#define xsuio(x) (*(suio **) &(x)->x_private)
#endif /* !__cplusplus */

#ifdef __cplusplus
extern "C" {
#endif /* __cplusplus */

extern void xdrsuio_create (XDR *, enum xdr_op);
extern void xdrsuio_destroy (XDR *);

static inline bool_t
isxdrsuio (XDR *x)
{
  return x->x_ops->x_destroy == xdrsuio_destroy;
}

static inline int32_t *
xdrsuio_inline (XDR *xdrs, u_int count)
{
  register suio *const uio = xsuio (xdrs);
  char *space;
  assert (! (count & 3));
  if (count <= SBUFSIZ) {
    space = suio_getspace_aligned (uio, count);
    suio_print (uio, space, count);
  }
  else {
    space = (char *) xmalloc (count);
    suio_print (uio, space, count);
    suio_callfree (uio, space);
  }
  assert (!((u_long) space & 0x3));
  return (int32_t *) space;
}

static inline bool_t
xdrsuio_putlong (XDR *xdrs, long *lp)
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
  return (xsuio(xdrs)->uio_resid);
}

#ifdef __cplusplus
}
#endif /* __cplusplus */
