/* $Id: clock_gettime.c 1117 2005-11-01 16:20:39Z max $ */

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


#include "sysconf.h"

#ifndef HAVE_CLOCK_GETTIME

#include <sys/resource.h>

int
clock_gettime (int id, struct timespec *tp)
{
  struct timeval tv;
  struct rusage ru;

  switch (id) {
  case CLOCK_REALTIME:
    if (gettimeofday (&tv, NULL) < 0)
      return -1;
    tp->tv_sec = tv.tv_sec;
    tp->tv_nsec = tv.tv_usec * 1000;
    return 0;
  case CLOCK_VIRTUAL:
    if (getrusage (RUSAGE_SELF, &ru) < 0)
      return -1;
    tp->tv_sec = ru.ru_utime.tv_sec;
    tp->tv_nsec = ru.ru_utime.tv_usec * 1000;
    return 0;
  case CLOCK_PROF:
    if (getrusage (RUSAGE_SELF, &ru) < 0)
      return -1;
    tp->tv_sec = (ru.ru_utime.tv_sec + ru.ru_stime.tv_sec);
    tp->tv_nsec = (ru.ru_utime.tv_usec + ru.ru_stime.tv_usec) * 1000;
    while (tp->tv_nsec > 1000000000) {
      tp->tv_sec++;
      tp->tv_nsec -= 1000000000;
    }
    return 0;
  default:
    errno = EINVAL;
    return -1;
  }
}

#endif /* !HAVE_CLOCK_GETTIME */
