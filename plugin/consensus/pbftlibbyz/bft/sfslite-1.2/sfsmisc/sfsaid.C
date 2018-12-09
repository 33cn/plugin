/* $Id: sfsaid.C 1754 2006-05-19 20:59:19Z max $ */

/*
 *
 * Copyright (C) 1999 David Mazieres (dm@uun.org)
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

/* Agent ID (aid) is a generalized notion of user-ID which may take
 * sessions into account. */

#include "sfsmisc.h"

#ifdef HAVE_EGID_IN_GROUPLIST
/* On some operating systems, the first element of the grouplist is
 * the effective group id. */
const bool sfsaid_shift = true;
#else /* HAVE_EGID_IN_GROUPLIST */
const bool sfsaid_shift = false;
#endif /* HAVE_EGID_IN_GROUPLIST */

const sfs_aid sfsaid_sfs = INT64(0x8000000000000000);
const sfs_aid sfsaid_nobody = INT64(0x8000000000000001);

#if 0
enum { sfs_grouplist_size = 10 };
static gid_t sfs_maingroup;
static GETGROUPS_T sfs_grouplist[sfs_grouplist_size];
static bool sfs_groups_initialized;

void
sfsaid_init_groups ()
{
  assert (!sfs_groups_initialized);
  sfs_groups_initialized = true;
  random_init ();
  sfs_maingroup = getgid ();
  for (int i = 0; i < sfs_grouplist_size; i++)
    sfs_grouplist[i] = rnd.getword ();
}

void
sfsaid_set_groups ()
{
  assert (sfs_groups_initialized);
  if (setgroups (sfs_grouplist_size, sfs_grouplist) < 0)
    fatal ("setgroups: %m\n");
}
#endif

inline bool
isauxgid (u_int32_t gid)
{
  return gid >= sfs_resvgid_start
    && gid < sfs_resvgid_start + sfs_resvgid_count;
}

bool
sfs_specaid (sfs_aid aid)
{
  return aid & INT64(0x8000000000000000);
}

inline sfs_aid
mkaid (u_int32_t uid, u_int32_t gid)
{
  assert (isauxgid (gid));
  return implicit_cast<u_int64_t> (gid - sfs_resvgid_start + 1) << 32 | uid;
}

sfs_aid
sfs_mkaid (u_int32_t uid, u_int32_t gid)
{
  return mkaid (uid, gid);
}

static sfs_aid
calcaid (u_int32_t uid, u_int32_t gid, size_t ngid,
	 u_int32_t g0 = 0, u_int32_t g1 = 0, u_int32_t g2 = 0)
{
  const u_int32_t sfsgid = sfs_gid;

  if (uid == implicit_cast<u_int32_t> (sfs_uid)
      || (!uid && gid == implicit_cast<u_int32_t> (sfsgid)))
    return sfsaid_sfs;

  /* For uid = root, a grouplist of { UID, sfs_gid } means use the
   * same agent as uid = UID; */
  if (!uid && ngid == 2 && g1 == sfsgid)
    return g0;

  /* For uid = root, a grouplist of { UID, RESVGID, sfs_gid } means
   * use the same agent as uid = UID, grouplist = { RESVGID, ... } */
  if (!uid && ngid == 3 && isauxgid (g1) && g2 == sfsgid)
    return mkaid (g0, g1);

  if (ngid > 0 && isauxgid (g0))
    return mkaid (uid, g0);

  return uid;
}

sfs_aid
aup2aid (const authunix_parms *aup)
{
  if (!aup)
    return sfsaid_nobody;

  switch (aup->aup_len) {
  case 0:
    return calcaid (aup->aup_uid, aup->aup_gid, aup->aup_len);
  case 1:
    return calcaid (aup->aup_uid, aup->aup_gid, aup->aup_len,
		    aup->aup_gids[0]);
  case 2:
    return calcaid (aup->aup_uid, aup->aup_gid, aup->aup_len,
		    aup->aup_gids[0], aup->aup_gids[1]);
  default:
    return calcaid (aup->aup_uid, aup->aup_gid, aup->aup_len,
		    aup->aup_gids[0], aup->aup_gids[1], aup->aup_gids[2]);
  }
}

sfs_aid
myaid ()
{
  GETGROUPS_T group_buf[NGROUPS_MAX];
  GETGROUPS_T *groups = group_buf;
  int ngroups = getgroups (NGROUPS_MAX, groups);

  if (sfsaid_shift) {
    groups++;
    ngroups--;
  }
  if (ngroups < 0)
    ngroups = 0;

  switch (ngroups) {
  case 0:
    return calcaid (getuid (), getgid (), ngroups);
  case 1:
    return calcaid (getuid (), getgid (), ngroups,
		    groups[0]);
  case 2:
    return calcaid (getuid (), getgid (), ngroups,
		    groups[0], groups[1]);
  default:
    return calcaid (getuid (), getgid (), ngroups,
		    groups[0], groups[1], groups[2]);
  }
}
