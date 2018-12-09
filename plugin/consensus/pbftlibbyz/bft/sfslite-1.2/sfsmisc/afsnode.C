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

#include "afsnode.h"
#include "crypt.h"
#ifdef HAVE_SYS_MKDEV_H
#include <sys/mkdev.h>
#endif /* HAVE_SYS_MKDEV_H */

const afsnode::inum_t maxino = INT64 (0x8000000000000000);
bool afsnode::fhsecret_initialized;
char afsnode::fhsecret[16];
sfs_aid (*afsnode::sbp2aid) (const svccb *) = afsnode::sbp2aid_default;

typedef ihash<const afsnode::inum_t, afsnode,
  &afsnode::ino, &afsnode::itblink> inotbl_t;
static inotbl_t &inotbl = *New inotbl_t;

struct stalenode_t : public afsnode {
  stalenode_t () : afsnode (ftype3 (0)) {}

  void bumpctime () { panic ("stalenode_t::bumpctime\n"); }

  void nfs_getattr (svccb *sbp)
    { nfs_error (sbp, NFSERR_STALE); }
  void nfs_lookup (svccb *sbp, str)
    { nfs_error (sbp, NFSERR_STALE); }
  void nfs_readlink (svccb *sbp)
    { nfs_error (sbp, NFSERR_STALE); }
  void nfs_read (svccb *sbp)
    { nfs_error (sbp, NFSERR_STALE); }
  void nfs_readdir (svccb *sbp)
    { nfs_error (sbp, NFSERR_STALE); }
  void nfs_statfs (svccb *sbp)
    { nfs_error (sbp, NFSERR_STALE); }

  void nfs3_access (svccb *sbp)
    { nfs3_err (sbp, NFS3ERR_STALE); }
  void nfs3_fsstat (svccb *sbp)
    { nfs3_err (sbp, NFS3ERR_STALE); }
  void nfs3_fsinfo (svccb *sbp)
    { nfs3_err (sbp, NFS3ERR_STALE); }
};

void
afsnode::getnfstime (nfstime3 *ntp)
{
  timespec ts;
  clock_gettime (CLOCK_REALTIME, &ts);
  ntp->seconds = ts.tv_sec;
  ntp->nseconds = ts.tv_nsec;
}

afsnode::inum_t
afsnode::geninum (afsnode::inum_t no)
{
  static inum_t inoctr = 1;

  if (no) {
    if (inotbl[no])
      warn ("mkinum: inum 0x%" U64F "x already in use\n", no);
    else if (no < 2 || no >= maxino)
      warn ("mkinum: invalid inum 0x%" U64F "x requested\n", no);
    else
      return no;
  }

  while (inotbl[++inoctr])
    if (inoctr >= maxino)
      inoctr = 1;
  return inoctr;
}

static void
pathconf3 (svccb *sbp)
{
  pathconf3res res (NFS3_OK);
  res.resok->linkmax = 0xffff;
  res.resok->name_max = 0xff;
  res.resok->no_trunc = false;
  res.resok->chown_restricted = true;
  res.resok->case_insensitive = false;
  res.resok->case_preserving = true;
  sbp->reply (&res);
}

afsnode::afsnode (ftype3 type, inum_t ino)
  : nlinks (0), type (type), ino (geninum (ino))
{
  getnfstime (&ctime);
  mtime = ctime;
  inotbl.insert (this);
}

afsnode::~afsnode ()
{
  inotbl.remove (this);
}

void
afsnode::mkfh (nfs_fh *fhp)
{
  if (!fhsecret_initialized) {
    rnd.getbytes (fhsecret, sizeof (fhsecret));
    fhsecret_initialized = true;
  }
  bzero (fhp->data.base (), fhp->data.size ());
  puthyper (fhp->data.base (), ino);
  memcpy (fhp->data.base () + 8, fhsecret, sizeof (fhsecret));
}

bool
afsnode::chkfh (const nfs_fh *fhp)
{
  if (gethyper (fhp->data.base ()) != ino
      || memcmp (fhp->data.base () + 8, fhsecret, sizeof (fhsecret)))
    return false;
#if 0
  for (const char *s = fhp->data.base () + 8 + sizeof (fhsecret),
	 *e = fhp->data.lim (); s < e; s++)
    if (*s)
      return false;
#endif
  return true;
}

