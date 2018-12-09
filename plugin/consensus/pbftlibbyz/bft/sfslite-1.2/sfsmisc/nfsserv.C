/* $Id: nfsserv.C 2679 2007-04-04 20:53:20Z max $ */

/*
 *
 * Copyright (C) 2000 David Mazieres (dm@uun.org)
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
#include "sfsmisc.h"
#include "nfstrans.h"
#include "nfs3_nonnul.h"
#include "crypt.h"

static void stalereply (nfscall *nc) { nc->error (NFS3ERR_BADHANDLE); }
const nfsserv::cb_t nfsserv::stalecb (gwrap (stalereply));

/* A simulated close call. */
const rpcgen_table nfscall::closert = {
  "NFSPROC_CLOSE",
  &typeid (nfs_fh3), nfs_fh3_alloc, xdr_nfs_fh3, print_nfs_fh3,
  &typeid (nfsstat3), nfsstat3_alloc, xdr_nfsstat3, print_nfsstat3
};

nfscall::nfscall (const authunix_parms *au, u_int32_t p, void *a)
  : aup (au), procno (p), argp (a), resp (NULL), xdr_res (NULL),
    acstat (SUCCESS), austat (AUTH_OK), rqtime (sfs_get_timenow()), 
    nocache (false),
    nofree (false), stopserv (NULL), curserv (NULL)
{
}

void
nfscall::seterr (nfsstat3 err)
{
  /* After JUKEBOX errors, FreeBSD resends requests using the same xid. */
  nocache = err == NFS3ERR_JUKEBOX;
  clearres ();

#define mkerr(proc, arg, res)			\
case proc:					\
  {						\
    resp = New res (err);			\
    break;					\
  }

  switch (proc ()) {
    NFS_PROGRAM_3_APPLY_NONULL (mkerr);
  case NFSPROC_CLOSE:
    resp = New nfsstat3 (err);
    break;
  default:
    panic ("nfscall::error: invalid proc %d\n", proc ());
  }
#undef mkerr
}

void
nfscall::sendreply ()
{
  if (!curserv || curserv == stopserv)
    delete this;
  else {
    nfsserv *s = curserv;
    curserv = s->nextserv;
    s->getreply (this);
  }
}

void
nfscall::setreply (void *r, sfs::xdrproc_t xdr, bool nc)
{
  if (resp != r) {
    clearres ();
    resp = r;
    nofree = true;
  }
  if (xdr)
    xdr_res = xdr;
  if (nc)
    nocache = nc;
}

void
nfscall::reply (void *r, sfs::xdrproc_t xdr, bool nc)
{
  setreply (r, xdr, nc);
  sendreply ();
}

const rpcgen_table &
nfscall::getrpct () const
{
  u_int32_t pn = proc ();
  if (pn <= NFSPROC3_COMMIT)
    return nfs_program_3.tbl[pn];
  else if (pn == NFSPROC_CLOSE)
    return closert;
  else
    panic ("Bad NFS proc number %d\n", pn);
}

void
nfscall::pinres ()
{
  assert (!xdr_res);		// see nfscall_cb::~nfscall_cb

  if (!nofree)
    return;
  if (proc () == NFSPROC3_NULL) {
    clearres ();
    return;
  }

  void *newres = NULL;
#define mkcopy(proc, arg, res)				\
case proc:						\
  {							\
    newres = New res (*static_cast<res *> (resp));	\
    break;						\
  }

  switch (proc ()) {
    NFS_PROGRAM_3_APPLY_NONULL (mkcopy);
  case NFSPROC_CLOSE:
    resp = New nfsstat3 (*static_cast <nfsstat3 *> (resp));
    break;
  default:
    panic ("nfscall::pinres: invalid proc %d\n", proc ());
  }
#undef mkcopy

  clearres ();
  resp = newres;
}

void
nfscall::clearres ()
{
  if (nofree)
    nofree = false;
  else if (resp)
    xdr_delete (xdr_res ? xdr_res : getrpct ().xdr_res, resp);
  resp = NULL;
  nocache = false;
  xdr_res = NULL;
}

void *
nfscall::getvoidres ()
{
  if (!resp)
    resp = getrpct ().alloc_res ();
  return resp;
}

nfscall_rpc::nfscall_rpc (svccb *s)
  : nfscall (s->getaup (), s->proc (), s->getvoidarg ()),
    sbp (s)
{
}

nfscall_rpc::~nfscall_rpc ()
{
  if (sbp) {
    if (acstat != SUCCESS)
      sbp->reject (acstat);
    else if (austat != AUTH_OK)
      sbp->reject (austat);
    else {
#if 0
      if (proc () == NFSPROC3_LOOKUP) {
	lookup3res *lres = static_cast<lookup3res *> (resp);
	if (lres->status == NFS3_OK)
	  assert (lres->resok->obj_attributes.present);
      }
#endif
      sbp->reply (resp, xdr_res, nocache);
    }
  }
}

nfsserv::nfsserv (ptr<nfsserv> n)
  : cb (stalecb), nextserv (n)
{
  if (nextserv)
    nextserv->setcb (wrap (this, &nfsserv::getcall));
}

