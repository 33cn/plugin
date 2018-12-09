/* $Id: sfskeymgr.h 3758 2008-11-13 00:36:00Z max $ */

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

#ifndef _SFSMISC_SFSKEYMGR_H
#define _SFSMISC_SFSKEYMGR_H

#include <string.h>
#include "srp.h"
#include "qhash.h"
#include "sfscrypt.h"
#include "sfskeymisc.h"
#include "agentmisc.h"

typedef callback<void, str, bool>::ref km_update_cb;
typedef callback<void, str, ptr<sfscon>, ptr<sfspriv> >::ref km_login_cb;
typedef callback<void, str, ptr<sfspub> >::ref fpkcb;
typedef callback<void, str, ptr<sfscon> >::ref concb;

/* options per function call */
#define KM_NOSRP   (1 << 0)    /* Don't use SRP */
#define KM_NOESK   (1 << 1)    /* Don't Store Encrypted Secret Key */
#define KM_SKH     (1 << 2)    /* Store Server Key Half */
#define KM_DLT     (1 << 3)    /* Update Server Key Half with Delta-x */
#define KM_REREG   (1 << 4)    /* Allow reregistration */
#define KM_NOKBD   (1 << 5)    /* No keyboard noise in registration */
#define KM_PROAC   (1 << 6)    /* Generate Proactive / 2Schnorr Key */
#define KM_GEN     (1 << 7)    /* If ambiguous between fetch or gen, use gen */
#define KM_FGEN    (1 << 8)    /* Always generate */
#define KM_NOPWD   (1 << 9)    /* Do not encrypt secret key */
#define KM_FRC     (1 << 11)   /* Force key overwrites */
#define KM_NOPK    (1 << 12)   /* Don't set public key */
#define KM_NOLNK   (1 << 13)   /* Don't make a symlink when writing keys */
#define KM_UNX     (1 << 14)   /* Attempt Unix Login */
#define KM_KPSRP   (1 << 15)   /* Keep SRP values on server */
#define KM_KPESK   (1 << 16)   /* Keep Encrypted Secret Key on server */
#define KM_KPPK    (1 << 17)   /* Keep public key value on server */
#define KM_NOSRC   (1 << 18)   /* Do not search through key repository */
#define KM_NWPWD   (1 << 21)   /* Make a new password. */
#define KM_ALL     (1 << 22)   /* Apply to all backup keys */
#define KM_CHNGK   (1 << 23)   /* Change Proac key hostname/user b4 update */
#define KM_PKONLY  (1 << 24)   /* Fetch public key from server only */
#define KM_FRCLNK  (1 << 25)   /* Force link overwrite of regular file */
#define KM_REALM   (1 << 26)   /* Allow changes to SRP realm */
#define KM_ESIGN   (1 << 27)   /* Generate Esign Keys */

/* options per object */
#define KM_NOHM    (1 << 10)   /* No access to home directory */
#define KM_NOCRT   (1 << 19)   /* Do not create new .sfs directory, etc */
#define KM_NODCHK  (1 << 20)   /* Do not check default key symlink */

typedef enum { SFSKI_NONE = 0, SFSKI_STD = 1, SFSKI_PROAC = 2 } sfski_type;

class sfskeystore;
struct sfskeyinfo {
  sfskeyinfo (ptr<sfspriv> p) 
    : kn (NULL), dir (NULL), kt (p->is_proac () ? SFSKI_PROAC : SFSKI_STD), 
      rkt (SFSKI_NONE), version (1), exists (true), remote (false), 
      gen (false), bad (false), defkey (false), flagged (false), next (NULL), 
      privk (p) {}
  sfskeyinfo (const str &f, const str &d = NULL, sfski_type k = SFSKI_NONE,
	      int v = 1) 
    : kn (f), dir (d), kt (k), rkt (SFSKI_NONE), version (v), exists (true), 
      remote (false), gen (false), bad (false), defkey (false), 
      flagged (false), next (NULL), privk (NULL) {}
  virtual ~sfskeyinfo () {}

  static sfskeyinfo *alloc (str r, sfskeystore *ks, u_int32_t opts = 0);
  static sfskeyinfo *alloc (const str &raw, str d, sfski_type ki,
			    sfskeyinfo *m);
  static sfskeyinfo *alloc (const str &r, str d = NULL);

