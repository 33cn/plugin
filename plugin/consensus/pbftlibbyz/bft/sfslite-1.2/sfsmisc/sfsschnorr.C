

#include "sfsschnorr.h"
#include "rxx.h"

static rxx semic (";");
static rxx comma (",");

bool
sfs_2schnorr_priv::export_privkey (str *s) const
{
  sfs_2schnorr_priv_xdr sxdr[1];
  if (!export_privkey (&sxdr[0])) return false;
  if (!export_privkey (s, sxdr, 1)) return false;
  return true;
}

bool
sfs_schnorr_pub::verify (const sfs_sig2 &sig, const str &msg, str *e) const 
{
  if (sig.type != SFS_SCHNORR) return false;
  if (!check_keysize (e)) {
    return false;
  }
  bool ret = pubk->verify (msg, sig.schnorr->r, sig.schnorr->s);
  if (!ret && e)
    *e = "signature verification failed";
  return ret;
}

bool
sfs_schnorr_pub::export_pubkey (sfs_pubkey2 *k) const 
{
  k->set_type (SFS_SCHNORR);
  k->schnorr->p = pubk->modulus ();
  k->schnorr->q = pubk->order ();
  k->schnorr->g = pubk->generator ();
  k->schnorr->y = pubk->public_key ();
  return true;
}

bool
sfs_schnorr_pub::export_pubkey (strbuf &b, bool prefix) const 
{
  bigint p, q, g, y;
  p = pubk->modulus ();
  q = pubk->order ();
  g = pubk->generator ();
  y = pubk->public_key ();

  if (prefix) 
    b << keylabel << "," ;
  b << "p=0x" << p.getstr (16) 
    << ",q=0x" << q.getstr (16)
    << ",g=0x" << g.getstr (16) 
    << ",y=0x" << y.getstr (16) ;
  return true;
}

bool
sfs_2schnorr_priv::sign (sfs_sig2 *sig, const str &msg)
{
  assert (get_opt (SFS_SIGN));
  if (!msg) return false;
  if (wsk) {
    sig->set_type (SFS_SCHNORR);
    bigint r, s;
    if (wsk->sign (&r, &s, msg)) {
      sig->schnorr->r = r;
      sig->schnorr->s = s;
      return true;
    }
  }
  if (!sprivk || !privk) return false;
  return schnorr_client::ni_sign (privk, sprivk, sig, msg);
}

bool
sfs_1schnorr_priv::sign (sfs_sig2 *sig, const str &msg)
{
  assert (get_opt (SFS_SIGN));
  sig->set_type (SFS_SCHNORR);
  bigint r, s;
  if (!privk->sign (&r, &s, msg))
    return false;
  sig->schnorr->r = r;
  sig->schnorr->s = s;
  return true;
}

bool
sfs_1schnorr_priv::export_privkey (str *s) const
{
  sfs_1schnorr_priv_export sexp;
  sexp.kt = SFS_1SCHNORR;
  if (!export_privkey (&sexp.privkey))
    return false;
  if (!sha1_hashxdr (&sexp.cksum, sexp.privkey, true))
    return false;
  if (!(*s = xdr2str_pad (sexp, true, 8)))
    return false;
  return true;
}

bool
sfspriv::get_privkey_hash (u_int8_t *buf, const sfs_hash &hostid) const
{
  str s;
  sha1ctx sc;
  if (!export_privkey (&s))
    return false;
  sc.update (s, s.len ());
  sc.update (hostid.base (), hostid.size ());
  sc.final (buf);
  return true;
}

