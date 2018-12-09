/* $Id: sfsauthorizer.C 1754 2006-05-19 20:59:19Z max $ */

/*
 *
 * Copyright (C) 2004 David Mazieres (dm@uun.org)
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

#include "sfsmisc.h"
#include "sfscrypt.h"
#include "sfssesscrypt.h"
#include "sfskeymisc.h"
#include "rxx.h"

bool
sfs_authorizer::setres (sfsagent_auth_res *resp, const sfs_autharg2 &aarg)
{
  resp->set_authenticate (true);
  if (xdr2bytes (*resp->certificate, aarg))
    return true;
  warn ("sfs_authorizer: could not marshal autharg\n");
  resp->set_authenticate (false);
  return false;
}

bool
sfs_authorizer::reqinit (sfs_authreq2 *reqp, const sfsagent_authinit_arg *ap)
{
  if (!sha1_hashxdr (reqp->authid.base (), ap->authinfo)) {
    warn ("sfs_authorizer: could not marshal authinfo\n");
    return false;
  }
  reqp->type = SFS_SIGNED_AUTHREQ;
  reqp->seqno = ap->seqno;
  reqp->user = ap->user;
  return true;
}

bool
sfs_authorizer::reqinit (sfs_authreq2 *reqp, const sfsagent_authmore_arg *ap)
{
  if (!sha1_hashxdr (reqp->authid.base (), ap->authinfo)) {
    warn ("sfs_authorizer: could not marshal authinfo\n");
    return false;
  }
  reqp->type = SFS_SIGNED_AUTHREQ;
  reqp->seqno = ap->seqno;
  reqp->user = "";		// XXX - but server doesn't check anyway
  return true;
}

void
sfs_authorizer::authmore (const sfsagent_authmore_arg *argp,
			     sfsagent_auth_res *resp, cbv cb)
{
  resp->set_authenticate (argp->checkserver ? true : false);
  (*cb) ();
}

void
sfskey_authorizer::authinit (const sfsagent_authinit_arg *argp,
			     sfsagent_auth_res *resp, cbv cb)
{
  assert (resp);

  if (!ntries_ok (argp->ntries)) {
    resp->set_authenticate (false);
    (*cb) ();
    return;
  }

  if (argp->server_release >= 7 || k->is_v2 ())
    authinit_v2 (argp, resp, cb);
  else
    authinit_v1 (argp, resp, cb);
}

void
sfskey_authorizer::authinit_v1 (const sfsagent_authinit_arg *argp,
				sfsagent_auth_res *resp, cbv cb)
{
  resp->set_authenticate (false);
  if (!cred) {
    (*cb) ();
    return;
  }

  sfs_autharg2 aarg (SFS_AUTHREQ);
  if (!k->export_pubkey (&aarg.authreq1->usrkey)) {
    warn ("sfskey_authorizer::authinit_v1: could not export public key\n");
    (*cb) ();
    return;
  }

  sfs_signed_authreq ar;
  ar.type = SFS_SIGNED_AUTHREQ;
  if (!sha1_hashxdr (ar.authid.base (), argp->authinfo)) {
    warn ("sfskey_authorizer::authinit_v1: could not marshal authinfo\n");
    (*cb) ();
    return;
  }
  ar.seqno = argp->seqno;
  bzero (ar.usrinfo.base (), ar.usrinfo.size ());
  if (argp->user.len () <= ar.usrinfo.size ())
    memcpy (ar.usrinfo.base (), argp->user, argp->user.len ());

  str rawar = xdr2str (ar);
  if (!rawar || !k->sign_r (&aarg.authreq1->signed_req, rawar)) {
    warn ("sfskey_authorizer::authinit_v1: could not sign\n");
    (*cb) ();
    return;
  }

  resp->set_authenticate (true);
  if (!xdr2bytes (*resp->certificate, aarg)) {
    warn << "sfskey_authorizer::authinit_v1: could not marshal sfs_autharg2\n";
    resp->set_authenticate (false);
  }
  (*cb) ();
}

void
sfskey_authorizer::authinit_v2 (const sfsagent_authinit_arg *argp,
				sfsagent_auth_res *resp, cbv cb)
{
  resp->set_authenticate (false);
  if (!ntries_ok (argp->ntries)) {
    (*cb) ();
    return;
  }

  ref<sfs_autharg2> aarg (New refcounted<sfs_autharg2> (SFS_AUTHREQ2));
  if (!reqinit (&aarg->sigauth->req, argp)) {
    (*cb) ();
    return;
  }
  if (!cred)
    aarg->sigauth->req.type = SFS_SIGNED_AUTHREQ_NOCRED;
  if (!k->export_pubkey (&aarg->sigauth->key)) {
    warn ("sfskey_authorizer::authinit could export public key\n");
    resp->set_authenticate (false);
    (*cb) ();
    return;
  }

  sfsauth2_sigreq sr (SFS_SIGNED_AUTHREQ);
  *sr.authreq = aarg->sigauth->req;
  k->sign (sr, argp->authinfo,
	   wrap (this, &sfskey_authorizer::sigdone, aarg, resp, cb));
}

void
sfskey_authorizer::sigdone (ref<sfs_autharg2> aarg, sfsagent_auth_res *resp,
			    cbv cb, str err, ptr<sfs_sig2> sig)
{
  if (err) {
    warn << "signature failed at server: " << err << "\n";
    resp->set_authenticate (false);
    (*cb) ();
    return;
  }
  resp->set_authenticate (true);
  aarg->sigauth->sig = *sig;
  if (!xdr2bytes (*resp->certificate, *aarg)) {
    warn << "sfskey_authorizer::sigdone: could not marshal sfs_autharg2\n";
    resp->set_authenticate (false);
  }
  (*cb) ();
}

str
sfspw_authorizer::printable (str msg)
{
  /* Who knows what a bad server might accomplish by sending a prompt
   * with weird control characters to the user's terminal... */
  strbuf sb;
  for (u_int i = 0; i < msg.len (); i++) {
    u_char c = msg[i];
    if (c == 0x7f)
      sb.tosuio ()->copy ("^?", 2);
    else if (c >= ' ')
      sb.tosuio ()->copy (&c, 1);
    else if (c == '\n' || c == '\a')
      sb.tosuio ()->copy (&c, 1);
    else if (c == '\f' || c == '\v')
      sb.tosuio ()->copy ("\n", 1);
    else if (c == '\t')
      sb.tosuio ()->copy (" ", 1);
    else if (c != '\r') {
      sb.tosuio ()->copy ("^", 1);
      c = c + '@';
      sb.tosuio ()->copy (&c, 1);
    }
  }
  return sb;
}

