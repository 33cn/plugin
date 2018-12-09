// -*-c++-*-
/* $Id: nfstrans.h 1754 2006-05-19 20:59:19Z max $ */

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

#ifndef _SFSMISC_NFSTRANS_H_
#define _SFSMISC_NFSTRANS_H_ 1

#include "nfs3exp_prot.h"
#include "blowfish.h"

template<>
struct hashfn<nfs_fh3> {
  hash_t operator() (const nfs_fh3 &fh) const
    { return hash_bytes (fh.data.base (), fh.data.size ()); }
};

template<>
struct equals<nfs_fh3> {
  bool operator() (const nfs_fh3 &a, const nfs_fh3 &b) const {
    return a.data.size () == b.data.size () && 
      !memcmp (a.data.base (), b.data.base (), a.data.size ());
  }
};

#define LOG2_FHSIZE 6

struct fh3trans {
  enum { srvno_mask = 0x03ffffff };

  enum mode_t { ENCODE, DECODE };
  typedef callback<int, nfs_fh3 *, u_int32_t *>::ptr pretrans_t;
  typedef callback<int, fattr3 *, u_int32_t>::ptr fattr_hook_t;

  const mode_t mode;
  const block64cipher &cipher;
  u_int32_t srvno;
  bool srvno_ok;
  int err;
  const pretrans_t pretrans;
  fattr_hook_t fattr_hook;

  fh3trans (mode_t m, const block64cipher &bc, pretrans_t pt = NULL)
    : mode (m), cipher (bc), srvno_ok (false), err (0), pretrans (pt)
    { assert (mode != ENCODE); }
  fh3trans (mode_t m, const block64cipher &bc,
	    u_int32_t sn, pretrans_t pt = NULL)
    : mode (m), cipher (bc), srvno (sn), srvno_ok (true),
      err (0), pretrans (pt)
    { assert (srvno < (u_int32_t) -1 >> LOG2_FHSIZE); }
};

DUMBTRAVERSE (fh3trans)

inline bool
rpc_traverse (fh3trans &fht, fattr3 &attr)
{
  return !fht.fattr_hook
    || !(fht.err = (*fht.fattr_hook) (&attr, fht.srvno));
}
inline bool
rpc_traverse (fh3trans &fht, ex_fattr3 &attr)
{
  return !fht.fattr_hook
    || !(fht.err = (*fht.fattr_hook) (reinterpret_cast<fattr3 *> (&attr),
				      fht.srvno));
}

bool rpc_traverse (fh3trans &fht, nfs_fh3 &fh);
bool nfs3_transarg (fh3trans &fht, void *objp, u_int32_t proc);
bool nfs3_transres (fh3trans &fht, void *objp, u_int32_t proc);
bool nfs3exp_transarg (fh3trans &fht, void *objp, u_int32_t proc);
bool nfs3exp_transres (fh3trans &fht, void *objp, u_int32_t proc);


/* nfs3attr.C */
/*
 * nfs3_getattrinfo returns a vector of all the file attributes in an
 * NFS3 call result.  It groups file attributes (fattr3 structures)
 * with file handles and associated pre-operation attributes
 * (wcc_attr), if they are present.
 *
 * Note that if optional attributes are not present in an NFS3 reply,
 * the fattr field will be NULL.  If you wish to clear the
 * pre-operation attributes (for instance because you need to crrect
 * them based on the post-operation attributes, and fattr is NULL),
 * you can do this with:  wdata->before.set_present (false);
 */
struct attrinfo {
  nfs_fh3 *fh;
  fattr3exp *fattr;
  wcc_attr *wattr;
  wcc_data *wdata;

  attrinfo ();
  void set_wcc (wcc_data *wd);
};
typedef vec<attrinfo, 2> attrvec;
void nfs3_getattrinfo (attrvec *avp, u_int32_t proc, void *argp, void *resp);


/* nfsxattr.C -- same as above but for ex_ versions. */
struct xattr {
  nfs_fh3 *fh;
  ex_fattr3 *fattr;
  wcc_attr *wattr;
  ex_wcc_data *wdata;
  void set_wcc (ex_wcc_data *wd) {
    wdata = wd;
    if (wd->before.present)
      wattr = wd->before.attributes;
    if (wd->after.present)
      fattr = wd->after.attributes;
  }

  xattr () : fh (NULL), fattr (NULL), wattr (NULL), wdata (NULL) {}
};
typedef vec<xattr, 2> xattrvec;

DUMBTRAVERSE (xattrvec)

inline bool
rpc_traverse (xattrvec &xv, ex_fattr3 &obj)
{
  assert (!xv[0].fattr);
  xv[0].fattr = &obj;
  return true;
}

inline bool
rpc_traverse (xattrvec &xv, ex_wcc_data &obj)
{
  assert (!xv[0].wdata);
  xv[0].set_wcc (&obj);
  return true;
}

bool rpc_traverse (xattrvec &xv, ex_lookup3resok &obj);
bool rpc_traverse (xattrvec &xv, ex_diropres3ok &obj);
bool rpc_traverse (xattrvec &xv, dirlist3 &obj);
bool rpc_traverse (xattrvec &xv, ex_entryplus3 &obj);
bool rpc_traverse (xattrvec &xv, ex_rename3wcc &obj);
bool rpc_traverse (xattrvec &xv, ex_link3wcc &obj);
void nfs3_getxattr (xattrvec *xvp, u_int32_t proc, void *argp, void *resp);
void nfs3_exp_enable (u_int32_t proc, void *resp);
void nfs3_exp_disable (u_int32_t proc, void *resp);


inline bool
operator== (const nfstime3 &a, const nfstime3 &b)
{
  return a.seconds == b.seconds && a.nseconds == b.nseconds;
}
inline bool
operator!= (const nfstime3 &a, const nfstime3 &b)
{
  return a.seconds != b.seconds || a.nseconds != b.nseconds;
}

#endif /* _SFSMISC_NFSTRANS_H_ */
