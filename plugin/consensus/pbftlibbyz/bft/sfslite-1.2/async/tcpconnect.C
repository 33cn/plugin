/* $Id: tcpconnect.C 1117 2005-11-01 16:20:39Z max $ */

/*
 *
 * Copyright (C) 2003 David Mazieres (dm@uun.org)
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

#include "async.h"
#include "dns.h"

struct tcpconnect_t {
  virtual ~tcpconnect_t () {}
};

struct tcpportconnect_t : tcpconnect_t {
  u_int16_t port;
  cbi cb;
  int fd;
  dnsreq_t *dnsp;
  str *namep;

  tcpportconnect_t (const in_addr &a, u_int16_t port, cbi cb);
  tcpportconnect_t (str hostname, u_int16_t port, cbi cb,
		bool dnssearch, str *namep);
  ~tcpportconnect_t ();

  void reply (int s) { if (s == fd) fd = -1; (*cb) (s); delete this; }
  void fail (int error) { errno = error; reply (-1); }
  void connect_to_name (str hostname, bool dnssearch);
  void name_cb (ptr<hostent> h, int err);
  void connect_to_in_addr (const in_addr &a);
  void connect_cb ();
};

tcpportconnect_t::tcpportconnect_t (const in_addr &a, u_int16_t p, cbi c)
  : port (p), cb (c), fd (-1), dnsp (NULL), namep (NULL)
{
  connect_to_in_addr (a);
}

tcpportconnect_t::tcpportconnect_t (str hostname, u_int16_t p, cbi c,
			    bool dnssearch, str *np)
  : port (p), cb (c), fd (-1), dnsp (NULL), namep (np)
{
  connect_to_name (hostname, dnssearch);
}

tcpportconnect_t::~tcpportconnect_t ()
{
  if (dnsp)
    dnsreq_cancel (dnsp);
  if (fd >= 0) {
    fdcb (fd, selwrite, NULL);
    close (fd);
  }
}

void
tcpportconnect_t::connect_to_name (str hostname, bool dnssearch)
{
  dnsp = dns_hostbyname (hostname, wrap (this, &tcpportconnect_t::name_cb),
			 dnssearch);
}

void
tcpportconnect_t::name_cb (ptr<hostent> h, int err)
{
  dnsp = NULL;
  if (!h) {
    if (dns_tmperr (err))
      fail (EAGAIN);
    else
      fail (ENOENT);
    return;
  }
  if (namep)
    *namep = h->h_name;
  connect_to_in_addr (*(in_addr *) h->h_addr);
}

void
tcpportconnect_t::connect_to_in_addr (const in_addr &a)
{
  sockaddr_in sin;
  bzero (&sin, sizeof (sin));
  sin.sin_family = AF_INET;
  sin.sin_port = htons (port);
  sin.sin_addr = a;

  fd = inetsocket (SOCK_STREAM);
  if (fd < 0) {
    delaycb (0, wrap (this, &tcpportconnect_t::fail, errno));
    return;
  }
  make_async (fd);
  close_on_exec (fd);
  if (connect (fd, (sockaddr *) &sin, sizeof (sin)) < 0
      && errno != EINPROGRESS) {
    delaycb (0, wrap (this, &tcpportconnect_t::fail, errno));
    return;
  }
  fdcb (fd, selwrite, wrap (this, &tcpportconnect_t::connect_cb));
}

void
tcpportconnect_t::connect_cb ()
{
  fdcb (fd, selwrite, NULL);

  sockaddr_in sin;
  socklen_t sn = sizeof (sin);
  if (!getpeername (fd, (sockaddr *) &sin, &sn)) {
    reply (fd);
    return;
  }

  int err = 0;
  sn = sizeof (err);
  getsockopt (fd, SOL_SOCKET, SO_ERROR, (char *) &err, &sn);
  fail (err ? err : ECONNREFUSED);
}

tcpconnect_t *
tcpconnect (in_addr addr, u_int16_t port, cbi cb)
{
  return New tcpportconnect_t (addr, port, cb);
}

tcpconnect_t *
tcpconnect (str hostname, u_int16_t port, cbi cb,
	    bool dnssearch, str *namep)
{
  return New tcpportconnect_t (hostname, port, cb, dnssearch, namep);
}

void
tcpconnect_cancel (tcpconnect_t *tc)
{
  delete tc;
}

struct tcpsrvconnect_t : tcpconnect_t {
  u_int16_t defport;
  cbi cb;
  int dnserr;
  dnsreq_t *areq;
  ptr<hostent> h;
  dnsreq_t *srvreq;
  ptr<srvlist> srvl;
  timecb_t *tmo;
  vec<tcpconnect_t *> cons;
  int cbad;
  int error;
  ptr<srvlist> *srvlp;
  str *namep;

  tcpsrvconnect_t (str name, str service, cbi cb, u_int16_t dp,
		   bool search, ptr<srvlist> *sp, str *np);
  tcpsrvconnect_t (ref<srvlist> sl, cbi cb, str *np);
  ~tcpsrvconnect_t ();
  void dnsacb (ptr<hostent>, int err);
  void dnssrvcb (ptr<srvlist>, int err);
  void maybe_start (int err);
  void connectcb (int cn, int fd);
  void nextsrv (bool timeout = false);
};

void
tcpsrvconnect_t::nextsrv (bool timeout)
{
  if (!timeout)
    timecb_remove (tmo);
  tmo = NULL;

  u_int n = cons.size ();

  if (n >= srvl->s_nsrv)
    return;

  // warn ("nextsrv %d (port %d)\n", n, srvl->s_srvs[n].port);

  if (!srvl->s_srvs[n].port || !srvl->s_srvs[n].name[0]) {
    cons.push_back (NULL);
    errno = ENOENT;
    connectcb (n, -1);
    return;
  }
  else if (h && !strcasecmp (srvl->s_srvs[n].name, h->h_name))
    cons.push_back (tcpconnect (*(in_addr *) h->h_addr, srvl->s_srvs[n].port,
				wrap (this, &tcpsrvconnect_t::connectcb, n)));
  else {
    str name = srvl->s_srvs[n].name;
    addrhint **hint;
    for (hint = srvl->s_hints;
	 *hint && ((*hint)->h_addrtype != AF_INET
		   || strcasecmp ((*hint)->h_name, name));
	 hint++)
      ;
    if (*hint)
      cons.push_back (tcpconnect (*(in_addr *) (*hint)->h_address,
				  srvl->s_srvs[n].port,
				  wrap (this, &tcpsrvconnect_t::connectcb,
					n)));
    else
      cons.push_back (tcpconnect (srvl->s_srvs[n].name, srvl->s_srvs[n].port,
				  wrap (this, &tcpsrvconnect_t::connectcb, n),
				  false));
  }

  tmo = delaycb (4, wrap (this, &tcpsrvconnect_t::nextsrv, true));
}

void
tcpsrvconnect_t::connectcb (int cn, int fd)
{
  cons[cn] = NULL;

  if (fd >= 0) {
    errno = 0;
    if (namep) {
      if (srvl) {
	*namep = srvl->s_srvs[cn].name;
	srvl->s_srvs[cn].port = 0;
      }
      else
	*namep = h->h_name;
    }
    (*cb) (fd);
    delete this;
    return;
  }

  // warn ("%s:%d %m\n", srvl->s_srvs[cn].name, srvl->s_srvs[cn].port);

  if (!error)
    error = errno;
  else if (errno == EAGAIN)
    error = errno;
  else if (error != EAGAIN && errno != ENOENT)
    error = errno;

  if (srvl)
    srvl->s_srvs[cn].port = 0;

  if (!srvl || ++cbad >= srvl->s_nsrv) {
    errno = error;
    (*cb) (-1);
    delete this;
    return;
  }

  if (!cons.back ())
    nextsrv ();
}

tcpsrvconnect_t::tcpsrvconnect_t (str name, str s, cbi cb, u_int16_t dp,
				  bool search, ptr<srvlist> *sp, str *np)
  : defport (dp), cb (cb), dnserr (0), tmo (NULL), cbad (0),
    error (0), srvlp (sp), namep (np)
{
  areq = dns_hostbyname (name, wrap (this, &tcpsrvconnect_t::dnsacb), search);
  srvreq = dns_srvbyname (name, "tcp", s,
			  wrap (this, &tcpsrvconnect_t::dnssrvcb), search);
}

tcpsrvconnect_t::tcpsrvconnect_t (ref<srvlist> sl, cbi cb, str *np)

  : defport (0), cb (cb), dnserr (0), areq (NULL), srvreq (NULL),
    tmo (NULL), cbad (0), error (0), srvlp (NULL), namep (np)
{
  delaycb (0, wrap (this, &tcpsrvconnect_t::dnssrvcb, sl, 0));
}

tcpsrvconnect_t::~tcpsrvconnect_t ()
{
  for (tcpconnect_t **cp = cons.base (); cp < cons.lim (); cp++)
    tcpconnect_cancel (*cp);
  dnsreq_cancel (areq);
  dnsreq_cancel (srvreq);
  timecb_remove (tmo);
}

void
tcpsrvconnect_t::maybe_start (int err)
{
  if (err && err != NXDOMAIN && err != ARERR_NXREC) {
    if (!dnserr)
      dnserr = err;
    else if (!dns_tmperr (dnserr) && dns_tmperr (err))
      dnserr = err;
  }
  if (srvreq || (!srvl && areq))
    return;
  if (srvl)
    nextsrv ();
  else if (h && defport) {
    cons.push_back (tcpconnect (*(in_addr *) h->h_addr, defport,
				wrap (this, &tcpsrvconnect_t::connectcb, 0)));
  }
  else {
    if (dns_tmperr (dnserr))
      errno = EAGAIN;
    else
      errno = ENOENT;
    (*cb) (-1);
    delete this;
  }
}

void
tcpsrvconnect_t::dnsacb (ptr<hostent> hh, int err)
{
  areq = NULL;
  h = hh;
  maybe_start (err);
}

void
tcpsrvconnect_t::dnssrvcb (ptr<srvlist> s, int err)
{
  srvreq = NULL;
  srvl = s;
  if (srvlp)
    *srvlp = srvl;
  maybe_start (err);
}

tcpconnect_t *
tcpconnect_srv (str hostname, str service, u_int16_t defport,
		cbi cb, bool dnssearch, ptr<srvlist> *srvlp, str *np)
{
  if (srvlp && *srvlp)
    return New tcpsrvconnect_t (*srvlp, cb, np);
  else
    return New tcpsrvconnect_t (hostname, service, cb, defport,
				dnssearch, srvlp, np);
}

tcpconnect_t *
tcpconnect_srv_retry (ref<srvlist> srvl, cbi cb, str *np)
{
  return New tcpsrvconnect_t (srvl, cb, np);
}

