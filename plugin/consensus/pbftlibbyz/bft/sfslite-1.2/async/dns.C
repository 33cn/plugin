/* $Id: dns.C 3769 2008-11-13 20:21:34Z max $ */

/*
 *
 * Copyright (C) 1998-2003 David Mazieres (dm@uun.org)
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


/* Asynchronous resolver.  This code is easy to understand if and only
 * if you also have a copy of RFC 1035.  You can get this by anonymous
 * ftp from, among other places, ftp.internic.net:/rfc/rfc1035.txt.
 */

#include "dnsimpl.h"
#include "arena.h"

#include "vec.h"
#include "ihash.h"
#include "list.h"
#include "backoff.h"

#define NSPORT NAMESERVER_PORT

static resolv_conf *_resconf;
static resolv_conf *
resconf ()
{
  if (!_resconf)
    _resconf = New resolv_conf();
  return _resconf;
}

dnssock_udp::dnssock_udp (int f, cb_t cb)
  : dnssock (false, cb), fd (f)
{
  fdcb (fd, selread, wrap (this, &dnssock_udp::rcb));
}

dnssock_udp::~dnssock_udp ()
{
  fdcb (fd, selread, NULL);
  //fdcb (fd, selwrite, NULL);
  close (fd);
}

void
dnssock_udp::sendpkt (const u_char *pkt, size_t size)
{
  if (send (fd, pkt, size, 0) < 0 && errno != EAGAIN)
    (*cb) (NULL, -1);
}

void
dnssock_udp::rcb ()
{
  ref<bool> d = destroyed;
  int n;
  do {
    u_char qb[QBSIZE];
    n = recv (fd, reinterpret_cast<char *> (qb), sizeof (qb), 0);
    if (n > 0)
      (*cb) (qb, n);
    else if (!n || errno != EAGAIN)
      (*cb) (NULL, -1);
  } while (n > 0 && !*d);
}


dnssock_tcp::dnssock_tcp (int f, cb_t cb)
  : dnssock (true, cb), fd (f), write_ok (false)
{
  fdcb (fd, selread, wrap (this, &dnssock_tcp::rcb));
  fdcb (fd, selwrite, wrap (this, &dnssock_tcp::wcb, true));
}

dnssock_tcp::~dnssock_tcp ()
{
  fdcb (fd, selread, NULL);
  fdcb (fd, selwrite, NULL);
  close (fd);
}

void
dnssock_tcp::rcb ()
{
  if (tcpstate.input (fd) < 0) {
    (*cb) (NULL, -1);
    return;
  }

  ref<bool> d = destroyed;
  u_char *qb;
  size_t n;
  while (!*d && tcpstate.getpkt (&qb, &n))
    (*cb) (qb, n);
}

void
dnssock_tcp::wcb (bool selected)
{
  if (selected)
    write_ok = true;
  if (!write_ok)
    return;
  int n = tcpstate.output (fd);
  if (n < 0) {
    fdcb (fd, selwrite, NULL);
    (*cb) (NULL, -1);
  }
  else if (n > 0)
    fdcb (fd, selwrite, NULL);
  else
    fdcb (fd, selwrite, wrap (this, &dnssock_tcp::wcb, true));
}

void
dnssock_tcp::sendpkt (const u_char *pkt, size_t size)
{
  tcpstate.putpkt (pkt, size);
  wcb ();
}


resolver::resolver ()
  : nbump (0), addr (NULL), addrlen (0), udpcheck_req (NULL),
    last_resp (0), last_bump (0), destroyed (New refcounted<bool> (false))
{
}

resolver::~resolver ()
{
  //cantsend ();			// bad if exit called from callback
  delete udpcheck_req;
  *destroyed = true;
}

bool
resolver::udpinit ()
{
  udpsock = NULL;
  int fd = socket (addr->sa_family, SOCK_DGRAM, 0);
  if (fd < 0) {
    warn ("resolver::udpsock: socket: %m\n");
    return false;
  }
  make_async (fd);
  close_on_exec (fd);
  if (connect (fd, addr, addrlen) < 0) {
    warn ("resolver::udpsock: connect: %m\n");
    close (fd);
    return false;
  }
  udpsock = New refcounted<dnssock_udp> (fd, wrap (this, &resolver::pktready,
						   false));
  return true;
}