void
sfs_2schnorr_priv::sign (const sfsauth2_sigreq &sreq, sfs_authinfo ainfo, 
			cbsign cb) 
{
  assert (get_opt (SFS_SIGN));
  if (sprivk) {
    str msg = sigreq2str (sreq);
    if (!msg) {
      (*cb) ("Cannot marhsal signature request.", NULL);
      return;
    }
    ptr<sfs_sig2> ret = New refcounted <sfs_sig2> ();
    if (!sign (ret, msg)) {
      (*cb) ("Cannot run non-interactive signature.", NULL);
      return ;
    }
    (*cb) (NULL, ret);
    return;
  }
  if (sclnt) {
    fixup (&ainfo);
    sclnt->sign (sreq, ainfo, cb);
    return;
  }
  if (!hostname || !hostname.len ()) {
    (*cb) ("2-Schnorr key does not have an associated server", NULL);
    return;
  }
  if (connecting) {
    cqueue.push_back (New sign_block_t (sreq, ainfo, cb));
    return ;
  }
  connecting = true;

  sfs_connect_path (hostname, SFS_AUTHSERV, 
		    wrap (this, &sfs_2schnorr_priv::gotcon, sreq, ainfo, cb),
		    true, true);
}

void
sfs_2schnorr_priv::gotcon (sfsauth2_sigreq sreq, sfs_authinfo ainfo, 
			  cbsign cb, ptr<sfscon> sc, str err)
{
  if (!sc) {
    con_finish (cb, 
		strbuf ("Error connecting to " << hostname << ": " << err));
    return;
  }
  scon = sc;
  sclnt = New schnorr_client (privk, get_aclnt (sc->x), this, NULL, uname);
  if (!sclnt) {
    con_finish (cb, "Cannot make new schnorr client");
    return;
  }

#if 0
  ref<sfskey_authorizer> ska (New refcounted<sfskey_authorizer>);
  ska->setkey (this);
  ska->cred = false;
  sfs_connect_crypt (sc, 
		     wrap (this, &sfs_2schnorr_priv::gotlogin,
			   sreq, ainfo, ska, cb),
		     ska);
#else
  sfs_dologin (scon, this, 0, 
	       wrap (this, &sfs_2schnorr_priv::gotlogin, sreq, ainfo, cb),
	       false);
#endif
}

ptr<sfspriv>
sfs_2schnorr_priv::regen () const 
{
  bigint d;
  if (!sprivk)
    return NULL;

  ptr<schnorr_clnt_priv> privk2 = privk->update (&d);
  ptr<schnorr_srv_priv> sprivk2 = sprivk->update (d);

  ptr<sfs_2schnorr_priv> ret =
    New refcounted<sfs_2schnorr_priv, vbase>
    (scon, privk2, sprivk2, wsk, hostname, uname, opts);
  ret->audit = audit;
  return ret;
}

ptr<sfspriv>
sfs_2schnorr_priv::wholekey () const 
{
  if (!wsk)
    return NULL;
  ptr<sfspriv> ret = New refcounted<sfs_1schnorr_priv, vbase> (wsk, opts);
  return ret;
}

ptr<sfspriv>
sfs_2schnorr_priv::update () const
{
  bigint d;
  ptr<schnorr_srv_priv> sp = NULL;
  ptr<schnorr_priv> w = NULL;
  ptr<sfs_2schnorr_priv> ret = 
    New refcounted<sfs_2schnorr_priv, vbase>
    (scon, privk->update (&d), sp, w, hostname, uname, opts);
  strbuf b;
  b << "Updated: " << timestr ();
  ret->audit = b;
  ret->delta = d;
  return ret;
}

void
sfs_2schnorr_priv::gotlogin (sfsauth2_sigreq sreq, sfs_authinfo ainfo,
			     //ref<sfskey_authorizer>,
			     cbsign cb,
			     //ptr<sfscon> nsc,
			     str err)
{
  if (err) {
    strbuf b;
    b << "Error logging into " << hostname << ": " << err;
    con_finish (cb, b);
    return;
  }

  if (!scon) {
    con_finish (cb, "Signature server died in login process");
    return;
  }

  if (!sclnt) {
    sclnt = New schnorr_client (privk, get_aclnt (scon->x), this, scon->auth, 
				uname);
    if (!sclnt) {
      con_finish (cb, "Cannot make new schnorr client");
      return;
    }
  } else {
    sclnt->auth = scon->auth;
  }
  fixup (&ainfo);
  sclnt->sign (sreq, ainfo, cb);
  connecting = false;
  con_finish ();
  return;
}

