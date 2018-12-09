// -*-c++-*-
/* $Id: afsnode.h 1754 2006-05-19 20:59:19Z max $ */

/*
 *
 * Copyright (C) 1998-2001 David Mazieres (dm@uun.org)
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

#ifndef _AFSNODE_H_
#define _AFSNODE_H_ 1

#include "arpc.h"
#include "nfs_prot.h"
#include "itree.h"
#include "vec.h"
#include "sfsmisc.h"

#ifndef fattr3
#define fattr3 fattr3exp
#endif /* !fattr3 */

inline bool
operator< (const nfstime3 &a, const nfstime3 &b)
{
  if (a.seconds < b.seconds)
    return true;
  if (a.seconds == b.seconds)
    return a.nseconds < b.nseconds;
  return false;
}

extern void nfs3_err (svccb *sbp, nfsstat3 status);
inline void
nfs_error (svccb *sbp, int err)
{
  if (sbp->vers () != 3) {
    sbp->replyref (nfsstat (err));
    return;
  }
  nfsstat3 stat = nfsstat3 (err);
  if (stat == EINVAL)
    stat = NFS3ERR_INVAL;
  nfs3_err (sbp, stat);
}

class afsnode : public virtual refcount {
public:
  typedef u_int64_t inum_t;

protected:
  u_int nlinks;
  nfstime3 mtime;
  nfstime3 ctime;

  static inum_t geninum (inum_t no = 0);

  afsnode (ftype3, inum_t = 0);
  virtual ~afsnode ();
  void lookup_reply (svccb *sbp, afsnode *e);
  void dirop_reply (svccb *sbp, afsnode *e);

public:
  static bool fhsecret_initialized;
  static char fhsecret[16];
  static sfs_aid (*sbp2aid) (const svccb *);

  ihash_entry<afsnode> itblink;
  const ftype3 type;
  const inum_t ino;

  static void getnfstime (nfstime3 *ntp);
  static sfs_aid sbp2aid_default (const svccb *sbp)
    { return aup2aid (sbp->getaup ()); }

  virtual u_int getnlinks () { return nlinks; }
  virtual void bumpctime () { getnfstime (&ctime); }
  void addlink () { nlinks++; bumpctime (); }
  void remlink () { nlinks--; bumpctime (); }

  const nfstime3 &getmtime () const { return mtime; }
  const nfstime3 &getctime () const { return ctime; }

  virtual void mkfh (nfs_fh *);
  virtual bool chkfh (const nfs_fh *);
  virtual afsnode *nodefh2node (const nfs_fh *) { return this; }
  static afsnode *fh2node (const nfs_fh *);

  void mkfh3 (nfs_fh3 *);
  bool chkfh3 (const nfs_fh *);
  static afsnode *fh3node (const nfs_fh3 *);

  virtual str readlink () const { return NULL; }
  virtual void mkfattr (fattr *, sfs_aid aid);
  virtual void mkfattr3 (fattr3 *, sfs_aid aid);
  void mkpoattr (post_op_attr &poa, sfs_aid aid)
    { poa.set_present (true); mkfattr3 (poa.attributes.addr (), aid); }

  virtual void nfs_getattr (svccb *);
  virtual void nfs_lookup (svccb *, str);
  virtual void nfs_readlink (svccb *);
  virtual void nfs_read (svccb *);
  virtual void nfs_readdir (svccb *);
  virtual void nfs_statfs (svccb *);
  virtual void nfs_remove (svccb *sbp) { nfs_error (sbp, NFSERR_ACCES); }
  virtual void nfs_symlink (svccb *sbp) { nfs_error (sbp, NFSERR_ACCES); }
  virtual void nfs_mkdir (svccb *sbp) { nfs_error (sbp, NFSERR_ACCES); }
  virtual void nfs_rmdir (svccb *sbp) { nfs_error (sbp, NFSERR_ACCES); }
  virtual void nfs_write (svccb *sbp) { nfs_error (sbp, NFSERR_ACCES); }
  virtual void nfs_create (svccb *sbp) { nfs_error (sbp, NFSERR_ACCES); }
  virtual void nfs_setattr (svccb *sbp) { nfs_error (sbp, NFSERR_ACCES); }
  virtual void nfs_rename (svccb *sbp) { nfs_error (sbp, NFSERR_ACCES); }
  virtual void nfs_link (svccb *sbp) { nfs_error (sbp, NFSERR_ACCES); }
  static void dispatch (svccb *);

