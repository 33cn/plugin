/* $Id: sfskeymgr.C 3758 2008-11-13 00:36:00Z max $ */

/*
 *
 * Copyright (C) 2002 Maxwell Krohn (max@cs.nyu.edu)
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

#include "parseopt.h"
#include "crypt.h"
#include "sfskeymisc.h"
#include "sfskeymgr.h"
#include "rxx.h"

#include <unistd.h>
#include <sys/types.h>
#include <dirent.h>

static rxx dot ("\\.");
static rxx pound ("#");
static rxx comma (",");

static bool
get_yesno (const str &prompt) 
{
  for (int i = 0; i < 5; i++) {
    str r = getline (prompt, NULL);
    if (!r || r.len () == 0) 
      return false;
    const char *cp = r.cstr ();
    if (*cp == 'Y' || *cp == 'y')
      return true;
    if (*cp == 'N' || *cp == 'n')
      return false;
  }
  return false;
}

static str
dir_split (const str &s, str *d = NULL)
{
  if (d) *d = NULL;
  assert (s && s.len ());
  const char *cp1, *cp2;
  cp2 = NULL;
  cp1 = s.cstr ();
  while ((cp2 = strchr (cp1, '/')))
    cp1 = cp2 + 1;
  u_int len = s.len () - (cp1 - s.cstr ());
  if (len) {
    if (d) *d = substr (s.cstr (), 0, s.len () - len);
    return substr (cp1, 0, len);
  }
  return s;
}

str 
sfskeystore::defkey2 ()
{ 
  init (false);
  if (!dotsfs) 
    return NULL; 
  return strbuf (dotsfs << "/identity"); 
}

str 
sfskeystore::defkey2_readlink ()
{
  init (false);
  char buf[1024];
  str nln;
  str ln = defkey2 ();
  if (!ln) return NULL;
  struct stat sb;
  for (int i = 0; i < 32; i++) {
    if (lstat (ln.cstr (), &sb) < 0)
      return NULL;
    if (!S_ISLNK (sb.st_mode))
      return ln;
    int rc = readlink (ln.cstr (), buf, 1024);
    if (rc <= 0) 
      return NULL;
    nln = str (buf, rc);
    if (nln.cstr () [0] != '/') {
      str dir;
      dir_split (ln, &dir);
      if (!dir) 
	return NULL;
      strbuf sb (dir);
      if (dir.cstr () [dir.len () - 1] != '/')
	sb << "/";
      sb << nln;
      nln = sb;
    }
    ln = nln;
  }
  return NULL;
}

int
sfskeyinfo_proac::cmp (sfskeyinfo *k) const
{
  int d; 
  if (k->kt != SFSKI_PROAC)
    return -1;
  sfskeyinfo_proac *s = reinterpret_cast<sfskeyinfo_proac *> (k);
  if ((d = version - s->version)) return d;
  if ((d = s->host_priority - host_priority)) return d;
  if ((d = strcmp (host.cstr (), s->host.cstr ()))) return d;
  return (privk_version - s->privk_version);
}

sfskeymgr::sfskeymgr (const str &u, u_int32_t o)
  : g (0), N (0), fetching (false), fflerr (false), g_opts (o)
{
  setuser (u);
  ks = New sfskeystore (user, g_opts);
}

void
sfskeymgr::setuser (str u)
{
  sfs_aid aid;
  if ((uid = getuid ()) < 0)
    fatal << "Cannot find user ID\n";
#if 0  /* let the server enforce this -- it's not our problem */
  if (uid != 0 && u && u != myusername ())
    fatal << "Can only use -u flag as root\n";
#endif
  user = u;
  if (uid == 0) {
    if (!user) {
      if ((aid = myaid ())) {
	g_opts |= KM_NOHM;
	uid = aid;
      }
      struct passwd *pw = NULL;
      if (!(pw = getpwuid (uid)))
	fatal << "Cannot find user for uid " << uid << "\n";
      user = pw->pw_name;
    }
    else
      g_opts |= KM_NOHM;
  }
  else if (!user && !(user = myusername ()))
    fatal << "Cannot look up my username\n";
}

bool
sfskeymgr::getsrp (const str &filename, ptr<sfscon> sc)
{
  if (g > 0)
    return true;
  if (filename) {
    str s = file2str (filename);
    if (!s) 
      warn << "Cannot read SRP file: " << filename << "\n";
    else if (import_srp_params (s, &N, &g)) 
      return true;
  }
  ptr<aclnt> c = NULL;
  if (sc)
    c = aclnt::alloc (sc->x, sfsauth_prog_2);
  if (get_srp_params (c, &N, &g))
    return true;
  return false;
}

void
sfskeymgr::update (sfskey *nk, ptr<sfspriv> ok, const str &path,
		   u_int32_t l_opts, km_update_cb cb)
{
  ptr<sfscon> sc = NULL;
  user_host_t uh;
  if (!get_userhost (path, &uh))
    warn << path << ": cannot parse\n";
  ustate *st = New ustate (nk, ok, sc, uh.hostname, uh.user, cb, l_opts);
  login (path, wrap (this, &sfskeymgr::u_gotlogin, st), ok, l_opts);
}


void
sfskeymgr::u_gotlogin (ustate *st, str err, ptr<sfscon> sc, ptr<sfspriv> k)
{
  if (err) {
    done (st, err);
    return;
  }
  st->setcon (sc);
  st->ok = k;
  update (st);
}

void 
sfskeymgr::update (sfskey *nk, ptr<sfspriv> ok, ptr<sfscon> scon,
		   const str &path, const str &user, u_int32_t u_opts,
		   km_update_cb cb)
{
  ustate *st = New ustate (nk, ok, scon, path, user, cb, u_opts);
  update (st);
}

void
sfskeymgr::update (ustate *st)
{
  if (!(st->u_opts & KM_NOSRP)) getcertinfo (st);
  else getuinfo (st);
}

void
sfskeymgr::getcertinfo (ustate *st)
{
  sfsauth2_query_arg aqa;
  aqa.type = SFSAUTH_CERTINFO;
  aqa.key.set_type (SFSAUTH_DBKEY_NULL);
  st->c->call (SFSAUTH2_QUERY, &aqa, &st->aqr,
	       wrap (this, &sfskeymgr::gotcertinfo, st));

}