bool
resolver::tcpinit ()
{
  tcpsock = NULL;
  int fd = socket (addr->sa_family, SOCK_STREAM, 0);
  if (fd < 0) {
    warn ("resolver::tcpsock: socket: %m\n");
    return false;
  }
  make_async (fd);
  close_on_exec (fd);
  if (connect (fd, addr, addrlen) < 0 && errno != EINPROGRESS) {
    close (fd);
    return false;
  }
  tcpsock = New refcounted<dnssock_tcp> (fd, wrap (this, &resolver::pktready,
						   true));
  return true;
}

void
resolver::cantsend ()
{
  ref<bool> d = destroyed;
  for (dnsreq *r = reqtab.first (), *nr; !*d && r; r = nr) {
    nr = reqtab.next (r);
    failreq (ARERR_CANTSEND, r);
  }
}

bool
resolver::resend (bool udp, bool tcp)
{
  ref<bool> d = destroyed;
  for (dnsreq *r = reqtab.first (), *nr; !*d && r; r = nr) {
    nr = reqtab.next (r);
    if (r->usetcp) {
      if (tcp && tcpsock)
	sendreq (r);
      else if (tcp)
	failreq (ARERR_CANTSEND, r);
    }
    else if (udp && udpsock) {
      reqtoq.remove (r);
      reqtoq.start (r);
    }
  }
  return !*d;
}

bool
resolver::setsock (bool failure)
{
  if (udpcheck_req) {
    delete udpcheck_req;
    udpcheck_req = NULL;
  }

  do {
    if ((failure || !addr) && !bumpsock (failure))
      return false;
    failure = true;
    nbump++;
    last_resp = 0;
    last_bump = sfs_get_timenow();
    tcpsock = NULL;
  } while (!udpinit () || !tcpinit ());

  return resend (true, true);
}

void
resolver::sendreq (dnsreq *r)
{
  if (!udpsock) {
    setsock (false);
    return;
  }

  ptr<dnssock> sock;
  if (!r->usetcp)
    sock = udpsock;
  else if (!tcpsock && !tcpinit ()) {
    setsock (true);
    return;
  }
  else
    sock = tcpsock;

  u_char qb[QBSIZE];
  int n;
  n = res_mkquery (QUERY, r->name, C_IN, r->type,
		   NULL, 0, NULL, qb, sizeof (qb));
  //warn ("query (%s, %d): %d\n", r->name.cstr (), r->type, n);
  if (n < 0) {
    r->fail (ARERR_REQINVAL);
    return;
  }

  HEADER *const h = (HEADER *) qb;
  h->id = r->id;
  h->rd = 1;

  /* FreeBSD (and possibly other OSes) have a broken dn_expand
   * function that doesn't properly invert dn_comp.
   */
  {
    dnsparse query (qb, n, false);
    question q;
    if (query.qparse (&q))
      r->name = q.q_name;
  }

  sock->sendpkt (qb, n);
}

void
resolver::pktready (bool tcp, u_char *qb, ssize_t n)
{
  if (n <= 0) {
    if (tcp) {
      tcpsock = NULL;
      if (!last_resp)
	setsock (true);
      last_resp = 0;
      resend (false, true);
    }
    else {
      udpsock = NULL;
      setsock (true);
    }
    return;
  }

  nbump = 0;
  last_resp = sfs_get_timenow();

  dnsparse reply (qb, n);
  question q;
  if (!reply.qparse (&q) || q.q_class != C_IN)
    return;

  dnsreq *r;
  for (r = reqtab[reply.hdr->id];
       r && (r->usetcp != tcp || r->type != q.q_type
	     || strcasecmp (r->name, q.q_name));
       r = reqtab.nextkeq (r))
    ;
  if (!r)
    return;

  if (reply.error && !r->error)
    r->error = reply.error;
  if (r->error == NXDOMAIN) {
    r->error = 0;
    r->start (true);
  }
  else if (!r->error && !r->usetcp && reply.hdr->tc) {
    reqtoq.remove (r);
    r->usetcp = true;
    r->xmit (0);
  }
  else
    r->readreply (r->error ? NULL : &reply);
}

u_int16_t
resolver::genid ()
{
  u_int16_t id;
  int i = 0;
  do {
    id = arandom () % 0xffff;
  } while (reqtab[id] && ++i < 8);
  return id;
}

void
resolver::udpcheck ()
{
  if (!udpcheck_req)
    udpcheck_req = New dnsreq_a (this, "",
				 wrap (this, &resolver::udpcheck_cb),
				 false);
}

void
resolver::udpcheck_cb (ptr<hostent> h, int err)
{
  udpcheck_req = NULL;
  if (err == ARERR_TIMEOUT)
    setsock (true);
}


