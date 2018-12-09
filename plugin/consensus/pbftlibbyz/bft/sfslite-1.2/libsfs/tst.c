/* $Id: tst.c 435 2004-06-02 15:46:36Z max $ */

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

#include "sfs-internal.h"

#if 0

static void
getfh (const char *path)
{
  int fd;
  sfs_desc sd;

  fd = open (path, O_RDONLY);
  if (fd < 0) {
    perror (path);
    return;
  }
  if (sfs_fgetdesc (&sd, fd) < 0)
    fprintf (stderr, "%s: not found\n", path);
  else {
    unsigned i;
    printf ("%s %s ", path, sd.fsname);
    for (i = 0; i < sd.fhsize; i++)
      printf ("%02x", (u_char) sd.fhdata[i]);
    printf ("\n");
  }

  close (fd);
}

static void
getfh2 (const char *path)
{
  sfs_desc sd;
  if (sfs_getdesc (&sd, path) < 0)
    fprintf (stderr, "%s: not found\n", path);
  else {
    unsigned i;
    printf ("%s %s ", path, sd.fsname);
    for (i = 0; i < sd.fhsize; i++)
      printf ("%02x", (u_char) sd.fhdata[i]);
    printf ("\n");
  }

}

static void
pstat (const char *path)
{
  struct stat sb;
  struct sfs_names sn;

  if (stat (path, &sb) < 0) {
    perror (path);
    return;
  }
  sfs_stat2names (&sn, &sb);

  printf ("%s %s %s\n", sn.uidname, sn.gidname, path);
}

static void
pids (const char *path)
{
  struct stat sb;
  sfs_remoteid id;

  if (stat (path, &sb) < 0) {
    perror (path);
    return;
  }
  sfs_getremoteid (&id, sb.st_dev);
  if (id.valid) {
    printf ("%s: uid = %u, gid = %u", path, id.uid, id.gid);
    if (id.ngroups > 0) {
      u_int i;
      printf (", groups = { %u", id.groups[0]);
      for (i = 1; i < id.ngroups; i++)
	printf (", %u", id.groups[i]);
      printf (" }\n");
    }
    else
      printf ("\n");
  }
  else
    printf ("%s: no credentials\n", path);
}

int
main (int argc, char **argv)
{
  int i;

  if (argc <= 1) {
    getfh (".");
    getfh2 (".");
    pstat (".");
    pids (".");
  }
  else
    for (i = 1; i < argc; i++) {
      getfh (argv[i]);
      getfh2 (argv[i]);
      pstat (argv[i]);
      pids (argv[i]);
    }
  return 0;
}

#else

int
main (int argc, char **argv)
{
  return 0;
}

#endif