afsnode *
afsnode::fh2node (const nfs_fh *fhp)
{
  afsnode *n = inotbl[gethyper (fhp->data.base ())];
  if (n && (n = n->nodefh2node (fhp)) && n->chkfh (fhp))
    return n;
  return NULL;
}

void
afsnode::mkfh3 (nfs_fh3 *fhp)
{
  fhp->data.setsize (NFS_FHSIZE);
  mkfh (reinterpret_cast<nfs_fh *> (fhp->data.base ()));
}

bool
afsnode::chkfh3 (const nfs_fh *fhp)
{
  return fhp->data.size () == NFS_FHSIZE
    && chkfh (reinterpret_cast<const nfs_fh *> (fhp->data.base ()));
}

afsnode *
afsnode::fh3node (const nfs_fh3 *fhp)
{
  return fhp->data.size () != NFS_FHSIZE ? NULL
    : fh2node (reinterpret_cast<const nfs_fh *> (fhp->data.base ()));
}

void
afsnode::mkfattr3 (fattr3 *f, sfs_aid)
{
  bzero (f, sizeof (*f));
  f->type = type;
  f->mode = type == NF3DIR ? 0555 : 0444;
  f->nlink = getnlinks ();
  f->gid = sfs_gid;
  f->size = 512;
  f->used = 512;
  f->fileid = ino;
  getnfstime (&f->atime);
  f->mtime = mtime;
  f->ctime = ctime;
}

void
afsnode::mkfattr (fattr *f, sfs_aid aid)
{
  static const int modes[] = {
    0, NFSMODE_REG, NFSMODE_DIR, NFSMODE_BLK, NFSMODE_CHR,
    NFSMODE_LNK, NFSMODE_SOCK, NFSMODE_FIFO
  };
  static const int nmodes = sizeof (modes) / sizeof (modes[0]);

  fattr3 f3;
  mkfattr3 (&f3, aid);
  bzero (f, sizeof (*f));
  f->type = f3.type == NF3FIFO ? NFFIFO : ftype (f3.type);
  f->mode = f3.mode;
  if (f3.type < nmodes)
    f->mode |= modes[f3.type];
  f->nlink = f3.nlink;
  f->uid = f3.uid;
  f->gid = f3.gid;
  f->size = f3.size;
  f->blocksize = 512;
#ifndef __linux__		// XXX
  f->rdev = makedev (f3.rdev.major, f3.rdev.minor);
#endif /* !__linux__ */
  f->blocks = f3.used >> 9;
  f->fsid = f3.fsid;
  f->fileid = f3.fileid;
  f->atime.seconds = f3.atime.seconds;
  f->atime.useconds = f3.atime.nseconds / 1000;
  f->mtime.seconds = f3.mtime.seconds;
  f->mtime.useconds = f3.mtime.nseconds / 1000;
  f->ctime.seconds = f3.ctime.seconds;
  f->ctime.useconds = f3.ctime.nseconds / 1000;
}

void
afsnode::lookup_reply (svccb *sbp, afsnode *e)
{
  const sfs_aid aid = sbp2aid (sbp);
  if (sbp->vers () == 2) {
    diropres res (NFS_OK);
    if (e) {
      e->mkfh (&res.reply->file);
      e->mkfattr (&res.reply->attributes, aid);
    }
    else
      res.set_status (NFSERR_NOENT);
    sbp->reply (&res);
  }
  else {
    lookup3res res (NFS3_OK);
    if (e) {
      e->mkfh3 (&res.resok->object);
      e->mkpoattr (res.resok->obj_attributes, aid);
      mkpoattr (res.resok->dir_attributes, aid);
    }
    else {
      res.set_status (NFS3ERR_NOENT);
      mkpoattr (*res.resfail, aid);
    }
    sbp->reply (&res);
  }
}

