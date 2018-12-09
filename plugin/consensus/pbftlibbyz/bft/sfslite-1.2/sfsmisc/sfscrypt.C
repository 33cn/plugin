
#include <sys/time.h>
#include <time.h>
#include "arpc.h"
#include "rxx.h"
#include "sfskeymisc.h"
#include "parseopt.h"
#include "sfscrypt.h"
#include "sfsschnorr.h"

sfscrypt_t sfscrypt;
static rxx comma (",");
static rxx hexrx ("0x[\\da-fA-F]+");

ptr<sfspriv>
sfscrypt_t::alloc (const sfs_privkey2_clear &pk, u_char o) const
{
  sfsca *c = xttab[pk.type];
  if (!c) return NULL;
  return c->alloc (pk, o);
}

ptr<sfspub>
sfscrypt_t::alloc (const sfs_pubkey2 &pk, u_char o) const
{
  sfsca *c = xttab[pk.type];
  if (!c) return NULL;
  return c->alloc (pk, o);
}

ptr<sfspub>
sfscrypt_t::alloc (sfs_keytype xt, const str &pk, u_char o) const
{
  sfsca *c = xttab[xt];
  if (!c) return NULL;
  return c->alloc (pk, o);
}

ptr<sfspub>
sfscrypt_t::alloc (const str &s, u_char o) const
{
  vec<str> kv;
  if (split (&kv, comma, s, 2, true) != 2)
    return NULL;
  sfsca *c = strtab[kv[0]];
  if (!c) return NULL;
  return c->alloc (kv[1], o);
}

ptr<sfspriv>
sfscrypt_t::alloc (sfs_keytype xt, const str &esk, const eksblowfish *eksb,
		   ptr<sfscon> scon, u_char o) const
{
  sfsca *c = xttab[xt];
  str sk = sk_decrypt (esk, eksb);
  ptr<sfspriv> ret = NULL;
  u_int nskt;
  const sfs_keytype *skt = c->get_private_keytypes (&nskt);
  if (skt) {
    for (u_int i = 0; i < nskt; i++) {
      c = xttab[skt[i]];
      if (c && (ret = c->alloc (sk, scon, o))) 
	break;
    }
  } else {
    ret = c->alloc (sk, scon, o);
  }
  return ret;
}

str
sk_decrypt (const str &esk, const eksblowfish *eksb) 
{
  if (!esk || !esk.len () ) return NULL;
  if (esk.len () & (eksb ? 7 : 3)) return NULL;

  wmstr m (esk.len ());
  memcpy (m, esk, m.len ());
  if (eksb) {
    cbc64iv iv (*eksb);
    iv.decipher_bytes (m, m.len ());
  }
  return m;
}

str
sk_encrypt (const str &sk, const eksblowfish *eksb)
{
  wmstr m (sk.len ());
  memcpy (m, sk, m.len ());
  if (eksb) {
    cbc64iv iv (*eksb);
    iv.encipher_bytes (m, m.len ());
  }
  return m;
}

ptr<sfspriv>
sfscrypt_t::alloc_priv (const str &raw, u_char o) const
{
  str salt, sk, pk, kn;
  sfs_keytype xt;
  ptr<sfspriv> ret;
  if (!parse (raw, &xt, &salt, &sk, &pk, &kn)) return NULL;
  if (salt && salt.len ()) return NULL;
  sfsca *c = xttab[xt];
  if (!c) return NULL;
  ret = c->alloc (dearmor64 (sk), NULL, o);
  if (!ret) return NULL;
  if (!verify_sk (ret, xt, pk))  return NULL;
  return ret;
}

bool
sfscrypt_t::verify_sk (ref<sfspriv> k, sfs_keytype xt, const str &pk) const 
{
  if (!pk || !pk.len ()) return false;
  ptr<sfspub> k2 = alloc (xt, pk);
  if (!k2 || !(*k == *k2)) {
    strbuf b;
    k->export_pubkey (b, false);
    warn << "Error:  Public and private keys do not match!\n";
    warn << "(Public key should be " << b << ")\n";
    return false;
  }
  return true;
}

ptr<sfspriv>
sfscrypt_t::gen (sfs_keytype xt, u_int nbits, u_char o) const
{
  sfsca *c = xttab[xt];
  if (!c) return NULL;
  ptr<sfspriv> r = c->gen (nbits, o);
  return r;
}

ptr<sfspub>
sfscrypt_t::alloc (const sfs_pubkey &pk, u_char o) const 
{
  sfsca *c = xttab[SFS_RABIN];
  if (!c) return NULL;
  return c->alloc (pk, o);
}

