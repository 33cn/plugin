// -*-c++-*-
/* $Id: getfh3.h 1754 2006-05-19 20:59:19Z max $ */

/*
 *
 * Copyright (C) 2001 David Mazieres (dm@uun.org)
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

#ifndef _SFSMISC_GETFH3_H_
#define _SFSMISC_GETFH3_H_ 1

#include "arpc.h"
#include "nfs3_prot.h"
#include "mount_prot.h"

/*
 * getfh3.C
 */
const strbuf &strbuf_cat (const strbuf &, mountstat3);
void getfh3 (str host, str path,
	     callback<void, const nfs_fh3 *, str>::ref);
void getfh3 (ref<aclnt> c, str path,
	     callback<void, const nfs_fh3 *, str>::ref);
void lookupfh3 (ref<aclnt> c, const nfs_fh3 &start, str path,
		callback<void, const nfs_fh3 *,
		const fattr3exp *, str>::ref cb,
		const authunix_parms *aup = NULL);

/*
 * findfs.C
 */
struct nfsinfo {
  const ref<aclnt> c;
  str hostname;
  const nfs_fh3 fh;
  const u_int64_t rdev;
  int fd;
  nfsinfo (ref<aclnt> cc, const str &hh, const nfs_fh3 &ff, u_int64_t dd)
    : c (cc), hostname (hh), fh (ff), rdev (dd), fd (-1) {}
  ~nfsinfo () { close (fd); }
};
typedef callback<void, ptr<nfsinfo>, str>::ref findfscb_t;
enum {
  FINDFS_NOLOCAL = 1,
  FINDFS_NOSFS = 2,
};
void findfs (const authunix_parms *aup, str path,
	     findfscb_t cb, int flags = 0);

void pathinfofetch (const authunix_parms *aup, str path,
		    callback<void, u_int64_t, str, str>::ref cb,
		    int *fdp = NULL);

#endif /* !_SFSMISC_GETFH3_H_ */
