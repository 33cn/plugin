/* $Id: sfsconnect.C 3140 2007-11-27 17:05:50Z max $ */

/*
 *
 * Copyright (C) 1999-2004 David Mazieres (dm@uun.org)
 * Copyright (C) 2000 Michael Kaminsky (kaminsky@lcs.mit.edu)
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

#include "sfsconnect.h"
#include "rxx.h"
#include "parseopt.h"
#include "sfscrypt.h"
#include "dns.h"

ihash<const sfs_hash, sfs_pathcert,
      &sfs_pathcert::hostid, &sfs_pathcert::hlink> pathcert_tab;

void
sfs_initci (sfs_connectinfo *ci, str path, sfs_service service,
	    sfs_extension *ext_base, size_t ext_size)
{
  ci->set_civers (5);
  ci->ci5->release = sfs_release;
  ci->ci5->service = service;
  ci->ci5->sname = path;
  if (ext_size)
    ci->ci5->extensions.set (ext_base, ext_size);
  else
    ci->ci5->extensions.clear ();
}

bool
sfs_nextci (sfs_connectinfo *ci)
{
  if (ci->civers == 5) {
    sfs_service service = ci->ci5->service;
    str sname = ci->ci5->sname;
    rpc_vec<sfs_extension, RPC_INFINITY> exts;
    exts.swap (ci->ci5->extensions);

    ci->set_civers (4);
    ci->ci4->service = service;
    if (!sfs_parsepath (sname, &ci->ci4->name, &ci->ci4->hostid))
      ci->ci4->name = sname;
    exts.swap (ci->ci4->extensions);
    return true;
  }
  return false;
}

ptr<sfspriv> sfs_connect_t::ckey;

void
sfs_connect_t::srvfail (int e, str msg)
{
  if (srvl) {
    if (msg)
      warn << "server " << dnsname << " for " << msg << "\n";
    carg_reset ();
    last_srv_err = e;
    tcpc = tcpconnect_srv_retry (srvl,
				 wrap (this, &sfs_connect_t::getfd, destroyed),
				 &dnsname);
  }
  else
    fail (e, msg);
}

void
sfs_connect_t::succeed ()
{
  c = NULL;
  (*cb) (sc, NULL);
  delete this;
}

bool
sfs_connect_t::setsock (int fd)
{
  bzero (&sc->sa, sizeof (sc->sa));
  sc->sa_len = sizeof (sc->sa);

  if (!isunixsocket (fd)) {
    tcp_nodelay (fd);
    if (getpeername (fd, &sc->sa.sa, &sc->sa_len)) {
      fail (errno, location << ": getpeername: " << strerror (errno));
      return false;
    }
    int portno = -1;
    switch (sc->sa.sa.sa_family) {
    case AF_INET:
      portno = ntohs (sc->sa.sa_in.sin_port);
      break;
    default:
      fail (EPROTONOSUPPORT, location << ": unknown protocol family "
	    << sc->sa.sa.sa_family << "\n");
      return false;
    }
    sc->port = portno;
  }
  else
    sc->port = 0;

  sc->x = axprt_crypt::alloc (fd);
  c = aclnt::alloc (sc->x, sfs_program_1);
  return true;
}

bool
sfs_connect_t::start_common ()
{
#if 0
  if (encrypt)
    random_start ();
#endif

  c = NULL;
  srvl = NULL;
  dnsname = NULL;
  local_authd = false;
  last_srv_err = 0;
  sc->path = sname ();
  sc->service = service ();

  int v;

 restart:
  carg_reset ();
  if (!location_cache.insert (sname ())) {
    fail (ELOOP, sname () << ": redirect loop");
    return false;
  }
  else if (location_cache.size () > 32) {
    fail (ELOOP, sname () << ": too many levels of connect redirection");
    return false;
  }
    
  if (sname () == "-") {
    location = "localhost";
    port = 0;
    bzero (hostid.base (), hostid.size ());
  }
  else if (!sfs_parsepath (sname (), &location, &hostid, &port, &v)) {
    const char *p = sname ();
    if (check_hostid
	|| !sfsgetlocation (p, &location, &port, false) || *p) {
      fail (ENOENT, sname () << ": cannot parse path");
      return false;
    }
    bzero (hostid.base (), hostid.size ());
  }
  else if (v == 2) {
    if (sfs_pathcert *pc = sfs_pathcert::lookup (hostid)) {
      if (str dest = pc->dest ()) {
	sname () = dest;
	goto restart;
      }
      else
	srvfail (ESTALE, ": Public key has been revoked");
    }
  }

  return true;
}

bool
sfs_connect_t::start ()
{
  if (!start_common ())
    return false;

  if (sname () == "-") {
    int fd = suidgetfd ("authd");
    local_authd = true;
    errno = ECONNREFUSED;
    getfd (destroyed, fd);
    return true;
  }

  sfs_hosttab_init ();
  if (const sfs_host *hp = sfs_hosttab.lookup (location)) {
    switch (hp->sa.sa.sa_family) {
    case AF_INET:
      dnsname = inet_ntoa (hp->sa.sa_in.sin_addr);
      tcpc = tcpconnect (hp->sa.sa_in.sin_addr, ntohs (hp->sa.sa_in.sin_port),
			 wrap (this, &sfs_connect_t::getfd,
			       destroyed));
      return true;
    default:
      fail (EPROTONOSUPPORT, location << ": unknown protocol family "
	    << hp->sa.sa.sa_family << "\n");
      return false;
    }
  }

  if (port)
    tcpc = tcpconnect (location, port, wrap (this, &sfs_connect_t::getfd,
					     destroyed));
  else
    tcpc = tcpconnect_srv (location, "sfs", SFS_PORT,
			   wrap (this, &sfs_connect_t::getfd, destroyed),
			   true, &srvl, &dnsname);
  return true;
}

bool
sfs_connect_t::start (ptr<aclnt> cc)
{
  if (!start_common ())
    return false;
  c = cc;
  sendconnect ();
  return true;
}

bool
sfs_connect_t::start (ptr<srvlist> sl)
{
  if (!start_common ())
    return false;

  assert (!port);

  srvl = sl;
  tcpc = tcpconnect_srv_retry (srvl,
			       wrap (this, &sfs_connect_t::getfd, destroyed),
			       &dnsname);
  return true;
}

bool
sfs_connect_t::start (in_addr a, u_int16_t p)
{
  if (!start_common ())
    return false;

  if (p)
    port = p;
  else if (!port)
    port = SFS_PORT;

  tcpc = tcpconnect (a, port, wrap (this, &sfs_connect_t::getfd, destroyed));
  return true;
}

bool
sfs_connect_t::start_scon ()
{
  switch (sc->ci.civers) {
  case 4:
    sname () = strbuf () << sc->ci.ci4->name << ":"
			 << armor32 (sc->ci.ci4->hostid.base (),
				     sc->ci.ci4->hostid.size ());
    service () = sc->ci.ci4->service;
    exts () = sc->ci.ci4->extensions;
    break;
  case 5:
    ci5 = *sc->ci.ci5;
    break;
  default:
    panic ("unknown sfs_connectarg version\n");
    return false;
  }

  if (!start_common ())
    return false;

  if (sc->auth) {
    AUTH_DESTROY (sc->auth);
    sc->auth = NULL;
  }

  cres.set_status (SFS_OK);
  *cres.reply = sc->cres;
  dnsname = sc->dnsname;
  if (dnsname && !dnsname.len ())
    dnsname = NULL;
  c = aclnt::alloc (sc->x, sfs_program_1);
  // return dogetconres ();
  if (encrypt)
    docrypt ();
  else
    succeed ();
  return true;
}

void
sfs_connect_t::cancel ()
{
  delete this;
}

sfs_connect_t::~sfs_connect_t ()
{
#ifdef DMALLOC
  assert (!*destroyed);
#endif /* DMALLOC */
  *destroyed = true;
  if (tcpc)
    tcpconnect_cancel (tcpc);
  if (cbase)
    cbase->cancel ();
}

