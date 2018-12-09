/* $Id: sfsops.c 435 2004-06-02 15:46:36Z max $ */

/*
 *
 * Copyright (C) 1999 David Mazieres (dm@uun.org)
 * Copyright (C) 2000 Kevin Fu (fubob@mit.edu)
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
#include "sfsagent.h"
#include "hashtab.h"
#include <grp.h>

#define SFSPREF ".SFS "
#define SFSFH SFSPREF "FH"
#define SFSFS SFSPREF "FS"
#define SFSSPECLEN (2 + sizeof (SFSFH))

#define MIN2(a, b) ((a) < (b) ? (a) : (b))

struct id_dat {
  char name [sfs_idnamelen + 2];
  unsigned int id;
  struct hashtab_entry id_link;
  struct hashtab_entry name_link;
};

hashtab_decl (id2name, id_dat, id_link);
hashtab_decl (name2id, id_dat, name_link);

struct id_cache {
  dev_t dev;

  struct id2name uidtab;
  struct id2name gidtab;
  struct name2id unametab;
  struct name2id gnametab;

  struct hashtab_entry link;
};

hashtab_decl (dev2ic, id_cache, link) ictab;


struct id_cache *
lookup_ic (dev_t dev)
{
  struct id_cache *ic_elm;

  for (ic_elm = dev2ic_chain (&ictab, dev); 
       ic_elm && (dev != ic_elm->dev);
       ic_elm = dev2ic_next (ic_elm))
    ;
  if (ic_elm)
    return ic_elm;

  ic_elm = malloc (sizeof (*ic_elm));
  if (!ic_elm)
    return NULL;

  ic_elm->dev = dev;
  id2name_init (&ic_elm->uidtab);
  id2name_init (&ic_elm->gidtab);
  name2id_init (&ic_elm->unametab);
  name2id_init (&ic_elm->gnametab);

  dev2ic_insert (&ictab, ic_elm, dev);
  return ic_elm;
}

static void
namecpy (char *dst, const char *src)
{
  strncpy (dst, src, sfs_idnamelen);
  dst[sfs_idnamelen] = '\0';
}


int
lookup_id (struct id_cache *ic, bool_t gid, char *name)
{
  struct id_dat *id_elm;
  char *orig_name = name;
  const char *fs;
  int cdfd;
  u_int hval;

  if (*name == '%')
    name++;

  hval = hash_string (name);

  for (id_elm = name2id_chain ((gid ? &ic->gnametab : &ic->unametab), 
			       hval); 
       id_elm && strcmp (name, id_elm->name);
       id_elm = name2id_next (id_elm))
    ;
  if (id_elm)
    return id_elm->id;

  id_elm = malloc (sizeof (*id_elm));
  if (!id_elm)
    return -1;

  bzero (id_elm->name, sizeof (id_elm->name));
  namecpy (id_elm->name, name);

  name = orig_name;

  /* Get the id either locally or via SFS */
  if (*name != '%' || (++name, !devcon_lookup (&cdfd, &fs, ic->dev))) {
    if (gid) {
      struct group *gr = getgrnam (name);
      if (gr)
	id_elm->id = gr->gr_gid;
      else
	id_elm->id = -1;
    }
    else {
      struct passwd *pw = getpwnam (name);
      if (pw)
	id_elm->id = pw->pw_uid;
      else 
	id_elm->id = -1;
    }
  }
  else if (strlen (name) > sfs_idnamelen)
    {
      free (id_elm);
      return -1;
    }
  else {

    sfsctl_getidnums_arg arg;
    sfsctl_getidnums_res res;
    sfs_opt_idname *sid;

    bzero (&arg, sizeof (arg));
    bzero (&res, sizeof (res));
    arg.filesys = (char *) fs;
    sid = gid ? &arg.names.gidname : &arg.names.uidname;
    sid->present = TRUE;
    sid->u.name = (char *) name;
    
    srpc_call (&sfsctl_prog_1, cdfd, SFSCTL_GETIDNUMS, &arg, &res);

    if (!res.status)
      id_elm->id = gid ? res.u.nums.gid : res.u.nums.uid;
    else {
      free (id_elm);
      xdr_free ((xdrproc_t) xdr_sfsctl_getidnums_res, (char *) &res);
      return -1;
    }

    
    xdr_free ((xdrproc_t) xdr_sfsctl_getidnums_res, (char *) &res);


  }

  id2name_insert (gid ? &ic->gidtab   : &ic->uidtab,   id_elm, id_elm->id);
  name2id_insert (gid ? &ic->gnametab : &ic->unametab, id_elm, hval);

  return id_elm->id;
}



void
delid (void *_ic, struct id_dat *i)
{
  free (i);
}


void 
delic (void *_null, struct id_cache *ic)
{

  id2name_traverse (&ic->uidtab, delid, ic);
  id2name_clear (&ic->uidtab);
  id2name_init (&ic->uidtab);

  id2name_traverse (&ic->gidtab, delid, &ic->gidtab); 
  id2name_clear (&ic->gidtab);
  id2name_init (&ic->gidtab);

  dev2ic_delete (&ictab, ic);
  free (ic);
}

