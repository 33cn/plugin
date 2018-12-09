/* $Id: nfsxattr.C 1754 2006-05-19 20:59:19Z max $ */

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

#include "nfstrans.h"
#include "nfs3_nonnul.h"

bool
rpc_traverse (xattrvec &xv, ex_lookup3resok &obj)
{
  if (obj.dir_attributes.present)
    xv[0].fattr = obj.dir_attributes.attributes.addr ();
  if (obj.obj_attributes.present) {
    xv.push_back ();
    xv[1].fh = &obj.object;
    xv[1].fattr = obj.obj_attributes.attributes.addr ();
  }
  return true;
}

bool
rpc_traverse (xattrvec &xv, ex_diropres3ok &obj)
{
  xv[0].set_wcc (&obj.dir_wcc);
  if (obj.obj_attributes.present) {
    xv.push_back ();
    if (obj.obj.present)
      xv[1].fh = obj.obj.handle.addr ();
    xv[1].fattr = obj.obj_attributes.attributes.addr ();
  }
  return true;
}

bool
rpc_traverse (xattrvec &xv, dirlist3 &obj)
{
  /* No need to waste time recursing the chain. */
  return true;
}

bool
rpc_traverse (xattrvec &xv, ex_entryplus3 &obj)
{
  for (ex_entryplus3 *p = &obj; p; p = p->nextentry)
    if (p->name_attributes.present) {
      xattr &x = xv.push_back ();
      x.fattr = p->name_attributes.attributes.addr ();
      if (p->name_handle.present)
	x.fh = p->name_handle.handle.addr ();
    }
  return true;
}

bool
rpc_traverse (xattrvec &xv, ex_rename3wcc &obj)
{
  rpc_traverse (xv, obj.fromdir_wcc);
  xattr &x = xv.push_back ();
  x.fh = &reinterpret_cast<rename3args *> (xv[0].fh)->to.dir;
  x.set_wcc (&obj.todir_wcc);
  return true;
}

bool
rpc_traverse (xattrvec &xv, ex_link3wcc &obj)
{
  rpc_traverse (xv, obj.file_attributes);
  xattr &x = xv.push_back ();
  x.fh = &reinterpret_cast<link3args *> (xv[0].fh)->link.dir;
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

#define getxattr(proc, arg, res)			\
  case proc:						\
    rpc_traverse (*xvp, *static_cast<res *> (resp));	\
    break;

void
nfs3_getxattr (xattrvec *xvp, u_int32_t proc, void *argp, void *resp)
{
  xvp->clear ();
  xvp->push_back ().fh = static_cast<nfs_fh3 *> (argp);
  switch (proc) {
    ex_NFS_PROGRAM_3_APPLY_NOVOID (getxattr, nfs3void);
  default:
    panic ("nfs3_getxattr: bad proc %d\n", proc);
    break;
  }
  if (!(*xvp)[0].fattr && nfs_constop (proc)
      && !*static_cast<nfsstat3 *> (resp))
    xvp->pop_front ();
}

/*
 *  Convert to/from the ex_ version of data structures
 */

#define stompit(proc, arg, res)			\
  case proc:					\
    stompcast (*static_cast<res *> (resp));	\
    break;

void
nfs3_exp_enable (u_int32_t proc, void *resp)
{
  switch (proc) {
    ex_NFS_PROGRAM_3_APPLY_NOVOID (stompit, nfs3void);
  default:
    panic ("nfs3_exp_enable: bad proc %d\n", proc);
    break;
  }
}

void
nfs3_exp_disable (u_int32_t proc, void *resp)
{
  switch (proc) {
    NFS_PROGRAM_3_APPLY_NOVOID (stompit, nfs3void);
  default:
    panic ("nfs3_exp_enable: bad proc %d\n", proc);
    break;
  }
}