void
sfspw_authorizer::getpwd (str prompt, bool echo, cbs cb)
{
  if (dont_get_pwd)
    (*cb) (NULL);
  else if (!pwd_fds.empty ()) {
    /* We allow users to type a password multiple times in case of
     * mistake.  However, if we read a password from a file descriptor
     * and the password doesn't work, whatever script has handed us
     * the file descriptor is probably going to fail anyway.
     * Subsequently reading the password from the next file descriptor
     * and applying it to this authentication is only going to confuse
     * matters more (and possibly send a password to the wrong place).
     */
    dont_get_pwd = true;
    pipe2str (pwd_fds.pop_front (), wrap (getpwd_2, destroyed, cb));
  }
  else if (echo)
    getkbdline (printable (prompt), &rnd_input,
		wrap (getpwd_2, destroyed, cb));
  else
    getkbdpwd (printable (prompt), &rnd_input,
	       wrap (getpwd_2, destroyed, cb));
}

void
sfsunixpw_authorizer::authinit (const sfsagent_authinit_arg *argp,
				sfsagent_auth_res *resp, cbv cb)
{
  resp->set_authenticate (false);
  if (!ntries_ok (argp->ntries)) {
    (*cb) ();
    return;
  }
  sfs_autharg2 aarg (SFS_UNIXPWAUTH);
  if (reqinit (&aarg.pwauth->req, argp))
    setres (resp, aarg);
  (*cb) ();
}

void
sfsunixpw_authorizer::authmore (const sfsagent_authmore_arg *argp,
				sfsagent_auth_res *resp, cbv cb)
{
  if (argp->checkserver) {
    static rxx badchar ("[^\\w\\d_\\.\\-]");
    unixuser.setbuf (argp->more.base (), argp->more.size ());
    if (!unixuser.len () || badchar.search (unixuser)) {
      warn << "invalid unix username from auth server\n";
      unixuser = NULL;
    }
    resp->set_authenticate (true);
    (*cb) ();
    return;
  }
  resp->set_authenticate (false);

  sfs_unixpwauth_res res;
  if (!bytes2xdr (res, argp->more)) {
    warn ("could not unmarshal sfs_unixpwauth_res\n");
    (*cb) ();
    return;
  }

  ref<sfs_autharg2> aargp (New refcounted<sfs_autharg2> (SFS_UNIXPWAUTH));
  if (!reqinit (&aargp->pwauth->req, argp)) {
    (*cb) ();
    return;
  }

  str prompt = res.prompt;
  if (!prompt.len ())
    prompt = "Password: ";
  getpwd (prompt, res.echo,
	  wrap (this, &sfsunixpw_authorizer::authmore_2, resp, aargp, cb));
}

void
sfsunixpw_authorizer::authmore_2 (sfsagent_auth_res *resp,
				  ref<sfs_autharg2> aargp, cbv cb, str pwd)
{
  if (pwd && pwd.len ()) {
    aargp->pwauth->password = pwd;
    setres (resp, *aargp);
  }
  (*cb) ();
}
