/* $Id: ifchg.C 1117 2005-11-01 16:20:39Z max $ */

/*
 *
 * Copyright (C) 2002 David Mazieres (dm@uun.org)
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

#include "async.h"
#include "dns.h"
#include "list.h"

vec<in_addr> ifchg_addrs;
u_int64_t ifchg_count = 1;

struct ifchgcb_t {
  const cbv cb;
  list_entry<ifchgcb_t> link;
  ifchgcb_t (cbv c);
  ~ifchgcb_t ();
};

static list<ifchgcb_t, &ifchgcb_t::link> chglist;
static lazycb_t *lazy;

ifchgcb_t::ifchgcb_t (cbv c)
  : cb (c)
{
}

ifchgcb_t::~ifchgcb_t ()
{
  list<ifchgcb_t, &ifchgcb_t::link>::remove (this);
}

void
ifchgcb_test ()
{
  vec<in_addr> newaddrs;
  if (!myipaddrs (&newaddrs))
    return;
  if (newaddrs.size () == ifchg_addrs.size ()
      && !memcmp (newaddrs.base (), ifchg_addrs.base (),
		  ifchg_addrs.size () * sizeof (in_addr)))
    return;
  ifchg_addrs.swap (newaddrs);
  ++ifchg_count;
  list<ifchgcb_t, &ifchgcb_t::link> olist;
  chglist.swap (olist);
  while (ifchgcb_t *chg = olist.first) {
    olist.remove (chg);
    chglist.insert_head (chg);
    (*chg->cb) ();
  }
}

ifchgcb_t *
ifchgcb (cbv cb)
{
  if (!lazy) {
    if (!myipaddrs (&ifchg_addrs))
      fatal ("myipaddrs: %m\n");
    lazy = lazycb (60, wrap (ifchgcb_test));
  }
  ifchgcb_t *chg = New ifchgcb_t (cb);
  chglist.insert_head (chg);
  return chg;
}

void
ifchgcb_remove (ifchgcb_t *chg)
{
  delete chg;
}