void
sfs_connect_t::init ()
{
    ci5.release = SFS_RELEASE;
    ci5.service = SFS_SFS;
    aarg.ntries = 0;
    aarg.seqno = 0;
    aarg.requestor = progname ? progname.cstr () : "sfs_connect_t";
}

void
sfs_connect_t::getfd (ref<bool> dest, int fd)
{
  if (*dest)
    return;

  tcpc = NULL;
  if (fd < 0) {
    fail (last_srv_err ? last_srv_err : errno,
	  location << ": " << strerror (errno));
    return;
  }
  if (!setsock (fd))
    return;
  sendconnect ();
}

void
sfs_connect_t::sendconnect ()
{
  cbase = c->call (SFSPROC_CONNECT, &sc->ci, &cres,
		   wrap (this, &sfs_connect_t::getconres, destroyed));
}

void
sfs_connect_t::getconres (ref<bool> dest, enum clnt_stat err)
{
  if (*dest)
    return;

  cbase = NULL;
  if (err == RPC_CANTDECODEARGS && sfs_nextci (&sc->ci))
    sendconnect ();
  else if (err)
    srvfail (EIO, location << ": CONNECT RPC: " << err);
  else
    dogetconres ();
}

bool
sfs_connect_t::dogetconres ()
{
  if (cres.status != SFS_OK && cres.status != SFS_REDIRECT) {
    srvfail (EIO, location << ": " << cres.status);
    return false;
  }

  sc->cres = *cres.reply;
  ptr<const sfs_servinfo_w> si;
  sfs_pathcert *pc = NULL;
  if (cres.status == SFS_REDIRECT) {
    pc = sfs_pathcert::store (*cres.revoke);
    if (!pc) {
      srvfail (EIO, location << ": Server returned garbage revocation");
      return false;
    }
  }
  else {
    si = sfs_servinfo_w::alloc (cres.reply->servinfo);
    if (!si || !si->mkhostid_client (&sc->hostid)) {
      srvfail (EIO, location << ": Server returned garbage hostinfo");
      return false;
    }
    pc = sfs_pathcert::lookup (*cres.reply);
  }

  if (pc) {
    if (str dest = pc->dest ()) {
      sname () = dest;
      return start ();
    }
    else {
      srvfail (ESTALE, ": Public key has been revoked");
      return false;
    }
  }

  sc->servinfo = si;
  if (!si->mkhostid_client (&sc->hostid)) {
    srvfail (EIO, location << ": Server returned garbage hostinfo");
    return false;
  }

  if (local_authd) {
    if (ci5.service != SFS_AUTHSERV) {
      sname () = si->mkpath ();
      check_hostid = true;
      in_addr a;
      a.s_addr = htonl (INADDR_LOOPBACK);
      port = si->get_port ();
      if (port <= 0)
	// XXX - should really look up SRV records if there are any
	port = sfs_defport ? sfs_defport : SFS_PORT;
      return start (a, port);
    }
  }

  if (si->ckpath (sname ()))
    sc->hostid_valid = true;
  else if (check_hostid) {
    srvfail (ENOENT, location << ": Server key does not match hostid");
    return false;
  }
  sc->path = si->mkpath (2, port); 
  sc->dnsname = dnsname;

  if (encrypt)
    docrypt ();
  else
    succeed ();
  return true;
}

