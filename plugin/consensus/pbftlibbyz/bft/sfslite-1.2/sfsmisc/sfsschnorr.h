// -*-c++-*-
// The above line tells emacs that this header likes C++

#ifndef _SFSMISC_SCHNORR_H_
#define _SFSMISC_SCHNORR_H_

#include <assert.h>
#include "schnorr.h"

#include "sfs_prot.h"
#include "sfsauth_prot.h"
#include "sfsagent.h"
#include "sfsmisc.h"
#include "arpc.h"
#include "schnorr.h"
#include "sfscrypt.h"

#define MAX_SCHNORR_SESSIONS 20

struct sign_block_t {
  sign_block_t (const sfsauth2_sigreq &s, const sfs_authinfo &a,
		const cbsign c) : sr (s), ainfo (a), cb (c) {}
  sfsauth2_sigreq sr;
  sfs_authinfo ainfo;
  cbsign cb;
};


class schnorr_client {
protected:
  ref<schnorr_clnt_priv> sign_share;
  ref<aclnt> sign_client;
  sfs_hash pkh;
  ref<ephem_key_pair> ekp;
  int nsessions;
  vec<sign_block_t *> squeue;      /* sign queue */

  void complete_sign (ptr<ephem_key_pair> clnt_rnd, 
		      ptr<sfsauth2_sign_arg> clnt_arg, 
		      ptr<sfsauth2_sign_res> srv_res,
		      cbsign use_it, clnt_stat stat);

  void scb (int *rc, clnt_stat stat);

public:
  AUTH *auth;
  str uname;

  schnorr_client (ref<schnorr_clnt_priv> share, ref<aclnt> client, sfspub *p,
		  AUTH *a = NULL, str un = NULL)
    : sign_share (share), sign_client (client),
      ekp (sign_share->make_ephem_key_pair ()), nsessions (0),
      auth (a), uname (un) 
  { assert (p); p->get_pubkey_hash (&pkh); }

  void sign (const sfsauth2_sigreq &req, const sfs_authinfo &authinfo,
	     cbsign cb);
  
  static bool ni_sign (ptr<schnorr_clnt_priv> cpriv, 
		       ptr<schnorr_srv_priv> spriv,
		       sfs_sig2 *sig, const str &msg);
};

class sfs_schnorr_pub : virtual public sfspub {
public:
  sfs_schnorr_pub (ref<schnorr_pub> k, u_char o = 0) 
    : sfspub (SFS_SCHNORR, o, "schnorr"),  pubk (k)
    { assert (!get_opt (SFS_ENCRYPT | SFS_DECRYPT)); }
  bool verify (const sfs_sig2 &sig, const str &msg, str *e) const;
  bool export_pubkey (sfs_pubkey2 *k) const;
  bool export_pubkey (strbuf &b, bool prefix = true) const;
  bool check_keysize (str *s = NULL) const;
  bool encrypt (sfs_ctext2 *ct, const str &msg) const { return false; }
  bool operator== (const sfsauth_keyhalf &kh) const;
  
  static int find (const sfsauth_keyhalf &kh, const sfs_hash &h);
  static bool check_keysize (size_t pbits, size_t qbits, str *s = NULL);

 private:
  ref<schnorr_pub> pubk;
};

class sfs_1schnorr_priv : public sfs_schnorr_pub, public sfspriv {
public:
  sfs_1schnorr_priv (ref<schnorr_priv> sp, u_char o = 0)
    : sfspub (SFS_1SCHNORR, o, "oneschnorr"), sfs_schnorr_pub (sp, o),
      sfspriv (SFS_1SCHNORR, o, "oneschnorr"), privk (sp) {}
  bool export_privkey (sfs_privkey2_clear *k) const;
  bool decrypt (const sfs_ctext2 &ct, str *msg, u_int sz) 
    const { return false; }
  bool sign (sfs_sig2 *sig, const str &msg);
  str get_desc () const;

protected:
  bool export_privkey (str *s) const;

private:
  bool export_privkey (sfs_1schnorr_priv_xdr *sxdr) const;
  ref<schnorr_priv> privk;
};

class sfs_2schnorr_priv : public sfs_schnorr_pub, public sfspriv {
public:

  sfs_2schnorr_priv (ptr<sfscon> c, ref<schnorr_clnt_priv> cp, 
		     ptr<schnorr_srv_priv> sp = NULL,
		     ptr<schnorr_priv> w = NULL,
		     str hostname = NULL, str username = NULL, 
		     u_char o = 0) ;
  ~sfs_2schnorr_priv () { if (sclnt) delete sclnt; }

  bool get_privkey_hash (u_int8_t *buf, const sfs_hash &hostid) const 
    { return false; }
  bool decrypt (const sfs_ctext2 &ct, str *msg, u_int sz) const
    { return false; }
  bool sign (sfs_sig2 *sig, const str &msg);
  void sign (const sfsauth2_sigreq &sr, sfs_authinfo ainfo, cbsign cb);
  bool export_privkey (sfs_privkey2_clear *k) const;

