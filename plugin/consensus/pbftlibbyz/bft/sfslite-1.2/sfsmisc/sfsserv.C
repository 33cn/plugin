/* $Id: sfsserv.C 1754 2006-05-19 20:59:19Z max $ */

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

#include "sfsserv.h"
#include "sfsauth_prot.h"
#include <grp.h>

ptr<aclnt>
getauthclnt ()
{
  static ptr<axprt> authxprt;
  static ptr<aclnt> authclnt;
  if (authclnt && !authxprt->ateof ())
    return authclnt;
  int fd = suidgetfd ("authd");
  if (fd >= 0)
    return authclnt = aclnt::alloc (authxprt = axprt_stream::alloc (fd),
				    sfsauth_prog_2);
  fd = suidgetfd ("authserv");
  if (fd >= 0)
    return authclnt = aclnt::alloc (authxprt = axprt_stream::alloc (fd),
				    sfsauth_program_1);
  authxprt = NULL;
  authclnt = NULL;
  return NULL;
}

sfsserv::sfsserv (ref<axprt_crypt> xxc, ptr<axprt> xx)
  : xc (xxc), x (xx ? xx : implicit_cast<ptr<axprt> > (xxc)),
    destroyed (New refcounted<bool> (false)),
    sfssrv (asrv::alloc (x, sfs_program_1, wrap (this, &sfsserv::dispatch))),
    seqstate (128), authid_valid (false)
{
  authtab.push_back ();
  credtab.push_back ();
  struct sockaddr_in sin;
  bzero (&sin, sizeof (sin));
  socklen_t sinlen = sizeof (sin);
  if (getpeername (xxc->getfd (), (sockaddr *) &sin, &sinlen)
      || sin.sin_family != AF_INET)
    client_name = "";
  else
    client_name = inet_ntoa (sin.sin_addr);
}

sfsserv::~sfsserv ()
{
  *destroyed = true;

  if (authid_valid && authc) {
    /* We want to clear up any potential stateful login information
     * still sitting around in sfsauthd.  After all, we might been in
     * the middle of some authentication.  The auth server only keeps
     * one pending login state per authid, so we just need to send an
     * SFS_NOAUTH login. */
    sfsauth2_loginarg arg;
    arg.arg.seqno = 0;
    arg.arg.certificate.setsize (4);
    putint (arg.arg.certificate.base (), SFS_NOAUTH);
    arg.authid = authid;
    arg.source = strbuf () << client_name << "!"
			   << (progname ? progname : str ("???"));
    authc->call (SFSAUTH2_LOGIN, &arg, NULL, aclnt_cb_null,
		 NULL, NULL, xdr_void);
  }
}

u_int32_t
sfsserv::authnoalloc ()
{
  u_int32_t authno;

  if (authfreelist.size ())
    authno = authfreelist.pop_back ();
  else if (authtab.size () >= 0x10000) {
    warn << "no authnos available";
    authno = 0;
  }
  else {
    authno = authtab.size ();
    authtab.push_back ();
    credtab.push_back ();
  }
  return authno;
}

u_int32_t
sfsserv::authalloc (const sfsauth_cred *cp, u_int n)
{
  u_int32_t authno;
  for (u_int i = 0; i < n; i++) {
    const sfsauth_cred &c = cp[i];
    switch (c.type) {
    case SFS_UNIXCRED:
    case SFS_NOCRED:
      {
	if (!(authno = authnoalloc ()))
	  return 0;
	if (c.type == SFS_UNIXCRED) {
	  const sfs_unixcred &uc = *c.unixcred;
	  authtab[authno] = authunixint_create ("localhost", uc.uid, uc.gid,
						uc.groups.size (),
						uc.groups.base ());
	  credtab[authno] = c;
	} else {
	  authtab[authno] = authuint_create (authno);
	  credtab[authno].set_type (SFS_NOCRED);
	}
	return authno;
      }
      
    default:
      break;
    }
  }
  return 0;
}
void
sfsserv::authfree (size_t n)
{
  if (n && n < authtab.size () && authtab[n]) {
    authtab[n] = NULL; 
    credtab[n].set_type (SFS_NOCRED);
    authfreelist.push_back (n);
  }
}

