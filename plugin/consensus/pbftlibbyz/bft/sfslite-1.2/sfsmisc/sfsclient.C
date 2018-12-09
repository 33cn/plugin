/* $Id: sfsclient.C 2679 2007-04-04 20:53:20Z max $ */

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

#include "sfsclient.h"
#include "axprt_crypt.h"
#include "sfscd_prot.h"
#include "nfstrans.h"
#include "crypt.h"
#include "nfs3close_prot.h"

#define SFSPREF ".SFS \177"
#define SFSFH SFSPREF "FH"

time_t mount_idletime = 300;

static int
getportno (int fd)
{
  int port = -1;
  struct sockaddr_in sin;
  socklen_t sinlen = sizeof (sin);
  bzero (&sin, sizeof (sin));
  if (!getpeername (fd, (sockaddr *) &sin, &sinlen))
    port = ntohs (sin.sin_port);
  return port;
}

sfsserver::sfsserver (const sfsserverargs &a)
  : lock_flag (true), condemn_flag (false), destroyed (false),
    recon_backoff (1), recon_tmo (NULL), ns (a.ns), prog (*a.p),
    carg (a.ma->carg), fsinfo (NULL), lastuse (sfs_get_timenow())
{
  si = sfs_servinfo_w::alloc (a.ma->cres->servinfo);
  refcount_inc ();
  authinfo.type = SFS_AUTHINFO;
  portno = getportno (a.fd);
  if (a.ma->carg.civers == 4) {
    authinfo.service = carg.ci4->service;
    path = si->mkpath (2, portno);
  }
  else {
    path = a.ma->carg.ci5->sname;
    authinfo.service = carg.ci5->service;
  }

  if (a.ma->hostname.len ())
    dnsname = a.ma->hostname;
  authinfo.name = si->get_hostname ();
  assert (si->ckpath (path));
  si->mkhostid_client (&authinfo.hostid);

  prog.pathtab.insert (this);
  prog.idleq.insert_tail (this);

  waitq.push_back (wrap (this, &sfsserver::retrootfh, a.cb));
  if (a.fd >= 0) {
    setfd (a.fd);
    // Yuck... can't call virtual function from supertype constructor
    delaycb (0, wrap (this, &sfsserver::crypt, *a.ma->cres, si,
		      wrap (mkref (this), &sfsserver::getsessid)));
  }
  else
    reconnect_0 ();
  ns->setcb (wrap (this, &sfsserver::getnfscall));
}

sfsserver::~sfsserver ()
{
  if (fsinfo)
    fsinfo_free (fsinfo);
  prog.pathtab.remove (this);
  prog.idleq.remove (this);
}

void
sfsserver::setfd (int fd)
{
  assert (fd >= 0);
  x = axprt_crypt::alloc (fd);	// XXX - shouldn't always be crypt
  sfsc = aclnt::alloc (x, sfs_program_1);
  sfscbs = asrv::alloc (x, sfscb_program_1,
			wrap (this, &sfsserver::sfsdispatch));
}

void
sfsserver::touch ()
{
  lastuse = sfs_get_timenow();
  prog.idleq.remove (this);
  prog.idleq.insert_tail (this);
}

void
sfsserver::reconnect ()
{
  if (!locked ()) {
    lock ();
    assert (!recon_tmo);
    if (x) {
      sfsc = NULL;
      sfscbs = NULL;
      close (x->reclaim ());
      x = NULL;
    }
    srvl = NULL;
    flushstate ();
    // prog.cdc->call (SFSCDCBPROC_HIDEFS, &path, NULL, aclnt_cb_null);
    reconnect_0 ();
  }
}

void
sfsserver::reconnect_0 ()
{
  recon_tmo = NULL;
  if (dnsname && dnsname.len ()
      && ((portno >= 0 && !srvl) || isdigit (dnsname[dnsname.len () - 1])))
    tcpconnect (dnsname, portno,
		wrap (mkref (this), &sfsserver::reconnect_1), false);
  else {
    str hname;
    u_int16_t port;
    if (!sfs_parsepath (path, &hname, NULL, &port))
      panic << path << ": cannot parse\n";
    if (port)
      tcpconnect (hname, port,
		  wrap (mkref (this), &sfsserver::reconnect_1), false,
		  &dnsname);
    else
      tcpconnect_srv (hname, "sfs", SFS_PORT,
		      wrap (mkref (this), &sfsserver::reconnect_1), false,
		      &srvl, &dnsname);
  }
}

