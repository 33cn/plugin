/* $Id: sfsusermgr.h 1754 2006-05-19 20:59:19Z max $ */

/*
 *
 * Copyright (C) 2004 Michael Kaminsky (kaminsky@lcs.mit.edu)
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

#ifndef _SFSMISC_SFSUSERMGR_H
#define _SFSMISC_SFSUSERMGR_H

#include "sfskeymisc.h"
#include "agentmisc.h"
#include "sfskeymgr.h"

typedef callback<void, ptr<sfsauth2_query_res>, str>::ref authdb_query_cb;

class sfsusermgr {
public:
  sfsusermgr (ptr<sfskeymgr> km) : kmgr (km) {};

  void query (str group, authdb_query_cb cb);
  //void update (str group, vec<str> *members, vec<str> *owners, bool create = false);

  static bool parseuser (str user, str *name, str *host);

private:
  struct ustate {
    typedef callback<void, ptr<sfsauth2_query_res>, ustate *>::ref usermgr_cb;

    ustate (ptr<sfscon> sc, ptr<sfspriv> k, const str &n, const str &h,
            usermgr_cb cb) : key (k), name (n), host (h), cb (cb)
    { setcon (sc); }

    void setcon (ptr<sfscon> sc) 
    { if (sc) { scon = sc; c = aclnt::alloc (sc->x, sfsauth_prog_2); } }

    ptr<sfscon> scon;
    ptr<aclnt> c;
    ptr<sfspriv> key;

    const str name;
    const str host;

    usermgr_cb cb;
  };

  ptr<sfskeymgr> kmgr;

  void connect_cb (ustate *st, str err, ptr<sfscon> sc);
  void login_cb (ustate *st, str err, ptr<sfscon> sc, ptr<sfspriv> k);
  void getuinfo (ustate *st);
  void gotuinfo (ptr<sfsauth2_query_res> aqr, ustate *st, clnt_stat err);
  void query_done (authdb_query_cb qcb, ptr<sfsauth2_query_res> aqr, ustate *st);
#if 0
  void update_group (vec<str> *members, vec<str> *owners, str name,
                     ptr<sfsauth2_query_res> aqr, ustate *st);
  void gotsig (ptr<sfsauth2_update_arg> ua, ustate *st,
               str err, ptr<sfs_sig2> sig);
  void gotres (ptr<sfsauth2_update_res> ur, ustate *st, clnt_stat err);
#endif
};

#endif