void
sfsserv::dispatch (svccb *sbp)
{
  if (!sbp)
    return;

  switch (sbp->proc ()) {
  case SFSPROC_NULL:
    sbp->reply (NULL);
    break;
  case SFSPROC_CONNECT:
    sfs_connect (sbp);
    break;
  case SFSPROC_ENCRYPT:
    sfs_encrypt (sbp, 1);
    break;
  case SFSPROC_ENCRYPT2:
    sfs_encrypt (sbp, 2);
    break;
  case SFSPROC_GETFSINFO:
    sfs_getfsinfo (sbp);
    break;
  case SFSPROC_LOGIN:
    sfs_login (sbp);
    break;
  case SFSPROC_LOGOUT:
    sfs_logout (sbp);
    break;
  case SFSPROC_IDNAMES:
    sfs_idnames (sbp);
    break;
  case SFSPROC_IDNUMS:
    sfs_idnums (sbp);
    break;
  case SFSPROC_GETCRED:
    sfs_getcred (sbp);
    break;
  default:
    sfs_badproc (sbp);
    return;
  }
}

void
sfsserv::sfs_connect (svccb *sbp)
{
  if (cd || authid_valid) {
    sbp->reject (PROC_UNAVAIL);
    return;
  }
  cd.alloc ();
  cd->ci = *sbp->Xtmpl getarg <sfs_connectarg> ();
  cd->cr.set_status (SFS_OK);
  cd->cr.reply->charge.bitcost = sfs_hashcost;
  rnd.getbytes (cd->cr.reply->charge.target.base (), charge.target.size ());
  privkey = doconnect (&cd->ci, &cd->cr.reply->servinfo);
  si = sfs_servinfo_w::alloc (cd->cr.reply->servinfo);
  if (!privkey && !cd->cr.status)
    cd->cr.set_status (SFS_NOSUCHHOST);
  sbp->reply (&cd->cr);
}

static sfs_service
ci2service (const sfs_connectinfo &ci)
{
  switch (ci.civers) {
  case 4:
    return ci.ci4->service;
  case 5:
    return ci.ci5->service;
  default:
    return sfs_service (0);
  }
}

void
sfsserv::sfs_encrypt (svccb *sbp, int pvers)
{
  if (!cd || cd->cr.status || !si) {
    sbp->reject (PROC_UNAVAIL);
    return;
  }
  sfs_server_crypt (sbp, privkey, cd->ci, si,
		    &sessid, cd->cr.reply->charge, xc, pvers);
  sfs_hash hostid;
  bool hostid_ok = si->mkhostid (&hostid);
  assert (hostid_ok);
  sfs_get_authid (&authid, ci2service (cd->ci), si->get_hostname (), 
		  &hostid, &sessid);
  authid_valid = true;
  // cd.clear ();
}