resolv_conf::resolv_conf ()
  : ns_idx (0), last_reload (0), reload_lock (false),
    destroyed (New refcounted<bool> (false))
{
  if (!(_res.options & RES_INIT))
    res_init ();
  bzero (&srvaddr, sizeof (srvaddr));
  srvaddr.sin_family = AF_INET;
  srvaddr.sin_port = htons (NSPORT);
  ifc = ifchgcb (wrap (this, &resolv_conf::reload, false));
  ns_idx =  _res.nscount ? _res.nscount - 1 : 0;
}

resolv_conf::~resolv_conf ()
{
  *destroyed = true;
  ifchgcb_remove (ifc);
}

void
resolv_conf::reload (bool failure)
{
  if (reload_lock)
    return;
  reload_lock = true;
  chldrun (wrap (&reload_dumpres),
	   wrap (this, &resolv_conf::reload_cb, destroyed, failure));
}

void
resolv_conf::reload_dumpres (int fd)
{
  make_sync (fd);
  bzero (&_res, sizeof (_res));
  res_init ();
  v_write (fd, &_res, sizeof (_res));
  _exit (0);
}

void
resolv_conf::reload_cb (ref<bool> d, bool failure, str newres)
{
  if (*d)
    return;

  nbump = 0;
  reload_lock = false;
  last_reload = sfs_get_timenow();
  if (!newres) {
    warn ("resolv_conf::reload_cb: fork: %m\n");
    setsock (true);
    return;
  }
  if (newres.len () != sizeof (_res)) {
    warn ("resolv_conf::reload_cb: short read\n");
    setsock (true);
    return;
  }

  char oldnsaddr[sizeof (_res.nsaddr_list)];
  memcpy (oldnsaddr, _res.nsaddr_list, sizeof (oldnsaddr));
  memcpy (&_res, newres, sizeof (_res));
  if (memcmp (oldnsaddr, _res.nsaddr_list, sizeof (oldnsaddr))) {
    warn ("reloaded DNS configuration (resolv.conf)\n");
    ns_idx =  _res.nscount ? _res.nscount - 1 : 0;
    //nbump = 0;
    last_reload = sfs_get_timenow();
    setsock (true);
  }
  else
    setsock (failure);
}

bool
resolv_conf::bumpsock (bool failure)
{
  if (reload_lock)
    return false;
  if (sfs_get_timenow() > last_reload + 60) {
    reload (failure);
    return false;
  }

  if (nbump >= _res.nscount) {
    cantsend ();
    return false;
  }

  ns_idx = (ns_idx + 1) % _res.nscount;
  if (failure
      && (!addr || addrlen != sizeof (srvaddr)
	  || !addreq (addr, (reinterpret_cast<sockaddr *>
			     (&_res.nsaddr_list[ns_idx])),
		      addrlen)))
    warn ("changing nameserver to %s\n",
	  inet_ntoa (_res.nsaddr_list[ns_idx].sin_addr));

  srvaddr = _res.nsaddr_list[ns_idx];
  if (!srvaddr.sin_addr.s_addr)
    srvaddr.sin_addr.s_addr = htonl (INADDR_LOOPBACK);

  addr = reinterpret_cast<sockaddr *> (&srvaddr);
  addrlen = sizeof (srvaddr);

  return true;
}

const char *
resolv_conf::srchlist (int n)
{
  if (n <= 0)
    return "";
  return _res.dnsrch[n - 1];
}


dnsreq::dnsreq (resolver *rp, str n, u_int16_t t, bool search)
  : ntries (0), resp (rp), usetcp (false), constructed (false),
    error (0), type (t)
{
  while (n.len () && n[n.len () - 1] == '.') {
    search = false;
    n = substr (n, 0, n.len () - 1);
  }
  if (!search) {
    srchno = -1;
    basename = NULL;
    name = n;
  }
  else {
    srchno = 0;
    basename = n;
    name = NULL;
  }
  start (false);
  constructed = true;
}

void
dnsreq::remove ()
{
  if (intable) {
    intable = false;
    resp->reqtab.remove (this);
    if (!usetcp)
      resp->reqtoq.remove (this);
  }
}

dnsreq::~dnsreq ()
{
  remove ();
}