void
sfs_connect_t::docrypt ()
{
  assert (encrypt);
  if (sc->encrypting) {
    doauth ();
    return;
  }
  if (!ckey) {
    ckey = sfscrypt.gen (SFS_RABIN, 0, SFS_DECRYPT);
    delaycb (3600, wrap (&ckey_clear));
  }
  cbase = sfs_client_crypt (c, ckey, sc->ci, *cres.reply, sc->servinfo,
			    wrap (this, &sfs_connect_t::cryptcb, destroyed));
}

void
sfs_connect_t::cryptcb (ref<bool> dest, const sfs_hash *sidp)
{
  if (*dest)
    return;

  cbase = NULL;
  if (!sidp) {
    srvfail (EIO, location << ": Session key negotiation failed");
    return;
  }
  sc->sessid = *sidp;
  sfs_get_authid (&sc->authid,
		  (sc->ci.civers == 4 ? sc->ci.ci4->service
		   : sc->ci.ci5->service),
		  sc->servinfo->get_hostname (), &sc->hostid, 
		  &sc->sessid, &sc->authinfo);
  sc->encrypting = true;
  doauth ();
}

void
sfs_connect_t::doauth ()
{
  if (!authorizer) {
    succeed ();
    return;
  }

  // XXX - remove once compatibility no longer an issue
  if (sc->servinfo->get_relno () < 8)
    authorizer->mutual = false;

  if (!marg) {
    aarg.authinfo = sc->authinfo;
    aarg.server_release = sc->servinfo->get_relno ();
    marg.alloc ();
    ares.alloc ();
    if (authorizer->mutual)
      lres.alloc ();
    else
      olres.alloc ();
  }
  authorizer->authinit (&aarg, ares,
			wrap (this, &sfs_connect_t::dologin, destroyed));
}