static void
sfs_login_cb (ref<bool> destroyed, sfsserv *srv, svccb *sbp,
	      sfsauth_loginres *_resp, clnt_stat stat)
{
  auto_ptr<sfsauth_loginres> resp (_resp);
  if (stat || *destroyed) {
    if (stat)
      warn << "authserv: " << stat << "\n";
    sbp->replyref (sfs_loginres (SFSLOGIN_ALLBAD));
    return;
  }

  sfs_loginres res (resp->status);
  switch (resp->status) {
  case SFSLOGIN_OK:
    {
      if (resp->resok->authid != srv->authid
	  || !srv->seqstate.check (resp->resok->seqno)) {
	res.set_status (SFSLOGIN_BAD);
	break;
      }

      res.resok->authno = srv->authalloc (&resp->resok->cred, 1);
      if (!res.resok->authno) {
	warn << "credential type not supported (v1)\n";
	res.set_status (SFSLOGIN_BAD);
      }
      break;
    }
  case SFSLOGIN_MORE:
    *res.resmore = *resp->resmore;
    break;
  default:
    break;
  }
  sbp->reply (&res);
}
static void
sfs_login2_cb (ref<bool> destroyed, sfsserv *srv, svccb *sbp,
	       ref<sfsauth2_loginres> resp, clnt_stat stat)
{
  if (stat || *destroyed) {
    if (stat)
      warn << "credential type not supported (v2)\n";
    sbp->replyref (sfs_loginres (SFSLOGIN_ALLBAD));
    return;
  }

  sfs_loginarg *argp = sbp->Xtmpl getarg<sfs_loginarg> ();
  sfs_loginres res (resp->status);
  switch (resp->status) {
  case SFSLOGIN_OK:
    {
      if (resp->resok->creds.size () < 1) {
	res.set_status (SFSLOGIN_BAD) ;
	break;
      }
      res.resok->authno = srv->authalloc (resp->resok->creds.base (),
					  resp->resok->creds.size ());
      if (!res.resok->authno) {
	//warn << "ran out of authnos (or bad cred type v2)\n";
	res.set_status (SFSLOGIN_BAD);
	break;
      }
      else if (!srv->seqstate.check (argp->seqno))
	res.set_status (SFSLOGIN_BAD);
      else {
	res.resok->resmore = resp->resok->resmore;
	res.resok->hello = resp->resok->hello;
      }
      break;
    }
  case SFSLOGIN_MORE:
    *res.resmore = *resp->resmore;
    break;
  default:
    break;
  }
  sbp->reply (&res);
}
void
sfsserv::sfs_login (svccb *sbp)
{
  if (!authid_valid
      || ((!authc || authc->xi->ateof ()) && !(authc = getauthclnt ()))) {
    sbp->replyref (sfs_loginres (SFSLOGIN_ALLBAD));
    return;
  }
  if (authc->rp.versno == 1) {
    sfsauth_loginres *resp = New sfsauth_loginres;
    authc->call (SFSAUTHPROC_LOGIN,
		 sbp->Xtmpl getarg<sfs_loginarg> (), resp,
		 wrap (sfs_login_cb, destroyed, this, sbp, resp));
    return;
  }
  ref<sfsauth2_loginres> resp = New refcounted<sfsauth2_loginres> ();
  sfsauth2_loginarg arg;
  arg.arg = *sbp->Xtmpl getarg<sfs_loginarg> ();
  arg.authid = authid;
  arg.source = strbuf () << client_name << "!"
			 << (progname ? progname : str ("???"));
  authc->call (SFSAUTH2_LOGIN, &arg, resp,
	       wrap (sfs_login2_cb, destroyed, this, sbp, resp));
}

void
sfsserv::sfs_logout (svccb *sbp)
{
  authfree (*sbp->Xtmpl getarg<u_int32_t> ());
  sbp->reply (NULL);
}

// XXX - MAJOR DEADLOCK PROBLEMS HERE
// XXX - should never call getpw* or getgr*
void
sfsserv::sfs_idnames (svccb *sbp)
{
  if (!getauth (sbp->getaui ())) {
    sbp->reject (AUTH_REJECTEDCRED);
    return;
  }

  ::sfs_idnums *argp = sbp->Xtmpl getarg< ::sfs_idnums> ();
  ::sfs_idnames res;
  if (argp->uid != -1)
    if (struct passwd *p = getpwuid (argp->uid)) {
      res.uidname.set_present (true);
      *res.uidname.name = p->pw_name;
    }
  if (argp->gid != -1)
    if (struct group *g = getgrgid (argp->gid)) {
      res.gidname.set_present (true);
      *res.gidname.name = g->gr_name;
    }
  sbp->reply (&res);
}

