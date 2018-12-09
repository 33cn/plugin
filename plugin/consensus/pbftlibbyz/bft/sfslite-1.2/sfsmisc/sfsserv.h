// -*-c++-*-
/* $Id: sfsserv.h 1754 2006-05-19 20:59:19Z max $ */

/*
 *
 * Copyright (C) 2000 David Mazieres (dm@uun.org)
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

#ifndef _SFSSERV_H
#define _SFSSERV_H 1

#include "sfsmisc.h"
#include "arpc.h"
#include "crypt.h"
#include "seqno.h"
#include "sfssesscrypt.h"

ptr<aclnt> getauthclnt ();

class sfsserv {
  ptr<sfspriv> privkey;
public:
  const ref<axprt_crypt> xc;
  const ref<axprt> x;
  const ref<bool> destroyed;
private:
  vec<size_t> authfreelist;
  sfs_hashcharge charge;
  ptr<sfs_servinfo_w> si;

protected:
  struct condat {
    sfs_connectinfo ci;
    sfs_connectres cr;
  };

  rpc_ptr<condat> cd;
  ref<asrv> sfssrv;
  void dispatch (svccb *sbp);
  virtual ptr<sfspriv> doconnect (const sfs_connectarg *,
				  sfs_servinfo *) = 0;

public:
  seqcheck seqstate;
  bool authid_valid;
  sfs_hash sessid;
  sfs_hash authid;
  vec<auto_auth> authtab;
  vec<sfsauth_cred> credtab;
  str client_name;
  ptr<aclnt> authc;

  explicit sfsserv (ref<axprt_crypt> xc, ptr<axprt> x = NULL);
  virtual ~sfsserv ();
  virtual u_int32_t authnoalloc ();
  virtual u_int32_t authalloc (const sfsauth_cred *cp, u_int n);
  virtual void authfree (size_t n);
  AUTH *getauth (size_t n) {
    return n < authtab.size () ? implicit_cast<AUTH *> (authtab[n])
      : (AUTH *) NULL;
  }

  virtual ptr<aclnt> getauthclnt () { return ::getauthclnt (); }
  virtual void sfs_connect (svccb *sbp);
  virtual void sfs_encrypt (svccb *sbp, int PVERS = 2); // see sfssesskey.C
  virtual void sfs_getfsinfo (svccb *sbp) { sbp->reject (PROC_UNAVAIL); }
  virtual void sfs_login (svccb *sbp);
  virtual void sfs_logout (svccb *sbp);
  virtual void sfs_idnames (svccb *sbp);
  virtual void sfs_idnums (svccb *sbp);
  virtual void sfs_getcred (svccb *sbp);
  virtual void sfs_badproc (svccb *sbp) { sbp->reject (PROC_UNAVAIL); }
};

typedef callback<void, ptr<axprt_crypt> >::ref sfsserv_cb;
void sfssd_slave (sfsserv_cb cb);
bool sfssd_slave (sfsserv_cb cb, bool allowstandalone, u_int port);
void sfssd_slavegen (str sock, sfsserv_cb cb);

typedef callback<ptr<axprt_crypt>, ptr<axprt_crypt> >::ref sfsserv_axprt_cb;
bool sfssd_slave_axprt (sfsserv_axprt_cb cb, bool allowstandalone, u_int port);
void sfssd_slavegen_axprt (str sock, sfsserv_axprt_cb cb);

#endif /* _SFSSERV_H */