void
sfs_connect_t::dologin (ref<bool> dest)
{
  if (*dest)
    return;
  if (!ares->authenticate) {
    fail (EACCES, "access denied by server");
    return;
  }

  sfs_loginarg larg;
  larg.seqno = aarg.seqno;
  swap (larg.certificate, *ares->certificate);
  if (authorizer->mutual)
    cbase = c->call (SFSPROC_LOGIN, &larg, lres,
		     wrap (this, &sfs_connect_t::donelogin, destroyed));
  else
    cbase = c->call (SFSPROC_LOGIN, &larg, olres,
		     wrap (this, &sfs_connect_t::donelogin, destroyed),
		     NULL, NULL, xdr_sfs_loginres_old);
}

void
sfs_connect_t::donelogin (ref<bool> dest, clnt_stat err)
{
  if (*dest)
    return;

  cbase = NULL;
  if (err) {
    fail (EIO, location << ": LOGIN RPC: " << err);
    return;
  }

  if (!authorizer->mutual) {
    lres.alloc ();
    lres->set_status (olres->status);
    switch (olres->status) {
    case SFSLOGIN_OK:
      sc->auth = authuint_create (*olres->authno);
      succeed ();
      return;
    case SFSLOGIN_MORE:
      *lres->resmore = *olres->resmore;
      break;
    default:
      break;
    }
  }

  switch (lres->status) {
  case SFSLOGIN_OK:
    marg->authinfo = aarg.authinfo;
    marg->seqno = aarg.seqno;
    marg->checkserver = true;
    swap (marg->more, lres->resok->resmore);
    authorizer->authmore (marg, ares,
			  wrap (this, &sfs_connect_t::checkedserver,
				destroyed));
    break;
  case SFSLOGIN_ALLBAD:
    fail (EPERM, "server refuses all authentication");
    break;
  case SFSLOGIN_BAD:
    aarg.ntries++;
    /* cascade */
  case SFSLOGIN_AGAIN:
    doauth ();
    return;
  case SFSLOGIN_MORE:
    marg->authinfo = aarg.authinfo;
    marg->seqno = aarg.seqno;
    marg->checkserver = false;
    swap (marg->more, *lres->resmore);
    authorizer->authmore (marg, ares,
			  wrap (this, &sfs_connect_t::dologin, destroyed));
    break;
  default:
    fail (EINVAL, "bad status in login reply!");
    break;
  }
}

void
sfs_connect_t::checkedserver (ref<bool> d)
{
  if (*d)
    return;
  if (!ares->authenticate)
    fail (EACCES, "could not authenticate server after user login");
  else {
    sc->auth = authuint_create (lres->resok->authno);
    succeed ();
  }
}

