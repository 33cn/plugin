/* $Id: pathinfo.c 3769 2008-11-13 20:21:34Z max $ */

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

#include "sysconf.h"
#include "rwfd.h"

int suidprotect = 1;

char *progname;

static void fperror (char *msg) __attribute__ ((noreturn));
static void
fperror (char *msg)
{
  perror (msg);
  exit (1);
}

static int
isunixsocket (int fd)
{
  struct sockaddr_un su;
  socklen_t sulen = sizeof (su);
  bzero (&su, sizeof (su));
  su.sun_family = AF_UNIX;
  if (getsockname (fd, (struct sockaddr *) &su, &sulen) < 0
      || su.sun_family != AF_UNIX)
    return -1;
  return 0;
}

static void
pathinfo (char *path)
{
  int fd;
  struct stat sb;
  dev_t dev;
  char cwd[PATH_MAX+1];
  char buf[2*PATH_MAX+3];
  char res[3*PATH_MAX+40];
  char *rp;
  FILE *dfpipe;

  if (chdir (path) < 0
      || (fd = open (".", O_RDONLY)) < 0
      || fstat (fd, &sb) < 0)
    fperror (path);
  dev = sb.st_dev;

  if (!getcwd (cwd, sizeof (cwd)))
    fperror ("getcwd");

  strcpy (buf, cwd);
  while ((rp = strrchr (buf, '/'))) {
    rp[1] = '\0';
    if (stat (buf, &sb) < 0)
      fperror (buf);
    if (sb.st_dev != dev) {
      if (!(rp = strchr (cwd + (rp - buf) + 1, '/')))
	rp = "/";
      break;
    }
    rp[0] = '\0';
  }
  if (!rp)
    rp = cwd;

  dfpipe = popen (PATH_DF
#ifdef DF_NEEDS_DASH_K
		  " -k"
#endif /* DF_NEEDS_DASH_K */
		  " . | sed -ne '2s/ .*//; 2p'", "r");
  if (!dfpipe || !fgets (buf, sizeof (buf) - 1, dfpipe)
      || pclose (dfpipe) < 0)
    fperror (PATH_DF);
  if (!strchr (buf, '\n'))
    strcat (buf, "\n");

  sprintf (res, "0x%" U64F "x\n%s%s\n", (u_int64_t) dev, buf, rp);
  if (isunixsocket (1) < 0) {
    int i = write (1, res, strlen (res));
    i++;
  } else
    writefd (1, res, strlen (res), fd);
}

static void
usage (void)
{
  fprintf (stderr, "usage: %s [-u uid] path\n", progname);
  exit (1);
}

int
main (int argc, char **argv)
{
  int ch;

  if ((progname = strrchr (argv[0], '/')))
    progname++;
  else
    progname = argv[0];

  while ((ch = getopt (argc, argv, "u:")) != -1)
    switch (ch) {
    case 'u':
      if (setuid (atoi (optarg)) < 0)
	fperror ("setuid");
      break;
    default:
      usage ();
    }
  argc -= optind;
  argv += optind;

  if (argc != 1)
    usage ();

  pathinfo (argv[0]);
  exit (0);
}