void
sfs_2schnorr_priv::con_finish (cbsign cb, const str &err)
{
  connecting = false;
  (*cb) (err, NULL);
  con_finish (err);
}

void
sfs_2schnorr_priv::con_finish (const str &err)
{
  sign_block_t *s;
  while ( cqueue.size () && (s = cqueue.pop_front ())) {
    if (err) 
      (*s->cb) (err, NULL);
    else
      sclnt->sign (s->sr, s->ainfo, s->cb);
    delete s;
  }
}

void
sfs_2schnorr_priv::fixup (sfs_authinfo *a)
{
  if (a->type != SFS_AUTHINFO)
    *a = scon->authinfo;
}


ptr<aclnt>
sfs_2schnorr_priv::get_aclnt (ptr<axprt> x)
{
  ptr<aclnt> r = aclnt::alloc (x, sfsauth_prog_2);
  if (!r) return r;
  r->seteofcb (wrap (this, &sfs_2schnorr_priv::eofcb));
  return r;
}

void
sfs_2schnorr_priv::eofcb ()
{
  if (sclnt) delete sclnt;
  sclnt = NULL;
  scon = NULL;
}

bool
sfs_schnorr_pub::check_keysize (str *s) const
{
  return check_keysize (pubk->modulus ().nbits (), pubk->order ().nbits (), s);
}

bool
sfs_schnorr_pub::check_keysize (size_t pbits, size_t qbits, str *s) 
{
  size_t qminsize = 100;
  size_t qmaxsize = 300;
  strbuf b;

  if (qbits && qbits > qmaxsize) {
    b << "2Schnorr q value is " << qbits << " bits but maximum "
      << "allowed size is " << qmaxsize << " bits";
  } else if (qbits && qbits < qminsize) {
    b << "2Schnorr q value is " << qbits << " bits but minimum "
      << "allowed size is " << qminsize << " bits";
  } else if (pbits > sfs_maxdlogsize) {
    b << "2Schnorr p value is " << pbits << " bits but maxiumum "
      << "allowed size is " << sfs_maxdlogsize << " bits";
  } else if (pbits < sfs_mindlogsize) {
    b << "2Schnorr p value is " << pbits << " bits but minimum "
      << "allowed size is " << sfs_mindlogsize << " bits";
  } else {
    return true;
  }
  if (s) *s = b;
  return false;
}

ptr<sfspub>  
sfs_schnorr_alloc::alloc (const sfs_pubkey2 &pk, u_char o) const
{
  if (pk.type != SFS_SCHNORR) 
    return NULL;
  ref<schnorr_pub> pubk = 
    New refcounted<schnorr_pub> (pk.schnorr->p, pk.schnorr->q,
				 pk.schnorr->g, pk.schnorr->y);
  ptr<sfspub> pub = New refcounted<sfs_schnorr_pub, vbase> (pubk, o);
  return pub;
}

ptr<sfspriv> 
sfs_2schnorr_alloc::alloc (const sfs_privkey2_clear &pk, u_char o) const
{
  if (pk.type != SFS_2SCHNORR) return NULL;
  return alloc (*pk.schnorr2, NULL, o);
}

ptr<sfspriv>
sfs_1schnorr_alloc::alloc (const sfs_privkey2_clear &pk, u_char o) const
{
  if (pk.type != SFS_1SCHNORR) return NULL;
  return alloc (*pk.schnorr1, o);
}

ptr<sfspriv>
sfs_1schnorr_alloc::alloc (const sfs_1schnorr_priv_xdr &k, u_char o) const
{
  ptr<schnorr_priv> privk = 
    New refcounted<schnorr_priv> (k.p, k.q, k.g, k.y, k.x);
  ptr<sfspriv> ret = New refcounted<sfs_1schnorr_priv, vbase> (privk, o);
  return ret;
}

