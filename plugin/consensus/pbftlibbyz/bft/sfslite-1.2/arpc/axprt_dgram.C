/* $Id: axprt_dgram.C 1693 2006-04-28 23:17:35Z max $ */

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

#include "arpc.h"

bool
axprt_dgram::isconnected (int fd)
{
  sockaddr_in sin;
  bzero (&sin, sizeof (sin));
  sin.sin_family = AF_INET;
  socklen_t len = sizeof (sin);
  return !getpeername (fd, (sockaddr *) &sin, &len);
}

axprt_dgram::axprt_dgram (int f, bool c, size_t ss, size_t ps)
  : axprt (false, c, c ? 0 : ss), pktsize (ps), fd (f), cb (NULL)
{
  make_async (fd);
  close_on_exec (fd);

#ifdef SO_RCVBUF
  int n = 0;
  socklen_t sn = sizeof (n);
  if (getsockopt (fd, SOL_SOCKET, SO_RCVBUF, &n, &sn) >= 0
      && pktsize > implicit_cast<size_t> (n)) {
    n = pktsize;
    if (setsockopt (fd, SOL_SOCKET, SO_RCVBUF, &n, sizeof (n)) < 0)
      warn ("SO_RCVBUF -> %d bytes: %m\n", n);
  }
#endif /* SO_RCVBUF */

  if (c)
    sabuf = NULL;
  else
    sabuf = (sockaddr *) xmalloc (socksize);
  pktbuf = (char *) xmalloc (pktsize);
}

axprt_dgram::~axprt_dgram ()
{
  fdcb (fd, selread, NULL);
  close (fd);
  xfree (sabuf);
  xfree (pktbuf);
}

void
axprt_dgram::sendv (const iovec *iov, int cnt, const sockaddr *sap)
{
  assert (connected == !sap);
  msghdr mh;
  bzero (&mh, sizeof (mh));
  if (connected)
    mh.msg_name = NULL;
  else
    mh.msg_name = (char *) sap;
  mh.msg_namelen = socksize;
  mh.msg_iov = const_cast<iovec *> (iov);
  mh.msg_iovlen = cnt;
  sendmsg (fd, &mh, 0);
}

void
axprt_dgram::setrcb (recvcb_t c)
{
  cb = c;
  fdcb (fd, selread, c ? wrap (this, &axprt_dgram::input) : NULL);
}

void
axprt_dgram::input ()
{
  ref<axprt> hold (mkref (this)); // Don't let this be freed under us
  for (size_t tot = 0; cb && tot < pktsize;) {
    socklen_t ss = socksize;
    bzero (sabuf, ss);
    ssize_t ps = recvfrom (fd, pktbuf, pktsize, 0, sabuf, &ss);
    if (ps < 0) {
      if (errno != EAGAIN && connected)
	(*cb) (NULL, -1, NULL);
      return;
    }
    tot += ps;
    (*cb) (pktbuf, ps, sabuf);
  }
}

void
axprt_dgram::poll ()
{
  assert (cb);

  make_sync (fd);

  socklen_t ss = socksize;
  bzero (sabuf, ss);
  ssize_t ps = recvfrom (fd, pktbuf, pktsize, 0, sabuf, &ss);

  make_async (fd);

  if (ps < 0) {
    if (errno != EAGAIN && connected)
      (*cb) (NULL, -1, NULL);
    return;
  }
  (*cb) (pktbuf, ps, sabuf);
}