ptr<sfspub>
sfscrypt_t::alloc_from_priv (const str &raw) const
{
  str salt, sk, pk, kn;
  sfs_keytype xt;
  if (!parse (raw, &xt, &salt, &sk, &pk, &kn))
    return NULL;
  return alloc (xt, pk, (u_char)0);
}

ptr<sfspriv>
sfscrypt_t::alloc (const str &raw, str *kn, str *pwd, u_int *cost, u_char o) 
  const
{
  assert (kn && pwd && cost);
  str salt, sk, pk;
  sfs_keytype xt;
  ptr<sfspriv> ret;
  if (!parse (raw, &xt, &salt, &sk, &pk, kn) || !sk || !sk.len ())
    return NULL;
  sfsca *c = xttab[xt];
  if (!c) return NULL;

  eksblowfish eksb;
  eksblowfish *eksbp = NULL;
  sk = dearmor64 (sk);
  if (salt && salt.len ()) {
    if (!pw_dearmorsalt (cost, NULL, NULL, salt))
      return NULL;
    bool guess = false;
    for (int i = 0; i < 3; i++) {
      if (!*pwd) {
	*pwd = getpwd ("Passphrase for " << *kn << ": ");
	guess = true;
      } else 
	i--;
      eksbp = &eksb;
      pw_crypt (*pwd, salt, SALTBITS, eksbp);
      str dsk = sk_decrypt (sk, eksbp);
      if (dsk && (ret = c->alloc (dsk, NULL, o)))
	break;
      *pwd = NULL;
    }
    if (!ret) {
      warn << "Too many tries.\n";
      return NULL;
    }
  } else {
    ret = c->alloc (sk, NULL, o);
  }
  if (!ret) 
    return NULL;
  if (!verify_sk (ret, xt, pk)) {
    warn << "Secret key corruption encountered.\n";
    return NULL;
  }
  return ret;
}

#define A64STR "[A-Za-z0-9+/]+={0,2}"

bool
sfscrypt_t::parse (const str &raw, sfs_keytype *xt, str *salt, 
		   str *ske, str *pk, str *kn) const
{
  static rxx r_v1 ("^SK(\\d+),(\\d+\\$" A64STR "\\$)?,"
		   "(" A64STR "),([^,]+),(.*)$", "");
  static rxx r_v2 ("^SK(\\d+):(\\d+\\$" A64STR "\\$)?:"
		   "(" A64STR "):(.*?):(.*)$", "");
  rxx *r = NULL;
  if (r_v1.search (raw)) r = &r_v1;
  else if (r_v2.search (raw)) r = &r_v2;
  else return false;
  int id;
  if (!convertint ((*r)[1], &id)) return false;
  if (xt) *xt = (sfs_keytype )id;
  if (salt) { *salt = (*r)[2]; }
  if (ske) { *ske = (*r)[3]; }
  if (pk) { *pk = (*r)[4]; }
  if (kn) { *kn = (*r)[5]; }
  return true;
}

ptr<sfspub>
sfs_rabin_alloc::alloc (const sfs_pubkey &k, u_char o) const
{
  ptr<rabin_pub> rk = New refcounted<rabin_pub> (k);
  ref<sfspub> ret = New refcounted<sfs_rabin_pub, vbase> (rk, o);
  return ret;
}

ptr<sfspriv>
sfs_esign_alloc::gen (u_int nbits, u_char o) const 
{
  str s;
  if (nbits == 0)
    nbits = ESIGN_SCALE(sfs_rsasize);
  if (!sfs_rabin_pub::check_keysize (nbits, &s)) {
    warn << s << "\n";
    return NULL;
  }
  ptr<esign_priv> k = New refcounted<esign_priv> (esign_keygen (nbits));
  ref<sfspriv> ret = New refcounted<sfs_esign_priv, vbase> (k, o);
  return ret;
}

ptr<sfspriv>
sfs_rabin_alloc::gen (u_int nbits, u_char o) const
{
  str s;
  if (nbits == 0)
    nbits = sfs_rsasize;
  if (!sfs_rabin_pub::check_keysize (nbits, &s)) {
    warn << s << "\n";
    return NULL;
  }
  ptr<rabin_priv> k = New refcounted<rabin_priv> (rabin_keygen (nbits));
  ref<sfspriv> ret = New refcounted<sfs_rabin_priv, vbase> (k, o);
    
  return ret;
}