bool
nfsserv::encodefh (nfs_fh3 &fh)
{
  return !nextserv || nextserv->encodefh (fh);
}

nfsserv_udp::nfsserv_udp (const rpc_program &program)
{
  fd = inetsocket (SOCK_DGRAM, 0, INADDR_LOOPBACK);
  if (fd < 0)
    fatal ("inetsocket: %m\n");
  x = axprt_dgram::alloc (fd);
  s = asrv::alloc (x, program, wrap (this, &nfsserv_udp::getsbp));
}

nfsserv_udp::nfsserv_udp ()
{
  fd = inetsocket (SOCK_DGRAM, 0, INADDR_LOOPBACK);
  if (fd < 0)
    fatal ("inetsocket: %m\n");
  x = axprt_dgram::alloc (fd);
  s = asrv::alloc (x, nfs_program_3, wrap (this, &nfsserv_udp::getsbp));
}

void
nfsserv_udp::getsbp (svccb *sbp)
{
  if (sbp)
    mkcb (New nfscall_rpc (sbp));
  else
    warn << "nfsserv_udp::getsbp: NULL sbp\n";
}

void
nfsserv_fixup::getattr (nfscall *nc, nfs_fh3 *fhp, getattr3res *res)
{
  delete fhp;
  if (res->status)
    nc->error (res->status);
  else {
    lookup3res *lres = static_cast<lookup3res *> (nc->resp);
    lres->resok->obj_attributes.set_present (true);
    *lres->resok->obj_attributes.attributes = *res->attributes;
    nc->sendreply ();
  }
}

void
nfsserv_fixup::getreply (nfscall *nc)
{
  /* After JUKEBOX errors, FreeBSD resends requests using the same xid. */
  if (nc->proc () != NFSPROC3_NULL
      && *nc->Xtmpl getres<nfsstat3> () == NFS3ERR_JUKEBOX)
    nc->nocache = true;

  /* Many NFS3 clients flip out if lookups replies don't have attributes */
  lookup3res *lres = static_cast<lookup3res *> (nc->resp);
  if (nc->proc () == NFSPROC3_LOOKUP && lres->status == NFS3_OK
      && !lres->resok->obj_attributes.present) {
    nc->pinres ();
    nfs_fh3 *fhp = New nfs_fh3 (lres->resok->object);
    vNew nfscall_cb<NFSPROC3_GETATTR> (nc->aup, fhp,
				       wrap (this, &nfsserv_fixup::getattr,
					     nc, fhp),
				       this);
    return;
  }

  nc->sendreply ();
}

nfsdemux::nfsserv_cryptfh::nfsserv_cryptfh (const ref<nfsdemux> &dd,
					    u_int32_t s)
  : nfsserv (dd->ns), d (dd), srvno (s)
{
  d->srvnotab.insert (this);
  d->ns->setcb (wrap (d.get (), &nfsdemux::getcall)); // XXX - gross
}

nfsdemux::nfsserv_cryptfh::~nfsserv_cryptfh ()
{
  d->srvnotab.remove (this);
}

void
nfsdemux::nfsserv_cryptfh::getreply (nfscall *nc)
{
  if (!nc->xdr_res) {
    fh3trans fht (fh3trans::ENCODE, d->fhkey, srvno);
    if (nc->proc () != NFSPROC_CLOSE &&
	!nfs3_transres (fht, nc->getvoidres (), nc->proc ())) {
      warn ("Cannot encrypt file handles (err %d)\n", fht.err);
      nc->seterr (nfsstat3 (fht.err));
    }
  }
  nc->sendreply ();
}

bool
nfsdemux::nfsserv_cryptfh::encodefh (nfs_fh3 &fh)
{
  fh3trans fht (fh3trans::ENCODE, d->fhkey, srvno);
  if (!rpc_traverse (fht, fh)) {
    warn ("file handle too large\n");
    return false;
  }
  return nextserv->encodefh (fh);
}

nfsdemux::nfsdemux (const ref<nfsserv> &n)
  : ns (n), srvnoctr (0)
{
  char fhkeydat[53];
  rnd.getbytes (fhkeydat, sizeof (fhkeydat));
  fhkey.setkey (fhkeydat, sizeof (fhkeydat));
  bzero (fhkeydat, sizeof (fhkeydat));

  ns->setcb (wrap (this, &nfsdemux::getcall));
}

void
nfsdemux::getcall (nfscall *nc)
{
  if (nc->proc () == NFSPROC3_NULL) {
    nc->reply (NULL);
    return;
  }

  fh3trans fht (fh3trans::DECODE, fhkey);
  nfsserv_cryptfh *nsc;
  if (!nfs3_transarg (fht, nc->getvoidarg (), nc->proc ())
      || !(nsc = srvnotab[fht.srvno]) || !nsc->cb) {
    nc->error (NFS3ERR_BADHANDLE);
    return;
  }

  nsc->getcall (nc);
}

ref<nfsdemux::nfsserv_cryptfh>
nfsdemux::servalloc ()
{
  while (srvnotab[++srvnoctr])
    ;
  return New refcounted<nfsserv_cryptfh> (mkref (this), srvnoctr);
}
