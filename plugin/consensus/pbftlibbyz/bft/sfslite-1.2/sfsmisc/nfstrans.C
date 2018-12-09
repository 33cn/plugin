/* $Id: nfstrans.C 3758 2008-11-13 00:36:00Z max $ */

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

#include "nfstrans.h"
#include "nfsserv.h"
#include "nfs3_nonnul.h"
#include "sha1.h"

#if 0
static void
fhhash (u_int64_t *out, void *fhp, size_t len)
{
  assert (len < 64);

  u_char buf[64] = {
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0
  };

  memcpy (buf, fhp, len);
  buf[sizeof (buf) - 1] = len;

  u_int32_t state[5] = {
    0x67452301, 0xefcdab89, 0x98badcfe, 0x10325476, 0xc3d2e1f0,
  };
  sha1::transform (state, buf);
  out[0] = state[0];
  out[1] = state[1];
}
#endif

bool
rpc_traverse (fh3trans &fht, nfs_fh3 &fh)
{
  static const char zeros[7] = {0, 0, 0, 0, 0, 0, 0};

  if (fht.err)
    return false;

  nfs_fh3 res;
  cbc64iv iv (fht.cipher);
  u_int32_t snl, len, srvno;

  switch (fht.mode) {
  case fh3trans::ENCODE:
    srvno = fht.srvno;
    if (fht.pretrans && (fht.err = (*fht.pretrans) (&fh, &srvno)))
      return false;
    if (fh.data.size () > NFS3_FHSIZE - 4) {
      // XXX
      fht.err = NFS3ERR_BADHANDLE;
      return false;
    }
    res.data.setsize ((fh.data.size () + 11) & ~7);
    bzero (res.data.lim () - 7, 7);
    snl = fh.data.size () | (srvno << LOG2_FHSIZE);
    memcpy (res.data.base (), &snl, 4);
    memcpy (res.data.base () + 4, fh.data.base (), fh.data.size ());
    iv.encipher_words (reinterpret_cast<u_int32_t *> (res.data.base ()),
		       res.data.size ());
    fh = res;
    break;
  case fh3trans::DECODE:
    if (fh.data.size () & 7 || !fh.data.size ()) {
      fht.err = NFS3ERR_STALE;
      return false;
    }
    iv.decipher_words (reinterpret_cast<u_int32_t *> (fh.data.base ()),
		       fh.data.size ());
    memcpy (&snl, fh.data.base (), 4);
    len = snl % NFS3_FHSIZE;
    if (fh.data.size () != ((len + 11) & ~7)) {
      fht.err = NFS3ERR_STALE;
      return false;
    }
    if (memcmp (zeros, fh.data.base () + 4 + len,
		fh.data.size () - len - 4)) {
      fht.err = NFS3ERR_STALE;
      return false;
    }

    res.data.setsize (len);
    memcpy (res.data.base (), fh.data.base () + 4, len);
    fh = res;
    srvno = snl >> LOG2_FHSIZE;
    if (fht.pretrans && (fht.err = (*fht.pretrans) (&fh, &srvno)))
      return false;

    if (!fht.srvno_ok) {
      fht.srvno_ok = true;
      fht.srvno = srvno;
    }
    else if (fht.srvno != srvno) {
	fht.err = NFS3ERR_XDEV;
	return false;
    }
    break;
  }
  return true;
}

#define trans(proc, type)					\
  case proc:							\
    if (rpc_traverse (fht, *static_cast<type *> (objp)))	\
      return true;						\
    if (!fht.err)						\
      fht.err = NFS3ERR_INVAL;					\
    return false;
#define transarg(proc, arg, res) trans (proc, arg)
#define transres(proc, arg, res) trans (proc, res)

bool
nfs3_transarg (fh3trans &fht, void *objp, u_int32_t proc)
{
  switch (proc) {
    NFS_PROGRAM_3_APPLY_NONULL (transarg);
  case NFSPROC_CLOSE:
    if (rpc_traverse (fht, *static_cast<nfs_fh3 *>(objp)))
      return true;
    if (!fht.err)
      fht.err = NFS3ERR_INVAL;
    return false;
    break;
  default:
    panic ("nfs3_transarg: bad proc %d\n", proc);
  }
}

bool
nfs3_transres (fh3trans &fht, void *objp, u_int32_t proc)
{
  switch (proc) {
    NFS_PROGRAM_3_APPLY_NONULL (transres);
  default:
    panic ("nfs3_transres: bad proc %d\n", proc);
  }
}

bool
nfs3exp_transarg (fh3trans &fht, void *objp, u_int32_t proc)
{
  switch (proc) {
    ex_NFS_PROGRAM_3_APPLY_NONULL (transarg);
  default:
    panic ("nfs3exp_transarg: bad proc %d\n", proc);
  }
}

bool
nfs3exp_transres (fh3trans &fht, void *objp, u_int32_t proc)
{
  switch (proc) {
    ex_NFS_PROGRAM_3_APPLY_NONULL (transres);
  default:
    panic ("nfs3exp_transres: bad proc %d\n", proc);
  }
}
