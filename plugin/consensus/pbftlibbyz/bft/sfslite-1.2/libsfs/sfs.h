/* -*-c++-*- */
/* $Id: sfs.h 435 2004-06-02 15:46:36Z max $ */

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

#ifndef _SFS_H_INCLUDED_
#define _SFS_H_INCLUDED_ 1

#ifdef __cplusplus
extern "C" {
#endif /* __cplusplus */

#include <sys/types.h>
#include <sys/stat.h>

struct stat;

struct sfs_desc {
  char fsname[256];
  unsigned fhsize;
  char fhdata[64];
};
typedef struct sfs_desc sfs_desc;
int sfs_getdesc (struct sfs_desc *sdp, const char *path);
int sfs_lgetdesc (struct sfs_desc *sdp, const char *path);
int sfs_fgetdesc (struct sfs_desc *sdp, int fd);

#define sfs_idnamelen 32
struct sfs_names {
  char uidname[sfs_idnamelen + 2];
  char gidname[sfs_idnamelen + 2];
};
typedef struct sfs_names sfs_names;
int sfs_stat2names (sfs_names *snp, const struct stat *sb);
int sfs_uidbyname (const char *uidname, dev_t dev);
int sfs_gidbyname (const char *gidname, dev_t dev);

#define sfs_maxgroups 16
struct sfs_remoteid {
  int valid;
  unsigned uid;
  unsigned gid;
  unsigned ngroups;
  unsigned groups[sfs_maxgroups];
};
typedef struct sfs_remoteid sfs_remoteid;
int sfs_getremoteid (sfs_remoteid *rip, dev_t dev);
void sfs_flush_idcache ();

#ifdef __cplusplus
}
#endif /* __cplusplus */

#endif /* !_SFS_H_INCLUDED_ */