void
sfskeymgr::gotcertinfo (ustate *st, clnt_stat err)
{
  if (err) {
    done (st, cse2str (err));
    return;
  }
  if (st->aqr.type == SFSAUTH_ERROR) {
    done (st, *st->aqr.errmsg);
    return;
  }
  sfsauth_certinfores certres = *(st->aqr.certinfo);
  if (certres.name.len () <= 0) {
    done (st, "Could not get server's name/realm");
    return;
  }
  if (certres.info.status == SFSAUTH_CERT_REALM)
    warnx << st->hostname << ": authserver is in realm " 
	  << certres.name << "\n";
  st->realm = certres.name;

  getuinfo (st);
}

void
sfskeymgr::getuinfo (ustate *st)
{
  sfsauth2_query_arg aqa;
  aqa.type = SFSAUTH_USER;
  if (st->ok) {
    aqa.key.set_type (SFSAUTH_DBKEY_PUBKEY);
    if (!st->ok->export_pubkey (aqa.key.key)) {
      done (st, "could not export public key");
      return;
    }
  } else if (st->user) {
    aqa.key.set_type (SFSAUTH_DBKEY_NAME);
    *aqa.key.name = st->user;
  } else {
    done (st, "must provide either an oldkey or a user name");
    return;
  }
  st->c->call (SFSAUTH2_QUERY, &aqa, &st->aqr,
	       wrap (this, &sfskeymgr::gotuinfo, st));
}

void
sfskeymgr::gotuinfo (ustate *st, clnt_stat err) 
{
  if (err) {
    done (st, cse2str (err));
    return;
  }
  str rc;
  if ((rc = check_uinfo (st->aqr))) {
    done (st, rc);
    return; 
  }
  if (st->ok && st->aqr.userinfo->vers == 0) {
    done (st, "No account found for user");
    return;
  }
  if (!st->ok && st->aqr.userinfo->vers > 0 && !(st->u_opts & KM_REREG)) {
    done (st, "Account exists; use -f flag to force reregister");
    return;
  }
  if (!setup_uinfo (st))
    return;
  sign_updatereq (st);
}

str
sfskeymgr::check_uinfo (const sfsauth2_query_res &aqr)
{
  if (aqr.type == SFSAUTH_ERROR) {
    return *(aqr.errmsg);
  }
  if (aqr.type != SFSAUTH_USER) {
    return ("Unexpected server response");
  }
  return NULL;
}

void
sfskeymgr::doupdate (ustate *st)
{
  st->c->call (SFSAUTH2_UPDATE, &st->aua, &st->aur,
	       wrap (this, &sfskeymgr::gotres, st), st->scon->auth);
}

bool
sfskeymgr::setup_uinfo (ustate *st)
{
  // Needed for proactive signatures
  if (!(st->u_opts & KM_DLT) && (st->u_opts & KM_CHNGK)) {
    st->nk->key->set_hostname (st->scon->servinfo->mkpath ());
    st->nk->key->set_username (st->user);
  }

  sfsauth_userinfo uinfo = *st->aqr.userinfo;
  if (!(st->u_opts & KM_NOPK)) {
    if (!st->nk->key->export_pubkey (&uinfo.pubkey)) {
      done (st, "cannot export new public key");
      return false;
    }
  }
  if ((st->u_opts & (KM_SKH | KM_DLT))) {
    if (!st->nk->key->export_keyhalf (&uinfo.srvprivkey, &st->delta)) {
      done (st, "cannot export server keyhalf");
      return false;
    }
  } else {
    uinfo.srvprivkey.set_type (SFSAUTH_KEYHALF_NONE);
  }
  
  if (!(st->u_opts & KM_NOSRP)) {
    if (!st->nk->pwd) {
      done (st, "Cannot use SRP without a password");
      return false;
    }
    srp_client srpc;
    if (!(uinfo.pwauth = srpc.create (N, g, st->nk->pwd, st->realm, 
				      st->nk->cost, 0))) {
      done (st, "could not create SRP info");
      return false;
    }
    if (!(st->u_opts & KM_NOESK)) {
      str epk;
      str2wstr (epk);
      if (!st->nk->key->export_privkey (&epk, &srpc.eksb)) {
	done (st, "could not encrypt private key");
	return false;
      }
      uinfo.privkey = epk;
    }
    warnx << st->hostname << ": New SRP key: " << st->user << "@" 
	  << st->realm << "/" << N.nbits () << "\n";
  } 
  uinfo.vers ++;
  st->uinfo = uinfo;
  return true;
}

void
sfskeymgr::sign_updatereq (ustate *st)
{
  sfsauth2_update_arg &ua = st->aua;

  u_int32_t opts = 0;
  if (st->u_opts & KM_KPSRP) opts |= SFSUP_KPSRP;
  if (st->u_opts & KM_KPESK) opts |= SFSUP_KPESK;
  if (st->u_opts & KM_KPPK) opts |= SFSUP_KPPK;
  ua.req.opts = opts;

  ua.req.type = SFS_UPDATEREQ;
  ua.req.authid = st->scon->authid;
  ua.req.rec.set_type (SFSAUTH_USER);
  *ua.req.rec.userinfo = st->uinfo;
  sfsauth2_sigreq sr;
  sr.set_type (SFS_UPDATEREQ);
  *sr.updatereq = ua.req;

  if (st->ok)
    st->sigs++;
  if (!st->delta)
    st->sigs++;

  if (st->ok) {
    ua.authsig.alloc ();
    st->ok->sign (sr, st->scon->authinfo, 
		  wrap (this, &sfskeymgr::gotsig, st, ua.authsig));
  }
  if (!st->delta) {
    ua.newsig.alloc ();
    st->nk->key->sign (sr, st->scon->authinfo,
		       wrap (this, &sfskeymgr::gotsig, st, ua.newsig));
  }
}

void
sfskeymgr::gotsig (ustate *st, sfs_sig2 *target, str err, ptr<sfs_sig2> sig)
{
  if (err)
    st->sigerr = err;
  if (sig)
    *target = *sig;

  if (!(--st->sigs)) {
    if (st->sigerr) {
      done (st, st->sigerr);
    } else {
       doupdate (st);
    }
  }
}

void
sfskeymgr::gotres (ustate *st, clnt_stat err) 
{
  if (err) {
    done (st, cse2str (err), false);
  } else if (!st->aur.ok) {
    done (st, *st->aur.errmsg);
  } else {
    done (st, NULL);
  }
}    

void
sfskeymgr::done (ustate *st, str err, bool noconf)
{
  (*st->cb) (err, noconf);
  delete st;
  return;
}

void
sfskeymgr::done (lstate *st, str err)
{
  if (st->fetching)
    fetching = false;
  blocked_fetch_t *bf;
  while (fqueue.size ()) 
    if ((bf = fqueue.pop_front ())) {
      dologin (bf->st, bf->ki);
      delete bf;
    }

  (*st->cb) (err, st->scon, st->key);
  delete st;
  return;
}