ptr<sfspriv>
sfs_rabin_alloc::alloc (const sfs_privkey2_clear &k, u_char o) const
{
  assert (k.type == ktype);
  ptr<rabin_priv> rpk = New refcounted<rabin_priv> (k.rabin->p, k.rabin->q);
  if (!rpk) return NULL;
  ref<sfspriv> ret = New refcounted<sfs_rabin_priv, vbase> (rpk, o);
  return ret;
}

ptr<sfspub>
sfs_rabin_alloc::alloc (const sfs_pubkey2 &k, u_char o) const
{
  assert (k.type == ktype) ;
  ptr<rabin_pub> rpk = New refcounted<rabin_pub> (*(k.rabin));
  if (!rpk) return NULL;
  ref<sfspub> ret = New refcounted<sfs_rabin_pub, vbase> (rpk, o);
  return ret;
}

ptr<sfspub>
sfs_rabin_alloc::alloc (const str &pk, u_char o) const
{
  if (!hexrx.match (pk)) {
    warn << "Malformed Rabin public key given\n";
    return NULL;
  }
  ptr<rabin_pub> rpk = New refcounted<rabin_pub> (bigint (pk.cstr ()));
  if (!rpk)
    return NULL;
  ref<sfspub> ret = New refcounted<sfs_rabin_pub, vbase> (rpk, o);
  return ret;
}

bool
sfscrypt_t::verify (const sfs_pubkey2 &pk, const sfs_sig2 &sig,
		    const str &msg, str *e) const
{
  ptr<sfspub> p = alloc (pk, SFS_VERIFY);
  if (!p) {
    if (e) *e = "Could not use public key to verify";
    return false;
  }
  bool ret = p->verify (sig, msg, e);
  return ret;
}

bool 
sfspriv::export_keyhalf (sfsauth_keyhalf *kh, bool *dlt) const 
{
  *dlt = false;
  kh->set_type (SFSAUTH_KEYHALF_NONE);
  return true;
}

bool
sfspub::operator== (const sfspub &p) const
{
  sfs_pubkey2 pk;
  return (export_pubkey (&pk) && p == pk);
}

bool
sfspub::operator== (const sfs_pubkey2 &p1) const
{
  sfs_pubkey2 p2;
  return (export_pubkey (&p2) && (xdr2str (p2)  == xdr2str (p1)));
}

bool
sfspub::operator== (const str &b) const
{
  strbuf f;
  return (export_pubkey (f) && (str (f) == b));
}

bool 
sfspub::verify_init (const sfs_sig2 &sig, const str &msg, str *e) const
{
  if (!check_keysize (e)) {
    return false;
  }
  if (sig.type != ktype) {
    if (e) *e = "public key type mismatch!";
    return false;
  }
  return true;
}

bool
sfspriv::export_privkey (str *s, const eksblowfish *eksb) const
{
  str tmp;
  str2wstr (tmp);
  if (!export_privkey (&tmp) || !tmp) 
    return false;
  if (eksb && (tmp.len () & 4))
    return false;
  *s = eksb ? sk_encrypt (tmp, eksb) : tmp;
  return true;
}


bool
sfspriv::export_privkey (str *s, const str &kn, str pwd, u_int cost) const
{
  str salt = "";
  str seckey;

  if (pwd) {
    salt = pw_gensalt (cost);
    if (!salt)
      return NULL;

    eksblowfish eksb;
    pw_crypt (pwd, salt, SALTBITS, &eksb);
    if (!export_privkey (&seckey, &eksb))
      return false;
  } else 
    if (!export_privkey (&seckey, NULL))
      return false;

  return export_privkey (s, salt, seckey, kn);
}

bool 
sfs_rabin_priv::sign (sfs_sig2 *sig, const str &msg) 
{
  assert (get_opt (SFS_SIGN));
  sig->set_type (SFS_RABIN);
  if ((*sig->rabin = privk->sign (msg)) != 0) return true;
  else return false;
}

bool
sfs_esign_priv::sign (sfs_sig2 *sig, const str &msg)
{
  assert (get_opt (SFS_SIGN));
  sig->set_type (SFS_ESIGN);
  if ((*sig->esign = privk->sign (msg)) != 0) return true;
  else return false;
}

bool 
sfs_rabin_pub::verify (const sfs_sig2 &sig, const str &msg, str *e) const
{
  assert (get_opt (SFS_VERIFY));
  if (!verify_init (sig, msg, e)) return false;
  bool ret = pubk->verify (msg, *(sig.rabin));
  if (!ret && e)
    *e = "signature verification failed";
  return ret;
}

