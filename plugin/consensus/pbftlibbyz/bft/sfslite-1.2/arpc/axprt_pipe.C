/* $Id: axprt_pipe.C 3709 2008-10-07 21:29:50Z max $ */

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

inline void
axprt_pipe::wrsync ()
{
  u_int64_t iovno = out->iovno () + out->iovcnt ();
  if (!syncpts.empty () && syncpts.back () == iovno)
    return;
  syncpts.push_back (iovno);
  out->breakiov ();
}

axprt_pipe::axprt_pipe (int rfd, int wfd, size_t ps, size_t bs)
  : axprt (true, true), destroyed (false), ingetpkt (false), pktsize (ps),
    bufsize (bs ? bs : pktsize + 4), fdread (rfd), fdwrite (wfd), cb (NULL),
    pktlen (0), wcbset (false), raw_bytes_sent (0)
{
  make_async (fdread);
  make_async (fdwrite);
  close_on_exec (fdread);
  close_on_exec (fdwrite);
  out = New suio;
  pktbuf = NULL;
  bytes_sent = bytes_recv = 0;

#if defined (SO_SNDBUF)
  socklen_t sn = sizeof (sndbufsz);
  if (getsockopt (fdwrite, SOL_SOCKET, SO_SNDBUF, (char *) &sndbufsz, &sn))
    sndbufsz = -1;
#else /* ! defined (SO_SNDBUF) */
  sndbufsz = -1;
#endif /* ! defined (SO_SNDBUF) */
}

axprt_pipe::~axprt_pipe ()
{
  destroyed = true;
  if (fdwrite >= 0 && out->resid ())
    output ();
  fail ();
  delete out;
  xfree (pktbuf);
}

void
axprt_pipe::setrcb (recvcb_t c)
{
  assert (!destroyed);
  cb = c;
  if (fdread >= 0) {
    if (cb) {
      fdcb (fdread, selread, wrap (this, &axprt_pipe::input));
      if (pktlen)
	callgetpkt ();
    }
    else
      fdcb (fdread, selread, NULL);
  }
  else if (cb)
    (*cb) (NULL, -1, NULL);
}

void
axprt_pipe::setwcb (cbv c)
{
  assert (!destroyed);
  if (out->resid ())
    out->iovcb (c);
  else
    (*c) ();
}

void
axprt_pipe::recvbreak ()
{
  warn ("axprt_pipe::recvbreak: unanticipated break\n");
  fail ();
}

void
axprt_pipe::sendbreak (cbv::ptr cb)
{
  static const u_int32_t zero[2] = {};
  suio_print (out, zero + 1, 4);
  if (cb)
    out->iovcb (cb);
  wrsync ();
  output ();
}

void
axprt_pipe::fail ()
{
  if (fdread >= 0) {
    fdcb (fdread, selread, NULL);
    close (fdread);
  }
  if (fdwrite >= 0) {
    fdcb (fdwrite, selwrite, NULL);
    wcbset = false;
    close (fdwrite);
  }
  fdread = fdwrite = -1;
  if (!destroyed) {
    ref<axprt> hold (mkref (this)); // Don't let this be freed under us
    if (cb && !ingetpkt)
      (*cb) (NULL, -1, NULL);
    out->clear ();
  }
}

void
axprt_pipe::reclaim (int *rfd, int *wfd)
{
  if (fdread >= 0) {
    fdcb (fdread, selread, NULL);
  }
  if (fdwrite >= 0) {
    fdcb (fdwrite, selwrite, NULL);
    wcbset = false;
  }
  *rfd = fdread;
  *wfd = fdwrite;
  fdread = fdwrite = -1;
  fail ();
}

void
axprt_pipe::_sockcheck(int fd)
{
  if (fd < 0)
    return;
  sockaddr_in sin;
  bzero (&sin, sizeof (sin));
  socklen_t sinlen = sizeof (sin);
  if (getsockname (fd, (sockaddr *) &sin, &sinlen) < 0)
    return;
  if (sin.sin_family == AF_INET) {
    for (in_addr *ap = ifchg_addrs.base (); ap < ifchg_addrs.lim (); ap++)
      if (*ap == sin.sin_addr)
	return;
    fail ();
  }
}

void
axprt_pipe::sockcheck ()
{
  this->_sockcheck(fdread);
  this->_sockcheck(fdwrite);
}