ptr<sfscon>
sfskeymgr::getsrpcon (sfskeyinfo *ki)
{
  if (!ki->remote)
    return NULL;
  user_host_t uh;
  if (!get_userhost (ki->fn (), &uh))
    return NULL;
  key_con_t **k = keycontab[uh.hash];
  if (k)
    return (*k)->con;
  return NULL;
}


bool
sfskeymgr::fetchpub (const user_host_t &uh, fpkcb cb)
{
  key_con_t **k = keycontab[uh.hash];
  if (k) {
    (*cb) (NULL, (*k)->pubkey);
    return true;
  }
  fpkstate_t *fps = New fpkstate_t (uh, cb);
  if (!connect (uh, wrap (this, &sfskeymgr::fetchpub_gotcon, fps))) {
    delete fps;
    return false;
  }
  return true;
}

void
sfskeymgr::done (fpkstate_t *fps, str err)
{
  (*fps->cb) (err, fps->pub);
  delete fps;
  return;
}

void
sfskeymgr::fetchpub_gotcon (fpkstate_t *fps, str err, ptr<sfscon> sc)
{
  if (!sc || err) {
    done (fps, err);
    return;
  }
  fps->scon = sc;
  getpubkey (fps->scon->x, fps->uh.user, 
	     wrap (this, &sfskeymgr::fetchpub_gotpubkey, fps));
}

void
sfskeymgr::fetchpub_gotpubkey (fpkstate_t *fps, str err, ptr<sfspub> p)
{
  if (err || !p) {
    done (fps, err);
    return;
  }
  fps->pub = p;
  insertcon (fps->uh, fps->scon, p);
  done (fps, NULL);
}

void
sfskeymgr::login (const str &hostname, km_login_cb cb,
		  ptr<sfspriv> key, u_int32_t opts)
{
  if (!hostname) {
    (*cb) ("No hostname given", NULL, NULL);
    return;
  }
  user_host_t uh;
  if (!get_userhost (hostname, &uh)) {
    (*cb) (strbuf (hostname) << ": malformed hostname", NULL, NULL);
    return;
  }
  
  key_con_t **k = keycontab[uh.hash];
  if (k && (*k)->key) {
    (*cb) (NULL, (*k)->con, (*k)->key);
    return;
  }

  lstate *ls = New lstate (uh, key, cb, opts);
  if (k && (*k)->con) { // connection was opened previously without login
    gotcon (ls, NULL, (*k)->con);
  }
  else if (!connect (uh, wrap (this, &sfskeymgr::gotcon, ls))) {
    delete ls;
    str err = strbuf () << uh.hostname 
			<< ": full SFS path or SRP connection needed";
    (*cb) (err, NULL, NULL);
  }
}

void
sfskeymgr::connect (const str &h, concb c)
{
  host_t host;
  if (!get_host (h, &host)) {
    (*c) ("Could not parse hostname", NULL);
  }
  else if (!connect (host, c)) {
    (*c) ("No complete SFS path given, and connection is not cached", NULL);
  }
}

bool
sfskeymgr::connect (const host_t &h, concb c)
{
  ptr<sfscon> *con;
  vec<concb> **q;
  bool ret = false;
  if ((con = anoncontab[h.ahash])
      || (h.sfspath && (con = anoncontab[h.sfspath]))) {
    // (*c) (NULL, New refcounted<sfscon> (**con));
    /* XXX - I'm replacing the above line with the following one.  I
     * don't fully understand this, and so it may not be correct.
     * However, the way the authentication code works, if you tried to
     * log in multiple times to copies of the same sfscon (i.e., same
     * file descriptor), there would be no way to coordinate the
     * sequence numbers.  So I don't see how these anon connections
     * can really be re-used anyway.  -dm
     */
    (*c) (NULL, *con);
    ret = true;
  }
  else if ((q = cqueue[h.ahash])
	   || (h.sfspath && (q = cqueue[h.sfspath]))) {
    (*q)->push_back (c);
    ret = true;
  }
  else if (h.sfspath) {
    vec<concb> *v = New vec<concb> ();
    cqueue.insert (h.sfspath, v);
    if (h.sfspath != h.ahash)
      cqueue.insert (h.ahash, v);
    bool local = (h.sfspath == "-");
    sfs_connect_path (h.sfspath, SFS_AUTHSERV,
		      wrap (this, &sfskeymgr::connected, c, h),
		      true, !local);
    ret = true;
  }
  return ret;
}

void
sfskeymgr::connected (concb c, host_t h, ptr<sfscon> con, str err)
{
  if (con) {
    anoncontab.insert (h.ahash, con);
    anoncontab.insert (con->servinfo->mkpath (), con);
    assert (!con->auth);
  }
  
  vec<concb> **q = cqueue[h.sfspath];
  assert (q);
  concb::ptr i;
  while ((*q)->size ()) 
    (* (*q)->pop_front ()) (err, con ? New refcounted<sfscon> (*con) : NULL);

  delete *q;
  cqueue.remove (h.sfspath);
  if (h.sfspath != h.ahash)
    cqueue.remove (h.ahash);
  (*c) (err, con ? New refcounted<sfscon> (*con) : NULL);
}

void
sfskeymgr::gotcon (lstate *ls, str err, ptr<sfscon> sc)
{
  if (!sc || err) {
    done (ls, err);
    return;
  }
  ls->scon = sc;
  ls->c = aclnt::alloc (sc->x, sfs_program_1);
  if (sc->auth) {
    gotlogin (ls, NULL);
  }
  else if (ls->uh.sfspath && ls->uh.sfspath == "-" && (ls->opts & KM_UNX)) {
    unixlogin (ls);
  }
  else if (!ls->key) {
    getpubkey (ls->scon->x, ls->uh.user, 
	       wrap (this, &sfskeymgr::login_gotpubkey, ls));
  }
  else {
    sfs_dologin (sc, ls->key, 0,
		 wrap (this, &sfskeymgr::gotlogin, ls));
  }
}

void
sfskeymgr::getpubkey (ptr<axprt> x, const str &un, fpkcb cb) 
{
  ref<aclnt> c (aclnt::alloc (x, sfsauth_prog_2));
  sfsauth2_query_arg aqa;
  aqa.type = SFSAUTH_USER;
  aqa.key.set_type (SFSAUTH_DBKEY_NAME);
  *aqa.key.name = un;
  ptr<sfsauth2_query_res> aqr = New refcounted<sfsauth2_query_res> ();
  c->call (SFSAUTH2_QUERY, &aqa, aqr, 
	   wrap (this, &sfskeymgr::gotpubkey, aqr, cb));
}