void
sfsserver::reconnect_1 (int fd)
{
  if (condemned ()) {
    close (fd);
    unlock ();
  }
  else if (fd < 0) {
    srvl = NULL;
    // dnsname = NULL;
    connection_failure ();
  }
  else {
    int port = getportno (fd);
    if (port < 0) {
      close (fd);
      connection_failure ();
    }
    else {
      portno = port;
      setfd (fd);
      ref<sfs_connectres> cres = New refcounted<sfs_connectres>;
      sfsc->call (SFSPROC_CONNECT, &carg, cres,
		  wrap (mkref (this), &sfsserver::reconnect_2, cres));
    }
  }
}

void
sfsserver::reconnect_2 (ref<sfs_connectres> cres, clnt_stat err)
{
  ref<sfs_servinfo_w> nsi = sfs_servinfo_w::alloc (cres->reply->servinfo);

  if (condemned ())
    unlock ();
  else if (err)
    connection_failure ();
  else if (cres->status == SFS_REDIRECT)
    connection_failure (true);
  else if (cres->status)
    connection_failure ();
  else if (srvl && si->get_hostname () != nsi->get_hostname ())
    connection_failure ();
  else if (!(*si == *nsi) || !nsi->ckpath (path)) {
    warn << path << ": server has changed name or public key\n";
    connection_failure (true);
  }
  else
    crypt (*cres->reply, si, wrap (mkref (this), &sfsserver::getsessid));
}

void
sfsserver::getsessid (const sfs_hash *sessid)
{
  if (condemned ())
    unlock ();
  else if (!sessid)
    connection_failure ();
  else {
    srvl = NULL;
    authinfo.sessid = *sessid;
    sfs_fsinfo *fsi = fsinfo_alloc ();
    sfsc->call (SFSPROC_GETFSINFO, NULL, fsi,
		wrap (mkref (this), &sfsserver::getfsinfo, fsi),
		NULL, NULL, fsinfo_marshall ());
  }
}

void
sfsserver::getfsinfo (sfs_fsinfo *fsi, clnt_stat err)
{
  if (condemned ())
    unlock ();
  else if (err || !fsi)
    connection_failure ();
  else {
    setrootfh (fsi, wrap (this, &sfsserver::getfsinfo2, fsi));
    return;
  }

  fsinfo_free (fsi);
}

void
sfsserver::getfsinfo2 (sfs_fsinfo *fsi, bool err)
{
  if (err)
    connection_failure (true);
  else {
    nfs_fh3 fh (rootfh);
    if (fsinfo && fh.data != rootfh.data)
      connection_failure (true);
    else {
      recon_backoff = 1;
      if (fsinfo) {
	warn << "Reconnected to " << path << "\n";
	prog.cdc->call (SFSCDCBPROC_SHOWFS, &path, NULL, aclnt_cb_null);
	fsinfo_free (fsinfo);
      }
      fsinfo = fsi;
      initstate ();
      unlock ();
      return;
    }
  }

  fsinfo_free (fsi);
}

void
sfsserver::connection_failure (bool permanent)
{
  if (!fsinfo) {
    // XXX - doesn't work - what if GETFSINFO RPC itself fails
    condemn ();
    unlock ();
    destroy ();
  }
  else if (permanent) {
    warn << "Must remount " << path << "\n";
    condemn ();
    prog.cdc->call (SFSCDCBPROC_DELFS, &path, NULL, aclnt_cb_null);
    unlock ();
  }
  else {
    if (recon_backoff < 64)
      recon_backoff <<= 1;
    prog.cdc->call (SFSCDCBPROC_HIDEFS, &path, NULL, aclnt_cb_null);
    recon_tmo = delaycb (recon_backoff, 0,
			 wrap (mkref (this), &sfsserver::reconnect_0));
  }
}

void
sfsserver::unlock ()
{
  assert (lock_flag);
  lock_flag = false;
  vec<cbv> q;
  q.swap (waitq);
  while (!q.empty ())
    (*q.pop_front ()) ();
}

