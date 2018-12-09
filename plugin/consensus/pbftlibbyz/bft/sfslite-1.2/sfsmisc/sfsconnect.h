// -*-c++-*-
/* $Id: sfsconnect.h 3140 2007-11-27 17:05:50Z max $ */

/*
 *
 * Copyright (C) 1999 David Mazieres (dm@uun.org)
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

#ifndef _SFSCONNECT_H_
#define _SFSCONNECT_H_ 1

#include "async.h"
#include "arpc.h"
#include "crypt.h"
#include "sfsmisc.h"
#include "sfsauth_prot.h"
#include "qhash.h"
#include "sfssesscrypt.h"

union sfs_sa {
  sockaddr sa;
  sockaddr_in sa_in;
};

struct sfs_host {
  const str host;
  const sfs_sa sa;
  const socklen_t sa_len;

  ihash_entry<sfs_host> hlink;

  sfs_host (const str &host, const sfs_sa &sa, socklen_t sa_len);
};

class sfs_hosttab_t {
  str filename;

  void errmsg (int line, str);
  void loadfd (int fd);

public:
  bool loaded;
  ihash<const str, sfs_host, &sfs_host::host, &sfs_host::hlink> tab;

  sfs_hosttab_t () : loaded (false) {}
  ~sfs_hosttab_t () { clear (); }
  bool load (const char *path);
  void clear () { tab.deleteall (); loaded = false; }
  const sfs_host *lookup (str name) const;
};

extern sfs_hosttab_t sfs_hosttab;

void sfs_hosttab_init ();

struct sfscon {
  sfs_sa sa;
  socklen_t sa_len;

  str dnsname;
  u_int16_t port;
  ptr<axprt_crypt> x;
  str path;
  sfs_service service;
  sfs_connectinfo ci;
  sfs_connectok cres;
  ptr<const sfs_servinfo_w> servinfo;
  sfs_authinfo authinfo;
  sfs_hash hostid;
  sfs_hash sessid;
  sfs_hash authid;
  bool hostid_valid;
  bool encrypting;
  AUTH *auth;
  str user;
  sfscon () : sa_len (0), port (0), hostid_valid (false),
	      encrypting (false), auth (NULL) {}
  sfscon (const sfscon &c)
    : sa (c.sa), sa_len (c.sa_len), dnsname (c.dnsname), port (c.port),
      x (c.x), path (c.path), service (c.service), ci (c.ci), cres (c.cres),
      servinfo (c.servinfo), authinfo (c.authinfo), hostid (c.hostid),
      sessid (c.sessid), authid (c.authid), hostid_valid (c.hostid_valid),
      encrypting (c.encrypting), auth (NULL)
    { assert (encrypting); }
  ~sfscon () { if (auth) AUTH_DESTROY (auth); }
};

struct sfs_connect_t;
typedef callback<void, ptr<sfscon>, str>::ref sfs_connect_cb;
typedef struct dnsreq dnsreq_t;

class sfs_reconnect_t {
  const sfs_connect_cb cb;
  const ref<sfscon> sc;
  bool force;
  bool encrypt;
  str host;
  u_int16_t port;
  sfs_authorizer *authorizer;
  str user;

  dnsreq_t *areq, *srvreq;
  sfs_connect_t *sfscl;
  ptr<hostent> h;
  ptr<srvlist> srvl;

  sfs_reconnect_t (sfs_connect_cb cb, ref<sfscon> sc, bool force,
		   bool encrypt, sfs_authorizer *a, str user);
  ~sfs_reconnect_t ();
  void fail (str msg) { (*cb) (NULL, msg); delete this; }
  void dnscb_a (ptr<hostent> hh, int dnserr);
  void dnscb_srv (ptr<srvlist> s, int dnserr);
  void dnscb (int dnserr);
  void connectcb (ptr<sfscon> nsc, str err);

public:
  static sfs_reconnect_t *alloc (ref<sfscon> sc, sfs_connect_cb cb,
				 bool force = true, bool encrypt = true,
				 sfs_authorizer *a = NULL, str user = NULL)
    { return New sfs_reconnect_t (cb, sc, force, encrypt, a, user); }
  void cancel () { delete this; }
};

inline sfs_reconnect_t *
sfs_reconnect (ref<sfscon> sc, sfs_connect_cb cb, bool force = true)
{
  return sfs_reconnect_t::alloc (sc, cb, force);
}

class sfs_connect_t {
  friend void sfs_connect_local_cb (sfs_connect_t *cs,
				    ptr<sfscon> sc, str err);
  friend sfs_connect_t *sfs_connect_host (str host, sfs_service service,
					  sfs_connect_cb cb, bool encrypt);

  static ptr<sfspriv> ckey;

  const sfs_connect_cb cb;
  tcpconnect_t *tcpc;
  callbase *cbase;

  sfs_connectres cres;
  ptr<srvlist> srvl;
  str dnsname;

  ref<sfscon> sc;
  ptr<aclnt> c;
  bool local_authd;
  int last_srv_err;

  str location;
  sfs_hash hostid;
  u_int16_t port;
  bhash<str> location_cache;

  ref<bool> destroyed;
  rpc_ptr<sfsagent_authmore_arg> marg;
  rpc_ptr<sfsagent_auth_res> ares;
  rpc_ptr<sfs_loginres> lres;
  rpc_ptr<sfs_loginres_old> olres;
  
  static void ckey_clear () { ckey = NULL; }

  ~sfs_connect_t ();
  void init ();
  void fail (int e, str msg) { errno = e; (*cb) (NULL, msg); delete this; }
  void srvfail (int e, str msg);
  void succeed ();
  void carg_reset () { sc->ci.set_civers (5); *sc->ci.ci5 = ci5; }
  bool setsock (int fd);

  bool start_common ();
  void getfd (ref<bool> d, int fd);
  void sendconnect ();
  void getconres (ref<bool> d, enum clnt_stat err);
  bool dogetconres ();
  void docrypt ();
  void cryptcb (ref<bool> d, const sfs_hash *sessidp);
  void doauth ();
  void dologin (ref<bool> destroyed);
  void donelogin (ref<bool> dest, clnt_stat);
  void checkedserver (ref<bool> d);

public:
  sfsagent_authinit_arg aarg;
  sfs_connectinfo_5 ci5;
  bool encrypt;
  bool check_hostid;
  sfs_authorizer *authorizer;
  
  sfs_connect_t (const sfs_connect_cb &c)
    : cb (c), tcpc (NULL), cbase (NULL),
      sc (New refcounted<sfscon>), local_authd (false), last_srv_err (0),
      port (0), destroyed (New refcounted<bool> (false)), encrypt (true),
      check_hostid (true), authorizer (NULL) { init (); }
  sfs_connect_t (ref<sfscon> s, const sfs_connect_cb &c)
    : cb (c), tcpc (NULL), cbase (NULL), sc (s), local_authd (false),
    last_srv_err (0), port (0), destroyed (New refcounted<bool> (false)),
      encrypt (true), check_hostid (true), authorizer (NULL) { init (); }

  str &sname () { return ci5.sname; }
  sfs_service &service () { return ci5.service; }
  rpc_vec<sfs_extension, RPC_INFINITY> &exts () { return ci5.extensions; }

  bool start ();
  bool start (ptr<aclnt> c);
  bool start (ptr<srvlist> sl);
  bool start (in_addr a, u_int16_t port);
  bool start_scon ();
  void cancel ();
};

class sfs_pathcert {
  sfs_pathcert (const sfs_pathrevoke &cert, const sfs_hash &hostid);

public:
  const sfs_pathrevoke cert;
  const sfs_hash hostid;
  ihash_entry<sfs_pathcert> hlink;

  ~sfs_pathcert ();
  bool valid () const {
    return !cert.msg.redirect
      || cert.msg.redirect->expire < 
      implicit_cast<sfs_time> (sfs_get_timenow());
  }
  str dest () const {
    if (cert.msg.redirect && valid ())
      return sfs_servinfo_w::alloc (cert.msg.redirect->hostinfo)->mkpath ();
    return NULL;
  }
  bool isbetter (const sfs_pathrevoke &c2) {
    if (!valid ())
      return true;
    if (!cert.msg.redirect)
      return false;
    if (!c2.msg.redirect)
      return true;
    return c2.msg.redirect->serial > cert.msg.redirect->serial;
  }

  static sfs_pathcert *store (const sfs_pathrevoke &r);
  static sfs_pathcert *lookup (const sfs_hash &hostid);
  static sfs_pathcert *lookup (const sfs_connectok &cr);
};
extern ihash<const sfs_hash, sfs_pathcert,
	     &sfs_pathcert::hostid, &sfs_pathcert::hlink> pathcert_tab;

void sfs_initci (sfs_connectinfo *ci, str path, sfs_service service,
		 sfs_extension *ext_base = NULL, size_t ext_size = 0);
inline void
sfs_initci (sfs_connectinfo *ci, str path, sfs_service service,
	    vec<sfs_extension> *exts)
{
  sfs_initci (ci, path, service, exts->base(), exts->size ());
}
bool sfs_nextci (sfs_connectinfo *ci);

sfs_connect_t *sfs_connect (const sfs_connectarg &carg, sfs_connect_cb cb,
                            bool encrypt = true, bool check_hostid = true);
sfs_connect_t *sfs_connect_path (str path, sfs_service service,
                                 sfs_connect_cb cb, bool encrypt = true,
                                 bool check_hostid = true,
				 sfs_authorizer *a = NULL, str user = NULL);
sfs_connect_t *sfs_connect_host (str host, sfs_service service,
                                 sfs_connect_cb cb, bool encrypt = true);
sfs_connect_t *sfs_connect_crypt (ref<sfscon> sc, sfs_connect_cb cb,
				  sfs_authorizer *a = NULL, str user = NULL,
				  sfs_seqno seq = 0);
void sfs_connect_cancel (sfs_connect_t *);

typedef callback<void, str>::ref sfs_dologin_cb;
void sfs_dologin (ref<sfscon> scon, sfspriv *key, int seqno, sfs_dologin_cb cb,
		  bool cred = true);

#endif /* _SFSCONNECT_H_ */