void
sfskeymgr::gotpubkey (ptr<sfsauth2_query_res> r, fpkcb c, clnt_stat err)
{
  if (err) {
    (*c) (cse2str (err), NULL);
    return ;
  }
  str rc;
  if ((rc = check_uinfo (*r))) {
    (*c) (rc, NULL);
    return ;
  }
  ptr<sfspub> pub = sfscrypt.alloc (r->userinfo->pubkey);
  if (!pub) {
    (*c) ("No such public key on server.", NULL);
    return;
  }
  (*c) (NULL, pub);
}

void
sfskeymgr::login_gotpubkey (lstate *ls, str err, ptr<sfspub> p)
{
  if (err) {
    done (ls, err);
    return;
  }
  sfskeyinfo *ki = NULL;
  if (!(ki = ks->getkeys (p))) {
    done (ls, "cannot find public key");
    return ;
  }
  dologin (ls, ki);
}

void
sfskeymgr::dologin (lstate *ls, sfskeyinfo *ki)
{
  str err;
  while (ki && ki->bad) 
    ki = ki->next;
  if (!ki) {
    done (ls, "No suitable keys found");
    return;
  }
  ptr<sfspriv> priv = ki->privk;
  if (!priv) {
    if (fetching && !ls->fetching) {
      fqueue.push_back (New blocked_fetch_t (ls, ki));
      return ;
    }
    fetching = true;
    ls->fetching = true;
    sfskey *k = fetch (ki, &err);
    if (err) {
      delete k;
      done (ls, err);
      return ;
    }
    priv = k->key;
    delete k;
  }
  if (!priv) {
    done (ls, "Internal key tables are corrupted");
    return;
  }
  ls->key = priv;
  sfs_dologin (ls->scon, priv, 0,
	       wrap (this, &sfskeymgr::gotlogin_r, ls, ki));
}

void
sfskeymgr::gotlogin_r (lstate *ls, sfskeyinfo *ki, str err) 
{
  if (!err) {
    ki->privk = ls->key;
    gotlogin (ls, NULL);
    return ;
  }
  ki->bad = true;
  while (ki && ki->bad)
    ki = ki->next;
  if (ki) warn << "Retryable error: ";
  else warn << "Final retry failed: ";
  warn << err << "\n";
  if (err && ki) {
    dologin (ls, ki);
    return;
  }
  done (ls, "All suitable keys failed");
  return ;
}

void
sfskeymgr::gotlogin (lstate *ls, str err)
{
  if (err) {
    done (ls, err);
    return ;
  }
  insertcon (ls->uh, ls->scon, ls->key);
  done (ls, NULL);
}

void
sfskeymgr::unixlogin (lstate *ls, int ntries)
{
  ls->a = New sfsunixpw_authorizer;
  sfs_connect_crypt (ls->scon, wrap (this, &sfskeymgr::gotunixlogin, ls),
		     ls->a, user, ls->seqno);
}

void
sfskeymgr::gotunixlogin (lstate *ls, ptr<sfscon> sc, str err)
{
  /* XXX - don't like this "global" user, but the point is that the
   * username may be different from what you log in as.  For example,
   * on OpenBSD you could run:
   *
   *    sfskey register -f -u dm:skey
   *
   * to get skey authentication, but the username that comes back
   * would be dm.
   */ 
  if (str unixuser = dynamic_cast<sfsunixpw_authorizer *> (ls->a)->unixuser)
    user = unixuser;
  done (ls, err);
}

sfskey *
sfskeystore::fetch (sfskeyinfo *ki, sfskeymgr *km, str *err, str *pwd,
		    u_int32_t l_opts)
{
  assert (ki && err && pwd);
  str keyname = ki->afn ();
  sfskey *k = New sfskey ();
  if (!ki->remote)
    k->pwd = *pwd;
  ptr<sfscon> scon;
  bool changerealm;
  bool *bp = (l_opts & KM_REALM) ? &changerealm : NULL;
  str rc = sfskeyfetch (k, keyname, &scon, NULL, false, bp);
  if (!rc && scon && bp && *bp && !(l_opts & KM_FRC) &&
      !get_yesno ("Accept changes to SRP realm? [y/N] "))
    rc = "Changes to SRP realm not accepted";
  if (rc) {
    if (err) *err = rc;
    return NULL;
  }
  if (!(g_opts & (KM_NODCHK | KM_NOHM))) {
    str dfk = defkey2_readlink ();
    if (dfk) {
      dfk = dir_split (dfk);
      if (ki->fn () == dfk)
	ki->defkey = true;
    }
  }

  *pwd = k->pwd;
  if (scon) {
    if (!km->insert (keyname, scon, k->key, k, err)) 
      return NULL;
    ki->rkt = k->key->is_proac () ? SFSKI_PROAC : SFSKI_STD;
  }
  return k;
}

bool
sfskeymgr::bump_privk_version (sfskeyinfo *ki, u_int32_t l_opts)
{
  bool ret = false;
  if (ki->remote) {
    if (l_opts & (KM_NOSRP | KM_NOESK))
      warn << ki->fn () << ": refusing to update; client share will be lost\n";
    else 
      ret = true;
  } else {
    if (ki->bump_privk_version ())
      ret = ks->check_local (ki, l_opts);
  }
  return ret;
}

bool
sfskeymgr::insert (const str &kn, ptr<sfscon> con, ptr<sfspriv> key, 
		   sfskey *kw, str *err)
{
  user_host_t uh;
  if (!get_userhost (kn, &uh)) {
    *err = strbuf (kn << ": invalid keyname");
    return false;
  }
  insertcon (uh, con, key, kw);
  return true;
}

sfskeyinfo *
sfskeystore::getkeys (ptr<sfspub> pub)
{
  lsdir ();
  str hv = pub->get_pubkey_hash ();
  sfskeyinfo **k;
  sfskeyinfo_proac **s;
  if ((k = pktab_std[hv]))
    return *k;
  else if ((s = pktab_proac[hv]))
    return *s;
  return NULL;
}

void
sfskeystore::prepend (ptr<sfspriv> priv)
{
  lsdir ();
  str hv = priv->get_pubkey_hash ();
  if (!priv->is_proac ()) {
    sfskeyinfo *ki = New sfskeyinfo (priv);
    pktab_std.insert (hv, ki);
  } else {
    sfskeyinfo_proac *ks = New sfskeyinfo_proac (priv);
    sfskeyinfo_proac **k = pktab_proac[hv];
    if (k) ks->next = *k;
    pktab_proac.insert (hv, ks);
  }
}

