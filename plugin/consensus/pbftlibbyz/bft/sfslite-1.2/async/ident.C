/* $Id: ident.C 1117 2005-11-01 16:20:39Z max $ */

/*
 *
 * Copyright (C) 1999 David Mazieres (dm@uun.org)
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

#include "rxx.h"
#include "async.h"
#include "dns.h"
#include "aios.h"

#define AUTH_PORT 113

static rxx identrx ("^([^:]*:){3}\\s*(.*?)\\s*$");

struct identstat {
  ptr<aios> a;
  int ncb;
  int err;
  str user;
  str host;
  ptr<hostent> h;
  callback<void, str, ptr<hostent>, int>::ptr cb;

  void cbdone ();
  void identcb (str u, int e) {
    if (u && identrx.search (u))
      user = identrx[2];
    a = NULL;
    cbdone ();
  }
  void dnscb (ptr<hostent> hh, int e) {
    h = hh;
    err = e;
    if (h && *h->h_name)
      host = h->h_name;
    cbdone ();
  }
};

void
identstat::cbdone ()
{
  if (--ncb)
    return;
  str res;
  if (user)
    res = user << "@" << host;
  else
    res = host;
  (*cb) (res, h, err);
  delete this;
}

void
identptr (int fd, callback<void, str, ptr<hostent>, int>::ref cb)
{
  struct sockaddr_in la, ra;
  socklen_t len;

  len = sizeof (la);
  bzero (&la, sizeof (la));
  bzero (&ra, sizeof (ra));
  errno = 0;
  if (getsockname (fd, (struct sockaddr *) &la, &len) < 0
      || la.sin_family != AF_INET
      || getpeername (fd, (struct sockaddr *) &ra, &len) < 0
      || ra.sin_family != AF_INET
      || len != sizeof (la)) {
    warn ("ident: getsockname/getpeername: %s\n", strerror (errno));
    (*cb) ("*disconnected*", NULL, ARERR_CANTSEND);
    return;
  }

  u_int lp = ntohs (la.sin_port);
  la.sin_port = htons (0);
  u_int rp = ntohs (ra.sin_port);
  ra.sin_port = htons (AUTH_PORT);

  int ifd = socket (AF_INET, SOCK_STREAM, 0);
  if (ifd >= 0) {
    close_on_exec (ifd);
    make_async (ifd);
    if (connect (ifd, (sockaddr *) &ra, sizeof (ra)) < 0
	&& errno != EINPROGRESS) {
      close (ifd);
      ifd = -1;
    }
  }

  identstat *is = New identstat;
  is->err = 0;
  is->cb = cb;
  is->host = inet_ntoa (ra.sin_addr);

  if (ifd >= 0) {
    is->ncb = 2;
    close_on_exec (ifd);
    is->a = aios::alloc (ifd);
    is->a << rp << ", " << lp << "\r\n";
    is->a->settimeout (15);
    is->a->readline (wrap (is, &identstat::identcb));
  }
  else
    is->ncb = 1;

  dns_hostbyaddr (ra.sin_addr, wrap (is, &identstat::dnscb));
}

static void
strip_hostent (callback<void, str, int>::ref cb, str id, int err)
{
  (*cb) (id, err);
}
void
ident (int fd, callback<void, str, int>::ref cb)
{
  ident (fd, wrap (strip_hostent, cb));
}
