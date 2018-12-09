/* $Id: xhinfo.C 2728 2007-04-16 13:29:12Z max $ */

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

xhinfo::xhinfo (const ref<axprt> &x)
  : nsvc (0), xh (x), max_acked_offset (0)
{
  xh->xhip = this;
  xh->setrcb (wrap (this, &xhinfo::dispatch));
}

xhinfo::~xhinfo ()
{
  xh->xhip = NULL;
  xh->setrcb (NULL);
}

ptr<xhinfo>
xhinfo::lookup (const ref<axprt> &x)
{
  if (x->ateof ())
    return NULL;
  if (xhinfo *xip = x->xhip)
    return mkref (xip);
  return New refcounted<xhinfo> (x);
}

void
xhinfo::seteof (ref<xhinfo> xi, const sockaddr *src)
{
  if (xh->connected) {
    xh->setrcb (NULL);
    if (clist.first)
      aclnt::dispatch (xi, NULL, 0, src);
    if (stab.first ())
      asrv::dispatch (xi, NULL, 0, src);
  }
}

void
xhinfo::dispatch (const char *msg, ssize_t len, const sockaddr *src)
{
  ref<xhinfo> xi = mkref (this);
  if (len < 8) {
    if (len > 0)
      warn ("xhinfo::dispatch: packet too short\n");
    seteof (xi, src);
    return;
  }
  if (len & 3) {
    if (len > 0)
      warn ("xhinfo::dispatch: packet not multiple of 4 bytes\n");
    seteof (xi, src);
    return;
  }
  switch (getint (msg +  4)) {
  case CALL:
    if (stab.first ())
      asrv::dispatch (xi, msg, len, src);
    else {
      warn ("xhinfo::dispatch: unanticipated RPC CALL\n");
      seteof (xi, src);
    }
    break;
  case REPLY:
    if (clist.first)
      aclnt::dispatch (xi, msg, len, src);
    else {
      warn ("xhinfo::dispatch: unanticipated RPC REPLY\n");
      seteof (xi, src);
    }
    break;
  default:
    warn ("xhinfo::dispatch: unknown RPC message type\n");
    seteof (xi, src);
    break;
  }
}

void
xhinfo::xon (const ref<axprt> &x, bool receive)
{
  ptr<xhinfo> xi = lookup (x);
  assert (xi);
  if (!receive)
    x->setrcb (NULL);
  else if (!xi->ateof ())
    x->setrcb (wrap (&*xi, &xhinfo::dispatch));
}
