/* $Id: suidconnect.c 2507 2007-01-12 20:28:54Z max $ */

/*
 *
 * Copyright (C) 1998, 1999 David Mazieres (dm@uun.org)
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

int suidprotect = 1;

int opt_quiet;

static int
sendfd (int fd, int wfd)
{
  char c = '\175';
  return writefd (fd, &c, 1, wfd);
}

static void
usage (void)
{
  fprintf (stderr, "usage: suidconnect [-q] program\n");
  exit (1);
}

static bool_t
checkprog (const char *prog)
{
  if (!*prog)
    return FALSE;
  for (; *prog; prog++)
    if (!isalnum (*prog) && *prog != '_')
      return FALSE;
  return TRUE;
}

static bool_t
getaddr (struct sockaddr_un *sun, const char *prog)
{
  static const char *runinplace = "";
  static const char suffix[] = ".sock";
  const char *sockdir;
#ifdef DMALLOC
  if (!suidsafe ()) {
    xputenv ("DMALLOC_LOGFILE=/dev/null");
    xputenv ("DMALLOC_OPTIONS=");
    dmalloc_debug (0);
  }
#endif /* DMALLOC */
  sockdir = getsfssockdir_c ();

  if (!checkprog (prog)
      || (strlen (sockdir) + 1 + strlen (runinplace)
	  + strlen (prog) + sizeof (suffix) > sizeof (sun->sun_path)))
    return FALSE;
  strcpy (sun->sun_path, sockdir);
  strcat (sun->sun_path, runinplace);
  strcat (sun->sun_path, "/");
  strcat (sun->sun_path, prog);
  strcat (sun->sun_path, suffix);
  sun->sun_family = AF_UNIX;
  return TRUE;
}

int
main (int argc, char **argv)
{
  struct sockaddr_un sun;
  int fd;
  AUTH *auth;
  int res;
  enum clnt_stat err;
  int argn;

  for (argn = 1; argn < argc && argv[argn][0] == '-'; argn++) {
    if (!strcmp (argv[argn], "-q"))
      opt_quiet = 1;
    else
      usage ();
  }

  if (argn + 1 != argc || !getaddr (&sun, argv[argn]))
    usage ();

  if ((fd = socket (AF_UNIX, SOCK_STREAM, 0)) < 0
      || connect (fd, (struct sockaddr *) &sun, sizeof (sun)) < 0) {
    if (!opt_quiet)
      perror (sun.sun_path);
    exit (1);
  }
  
  auth = authunix_create_realids ();
  if (!auth) {
    fprintf (stderr, "could not create auth_unix structure\n");
    exit (1);
  }

  err = srpc_callraw (fd, SETUID_PROG, SETUID_VERS, SETUIDPROC_SETUID,
		      (xdrproc_t) xdr_void, NULL, (xdrproc_t) xdr_int, &res,
		      auth);
  if (err)
    fprintf (stderr, "%s: %s\n", argv[1], clnt_sperrno (err));
  else if (res)
    fprintf (stderr, "%s: %s\n", argv[1], strerror (res));
  else if (sendfd (0, fd) >= 0) {
#ifdef __APPLE__
    /* Work around garbage-collection bug in MacOS where passed socket
     * becomes invalid if we exit too soon. */
    if (fd < FD_SETSIZE) {
      fd_set fds;
      struct timeval tv;

      FD_ZERO (&fds);
      FD_SET (fd, &fds);
      tv.tv_sec = 60;
      tv.tv_usec = 0;
      select (fd + 1, &fds, NULL, NULL, &tv);
    }
#endif /* __APPLE__ */
    exit (0);
  }
  else
    perror ("suidconnect: sendfd");
  exit (1);
}
