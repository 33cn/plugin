/* $Id: nfs3attr.C 1754 2006-05-19 20:59:19Z max $ */

/*
 *
 * Copyright (C) 2001 David Mazieres (dm@uun.org)
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

#include "nfstrans.h"
#include "nfs3_nonnul.h"

DUMBTRAVERSE (attrvec)
inline bool
rpc_traverse (attrvec &av, fattr3exp &obj)
{
  assert (!av[0].fattr);
  av[0].fattr = &obj;
  return true;
}

inline bool
rpc_traverse (attrvec &av, wcc_data &obj)
{
  assert (!av[0].wdata);
  av[0].set_wcc (&obj);
  return true;
}
bool rpc_traverse (attrvec &av, lookup3resok &obj);
bool rpc_traverse (attrvec &av, diropres3ok &obj);
bool rpc_traverse (attrvec &av, dirlist3 &obj);
bool rpc_traverse (attrvec &av, entryplus3 &obj);
bool rpc_traverse (attrvec &av, rename3wcc &obj);
bool rpc_traverse (attrvec &av, link3wcc &obj);

attrinfo::attrinfo ()
  : fh (NULL), fattr (NULL), wattr (NULL), wdata (NULL)
{
}

void
attrinfo::set_wcc (wcc_data *wd)
{
  wdata = wd;
  if (wd->before.present)
    wattr = wd->before.attributes;
  if (wd->after.present)
    fattr = wd->after.attributes;
}

bool
rpc_traverse (attrvec &av, lookup3resok &obj)
{
  if (obj.dir_attributes.present)
    av[0].fattr = obj.dir_attributes.attributes.addr ();
  if (obj.obj_attributes.present) {
    av.push_back ();
    av[1].fh = &obj.object;
    av[1].fattr = obj.obj_attributes.attributes.addr ();
  }
  return true;
}

bool
rpc_traverse (attrvec &av, diropres3ok &obj)
{
  av[0].set_wcc (&obj.dir_wcc);
  if (obj.obj_attributes.present) {
    av.push_back ();
    if (obj.obj.present)
      av[1].fh = obj.obj.handle.addr ();
    av[1].fattr = obj.obj_attributes.attributes.addr ();
  }
  return true;
}

bool
rpc_traverse (attrvec &av, dirlist3 &obj)
{
  /* No need to waste time recursing the chain. */
  return true;
}

bool
rpc_traverse (attrvec &av, entryplus3 &obj)
{
  for (entryplus3 *p = &obj; p; p = p->nextentry)
    if (p->name_attributes.present) {
      attrinfo &x = av.push_back ();
      x.fattr = p->name_attributes.attributes.addr ();
      if (p->name_handle.present)
	x.fh = p->name_handle.handle.addr ();
    }
  return true;
}

bool
rpc_traverse (attrvec &av, rename3wcc &obj)
{
  rpc_traverse (av, obj.fromdir_wcc);
  attrinfo &x = av.push_back ();
  x.fh = &reinterpret_cast<rename3args *> (av[0].fh)->to.dir;
  x.set_wcc (&obj.todir_wcc);
  return true;
}

bool
rpc_traverse (attrvec &av, link3wcc &obj)
{
  rpc_traverse (av, obj.file_attributes);
  attrinfo &x = av.push_back ();
  x.fh = &reinterpret_cast<link3args *> (av[0].fh)->link.dir;
  x.set_wcc (&obj.linkdir_wcc);
  return true;
}

inline bool
nfs_constop (u_int32_t proc)
{
  switch (proc) {
  case NFSPROC3_SETATTR:
  case NFSPROC3_WRITE:
  case NFSPROC3_CREATE:
  case NFSPROC3_MKDIR:
  case NFSPROC3_SYMLINK:
  case NFSPROC3_MKNOD:
  case NFSPROC3_REMOVE:
  case NFSPROC3_RMDIR:
  case NFSPROC3_RENAME:
  case NFSPROC3_LINK:
    return false;
  default:
  case NFSPROC3_COMMIT:		// sic
    return true;
  }
}

void
nfs3_getattrinfo (attrvec *avp, u_int32_t proc, void *argp, void *resp)
{
  if (proc == NFSPROC3_NULL)
    return;
  avp->clear ();
  avp->push_back ().fh = static_cast<nfs_fh3 *> (argp);
  nfs3_traverse_res (*avp, proc, resp);
  if (!(*avp)[0].fattr
      && (nfs_constop (proc) || static_cast<nfsstat3 *> (resp)))
    avp->pop_front ();
}