void
sfskeystore::hashit (ptr<sfspub> pub, sfskeyinfo *k_new)
{
  str hv = pub->get_pubkey_hash ();
  if (k_new->kt == SFSKI_STD) {
    pktab_std.insert (hv, k_new);
    return;
  } else if (k_new->kt != SFSKI_PROAC) {
    return;
  }
  sfskeyinfo_proac *ks = reinterpret_cast<sfskeyinfo_proac *> (k_new);
  sfskeyinfo_proac **k = pktab_proac [hv];
  if (!k) {
    pktab_proac.insert (hv, ks);
    return;
  }

  sfskeyinfo_proac *pp = NULL;
  for (sfskeyinfo_proac *p = *k; p; 
       p = reinterpret_cast<sfskeyinfo_proac *> (p->next)) {
    if (ks->host_priority < p->host_priority ||
	(ks->host_priority == p->host_priority && !(ks->host == p->host))) {
      ks->next = p;
      if (pp) pp->next = ks;
      else *k = ks;
      break;
    } else if (ks->host_priority == p->host_priority) {
      if (ks->privk_version > p->privk_version) {
	ks->next = p->next;
	if (pp) pp->next = ks;
	else *k = ks;
      }
      break;
    } else if (!p->next) {
      p->next = ks;
      break;
    }
    pp = p;
  }
}

void
sfskeystore::init (bool needkeysdir)
{
  if (!(g_opts & KM_NOHM)) {
    if (g_opts & KM_NOCRT) {
      if (needkeysdir) {
	agent_ckdir (true);
      } else if (!init_ck) {
	agent_ckdir (false);
	init_ck = true;
      }
    } else {
      agent_mkdir ();
    }
    dir = userkeysdir;
  }
}

bool
sfskeystore::lsdir ()
{
  if (read)
    return true;
  if (g_opts & KM_NOHM)
    return true;
  init (true);
  if (!dir) 
    fatal << "No keys directory found ($HOME/.sfs/authkeys/)\n";

  DIR *dp;
  struct dirent *dip;
  
  dp = opendir (dir);
  if (!dp)
    fatal << dir << ": Cannot open directory.\n";

  str dfk = defkey2_readlink ();
  if (dfk)
    dfk = dir_split (dfk);

  sfskeyinfo *k;
  while ((dip = readdir (dp))) {
    if (!strcmp (dip->d_name, ".") || !strcmp (dip->d_name, ".."))
      continue;
    k = sfskeyinfo::alloc (str (dip->d_name), dir);
    str file = file2str (k->afn ());
    if (file) {
      ptr<sfspub> pub = sfscrypt.alloc_from_priv (file);
      if (pub) 
	hashit (pub, k);
    }
    if (k) {
      if (dfk && k->fn () == dfk)
	k->defkey = true;
      k->exists = true;
      ls[k->kt].push_back (k);
    }
  }
  closedir (dp);
  read = true;
  return true;
}

sfskeyinfo *
sfskeystore::generate (sfskey *k, u_int32_t l_opts)
{
  lsdir ();
  str kn = k->keyname;
  str knp = kn;
  if (kn) {
    char *cp = strchr (k->keyname, '#');
    if (cp) 
      kn = substr (kn, 0, cp - k->keyname - 1);
    knp = strbuf (kn << "#");
  }
  ptr<sfspriv> proac = k->key;
  sfski_type kt = proac->is_proac () ? SFSKI_PROAC : SFSKI_STD;
  sfskeyinfo *ki = NULL;
  sfskeyinfo_proac *s = NULL;
  str path, hn;
  if (kt != SFSKI_PROAC || 
      ((ki = getkeys (proac)) && ki->kt != SFSKI_PROAC) ||
      !(path = proac->get_hostname ()) || 
      !sfs_parsepath (path, &hn, NULL, NULL, NULL) || !hn)
    return generate (knp, kt, false);
  
  sfskeyinfo_proac *sret = NULL;
  int hpmax = 0;
  for ( ; ki ; ki = ki->next) {
    s = reinterpret_cast<sfskeyinfo_proac *> (ki);
    if (s->host_priority > hpmax)
      hpmax = s->host_priority;
    if (s->host == hn) {
      sret = New sfskeyinfo_proac (*s);
      sret->kn = kn;
      sret->privk_version ++;
      break;
    }
  }
  sfskeyinfo *ret = NULL;
  if (!sret) {
    if (s) {
      sret = New sfskeyinfo_proac (*s);
      sret->kn = kn;
      sret->host = hn;
      sret->privk_version = 1;
      sret->host_priority = hpmax + 1;
      ret = sret;
    } else {
      ret = generate (knp, kt, false);
      ret->set_hostname (hn);
    }
  } else {
    ret = sret;
  }
  if (!ret)
    return NULL;
  if (!check_local (ret, l_opts)) {
    warn << ret->afn () << ": refusing to overwite\n";
    return NULL;
  }
  return ret;
}

sfskeyinfo *
sfskeystore::generate (str raw, sfski_type xt, bool kcomplete)
{
  lsdir ();
  bool exists = false;
  if (!raw) {
    str hn = sfshostname ();  // XXX - this should be more robust 
    if (!hn) {
      warn << "Could not fetch local hostname\n";
      return NULL;
    }
    raw = strbuf (user << "@" << hn);
  } else if (xt != SFSKI_NONE && raw.cstr () [raw.len () - 1] == '#') {
    raw = substr (raw, 0, raw.len () - 1);
  }
  sfskeyinfo *m = NULL;
  if (!(g_opts & KM_NOHM)) {
    const vec<sfskeyinfo *> &k = ls [xt];
    for (u_int i = 0; i < k.size (); i++) {
      if (k[i]->cmp (raw, kcomplete)) continue;
      if (kcomplete) {
	exists = true;
	break;
      }
      if (!m || k[i]->gcmp (*m) > 0) m = k[i];
    }
  }
  sfskeyinfo *r = sfskeyinfo::alloc (raw, dir, xt, m);
  r->exists = exists;
  r->gen = true;
  return r;
}

sfskeyinfo *
sfskeystore::search (const str &raw, bool kcomplete)
{
  bool mmk = false;
  lsdir ();
  sfskeyinfo *p = NULL, *tp = NULL;
  for (int i = 0; i <= SFSKI_PROAC; i++) {
    tp = search (raw, sfski_type (i), kcomplete);
    if (p && tp) {
      warn << raw << ": multiple matching keys found\n";
      mmk = true;
    }
    if (tp) p = tp;
  }
  if (mmk)
    warn << "using key: " << p->fn () << "\n";
  return p;
}