  virtual void nfs3_access (svccb *);
  virtual void nfs3_fsstat (svccb *);
  virtual void nfs3_fsinfo (svccb *);
  virtual void nfs3_commit (svccb *sbp)
    { sbp->replyref (commit3res (NFS3_OK)); }
  static void dispatch3 (svccb *);
};

class afsreg : public afsnode {
protected:
  str contents;
  afsreg (const str &c)
    : afsnode (NF3REG), contents (c) {}
public:
  str read () const { return contents; }
  void setcontents (const str &s) { contents = s; getnfstime (&mtime); }
  void mkfattr3 (fattr3 *f, sfs_aid aid);
  void nfs_read (svccb *);
  static ref<afsreg> alloc (const str &contents = "")
    { return New refcounted<afsreg> (contents); }
};

class afslink : public afsnode {
  vec<svccb *> sbps;

  bool resok;
  readlinkres res;
  sfs_aid lastaid;

  void reply ();

protected:
  afslink ();
  virtual ~afslink ();
  void sendreply (svccb *sbp);

public:
  void bumpctime () { ctime.seconds++; }
  void mkfattr3 (fattr3 *, sfs_aid aid);
  void setres (nfsstat err);
  void setres (nfspath path);
  str readlink () const { return res.status ? str (NULL) : str (*res.data); }
  bool resset () { return resok; }

  void nfs_readlink (svccb *sbp);

  static ref<afslink> alloc () { return New refcounted<afslink>; }
  static ref<afslink> alloc (nfspath res) {
    ref<afslink> ret = alloc ();
    ret->setres (res);
    return ret;
  }
};

class afsdir;
struct afsdirentry {
  afsdir *const dir;
  const str name;
  const ref<afsnode> node;
  const u_int32_t cookie;

  ihash_entry<afsdirentry> clink;
  itree_entry<afsdirentry> dlink;

  afsdirentry (afsdir *dir, const str &name, afsnode *node);
  ~afsdirentry ();

  static void del (afsdirentry *e) { delete e; }
};

class afsdir : public afsnode {
  friend class afsdirentry;

protected:
  itree<const str, afsdirentry, &afsdirentry::name,
    &afsdirentry::dlink> entries;
  afsdir *const parent;
  afsdir (afsdir *parent);
  virtual ~afsdir ();

  static BOOL xdr (XDR *, void *);

public:
  virtual void bumpmtime () { getnfstime (&mtime); }
  virtual afsnode *lookup (const str &name, sfs_aid aid);

  virtual bool entryok (afsdirentry *, sfs_aid aid);
  virtual afsdirentry *firstentry (sfs_aid aid);
  virtual afsdirentry *nextentry (afsdirentry *, sfs_aid aid);

  virtual bool link (afsnode *node, const str &name);
  virtual bool unlink (const str &name);
  virtual ptr<afsdir> mkdir (const str &name);
  virtual ptr<afslink> symlink (const str &contents, const str &name);

  virtual void nfs_lookup (svccb *, str);
  virtual void nfs_readdir (svccb *sbp)
    { sbp->reply (sbp, &xdr); }

  // XXX - extremely fragile.  If New refcounted<afsdir> appears more
  // than once in any translation unit it causes an internal compiler
  // error in gcc.  Be sure to test with gcc when changing this class.
  static ptr<afsdir> allocsubdir (afsdir *parent)
    { return New refcounted<afsdir> (parent); }
  static ptr<afsdir> alloc () { return allocsubdir (NULL); }

  void ls ();
};

#endif /* !_AFSNODE_H_ */