void
afsnode::dirop_reply (svccb *sbp, afsnode *e)
{
  const sfs_aid aid = sbp2aid (sbp);
  if (sbp->vers () == 2)
    lookup_reply (sbp, e);
  else {
    diropres3 res (NFS3_OK);
    if (e) {
      res.resok->obj.set_present (true);
      e->mkfh3 (res.resok->obj.handle.addr ());
      e->mkpoattr (res.resok->obj_attributes, aid);
      mkpoattr (res.resok->dir_wcc.after, aid);
    }
    else {
      res.set_status (NFS3ERR_NOENT);
      mkpoattr (res.resfail->after, aid);
    }
    sbp->replyref (res);
  }
}

void
afsnode::nfs_getattr (svccb *sbp)
{
  const sfs_aid aid = sbp2aid (sbp);
  if (sbp->vers () == 2) {
    attrstat res (NFS_OK);
    mkfattr (res.attributes.addr (), aid);
    sbp->reply (&res);
  }
  else {
    getattr3res res (NFS3_OK);
    mkfattr3 (res.attributes.addr (), aid);
    sbp->reply (&res);
  }
}

void
afsnode::nfs_lookup (svccb *sbp, str)
{
  nfs_error (sbp, NFSERR_NOTDIR);
}

void
afsnode::nfs_readlink (svccb *sbp)
{
  nfs_error (sbp, nfsstat (EINVAL));
}

void
afsnode::nfs_read (svccb *sbp)
{
  nfs_error (sbp, nfsstat (EINVAL));
}

void
afsnode::nfs_readdir (svccb *sbp)
{
  nfs_error (sbp, NFSERR_NOTDIR);
}

void
afsnode::nfs_statfs (svccb *sbp)
{
  statfsres res;
  res.set_status (NFS_OK);
  res.reply->tsize = 8192;
  res.reply->bsize = 512;
  res.reply->blocks = 0;
  res.reply->bfree = 0;
  res.reply->bavail = 0;
  sbp->reply (&res);
}

void
afsnode::nfs3_access (svccb *sbp)
{
  access3res res (NFS3_OK);
  mkpoattr (res.resok->obj_attributes, sbp2aid (sbp));
  res.resok->access = ((ACCESS3_READ | ACCESS3_LOOKUP | ACCESS3_EXECUTE)
		       & sbp->Xtmpl getarg<access3args> ()->access);
  sbp->reply (&res);
}

void
afsnode::nfs3_fsstat (svccb *sbp)
{
  fsstat3res res (NFS3_OK);
  rpc_clear (res);
  sbp->reply (&res);
}

void
afsnode::nfs3_fsinfo (svccb *sbp)
{
  fsinfo3res res (NFS3_OK);
  res.resok->rtmax = 8192;
  res.resok->rtpref = 8192;
  res.resok->rtmult = 512;
  res.resok->wtmax = 8192;
  res.resok->wtpref = 8192;
  res.resok->wtmult = 8192;
  res.resok->dtpref = 8192;
  res.resok->maxfilesize = INT64 (0x7fffffffffffffff);
  res.resok->time_delta.seconds = 0;
  res.resok->time_delta.nseconds = 1;
  res.resok->properties = (FSF3_LINK | FSF3_SYMLINK | FSF3_HOMOGENEOUS
			   | FSF3_CANSETTIME);
  sbp->reply (&res);
}

static afsnode *
sbp2node (svccb *sbp)
{
  static ref<stalenode_t> stalenode = New refcounted<stalenode_t>;
  switch (sbp->vers ()) {
  case 2:
    if (afsnode *a = afsnode::fh2node (sbp->Xtmpl getarg<nfs_fh> ()))
      return a;
    break;
  case 3:
    if (afsnode *a = afsnode::fh3node (sbp->Xtmpl getarg<nfs_fh3> ()))
      return a;
    break;
  }
  return stalenode;
}