ptr<sfspub>
sfs_schnorr_alloc::alloc (const str &s, u_char o) const
{
  bigint p,q,g,y;
  vec<str> kv;
  int num = 4;
  if (split (&kv, comma, s, num) != num) return NULL;
  for (int i = 0; i < num; i++) { if (!kv[i]) return NULL; }
  p = kv[0] + 2;
  q = kv[1] + 2;
  g = kv[2] + 2;
  y = kv[3] + 2;
  if (!p || !q || !g || !y) return NULL;
  ref<schnorr_pub> pubk = New refcounted<schnorr_pub> (p,q,g,y);
  ptr<sfspub> pub = New refcounted<sfs_schnorr_pub, vbase> (pubk, o);
  return pub;
}

ptr<sfspriv>
sfs_1schnorr_alloc::alloc (const str &raw, ptr<sfscon> c, u_char o) const
{
  sfs_1schnorr_priv_export sxdr;
  if (!str2xdr (sxdr, raw)) return NULL;
  if (sxdr.kt != SFS_1SCHNORR) return NULL;
  sfs_hash cksum;
  if (!sha1_hashxdr (&cksum, sxdr.privkey, true))
    return NULL;
  if (memcmp (&cksum, &sxdr.cksum, sizeof (cksum)))
    return NULL;
  return alloc (sxdr.privkey, o);
}

ptr<sfspriv>
sfs_2schnorr_alloc::alloc (const str &raw, ptr<sfscon> c, u_char o) const
{
  sfs_2schnorr_priv_xdr keys[2];
  int nkeys;
  if (!import_priv (raw, keys, &nkeys) || nkeys < 1) 
    return NULL;
  const sfs_2schnorr_priv_xdr &k = keys[0];
  return alloc (k, c, o);
}

ptr<sfspriv>
sfs_2schnorr_alloc::alloc (const sfs_2schnorr_priv_xdr &k, ptr<sfscon> c, 
			   u_char o) const
{
  ptr<schnorr_clnt_priv> privk = 
    New refcounted<schnorr_clnt_priv> (k.p, k.q, k.g, k.y, k.x);
  str hostname (k.hostname.cstr (), k.hostname.len ());
  str username (k.uname.cstr (), k.uname.len ());
  ptr<schnorr_srv_priv> np = NULL;
  ptr<schnorr_priv> w = NULL;
  ptr<sfspriv> ret = 
    New refcounted<sfs_2schnorr_priv, vbase> (c, privk, np, w, hostname, 
					     username, o);
  return ret;
}

ptr<sfspriv>
sfs_2schnorr_alloc::gen (u_int nbits, u_char o) const
{
  str serr;
  if (nbits == 0) nbits = sfs_dlogsize;
  if (!sfs_schnorr_pub::check_keysize (nbits, 0, &serr)) {
    warn << serr << "\n";
    return NULL;
  }
  ptr<schnorr_gen> s = schnorr_gen::rgen (nbits);
  ptr<sfscon> scon = NULL;
  str hostname = NULL;
  str uname = NULL;
  ptr<sfs_2schnorr_priv> ret =
    New refcounted<sfs_2schnorr_priv, vbase> (scon, s->csk, s->ssk, s->wsk,
					      hostname, uname, o);
  strbuf b = "Generated: " << timestr ();
  ret->audit = b;
  return ret;
}

sfs_2schnorr_priv::sfs_2schnorr_priv (ptr<sfscon> c, 
				      ref<schnorr_clnt_priv> cp, 
				      ptr<schnorr_srv_priv> sp, 
				      ptr<schnorr_priv> w, str hn, 
				      str un, u_char o)
  : sfspub (SFS_2SCHNORR, o, "twoschnorr"), sfs_schnorr_pub (cp, o),
    sfspriv (SFS_2SCHNORR, o, "twoschnorr"), scon (c), sprivk (sp), wsk (w),
    hostname (hn), uname (un), delta (0), sclnt (NULL), privk (cp),
    connecting (false)
{
  if (c && hn == c->servinfo->mkpath ()) {
    if (!hn) hn = c->servinfo->mkpath ();
    if (!un) un = c->user;
    sclnt =  New schnorr_client (privk, get_aclnt (c->x), this, c->auth, un);
  }
}