sfskeyinfo *
sfskeystore::search (const str &raw, sfski_type kt, bool kcomplete)
{
  sfskeyinfo *m = NULL;
  if (g_opts & KM_NOHM) return m;
  const vec<sfskeyinfo *> &k = ls [kt];
  for (u_int i = 0; i < k.size (); i++) {
    if (k[i]->cmp (raw, kcomplete)) continue;
    if (!m || k[i]->cmp (m) > 0) m = k[i];
    if (kcomplete && m) break;
  }
  return m;
}

bool
sfskeyinfo_proac::set_hostname (const str &s)
{
  // XXX - should be more robust
  if (s) {
    host = s;
  } else if (!(host = sfshostname ())) {
    warn << "Cannot get my hostname\n";
    return false;
  }
  return true;
}

sfskeyinfo *
sfskeyinfo::alloc (const str &raw, str dir, sfski_type kt, sfskeyinfo *m)
{
  int version = (m ? m->version + 1 : 1);
  if (m && !dir) dir = m->dir;
  switch (kt) {
  case SFSKI_STD:
    return New sfskeyinfo_std (raw, dir, version);
  case SFSKI_PROAC:
    {
      sfskeyinfo_proac *k = New sfskeyinfo_proac (raw, dir, version);
      if (!k->set_hostname ()) {
	delete k;
	k = NULL;
      }
      return k;
    }
  default:
    return New sfskeyinfo (raw, dir, kt, version);
  }
}

sfskeyinfo *
sfskeyinfo::alloc (const str &raw, str dir)
{
  assert (raw && raw.len ());
  str file;
  if (!dir)
    file = dir_split (raw, &dir);
  else
    file = raw;
  vec<str> kv1, kv2;
  int c = split (&kv1, pound, file, 2);
  if (c != 2)
    return New sfskeyinfo (file, dir);
  c = split (&kv2, comma, kv1[1], 3);
  sfskeyinfo *ret = NULL;
  if (c == 1) {
    int v;
    if (!convertint (kv2[0], &v))
      ret = New sfskeyinfo (file, dir);
    else
      ret = New sfskeyinfo_std (kv1[0], dir, v);
  } else if (c == 3) {
    int pv, hp, sv;
    vec<str> kv3;
    if (!convertint (kv2[0], &pv) || !convertint (kv2[2], &sv) ||
	split (&kv3, dot, kv2[1], 2) != 2 || !convertint (kv3[0], &hp))
      ret = New sfskeyinfo (file, dir);
    else 
      ret = New sfskeyinfo_proac (kv1[0], dir, pv, kv3[1], hp, sv);
  } else {
    ret = New sfskeyinfo (file, dir);
  }
  return ret;
}

bool
sfskeystore::check_local (sfskeyinfo *ki, u_int32_t l_opts)
{
  int rc = access (ki->afn (), F_OK);
  if (rc < 0 && errno == ENOENT) {
    ki->exists = false;
    ki->gen = true;
  } else if (rc == 0) {
    ki->exists = true;
    if (l_opts & KM_FGEN) { 
      if (!(l_opts & KM_FRC))
	return false;
      rc = access (ki->afn (), W_OK);
      if (rc < 0)
	return false;
    }
  } else {
    return false;
  }
  return true;
}

sfskeyinfo *
sfskeyinfo::alloc (str raw, sfskeystore *ks, u_int32_t l_opts)
{
  sfskeyinfo *ret = NULL;
  sfski_type kt = (l_opts & KM_PROAC) ? SFSKI_PROAC : SFSKI_STD;
  if (!raw || raw == "#") {
    if (l_opts & KM_FGEN)
      return ks->generate (NULL, kt, false);
    str kn = ks->defkey2_readlink ();
    if (kn) {
      ret = sfskeyinfo::alloc (kn);
      if (!ks->check_local (ret, l_opts)) {
	warn << kn << ": refusing to overwrite\n";
	return NULL;
      }
    } else if (l_opts & KM_GEN) {
      ret = ks->generate (NULL, kt, false);
    }
    return ret;
  }
  char *cp = strchr (raw.cstr (), '#');
  char *cp2 = strchr (raw.cstr (), '/');
  char *cp3 = strchr (raw.cstr (), '@');
  if (!cp || cp2 || (l_opts & KM_NOSRC)) {
    if (iskeyremote (raw, l_opts & KM_PKONLY)) {
      if (cp3 && raw.len () && (raw.cstr () + raw.len () - 1 == cp3)) {
	warn << raw << ": illegal keyname\n";
	return NULL;
      }
      if (cp3 && cp3 == raw.cstr ()) 
	raw = strbuf () << ks->getuser () << raw;
      ret = New sfskeyinfo (raw);
      ret->remote = true;
      ret->exists = true;
    } else {
      if (!cp2)
	raw = strbuf () << "./" << raw;
      ret = sfskeyinfo::alloc (raw);
      if (!ks->check_local (ret, l_opts)) {
	warn << raw << ": refusing to overwrite\n";
	return NULL;
      }
    }
  } else {
    bool kcomplete = (cp[1] != 0);
    if (!kcomplete) 
      raw = substr (raw, 0, raw.len () - 1);
    if ((l_opts & KM_FGEN) || 
	(!(ret = ks->search (raw, kcomplete)) && (l_opts & KM_GEN)))
      ret = ks->generate (raw, kt, kcomplete);
  }
  return ret;
}

sfskeyinfo *
sfskeymgr::getkeyinfo (const str &keyname, u_int32_t l_opts)
{
  return sfskeyinfo::alloc (keyname, ks, l_opts);
}

sfskeyinfo *
sfskeymgr::getkeyinfo (sfskey *k, u_int32_t l_opts)
{
  return ks->generate (k, l_opts);
}

sfskeyinfo *
sfskeymgr::getkeyinfo_list (const str &keyname, u_int32_t l_opts)
{
  sfskeyinfo *ki = getkeyinfo (keyname, l_opts);
  if (ki && ki->has_backup_keys ()) {
    sfskeyinfo *nki = ks->fill_list (ki);
    for (sfskeyinfo *p = nki; p; p = p->next)
      if (p->cmp (ki) == 0)
	return nki;
  }
  return ki;
}

sfskey *
sfskeymgr::fetch (sfskeyinfo *ki, str *err, u_int32_t l_opts)
{
  if (!ki->exists) {
    *err = strbuf (ki->afn () << ": cannot find key");
    return NULL;
  }
  user_host_t uh;
  key_con_t **kct = NULL;
  sfskey *k = NULL;
  if (ki->remote && get_userhost (ki->afn (), &uh) && 
      (kct = keycontab[uh.hash]) && (k = (*kct)->kw)) {
    return (k);
  }

  return ks->fetch (ki, this, err, &pwd, l_opts);
}

