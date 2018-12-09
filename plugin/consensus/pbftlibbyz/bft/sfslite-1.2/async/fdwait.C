/* $Id: fdwait.C 3758 2008-11-13 00:36:00Z max $ */

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

#include "amisc.h"

int
fdwait (int fd, bool r, bool w, timeval *tvp)
{
  return fdwait(fd, fd, r, w, tvp);
}

int
fdwait (int rfd, int wfd, bool r, bool w, timeval *tvp)
{
  static int nfd;
  static fd_set *rfds;
  static fd_set *wfds;
  int maxfd = rfd > wfd ? rfd : wfd;

  assert (rfd >= 0 && wfd >=0);

  if (maxfd >= nfd) {
    // nfd = max (fd + 1, FD_SETSIZE);
    nfd = maxfd + 1;
    nfd = (nfd + 0x3f) & ~0x3f;
    xfree (rfds);
    xfree (wfds);
    rfds = (fd_set *) xmalloc (nfd >> 3);
    wfds = (fd_set *) xmalloc (nfd >> 3);
    bzero (rfds, nfd >> 3);
    bzero (wfds, nfd >> 3);
  }

  FD_SET (rfd, rfds);
  FD_SET (wfd, wfds);
  int res = select (maxfd + 1, r ? rfds : NULL, w ? wfds : NULL, NULL, tvp);
  FD_CLR (rfd, rfds);
  FD_CLR (wfd, wfds);
  return res;
}

int
fdwait (int fd, selop op, timeval *tvp)
{
  switch (op) {
  case selread:
    return fdwait (fd, true, false, tvp);
  case selwrite:
    return fdwait (fd, false, true, tvp);
  default:
    panic ("fdwait: bad operation\n");
  }
}

