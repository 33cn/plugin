/* $Id: sfsgroupmgr.C 1754 2006-05-19 20:59:19Z max $ */

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

#include "parseopt.h"
#include "crypt.h"
#include "sfskeymisc.h"
#include "sfsgroupmgr.h"
#include "rxx.h"

#include <unistd.h>
#include <sys/types.h>
#include <dirent.h>

bool
sfsgroupmgr::parsegroup (str group, str *gname, str *ghost)
{
  static rxx grouphostrx ("([A-Za-z][\\-\\w\\.]{0,31})(@[A-Za-z].+)?");
  if (!grouphostrx.match (group)) {
    warn << "Could not parse group[@host]: " << group << "\n";
    return false;
  }
  *gname = grouphostrx[1];
  *ghost = grouphostrx[2] ? grouphostrx[2].cstr () : "-";
  return true;
}

void
sfsgroupmgr::do_query (str group, grpmgr_query_cb qcb, bool expanded)
{
  str gname;
  str ghost;

  if (!parsegroup (group, &gname, &ghost)) {
    (*qcb) (NULL, NULL);
    return;
  }
  gstate *st = New gstate (NULL, NULL, gname, ghost, expanded, 0,
                           wrap (this, &sfsgroupmgr::query_done, qcb));

  if (issrpkey (ghost))
    kmgr->login (ghost, wrap (this, &sfsgroupmgr::login_cb, st),
	         NULL, KM_NOSRP | KM_NOESK);
  else
    kmgr->connect (ghost, wrap (this, &sfsgroupmgr::connect_cb, st));
}

void
sfsgroupmgr::connect_cb (gstate *st, str err, ptr<sfscon> sc)
{
  if (err) {
    warn << st->ghost << ": could not connect: " << err << "\n";
    (*st->cb) (NULL, st);
    return;
  }
  st->setcon (sc);
  getginfo (st);
}

void
sfsgroupmgr::changelogquery (str group, unsigned int vers, grpmgr_query_cb qcb)
{
  str gname;
  str ghost;

  if (!parsegroup (group, &gname, &ghost)) {
    (*qcb) (NULL, NULL);
    return;
  }
  gstate *st = New gstate (NULL, NULL, gname, ghost, false, vers,
                           wrap (this, &sfsgroupmgr::clquery_done, qcb));

  kmgr->connect (ghost, wrap (this, &sfsgroupmgr::clquery_login_cb, st));
}

void
sfsgroupmgr::clquery_login_cb (gstate *st, str err, ptr<sfscon> sc)
{
  if (err) {
    warn << st->ghost << ": could not login: " << err << "\n";
    (*st->cb) (NULL, st);
    return;
  }
  st->setcon (sc);
  getclinfo (st);
}

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
sfsgroupmgr::login_cb (gstate *st, str err, ptr<sfscon> sc, ptr<sfspriv> k)
{
  if (err) {
    warn << st->ghost << ": could not login:\n" << err << "\n";
    (*st->cb) (NULL, st);
    return;
  }
  st->setcon (sc);
  st->key = k;
  getginfo (st);
}

void
sfsgroupmgr::getginfo (gstate *st)
{
  assert (st);
  assert (st->c);

  sfsauth2_query_arg aqa;
  aqa.type = st->expanded ? SFSAUTH_EXPANDEDGROUP : SFSAUTH_GROUP;
  aqa.key.set_type (SFSAUTH_DBKEY_NAMEVERS);
  aqa.key.namevers->name = st->gname;
  aqa.key.namevers->vers = st->vers;

  ref<sfsauth2_query_res> aqr = New refcounted<sfsauth2_query_res>;
  st->c->call (SFSAUTH2_QUERY, &aqa, aqr,
	       wrap (this, &sfsgroupmgr::gotginfo, aqr, st), st->scon->auth);
}