void
sfsserver::sfsdispatch (svccb *sbp)
{
  if (!sbp) {
    if (!condemned ()) {
      warn << "EOF from " << path << "\n";
      reconnect ();
    }
    return;
  }

  switch (sbp->proc ()) {
  case SFSCBPROC_NULL:
    sbp->reply (NULL);
    break;
  default:
    sbp->reject (PROC_UNAVAIL);
    break;
  }
}

void
sfsserver::retrootfh (fhcb cb)
{
  nfs_fh3 fh (rootfh);
  if (condemned ())
    (*cb) (NULL);
  else if (!ns->encodefh (fh)) {
    (*cb) (NULL);
    destroy ();
  }
  else
    (*cb) (&fh);
}

void
sfsserver::condemn ()
{
  if (!condemn_flag) {
    condemn_flag = true;
    if (recon_tmo) {
      timecb_remove (recon_tmo);
      recon_tmo = NULL;
      unlock ();
    }
    flushstate ();
    vec<cbv> q;
    q.swap (waitq);
    while (!q.empty ())
      (*q.pop_front ()) ();
  }
}

void
sfsserver::destroy ()
{
  assert (!destroyed);
  destroyed = true;
  condemn();
  refcount_dec ();
}

void
sfsserver::getnfscall (nfscall *nc)
{
  if (condemned ()) {
    nc->error (NFS3ERR_STALE);
    return;
  }
  if (locked () || !x || x->ateof ()) {
    waitq.push_back (wrap (this, &sfsserver::getnfscall, nc));
    return;
  }

  if (nc->proc () == NFSPROC3_NULL) {
    nc->reply (NULL);
    return;
  }
  if (nc->proc () != NFSPROC3_GETATTR
      || nc->Xtmpl getarg<nfs_fh3> ()->data != rootfh.data) {
    touch ();
    if (!authok (nc))
      return;
  }
  if (!prog.intercept (this, nc))
    dispatch (nc);
}

void
sfsserver::crypt (sfs_connectok cres, ref<const sfs_servinfo_w> si,
		  const sfsserver::crypt_cb cb)
{
  sfs_hash h;
  (*cb) (&h);
}

sfsprog::sfsprog (ref<axprt_unix> cdx, sfsprog::allocfn_t f,
		  bool ncl, bool mcl)
  : newserver (f), needclose (ncl), mntwithclose (mcl), x (cdx),
    cdc (aclnt::alloc (x, sfscdcb_program_1)),
    cds (asrv::alloc (x, sfscd_program_1, wrap (this, &sfsprog::cddispatch))),
    ns (mntwithclose ? New refcounted<nfsserv_udp>(cl_nfs_program_3)
	: New refcounted<nfsserv_udp>),
    nd (New refcounted<nfsdemux> (New refcounted<nfsserv_fixup>(ns))),
    linkserv (nd->servalloc ()), idletmo (NULL)
{
  linkserv->setcb (wrap (this, &sfsprog::linkdispatch));
  ifchgcb (wrap (this, &sfsprog::sockcheck));
}

void
sfsprog::cddispatch (svccb *sbp)
{
  if (!sbp)
    fatal ("EOF from sfscd\n");
  switch (sbp->proc ()) {
  case SFSCDPROC_NULL:
    sbp->reply (NULL);
    break;
  case SFSCDPROC_INIT:
    sfs_suidserv (sbp->Xtmpl getarg<sfscd_initarg> ()->name,
		  wrap (this, &sfsprog::ctlaccept));
    sbp->reply (NULL);
    break;
  case SFSCDPROC_MOUNT:
    {
      sfscd_mountarg *arg = sbp->Xtmpl getarg<sfscd_mountarg> ();
      ref<nfsserv> nns = nd->servalloc ();
      if (needclose)
	nns = close_simulate (nns);
      int fd = arg->cres ? x->recvfd () : -1;
      newserver (this, nns, fd, arg,
		 wrap (this, &sfsprog::mountcb, sbp));
      tmosched ();
      break;
    }
  case SFSCDPROC_UNMOUNT:
    if (sfsserver *s = pathtab[*sbp->Xtmpl getarg<nfspath3> ()])
      s->destroy ();
    sbp->reply (NULL);
    break;
  case SFSCDPROC_FLUSHAUTH:
    {
      sfs_aid aid = *sbp->Xtmpl getarg<sfs_aid> ();
      for (sfsserver *s = pathtab.first (); s; s = pathtab.next (s))
	s->authclear (aid);
      sbp->reply (NULL);
      break;
    }
  case SFSCDPROC_CONDEMN:
    if (sfsserver *s = pathtab[*sbp->Xtmpl getarg<nfspath3> ()])
      s->condemn ();
    sbp->reply (NULL);
    break;
  default:
    sbp->reject (PROC_UNAVAIL);
    break;
  }
}

