/* $Id: nfs3_err.C 1754 2006-05-19 20:59:19Z max $ */

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
#include "nfs3exp_prot.h"
#include "nfs3_nonnul.h"
#include "sfsmisc.h"

#define mkerr(proc, arg, res)			\
case proc:					\
  sbp->replyref (res (status), nocache);	\
  break;

void
nfs3_err (svccb *sbp, nfsstat3 status)
{
  assert (status);
  /* After JUKEBOX errors, FreeBSD resends requests with the same xid. */
  bool nocache = status == NFS3ERR_JUKEBOX;

  switch (sbp->proc ()) {
    NFS_PROGRAM_3_APPLY_NONULL (mkerr);
  default:
    panic ("nfs3_err: invalid proc %d\n", sbp->proc ());
  }
}

void
nfs3exp_err (svccb *sbp, nfsstat3 status)
{
  assert (status);
  /* After JUKEBOX errors, FreeBSD resends requests with the same xid. */
  bool nocache = status == NFS3ERR_JUKEBOX;

  switch (sbp->proc ()) {
    ex_NFS_PROGRAM_3_APPLY_NOVOID (mkerr, nfs3void);
  default:
    panic ("nfs3exp_err: invalid proc %d\n", sbp->proc ());
  }
}

const strbuf &
strbuf_cat (const strbuf &sb, nfsstat3 err)
{
  switch (err) {
  case NFS3_OK:
    return strbuf_cat (sb, "no error", false);
  case NFS3ERR_BADHANDLE:
    return strbuf_cat (sb, "illegal file handle", false);
  case NFS3ERR_NOT_SYNC:
    return strbuf_cat (sb, "setattr synchronization failure", false);
  case NFS3ERR_BAD_COOKIE:
    return strbuf_cat (sb, "stale directory cookie", false);
  case NFS3ERR_NOTSUPP:
    return strbuf_cat (sb, strerror (EOPNOTSUPP), false);
  case NFS3ERR_TOOSMALL:
    return strbuf_cat (sb, "buffer or request too small", false);
  case NFS3ERR_SERVERFAULT:
    return strbuf_cat (sb, "generic server error", false);
  case NFS3ERR_BADTYPE:
    return strbuf_cat (sb, "file type not supported", false);
  case NFS3ERR_FPRINTNOTFOUND:
    return strbuf_cat (sb, "finger print not found", false);
  case NFS3ERR_JUKEBOX:
    return strbuf_cat (sb, "try again later", false);
  default:
    return strbuf_cat (sb, strerror (err));
  }
}
