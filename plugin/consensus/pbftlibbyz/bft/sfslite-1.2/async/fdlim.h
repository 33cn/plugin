/* $Id: fdlim.h 2980 2007-08-11 04:08:55Z max $ */

/*
 *
 * Copyright (C) 1998 David Mazieres (dm@uun.org)
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


#ifndef FDLIM_H
#define FDLIM_H

#include <unistd.h>
#include <sys/types.h>
#include <sys/time.h>
#include <sys/resource.h>

#define FDLIM_MAX 0x18000

inline int
fdlim_get (int hard)
{
#ifdef RLIMIT_NOFILE
  struct rlimit rlfd;
  if (getrlimit (RLIMIT_NOFILE, &rlfd) < 0)
    return -1;
#ifdef RLIM_INFINITY /* not defined on HPSUX */
  if ((hard ? rlfd.rlim_max : rlfd.rlim_cur) == RLIM_INFINITY)
    return FDLIM_MAX;
  else
#endif /* RLIM_INFINITY */
    return hard ? rlfd.rlim_max : rlfd.rlim_cur;
#else /* !RLIMIT_NOFILE */
#ifdef HAVE_GETDTABLESIZE
  return getdtablesize ();
#else /* !HAVE_GETDTABLESIZE */
#ifdef _SC_OPEN_MAX
  return (sysconf (_SC_OPEN_MAX));
#else /* !_SC_OPEN_MAX */
#ifdef NOFILE
  return (NOFILE);
#else /* !NOFILE */
  return 25;
#endif /* !NOFILE */
#endif /* !_SC_OPEN_MAX */
#endif /* !HAVE_GETDTABLESIZE */
#endif /* !RLIMIT_NOFILE */
}

inline int
fdlim_set (rlim_t lim, int hard)
{
#ifdef RLIMIT_NOFILE
  struct rlimit rlfd;
  if (lim <= 0)
    return -1;
  if (getrlimit (RLIMIT_NOFILE, &rlfd) < 0)
    return -1;

#ifdef RLIMIT_INFINITY
  if (lim >= FDLIM_MAX)
    lim = RLIM_INFINITY;
#endif /* RLIMIT_INFINITY */
  switch (hard) {
  case 0:
    rlfd.rlim_cur = (lim <= rlfd.rlim_max) ? lim : rlfd.rlim_max;
    break;
  case 1:
    rlfd.rlim_cur = lim;
    if (lim > rlfd.rlim_max)
      rlfd.rlim_max = lim;
    break;
  case -1:
    rlfd.rlim_max = lim;
    if (rlfd.rlim_cur > lim)
      rlfd.rlim_cur = lim;
    break;
  default:
#ifdef assert
    assert (0);
#else /* !assert */
    return -1;
#endif /* !assert */
  }

  if (setrlimit (RLIMIT_NOFILE, &rlfd) < 0)
    return -1;
  return 0;
#else /* !RLIMIT_NOFILE */
  return -1;
#endif /* !RLIMIT_NOFILE */
}

#endif /* FDLIM_H */
