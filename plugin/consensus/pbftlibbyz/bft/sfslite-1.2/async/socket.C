/* $Id: socket.C 2967 2007-07-24 15:20:57Z max $ */

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
#include "init.h"
extern "C" {
#ifdef HAVE_BINDRESVPORT
# include <rpc/rpc.h>
# ifdef NEED_BINDRESVPORT_DECL
int bindresvport (int, struct sockaddr_in *);
# endif /* NEED_BINDRESVPORT_DECL */
#endif /* HAVE_BINDRESVPORT */
#include <netinet/in_systm.h>
#include <netinet/tcp.h>
#include <netinet/ip.h>
}

#ifdef SFS_ALLOW_LARGE_BUFFER
enum { maxsobufsize = 0x11000 }; /* 64K + header */
#else /* !SFS_ALLOW_LARGE_BUFFER */
enum { maxsobufsize = 0x05000 }; /* 16K + header */
#endif /* !SFS_ALLOW_LARGE_BUFFER */
int sndbufsize = maxsobufsize;
int rcvbufsize = maxsobufsize;
in_addr inet_bindaddr;
INITFN(init_env);
static void
init_env ()
{
  if (char *p = safegetenv ("SNDBUFSIZE"))
    sndbufsize = atoi (p);
  if (char *p = safegetenv ("RCVBUFSIZE"))
    rcvbufsize = atoi (p);

  char *p = safegetenv ("BINDADDR");
  if (!p || inet_aton (p, &inet_bindaddr) <= 0)
    inet_bindaddr.s_addr = htonl (INADDR_ANY);
}


u_int16_t inetsocket_lastport;

int
inetsocket_resvport (int type, u_int32_t addr)
{
#ifndef NORESVPORTS
  int s;
  struct sockaddr_in sin;

  bzero (&sin, sizeof (sin));
  sin.sin_family = AF_INET;
  sin.sin_port = htons (0);
  if (addr == INADDR_ANY)
    sin.sin_addr = inet_bindaddr;
  else
    sin.sin_addr.s_addr = htonl (addr);
  if ((s = socket (AF_INET, type, 0)) < 0)
    return (-1);

  /* Don't bother if we aren't root */
  if (geteuid ()) {
  again0:
    if (bind (s, (struct sockaddr *) &sin, sizeof (sin)) >= 0)
      return s;
    if (errno == EADDRNOTAVAIL && sin.sin_addr.s_addr != htonl (addr)) {
      sin.sin_addr.s_addr = htonl (addr);
      goto again0;
    }
    close (s);
    return -1;
  }
#ifdef HAVE_BINDRESVPORT
 again1:
  if (bindresvport (s, &sin) >= 0) {
    inetsocket_lastport = ntohs (sin.sin_port);
    return s;
  }
  if (errno == EADDRNOTAVAIL && sin.sin_addr.s_addr != htonl (addr)) {
    sin.sin_addr.s_addr = htonl (addr);
    goto again1;
  }
#else /* !HAVE_BINDRESVPORT */
  for (inetsocket_lastport = IPPORT_RESERVED - 1;
       inetsocket_lastport > IPPORT_RESERVED/2;
       inetsocket_lastport--) {
  again2:
    sin.sin_port = htons (inetsocket_lastport);
    if (bind (s, (struct sockaddr *) &sin, sizeof (sin)) >= 0)
      return (s);
    if (errno == EADDRNOTAVAIL && sin.sin_addr.s_addr != htonl (addr)) {
      sin.sin_addr.s_addr = htonl (addr);
      goto again2;
    }
    if (errno != EADDRINUSE)
      break;
  }
#endif /* !HAVE_BINDRESVPORT */
  close (s);
  return -1;
#else /* NORESVPORTS */
  return inetsocket (type, addr, 0);
#endif /* NORESVPORTS */
}

int
inetsocket (int type, u_int16_t port, u_int32_t addr)
{
  int s;
  int n;
  socklen_t sn;
  struct sockaddr_in sin;

  bzero (&sin, sizeof (sin));
  sin.sin_family = AF_INET;
  sin.sin_port = htons (port);
  if (addr == INADDR_ANY)
    sin.sin_addr = inet_bindaddr;
  else
    sin.sin_addr.s_addr = htonl (addr);
  if ((s = socket (AF_INET, type, 0)) < 0)
    return -1;

  sn = sizeof (n);
  n = 1;
  /* Avoid those annoying TIME_WAITs for TCP */
  if (port && type == SOCK_STREAM
      && setsockopt (s, SOL_SOCKET, SO_REUSEADDR, (char *) &n, sizeof (n)) < 0)
    fatal ("inetsocket: SO_REUSEADDR: %s\n", strerror (errno));
 again:
  if (bind (s, (struct sockaddr *) &sin, sizeof (sin)) >= 0) {
#if 0
    if (type == SOCK_STREAM)
      warn ("TCP (fd %d) bound: %s\n", s, __backtrace (__FL__));
#endif
    return s;
  }
  if (errno == EADDRNOTAVAIL && sin.sin_addr.s_addr != htonl (addr)) {
    sin.sin_addr.s_addr = htonl (addr);
    goto again;
  }
  close (s);
  return -1;
}

