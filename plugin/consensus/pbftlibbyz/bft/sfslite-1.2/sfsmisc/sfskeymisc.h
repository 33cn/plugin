// -*-c++-*-
/* $Id: sfskeymisc.h 1754 2006-05-19 20:59:19Z max $ */

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

#ifndef _SFSMISC_SFSKEYMISC_H_
#define _SFSMISC_SFSKEYMISC_H_ 1

#include "sfscrypt.h"
#include "sfsmisc.h"
#include "sfsconnect.h"
#include "password.h"
#include "srp.h"

struct sfskey {
  ptr<sfspriv> key;
  str keyname;
  str pwd;
  u_int cost;
  rpc_ptr<sfssrp_parms> srpparms;
  str esk; 
  sfs_keytype pkt;
  ptr<eksblowfish> eksb;
  bool generated;
  sfskey () : cost (0), generated (false) {}
};

/* sfssrpconnect.C */
sfs_connect_t *sfs_connect_srp (str user, srp_client *srpp, sfs_connect_cb cb,
				str *userp = NULL, str *pwdp = NULL,
				bool *serverokp = NULL);
bool get_srp_params (ptr<aclnt> c, bigint *g, bigint *N);

/* sfskeyfetch.C */
bool issrpkey (const str &keyname);
bool iskeyremote (str keyname, bool longkeyok = false);
bool parse_userhost (str s, str *user = NULL, str *host = NULL);

str sfskeyfetch (sfskey *sk, str keyname, ptr<sfscon> *scp = NULL,
		 ptr<sfsauth_certinfores> *cip = NULL,
		 bool prompt = true, bool *warnp = NULL);

inline void
sfsfetchkey (str keyname, callback<void, ptr<sfskey>, str, 
	     ptr<sfscon> >::ref cb)
{
  ptr<sfskey> sk = New refcounted<sfskey>;
  ptr<sfscon> sc;
  if (str err = sfskeyfetch (sk, keyname, &sc))
    (*cb) (NULL, err, NULL);
  else
    (*cb) (sk, NULL, sc);
}

/* sfskeymisc.C */
extern vec<int> pwd_fds;
extern bool opt_pwd_fd;
str myusername ();
str sfskeysave (str path, const sfskey *sk, bool excl = true);
str sfskeygen (sfskey *sk, u_int nbits, str prompt = NULL, bool askname = true,
	       bool nopwd = false, bool nokbdnoise = false,
	       bool encrypt = true, sfs_keytype kt = SFS_RABIN);
str sfskeygen_prompt (sfskey *sk, str prompt = NULL, 
		      bool askname = true, bool nopwd = false, 
		      bool nokbdnoise = false);
str defkeyname (str user = NULL);
void rndkbd ();
str getpwd (str prompt);
str getpwdconfirm (str prompt);
str getline (str prompt, str def);


#endif /* _SFSMISC_SFSKEYMISC_H_ */