bool
sfs_esign_pub::verify (const sfs_sig2 &sig, const str &msg, str *e) const
{
  assert (get_opt (SFS_VERIFY));
  if (!verify_init (sig, msg, e)) return false;
  bool ret = pubk->verify (msg, *(sig.esign));
  if (!ret && e)
    *e = "signature verification failed";
  return ret;
}

bool
sfs_rabin_pub::verify_r (const bigint &n, size_t len, str &msg, str *e) const
{
  assert (get_opt (SFS_VERIFY));
  if (!check_keysize (e))
    return false;
  msg = pubk->verify_r (n, len);
  if (msg)
    return true;
  else {
    if (e) *e = "signature verification failed";
    return false;
  }
}

bool
sfs_rabin_priv::sign_r (sfs_sig *sig, const str &msg) const
{
  assert (get_opt (SFS_SIGN));
  if (!(*sig = privk->sign_r (msg))) return false;
  return true;
}

bool
sfs_rabin_priv::export_privkey (sfs_privkey2_clear *k) const
{
  k->set_type (ktype);
  k->rabin->p = privk->p;
  k->rabin->q = privk->q;
  return true;
}

bool
sfs_esign_priv::export_privkey (sfs_privkey2_clear *k) const
{
  k->set_type (ktype);
  return export_privkey (k->esign);
}

bool
sfs_esign_priv::export_privkey (sfs_esign_priv_xdr *k) const 
{
  k->p = privk->p;
  k->q = privk->q;
  k->k = privk->k;
  return true;
}

bool
sfs_rabin_pub::export_pubkey (sfs_pubkey2 *k) const
{
  k->set_type (ktype);
  *k->rabin = pubk->n;
  return true;
}

bool
sfs_esign_pub::export_pubkey (sfs_pubkey2 *k) const
{
  k->set_type (ktype);
  k->esign->n = pubk->n;
  k->esign->k = pubk->k;
  return true;
} 

bool
sfs_esign_pub::export_pubkey (strbuf &b, bool prefix) const
{
  if (prefix)
    b << keylabel << "," ;
  b << "n=0x" << pubk->n.getstr (16)
    << ",k=0x" << strbuf ("%lx", pubk->k);
  return true;
}

bool
sfspub::check_keysize (size_t nbits, size_t ll, size_t ul, const str &l, 
		       str *s)
{
  strbuf b;
  if (nbits > ul) {
    b << l << " key is " << nbits << " bits but maximum allowed size is " 
      << ul << " bits";
  } else if (nbits < ll) {
    b << l << " key is " << nbits << " bits but minimum allowed size is "
      << ll << " bits";
  } else {
    return true;
  }
  if (s) *s = b;
  return false;
}

bool
sfs_esign_pub::check_keysize (str *s) const
{
  return check_keysize (pubk->n.nbits (), s);
}

bool
sfs_rabin_pub::check_keysize (str *s) const
{
  return check_keysize (pubk->n.nbits (), s);
}

bool
sfs_esign_pub::check_keysize (size_t nbits, str *s)
{
  return sfspub::check_keysize (nbits, ESIGN_SCALE(sfs_minrsasize),
				ESIGN_SCALE(sfs_maxrsasize), "Esign", s);
}

bool 
sfs_rabin_pub::check_keysize (size_t nbits, str *s)
{
  return sfspub::check_keysize (nbits, sfs_minrsasize, sfs_maxrsasize,
				"Rabin", s);
}

bool
sfs_rabin_pub::export_pubkey (strbuf &b, bool prefix) const
{
  if (prefix) b << keylabel << "," ;
  b << "0x" << pubk->n.getstr (16);
  return true;
}

bool
sfs_rabin_pub::encrypt (sfs_ctext2 *ct, const str &msg) const
{
  assert (get_opt (SFS_ENCRYPT));
  ct->set_type (ktype);
  if (!(*ct->rabin = pubk->encrypt (msg))) return false;
  return true;
}

bool
sfs_rabin_priv::decrypt (const sfs_ctext &ct, str *msg) const
{
  assert (get_opt (SFS_DECRYPT));
  if (!ct) return false;
  *msg = privk->decrypt (ct, sizeof (sfs_kmsg));
  if (!*msg) return false;
  return true;
}

bool
sfs_rabin_priv::decrypt (const sfs_ctext2 &ct, str *msg, u_int sz) const
{
  if (ct.type != ktype) return false;
  if (!ct.rabin) return false;
  *msg = privk->decrypt (*ct.rabin, sz);
  if (!*msg) return false;
  return true;
}