/* Delete the entire cache and free all cache memory */
void
sfs_flush_idcache ()
{
  dev2ic_traverse (&ictab, delic, NULL);
}

#if 0
static int
sfs_do_getdesc (struct sfs_desc *sdp, const char *path,
		int (*statfn) (const char *, struct stat *),
		int (*chownfn) (const char *, uid_t, gid_t))
{
  struct stat sb;
  const char *fs;
  int cdfd;
  enum clnt_stat err;
  sfsctl_getfh_arg arg;
  sfsctl_getfh_res res;
  int saved_errno;
  int ret;

  if (statfn (path, &sb) < 0)
    return -1;
  if (!devcon_lookup (&cdfd, &fs, sb.st_dev)) {
    errno = EINVAL;
    return -1;
  }

  saved_errno = errno;
  chownfn (path, -2, getpid ());
  if (errno != EPERM)
    return -1;
  errno = saved_errno;

  bzero (&arg, sizeof (arg));
  bzero (&res, sizeof (res));
  arg.filesys = (char *) fs;
  arg.fileid = sb.st_ino;
  err = srpc_call (&sfsctl_prog_1, cdfd, SFSCTL_GETFH, &arg, &res);

  if (!err && !res.status) {
    strcpy (sdp->fsname, fs);
    sdp->fhsize = res.u.fh.data.len;
    memcpy (sdp->fhdata, res.u.fh.data.val, sdp->fhsize);
    ret = 0;
  }
  else {
    errno = res.status == NFS3ERR_JUKEBOX ? EAGAIN : EIO;
    ret = -1;
  }

  xdr_free ((xdrproc_t) xdr_sfsctl_getfh_res, (char *) &res);
  return ret;
}

int
sfs_getdesc (struct sfs_desc *sdp, const char *path)
{
  assert (!"unsafe");
  return sfs_do_getdesc (sdp, path, stat, chown);
}

int
sfs_lgetdesc (struct sfs_desc *sdp, const char *path)
{
  assert (!"unsafe");
  return sfs_do_getdesc (sdp, path, lstat, lchown);
}

int
sfs_fgetdesc (sfs_desc *sdp, int fd)
{
  struct stat sb;
  const char *fs;
  int cdfd;
  enum clnt_stat err;
  sfsctl_getfh_arg arg;
  sfsctl_getfh_res res;
  int saved_errno;
  int ret;

  assert (!"unsafe");
  if (fstat (fd, &sb) < 0)
    return -1;
  if (!devcon_lookup (&cdfd, &fs, sb.st_dev)) {
    errno = EINVAL;
    return -1;
  }

  saved_errno = errno;
  fchown (fd, -2, getpid ());
  if (errno != EPERM) {
    close (fd);
    return -1;
  }
  errno = saved_errno;

  bzero (&arg, sizeof (arg));
  bzero (&res, sizeof (res));
  arg.filesys = (char *) fs;
  arg.fileid = sb.st_ino;
  err = srpc_call (&sfsctl_prog_1, cdfd, SFSCTL_GETFH, &arg, &res);

  if (!err && !res.status) {
    strcpy (sdp->fsname, fs);
    sdp->fhsize = res.u.fh.data.len;
    memcpy (sdp->fhdata, res.u.fh.data.val, sdp->fhsize);
    ret = 0;
  }
  else {
    errno = res.status == NFS3ERR_JUKEBOX ? EAGAIN : EIO;
    ret = -1;
  }

  xdr_free ((xdrproc_t) xdr_sfsctl_getfh_res, (char *) &res);
  return ret;
}
#endif