bool
sfs_2schnorr_priv::export_keyhalf (sfsauth_keyhalf *kh, bool *dlt) const
{
  if (sprivk) {
    if (!export_privkey (kh, sprivk))
      return false;
  } else if (delta > 0) {
    kh->set_type (SFSAUTH_KEYHALF_DELTA);
    *kh->delta = delta;
    *dlt = true;
  }
  return true;
}

bool
sfs_2schnorr_alloc::import_priv (const str &raw, sfs_2schnorr_priv_xdr priv[], 
				 int *num)
{
  sfs_2schnorr_priv_export sxdr;
  if (!str2xdr (sxdr, raw)) return false;
  if (!sxdr.kt == SFS_2SCHNORR) return false;
  sfs_hash cksum;
  u_int i;
  for (i = 0; i < sxdr.privkeys.size () ; i++)
    priv[i] = sxdr.privkeys[i];
  *num = i;
  if (i == 0) return false;
  if (!sha1_hashxdr (&cksum, sxdr.privkeys[0], true))
    return false;
  if (memcmp (&cksum, &sxdr.cksum, sizeof (cksum)))
    return false;
  return true;
}

bool
sfs_2schnorr_priv::export_privkey (str *raw, sfs_2schnorr_priv_xdr priv[], 
				  int num)
{
  sfs_2schnorr_priv_export sxdr; 
  sxdr.privkeys.setsize (num);
  sxdr.kt = SFS_2SCHNORR;
  if (num <= 0) return false;
  if (!sha1_hashxdr (&sxdr.cksum, priv[0], true))
    return false;
  for (int i = 0; i < num ; i++)
    sxdr.privkeys[i] = priv[i];
  *raw = xdr2str_pad (sxdr, true, 8); 
  if (!*raw) return false;
  return true;
}

bool
sfs_2schnorr_priv::export_privkey (sfs_2schnorr_priv_xdr *s, 
				  ptr<const schnorr_pub> o, 
				  str hostname, str uname, str audit)
{
  s->p = o->modulus ();
  s->q = o->order ();
  s->g = o->generator ();
  s->y = o->public_key ();
  s->x = o->private_share ();
  if (hostname) s->hostname = hostname;
  if (uname) s->uname = uname;
  if (audit) s->audit = audit;
  return true;
}

bool
sfs_2schnorr_priv::export_privkey (sfsauth_keyhalf *kh, 
				  ptr<const schnorr_srv_priv> s)
{
  kh->set_type (SFSAUTH_KEYHALF_PRIV);
  kh->priv->setsize (1);
  export_privkey (&((*kh->priv)[0]), s);
  return true;
}

bool 
sfs_2schnorr_priv::parse_keyhalf (sfsauth_keyhalf *k, const str &s)
{
  vec<str> keys;
  int nkeys;
  if (strchr (s, ';')) {
    if ((nkeys = split (&keys, semic, s, 2)) != 2) return false;
  } else {
    nkeys = 1;
    keys.push_back (s);
  }

  k->set_type (SFSAUTH_KEYHALF_PRIV);
  k->priv->setsize (nkeys);
  int num = 5;
  for (int i = 0; i < nkeys; i++) {
    bigint p,q,g,y,x;
    vec<str> kv;
  
    if (split (&kv, comma, keys[i], num) != num) return false;
    for (int j = 0; j < num; j++) { if (!kv[j]) return false; }

    p = kv[0] + 2;
    q = kv[1] + 2;
    g = kv[2] + 2;
    y = kv[3] + 2;
    x = kv[4] + 2;
    
    //if (!p || !q || !g || !y || !x) return false;

    (*k->priv)[i].p = p;
    (*k->priv)[i].q = q;
    (*k->priv)[i].g = g;
    (*k->priv)[i].y = y;
    (*k->priv)[i].x = x;
  }
  return true;
}

