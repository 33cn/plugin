/* $Id: devcon.c 2507 2007-01-12 20:28:54Z max $ */

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
#include "hashtab.h"
#include "sfsagent.h"

const char *path_suidconnect;
const char *path_devdb;

static pid_t sfspid;
static struct timespec devdb_mtime;

struct dev_prog {
  u_int64_t dev;
  char *prog;
  char *fs;
  struct hashtab_entry link;
};
hashtab_decl (dev2prog, dev_prog, link) devtab;

struct prog_fd {
  char *prog;
  int fd;
  struct hashtab_entry link;
};
hashtab_decl (prog2fd, prog_fd, link) progtab;

static void
set_paths (void)
{
#ifdef MAINTAINER
  const char *builddir = getenv ("SFS_RUNINPLACE");
  const char *rootdir = getenv ("SFS_ROOT");
#endif /* MAINTAINER */

  if (!path_suidconnect) {
#ifdef MAINTAINER
    static const char scpath[] = "/libsfs/suidconnect";
    if (builddir) {
      char *buf = malloc (strlen (builddir) + sizeof (scpath));
      assert (buf);
      strcpy (buf, builddir);
      strcat (buf, scpath);
      path_suidconnect = buf;
    }
    else
#endif /* MAINTAINER */
      path_suidconnect = EXECDIR "/suidconnect";
  }
  if (!path_devdb) {
#ifdef MAINTAINER
    static const char ddbpath[] = "/.devdb";
    if (rootdir) {
      char *buf = malloc (strlen (rootdir) + sizeof (ddbpath));
      assert (buf);
      strcpy (buf, rootdir);
      strcat (buf, ddbpath);
      path_devdb = buf;
    }
    else
#endif /* MAINTAINER */
      path_devdb = "/sfs/.devdb";
  }
}

static int
recvfd (int fd)
{
  int nfd = -1;
  char c;
  readfd (fd, &c, 1, &nfd);
  return nfd;
}

static int
suidgetfd (const char *prog)
{
  int fds[2];
  int nfd;
  char *av[3] = { "suidconnect", (char *) prog, NULL };

  if (!path_suidconnect)
    set_paths ();

  if (socketpair (AF_UNIX, SOCK_STREAM, 0, fds) < 0) {
    perror ("socketpair");
    return -1;
  }
  switch (fork ()) {
  case -1:
    perror ("fork");
    close (fds[0]);
    close (fds[1]);
    return -1;
  case 0:
    close (fds[1]);
    if (fds[0]) {
      dup2 (fds[0], 0);
      close (fds[0]);
    }
    execv (path_suidconnect, av);
    perror (path_suidconnect);	/* XXX */
    exit (1);
  }

  close (fds[0]);
  nfd = recvfd (fds[1]);
  close (fds[1]);
  return nfd;
}

static int
getprogfd (const char *prog)
{
  int fd;
  int pid;
  int flags;
  enum clnt_stat err;

  fd = suidgetfd (prog);
  if (fd < 0)
    return -1;

  flags = fcntl (fd, F_GETFL, 0);
  if (flags == -1) {
    close (fd);
    return -1;
  }
  flags &= ~O_NONBLOCK;
  fcntl (fd, F_SETFL, flags);

  pid = getpid ();
  err = srpc_call (&sfsctl_prog_1, fd, SFSCTL_SETPID, &pid, NULL);
  if (err) {
    clnt_perrno (err);		/* XXX */
    close (fd);
    return -1;
  }

  return fd;
}

static void
dev2str (char *out, u_int64_t dev)
{
  if (dev >= INT64 (0x100000000))
    sprintf (out, "0x%x%08x", (int) (dev >> 32), (int) (dev & 0xffffffffU));
  else
    sprintf (out, "0x%x", (int) (dev & 0xffffffffU));
}

