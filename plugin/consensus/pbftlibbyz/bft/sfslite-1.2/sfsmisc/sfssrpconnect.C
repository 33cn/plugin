/* $Id: sfssrpconnect.C 3758 2008-11-13 00:36:00Z max $ */

/*
 *
 * Copyright (C) 1999 David Mazieres (dm@uun.org)
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

#include "sfskeymisc.h"
#include "sfsauth_prot.h"
#include "rxx.h"

void
sfssrp_authorizer::authinit (const sfsagent_authinit_arg *argp,
			     sfsagent_auth_res *resp, cbv cb)
{
  resp->set_authenticate (false);
  if (!ntries_ok (argp->ntries)) {
    (*cb) ();
    return;
  }

  sfs_autharg2 aarg (SFS_SRPAUTH);
  if (!reqinit (&aarg.srpauth->req, argp)) {
    (*cb) ();
    return;
  }

  if (!srpc) {
    srpc = New srp_client;
    delete_srpc = true;
  }
  if (!srpc->init (&aarg.srpauth->msg, aarg.srpauth->req.authid,
		   aarg.srpauth->req.user, NULL,
		   // XXX -- eventually change next line to just 6
		   argp->server_release < 8 ? 3 : 6))
    warn ("SRP client initialization failed\n");
  else
    setres (resp, aarg);
  (*cb) ();
}

void
sfssrp_authorizer::authmore (const sfsagent_authmore_arg *argp,
			     sfsagent_auth_res *resp, cbv cb)
{
  resp->set_authenticate (false);

  sfs_autharg2 aarg (SFS_SRPAUTH);
  aarg.srpauth->req.type = SFS_SIGNED_AUTHREQ;
  aarg.srpauth->req.authid = srpc->sessid;
  aarg.srpauth->req.seqno = argp->seqno;
  aarg.srpauth->req.user = srpc->user;

  switch (srpc->next (&aarg.srpauth->msg, &argp->more)) {
  case SRP_SETPWD:
    if (!argp->checkserver) {
      getpwd (strbuf () << "Passphrase for " << srpc->getname () << ": ",
	      false,
	      wrap (this, &sfssrp_authorizer::authmore_2, argp, resp, cb));
      return;
    }
    break;
  case SRP_NEXT:
    if (!argp->checkserver) 
      setres (resp, aarg);
    break;
  case SRP_DONE:
    if (argp->checkserver)
      resp->set_authenticate (true);
    break;
  default:
    break;
  }
  (*cb) ();
}

void
sfssrp_authorizer::authmore_2 (const sfsagent_authmore_arg *argp,
			       sfsagent_auth_res *resp, cbv cb, str pw)
{
  pwd = pw;
  if (!pwd || !pwd.len ()) {
    resp->set_authenticate (false);
    (*cb) ();
  }
  else {
    srpc->setpwd (pwd);
    authmore (argp, resp, cb);
  }
}

static void
sfs_connect_srp_2 (ref<sfssrp_authorizer> a, str *userp, str *pwdp,
		   bool *serverokp, sfs_connect_cb cb, ptr<sfscon> sc, str err)
{
  if (!err && (!a->srpc->host_ok && !serverokp))
    err = "mutual authentication of server failed";
  if (err) {
    if (serverokp)
      *serverokp = false;
    (*cb) (NULL, err);
  }
  else {
    if (serverokp)
      *serverokp = a->srpc->host_ok;
    sc->user = a->srpc->user;
    if (userp)
      *userp = a->srpc->user << "@" << a->srpc->host;
    if (pwdp)
      *pwdp = a->pwd;
    (*cb) (sc, NULL);
  }
}

sfs_connect_t *
sfs_connect_srp (str u, srp_client *srpp, sfs_connect_cb cb,
		 str *userp, str *pwdp, bool *serverokp)
{
  static rxx usrhost ("^([^@]+)?@(.*)$");
  if (!usrhost.match (u)) {
    if (userp)
      *userp = u;
    (*cb) (NULL, "not of form [user]@hostname");
    return NULL;
  }

  str user (usrhost[1]), host (usrhost[2]);
  if (!user && !(user = myusername ())) {
    (*cb) (NULL, "could not get local username");
    return NULL;
  }

  ref<sfssrp_authorizer> a (New refcounted<sfssrp_authorizer>);
  a->srpc = srpp;

  sfs_connect_t *cs
    = New sfs_connect_t (wrap (sfs_connect_srp_2, a, userp, pwdp,
			       serverokp, cb));
  cs->sname () = host;
  cs->service () = SFS_AUTHSERV;
  cs->encrypt = true;
  cs->check_hostid = false;
  cs->authorizer = a;
  cs->aarg.user = user;
  if (!cs->start ())
    return NULL;
  return cs;
}

bool
get_srp_params (ptr<aclnt> c, bigint *Np, bigint *gp)
{
  bool valid = false;
  bigint N,g;
  str srpfile, parms;
  if ((srpfile = sfsconst_etcfile ("sfs_srp_parms")) &&
      (parms = file2str (srpfile)) &&
      import_srp_params (parms, &N, &g) ) {
    if (!srp_base::checkparam (N, g)) {
      warn << "Invalid SRP parameters read from file: "
	   << srpfile << "\n";
    }
    else
      valid = true;
  }

  if (!valid && c) {
    sfsauth2_query_arg aqa;
    sfsauth2_query_res aqr;
    aqa.type = SFSAUTH_SRPPARMS;
    aqa.key.set_type (SFSAUTH_DBKEY_NULL);
    clnt_stat err = c->scall (SFSAUTH2_QUERY, &aqa, &aqr);
    if (!err && aqr.type == SFSAUTH_SRPPARMS &&
	import_srp_params (aqr.srpparms->parms, &N, &g)) {
      if (!srp_base::checkparam (N, g)) {
	warn << "Invalid SRP parameters read from sfsauthd.\n";
	return false;
      } else {
	valid = true;
      }
    }
  }
  if (valid) {
    *Np = N;
    *gp = g;
    return true;
  }
  return false;
}
