/* $Id: sfsgroupmgr.h 1754 2006-05-19 20:59:19Z max $ */

/*
 *
 * Copyright (C) 2003 Michael Kaminsky (kaminsky@lcs.mit.edu)
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

#ifndef _SFSMISC_SFSGROUPMGR_H
#define _SFSMISC_SFSGROUPMGR_H

#include "sfskeymisc.h"
#include "agentmisc.h"
#include "sfskeymgr.h"

typedef callback<void, ptr<sfsauth2_query_res>, str>::ref grpmgr_query_cb;

class sfsgroupmgr {
public:
  //sfsgroupmgr ();
  sfsgroupmgr (ptr<sfskeymgr> km) : kmgr (km) {};

  void query (str group, grpmgr_query_cb cb) { do_query (group, cb, false); }
  void expandedquery (str group, grpmgr_query_cb cb) { do_query (group, cb, true); }
  void changelogquery (str group, unsigned int vers, grpmgr_query_cb cb);
  void update (str group, vec<str> *members, vec<str> *owners, bool create = false);

  static bool parsegroup (str group, str *gname, str *ghost);

private:
  struct gstate {
    typedef callback<void, ptr<sfsauth2_query_res>, gstate *>::ref grpmgr_cb;

    gstate (ptr<sfscon> sc, ptr<sfspriv> k,
	    const str &n, const str &h, bool e, unsigned int v, grpmgr_cb cb) :
      key (k), gname (n), ghost (h), expanded (e), vers (v), cb (cb)
    { setcon (sc); gi.vers = 0; }

    void setcon (ptr<sfscon> sc) 
    { if (sc) { scon = sc; c = aclnt::alloc (sc->x, sfsauth_prog_2); } }

    ptr<sfscon> scon;
    ptr<aclnt> c;
    ptr<sfspriv> key;

    const str gname;
    const str ghost;
    bool expanded;
    unsigned int vers;
    unsigned int pending;

    sfsauth_groupinfo gi;

    grpmgr_cb cb;
  };

  ptr<sfskeymgr> kmgr;

  void do_query (str group, grpmgr_query_cb qcb, bool expanded);
  void connect_cb (gstate *st, str err, ptr<sfscon> sc);
  void clquery_login_cb (gstate *st, str err, ptr<sfscon> sc);
  void login_cb (gstate *st, str err, ptr<sfscon> sc, ptr<sfspriv> k);
  void getginfo (gstate *st);
  void gotginfo (ptr<sfsauth2_query_res> aqr, gstate *st, clnt_stat err);
  void query_done (grpmgr_query_cb qcb, ptr<sfsauth2_query_res> aqr, gstate *st);
  void getclinfo (gstate *st);
  void gotclinfo (ptr<sfsauth2_query_res> aqr, gstate *st, clnt_stat err);
  void clquery_done (grpmgr_query_cb qcb, ptr<sfsauth2_query_res> aqr, gstate *st);
  void update_group (vec<str> *members, vec<str> *owners, str gname,
                     ptr<sfsauth2_query_res> aqr, gstate *st);
  void gotsig (ptr<sfsauth2_update_arg> ua, gstate *st,
               str err, ptr<sfs_sig2> sig);
  void gotres (ptr<sfsauth2_update_res> ur, gstate *st, clnt_stat err);
};

#endif
