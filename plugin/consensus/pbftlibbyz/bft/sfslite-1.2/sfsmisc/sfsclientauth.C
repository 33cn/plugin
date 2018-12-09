/* $Id: sfsclientauth.C 1754 2006-05-19 20:59:19Z max $ */

/*
 *
 * Copyright (C) 1999, 2000 David Mazieres (dm@uun.org)
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

#include "crypt.h"
#include "sfsclient.h"
#include "sfscd_prot.h"
#include "sfscrypt.h"

ptr<sfspriv> sfsserver_auth::privkey;
timecb_t *sfsserver_auth::keytmo;

void
sfsserver_auth::keyexpire ()
{
  keytmo = NULL;
  keygen ();
}

void
sfsserver_auth::keygen ()
{
  privkey = sfscrypt.gen (SFS_RABIN, 0, SFS_DECRYPT);
  if (!keytmo)
    keytmo = delaycb (3536 + (rnd.getword () & 0x3f), 0, wrap (&keyexpire));
}

void
sfsserver_auth::crypt (sfs_connectok cres, ref<const sfs_servinfo_w> si,
		       sfsserver::crypt_cb cb)
{
  if (!privkey)
    keygen ();
  sfs_client_crypt (sfsc, privkey, carg, cres, si, cb, xc);
}

ref<sfsserver_auth::userauth>
sfsserver_auth::userauth_alloc (sfs_aid aid)
{
  ref<userauth> uap (New refcounted<userauth> (aid, mkref (this)));
  uap->sendreq ();
  return uap;
}

bool
sfsserver_auth::authok (nfscall *nc)
{
  sfs_aid aid = nc->getaid ();
  if (sfs_specaid (aid) || authnos[aid])
    return true;
  ptr<userauth> uap = authpending[aid];
  if (!uap)
    uap = userauth_alloc (aid);
  uap->pushreq (nc);
  return false;
}

void
sfsserver_auth::flushstate ()
{
  seqno = 0;
  authnos.clear ();
  authpending.clear ();
}

void
sfsserver_auth::authclear (sfs_aid aid)
{
  if (AUTH *a = authnos[aid]) {
    if (sfsc) {
      u_int32_t authno = authuint_getval (a);
      sfsc->call (SFSPROC_LOGOUT, &authno, NULL, aclnt_cb_null);
    }
    authnos.remove (aid);
  }
  else if (ptr<userauth> uap = authpending[aid]) {
    authpending.remove (aid);
    uap->abort ();
  }
}

sfsserver_auth::userauth::userauth (sfs_aid a, const ref<sfsserver_auth> &s)
  : aid (a), sp (s), cbase (NULL), aborted (false), ntries (0), authno (0)
{
  sp->authpending.insert (aid, mkref (this));
  tmo = delaycb (3, 0, wrap (this, &sfsserver_auth::userauth::timeout));
}

sfsserver_auth::userauth::~userauth ()
{
  abort ();
}

void
sfsserver_auth::userauth::timeout ()
{
  tmo = NULL;
  while (!ncvec.empty ())
    ncvec.pop_front ()->error (NFS3ERR_JUKEBOX);
}

void
sfsserver_auth::userauth::pushreq (nfscall *nc)
{
  if (tmo)
    ncvec.push_back (nc);
  else
    nc->error (NFS3ERR_JUKEBOX);
}

void
sfsserver_auth::userauth::abort ()
{
  if (authno && sp->x && !sp->x->ateof () && sp->sfsc) {
    sp->sfsc->call (SFSPROC_LOGOUT, &authno, NULL, aclnt_cb_null);
    authno = 0;
  }
  if (!aborted) {
    aborted = true;
    if (tmo) {
      timecb_remove (tmo);
      tmo = NULL;
    }
    if (cbase) {
      cbase->cancel ();
      cbase = NULL;
    }
    while (!ncvec.empty ())
      sp->getnfscall (ncvec.pop_front ());
  }
}

void
sfsserver_auth::userauth::sendreq ()
{
  if (!ntries) {
    ++sp->seqno;
    seqno = sp->seqno;
  }
  sfscd_agentreq_arg arg;
  arg.aid = aid;
  arg.agentreq.set_type (AGENTCB_AUTHINIT);
  arg.agentreq.init->ntries = ntries;
  arg.agentreq.init->requestor = "";
  arg.agentreq.init->authinfo = sp->authinfo;
  arg.agentreq.init->seqno = seqno;
  arg.agentreq.init->server_release = sp->si->get_relno ();
  cbase = sp->prog.cdc->call (SFSCDCBPROC_AGENTREQ, &arg, &ares,
			      wrap (this, &sfsserver_auth::userauth::aresult));
}

void
sfsserver_auth::userauth::aresult (clnt_stat err)
{
  cbase = NULL;
  if (err) {
    warn << "sfscd: " << err << "\n";
    finish ();
  }
  else if (!ares.authenticate)
    finish ();
  else {
    sfs_loginarg arg;
    arg.seqno = seqno;
    arg.certificate = *ares.certificate;
    sp->sfsc->call (SFSPROC_LOGIN, &arg, &sres,
		    wrap (mkref (this),
			  &sfsserver_auth::userauth::sresult),
		    NULL, NULL, xdr_sfs_loginres_old);
  }
}

void
sfsserver_auth::userauth::sresult (clnt_stat err)
{
  assert (!cbase);		// Not canceled so we don't leak authnos
  if (err) {
    finish ();
    return;
  }
  if (sres.status == SFSLOGIN_OK)
    authno = *sres.authno;
    // authno = sres.resok->authno;
  if (aborted)
    return;

  switch (sres.status) {
  case SFSLOGIN_OK:
    finish ();
    break;
  case SFSLOGIN_MORE:
    {
      sfscd_agentreq_arg arg;
      arg.aid = aid;
      arg.agentreq.set_type (AGENTCB_AUTHMORE);
      arg.agentreq.more->authinfo = sp->authinfo;
      arg.agentreq.more->seqno = seqno;
      arg.agentreq.more->checkserver = false;
      arg.agentreq.more->more = *sres.resmore;
      cbase = sp->prog.cdc->call (SFSCDCBPROC_AGENTREQ, &arg, &ares,
				  wrap (this,
					&sfsserver_auth::userauth::aresult));
      break;
    }
  case SFSLOGIN_BAD:
    ntries++;
    sendreq ();
    break;
  case SFSLOGIN_ALLBAD:
    finish ();
    break;
  default:
    warn << "userauth: bad status in loginres!\n";
    finish ();
  }
}

void
sfsserver_auth::userauth::finish ()
{
  assert (!aborted);
  auto_auth aa (authuint_create (authno));
  authno = 0;
  sp->authnos.insert (aid, aa);
  sp->authpending.remove (aid);
}