  virtual int cmp (sfskeyinfo *k) const { return -1; }
  int cmp (const str &r, bool kcomplete) const 
    { return (strcmp (r, kcomplete ? fn () : kn)); }
  int gcmp (const sfskeyinfo &k) const { return (version - k.version); }
  virtual void setpriority (int i) {}
  virtual bool set_hostname (const str &s = NULL) { return true; }
  virtual str fn () const { return kn; }
  virtual str prefix () const { return kn; }
  str fn_std () const { return strbuf (kn << "#" << version); }
  str afn () const;
  virtual bool setlink () const { return false; }
  virtual bool has_backup_keys () const { return false; }
  virtual bool bump_privk_version () { return false; };
  bool is_proactive () const 
    { return (kt == SFSKI_PROAC || rkt == SFSKI_PROAC); }

  str kn;
  str dir;
  str keylabel;
  sfski_type kt;
  sfski_type rkt;
  int version;
  bool exists;  
  bool remote;  
  bool gen;
  bool bad;
  bool defkey;
  bool flagged;
  
  sfskeyinfo *next;
  ptr<sfspriv> privk;
};

typedef callback<void, sfskey *>::ref cbk;
typedef callback<void, sfskeyinfo *, sfskey *>::ref cbkik;

struct sfskeyinfo_std : public sfskeyinfo {
  sfskeyinfo_std (const str &r, const str &d = NULL, int v = 1)
    : sfskeyinfo (r, d, SFSKI_STD, v) {}
  int cmp (sfskeyinfo *k) const { return gcmp (*k); }
  str fn () const { return fn_std (); }
  bool setlink () const { return true; }

};

struct sfskeyinfo_proac : public sfskeyinfo {
  sfskeyinfo_proac (ptr<sfspriv> p) : sfskeyinfo (p) {}
  sfskeyinfo_proac (const str &f, const str &d = NULL, int pv = 1, 
		      const str &h = NULL, int hp = 1, int sv = 1)
    : sfskeyinfo (f, d, SFSKI_PROAC, pv), host (h), host_priority (hp), 
      privk_version (sv) {}
  int cmp (sfskeyinfo *i) const;
  void setpriority (int i) { host_priority = i; }
  bool set_hostname (const str &s = NULL) ;
  bool setlink () const { return true; }
  bool has_backup_keys () const { return true; }
  bool bump_privk_version () { privk_version++; return true; };
  str fn () const { return strbuf (kn << "#" << version << "," << 
				   host_priority << "." << host << "," << 
				   privk_version ); }
  str host;
  int host_priority;
  int privk_version;
};

class sfskeymgr;
class sfskeystore {
public:
  sfskeystore (const str &u, u_int32_t o) : 
    g_opts (o), read (false), user (u), init_ck (false) {}
  
  sfskeyinfo *search (const str &raw, bool kcomplete = false);
  sfskeyinfo *search (const str &raw, sfski_type kt, bool kcomplete = false);
  sfskeyinfo *generate (str raw, sfski_type kt, bool kcomplete = false);
  sfskeyinfo *generate (sfskey *k, u_int32_t l_opts);
  bool save (sfskey *k, sfskeyinfo *ki, u_int32_t l_opts);
  bool setlink (const str &target, u_int32_t l_opts = 0);
  sfskeyinfo *fill_list (sfskeyinfo *ki);

  sfskeyinfo *getkeys (ptr<sfspub> pub);
  void prepend (ptr<sfspriv> priv);
  sfskey *fetch (sfskeyinfo *, sfskeymgr *km, str *err, str *pwd, 
		 u_int32_t l_opts);
  bool check_local (sfskeyinfo *ki, u_int32_t l_opts);

  const str &getuser () const { return user; }

  str defkey2 ();
  str defkey2_readlink ();
  
  u_int32_t g_opts;
private:
  void hashit (ptr<sfspub> pub, sfskeyinfo *k_new);
  void init (bool needkeysdir = true);
  bool lsdir ();
  bool read;
  vec<sfskeyinfo *> ls[3];
  qhash<str, sfskeyinfo *> pktab_std;
  qhash<str, sfskeyinfo_proac *> pktab_proac;
  const str user;
  str dir;
  bool init_ck;
};

class sfskeymgr {
public:

  struct host_t {
    str hostname; // ludlow.scs.cs.nyu.edu
    str sfspath;  // @ludlow.scs.cs.nyu.edu%200,ffffaaaafff00000ff
    str ahash;    // ludlow.scs.cs.nyu.edu%200
  };