void
sfskeymgr::fetchpub (sfskeyinfo *ki, fpkcb cb)
{
  str fn = ki->afn ();
  if (!ki->remote) {
    str raw = file2str (fn);
    if (!raw) {
      (*cb) (strbuf () << fn << ": cannot open file", NULL);
      return ;
    }
    ptr<sfspub> k = sfscrypt.alloc_from_priv (raw);
    if (!k) {
      (*cb) (strbuf () << fn << ": could not parse key", NULL);
      return;
    }
    (*cb) (NULL, k);
    return;
  } else {
    user_host_t uh;
    key_con_t **kc;
    if (!get_userhost (fn, &uh))
      (*cb) (strbuf () << fn << ": cannot parse user/host", NULL);
    kc = keycontab[uh.hash];
    if (kc && (*kc)->pubkey)
      (*cb) (NULL, (*kc)->pubkey);
    else if (!fetchpub (uh, cb)) {
      str err;
      sfskey *k = fetch (ki, &err);
      if (!k) (*cb) (err, NULL);
      else (*cb) (NULL, k->key);
    }
  }
}

sfskeyinfo *
sfskeystore::fill_list (sfskeyinfo *ki)
{
  if (!ki->exists)
    return NULL;
  lsdir ();
  str k = file2str (ki->afn ());
  ptr<sfspub> pub;
  if (k && (pub = sfscrypt.alloc_from_priv (k)))
    ki = getkeys (pub);
  return ki;
}

void
sfskeymgr::fetch_all (sfskeyinfo *ki, cbkik cb)
{
  str err;
  sfskey *k = NULL;
  for ( ; ki ; ki = ki->next) {
    if (!ki->exists || ki->bad || ki->flagged)
      continue;
    k = fetch (ki, &err);
    if (err) {
      warn << ki->afn () << ": " << err << "\n";
    } else if (!k) {
      warn << ki->afn () << ": no key returned\n";
    } else {
      ki->flagged = true;
      (*cb) (ki, k);
    }
  }
}

void
sfskeymgr::fetch_from_list (sfskeyinfo *ki, cbk cb) 
{
  str err;
  sfskey *k = NULL;
  while (ki) {
    while (ki && (!ki->exists || ki->bad))
      ki = ki->next;
    if (ki) {
      k = fetch (ki, &err);
      if (err) {
	fflerr = true;
	warn << ki->afn () << ": " << err << "\n";
	k = NULL;
      } else if (!k) {
	fflerr = true;
	warn << ki->afn () << ": no key returned\n";
      } else {
	break;
      }
      ki = ki->next;
    }
  }
  if (!k) {
    (*cb) (NULL);
    return ;
  }
  k->key->init (wrap (this, &sfskeymgr::initcb, ki, k, cb));
}

void
sfskeymgr::initcb (sfskeyinfo *ki, sfskey *k, cbk cb, str err)
{
  if (!err) {
    if (fflerr)
      warn << ki->afn () << ": key initialize succeded\n";
    (*cb) (k);
    return;
  }
  fflerr = true;
  warn << ki->afn () << ": key initialization failed:\n" << err << "\n";
  ki = ki->next;
  fetch_from_list (ki, cb);
}


sfskey *
sfskeymgr::fetch_or_gen (sfskeyinfo *ki, str *errp, u_int nbits, u_int cost,
			 u_int32_t l_opts)
{
  assert (errp);
  if (ki->exists) {
    sfskey *k = fetch (ki, errp);
    return k;
  }
  sfskey *k = New sfskey ();
  if (cost) k->cost = cost;

  bool labprompt;
  if (ki->keylabel) {
    labprompt = false;
    k->keyname = ki->keylabel;
  } else {
    labprompt = !(ki->is_proactive ());
    k->keyname = ki->fn ();
  }
  str kts;
  sfs_keytype kt;
  if (l_opts & KM_PROAC) {
    kts = "2-Schnorr";
    kt = SFS_2SCHNORR;
  } else if (l_opts & KM_ESIGN) {
    kts = "ESign";
    kt = SFS_ESIGN;
  } else {
    kts = "Rabin";
    kt = SFS_RABIN;
  }

  strbuf prompt (ki->fn_std () << " (" << kts << ")" );
  str rc = sfskeygen (k, nbits, prompt, labprompt, (l_opts & KM_NOPWD), 
		      (l_opts & KM_NOKBD), false, kt);
  if (rc) {
    *errp = rc;
    delete k;
    return NULL;
  }
  return k;
}

void
sfskeymgr::insertcon (const user_host_t &uh, ptr<sfscon> con, ptr<sfspub> k)
{
  insertcon (uh, con, New key_con_t (con, k));
}

void
sfskeymgr::insertcon (const user_host_t &uh, ptr<sfscon> con, ptr<sfspriv> k,
		      sfskey *kw)
{
  key_con_t **p = keycontab[uh.hash];
  if (p)  // it's possible to "upgrade" a cached connection
    (*p)->upgrade (k, kw);
  else 
    insertcon (uh, con, New key_con_t (con, k, kw));
}

void
sfskeymgr::insertcon (const user_host_t &uh, ptr<sfscon> con, key_con_t *kct)
{
  keycontab.insert (uh.hash, kct);
  keycontab.insert (strbuf (uh.user) << con->servinfo->mkpath (), kct);
  keycontab.insert (strbuf (uh.user << "@" << con->servinfo->get_hostname ()),
		    kct);
}

ptr<const sfs_servinfo_w>
sfskeymgr::getservinfo ()
{
  if (servinfo)
    return servinfo;

  sfs_connect_path ("-", SFS_AUTHSERV,
		    wrap (this, &sfskeymgr::si_gotcon),
		    true, false);
  checkflag = true;
  while (checkflag)
    acheck ();
  return servinfo;
}

void
sfskeymgr::si_gotcon (ptr<sfscon> s, str err)
{
  if (err)
    warn << "Cannot connect to local authserver: " << err << "\n";
  if (s) 
    servinfo = s->servinfo;
  checkflag = false;
}

