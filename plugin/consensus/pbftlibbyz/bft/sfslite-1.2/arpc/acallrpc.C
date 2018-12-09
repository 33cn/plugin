/* $Id: acallrpc.C 2508 2007-01-12 23:39:52Z yipal $ */

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

#include "dns.h"
#include "arpc.h"
#include "xdr_suio.h"
#include "pmap_prot.h"

static ptr<axprt_dgram> _udpxprt;
static ptr<aclnt> _udpclnt;

static rpc_program noprog;

ptr<axprt_stream> aclnt_axprt_stream_alloc (size_t ps, int fd);
const axprtalloc_fn axprt_stream_alloc_default
  = gwrap (aclnt_axprt_stream_alloc, int (axprt_stream::defps));

static void
acallrpc_init ()
{
  int udpfd;
  if (suidsafe ())
    udpfd = inetsocket_resvport (SOCK_DGRAM);
  else
    udpfd = inetsocket (SOCK_DGRAM);
  if (udpfd < 0)
    fatal ("acallrpc_init: inetsocket: %m\n");
  close_on_exec (udpfd);
  if (!(_udpxprt = axprt_dgram::alloc (udpfd)))
    fatal ("acallrpc_init: axprt_dgram::alloc failed\n");
  if (!(_udpclnt = aclnt::alloc (_udpxprt, noprog, NULL)))
    fatal ("acallrpc_init: aclnt::alloc failed\n");
}

ptr<axprt_dgram> udpxprt() {
  if (!_udpxprt)
    acallrpc_init ();
  return _udpxprt;
}

ptr<aclnt> udpclnt() {
  if (!_udpclnt)
    acallrpc_init ();
  return _udpclnt;
}

ptr<axprt_stream>
aclnt_axprt_stream_alloc (size_t ps, int fd)
{
  return axprt_stream::alloc (fd, ps);
}

class rpc2sin {
  int port;

  void dnscb (ptr<hostent> h, int err) {
    if (h) {
      sin.sin_addr = *(in_addr *) h->h_addr;
      getport ();
    }
    else
      gotaddr (RPC_UNKNOWNHOST);
  }

  void getport () {
    if (port) {
      sin.sin_port = htons (port);
      gotaddr (RPC_SUCCESS);
      return;
    }

    sin.sin_port = htons (PMAP_PORT);
    mapping pm;
    pm.prog = prog;
    pm.vers = vers;
    pm.prot = prot;
    pm.port = 0;
    udpclnt ()->call (PMAPPROC_GETPORT, (void *) &pm, (void *) &port,
                      wrap (this, &rpc2sin::gotport),
                      (AUTH *) 0, xdr_mapping, xdr_int,
                      PMAP_PROG, PMAP_VERS, (sockaddr *) &sin);
  }

  void gotport (clnt_stat stat) {
    if (stat)
      gotaddr (RPC_PMAPFAILURE);
    else if (!port)
      gotaddr (RPC_PROGNOTREGISTERED);
    else {
      sin.sin_port = htons (port);
      gotaddr (RPC_SUCCESS);
    }
  }

protected:
  const u_int32_t prog;
  const u_int32_t vers;
  const u_int32_t prot;

  struct sockaddr_in sin;

  rpc2sin (u_int32_t prog, u_int32_t vers, u_int32_t prot)
    : prog (prog), vers (vers), prot (prot) {
    bzero (&sin, sizeof (sin));
    sin.sin_family = AF_INET;
  }
  virtual ~rpc2sin () {}

  void getaddr (const char *name, int portno = 0) {
    port = portno;
    dns_hostbyname (name, wrap (this, &rpc2sin::dnscb), true, true);
  }

  void getaddr (in_addr addr, int portno = 0) {
    sin.sin_addr = addr;
    port = portno;
    getport ();
  }

  virtual void gotaddr (clnt_stat) = 0;
};

class acallrpcobj : rpc2sin {
  char *callbuf;
  size_t calllen;

  bool used;

  u_int32_t proc;
  sfs::xdrproc_t outxdr;
  void *outmem;
  aclnt_cb cb;
  AUTH *auth;

  void setmsg (sfs::xdrproc_t inxdr, void *inmem) {
    callbuf = NULL;
    xdrsuio x (XDR_ENCODE);
    if (aclnt::marshal_call (x, auth, prog, vers,
			      proc, inxdr, inmem)) {
      calllen = x.uio ()->resid ();
      callbuf = suio_flatten (x.uio ());
    }
  }

  void gotaddr (clnt_stat stat) {
    if (stat)
      done (stat);
    else {
      char *msg = callbuf;
      callbuf = NULL;
      vNew rpccb_unreliable (udpclnt(), msg, calllen,
			     wrap (this, &acallrpcobj::done),
			     outmem, outxdr, (sockaddr *) &sin);
    }
  }

  void done (clnt_stat stat) {
    (*cb) (stat);
    delete this;
  }

  PRIVDEST ~acallrpcobj () { xfree (callbuf); }

public:

  acallrpcobj (u_int32_t prog, u_int32_t vers,
	       u_int32_t proc, sfs::xdrproc_t inxdr, void *inmem,
	       sfs::xdrproc_t outxdr, void *outmem,
	       aclnt_cb cb, AUTH *auth = NULL)
    : rpc2sin (prog, vers, IPPROTO_UDP),
      used (false), proc (proc),
      outxdr (outxdr), outmem (outmem), cb (cb), auth (auth)
    { setmsg (inxdr, inmem); }

  void call (const char *host, int port = 0) {
    assert (!used);
    used = true;
    if (callbuf)
      getaddr (host, port);
    else
      done (RPC_CANTENCODEARGS);
  }
  void call (const in_addr &addr, int port = 0) {
    assert (!used);
    used = true;
    if (callbuf)
      getaddr (addr, port);
    else
      done (RPC_CANTENCODEARGS);
  }
};

void
__acallrpc (const char *host, u_int port,
	    u_int32_t prog, u_int32_t vers, u_int32_t proc,
	    sfs::xdrproc_t inxdr, void *inmem, sfs::xdrproc_t outxdr, void *outmem,
	    aclnt_cb cb, AUTH *auth)
{
  acallrpcobj *co = New acallrpcobj (prog, vers, proc, inxdr, inmem,
				     outxdr, outmem, cb, auth);
  co->call (host, port);
}

void
__acallrpc (in_addr host, u_int port,
	    u_int32_t prog, u_int32_t vers, u_int32_t proc,
	    sfs::xdrproc_t inxdr, void *inmem, sfs::xdrproc_t outxdr, void *outmem,
	    aclnt_cb cb, AUTH *auth)
{
  acallrpcobj *co = New acallrpcobj (prog, vers, proc, inxdr, inmem,
				     outxdr, outmem, cb, auth);
  co->call (host, port);
}


void
acallrpc (const sockaddr_in *sinp, const rpc_program &rp, u_int32_t proc,
	  void *in, void *out, aclnt_cb cb, AUTH *auth)
{
  // XXX - the const part of the cast to sockaddr * is not quite right
  assert (proc < rp.nproc);
  udpclnt()->call (proc, in, out, cb, auth,
                   rp.tbl[proc].xdr_arg, rp.tbl[proc].xdr_res,
                   rp.progno, rp.versno,
                   (sockaddr *) (sinp));
}

class aclntudpobj : rpc2sin {
  const rpc_program &rp;
  aclntalloc_cb cb;

public:
  aclntudpobj (const char *host, int port, const rpc_program &rp,
	       aclntalloc_cb cb)
    : rpc2sin (rp.progno, rp.versno, IPPROTO_UDP), rp (rp), cb (cb)
    { getaddr (host, port); }
  aclntudpobj (const in_addr &addr, int port, const rpc_program &rp,
	       aclntalloc_cb cb)
    : rpc2sin (rp.progno, rp.versno, IPPROTO_UDP), rp (rp), cb (cb)
    { getaddr (addr, port); }

  PRIVDEST ~aclntudpobj () {}

private:
  void gotaddr (clnt_stat stat) {
    if (stat)
      (*cb) (NULL, stat);
    else
      (*cb) (aclnt::alloc (udpxprt(), rp, (sockaddr *) &sin), RPC_SUCCESS);
    delete this;
  }
};

void
aclntudp_create (const char *host, int port, const rpc_program &rp,
		 aclntalloc_cb cb)
{
  vNew aclntudpobj (host, port, rp, cb);
}
void
aclntudp_create (const in_addr &addr, int port, const rpc_program &rp,
		 aclntalloc_cb cb)
{
  vNew aclntudpobj (addr, port, rp, cb);
}

class aclnttcpobj : rpc2sin {
  const rpc_program &rp;
  callback<void, int, clnt_stat>::ref cb;
  int fd;

public:
  aclnttcpobj (const char *host, int port, const rpc_program &rp,
	       callback<void, int, clnt_stat>::ref cb)
    : rpc2sin (rp.progno, rp.versno, IPPROTO_TCP), rp (rp), cb (cb)
    { getaddr (host, port); }
  aclnttcpobj (const in_addr &addr, int port, const rpc_program &rp,
	       callback<void, int, clnt_stat>::ref cb)
    : rpc2sin (rp.progno, rp.versno, IPPROTO_TCP), rp (rp), cb (cb)
    { getaddr (addr, port); }

  PRIVDEST ~aclnttcpobj () {}

private:
  void finish (int f, clnt_stat stat) {
    (*cb) (f, stat);
    delete this;
  }

  void gotaddr (clnt_stat stat) {
    if (stat) {
      finish (NULL, stat);
      return;
    }
    fd = inetsocket_resvport (SOCK_STREAM);
    if (fd < 0) {
      finish (-1, RPC_FAILED);
      return;
    }
    make_async (fd);
    if (connect (fd, (sockaddr *) &sin, sizeof (sin)) < 0
	&& errno != EINPROGRESS) {
      close (fd);
      finish (-1, RPC_FAILED);
      return;
    }
    fdcb (fd, selwrite, wrap (this, &aclnttcpobj::connected));
  }