  void set_username (const str &s) { uname = s; }
  void set_hostname (const str &s) { hostname = s; }
  str  get_hostname () const { return hostname; }
  bool is_proac () const { return true; } 

  // should be private -- called as callbacks
  void gotcon (sfsauth2_sigreq sr, sfs_authinfo ainfo, 
	       cbsign cb, ptr<sfscon> sc, str err);
  void gotlogin (sfsauth2_sigreq sr, sfs_authinfo ainfo,
		 cbsign cb, str err);

  ptr<sfspriv> update () const; 
  ptr<sfspriv> regen () const;
  ptr<sfspriv> wholekey () const;
  bool get_coninfo (ptr<sfscon> *c, str *h) const
  { *c = scon; *h = hostname; return true; }
  bool export_keyhalf (sfsauth_keyhalf *kh, bool *dlt) const;
  
  // yuck -- this is a mess, but we need these to deal with server keys, too
  static bool export_privkey (str *raw, sfs_2schnorr_priv_xdr priv[], int num);
  static bool export_privkey (sfsauth_keyhalf *k, 
			      ptr<const schnorr_srv_priv> s);
  static bool parse_keyhalf (sfsauth_keyhalf *k, const str &s);
  static bool export_keyhalf (const sfsauth_keyhalf &k, strbuf &b);

  str get_desc () const;

protected:
  bool export_privkey (str *s) const;

public:
  ptr<sfscon> scon;
  ptr<schnorr_srv_priv> sprivk;
  ptr<schnorr_priv> wsk;
  str hostname;
  str uname;
  str audit;
  bigint delta;

private:
  ptr<aclnt> get_aclnt (ptr<axprt> x);
  void eofcb ();
  static bool export_privkey (sfs_2schnorr_priv_xdr *p, 
			      ptr<const schnorr_pub> k, str hostname = NULL,
			      str username = NULL, str audit = NULL);
  bool export_privkey (sfs_2schnorr_priv_xdr *x) const
  { return export_privkey (x, privk, hostname, uname, audit); }
  void fixup (sfs_authinfo *a);

  schnorr_client *sclnt;
  ref<schnorr_clnt_priv> privk;
  bool connecting;
  void con_finish (cbsign cb, const str &err);
  void con_finish (const str &err = NULL);
  vec<sign_block_t *> cqueue;      /* connect queue */
};

class sfs_schnorr_alloc : public sfsca {
 public:
  sfs_schnorr_alloc (sfs_keytype kt = SFS_SCHNORR, const str &s = "schnorr") 
    : sfsca (kt, s) { skt[0] = SFS_1SCHNORR; skt[1] = SFS_2SCHNORR; }

  virtual ptr<sfspriv> alloc (const sfs_privkey2_clear &pk, u_char o = 0) const
  { return NULL; }
  ptr<sfspub>  alloc (const sfs_pubkey2 &pk, u_char o = 0) const;
  ptr<sfspub>  alloc (const str &s, u_char o = 0) const;
  virtual ptr<sfspriv> alloc (const str &raw, ptr<sfscon> c, u_char o = 0) 
    const { return NULL; }
  virtual ptr<sfspriv> gen (u_int nbits, u_char o = 0) const { return NULL; }
  const sfs_keytype *get_private_keytypes (u_int *n) const 
  { *n = 2; return skt; }

private:
  sfs_keytype skt[2];
};

class sfs_2schnorr_alloc : public sfs_schnorr_alloc {
public:
  sfs_2schnorr_alloc () : sfs_schnorr_alloc (SFS_2SCHNORR, "twoschnorr") {}
  ptr<sfspriv> alloc (const sfs_privkey2_clear &pk, u_char o = 0) const;
  ptr<sfspriv> alloc (const str &raw, ptr<sfscon> c, u_char o = 0) const;
  ptr<sfspriv> gen (u_int nbits, u_char o = 0) const;
  static bool import_priv (const str &raw, sfs_2schnorr_priv_xdr priv[],
			   int *num);
 private:
  ptr<sfspriv> alloc (const sfs_2schnorr_priv_xdr &k, ptr<sfscon> c, 
		      u_char o = 0) const;
};

class sfs_1schnorr_alloc : public sfs_schnorr_alloc {
public:
  sfs_1schnorr_alloc () : sfs_schnorr_alloc (SFS_1SCHNORR, "oneschnorr") {}
  ptr<sfspriv> alloc (const sfs_privkey2_clear &pk, u_char o = 0) const;
  ptr<sfspriv> alloc (const str &raw, ptr<sfscon> c, u_char o = 0) const;
private:
  ptr<sfspriv> alloc (const sfs_1schnorr_priv_xdr &k, u_char o = 0) const;
};

str sigreq2str (const sfsauth2_sigreq &sr);
bool sigreq_authid_cmp (const sfsauth2_sigreq &sr, const sfs_hash &authid);
void null_sigreq (sfsauth2_sigreq *sr);


#endif /* _SFSMISC_SCHNORR_H_ */