void
dnsreq::start (bool again)
{
  if (again && (srchno < 0 || !resp->srchlist (srchno))) {
    fail (NXDOMAIN);
    return;
  }

  if (again) {
    resp->reqtab.remove (this);
    if (!usetcp)
      resp->reqtoq.remove (this);
  }
  if (srchno >= 0) {
    const char *suffix = resp->srchlist (srchno++);
    if (*suffix)
      name = strbuf ("%s.%s", basename.cstr (), suffix);
    else
      name = basename;
  }
  id = resp->genid ();
  intable = true;
  resp->reqtab.insert (this);
  if (usetcp)
    xmit (0);
  else
    resp->reqtoq.start (this);
}

void
dnsreq::xmit (int retry)
{
  error = 0;
  resp->sendreq (this);
}

void
dnsreq::timeout ()
{
  assert (!usetcp);
  if (sfs_get_timenow() - resp->last_resp < 90 || !name.len ())
    fail (ARERR_TIMEOUT);
  else {
    resp->reqtoq.keeptrying (this);
    resp->udpcheck ();
  }
}

void
dnsreq::fail (int err)
{
  assert (err);
  if (!error)
    error = err;
  if (constructed)
    readreply (NULL);
  else {
    remove ();
    delaycb (0, wrap (this, &dnsreq::readreply, (dnsparse *) NULL));
  }
}

void
dnsreq_cancel (dnsreq *rqp)
{
  delete rqp;
}


void
dnsreq_a::readreply (dnsparse *reply)
{
  ptr<hostent> h;
  if (!error) {
    assert (reply);
    if (!(h = reply->tohostent ()))
      error = reply->error;
    else if (checkaddr) {
      char **ap;
      for (ap = h->h_addr_list; *ap && *(in_addr *) *ap != addr; ap++)
	;
      if (!*ap) {
	h = NULL;
	error = ARERR_PTRSPOOF;
      }
    }
  }
  (*cb) (h, error);
  delete this;
}

dnsreq *
dns_hostbyname (str name, cbhent cb,
		bool search, bool addrok)
{
  if (addrok) {
    in_addr addr;
    if (name.len () && isdigit (name[name.len () - 1])
	&& inet_aton (name.cstr (), &addr)) {
      ptr<hostent> h = refcounted<hostent, vsize>::alloc
	(sizeof (*h) + 3 * sizeof (void *)
	 + sizeof (addr) + strlen (name) + 1);
      h->h_aliases = (char **) &h[1];
      h->h_addrtype = AF_INET;
      h->h_length = sizeof (addr);
      h->h_addr_list = &h->h_aliases[1];

      h->h_aliases[0] = NULL;
      h->h_addr_list[0] = (char *) &h->h_addr_list[2];
      h->h_addr_list[1] = NULL;

      *(struct in_addr *) h->h_addr_list[0] = addr;
      h->h_name = (char *) h->h_addr_list[0] + sizeof (addr);
      strcpy ((char *) h->h_name, name);

      (*cb) (h, 0);
      return NULL;
    }
  }
  return New dnsreq_a (resconf(), name, cb, search);
}


void
dnsreq_mx::readreply (dnsparse *reply)
{
  ptr<mxlist> m;
  if (!error) {
    if (!(m = reply->tomxlist ()))
      error = reply->error;
  }
  (*cb) (m, error);
  delete this;
}

dnsreq *
dns_mxbyname (str name, cbmxlist cb, bool search)
{
  return New dnsreq_mx (resconf(), name, cb, search);
}


void
dnsreq_srv::readreply (dnsparse *reply)
{
  ptr<srvlist> s;
  if (!error) {
    if (!(s = reply->tosrvlist ()))
      error = reply->error;
  }
  (*cb) (s, error);
  delete this;
}

dnsreq *
dns_srvbyname (str name, cbsrvlist cb, bool search)
{
  return New dnsreq_srv (resconf(), name, cb, search);
}


dnsreq_ptr::~dnsreq_ptr ()
{
  while (!vrfyv.empty ())
    delete vrfyv.pop_front ();
}

void
dnsreq_ptr::maybe_push (vec<str, 2> *sv, const char *s)
{
  for (const str *sp = sv->base (); sp < sv->lim (); sp++)
    if (!strcasecmp (sp->cstr (), s))
      return;
  sv->push_back (s);
}

str
dnsreq_ptr::inaddr_arpa (in_addr addr)
{
  u_char *a = reinterpret_cast<u_char *> (&addr);
  return strbuf ("%d.%d.%d.%d.in-addr.arpa", a[3], a[2], a[1], a[0]);
}