  void connected () {
    fdcb (fd, selwrite, NULL);
    sockaddr_in xsin;
    socklen_t xlen = sizeof (xsin);
    if (getpeername (fd, (sockaddr *) &xsin, &xlen) < 0) {
      close (fd);
      finish (-1, RPC_FAILED);
      return;
    }
    finish (fd, RPC_SUCCESS);
  }
};

static void
aclnttcp_create_finish (const rpc_program *rpp,
			aclntalloc_cb cb, axprtalloc_fn xa,
			int fd, clnt_stat stat)
{
  if (fd < 0)
    (*cb) (NULL, stat);
  else if (ptr<axprt> ax = (*xa) (fd))
    (*cb) (aclnt::alloc (ax, *rpp), stat);
  else
    (*cb) (NULL, RPC_FAILED);
}

void
aclnttcp_create (const char *host, int port, const rpc_program &rp,
		 aclntalloc_cb cb, axprtalloc_fn xa)
{
  vNew aclnttcpobj (host, port, rp,
		    wrap (aclnttcp_create_finish, &rp, cb, xa));
}
void
aclnttcp_create (const in_addr &addr, int port, const rpc_program &rp,
		 aclntalloc_cb cb, axprtalloc_fn xa)
{
  vNew aclnttcpobj (addr, port, rp,
		    wrap (aclnttcp_create_finish, &rp, cb, xa));
}

static sockaddr_in pmapaddr;
static vec<mapping> pmap_mappings;

static void
pmap_map_3 (callback<void, bool>::ptr cb, ref<bool> resp, size_t mpos,
	     clnt_stat stat)
{
  if (stat) {
    warn << "portmap: " << stat << "\n";
    if (cb)
      (*cb) (false);
    return;
  }
  if (cb)
    (*cb) (*resp);
}

static void
pmap_map_2 (callback<void, bool>::ptr cb, size_t mpos, clnt_stat stat)
{
  if (stat) {
    warn << "portmap: " << stat << "\n";
    if (cb)
      (*cb) (false);
    return;
  }
  ref<bool> resp (New refcounted<bool> (false));
  acallrpc (&pmapaddr, pmap_prog_2, PMAPPROC_SET, &pmap_mappings[mpos], resp,
	    wrap (pmap_map_3, cb, resp, mpos));
}

static void
pmap_map_1 (callback<void, bool>::ptr cb, size_t mpos, ref<u_int32_t> portp,
	     clnt_stat stat)
{
  if (stat) {
    warn << "portmap: " << stat << "\n";
    if (cb)
      (*cb) (false);
    return;
  }

  if (*portp) {
    static bool garbage;
    mapping m = pmap_mappings[mpos];
    m.port = *portp;
    acallrpc (&pmapaddr, pmap_prog_2, PMAPPROC_UNSET, &m, &garbage,
	      wrap (pmap_map_2, cb, mpos));
  }
  else
    pmap_map_2 (cb, mpos, RPC_SUCCESS);
}

void
pmap_map (int fd, const rpc_program &rp, callback<void, bool>::ptr cb)
{
  static bool pmapaddr_initted;
  if (!pmapaddr_initted) {
    pmapaddr.sin_family = AF_INET;
    pmapaddr.sin_port = htons (PMAP_PORT);
    pmapaddr.sin_addr.s_addr = htonl (INADDR_LOOPBACK);
  }

  sockaddr_in sin;
  bzero (&sin, sizeof (sin));
  socklen_t sn = sizeof (sin);
  if (getsockname (fd, (sockaddr *) &sin, &sn) < 0
      || sin.sin_family != AF_INET) {
    if (cb)
      (*cb) (false);
    return;
  }

  int n;
  sn = sizeof (n);
  if (getsockopt (fd, SOL_SOCKET, SO_TYPE, (char *) &n, &sn) < 0
      || (n != SOCK_STREAM && n != SOCK_DGRAM)) {
    if (cb)
      (*cb) (false);
    return;
  }

  mapping &m = pmap_mappings.push_back ();
  m.prog = rp.progno;
  m.vers = rp.versno;
  m.prot = n == SOCK_STREAM ? IPPROTO_TCP : IPPROTO_UDP;
  m.port = ntohs (sin.sin_port);

  ref<u_int32_t> resp = New refcounted<u_int32_t> (0);
  acallrpc (&pmapaddr, pmap_prog_2, PMAPPROC_GETPORT, &m, resp,
	    wrap (pmap_map_1, cb, pmap_mappings.size () - 1, resp));
}

EXITFN (pmap_unmapall);

static void
pmap_unmapall ()
{
  for (size_t i = 0; i < pmap_mappings.size (); i++) {
    static bool garbage;
    if (pmap_mappings[i].port)
      acallrpc (&pmapaddr, pmap_prog_2, PMAPPROC_UNSET,
		&pmap_mappings[i], &garbage, aclnt_cb_null);
  }
}
