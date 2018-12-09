// -*-c++-*-
/* $Id: xhinfo.h 2728 2007-04-16 13:29:12Z max $ */

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


class xhinfo : public virtual refcount {
  u_int nsvc;

protected:
  xhinfo (const ref<axprt> &);
  ~xhinfo ();

public:
  const ref<axprt> xh;
  list<aclnt, &aclnt::xhlink> clist;
  ihash<const progvers, asrv, &asrv::pv, &asrv::xhlink> stab;
  ihash<const u_int32_t, callbase, &callbase::xid, &callbase::hlink> xidtab;
  ihash_entry<xhinfo> hlink;

  void seteof (ref<xhinfo>, const sockaddr *);
  void dispatch (const char *, ssize_t, const sockaddr *);
  u_int svcnum () const { return nsvc; }
  u_int svcadd () { return nsvc++; }
  u_int svcdel () { assert (nsvc); return nsvc--; }

  bool ateof () { return xh->ateof (); }
  static ptr<xhinfo> lookup (const ref<axprt> &);
  static void xon (const ref<axprt> &x, bool receive = true);

  u_int64_t max_acked_offset;  // aclnts update this
};

