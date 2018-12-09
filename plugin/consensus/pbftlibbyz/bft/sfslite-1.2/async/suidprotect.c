/* $Id: suidprotect.c 1117 2005-11-01 16:20:39Z max $ */

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

int
suidsafe (void)
{
  static int safe;
  if (!safe) {
    if (suidprotect && getuid ()
#if defined (HAVE_ISSETUGID)
	     && issetugid ()
#elif defined (HAVE_GETEUID) && defined (HAVE_GETEGID)
	     && (getuid () != geteuid () || getgid () != getegid ())
#endif /* HAVE_GETEUID and HAVE_GETEGID */
	     )
      safe = -1;
    else
      safe = 1;
  }
  return safe < 0 ? 0 : 1;
}

int
execsafe (void)
{
  return (suidprotect || execprotect) ? 0 : 1;
}

char *
safegetenv (const char *name)
{
  return suidsafe () ? getenv (name) : NULL;
}

extern char **environ;

#ifndef PUTENV_COPIES_ARGUMENT
#if defined (__sun__) && defined (__svr4__)
# define BAD_PUTENV 1
#endif /* __sun__ && __svr4__ */

#if BAD_PUTENV
static char **newenvp;
static char **badenvp;
#endif /* BAD_PUTENV */

int
xputenv (const char *s)
{
  char *ss;
  int ret;

#if BAD_PUTENV
  /* Solaris appears to have a very broken putenv implementation that
   * calls realloc on memory not allocated by malloc.  The effects are
   * particularly bad when C++ global constructors call putenv before
   * that start of main.  (Perhaps some gcc initialization code fixes
   * the problem before main is invoked?)  Just to be safe, we make
   * sure environ really points to malloced memory.  */
  if (!newenvp) {
    int n;
    for (n = 0; environ[n]; n++)
      ;
    newenvp = malloc ((n + 2) * sizeof (*newenvp));
    if (!newenvp)
      return -1;
    memcpy (newenvp, environ, (n + 1) * sizeof (environ));
    badenvp = environ;
  }
  /* Somehow, Solaris stubbornly resets environ to point to crap
   * memory, so we need to keep resetting it. */
  if (environ == badenvp)
    environ = newenvp;
#endif /* BAD_PUTENV */

  /* Note:  Don't use xstrdup, because we don't always link this file
   * against libasync.  (suidconnect uses it, too) */
  ss = strdup (s);
  if (!ss)
    return -1;
  ret = putenv (ss);
  if (ret < 0)
    xfree (ss);
#if BAD_PUTENV
  else
    newenvp = environ;
#endif /* BAD_PUTENV */
  return ret;
}
#endif /* !PUTENV_COPIES_ARGUMENT */

#ifndef HAVE_UNSETENV
void
unsetenv (const char *name)
{
  int len = strlen (name);
  char **ep;
  for (ep = environ;; ep++) {
    if (!*ep)
      return;
    else if (!strncmp (name, *ep, len) && (*ep)[len] == '=')
      break;
  }
  while ((ep[0] = ep[1]))
    ep++;
}
#endif /* !HAVE_UNSETENV */
