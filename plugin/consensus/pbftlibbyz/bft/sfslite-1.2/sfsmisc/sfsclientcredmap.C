/* $Id: sfsclientcredmap.C 1754 2006-05-19 20:59:19Z max $ */

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

#include "sfsclient.h"

ref<sfsserver_auth::userauth>
sfsserver_credmap::userauth_alloc (sfs_aid aid)
{
  ref<userauth_credmap> uap (New refcounted<userauth_credmap> (aid,
							       mkref (this)));
  uap->sendreq ();
  return uap;
}

void
sfsserver_credmap::flushstate ()
{
  uidmap.clear ();
  credmap.clear ();
  sfsserver_auth::flushstate ();
}

void
sfsserver_credmap::authclear (sfs_aid aid)
{
  if (ptr<sfsauth_cred> cred = credmap[aid])
    if (cred->type == SFS_UNIXCRED
	&& cred->unixcred->uid == (aid & 0xffffffff))
      uidmap.remove (aid & 0xffffffff);
  credmap.remove (aid);
  sfsserver_auth::authclear (aid);
}

bool
sfsserver_credmap::nomap (const authunix_parms *aup)
{
  if (!aup)
    return false;
  if (!aup->aup_uid) {
    if (ptr<sfsauth_cred> cred = credmap[aup2aid (aup)])
      if (cred->unixcred->uid == (aup2aid (aup) & 0xffffffff))
	return true;
    return false;
  }
  else if (ptr<sfsauth_cred> cred = credmap[aup2aid (aup)]) {
    if (cred->unixcred->uid != implicit_cast<u_int32_t> (aup->aup_uid)
	|| cred->unixcred->gid != implicit_cast<u_int32_t> (aup->aup_gid))
      return false;
    /* XXX - could maybe compare group lists, but things might be out
     * of order or shifted around, especially since some OSes
     * duplicate the egid as the first element of the grouplist. */
    return true;
  }
  return false;
}

void
sfsserver_credmap::mapcred (const authunix_parms *aup, ex_fattr3 *fp,
			    u_int32_t unknown_uid, u_int32_t unknown_gid)
{
  if (!fp || nomap (aup))
    return;
  if (u_int32_t *uidp = uidmap[fp->uid])
    fp->uid = *uidp;
  else
    fp->uid = unknown_uid;
  if (aup)
    if (ptr<sfsauth_cred> cred = credmap[aup2aid (aup)]) {
      bool ingroup = cred->unixcred->gid == fp->gid;
      for (u_int i = 0; !ingroup && i < cred->unixcred->groups.size (); i++)
	if (cred->unixcred->groups[i] == fp->gid)
	  ingroup = true;
      if (ingroup) {
	fp->gid = aup->aup_gid;	// XXX - no really good way to do this
	return;
      }
    }
  fp->gid = unknown_gid;
}

sfsserver_credmap
::userauth_credmap::userauth_credmap (sfs_aid aid,
				      const ref<sfsserver_credmap> &s)
  : userauth (aid, s), cred (New refcounted<sfsauth_cred>)
{
}

void
sfsserver_credmap::userauth_credmap::finish ()
{
  if (!authno)
    userauth::finish ();
  else {
    auto_auth aa (authuint_create (authno));
    cbase = sp->sfsc->call (SFSPROC_GETCRED, NULL, cred,
			    wrap (this, &sfsserver_credmap
				  ::userauth_credmap::cresult),
			    aa);
  }
}

void
sfsserver_credmap::userauth_credmap::cresult (clnt_stat err)
{
  cbase = NULL;
  if (!err && cred->type == SFS_UNIXCRED) {
    sfsserver_credmap *scp = static_cast<sfsserver_credmap *> (sp.get ());
    scp->credmap.insert (aid, cred);
    scp->uidmap.insert (cred->unixcred->uid, aid & 0xffffffff);
  }
  userauth::finish ();
}