static bool_t
devdblookup (char **progp, char **fsp, u_int64_t dev)
{
  char linebuf[512];
  char devstr[20];
  size_t devlen;
  FILE *ddbf;

  if (!path_devdb)
    set_paths ();
  if (progp)
    *progp = NULL;
  if (fsp)
    *fsp = NULL;
  dev2str (devstr, dev);
  strcat (devstr, " ");
  devlen = strlen (devstr);

  ddbf = fopen (path_devdb, "r");
  if (!ddbf)
    return FALSE;

  while (fgets (linebuf, sizeof (linebuf), ddbf))
    if (!strncmp (linebuf, devstr, devlen)) {
      char *sdev, *sfs, *sprog;
      char *bufp = linebuf;
      if ((sdev = strnnsep_c (&bufp, " \r\n"))
	  && (sfs = strnnsep_c (&bufp, " \r\n")) && strlen (sfs) <= 255
	  && (sprog = strnnsep_c (&bufp, " \r\n")) && strlen (sprog) <= 255) {
	if (fsp) {
	  *fsp = malloc (strlen (sfs) + 1);
	  if (!*fsp) {
	    fclose (ddbf);
	    return FALSE;
	  }
	  strcpy (*fsp, sfs);
	}
	if (progp) {
	  *progp = malloc (strlen (sprog) + 1);
	  if (!*progp) {
	    if (fsp) {
	      free (*fsp);
	      *fsp = NULL;
	    }
	    fclose (ddbf);
	    return FALSE;
	  }
	  strcpy (*progp, sprog);
	}
	fclose (ddbf);
	return TRUE;
      }
    }

  fclose (ddbf);
  return FALSE;
}

static void
deldpp (void *_null, struct dev_prog *dpp)
{
  dev2prog_delete (&devtab, dpp);
  free (dpp->prog);
  free (dpp->fs);
  free (dpp);
}

static struct dev_prog *
getdpp (u_int64_t dev)
{
  struct dev_prog *dpp;
  for (dpp = dev2prog_chain (&devtab, dev); dpp && dpp->dev != dev;
       dpp = dev2prog_next (dpp))
    ;
  if (dpp)
    return dpp;

  dpp = malloc (sizeof (*dpp));
  if (!dpp)
    return NULL;
  dpp->dev = dev;
  devdblookup (&dpp->prog, &dpp->fs, dev);
  dev2prog_insert (&devtab, dpp, dev);
  return dpp;
}

static void
delpfp (void *_null, struct prog_fd *pfp)
{
  prog2fd_delete (&progtab, pfp);
  free (pfp->prog);
  if (pfp->fd >= 0)
    close (pfp->fd);
  free (pfp);
}

static struct prog_fd *
getpfp (const char *prog)
{
  u_int hval = hash_string (prog);
  struct prog_fd *pfp;

  for (pfp = prog2fd_chain (&progtab, hval); pfp && strcmp (prog, pfp->prog);
       pfp = prog2fd_next (pfp))
    ;
  if (pfp)
    return pfp;

  pfp = malloc (sizeof (*pfp));
  if (!pfp)
    return NULL;
  pfp->fd = getprogfd (prog);
  pfp->prog = malloc (strlen (prog) + 1);
  if (!pfp->prog) {
    free (pfp);
    return NULL;
  }
  strcpy (pfp->prog, prog);
  prog2fd_insert (&progtab, pfp, hval);
  return pfp;
}

void
devcon_flush (void)
{
  dev2prog_traverse (&devtab, deldpp, NULL);
  sfs_flush_idcache ();
}

void
devcon_close (void)
{
  dev2prog_traverse (&devtab, deldpp, NULL);
  prog2fd_traverse (&progtab, delpfp, NULL);
  sfs_flush_idcache ();
}

static void
devcon_ckfresh (void)
{
  pid_t mypid;
  struct stat sb;

  mypid = getpid ();
  if (sfspid != mypid) {
    sfspid = mypid;
    prog2fd_traverse (&progtab, delpfp, NULL);
  }

  if (!path_devdb)
    set_paths ();
  if (stat (path_devdb, &sb) >= 0) {
    struct timespec mtime;
#ifdef SFS_HAVE_STAT_ST_MTIMESPEC
    mtime = sb.st_mtimespec;
#else /* !SFS_HAVE_STAT_ST_MTIMESPEC */
    mtime.tv_sec = sb.st_mtime;
    mtime.tv_nsec = 0;
#endif /* !SFS_HAVE_STAT_ST_MTIMESPEC */
    if (mtime.tv_sec != devdb_mtime.tv_sec
	|| mtime.tv_nsec != devdb_mtime.tv_nsec) {
      devdb_mtime = mtime;
      dev2prog_traverse (&devtab, deldpp, NULL);
    }
  }
}

bool_t
devcon_lookup (int *fdp, const char **fsp, dev_t dev)
{
  struct dev_prog *dpp;
  struct prog_fd *pfp;

  devcon_ckfresh ();

  *fdp = -1;
  *fsp = NULL;
  dpp = getdpp (dev2int (dev));
  if (!dpp || !dpp->prog || !dpp->fs)
    return FALSE;
  *fsp = dpp->fs;
  pfp = getpfp (dpp->prog);
  if (!pfp || pfp->fd < 0)
    return FALSE;
  *fdp = pfp->fd;
  return TRUE;
}