int
unixsocket (const char *path)
{
  sockaddr_un sun;

  if (strlen (path) >= sizeof (sun.sun_path)) {
#ifdef ENAMETOOLONG
    errno = ENAMETOOLONG;
#else /* !ENAMETOOLONG */
    errno = E2BIG;
#endif /* !ENAMETOOLONG */
    return -1;
  }

  bzero (&sun, sizeof (sun));
  sun.sun_family = AF_UNIX;
  strcpy (sun.sun_path, path);

  int fd = socket (AF_UNIX, SOCK_STREAM, 0);
  if (fd < 0)
    return -1;
  if (bind (fd, (sockaddr *) &sun, sizeof (sun)) < 0) {
    close (fd);
    return -1;
  }
  return fd;
}

int
unixsocket_connect (const char *path)
{
  sockaddr_un sun;

  if (strlen (path) >= sizeof (sun.sun_path)) {
#ifdef ENAMETOOLONG
    errno = ENAMETOOLONG;
#else /* !ENAMETOOLONG */
    errno = E2BIG;
#endif /* !ENAMETOOLONG */
    return -1;
  }

  bzero (&sun, sizeof (sun));
  sun.sun_family = AF_UNIX;
  strcpy (sun.sun_path, path);

  int fd = socket (AF_UNIX, SOCK_STREAM, 0);
  if (fd < 0)
    return -1;
  if (connect (fd, (sockaddr *) &sun, sizeof (sun)) < 0) {
    close (fd);
    return -1;
  }
  return fd;
}

bool
isunixsocket (int fd)
{
  sockaddr_un sun;
  socklen_t sunlen = sizeof (sun);
  bzero (&sun, sizeof (sun));
  sun.sun_family = AF_UNIX;
  if (getsockname (fd, (sockaddr *) &sun, &sunlen) < 0
      || sun.sun_family != AF_UNIX)
    return false;
  return true;
}

void
close_on_exec (int s, bool set)
{
  if (fcntl (s, F_SETFD, int (set)) < 0)
    fatal ("F_SETFD: %s\n", strerror (errno));
}

int
_make_async (int s)
{
  int n;
  if ((n = fcntl (s, F_GETFL)) < 0
      || fcntl (s, F_SETFL, n | O_NONBLOCK) < 0)
    return -1;
  return 0;
}

void
make_async (int s)
{
  int n;
  int type;
  socklen_t sn;
  if (_make_async (s) < 0)
    fatal ("O_NONBLOCK: %s\n", strerror (errno));
  type = 0;
  sn = sizeof (type);
  if (getsockopt (s, SOL_SOCKET, SO_TYPE, (char *)&type, &sn) < 0)
    return;

  // Linux does its own autotuning
#if !defined(__linux__) && defined (SO_RCVBUF) && defined (SO_SNDBUF)
  if (type == SOCK_STREAM)
    n = rcvbufsize;
  else
    n = maxsobufsize;
  if (setsockopt (s, SOL_SOCKET, SO_RCVBUF, (char *) &n, sizeof (n)) < 0)
    warn ("SO_RCVBUF: %s\n", strerror (errno));

  if (type == SOCK_STREAM)
    n = sndbufsize;
  else
    n = maxsobufsize;
  if (setsockopt (s, SOL_SOCKET, SO_SNDBUF, (char *) &n, sizeof (n)) < 0)
    warn ("SO_SNDBUF: %s\n", strerror (errno));
#endif /* !__linux__ && SO_RCVBUF && SO_SNDBUF */

  /* Enable keepalives to avoid buffering write data for infinite time */
  n = 1;
  if (type == SOCK_STREAM
      && setsockopt (s, SOL_SOCKET, SO_KEEPALIVE,
		     (char *) &n, sizeof (n)) < 0)
    warn ("SO_KEEPALIVE: %s\n", strerror (errno));
}

void
make_sync (int s)
{
  int n;
  if ((n = fcntl (s, F_GETFL)) >= 0)
    fcntl (s, F_SETFL, n & ~O_NONBLOCK);
}

void
tcp_nodelay (int s)
{
#if defined (TCP_NODELAY) || defined (IPTOS_LOWDELAY)
  int n = 1;
#endif /* TCP_NODELAY || IPTOS_LOWDELAY */
#ifdef TCP_NODELAY
  if (setsockopt (s, IPPROTO_TCP, TCP_NODELAY, (char *) &n, sizeof (n)) < 0)
    warn ("TCP_NODELAY: %m\n");
#endif /* TCP_NODELAY */
#ifdef IPTOS_LOWDELAY
  setsockopt (s, IPPROTO_IP, IP_TOS, (char *) &n, sizeof (n));
#endif /* IPTOS_LOWDELAY */
}

void
tcp_abort (int fd)
{
  struct linger l;
  l.l_onoff = 1;
  l.l_linger = 0;
  setsockopt (fd, SOL_SOCKET, SO_LINGER, (char *) &l, sizeof (l));
  close (fd);
}

bool
addreq (const sockaddr *a, const sockaddr *b, socklen_t size)
{
  if (a->sa_family != b->sa_family)
    return false;
  switch (a->sa_family) {
  case AF_INET:
    if (implicit_cast<size_t> (size) >= sizeof (sockaddr_in)) {
      const sockaddr_in *aa = reinterpret_cast<const sockaddr_in *> (a);
      const sockaddr_in *bb = reinterpret_cast<const sockaddr_in *> (b);
      return (aa->sin_addr.s_addr == bb->sin_addr.s_addr
	      && aa->sin_port == bb->sin_port);
    }
    warn ("addreq: %d bytes is too small for AF_INET sockaddrs\n", size);
    return false;
  default:
    warn ("addreq: bad sa_family %d\n", a->sa_family);
    return false;
  }
}