void
sfsprog::tmosched (bool expired)
{
  if (expired)
    idletmo = NULL;
  sfsserver *si;
  while ((si = idleq.first)) {
    if (si->lastuse + mount_idletime > sfs_get_timenow()) {
      if (!idletmo)
	idletmo = delaycb (si->lastuse + mount_idletime - sfs_get_timenow(),
			   wrap (this, &sfsprog::tmosched, true));
      return;
    }
    if (!si->locked () && !si->condemned ())
      cdc->call (SFSCDCBPROC_IDLE, &si->path, NULL, aclnt_cb_null);
    si->touch ();
  }
}

void
sfsprog::mountcb (svccb *sbp, const nfs_fh3 *fhp)
{
  sfscd_mountres res (NFS_OK);
  if (fhp) {
    res.reply->mntflags = NMOPT_NFS3;
    if (mntwithclose)
      res.reply->mntflags |= NMOPT_SENDCLOSE;
    res.reply->fh = fhp->data;
    x->sendfd (ns->getfd (), false);
  }
  else
    res.set_err (EIO);
  sbp->reply (&res);
}

void
sfsprog::ctlaccept (ptr<axprt_unix> x, const authunix_parms *aup)
{
  if (x && !x->ateof ())
    vNew sfsctl (x, aup, this);
}

static void
mklnkfattr (fattr3exp *f, const nfs_fh3 *fh)
{
  bzero (f, sizeof (*f));
  f->type = NF3LNK;
  f->mode = 0444;
  f->nlink = 1;
  f->gid = sfs_gid;
  f->used = 0;
  f->fileid = 1;
  f->size = 2 * fh->data.size ();
}

void
sfsprog::linkdispatch (nfscall *nc)
{
  static const char hexchars[] = "0123456789abcdef";

  switch (nc->proc ()) {
  case NFSPROC3_GETATTR:
    {
      nfs_fh3 *arg = nc->Xtmpl getarg<nfs_fh3> ();
      getattr3res res (NFS3_OK);
      mklnkfattr (res.attributes.addr (), arg);
      nc->reply (&res);
      break;
    }
  case NFSPROC3_READLINK:
    {
      nfs_fh3 *arg = nc->Xtmpl getarg<nfs_fh3> ();
      readlink3res res (NFS3_OK);
      res.resok->symlink_attributes.set_present (true);
      mklnkfattr (res.resok->symlink_attributes.attributes.addr (), arg);

      mstr m (2 * arg->data.size ());
      for (size_t i = 0; i < arg->data.size (); i++) {
	u_char b = arg->data[i];
	m[2*i] = hexchars[b>>4];
	m[2*i+1] = hexchars[b&0xf];
      }
      res.resok->data = m;

      nc->reply (&res);
      break;
    }
  default:
    nc->error (NFS3ERR_ACCES);
    break;
  }
}

void
sfsprog::sockcheck ()
{
  for (sfsserver *np, *p = idleq.first; p; p = np) {
    np = idleq.next (p);
    if (p->x)
      p->x->sockcheck ();
  }
}