void
afsnode::dispatch (svccb *sbp)
{
  switch (sbp->proc ()) {
  case NFSPROC_NULL:
    sbp->reply (NULL);
    break;
  case NFSPROC_GETATTR:
    sbp2node (sbp)->nfs_getattr (sbp);
    break;
  case NFSPROC_LOOKUP:
    sbp2node (sbp)->nfs_lookup (sbp, sbp->Xtmpl getarg<diropargs> ()->name);
    break;
  case NFSPROC_READLINK:
    sbp2node (sbp)->nfs_readlink (sbp);
    break;
  case NFSPROC_READ:
    sbp2node (sbp)->nfs_read (sbp);
    break;
  case NFSPROC_READDIR:
    sbp2node (sbp)->nfs_readdir (sbp);
    break;
  case NFSPROC_STATFS:
    sbp2node (sbp)->nfs_statfs (sbp);
    break;
  case NFSPROC_REMOVE:
    sbp2node (sbp)->nfs_remove (sbp);
    break;
  case NFSPROC_SYMLINK:
    sbp2node (sbp)->nfs_symlink (sbp);
    break;
  case NFSPROC_RMDIR:
    sbp2node (sbp)->nfs_rmdir (sbp);
    break;
  case NFSPROC_MKDIR:
    sbp2node (sbp)->nfs_rmdir (sbp);
    break;
  case NFSPROC_SETATTR:
    sbp2node (sbp)->nfs_rmdir (sbp);
    break;
  case NFSPROC_WRITE:
    sbp2node (sbp)->nfs_rmdir (sbp);
    break;
  case NFSPROC_CREATE:
    sbp2node (sbp)->nfs_create (sbp);
    break;
  case NFSPROC_RENAME:
    sbp2node (sbp)->nfs_rename (sbp);
    break;
  case NFSPROC_LINK:
    sbp2node (sbp)->nfs_link (sbp);
    break;
  case NFSPROC_WRITECACHE:
    nfs_error (sbp, NFSERR_ACCES);
    break;
  case NFSPROC_ROOT:
    nfs_error (sbp, NFSERR_PERM);
    break;
  default:
    nfs_error (sbp, nfsstat (EINVAL));
    break;
  }
}

void
afsnode::dispatch3 (svccb *sbp)
{
  switch (sbp->proc ()) {
  case NFSPROC3_NULL:
    sbp->reply (NULL);
    break;
  case NFSPROC3_GETATTR:
    sbp2node (sbp)->nfs_getattr (sbp);
    break;
  case NFSPROC3_LOOKUP:
    sbp2node (sbp)->nfs_lookup (sbp,
				sbp->Xtmpl getarg<diropargs3> ()->name);
    break;
  case NFSPROC3_ACCESS:
    sbp2node (sbp)->nfs3_access (sbp);
    break;
  case NFSPROC3_READLINK:
    sbp2node (sbp)->nfs_readlink (sbp);
    break;
  case NFSPROC3_READ:
    sbp2node (sbp)->nfs_read (sbp);
    break;
  case NFSPROC3_SYMLINK:
    sbp2node (sbp)->nfs_symlink (sbp);
    break;
  case NFSPROC3_REMOVE:
    sbp2node (sbp)->nfs_remove (sbp);
    break;
  case NFSPROC3_RMDIR:
    sbp2node (sbp)->nfs_rmdir (sbp);
    break;
  case NFSPROC3_MKDIR:
    sbp2node (sbp)->nfs_mkdir (sbp);
    break;
  case NFSPROC3_SETATTR:
    sbp2node (sbp)->nfs_setattr (sbp);
    break;
  case NFSPROC3_WRITE:
    sbp2node (sbp)->nfs_write (sbp);
    break;
  case NFSPROC3_CREATE:
    sbp2node (sbp)->nfs_create (sbp);
    break;
  case NFSPROC3_MKNOD:
    nfs3_err (sbp, NFS3ERR_ACCES);
    break;
  case NFSPROC3_RENAME:
    sbp2node (sbp)->nfs_rename (sbp);
    break;
  case NFSPROC3_LINK:
    sbp2node (sbp)->nfs_link (sbp);
    break;
  case NFSPROC3_COMMIT:
    sbp2node (sbp)->nfs3_commit (sbp);
    break;
  case NFSPROC3_READDIR:
  case NFSPROC3_READDIRPLUS:
    sbp2node (sbp)->nfs_readdir (sbp);
    break;
  case NFSPROC3_FSSTAT:
    sbp2node (sbp)->nfs3_fsstat (sbp);
    break;
  case NFSPROC3_FSINFO:
    sbp2node (sbp)->nfs3_fsinfo (sbp);
    break;
  case NFSPROC3_PATHCONF:
    pathconf3 (sbp);
    break;
  }
}

void
afsreg::mkfattr3 (fattr3 *f, sfs_aid aid)
{
  afsnode::mkfattr3 (f, aid);
  f->size = contents.len ();
}