  struct user_host_t : public sfskeymgr::host_t {
    str user;     // max
    str hash;     // max@ludlow.scs.cs.nyu.edu%200
  };

private:
  struct ustate {

    ustate (sfskey *n, ptr<sfspriv> o, ptr<sfscon> sc,
	    const str &h, const str &u, const km_update_cb cbu, 
	    u_int32_t op) : 
      nk (n), ok (o), hostname (h), user (u), cb (cbu), u_opts (op),
      sigerr (NULL), sigs (0), delta (false) { setcon (sc); }

    void setcon (ptr<sfscon> sc) 
    { if (sc) { scon = sc; c = aclnt::alloc (sc->x, sfsauth_prog_2); } }

    sfskey *nk;
    ptr<sfspriv> ok;
    const str hostname;
    const str user;
    const km_update_cb cb;
    const u_int32_t u_opts;
    ptr<sfscon> scon;
    ptr<aclnt> c;
    str realm;
    sfsauth2_query_res aqr;
    sfsauth2_update_res aur;
    sfsauth2_update_arg aua;
    str sigerr;
    int sigs;
    sfsauth_userinfo uinfo;
    bool delta;
  };

  struct key_con_t {
    key_con_t (ptr<sfscon> c, ptr<sfspriv> k, sfskey *w) 
      : con (c), key (k), pubkey (key), kw (w) {}

    key_con_t (ptr<sfscon> c, ptr<sfspub> p)
      : con (c), pubkey (p), kw (NULL) {}

    void upgrade (ptr<sfspriv> k, sfskey *w)
    { 
      key = k;
      pubkey = k;
      kw = w;
    }
    const ptr<sfscon> con;
    ptr<sfspriv> key;
    ptr<sfspub> pubkey;
    sfskey *kw;
  };

  struct lstate {
    const user_host_t uh;
    ptr<sfspriv> key;
    ptr<sfspub> pkey;
    const km_login_cb cb;
    ptr<sfscon> scon;
    ptr<aclnt> c;
#if 0
    sfs_loginres res;
    sfs_autharg2 arg;
#endif
    sfsauth2_query_res aqr;
    u_int32_t opts;
    bool fetching;
    sfs_authorizer *a;
    sfs_seqno seqno;
    lstate (const user_host_t &u, ptr<sfspriv> k, km_login_cb c, 
	    u_int32_t op)
      : uh (u), key (k), cb (c), opts (op), fetching (false),
	 a (NULL), seqno (0) {}
    ~lstate () { delete a; }
  };

  struct fpkstate_t { // fetch pub key state
    fpkstate_t (const user_host_t &u, fpkcb c)
      : uh (u), cb (c) {}
    const user_host_t uh;
    const fpkcb cb;
    ptr<sfscon> scon;
    ptr<sfspub> pub;
  };

  struct blocked_fetch_t {
    blocked_fetch_t (lstate *s, sfskeyinfo *k) : st (s), ki (k) {}
    lstate *st;
    sfskeyinfo *ki;
  };


public:
  sfskeymgr (const str &user = NULL, u_int32_t g_opts = 0);
  bool getsrp (const str &filename = NULL, ptr<sfscon> sc = NULL);
  sfskeyinfo *getkeyinfo (const str &raw, u_int32_t opts = 0);
  sfskeyinfo *getkeyinfo (sfskey *k, u_int32_t opts = 0);
  sfskeyinfo *getkeyinfo_list (const str &kn, u_int32_t opts = 0);

  void update (sfskey *nk, ptr<sfspriv> ok, const str &path, 
	       u_int32_t opts, km_update_cb cb);
  void update (sfskey *nk, ptr<sfspriv> ok, ptr<sfscon> scon, 
	       const str &path, const str &user, u_int32_t opts,
	       km_update_cb cb);
  void fetchpub (sfskeyinfo *ki, fpkcb c);
  bool fetchpub (const user_host_t &uh, fpkcb c);
  sfskey *fetch (sfskeyinfo *ki, str *errp, u_int32_t opts = 0);
  sfskey *fetch_or_gen (sfskeyinfo *ki, str *errp, u_int nbits = 0, 
			u_int cost = 0, u_int32_t opts = 0);
  void fetch_from_list (sfskeyinfo *ki, cbk cb);
  void fetch_all (sfskeyinfo *ki, cbkik cb);
  bool save (sfskey *k, sfskeyinfo *ki, u_int32_t opts = 0);
  bool select (sfskeyinfo *ki, u_int32_t opts = 0);
  void login (const str &hostname, km_login_cb cb, ptr<sfspriv> key = NULL, 
	      u_int32_t opts = 0);
  void connect (const str &h, concb c);
  void add (ptr<sfspriv> priv) { ks->prepend (priv); }
  void add_keys (const vec<str> &keys);
  bool add_con (ptr<sfspriv> key, ptr<sfscon> *, str *);
  void check_connect (const vec<str> &servers, cbv cb);
  bool bump_privk_version (sfskeyinfo *ki, u_int32_t opts);
  ptr<sfscon> getsrpcon (sfskeyinfo *ki);