void
dnsreq_ptr::readreply (dnsparse *reply)
{
  vec<str, 2> names;
  if (!error) {
    const u_char *cp = reply->getanp ();
    for (u_int i = 0; i < reply->ancount; i++) {
      resrec rr;
      if (!reply->rrparse (&cp, &rr))
	break;
      if (rr.rr_type == T_PTR && rr.rr_class == C_IN)
	maybe_push (&names, rr.rr_ptr);
    }

    if (!names.empty ()) {
      napending = names.size ();
      remove ();
      for (u_int i = 0; i < names.size (); i++)
	vrfyv.push_back (New dnsreq_a (resp, names[i],
				       wrap (this, &dnsreq_ptr::readvrfy, i),
				       addr));
      return;
    }
  }

  if (!error && !(error = reply->error))
    error = ARERR_NXREC;
  (*cb) (NULL, error);
  delete this;
}

void
dnsreq_ptr::readvrfy (int i, ptr<hostent> h, int err)
{
  vrfyv[i] = NULL;
  if (err && (dns_tmperr (err) || !error))
    error = err;
  if (h) {
    maybe_push (&vnames, h->h_name);
    for (char **np = h->h_aliases; *np; np++)
      maybe_push (&vnames, *np);
  }
  if (--napending)
    return;

  if (vnames.empty () && !error)
    error = ARERR_PTRSPOOF;
  if (error) {
    (*cb) (NULL, error);
    delete this;
    return;
  }

  u_int namelen = 0;
  for (str *np = vnames.base (); np < vnames.lim (); np++)
    namelen += np->len () + 1;

  int hsize = (sizeof (*h)
		+ (vnames.size () + 1) * sizeof (char *)
	       + namelen
	       + 2 * sizeof (char *)
	       + sizeof (in_addr));
  h = refcounted<hostent, vsize>::alloc (hsize);
  h->h_addrtype = AF_INET;
  h->h_length = sizeof (in_addr);
  h->h_aliases = (char **) &h[1];
  h->h_addr_list = &h->h_aliases[vnames.size ()];

  h->h_addr_list[0] = (char *) &h->h_addr_list[2];
  h->h_addr_list[1] = NULL;
  *(in_addr *) h->h_addr_list[0] = addr;

  char *dp = h->h_addr_list[0] + sizeof (in_addr);
  memcpy (h->h_name = dp, vnames[0], vnames[0].len () + 1);
  dp += vnames[0].len () + 1;
  vnames.pop_front ();
  char **ap = h->h_aliases;
  while (!vnames.empty ()) {
    *ap = dp;
    memcpy (dp, vnames.front (), vnames.front ().len () + 1);
    dp += vnames.front ().len () + 1;
    ap++;
    vnames.pop_front ();
  }
  *ap = NULL;

  (*cb) (h, error);
  delete this;
}

dnsreq *
dns_hostbyaddr (in_addr addr, cbhent cb)
{
  return New dnsreq_ptr (resconf(), addr, cb);
}


void
dnsreq_txt::readreply (dnsparse *reply)
{
  ptr<txtlist> t;
  if (!error && !(t = reply->totxtlist ()))
    error = reply->error;
  (*cb) (t, error);
  delete this;
}

dnsreq_t *
dns_txtbyname (str name, cbtxtlist cb, bool search)
{
  return New dnsreq_txt (resconf(), name, cb, search);
}


const char *
dns_strerror (int no)
{
  switch (no) {
  case NOERROR:
    return "no error";
  case FORMERR:
    return "DNS format error";
  case SERVFAIL:
    return "name server failure";
  case NXDOMAIN:
    return "non-existent domain name";
  case NOTIMP:
    return "unimplemented DNS request";
  case REFUSED:
    return "DNS query refused";
  case ARERR_NXREC:
    return "no DNS records of appropriate type";
  case ARERR_TIMEOUT:
    return "name lookup timed out";
  case ARERR_PTRSPOOF:
    return "incorrect PTR record";
  case ARERR_BADRESP:
    return "malformed DNS reply";
  case ARERR_CANTSEND:
    return "cannot send to name server";
  case ARERR_REQINVAL:
    return "malformed domain name";
  case ARERR_CNAMELOOP:
    return "CNAME records form loop";
  default:
    return "unknown DNS error";
  }
}

int
dns_tmperr (int no)
{
  switch (no) {
  case SERVFAIL:
  case ARERR_TIMEOUT:
  case ARERR_CANTSEND:
  case ARERR_BADRESP:
    return 1;
  default:
    return 0;
  }
}