void
afsreg::nfs_read (svccb *sbp)
{
  if (sbp->vers () == 3) {
    read3args *arg = sbp->Xtmpl getarg<read3args> ();
    read3res res (NFS3_OK);
    res.resok->eof = arg->offset + arg->count >= contents.len ();
    if (arg->offset >= contents.len ())
      res.resok->count = 0;
    else {
      res.resok->count = min<u_int64_t> (arg->count,
					 contents.len () - arg->offset);
      res.resok->data.setsize (res.resok->count);
      memcpy (res.resok->data.base (),
	      contents.cstr () + arg->offset, res.resok->count);
    }
    mkpoattr (res.resok->file_attributes, sbp2aid (sbp));
    sbp->replyref (res);
  }
  else if (sbp->vers () == 2) {
    readargs *arg = sbp->Xtmpl getarg<readargs> ();
    readres res (NFS_OK);
    if (arg->offset < contents.len ()) {
      res.reply->data.setsize (min<u_int32_t> (arg->count,
					       contents.len () - arg->offset));
      memcpy (res.reply->data.base (), contents.cstr () + arg->offset,
	      res.reply->data.size ());
    }
    mkfattr (&res.reply->attributes, sbp2aid (sbp));
    sbp->replyref (res);
  }
}

void
afslink::mkfattr3 (fattr3 *f, sfs_aid aid)
{
  /* BSD needs the seconds (not just milliseconds/nanoseconds) of the
   * ctime to change on every lookup/getattr in order to defeat the
   * name cache. */
  if (aid != lastaid) {
    lastaid = aid;
    bumpctime ();
  }

  afsnode::mkfattr3 (f, aid);
  if (!resok)
    f->size = NFS_MAXPATHLEN;
  else if (!res.status)
    f->size = res.data->len ();
  else
    f->size = 0;
}

void
afslink::sendreply (svccb *sbp)
{
  if (sbp->vers () == 2)
    sbp->reply (&res);
  else if (res.status)
    nfs_error (sbp, res.status);
  else {
    readlink3res res3 (NFS3_OK);
    res3.resok->data = *res.data;
    mkpoattr (res3.resok->symlink_attributes, sbp2aid (sbp));
    sbp->reply (&res3);
  }
}

void
afslink::reply ()
{
  resok = true;
  if (!sbps.empty ()) {
    do {
      sendreply (sbps.pop_front ());
    } while (!sbps.empty ());
    sbps.clear ();
  }
}

afslink::afslink ()
  : afsnode (NF3LNK), resok (false), lastaid (0)
{
  ctime.seconds = ctime.nseconds = 0;
  res.set_status (NFSERR_STALE);
}

afslink::~afslink ()
{
  reply ();
}

void
afslink::setres (nfsstat err)
{
  assert (err);
#ifdef __linux__
  /* XXX -- linux ignores the return status of a readlink RPC */
  res.set_status (NFS_OK);
  *res.data = strbuf (":: ") << strerror (err);
#else /* !linux */
    res.set_status (err);
#endif /* !linux */
  reply ();
}

void
afslink::setres (nfspath path)
{
  res.set_status (NFS_OK);
  *res.data = path;
  reply ();
}

void
afslink::nfs_readlink (svccb *sbp)
{
  if (resok)
    sendreply (sbp);
  else
    sbps.push_back (sbp);
}

str
au2str (const authunix_parms *aup)
{
  if (!aup)
    return str ();

  time_t at = aup->aup_time;
  char buf[80];
  if (!strftime (buf, sizeof (buf), "%Y/%m/%d %T", localtime (&at)))
    panic ("strftime overflow\n");

  strbuf b ("AUP: %s (%s) uid=%d gid=%d groups={", aup->aup_machname, buf,
	    implicit_cast<int> (aup->aup_uid),
	    implicit_cast<int> (aup->aup_gid));
  for (u_int i = 0; i < aup->aup_len; i++)
    if (i)
      b.fmt (", %d", implicit_cast<int> (aup->aup_gids[i]));
    else
      b.fmt ("%d", implicit_cast<int> (aup->aup_gids[i]));
  b.fmt ("}");

  return b;
}