void
sfsgroupmgr::gotginfo (ptr<sfsauth2_query_res> aqr, gstate *st, clnt_stat err)
{
  if (err) {
    warn << err << "\n";
    (*st->cb) (NULL, st);
    return;
  }
  if (!aqr || (aqr->type != SFSAUTH_GROUP && aqr->type != SFSAUTH_ERROR)) {
    warn << "Unexpected response from server\n";
    (*st->cb) (NULL, st);
    return;
  }
  if (aqr->type == SFSAUTH_ERROR) {
    (*st->cb) (aqr, st);
    return;
  }

  if (st->gi.vers && st->gi.vers != aqr->groupinfo->vers) {
    warn << "Group was updated while fetching multiple chunks.\n";
    (*st->cb) (NULL, st);
    return;
  }
  else
    st->gi.vers = aqr->groupinfo->vers;

  while (aqr->groupinfo->owners.size () > 0)
    st->gi.owners.push_back (aqr->groupinfo->owners.pop_front ());
  while (aqr->groupinfo->members.size () > 0
         && !(aqr->groupinfo->members.front () == "..."))
    st->gi.members.push_back (aqr->groupinfo->members.pop_front ());
  if (aqr->groupinfo->members.size () > 0) {
    st->vers = st->gi.members.size () + st->gi.owners.size ();
    getginfo (st);
  }
  else {
    aqr->groupinfo->members = st->gi.members;
    aqr->groupinfo->owners = st->gi.owners;
    (*st->cb) (aqr, st);
  }
}

void
sfsgroupmgr::query_done (grpmgr_query_cb qcb,
                         ptr<sfsauth2_query_res> aqr, gstate *st)
{
  if (!aqr)
    (*qcb) (NULL, NULL);
  else
    (*qcb) (aqr, st->scon->path);
  delete st;
}

void
sfsgroupmgr::getclinfo (gstate *st)
{
  assert (st);
  assert (st->c);

  sfsauth2_query_arg aqa;
  aqa.type = SFSAUTH_LOGENTRY;
  aqa.key.set_type (SFSAUTH_DBKEY_NAMEVERS);
  aqa.key.namevers->name = st->gname;
  aqa.key.namevers->vers = st->vers;

  ref<sfsauth2_query_res> aqr = New refcounted<sfsauth2_query_res>;
  st->c->call (SFSAUTH2_QUERY, &aqa, aqr,
	       wrap (this, &sfsgroupmgr::gotclinfo, aqr, st));
}

void
sfsgroupmgr::gotclinfo (ptr<sfsauth2_query_res> aqr, gstate *st, clnt_stat err)
{
  if (err) {
    warn << err << "\n";
    (*st->cb) (NULL, st);
    return;
  }
  if (!aqr || (aqr->type != SFSAUTH_GROUP
	       && aqr->type != SFSAUTH_LOGENTRY
	       && aqr->type != SFSAUTH_ERROR)) {
    warn << "Unexpected response from server\n";
    (*st->cb) (NULL, st);
    return;
  }

  if (aqr->type == SFSAUTH_LOGENTRY) {
    while (aqr->logentry->members.size () > 0)
      st->gi.members.push_back (aqr->logentry->members.pop_front ());
    if (aqr->logentry->more) {
      st->vers = aqr->logentry->vers;
      getclinfo (st);
    }
    else {
      // audit string doesn't have correct starting version
      aqr->logentry->members = st->gi.members;
      (*st->cb) (aqr, st);
    }
  }
  else
    gotginfo (aqr, st, err);
}

void
sfsgroupmgr::clquery_done (grpmgr_query_cb qcb,
                           ptr<sfsauth2_query_res> aqr, gstate *st)
{
  if (!aqr)
    (*qcb) (NULL, NULL);
  else
    (*qcb) (aqr, st->scon->path);
  delete st;
}

static void
update_fixlist (sfs_groupmembers &o, vec<str> &n, unsigned int &current)
{
  // 250 is the maximum number of members+owners allowed in a single
  // RPC message; 250 * 256 (characters) < 64K (max RPC size)
  unsigned int spaces = 250 - current > 0 ? 250 - current : 0;
  if (spaces > n.size ())
    spaces = n.size ();

  o.setsize (spaces);

  for (unsigned int i = 0; i < spaces; i++)
    o[i] = n.pop_front ();

  current = current + spaces;
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
