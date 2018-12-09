/* $Id: authunixint.c 1117 2005-11-01 16:20:39Z max $ */

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

#include "sysconf.h"
#include <rpc/rpc.h>

#ifdef AUTHUNIX_GID_T
typedef gid_t authunix_gid_t;
#else /* !AUTHUNIX_GID_T */
typedef int32_t authunix_gid_t;
#endif /* !AUTHUNIX_GID_T */

AUTH *
authunixint_create (const char *host, u_int32_t uid, u_int32_t gid,
		    u_int32_t ngroups, const u_int32_t *groups)
{
  char *h = (char *) host;
  authunix_gid_t *gids;
  size_t i;
  AUTH *ret;

  if (ngroups > 16)
    ngroups = 16;
  if (sizeof (gid_t) == 4) {
    authunix_gid_t *gids = (authunix_gid_t *) groups;
    return authunix_create (h, uid, gid, ngroups, gids);
  }

  gids = malloc (ngroups * sizeof (*gids));
  if (!gids)
    return NULL;
  for (i = 0; i < ngroups; i++)
    gids[i] = groups[i];
  ret = authunix_create (h, uid, gid, ngroups, gids);
  free (gids);
  return ret;
}


/* On some OS's, authunix_create_default uses the effective rather
 * than the read uid and gid.  This makes authunix_create_default
 * unsuitable for use in setuid or setgid programs.  */
AUTH *
authunix_create_realids (void)
{
#ifdef HAVE_EGID_IN_GROUPLIST
  enum { first_group = 1 };
#else /* !HAVE_EGID_IN_GROUPLIST */
  enum { first_group = 0 };
#endif /* !HAVE_EGID_IN_GROUPLIST */
  u_int32_t uid = getuid (), gid = getgid ();
  GETGROUPS_T groups[NGROUPS_MAX];
  int ngroups;
  authunix_gid_t *gids;
  int i;
  AUTH *ret;

  ngroups = getgroups (NGROUPS_MAX, groups);
  if (ngroups < first_group)
    ngroups = first_group;

  if (sizeof (GETGROUPS_T) == sizeof (authunix_gid_t))
    return authunix_create ("localhost", uid, gid, ngroups - first_group,
			    (authunix_gid_t *) (groups + first_group));

  gids = malloc (ngroups * sizeof (*gids));
  if (!gids)
    return NULL;
  for (i = 0; i < ngroups; i++)
    gids[i] = groups[i];
  ret = authunix_create ("localhost", uid, gid,
			 ngroups - first_group, gids + first_group);
  free (gids);
  return ret;
}