void
sfs_connect_cancel (sfs_connect_t *cs)
{
  cs->cancel ();
}

sfs_connect_t *
sfs_connect (const sfs_connectarg &carg, sfs_connect_cb cb,
	     bool encrypt, bool check_hostid)
{
  sfs_connect_t *cs = New sfs_connect_t (cb);
  assert (carg.civers == 5);
  cs->ci5 = *carg.ci5;
  cs->encrypt = encrypt;
  cs->check_hostid = check_hostid;
  if (!cs->start ())
    return NULL;
  return cs;
}

sfs_connect_t *
sfs_connect_path (str path, sfs_service service, sfs_connect_cb cb,
		  bool encrypt, bool check_hostid, sfs_authorizer *a,
		  str user)
{
  sfs_connect_t *cs = New sfs_connect_t (cb);
  cs->sname () = path;
  cs->service () = service;
  cs->encrypt = encrypt;
  cs->check_hostid = check_hostid;
  cs->authorizer = a;
  if (user)
    cs->aarg.user = user;
  if (!cs->start ())
    return NULL;
  return cs;
}

sfs_connect_t *
sfs_connect_host (str host, sfs_service service, sfs_connect_cb cb,
		  bool encrypt)
{
  sfs_connect_t *cs = New sfs_connect_t (cb);
  cs->sname () = host;
  cs->service () = service;
  cs->encrypt = encrypt;
  cs->check_hostid = false;
  if (!cs->start ())
    return NULL;
  return cs;
}

sfs_connect_t *
sfs_connect_crypt (ref<sfscon> sc, sfs_connect_cb cb, sfs_authorizer *a,
		   str user, sfs_seqno seq)
{
  sfs_connect_t *cs = New sfs_connect_t (sc, cb);
  if (a) {
    cs->authorizer = a;
    cs->aarg.seqno = seq;
    if (user)
      cs->aarg.user = user;
  }
  if (!cs->start_scon ())
    return NULL;
  return cs;
}

sfs_pathcert::sfs_pathcert (const sfs_pathrevoke &c, const sfs_hash &h)
  : cert (c), hostid (h)
{
  pathcert_tab.insert (this);
}

sfs_pathcert::~sfs_pathcert ()
{
  pathcert_tab.remove (this);
}

sfs_pathcert *
sfs_pathcert::store (const sfs_pathrevoke &r)
{
  sfs_pathrevoke_w w (r);
  sfs_hash hostid;
  if (!w.check (&hostid))
    return NULL;

  if (sfs_pathcert *pc = pathcert_tab[hostid]) {
    if (!pc->isbetter (r))
      return pc;
    delete pc;
  }
  return New sfs_pathcert (r, hostid);
}

sfs_pathcert *
sfs_pathcert::lookup (const sfs_hash &hostid)
{
  if (sfs_pathcert *pc = pathcert_tab[hostid]) {
    if (pc->valid ())
      return pc;
    delete pc;
  }
  return NULL;
}

sfs_pathcert *
sfs_pathcert::lookup (const sfs_connectok &cr)
{
  if (ptr<sfs_servinfo_w> w = sfs_servinfo_w::alloc (cr.servinfo)) {
    sfs_hash hostid;
    w->mkhostid (&hostid);
    return lookup (hostid);
  }
  return NULL;
}

sfs_reconnect_t::sfs_reconnect_t (sfs_connect_cb c, ref<sfscon> s, bool f,
				  bool e, sfs_authorizer *a, str u)
  : cb (c), sc (s), force (f), encrypt (e), port (0), authorizer (a),
    user (u), areq (NULL), srvreq (NULL), sfscl (NULL)
{
  if (!sfs_parsepath (sc->path, &host, NULL, &port) || !strchr (host, '.')) {
    fail ("bad path in sfs_reconnect");
    return;
  }

  if (!sc->x || sc->x->ateof ())
    force = true;

  if (!port && !isdigit (host[host.len () - 1]))
    srvreq = dns_srvbyname (host, wrap (this, &sfs_reconnect_t::dnscb_srv));
  if (sc->dnsname)
    areq = dns_hostbyname (sc->dnsname,
			   wrap (this, &sfs_reconnect_t::dnscb_a));
  else
    areq = dns_hostbyname (host, wrap (this, &sfs_reconnect_t::dnscb_a));
}