bool
sfskeymgr::save (sfskey *k, sfskeyinfo *ki, u_int32_t l_opts)
{
  str hn = k->key->get_hostname ();
  if (hn) {
    user_host_t uh;
    if (!get_userhost (hn, &uh)) {
      warn << ki->fn () << ": cannot set hostname\n";
      return false;
    }
    ki->set_hostname (uh.hostname);
  }
  if (l_opts & KM_CHNGK)
    k->keyname = ki->fn ();
  if (!k->cost)
    k->cost = sfs_pwdcost;
  return ks->save (k, ki, l_opts);
}

bool
sfskeymgr::select (sfskeyinfo *ki, u_int32_t l_opts)
{
  return ks->setlink (ki->afn (), l_opts);
}

bool
sfskeystore::save (sfskey *k, sfskeyinfo *ki, u_int32_t l_opts)
{
  str rc;
  bool ret = true;
  if (ki->gen && (!ki->exists || (l_opts & KM_FRC))) {
    str kn = ki->afn ();
    rc = sfskeysave (kn, k, !(l_opts & KM_FRC));
    if (rc) {
      warn << rc << "\n";
      ret = false;
    } else {
      warnx << "wrote key: " << kn << "\n";
      if (!(l_opts & KM_NOLNK) && !(g_opts & KM_NOHM) && 
	  ki->setlink () && !setlink (kn, l_opts)) 
	return false;
    }
  }
  return ret;
}

bool
sfskeystore::setlink (const str &target, u_int32_t l_opts) 
{
  if (g_opts & KM_NOHM)
    return false;
  struct stat sb;
  str ln = defkey2 ();
  int rc = lstat (ln.cstr (), &sb);
  if (rc < 0 && errno != ENOENT) {
    warn << ln << ": cannot access file\n";
    return false;
  }
    
  if (rc == 0) {
    if (!S_ISLNK (sb.st_mode)) {
      if (S_ISREG (sb.st_mode) && (l_opts & KM_FRCLNK)) {
	if (unlink (ln.cstr ()) < 0) {
	  warn << ln << ": cannot delete\n";
	  return false;
	}
	if (!access (ln.cstr (), F_OK)) {
	  warn << ln << ": delete failed\n";
	  return false;
	}
      } else {
	warn << ln << ": file exists and is not a symlink; not overwriting\n";
	return false;
      }
    } else if (unlink (ln.cstr ()) < 0) {
      warn << ln << ": cannot overwrite\n";
      return false;
    }
  }
  if (symlink (target.cstr (), ln.cstr ())) {
    warn << ln << ": " << strerror (errno);
    return false;
  }
  return true;
}

bool
sfskeymgr::get_host (const str &s, host_t *h)
{
  user_host_t u;
  if (!get_userhost (s, &u))
    return false;
  *h = u;
  return true;
}

bool
sfskeymgr::get_userhost (const str &h, user_host_t *uh)
{
  if (!h || !h.len () || h == "-") {
    uh->user = user;
    uh->hostname = sfshostname ();
    uh->sfspath = "-";
    uh->hash = strbuf (user << "@-");
    uh->ahash = "-";
    return true;
  }

  static rxx re ("^(([^@]*)@)?([^@%,]+)(%(\\d+))?(,([a-zA-Z0-9]+))?$");
  if (!re.search (h))
    return false;
  str u2 = re[2];
  str hostname = re[3];
  str ports = re[5];
  str hostid = re[7];
  int port = SFS_PORT;
  
  assert (hostname);
  if (ports && ports.len ()) 
    if (!convertint (ports, &port))
      return false;
  if (hostname == "-" && (port != SFS_PORT))
    return false;
  if (u2 && u2.len ())
    uh->user = u2;
  else
    uh->user = user;
  uh->hostname = hostname;
  if (hostid) {
    if (port != SFS_PORT)
      uh->sfspath = strbuf ("@" << hostname << "%" << port << "," << hostid);
    else 
      uh->sfspath = strbuf ("@" << hostname << "," << hostid);
  } else {
    if (hostname == "-")
      uh->sfspath = hostname;
    else 
      uh->sfspath = NULL;
  }
  if (port != SFS_PORT) 
    uh->ahash = strbuf (hostname) << "%" << port;
  else 
    uh->ahash = hostname;
  uh->hash = strbuf (uh->user) << "@" << uh->ahash;

  return true;
}

void
sfskeymgr::add_keys (const vec<str> &keys)
{
  sfskeyinfo *ki;
  sfskey *k;
  str errstr;
  for (u_int i = 0; i < keys.size (); i++) {
    if (!(ki = getkeyinfo (keys[i])))
      fatal << keys[i] << ": invalid keyname\n";
    if (!(k = fetch (ki, &errstr))) 
      fatal << keys[i] << ": could not fetch key:\n" << errstr << "\n";
    add (k->key);
  }
}

void
sfskeymgr::check_connect (const vec<str> &servers, cbv cb)
{
  ncb = servers.size ();
  for (u_int i = 0; i < servers.size (); i++) 
    login (servers[i], wrap (this, &sfskeymgr::cc_cb, servers[i], cb));
	   
}

void
sfskeymgr::cc_cb (str srv, cbv cb, str err, ptr<sfscon> con, ptr<sfspriv> k)
{
  if (err) {
    fatal << srv << ": " << err << "\n";
  } else if (!con || !k) {
    fatal << srv << ": Cannot establish connection or login\n";
  }
  if (--ncb == 0)
    (*cb) ();
}

bool
sfskeymgr::add_con (ptr<sfspriv> key, ptr<sfscon> *sc, str *hn)
{
  if (!key->get_coninfo (sc, hn))
    return false;
  str err;
  if (!insert (*hn, *sc, key, NULL, &err)) {
    warn << err << "\n";
    return false;
  }
  return true;
}

str
sfskeyinfo::afn () const
{
  str ret;
  if (dir) {
    u_int dlen = dir.len ();
    const char *d = dir.cstr ();
    u_int i;
    for (i = dlen - 1; i >= 0 && d[i] == '/'; i--) ;
    ret = strbuf (substr (dir, 0, i+1) << "/" << fn ());
  } else {
    ret = fn ();
  }
  return ret;
}

str
seconds2str (sfs_time t)
{
  sfs_time wk = t / (7 * 24 * 60 * 60);
  t = t - (wk * 7 * 24 * 60 * 60);
  sfs_time dy = t / (24 * 60 * 60);
  t = t - (dy * 24 * 60 * 60);
  sfs_time hr = t / (60 * 60);
  t = t - (hr * 60 * 60);
  sfs_time mn = t / 60;
  t = t - (mn * 60);

  strbuf s;
  s.fmt ("%" U64F "dw, %" U64F "dd, %" U64F "dh, %" U64F "dm, "
         "%" U64F "ds", wk, dy, hr, mn, t);
  return s;
}

