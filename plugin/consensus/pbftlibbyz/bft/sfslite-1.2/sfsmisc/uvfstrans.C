/* $Id: uvfstrans.C 1754 2006-05-19 20:59:19Z max $ */

/*
 *
 * Copyright (C) 1999 Michael Kaminsky (kaminsky@lcs.mit.edu)
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

#ifdef USE_UVFS

#include "uvfstrans.h"
#include "nfs3_nonnul.h"

bool
uvfsstate::translate (uvfstrans &fht, fh_t &obj)
{
  switch (fht.dir) {
  case uvfstrans::TOUVFS:
    {
      fhpair *fhp = ntou_tab (obj, fht.srvno);
      if (!fhp) {
	fhp = New fhpair;
	fhp->srvno = fht.srvno;
	fhp->nfh = obj;
	fhp->ufh = obj = nextfh ();
	uton_tab.insert (fhp);
	ntou_tab.insert (fhp);
	return true;
      }
      obj = fhp->ufh;
      return true;
      break;
    }
  case uvfstrans::FROMUVFS:
    {
      fhpair *fhp = uton_tab[obj];
      if (!fhp) {
	fht.error = ESTALE;
	return false;
      }
      if (fht.srvno_valid && fht.srvno != fhp->srvno) {
	fht.error = EXDEV;
	return false;
      }
      fht.srvno_valid = true;
      fht.srvno = fhp->srvno;
      obj = fhp->nfh;
      return true;
      break;
    }
  default:
    panic ("Bad uvfs translation direction\n");
  }
}

void
uvfsstate::reclaim (fh_t &obj)
{
  fhpair *fhp = uton_tab[obj];
  uton_tab.remove (fhp);
  ntou_tab.remove (fhp);
}

struct uvfsvoid {
  uvfsvoid () {}
  template<class T> uvfsvoid (T) {}
};

template<class T> inline bool
rpc_traverse (T &, uvfsvoid &)
{
  return true;
}

#define UVFSPROG_1_APPLY_NONULL(macro)	\
  UVFSPROG_1_APPLY_NOVOID (macro, uvfsvoid)

#define trans(type)						\
    if (rpc_traverse (fht, *static_cast<type *> (objp)))	\
      ret = true;						\
    else if (!fht.error)					\
      fht.error = EINVAL;
#define transarg(proc, arg, res)		\
  case proc:					\
    trans (arg);				\
    break;
#define stomparg(proc, arg, res)		\
  case proc:					\
    stompcast (*static_cast<arg *> (objp));	\
    break;
#define transres(proc, arg, res)		\
  case proc:					\
    stompcast (*static_cast<res *> (objp));	\
    trans (res);				\
    break;

bool
uvfs_transarg (uvfstrans &fht, void *objp, u_int32_t proc)
{
  bool ret = false;
  switch (proc) {
    UVFSPROG_1_APPLY_NONULL (transarg);
  default:
    panic ("uvfs_transarg: bad proc %d\n", proc);
  }
  switch (proc) {
    NFS_PROGRAM_3_APPLY_NONULL(stomparg);
  }
  return ret;
}

bool
uvfs_transres (uvfstrans &fht, void *objp, u_int32_t proc)
{
  bool ret = false;
  switch (proc) {
    UVFSPROG_1_APPLY_NONULL (transres);
  default:
    panic ("uvfs_transres: bad proc %d\n", proc);
  }
  return ret;
}

int
uvfs_open ()
{
  int i = 0;

  for (;;) {
    int fd = open (str (strbuf ("/dev/uvfs%d", i)), O_RDWR);
    if (fd >= 0 || errno == ENOENT)
      return fd;
  }
}

#define mkerr(proc, arg, res)			\
case proc:					\
  sbp->replyref (res (nfsstat3 (status)));	\
  break;

void
uvfs_err (svccb *sbp, uvfsstat status)
{
  assert (status);

  switch (sbp->proc ()) {
    UVFSPROG_1_APPLY_NONULL (mkerr);
  default:
    panic ("uvfs_err: invalid proc %d\n", sbp->proc ());
  }
}

#endif /* USE_UVFS */
