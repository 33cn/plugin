/* $Id: sfspaths.c 2507 2007-01-12 20:28:54Z max $ */

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


#include "sfs-internal.h"

/* This is a strsep function that returns a null field for adjacent
 * separators.  This is the same as the 4.4BSD strsep, but different
 * from the one in the GNU libc. */
char *
xstrsep_c(char **str, const char *delim)
{
  char *s, *e;

  if (!**str)
    return (NULL);

  s = *str;
  e = s + strcspn (s, delim);

  if (*e != '\0')
    *e++ = '\0';
  *str = e;

  return (s);
}

/* Get the next non-null token (like GNU strsep).  Strsep() will
 * return a null token for two adjacent separators, so we may have to
 * loop. */
char *
strnnsep_c (char **stringp, const char *delim)
{
  char *tok;

  do {
    tok = xstrsep_c (stringp, delim);
  } while (tok && *tok == '\0');
  return (tok);
}

const char *
getsfssockdir_c (void)
{
  static char *sockdir;
  char *runinplace;
  char *paths[] = { ETCDIR "/sfs_config", DATADIR "/sfs_config", NULL };
  char **p;
  FILE *f = NULL;

  if (sockdir)
    return sockdir;

#ifdef MAINTAINER
  if ((runinplace = getenv ("SFS_RUNINPLACE"))) {
    char suff[] = "/runinplace";
    if (!suidsafe ()) {
      setgid (getgid ());
      setuid (getuid ());
    }
    sockdir = malloc (strlen (runinplace) + sizeof (suff));
    if (!sockdir) {
      fprintf (stderr, "out of memory\n");
      exit (1);
    }
    sprintf (sockdir, "%s%s", runinplace, suff);
    return sockdir;
  }
#endif /* MAINTAINER */

  for (p = paths; *p && !(f = fopen (*p, "r")); p++)
    ;
  if (f) {
    char *s, *t, buf[2048];
    while ((s = fgets (buf, sizeof (buf), f))) {
      if (strlen (buf) >= sizeof (buf) - 1)
	break;
      t = strnnsep_c (&s, " \r\n");
      if (strcasecmp (t, "sfsdir"))
	continue;
      t = strnnsep_c (&s, " \r\n");
      if (t && *t == '/') {
	sockdir = malloc (strlen (t) + sizeof ("/sockets"));
	if (!sockdir) {
	  fprintf (stderr, "out of memory\n");
	  exit (1);
	}
	sprintf (sockdir, "%s/sockets", t);
	break;
      }
    }
    fclose (f);
  }

  if (!sockdir)
    sockdir = SFSDIR "/sockets";
  return sockdir;
}
