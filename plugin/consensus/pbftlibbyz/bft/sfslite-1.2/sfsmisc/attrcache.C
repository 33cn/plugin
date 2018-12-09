/* $Id: attrcache.C 1754 2006-05-19 20:59:19Z max $ */

/*
 *
 * Copyright (C) 1999-2002 David Mazieres (dm@uun.org)
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

#include "nfsserv.h"

enum { max_attr_dat = 1024 };

struct attr_dat_compare {
  attr_dat_compare () {}
  int operator() (const attr_cache::attr_dat &a,
		  const attr_cache::attr_dat &b) const {
    return a.attr.expire < b.attr.expire ? -1 
      : a.attr.expire != b.attr.expire;
  }
};

//static itree_core<attr_cache::attr_dat,
//  &attr_cache::attr_dat::explink, attr_dat_compare> expirelist;

static tailq<attr_cache::attr_dat, &attr_cache::attr_dat::lrulink> lrulist;
static u_int num_attr_dat;

attr_cache::attr_dat::attr_dat (attr_cache *c, const nfs_fh3 &f,
				const fattr3exp *a)
  : cache (c), fh (f)
{
  attr = *a;
  lrulist.insert_tail (this);
  //expirelist.insert (this);
  cache->attrs.insert (this);
  num_attr_dat++;
  while (num_attr_dat > implicit_cast<u_int> (max_attr_dat))
    delete lrulist.first;
}

attr_cache::attr_dat::~attr_dat ()
{
  lrulist.remove (this);
  //expirelist.remove (this);
  cache->attrs.remove (this);
  num_attr_dat--;
}

void
attr_cache::attr_dat::touch ()
{
  lrulist.remove (this);
  lrulist.insert_tail (this);
}

void
attr_cache::attr_dat::set (const fattr3exp *a, const wcc_attr *w)
{
  //expirelist.remove (this);

  if (a->mode != attr.mode || a->uid != attr.uid || a->gid != attr.gid)
    access.clear ();
#if 0
  /* Maybe need something like this for non-unix?  We would need to
   * know also if the operation was a SETATTR. */
  else if (a->ctime != attr.ctime && (!w || w->ctime != attr.ctime))
    access.clear ();
#endif
	   
  attr = *a;
  //expirelist.insert (this);
}

void
attr_cache::flush_access (const nfs_fh3 &fh, sfs_aid aid)
{
  if (attr_dat *ad = attrs[fh])
    ad->access.remove (aid);
}

void
attr_cache::attr_enter (const nfs_fh3 &fh, const fattr3exp *a,
			const wcc_attr *w)
{
  attr_dat *ad = attrs[fh];
  if (!a) {
    if (ad)
      ad->attr.expire = 0;
  }
  else if (!ad)
    vNew attr_dat (this, fh, a);
  else {
    ad->set (a, w);
    ad->touch ();
  }
}

const fattr3exp *
attr_cache::attr_lookup (const nfs_fh3 &fh)
{
  attr_dat *ad = attrs[fh];
  if (ad && ad->valid ()) {
    ad->touch ();
    return &ad->attr;
  }
  return NULL;
}

void
attr_cache::access_enter (const nfs_fh3 &fh, sfs_aid aid,
			  u_int32_t mask, u_int32_t perm)
{
  if (attr_dat *ad = attrs[fh]) {
    ad->touch ();
    access_dat *ac = ad->access[aid];
    if (!ac)
      ad->access.insert (aid, access_dat (mask, perm));
    else {
      ac->mask |= mask;
      ac->perm = (ac->perm & ~perm) | (mask & perm);
    }
  }
}

int32_t
attr_cache::access_lookup (const nfs_fh3 &fh, sfs_aid aid, u_int32_t mask)
{
  if (attr_dat *ad = attrs[fh])
    if (ad->valid ())
      if (access_dat *ac = ad->access[aid])
	if ((mask & ac->mask) == mask) {
	  ad->touch ();
	  return ac->perm & mask;
	}
  return -1;
}

void
nfsserv_ac::getcall (nfscall *nc)
{
  if (nc->proc () == NFSPROC3_GETATTR) {
    const fattr3exp *f = ac.attr_lookup (*nc->Xtmpl getarg<nfs_fh3> ());
    if (f) {
      getattr3res res (NFS3_OK);
      *res.attributes = *reinterpret_cast<const fattr3 *> (f);
      nc->reply (&res);
      return;
    }
  }
  else if (nc->proc () == NFSPROC3_ACCESS) {
    access3args *a = nc->Xtmpl getarg<access3args> ();
    int32_t perm = ac.access_lookup (a->object, nc->getaid (), a->access);
    if (perm > 0) {
      access3res res (NFS3_OK);
      res.resok->obj_attributes.set_present (true);
      *res.resok->obj_attributes.attributes
	= *reinterpret_cast<const fattr3 *> (ac.attr_lookup (a->object));
      res.resok->access = perm;
      nc->reply (&res);
      return;
    }
  }

  mkcb (nc);
}

void
nfsserv_ac::getreply (nfscall *nc)
{
  attrvec xv;
  nfs3_getattrinfo (&xv, nc->proc (), nc->getvoidarg (), nc->resp);
  for (attrinfo *x = xv.base (); x < xv.lim (); x++) {
    if (x->fattr)
      x->fattr->expire += nc->rqtime;
    ac.attr_enter (*x->fh, x->fattr, x->wattr);
  }

  if (nc->proc () == NFSPROC3_ACCESS) {
    access3res *ares = static_cast<access3res *> (nc->resp);
    access3args *a = nc->Xtmpl getarg<access3args> ();
    if (ares->status)
      ac.flush_access (a->object, nc->getaid ());
    else
      ac.access_enter (a->object, nc->getaid (),
		       a->access, ares->resok->access);
  }

  nc->sendreply ();
}

