/* $Id: sfsusermgr.C 1754 2006-05-19 20:59:19Z max $ */

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

#include "parseopt.h"
#include "crypt.h"
#include "sfskeymisc.h"
#include "sfsusermgr.h"
#include "rxx.h"

#include <unistd.h>
#include <sys/types.h>
#include <dirent.h>

bool
sfsusermgr::parseuser (str user, str *name, str *host)
{
  static rxx userhostrx ("([A-Za-z][\\-\\w\\.]{0,31})(@.*)?$");
  if (!userhostrx.match (user)) {
    warn << "Could not parse user[@host]: " << user << "\n";
    return false;
  }
  *name = userhostrx[1];
  *host = userhostrx[2] ? userhostrx[2].cstr () : "-";
  return true;
}

void
sfsusermgr::query (str user, authdb_query_cb qcb)
{
  str name;
  str host;

  if (!parseuser (user, &name, &host)) {
    (*qcb) (NULL, NULL);
    return;
  }
  ustate *st = New ustate (NULL, NULL, name, host,
                           wrap (this, &sfsusermgr::query_done, qcb));

  if (issrpkey (host))
    kmgr->login (host, wrap (this, &sfsusermgr::login_cb, st),
	         NULL, KM_NOSRP | KM_NOESK);
  else
    kmgr->connect (host, wrap (this, &sfsusermgr::connect_cb, st));
}

void
sfsusermgr::connect_cb (ustate *st, str err, ptr<sfscon> sc)
{
  if (err) {
    warn << st->host << ": could not connect: " << err << "\n";
    (*st->cb) (NULL, st);
    return;
  }
  st->setcon (sc);
  getuinfo (st);
}

void
sfsusermgr::login_cb (ustate *st, str err, ptr<sfscon> sc, ptr<sfspriv> k)
{
  if (err) {
    warn << st->host << ": could not login:\n" << err << "\n";
    (*st->cb) (NULL, st);
    return;
  }
  st->setcon (sc);
  st->key = k;
  getuinfo (st);
}

void
sfsusermgr::getuinfo (ustate *st)
{
  assert (st);
  assert (st->c);

  sfsauth2_query_arg aqa;
  aqa.type = SFSAUTH_USER;
  aqa.key.set_type (SFSAUTH_DBKEY_NAME);
  *aqa.key.name = st->name;

  ref<sfsauth2_query_res> aqr = New refcounted<sfsauth2_query_res>;
  st->c->call (SFSAUTH2_QUERY, &aqa, aqr,
	       wrap (this, &sfsusermgr::gotuinfo, aqr, st), st->scon->auth);
}

void
sfsusermgr::gotuinfo (ptr<sfsauth2_query_res> aqr, ustate *st, clnt_stat err)
{
  if (err) {
    warn << err << "\n";
    (*st->cb) (NULL, st);
    return;
  }
  if (!aqr || (aqr->type != SFSAUTH_USER && aqr->type != SFSAUTH_ERROR)) {
    warn << "Unexpected response from server\n";
    (*st->cb) (NULL, st);
    return;
  }

  (*st->cb) (aqr, st);
}

void
sfsusermgr::query_done (authdb_query_cb qcb,
                         ptr<sfsauth2_query_res> aqr, ustate *st)
{
  if (!aqr)
    (*qcb) (NULL, NULL);
  else
    (*qcb) (aqr, st->scon->path);
  delete st;
}

#if 0
void
sfsgroupmgr::update (str group, vec<str> *members, vec<str> *owners, bool create)
{
  str gname;
  str ghost;

  if (!parsegroup (group, &gname, &ghost)) {
    exit (1);
  }
  gstate *st = New gstate (NULL, NULL, gname, ghost, false, 0,
                           wrap (this, &sfsgroupmgr::update_group,
			         members, owners, create ? gname : str (NULL)));
  kmgr->login (ghost,
               wrap (this, &sfsgroupmgr::login_cb, st),
	       NULL, KM_NOSRP | KM_NOESK);
}

void
sfsgroupmgr::update_group (vec<str> *members, vec<str> *owners, str gname,
                           ptr<sfsauth2_query_res> aqr, gstate *st)
{
  bool create = false;

  if (!aqr) {
    delete st;
    exit (1);
  }
  else if (aqr->type == SFSAUTH_ERROR) {
    if (gname && *aqr->errmsg == "group not found") {
      warn << "Group `" << gname << "' not found.  Creating new group.\n";
      aqr->set_type (SFSAUTH_GROUP);
      aqr->groupinfo->name = gname;
      aqr->groupinfo->id = 0;    // assigned by server
      aqr->groupinfo->vers = 0;
      create = true;
    }
    else {
      warn << "Error: " << *aqr->errmsg << "\n";
      delete st;
      exit (1);
    }
  }
  else if (aqr->type != SFSAUTH_GROUP) {
    warn << "Error: server returned a non-group record\n";
    delete st;
    exit (1);
  }
  else if (gname) {
    warn << "Error: group " << gname << " already exists\n";
    delete st;
    exit (1);
  }

  assert (st->key);

  unsigned int remaining = owners->size () + members->size ();
  unsigned int version = aqr->groupinfo->vers;
  st->pending = 0;

  while (remaining > 0 || create) {
    unsigned int current = 0;
    ref<sfsauth2_update_arg> ua = New refcounted<sfsauth2_update_arg>;

    ua->req.opts = SFSUP_KPSRP | SFSUP_KPESK | SFSUP_KPPK;
    ua->req.type = SFS_UPDATEREQ;
    ua->req.authid = st->scon->authid;
    ua->req.rec.set_type (SFSAUTH_GROUP);

    // Update groupinfo
    *ua->req.rec.groupinfo = *aqr->groupinfo;

    st->pending++;

    ua->req.rec.groupinfo->vers = ++version;
    if (!create) {
      update_fixlist (ua->req.rec.groupinfo->owners, *owners, current);
      update_fixlist (ua->req.rec.groupinfo->members, *members, current);
    }

    sfsauth2_sigreq sr;
    sr.set_type (SFS_UPDATEREQ);
    *sr.updatereq = ua->req;

    ua->authsig.alloc ();
    st->key->sign (sr, st->scon->authinfo, 
	           wrap (this, &sfsgroupmgr::gotsig, ua, st));

    remaining = remaining - current;

    //warn << "SIGN: rem = " << remaining << "; curr = " << current << "\n";

    if (create)
      break;
  }
}

void
sfsgroupmgr::gotsig (ptr<sfsauth2_update_arg> ua, gstate *st,
                     str err, ptr<sfs_sig2> sig)
{
  if (err) {
    delete st;
    fatal << "Signing error: " << err << "\n";
  }
  if (sig)
    *ua->authsig = *sig;

  ref<sfsauth2_update_res> ur = New refcounted<sfsauth2_update_res>;
  st->c->call (SFSAUTH2_UPDATE, ua, ur,
	       wrap (this, &sfsgroupmgr::gotres, ur, st), st->scon->auth);
}

void
sfsgroupmgr::gotres (ptr<sfsauth2_update_res> ur, gstate *st, clnt_stat err)
{
  if (err)
    fatal << "Update error: " << err << "\n";
  if (!ur->ok)
    fatal << "Update error: " << *ur->errmsg << "\n";

  if (!--st->pending) {
    delete st;
    warn << "Group update sucessfull\n";
    exit (0);
  }
}
#endif