bool 
sfs_2schnorr_priv::export_keyhalf (const sfsauth_keyhalf &k, strbuf &b)
{
  if (k.type != SFSAUTH_KEYHALF_PRIV)
    return false;
  bool first = true;
  for (u_int i = 0; i < k.priv->size (); i++) {
    if (!first) b << ";";
    b << "p=0x" << (*k.priv)[i].p.getstr (16) << ","
      << "q=0x" << (*k.priv)[i].q.getstr (16) << ","
      << "g=0x" << (*k.priv)[i].g.getstr (16) << ","
      << "y=0x" << (*k.priv)[i].y.getstr (16) << ","
      << "x=0x" << (*k.priv)[i].x.getstr (16);
    first = false;
  }
  return true;
}

void 
schnorr_client::complete_sign (ptr<ephem_key_pair> clnt_rnd, 
			       ptr<sfsauth2_sign_arg> clnt_arg,
			       ptr<sfsauth2_sign_res> srv_res,
			       cbsign cb, clnt_stat stat)
{
  --nsessions;
  for (int i = nsessions; squeue.size () && i < MAX_SCHNORR_SESSIONS; i++) {
    if (sign_block_t *s = squeue.pop_front ()) {
      sign (s->sr, s->ainfo, s->cb);
      delete s;
    }
  }

  if (stat) {
    strbuf b;
    b << stat;
    (*cb) (b, NULL);
    return;
  }
  if (!(srv_res->ok)) {
    (*cb) (*srv_res->errmsg, NULL);
    return;
  }     
  if (clnt_arg->presig.type != SFS_2SCHNORR ||
      srv_res->sig->type != SFS_SCHNORR) {
    (*cb) ("Server returned unrecognized signature half", NULL);
    return;
  }
  str msg = sigreq2str (clnt_arg->req);
  if (!msg) {
    (*cb) ("Could not marshall signature request", NULL);
    return;
  }
  bigint r,s;
  if ((clnt_rnd->public_half () != clnt_arg->presig.schnorr->r) ||
      (!sign_share->complete_signature (&r, &s, msg,
					clnt_arg->presig.schnorr->r, 
					clnt_rnd->private_half (), 
					srv_res->sig->schnorr->r, 
					srv_res->sig->schnorr->s))) {
    (*cb) ("Signature completion failed", NULL);
  }
  else {
    ptr<sfs_sig2> sig = New refcounted<sfs_sig2>;
    sig->set_type (SFS_SCHNORR);
    sig->schnorr->r = r;
    sig->schnorr->s = s;
    (*cb) (NULL, sig);
  }
}

void
schnorr_client::sign (const sfsauth2_sigreq &req,
		      const sfs_authinfo &authinfo, cbsign cb)
{
  if (nsessions >= MAX_SCHNORR_SESSIONS) {
    squeue.push_back (New sign_block_t (req, authinfo, cb));
    return;
  }
  nsessions ++;
  ref<sfsauth2_sign_arg> arg = New refcounted<sfsauth2_sign_arg>;
  ref<sfsauth2_sign_res> res = New refcounted<sfsauth2_sign_res>;

  arg->authinfo = authinfo;
  arg->req = req;
  arg->presig.set_type (SFS_2SCHNORR);
  arg->presig.schnorr->r = ekp->public_half ();
  arg->user = uname;
  arg->pubkeyhash = pkh;
  sign_client->call (SFSAUTH2_SIGN, arg, res, 
		     wrap (this, &schnorr_client::complete_sign, 
			   ekp, arg, res, cb), auth);
  ekp = sign_share->make_ephem_key_pair ();
}