bool
sfsprog::intercept (sfsserver *s, nfscall *nc)
{
  switch (nc->proc ()) {
  case NFSPROC3_SETATTR:
    {
      setattr3args *sar = nc->Xtmpl getarg<setattr3args> ();
      sattr3 &sa = sar->new_attributes;
      if (sa.mode.set || sa.size.set || sa.atime.set || sa.mtime.set
	  || !sa.uid.set || !sa.gid.set || *sa.uid.val != (u_int32_t) -2)
	return false;
      if (sfsctl *sc = ctltab (nc->getaid (), *sa.gid.val))
	sc->fip = New refcounted<sfsctl::fileinfo> (s, sar->object);
      nc->error (NFS3ERR_PERM);
      return true;
    }
  case NFSPROC3_LOOKUP:
    {
      diropargs3 *arg = nc->Xtmpl getarg<diropargs3> ();
      if (strncmp (arg->name, SFSPREF, sizeof (SFSPREF) - 1))
	return false;
      lookup3res res (NFS3_OK);
      res.resok->obj_attributes.set_present (true);
      if (arg->name == SFSFH) {
	res.resok->object = arg->dir;
	mklnkfattr (res.resok->obj_attributes.attributes.addr (), &arg->dir);
	linkserv->encodefh (res.resok->object);
	nc->reply (&res, xdr_lookup3res);
	return true;
      }
      return false;
    }
  default:
    return false;
    break;
  }
}

/* We have no idea what type aup.aup_gids is (could be gid_t,
 * u_int32_t, etc.)  Rather than autoconf it, just use templates to
 * work around the problem. */
template<class T> inline void
domalloc (T *&tp, size_t glen)
{
  tp = static_cast<T *> (xmalloc (glen));
}

sfsprog::sfsctl::sfsctl (ref<axprt_stream> x, const authunix_parms *naup,
			 sfsprog *p)
  : prog (p), aid (aup2aid (naup)), pid (0)
{
  const int glen = naup->aup_len * sizeof (naup->aup_gids[0]);
  aup = *naup;
  aup.aup_machname = xstrdup (naup->aup_machname);
  domalloc (aup.aup_gids, glen);
  memcpy (aup.aup_gids, naup->aup_gids, glen);
  s = asrv::alloc (x, sfsctl_prog_1, wrap (this, &sfsctl::dispatch));
  prog->ctltab.insert (this);
}

sfsprog::sfsctl::~sfsctl ()
{
  prog->ctltab.remove (this);
  xfree (aup.aup_machname);
  xfree (aup.aup_gids);
}

void
sfsprog::sfsctl::setpid (int32_t npid)
{
  prog->ctltab.remove (this);
  pid = npid;
  prog->ctltab.insert (this);
}

static void
sfsctl_err (svccb *sbp, nfsstat3 err)
{
  switch (sbp->proc ()) {
  case SFSCTL_LOOKUP:
    sbp->replyref (lookup3res (nfsstat3 (err)));
    break;
  default:
    sbp->replyref (err);
  }
}

inline nfsstat3
clnt2nfs (clnt_stat err)
{
  switch (err) {
  case RPC_SUCCESS:
    return NFS3_OK;
  case RPC_CANTSEND:
  case RPC_CANTRECV:
    return NFS3ERR_JUKEBOX;
  case RPC_AUTHERROR:
    return NFS3ERR_ACCES;
  default:
    return NFS3ERR_IO;
  }
}

static void
idnames_cb (svccb *sbp, sfs_idnames *resp, clnt_stat err)
{
  sfsctl_getidnames_res res (clnt2nfs (err));
  if (!res.status)
    *res.names = *resp;
  sbp->reply (&res);
  delete resp;
}

static void
idnums_cb (svccb *sbp, sfs_idnums *resp, clnt_stat err)
{
  sfsctl_getidnums_res res (clnt2nfs (err));
  if (!res.status)
    *res.nums = *resp;
  sbp->reply (&res);
  delete resp;
}

static void
getcred_cb (svccb *sbp, sfsauth_cred *resp, clnt_stat err)
{
  sfsctl_getcred_res res (clnt2nfs (err));
  if (!res.status)
    *res.cred = *resp;
  sbp->reply (&res);
  delete resp;
}

static void
lookup_cb (svccb *sbp, lookup3res *resp, clnt_stat err)
{
  if (err)
    resp->set_status (clnt2nfs (err));
  sbp->reply (resp);
  delete resp;
}

static void
getacl_cb (svccb *sbp, ex_read3res *resp, clnt_stat err)
{
  if (err)
    resp->set_status (clnt2nfs (err));
  nfs3_exp_disable (NFSPROC3_READ, resp);
  sbp->reply (resp);
  delete resp;
}