bool
sfs_rabin_pub::encrypt (sfs_ctext *n, const str &msg) const
{
  assert (get_opt (SFS_ENCRYPT));
  if (!(*n = pubk->encrypt (msg))) return false;
  return true;
}

bool
sfs_rabin_pub::export_pubkey (sfs_pubkey *k) const 
{
  if (!(*k = pubk->n)) return false;
  return true;
}

bool 
sfs_rabin_priv::get_privkey_hash (u_int8_t *buf, const sfs_hash &hostid) const
{
  str p;
  sha1ctx sc;
  p = str2wstr (privk->p.getraw ());
  sc.update (p, p.len ());
  sc.update (hostid.base (), hostid.size ());
  p = str2wstr (privk->q.getraw ());
  sc.update (p, p.len ());
  sc.final (buf);
  return true;
}

bool
sfs_esign_priv::export_privkey (str *s) const
{
  sfs_esign_priv_export sexp;
  sexp.kt = ktype;
  if (!export_privkey (&sexp.privkey))
    return false;
  if (!sha1_hashxdr (&sexp.cksum, sexp.privkey, true))
    return false;
  if (!(*s = xdr2str_pad (sexp, true, 8)))
    return false;
  return true;
}

bool
sfs_rabin_priv::export_privkey (str *s) const
{
  sfs_rabin_priv_export rexp;
  xdrsuio x;
  if (!xdr_putint (&x, SFS_RABIN) ||
      !xdr_putbigint (&x, privk->n))
    return false;
  sha1_hashv (&rexp.cksum, x.iov (), x.iovcnt ());
  rexp.p = privk->p;
  rexp.q = privk->q;
  if ((*s = xdr2str_pad (rexp, true, 8)))
    return true;
  return false;
}

bool
sfspriv::export_privkey (str *s, const str &salt, const str &sk,
			 const str &kn) const
{
  strbuf b ("SK%d", (int )ktype);
  b << ":" << salt << ":" << armor64 (sk) << ":" ;
  export_pubkey (b, false); 
  b << ":" << kn;
  *s = b;
  return true;
}

void
sfspriv::signcb (str *errp, sfs_sig2 *sigp, str err, ptr<sfs_sig2> sig)
{
  if (err) *errp = err;
  else *errp = ""; 
  if (sig) *sigp = *sig;
}

void
sfspriv::sign (const sfsauth2_sigreq &sr, sfs_authinfo ainfo, cbsign cb)
{
  ptr <sfs_sig2> sig = New refcounted<sfs_sig2> ();
  str msg = sigreq2str (sr);
  if (!msg) {
    (*cb) ("Could not convert sfs_updatereq to string", NULL);
    return;
  }
  bool rc = sign (sig, msg);
  if (!rc) 
    (*cb) ("Synchronous sign failed", NULL);
  else 
    (*cb) (NULL, sig);
}

ptr<sfspub>
sfs_esign_alloc::alloc (const str &s, u_char o) const
{
  static rxx comma (",");
  vec<str> kv;
  int num = 2;
  if (split (&kv, comma, s, num) != num) return NULL;
  for (int i = 0 ; i < num; i++) { if (!kv[i]) return NULL; }
  bigint n;
  n = kv[0] + 2;
  u_int k;
  if (!convertint (kv[1] + 2, &k))
    return NULL;
  if (!n || !k) return NULL;
  ref<esign_pub> pubk = New refcounted<esign_pub> (n, k);
  ptr<sfspub> pub = New refcounted<sfs_esign_pub, vbase> (pubk, o);
  return pub;
}

ptr<sfspub>
sfs_esign_alloc::alloc (const sfs_pubkey2 &k, u_char o) const
{
  assert (k.type == ktype);
  if (k.esign->k <= 4)
    return NULL;
  ptr<esign_pub> epk = New refcounted<esign_pub> (k.esign->n, k.esign->k);
  if (!epk) return NULL;
  ref<sfspub> ret = New refcounted<sfs_esign_pub, vbase> (epk, o);
  return ret;
}

ptr<sfspriv>
sfs_esign_alloc::alloc (const str &raw, ptr<sfscon> dummy, u_char o) const
{
  sfs_esign_priv_export sxdr;
  if (!str2xdr (sxdr, raw)) return NULL;
  if (sxdr.kt != SFS_ESIGN) return NULL;
  sfs_hash cksum;
  if (!sha1_hashxdr (&cksum, sxdr.privkey, true)) return NULL;
  if (memcmp (&cksum, &sxdr.cksum, sizeof (cksum))) return NULL;
  return alloc (sxdr.privkey, o);
}