/* Return -1 on error, 0 otherwise */
int
sfs_stat2names (sfs_names *snp, const struct stat *sb)
{


  const char *fs;
  int cdfd;
  sfsctl_getidnames_res res;
  struct passwd *pw;
  struct group *gr;
  bool_t remote = FALSE;

  bool_t uname = FALSE;
  bool_t gname = FALSE;
  struct id_dat *id_elm;
  struct id_cache *ic;
  u_int hval;

  ic = lookup_ic (sb->st_dev);
  
  if (!ic)
    return -1;

  for (id_elm = id2name_chain (&ic->uidtab, sb->st_uid); 
       id_elm && (sb->st_uid != id_elm->id);
       id_elm = id2name_next (id_elm))
    ;
  if (id_elm)
    {	 
      namecpy (snp->uidname, id_elm->name);
      uname = TRUE;
    }


  for (id_elm = id2name_chain (&ic->gidtab, sb->st_gid); 
       id_elm && (sb->st_gid != id_elm->id);
       id_elm = id2name_next (id_elm))
    ;
  if (id_elm)
    {	 
      namecpy (snp->gidname, id_elm->name);
      gname = TRUE;
    }


  if (gname && uname)
    return 0;

  bzero (&res, sizeof (res));
  if (devcon_lookup (&cdfd, &fs, sb->st_dev)) {
    sfsctl_getidnames_arg arg;
    remote = TRUE;
    arg.filesys = (char *) fs;
    arg.nums.uid = sb->st_uid;
    arg.nums.gid = sb->st_gid;
    srpc_call (&sfsctl_prog_1, cdfd, SFSCTL_GETIDNAMES, &arg, &res);
  }

  if (!uname)
    {
      pw = getpwuid (sb->st_uid);
      if (remote && !res.status && res.u.names.uidname.present) {
	if (pw && !strcmp (pw->pw_name, res.u.names.uidname.u.name))
	  namecpy (snp->uidname, pw->pw_name);
	else {
	  snp->uidname[0] = '%';
	  namecpy (snp->uidname + 1, res.u.names.uidname.u.name);
	}
      }
      else if (!remote && pw)
	namecpy (snp->uidname, pw->pw_name);
      else
	sprintf (snp->uidname, "%u", (int) sb->st_uid);

      id_elm = malloc (sizeof (*id_elm));

      if (!id_elm)
	{
	  xdr_free ((xdrproc_t) xdr_sfsctl_getidnames_res, (char *) &res);
	  return -1;
	}

      bzero (id_elm->name, sizeof (id_elm->name));
      namecpy (id_elm->name, snp->uidname);
      id_elm->id = sb->st_uid;
      hval = hash_string (id_elm->name);

      id2name_insert (&ic->uidtab,   id_elm, id_elm->id);
      name2id_insert (&ic->unametab, id_elm, hval);
    }

  if (!gname)
    {
      gr = getgrgid (sb->st_gid);
      if (remote && !res.status && res.u.names.gidname.present) {
	if (gr && !strcmp (gr->gr_name, res.u.names.gidname.u.name))
	  namecpy (snp->gidname, gr->gr_name);
	else {
	  snp->gidname[0] = '%';
	  namecpy (snp->gidname + 1, res.u.names.gidname.u.name);
	}
      }
      else if (!remote && gr)
	namecpy (snp->gidname, gr->gr_name);
      else
	sprintf (snp->gidname, "%u", (int) sb->st_gid);

      id_elm = malloc (sizeof (*id_elm));

      if (!id_elm)
	{
	  xdr_free ((xdrproc_t) xdr_sfsctl_getidnames_res, (char *) &res);
	  return -1;
	}

      bzero (id_elm->name, sizeof (id_elm->name));
      namecpy (id_elm->name, snp->gidname);
      id_elm->id = sb->st_gid;
      hval = hash_string (id_elm->name);

      id2name_insert (&ic->gidtab,   id_elm, id_elm->id);
      name2id_insert (&ic->gnametab, id_elm, hval);
    }

  xdr_free ((xdrproc_t) xdr_sfsctl_getidnames_res, (char *) &res);
  return 0;
}

int
sfs_getremoteid (sfs_remoteid *rip, dev_t dev)
{
  const char *fs;
  int cdfd;
  sfsctl_getcred_res res;
  unsigned int i;

  if (!devcon_lookup (&cdfd, &fs, dev)) {
    GETGROUPS_T gidset[sfs_maxgroups];
    rip->valid = SFS_UNIXCRED;
    rip->uid = geteuid ();
    rip->gid = getegid ();
    rip->ngroups = getgroups (sfs_maxgroups, gidset);
    if (rip->ngroups < 0)
      rip->ngroups = 0;
    for (i = 0; i < rip->ngroups; i++)
      rip->groups[i] = gidset[i];
    return 0;
  }

  bzero (rip, sizeof (*rip));
  bzero (&res, sizeof (res));
  if (srpc_call (&sfsctl_prog_1, cdfd, SFSCTL_GETCRED, &fs, &res)) {
    xdr_free ((xdrproc_t) xdr_sfsauth_cred, (char *) &res);
    return -1;
  }

  if (res.u.cred.type == SFS_UNIXCRED) {
    rip->valid = SFS_UNIXCRED;
    rip->uid = res.u.cred.u.unixcred.uid;
    rip->gid = res.u.cred.u.unixcred.gid;
    rip->ngroups = res.u.cred.u.unixcred.groups.len;
    if (rip->ngroups > sfs_maxgroups)
      rip->ngroups = sfs_maxgroups;
    for (i = 0; i < rip->ngroups; i++)
      rip->groups[i] = res.u.cred.u.unixcred.groups.val[i];
  }

  xdr_free ((xdrproc_t) xdr_sfsctl_getcred_res, (char *) &res);
  return 0;
}

static int
sfs_idbyname (const char *id, bool_t gid, dev_t dev)
{
  struct id_cache *ic_elm;

  ic_elm = lookup_ic (dev);
  
  if (!ic_elm)
    return -1;

  return lookup_id (ic_elm, gid, (char *)id);
}

int
sfs_uidbyname (const char *uidname, dev_t dev)
{
  return sfs_idbyname (uidname, FALSE, dev);
}

int
sfs_gidbyname (const char *gidname, dev_t dev)
{
  return sfs_idbyname (gidname, TRUE, dev);
}