static void
setacl_cb (svccb *sbp, ex_write3res *resp, clnt_stat err)
{
  if (err)
    resp->set_status (clnt2nfs (err));
  nfs3_exp_disable (NFSPROC3_WRITE, resp);
  sbp->reply (resp);
  delete resp;
}

void
sfsprog::sfsctl::dispatch (svccb *sbp)
{
  if (!sbp) {
    delete this;
    return;
  }

  switch (sbp->proc ()) {
  case SFSCTL_NULL:
    sbp->reply (NULL);
    return;
  case SFSCTL_SETPID:
    setpid (*sbp->Xtmpl getarg<int32_t> ());
    sbp->reply (NULL);
    return;
  }

  sfsserver *si = prog->pathtab[*sbp->Xtmpl getarg<filename3> ()];
  if (!si) {
    sfsctl_err (sbp, NFS3ERR_STALE);
    return;
  }
  if (!si->sfsc) {
    sfsctl_err (sbp, NFS3ERR_JUKEBOX);
    return;
  }
  AUTH *auth = si->authof (aid);

  switch (sbp->proc ()) {
  case SFSCTL_GETFH:
    {
      sfsctl_getfh_res res;
      if (fip && fip->fspath == si->path)
	*res.fh = fip->fh;
      else
	res.set_status (NFS3ERR_STALE);
      fip = NULL;
      sbp->reply (&res);
      break;
    }

  case SFSCTL_GETIDNAMES:
    {
      sfsctl_getidnames_arg *argp
	= sbp->Xtmpl getarg<sfsctl_getidnames_arg> ();
      sfs_idnames *resp = New sfs_idnames;
      si->sfsc->call (SFSPROC_IDNAMES, &argp->nums, resp,
		      wrap (idnames_cb, sbp, resp), auth);
      break;
    }

  case SFSCTL_GETIDNUMS:
    {
      sfsctl_getidnums_arg *argp
	= sbp->Xtmpl getarg<sfsctl_getidnums_arg> ();
      sfs_idnums *resp = New sfs_idnums;
      si->sfsc->call (SFSPROC_IDNUMS, &argp->names, resp,
		      wrap (idnums_cb, sbp, resp), auth);
      break;
    }

  case SFSCTL_GETCRED:
    {
      sfsauth_cred *resp = New sfsauth_cred;
      si->sfsc->call (SFSPROC_GETCRED, NULL, resp,
		      wrap (getcred_cb, sbp, resp), auth);
      break;
    }

  case SFSCTL_LOOKUP:
    {
      sfsctl_lookup_arg *argp
	= sbp->Xtmpl getarg<sfsctl_lookup_arg> ();
      lookup3res *resp = New lookup3res;
      si->sfsc->call (NFSPROC3_LOOKUP, &argp->arg, resp,
		      wrap (lookup_cb, sbp, resp), auth,
		      xdr_diropargs3, xdr_ex_lookup3res,
		      ex_NFS_PROGRAM, ex_NFS_V3);
      break;
    }
  case SFSCTL_GETACL:
    {
      sfsctl_getacl_arg *argp
	= sbp->Xtmpl getarg<sfsctl_getacl_arg> ();
      ex_read3res *resp = New ex_read3res;
      si->sfsc->call (ex_NFSPROC3_GETACL, &argp->arg, resp,
		      wrap (getacl_cb, sbp, resp), auth,
		      xdr_diropargs3, xdr_ex_read3res,
		      ex_NFS_PROGRAM, ex_NFS_V3);
      break;
    }
  case SFSCTL_SETACL:
    {
      sfsctl_setacl_arg *argp
	= sbp->Xtmpl getarg<sfsctl_setacl_arg> ();
      ex_write3res *resp = New ex_write3res;
      si->sfsc->call (ex_NFSPROC3_SETACL, &argp->arg, resp,
		      wrap (setacl_cb, sbp, resp), auth,
		      xdr_setaclargs, xdr_ex_write3res,
		      ex_NFS_PROGRAM, ex_NFS_V3);
      break;
    }

  default:
    sbp->reject (PROC_UNAVAIL);
    break;
  }
}
