// -*-c++-*-
/* $Id: sfssesscrypt.h 1754 2006-05-19 20:59:19Z max $ */

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

#ifndef _SFSMISC_SFSSESSCRYPT_H_
#define _SFSMISC_SFSSESSCRYPT_H_ 1

#include "sfsagent.h"
#include "srp.h"

struct sfs_autharg2;
struct sfs_authreq2;

/* sfsauthorizer.C */
struct sfs_authorizer {
  bool mutual;
  sfs_authorizer () : mutual (true) {}
  static bool setres (sfsagent_auth_res *resp, const sfs_autharg2 &aa);
  static bool reqinit (sfs_authreq2 *reqp, const sfsagent_authinit_arg *ap);
  static bool reqinit (sfs_authreq2 *reqp, const sfsagent_authmore_arg *ap);
  // Warning: in next 2 funcs can't delete argp's before cb is called
  virtual void authinit (const sfsagent_authinit_arg *argp,
			 sfsagent_auth_res *resp, cbv cb) = 0;
  virtual void authmore (const sfsagent_authmore_arg *argp,
			 sfsagent_auth_res *resp, cbv cb);
  virtual ~sfs_authorizer () {}
};

class sfskey_authorizer : virtual public sfs_authorizer {
  virtual void sigdone (ref<sfs_autharg2> aarg, sfsagent_auth_res *resp,
			cbv cb, str err, ptr<sfs_sig2> sig);
protected:
  ptr<sfspriv> khold;

  virtual bool ntries_ok (int ntries) { return !ntries; }
  void authinit_v1 (const sfsagent_authinit_arg *argp,
		    sfsagent_auth_res *resp, cbv cb);
  void authinit_v2 (const sfsagent_authinit_arg *argp,
		    sfsagent_auth_res *resp, cbv cb);
public:
  sfspriv *k;
  bool cred;
  sfskey_authorizer () : cred (true) { mutual = false; }
  void setkey (ptr<sfspriv> kk) { k = khold = kk; }
  void setkey (ref<sfspriv> kk) { k = khold = kk; }
  void setkey (sfspriv *kk) { k = kk; khold = NULL; }
  void authinit (const sfsagent_authinit_arg *argp,
		 sfsagent_auth_res *resp, cbv cb);
};

class sfspw_authorizer : virtual public sfs_authorizer {
  static void getpwd_2 (ref<bool> d, cbs cb, str pwd) { if (!*d) (*cb) (pwd); }
protected:
  const ref<bool> destroyed;
  bool ntries_ok (int ntries) { return ntries < 5; }
public:
  bool dont_get_pwd;

  static str printable (str msg);

  sfspw_authorizer ()
    : destroyed (New refcounted<bool> (false)), dont_get_pwd (false) {}
  ~sfspw_authorizer () { *destroyed = true; }
  virtual void getpwd (str prompt, bool echo, cbs cb);
};

class sfsunixpw_authorizer : public sfspw_authorizer {
  void authmore_2 (sfsagent_auth_res *resp,
		   ref<sfs_autharg2> aargp, cbv cb, str pwd);
public:
  str unixuser;
  void authinit (const sfsagent_authinit_arg *argp,
		 sfsagent_auth_res *resp, cbv cb);
  void authmore (const sfsagent_authmore_arg *argp,
		 sfsagent_auth_res *resp, cbv cb);
};

class sfssrp_authorizer : public sfspw_authorizer {
  void authmore_2 (const sfsagent_authmore_arg *argp,
		   sfsagent_auth_res *resp, cbv cb, str pwd);
public:
  str pwd;
  srp_client *srpc;
  bool delete_srpc;
  sfssrp_authorizer () : srpc (NULL), delete_srpc (false) {}
  ~sfssrp_authorizer () { if (delete_srpc) delete srpc; }
  void authinit (const sfsagent_authinit_arg *argp,
		 sfsagent_auth_res *resp, cbv cb);
  void authmore (const sfsagent_authmore_arg *argp,
		 sfsagent_auth_res *resp, cbv cb);
};


/* sfssesskey.C */
inline const sfs_kmsg *
sfs_get_kmsg (const str &s)
{
  return reinterpret_cast<const sfs_kmsg *> (s.cstr ());
}
void sfs_get_sesskey (sfs_hash *ksc, sfs_hash *kcs,
		      const sfs_servinfo &si, const sfs_kmsg *smsg, 
		      const sfs_connectinfo &ci, const bigint &kc,
		      const sfs_kmsg *cmsg);
void sfs_get_sessid (sfs_hash *sessid, const sfs_hash *ksc,
		     const sfs_hash *kcs);
void sfs_get_authid (sfs_hash *authid, sfs_service service, sfs_hostname name,
		     const sfs_hash *hostid, const sfs_hash *sessid,
		     sfs_authinfo *authinfo = NULL);
void sfs_server_crypt (svccb *sbp, sfspriv *sk,
		       const sfs_connectinfo &ci, ref<const sfs_servinfo_w> s,
		       sfs_hash *sessid, const sfs_hashcharge &charge,
		       axprt_crypt *cx = NULL, int PVERS = 2);
/* N.B., sfs_client_crypt might make callback immediately (on failure)
 * and return NULL.  The callback argument is sessid, for use with
 * sfs_get_authid. */
struct callbase;
callbase *sfs_client_crypt (ptr<aclnt> c, ptr<sfspriv> clntkey,
			    const sfs_connectinfo &ci,
			    const sfs_connectok &cres,
			    ref<const sfs_servinfo_w> si,
			    callback<void, const sfs_hash *>::ref cb,
			    ptr<axprt_crypt> cx = NULL);

#endif /* _SFSMISC_SFSSESSCRYPT_H_ */