ptr<sfspriv>
sfs_esign_alloc::alloc (const sfs_esign_priv_xdr &k, u_char o) const
{
  ptr<esign_priv> privk = New refcounted<esign_priv> (k.p, k.q, k.k);
  ptr<sfspriv> ret = New refcounted<sfs_esign_priv, vbase> (privk, o);
  return ret;
}

ptr<sfspriv>
sfs_esign_alloc::alloc (const sfs_privkey2_clear &pk, u_char o) const
{
  if (pk.type != SFS_ESIGN) return NULL;
  return (alloc (*pk.esign, o));
}

ptr<sfspriv>
sfs_rabin_alloc::alloc (const str &raw, ptr<sfscon> dummy, u_char o) const
{
  u_char h[sha1::hashsize];
  sfs_rabin_priv_export rexp;
  if (!str2xdr (rexp, raw) || rexp.p >= rexp.q || rexp.p <= 1 || rexp.q <= 1) 
    return NULL;
  ref<rabin_priv> rsk = New refcounted<rabin_priv> (rexp.p, rexp.q);
  xdrsuio x;
  if (!xdr_putint (&x, SFS_RABIN) || !xdr_putbigint (&x, rsk->n))
    return NULL;
  sha1_hashv (h, x.iov (), x.iovcnt ());
  if (memcmp (h, &rexp.cksum, sizeof (h)))
    return NULL;
  ref<sfspriv> ret = New refcounted<sfs_rabin_priv, vbase> (rsk, o);
  return ret;
}

sfscrypt_t::sfscrypt_t () 
{
  add (New sfs_rabin_alloc ());
  add (New sfs_schnorr_alloc ());
  add (New sfs_1schnorr_alloc ());
  add (New sfs_2schnorr_alloc ());
  add (New sfs_esign_alloc ());
}

str
timestr ()
{
  char buf[80];
  struct timeval t;
  struct tm *tmp;
  time_t time_tmp;
  if (gettimeofday (&t, NULL) < 0 || !(time_tmp = t.tv_sec) || 
      !(tmp = localtime (&time_tmp)))
    return "** TIME LOOKUP FAILED **";
  int n = strftime (buf, sizeof (buf), "%a, %d %b %Y %H:%M:%S %z", tmp);
  assert (implicit_cast<size_t> (n) < sizeof (buf));
  return buf;
}

bool
sfspub::get_pubkey_hash (sfs_hash *id, int vers) const
{
  switch (vers) {

    // V1 is for user management, mainly.
  case 1: 
    {
      sfs_pubkey2 p;
      if (!export_pubkey (&p)) return false;
      sha1_hashxdr (id->base (), p);
      return true;
      break;
    }

    // V2 is for turning pubkeys into hostid's, as in the case
    // of SFS paths
  case 2:
    {
      sha1ctx sha;
      sfs_pubkey2_hash k;
      k.type = SFS_PUBKEY2_HASH;
      if (!export_pubkey (&k.pubkey))
	return false;
      str s = xdr2str (k);
      u_int8_t h[sha1::hashsize];
      if (s.len () <= 0) {
	warn << "sfs_mkhostid (" << get_hostname () << "): XDR failed!\n";
	bzero (id->base (), id->size ());
	return false;
      }
      sha.update (s.cstr (), s.len ());
      sha.final (h);
      sha.reset ();
      sha.update (h, sha1::hashsize);
      sha.update (s.cstr (), s.len ());
      sha.final (id->base ());
      return true;
      break;

    }
  default:
    break;
  }
  return false;
}

str
sfspub::get_pubkey_hash (int vers) const
{
  sfs_hash h;
  if (!get_pubkey_hash (&h, vers))
    return NULL;
  return str (h.base (), h.size ());
}

void
sfspriv::init (cbs cb)
{
  sfs_authinfo ainfo;
  ainfo.type = SFS_NULL;
  sfsauth2_sigreq sr (SFS_NULL);
  null_sigreq (&sr);
  sign (sr, ainfo, wrap (this, &sfspriv::initcb, cb));
}

void
sfspriv::initcb (cbs cb, str err, ptr<sfs_sig2> sig)
{
  if (!sig && !err)
    err = "no valid signature returned from server.";
  (*cb) (err);
}