// XXX - MAJOR DEADLOCK PROBLEMS HERE
// XXX - should never call getpw* or getgr*
void
sfsserv::sfs_idnums (svccb *sbp)
{
  if (!getauth (sbp->getaui ())) {
    sbp->reject (AUTH_REJECTEDCRED);
    return;
  }

  ::sfs_idnames *argp = sbp->Xtmpl getarg< ::sfs_idnames> ();
  ::sfs_idnums res = { -1, -1 };
  if (argp->uidname.present)
    if (struct passwd *p = getpwnam (argp->uidname.name->cstr ()))
      res.uid = p->pw_uid;
  if (argp->gidname.present)
    if (struct group *g = getgrnam (argp->gidname.name->cstr ()))
      res.gid = g->gr_gid;
  sbp->reply (&res);
}

void
sfsserv::sfs_getcred (svccb *sbp)
{
  u_int32_t authno = sbp->getaui ();
  if (authno < credtab.size ())
    sbp->replyref (credtab[authno]);
  else
    sbp->replyref (sfsauth_cred (SFS_NOCRED));
}


static ptr<axprt_stream>
sfs_accept (bool primary, sfsserv_axprt_cb cb, int fd)
{
  if (fd < 0) {
    if (primary)
      (*cb) (NULL);
    return NULL;
  }
  tcp_nodelay (fd);
  ref<axprt_crypt> x = axprt_crypt::alloc (fd);
  return (*cb) (x);
}

static void
sfs_accept_standalone (sfsserv_axprt_cb cb, int sfssfd)
{
  sockaddr_in sin;
  bzero (&sin, sizeof (sin));
  socklen_t sinlen = sizeof (sin);
  int fd = accept (sfssfd, reinterpret_cast<sockaddr *> (&sin), &sinlen);
  if (fd >= 0)
    sfs_accept (true, cb, fd);
  else if (errno != EAGAIN)
    warn ("accept: %m\n");
}

static void
sfssd_slavegen_cb (sfsserv_axprt_cb cb, int fd)
{
  if (!cloneserv (fd, wrap (sfs_accept, false, cb))) {
    warn ("sfssd_slavegen_cb:  cloneserv:  %m\n");
    close (fd);
  }
}

static ptr<axprt_crypt>
sfssd_cb_with_axprt(sfsserv_cb cb, ptr<axprt_crypt> x)
{
  (*cb)(x);
  return x;
}

void
sfssd_slave (sfsserv_cb cb)
{
  u_int16_t port = sfs_defport ? sfs_defport : SFS_PORT;
  if (!sfssd_slave_axprt (wrap (sfssd_cb_with_axprt, cb), true, port))
    fatal ("binding TCP port %d: %m\n", port);
}

bool
sfssd_slave (sfsserv_cb cb, bool allowstandalone, u_int port)
{
  return sfssd_slave_axprt (wrap (sfssd_cb_with_axprt, cb), 
			    allowstandalone, port);
}

void
sfssd_slavegen (str sock, sfsserv_cb cb)
{
  sfssd_slavegen_axprt (sock, wrap (sfssd_cb_with_axprt, cb));
}

void
sfssd_slavegen_axprt(str sock, sfsserv_axprt_cb cb)
{
  sfs_unixserv (sock, wrap (sfssd_slavegen_cb, cb));
}

bool
sfssd_slave_axprt (sfsserv_axprt_cb cb, bool allowstandalone, u_int port)
{
  if (!port)
    port = SFS_PORT;
  if (cloneserv (0, wrap (sfs_accept, true, cb)))
    return true;
  else if (!allowstandalone) {
    warn ("No sfssd detected.\n");
    return true;
  }
  warn ("No sfssd detected, running in standalone mode.\n");
  int sfssfd = inetsocket (SOCK_STREAM, port);
  if (sfssfd < 0)
    return false;
  close_on_exec (sfssfd);
  listen (sfssfd, 5);
  fdcb (sfssfd, selread, wrap (sfs_accept_standalone, cb, sfssfd));
  return true;
}