void
axprt_pipe::sendv (const iovec *iov, int cnt, const sockaddr *)
{
  assert (!destroyed);
  u_int32_t len = iovsize (iov, cnt);

  if (fdwrite < 0)
    panic ("axprt_pipe::sendv: called after an EOF\n");

  if (len > pktsize) {
    warn ("axprt_pipe::sendv: packet too large\n");
    fail ();
    return;
  }
  bytes_sent += len;
  raw_bytes_sent += len + 4;
  len = htonl (0x80000000 | len);

  if (!out->resid () && cnt < min (16, UIO_MAXIOV)) {
    iovec *niov = New iovec[cnt+1];
    niov[0].iov_base = (iovbase_t) &len;
    niov[0].iov_len = 4;
    memcpy (niov + 1, iov, cnt * sizeof (iovec));

    ssize_t skip = writev (fdwrite, niov, cnt + 1);
    if (skip < 0 && errno != EAGAIN) {
      fail ();
      return;
    }
    else
      out->copyv (niov, cnt + 1, max<ssize_t> (skip, 0));

    delete[] niov;
  }
  else {
    out->copy (&len, 4);
    out->copyv (iov, cnt, 0);
  }
  output ();
}

void
axprt_pipe::output ()
{
  ssize_t n;
  int cnt;

  do {
    while (!syncpts.empty () && out->iovno () >= syncpts.front ())
      syncpts.pop_front ();
    cnt = syncpts.empty () ? (size_t) -1
      : int (syncpts.front () - out->iovno ());
  } while ((n = dowritev (cnt)) > 0);

  if (n < 0)
    fail ();
  else if (out->resid () && !wcbset) {
    wcbset = true;
    fdcb (fdwrite, selwrite, wrap (this, &axprt_pipe::output));
  }
  else if (!out->resid () && wcbset) {
    wcbset = false;
    fdcb (fdwrite, selwrite, NULL);
  }
}

void
axprt_pipe::ungetpkt (const void *pkt, size_t len)
{
  assert (len <= pktsize);
  assert (!pktlen);

  if (!pktbuf) {
    pktbuf = (char *) xmalloc (bufsize);
  }

  pktlen = len + 4;
  putint (pktbuf, 0x80000000|len);
  memcpy (pktbuf + 4, pkt, len);
  if (cb)
    callgetpkt ();
}

bool
axprt_pipe::checklen (int32_t *lenp)
{
  int32_t len = *lenp;
  if (!(len & 0x80000000)) {
    warn ("axprt_pipe::checklen: invalid packet length: 0x%x\n", len);
    fail ();
    return false;
  }
  len &= 0x7fffffff;
  if ((u_int32_t) len > pktsize) {
    warn ("axprt_pipe::checklen: 0x%x byte packet is too large\n", len);
    fail ();
    return false;
  }
  *lenp = len;
  return true;
}

bool
axprt_pipe::getpkt (char **cpp, char *eom)
{
  char *cp = *cpp;
  if (!cb || eom - cp < 4)
    return false;

  int32_t len = getint (cp);
  cp += 4;

  if (!len) {
    *cpp = cp;
    recvbreak ();
    return true;
  }
  if (!checklen (&len))
    return false;

  if ((eom - cp) < len)
    return false;
  *cpp = cp + len;
  (*cb) (cp, len, NULL);
  return true;
}

ssize_t
axprt_pipe::doread (void *buf, size_t maxlen)
{
  return read (fdread, buf, maxlen);
}

void
axprt_pipe::input ()
{
  if (fdread < 0)
    return;

  ref<axprt> hold (mkref (this)); // Don't let this be freed under us

  if (!pktbuf) {
    pktbuf = (char *) xmalloc (bufsize);
  }

  ssize_t n = doread (pktbuf + pktlen, bufsize - pktlen);
  if (n <= 0) {
    if (n == 0 || errno != EAGAIN)
      fail ();
    return;
  }
  bytes_recv += n;
  pktlen += n;

  callgetpkt ();
}

void
axprt_pipe::poll ()
{
  assert (cb);
  assert (!ateof ());
  if (ingetpkt)
    panic ("axprt_pipe: polling for more input from within a callback\n");

  struct timeval ztv = { 0, 0 };
  fdwait (fdread, fdwrite, true, wcbset, NULL);
  if (!wcbset || fdwait (fdread, selread, &ztv) > 0)
    input ();
  else
    output ();
}

void
axprt_pipe::callgetpkt ()
{
  if (ingetpkt)
    return;

  ref<axprt> hold (mkref (this)); // Don't let this be freed under us

  ingetpkt = true;
  char *cp = pktbuf, *eom = pktbuf + pktlen;
  while (cb && getpkt (&cp, eom))
    ;
  if (ateof ()) {
    if (cb)
      (*cb) (NULL, -1, NULL);
  }
  else {
    if (cp != pktbuf)
      memmove (pktbuf, cp, eom - cp);
    pktlen -= cp - pktbuf;
    if (!pktlen) {
      xfree (pktbuf);
      pktbuf = NULL;
    }
    assert (pktlen < pktsize);
  }
  ingetpkt = false;
}