sfs_reconnect_t::~sfs_reconnect_t ()
{
  dnsreq_cancel (areq);
  dnsreq_cancel (srvreq);
  if (sfscl)
    sfscl->cancel ();
}

void
sfs_reconnect_t::dnscb_a (ptr<hostent> hh, int dnserr)
{
  areq = NULL;
  h = hh;
  dnscb (dnserr);
}

void
sfs_reconnect_t::dnscb_srv (ptr<srvlist> s, int dnserr)
{
  srvreq = NULL;
  srvl = s;
  dnscb (dnserr);
}

void
sfs_reconnect_t::dnscb (int dnserr)
{
  if (dnserr && dnserr != NXDOMAIN && dnserr != ARERR_NXREC) {
    fail (dns_strerror (dnserr));
    return;
  }
  if (areq || srvreq)
    return;
  if (!h && !srvl) {
    fail ("host no longer exists in DNS\n");
    return;
  }

  bool use_h = false;
  if (!srvl)
    use_h = true;
  else if (h)
    for (u_int i = 0; !use_h && i < srvl->s_nsrv; i++)
      if (srvl->s_srvs[i].port == sc->port
	  && !strcasecmp (srvl->s_srvs[i].name, sc->dnsname))
	use_h = true;

  sfscl = New sfs_connect_t (wrap (this, &sfs_reconnect_t::connectcb));
  sfscl->sname () = sc->path;
  sfscl->service () = sc->service;
  sfscl->encrypt = encrypt;
  if (authorizer) {
    sfscl->authorizer = authorizer;
    if (user)
      sfscl->aarg.user = user;
  }
  if (!use_h) {
    sfscl->start (srvl);
    return;
  }

  bool use_a = false;
  if (!h)
    use_a = true;
  if (!use_a && sc->sa.sa.sa_family == AF_INET) {
    if (sc->sa.sa_in.sin_addr.s_addr == htonl (INADDR_LOOPBACK))
      use_a = true;
    else
      for (char **ap = h->h_addr_list; !use_a && *ap; ap++)
	if (*reinterpret_cast<in_addr *> (*ap) == sc->sa.sa_in.sin_addr)
	  use_a = true;
  }

  if (!use_a)
    sfscl->start (*reinterpret_cast<in_addr *> (h->h_addr), sc->port);
  else if (!force) {
    /* DNS records haven't changed, so no need to force reconnection */
    sfscl->cancel ();
    connectcb (sc, NULL);
  }
  else if (sc->sa.sa.sa_family == AF_INET)
    sfscl->start (sc->sa.sa_in.sin_addr, sc->port);
  else
    panic ("unknown socket type %d\n", sc->sa.sa.sa_family);
}

void
sfs_reconnect_t::connectcb (ptr<sfscon> nsc, str err)
{
  sfscl = NULL;
  (*cb) (nsc, err);
  delete this;
}

static void
sfs_logindone (ref<sfskey_authorizer> a, sfs_dologin_cb cb,
	       ptr<sfscon> sc, str err)
{
  if (!err && sc && !sc->auth)
    (*cb) ("Login / authentication refused by auth server");
  else
    (*cb) (err);
}
void
sfs_dologin (ref<sfscon> scon, sfspriv *key, int seqno, sfs_dologin_cb cb,
	     bool cred)
{
  ref<sfskey_authorizer> a (New refcounted<sfskey_authorizer>);
  a->setkey (key);
  a->cred = cred;
  sfs_connect_crypt (scon, wrap (sfs_logindone, a, cb), a, NULL, seqno);
}