  // helpers for update function (have to be public, but shouldn't be...)
  void gotres (ustate *s, clnt_stat err);
  void gotsig (ustate *s, sfs_sig2 *target, str err, ptr<sfs_sig2> sig);
  void gotcertinfo (ustate *s, clnt_stat err);
  void gotuinfo (ustate *s, clnt_stat err);
  
  // helpers for login function
  void gotcon (lstate *s, str err, ptr<sfscon> sc);
  //void gotunixlogin (lstate *s, int ntries, clnt_stat err);
  void gotunixlogin (lstate *s, ptr<sfscon> sc, str err);
  void gotlogin (lstate *s, str err);
  void gotlogin_r (lstate *ls, sfskeyinfo *ki, str err);

  // helper for getservinfo
  void si_gotcon (ptr<sfscon> s, str err);

  // helper for check_connect
  void cc_cb (str srv, cbv cb, str err, ptr<sfscon> con, ptr<sfspriv> k);
  
  // helper for fetch_from_list
  void initcb (sfskeyinfo *ki, sfskey *k, cbk cb, str err);

  bool insert (const str &kn, ptr<sfscon> con, ptr<sfspriv> key, 
	       sfskey *kw, str *err);
  bool get_userhost (const str &h, user_host_t *uh);
  bool get_host (const str &s, host_t *h);

private: 

  // helpers for update function
  void u_gotlogin (ustate *s, str err, ptr<sfscon> sc, ptr<sfspriv> k);
		   
  void update (ustate *s);
  void doupdate (ustate *s);
  void getuinfo (ustate *s);
  void getcertinfo (ustate *s);
  bool setup_uinfo (ustate *s);
  void sign_updatereq (ustate *s);
  void done (ustate *, str, bool gotconf = true);
  bool connect (const host_t &uh, concb c);

  // helpers for login function
  void done (lstate *, str );
  void unixlogin (lstate *s, int ntries = 0);
  void getpubkey (ptr<axprt> x, const str &un, fpkcb c);
  void dologin (lstate *ls, sfskeyinfo *ki);
  void gotpubkey (ptr<sfsauth2_query_res> r, fpkcb c, clnt_stat err);
  void login_gotpubkey (lstate *l, str err, ptr<sfspub> p);
  void connected (concb c, host_t h, ptr<sfscon> con, str err);

  // helper for fetchpub function
  void done (fpkstate_t *fps, str err);
  void fetchpub_gotcon (fpkstate_t *fps, str err, ptr<sfscon> con);
  void fetchpub_gotpubkey (fpkstate_t *fps, str err, ptr<sfspub> p);
  
  void setuser (str user);
  ptr<const sfs_servinfo_w> getservinfo ();
  str cse2str (clnt_stat e) { strbuf b; b << e; return b; }
  void insertcon (const user_host_t &uh, ptr<sfscon> c, ptr<sfspriv> k,
		  sfskey *kw = NULL);
  void insertcon (const user_host_t &uh, ptr<sfscon> c, key_con_t *k);
  void insertcon (const user_host_t &uh, ptr<sfscon> c, ptr<sfspub> p);

  str check_uinfo (const sfsauth2_query_res &aqr);

  bool checkflag;
  qhash<str, key_con_t *> keycontab;
  qhash<str, ptr<sfscon> > anoncontab;
  u_int32_t uid;
  str user;

  bigint g,N; // SRP params
  sfskeystore *ks;
  str pwd;

  ptr<const sfs_servinfo_w> servinfo;
  bool fetching;
  bool fflerr;
  vec< blocked_fetch_t *> fqueue;
  qhash<str, vec<concb> *> cqueue;
  u_int ncb;
  u_int32_t g_opts;

};

str seconds2str (sfs_time t);

#endif