bool 
schnorr_client::ni_sign (ptr<schnorr_clnt_priv> cpriv, 
			 ptr<schnorr_srv_priv> spriv,
			 sfs_sig2 *sig, const str &msg)
{
  assert(cpriv && spriv && sig);
  sig->set_type (SFS_SCHNORR);
  bigint r_srv, s_srv;

  ref<ephem_key_pair> c_ekp = cpriv->make_ephem_key_pair ();
  return spriv->endorse_signature (&r_srv, &s_srv, msg, 
				   c_ekp->public_half ()) &&
    cpriv->complete_signature (&sig->schnorr->r, &sig->schnorr->s, msg, 
			       c_ekp->public_half (), c_ekp->private_half (),
			       r_srv, s_srv); 
}

str 
sigreq2str (const sfsauth2_sigreq &sr)
{
  str ret;
  switch (sr.type) {
  case SFS_SIGNED_AUTHREQ:
    ret = xdr2str (*sr.authreq);
    break;
  case SFS_UPDATEREQ:
    ret = xdr2str (*sr.updatereq);
    break;
  case SFS_NULL:
    ret = str (sr.rnd->base (), sr.rnd->size ());
    break;
  default:
    ret = NULL;
  }
  return ret;
}

void
null_sigreq (sfsauth2_sigreq *sr)
{
  sr->set_type (SFS_NULL);
  rnd.getbytes (sr->rnd->base (), sr->rnd->size ());
}

bool
sigreq_authid_cmp (const sfsauth2_sigreq &sr, const sfs_hash &authid)
{
  switch (sr.type) {
  case SFS_SIGNED_AUTHREQ:
    return (authid == sr.authreq->authid);
  case SFS_UPDATEREQ:
    return (authid == sr.updatereq->authid);
  default:
    return false;
  }
}

int
sfs_schnorr_pub::find (const sfsauth_keyhalf &kh, const sfs_hash &h1)
{
  u_int nkeys = kh.priv->size ();
  u_int i;
  sfs_pubkey2 pub (SFS_SCHNORR);
  for (i = 0; i < nkeys; i++) {
    sfs_hash h2;
    pub.schnorr->p = (*kh.priv)[i].p;
    pub.schnorr->q = (*kh.priv)[i].q;
    pub.schnorr->g = (*kh.priv)[i].g;
    pub.schnorr->y = (*kh.priv)[i].y;
    sha1_hashxdr (h2.base (), pub);
    if (h2 == h1) break;
  }
  if (i == nkeys)
    return -1;
  return int (i);
}

bool
sfs_schnorr_pub::operator== (const sfsauth_keyhalf &kh) const
{
  sfs_hash h;
  if (kh.type != SFSAUTH_KEYHALF_PRIV)
    return false;
  get_pubkey_hash (&h);
  return (kh.type == SFSAUTH_KEYHALF_PRIV && find (kh, h) >= 0);
}

bool 
sfs_2schnorr_priv::export_privkey (sfs_privkey2_clear *k) const
{
  k->set_type (SFS_2SCHNORR);
  return export_privkey (&(*k->schnorr2));
}

bool
sfs_1schnorr_priv::export_privkey (sfs_privkey2_clear *k) const
{
  k->set_type (SFS_1SCHNORR);
  return export_privkey (&(*k->schnorr1));
}

bool
sfs_1schnorr_priv::export_privkey (sfs_1schnorr_priv_xdr *sexp) const
{
  sexp->p = privk->modulus ();
  sexp->q = privk->order ();
  sexp->g = privk->generator ();
  sexp->y = privk->public_key ();
  sexp->x = privk->private_share ();
  return true;
}

str
sfs_2schnorr_priv::get_desc () const
{
  strbuf b;
  b << keylabel << ",y=0x" << privk->public_key ().getstr (16) 
    << "," << hostname ;
  return b;
}

str
sfs_1schnorr_priv::get_desc () const
{
  strbuf b;
  b << keylabel << ",y=0x" << privk->public_key ().getstr (16) ;
  return b;
}
