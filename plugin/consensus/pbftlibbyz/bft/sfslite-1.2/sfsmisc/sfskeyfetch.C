/* $Id: sfskeyfetch.C 1754 2006-05-19 20:59:19Z max $ */

/*
 *
 * Copyright (C) 1999 David Mazieres (dm@uun.org)
 * Copyright (C) 1999 Michael Kaminsky (kaminsky@lcs.mit.edu)
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

#include "sfscrypt.h"
#include "sfsmisc.h"
#include "sfskeymisc.h"
#include "rxx.h"
#include "sfsauth_prot.h"



bool
issrpkey (const str &keyname)
{
  static rxx srprx ("^([^#/@]*@)?[^@,#/]+$");
  return (srprx.search (keyname));
}

bool
iskeyremote (str keyname, bool longkeyok)
{
  if (sfs_parsepath (keyname) && !longkeyok) return false;
  return !strchr (keyname, '/') && !strchr (keyname, '#') && 
    strchr (keyname, '@');
}

bool
parse_userhost (str s, str *user, str *host)
{
  static rxx userhost ("^(([^@]+)@)?([^@]+)$");
  if (user)
    *user = NULL;
  if (host)
    *host = NULL;

  if (!userhost.search (s))
    return false;

  if (user)
    *user = userhost[2];
  if (host)
    *host = userhost[3];
  if (user && !*user && !(*user = myusername ()))
    fatal << "Could not get local username\n";

  return true;
}

static void
keyfetch_srp_cb (ptr<sfscon> *scp, str *errp, ptr<sfscon> sc, str err)
{
  if (sc) {
    *scp = sc;
    *errp = "";
 }
  else
    *errp = err;
}

static str
keyfetch_srp_opaque (sfskey *sk, str keyname, ptr<sfscon> *scp,
		     ptr<sfsauth_certinfores> *cip, bool *warnp)
{

  static bool srpprimed;
  if (!srpprimed) {
    srpprimed = true;
    bigint N, g;
    if (str srpfile = sfsconst_etcfile ("sfs_srp_parms"))
      if (str parms = file2str (srpfile))
	if (import_srp_params (parms, &N, &g) && !srp_base::seedparam (N, g))
	  warn << "DANGER: " << srpfile << " contains bogus parameters!\n";
  }

  ptr<sfscon> sc;
  ptr<aclnt> c;
  srp_client srp;
  str serr;
  str pwd;
  if (sk->pwd) { pwd = sk->pwd; }

  bool serverok = false;
  sfs_connect_srp (sk->keyname, &srp, wrap (keyfetch_srp_cb, &sc, &serr),
		   &sk->keyname, &pwd, &serverok);
  while (!serr)
    acheck ();
  if (serr.len ())
    return serr;

  int vers = sc->servinfo->get_vers ();
  str errstr;

  // version 1
  sfsauth_fetchres fetchres;
  // version 2
  sfsauth2_query_arg aqa;
  sfsauth2_query_res aqr;
  sfsauth_dbkey dbk (SFSAUTH_DBKEY_NAME);
  AUTH *auth = NULL;

  if (vers == 1) {
    c = aclnt::alloc (sc->x, sfsauth_program_1);

    if (clnt_stat err = c->scall (SFSAUTHPROC_FETCH, NULL, &fetchres))
      return sk->keyname << ": fetch: " << err;
    if (fetchres.status != SFSAUTH_OK)
      return sk->keyname << ": server refused key fetch request";

    str tesk = fetchres.resok->privkey;
    // Version 1 armor64's keys sent over the network. Version 2 does not.
    if (!tesk || ! (sk->esk = dearmor64 (tesk))) 
      return sk->keyname << ": could not dearmor or retrieve key";
    sk->pkt = SFS_RABIN;
    
    if (fetchres.resok->hostid != sc->hostid) {
      warnx << sc->path << " returned incorrect hostid in fetch\n";
      if (scp)
	*scp = NULL;
      if (cip)
	*cip = NULL;
      return NULL;
    }

  }
  else { // vers == 2

    c = aclnt::alloc (sc->x, sfsauth_prog_2);
    
    auth = sc->auth;
    *(dbk.name) = sc->user;
    aqa.type = SFSAUTH_USER;
    aqa.key = dbk;
  
    if (clnt_stat err = c->scall (SFSAUTH2_QUERY, &aqa, &aqr, auth))
      return sk->keyname << ": fetch: " << err;
    if (aqr.type == SFSAUTH_ERROR)
      return sk->keyname << ": server refused key fetch request: " 
			 << *(aqr.errmsg) << "\n";
    sk->esk = str (aqr.userinfo->privkey.base (), 
		   aqr.userinfo->privkey.size ());
    sk->pkt = aqr.userinfo->pubkey.type;
  }

  if (sk->esk.len () == 0)
    return sk->keyname << ": server did not return a private key";

  sk->cost = srp.cost;
  sk->pwd = pwd;
  sk->srpparms.alloc ();
  sk->srpparms->N = srp.N;
  sk->srpparms->g = srp.g;
  sk->eksb = New refcounted<eksblowfish> (srp.eksb);

  ptr<sfsauth_certinfores> certinfores = New refcounted<sfsauth_certinfores>;
  
  clnt_stat err = RPC_SUCCESS;  // initialize it to keep gcc happy
  if (vers == 1) { 
    err = c->scall (SFSAUTHPROC_CERTINFO, NULL, certinfores);
  }
  else {  // vers == 2
    aqa.type = SFSAUTH_CERTINFO;
    sfsauth_dbkey nullkey (SFSAUTH_DBKEY_NULL);
    aqa.key = nullkey;
    clnt_stat err = c->scall (SFSAUTH2_QUERY, &aqa, &aqr, auth);
    if (err == RPC_SUCCESS) 
      *certinfores = *(aqr.certinfo);
  }

  if (warnp)
    *warnp = false;
  if (err == RPC_SUCCESS) {
    if (certinfores->name.len () <= 0)
      return sk->keyname << ": server did not return its name/realm\n";
    sc->hostid_valid = (srp.host == certinfores->name);
    if (!sc->hostid_valid) {
      warn << "Warning: host for " << sk->keyname << " is actually server\n"
	   << "        " << sc->path << "\n"
	   << "        This server is claiming to serve host (or realm) "
	   << certinfores->name << ",\n"
	   << "        but you originally registered on host (or in realm) "
	   << srp.host << "\n";
      if (warnp) 
	*warnp = true;
      else {
	sc = NULL;
	certinfores = NULL;
      }
    }
  }
  else {
    certinfores->name = srp.host;
    certinfores->info.set_status (SFSAUTH_CERT_SELF);
  }

  if (scp)
    *scp = sc;
  if (cip)
    *cip = certinfores;
  return NULL;
}

static str
keyfetch_srp (sfskey *sk, str keyname, ptr<sfscon> *scp,
	      ptr<sfsauth_certinfores> *cip, bool *warnp)
{
  ptr<sfscon> sc;
  str r = keyfetch_srp_opaque (sk, keyname, &sc, cip, warnp);
  if (scp)
    *scp = sc;
  if (r)
    return r;
  if (!(sk->key = sfscrypt.alloc (sk->pkt, sk->esk, sk->eksb, sc, 
				  SFS_SIGN))) 
    return sk->keyname << ": cannot decrypt private key returned from "
		       << "server.\n";
  return NULL;
}

str
sfskeyfetch (sfskey *sk, str keyname, ptr<sfscon> *scp,
	     ptr<sfsauth_certinfores> *cip, bool prompt, bool *warnp)
{
  sk->key = NULL;
  sk->keyname = keyname;
  if (prompt) sk->pwd = NULL;
  sk->cost = 0;
  if (scp)
    *scp = NULL;
  if (cip)
    *cip = NULL;

  if (warnp) *warnp = false;
  if (iskeyremote (keyname))
    return keyfetch_srp (sk, keyname, scp, cip, warnp);

  sk->esk = file2wstr (keyname);
  if (!sk->esk)
    return keyname << ": " << strerror (errno);

  if (!sk->keyname || !sk->keyname.len ())
    sk->keyname = keyname;

  sk->key = sfscrypt.alloc (sk->esk, &(sk->keyname), &(sk->pwd), &(sk->cost),
			    SFS_SIGN);
  if (!sk->key) 
    return "Cannot decrypt secret key.";

  return NULL;
}

