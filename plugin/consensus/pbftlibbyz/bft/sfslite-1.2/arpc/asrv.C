/* $Id: asrv.C 3714 2008-10-14 15:00:36Z max $ */

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
#include "xdr_suio.h"
#include "rpc_stats.h"

#ifdef MAINTAINER
int asrvtrace (getenv ("ASRV_TRACE") ? atoi (getenv ("ASRV_TRACE")) : 0);
bool asrvtime (getenv ("ASRV_TIME"));
void set_asrvtrace (int l) { asrvtrace = l; }
void set_asrvtime (bool b) { asrvtime = b; }
int get_asrvtrace (void) { return asrvtrace; }
bool get_asrvtime (void) { return asrvtime; }
#else /* !MAINTAINER */
enum { asrvtrace = 0, asrvtime = 0 };
#endif /* !MAINTAINER */

#define trace (traceobj (asrvtrace, "ASRV_TRACE: ", asrvtime))


inline u_int32_t
xidswap (u_int32_t xid)
{
  return htonl (xid);
}

str
sock2str (const struct sockaddr *sp)
{
  static str empty ("");
  if (sp)
    switch (sp->sa_family) {
    case AF_INET:
      {
	const sockaddr_in *sinp = reinterpret_cast<const sockaddr_in *> (sp);
	return strbuf (" in4=%s:%d", inet_ntoa (sinp->sin_addr),
		       ntohs (sinp->sin_port));
      }
    }
  return empty;
}

svccb::svccb ()
  : arg (NULL), aup (NULL), addr (NULL), addrlen (0),
    resdat (NULL), res (NULL), reslen (0)
{
  bzero (&msg, sizeof (msg));
}

svccb::~svccb ()
{
  xdr_free (reinterpret_cast<sfs::xdrproc_t> (xdr_callmsg), &msg);
  if (arg)
    xdr_delete (srv->tbl[proc ()].xdr_arg, arg);
  if (resdat)
    xdr_delete (srv->tbl[proc ()].xdr_res, resdat);
  if (aup)
    xdr_delete (reinterpret_cast<sfs::xdrproc_t> (xdr_authunix_parms), aup);
  if (srv)
    srv->xi->svcdel ();
  xfree (res);
  delete addr;
}

bool
svccb::operator== (const svccb &a) const
{
  return (xid () == a.xid () && prog () == a.prog ()
	  && vers () == a.vers () && proc () == a.proc ()
	  && addrlen == a.addrlen && addreq (addr, a.addr, addrlen));
}

u_int
svccb::hash_value () const
{
  return xid () ^ hash_bytes (addr, addrlen);
}

void
svccb::init (asrv *s, const sockaddr *src)
{
  srv = mkref (s);
  srv->xi->svcadd ();
  if (!s->xi->xh->connected) {
    addrlen = s->xi->xh->socksize;
    addr = (sockaddr *) opnew (addrlen);
    memcpy (addr, src, addrlen);
  }

  // keep track of when this RPC started
  ts_start = sfs_get_tsnow();
}

void *
svccb::getvoidres ()
{
  if (!resdat)
    resdat = srv->tbl[proc ()].alloc_res ();
  return resdat;
}

const authunix_parms *
svccb::getaup () const
{
  if (aup || msg.rm_call.cb_cred.oa_flavor != AUTH_UNIX)
    return aup;
  xdrmem x (msg.rm_call.cb_cred.oa_base,
	    msg.rm_call.cb_cred.oa_length, XDR_DECODE);
  aup = New authunix_parms;
  bzero (aup, sizeof (*aup));
  if (xdr_authunix_parms (x.xdrp (), aup))
    return aup;
  xdr_free (reinterpret_cast<sfs::xdrproc_t> (xdr_authunix_parms), aup);
  delete aup;
  // msg.rm_call.cb_cred.oa_flavor = AUTH_NONE;
  return aup = NULL;
}

u_int32_t
svccb::getaui () const
{
  if (msg.rm_call.cb_cred.oa_flavor != AUTH_UINT
      || msg.rm_call.cb_cred.oa_length != 4)
    return 0;
  return ntohl (*(u_int32_t *) msg.rm_call.cb_cred.oa_base);
}

bool
svccb::fromresvport () const
{
  const sockaddr_in *sinp = reinterpret_cast<const sockaddr_in *> (getsa ());
  return (sinp && sinp->sin_family == AF_INET
	  && ntohs (sinp->sin_port) < IPPORT_RESERVED);
}

void
svccb::reply (const void *reply, sfs::xdrproc_t xdr, bool nocache)
{
  rpc_msg rm;

  rm.rm_xid = xid ();
  rm.rm_direction = REPLY;
  rm.rm_reply.rp_stat = MSG_ACCEPTED;
  rm.acpted_rply.ar_verf = _null_auth;
  rm.acpted_rply.ar_stat = SUCCESS;
  rm.acpted_rply.ar_results.where = (char *) reply;
  rm.acpted_rply.ar_results.proc
    = reinterpret_cast<sun_xdrproc_t> (xdr ? xdr : srv->tbl[proc ()].xdr_res);

  get_rpc_stats ().end_call (this, ts_start);

  xdrsuio x (XDR_ENCODE);
  if (!xdr_replymsg (x.xdrp (), &rm)) {
    warn ("svccb::reply: xdr_replymsg failed\n");
    delete this;
    return;
  }

  trace (4, "reply %s:%s x=%x\n",
	 srv->rpcprog->name, srv->tbl[msg.rm_call.cb_proc].name,
	 xidswap (msg.rm_xid));
  if (asrvtrace >= 5 && !xdr && srv->tbl[msg.rm_call.cb_proc].print_res)
    srv->tbl[msg.rm_call.cb_proc].print_res (reply, NULL, asrvtrace - 4,
					     "REPLY", "");
  srv->sendreply (this, &x, nocache);
}

static void
asrv_rpc_mismatch (ref<xhinfo> xi, const sockaddr *addr, u_int32_t xid)
{
  rpc_msg m;

  bzero (&m, sizeof (m));
  m.rm_xid = xid;
  m.rm_direction = REPLY;
  m.rm_reply.rp_stat = MSG_DENIED;
  m.rjcted_rply.rj_stat = RPC_MISMATCH;
  m.rjcted_rply.rj_vers.low = RPC_MSG_VERSION;
  m.rjcted_rply.rj_vers.high = RPC_MSG_VERSION;

  xdrsuio x (XDR_ENCODE);
  if (xdr_replymsg (x.xdrp (), &m))
    xi->xh->sendv (x.iov (), x.iovcnt (), addr);
  else
    warn ("asrv_rpc_mismatch: xdr_replymsg failed\n");
}

static void
asrv_auth_reject (ref<xhinfo> xi, const sockaddr *addr,
		  u_int32_t xid, auth_stat stat)
{
  rpc_msg m;

  bzero (&m, sizeof (m));
  assert (stat != AUTH_OK);
  m.rm_xid = xid;
  m.rm_direction = REPLY;
  m.rm_reply.rp_stat = MSG_DENIED;
  m.rjcted_rply.rj_stat = AUTH_ERROR;
  m.rjcted_rply.rj_why = stat;

  xdrsuio x (XDR_ENCODE);
  if (xdr_replymsg (x.xdrp (), &m))
    xi->xh->sendv (x.iov (), x.iovcnt (), addr);
  else
    warn ("asrv_auth_reject: xdr_replymsg failed\n");
}

static void
asrv_accepterr (ref<xhinfo> xi, const sockaddr *addr,
		accept_stat stat, const rpc_msg *mp)
{
  rpc_msg m;

  bzero (&m, sizeof (m));
  m.rm_xid = mp->rm_xid;
  m.rm_direction = REPLY;
  m.rm_reply.rp_stat = MSG_ACCEPTED;

  switch (stat) {
  case PROG_UNAVAIL:
  case PROG_MISMATCH:
    {
      m.acpted_rply.ar_stat = PROG_UNAVAIL;
      m.acpted_rply.ar_vers.low = 0xffffffff;
      m.acpted_rply.ar_vers.high = 0;

      u_int32_t prog = mp->rm_call.cb_prog;
      u_int32_t vers = mp->rm_call.cb_vers;
      for (asrv *a = xi->stab.first (); a; a = xi->stab.next (a))
	if (a->hascb () && a->pv.prog == prog) {
	  if (a->pv.vers == vers)
	    panic ("asrv_accepterr: prog/vers exists\n");
	    // m.acpted_rply.ar_stat = PROC_UNAVAIL;
	  else if (m.acpted_rply.ar_stat != PROC_UNAVAIL) {
	    m.acpted_rply.ar_stat = PROG_MISMATCH;
	    if (m.acpted_rply.ar_vers.low > a->pv.vers)
	      m.acpted_rply.ar_vers.low = a->pv.vers;
	    if (m.acpted_rply.ar_vers.high < a->pv.vers)
	      m.acpted_rply.ar_vers.high = a->pv.vers;
	  }
	}
      break;
    }
  case PROC_UNAVAIL:
  case GARBAGE_ARGS:
  case SYSTEM_ERR:
    m.acpted_rply.ar_stat = stat;
    break;
  default:
    panic ("asrv_accepterr: bad stat %d\n", stat);
    break;
  }

  xdrsuio x (XDR_ENCODE);
  if (xdr_replymsg (x.xdrp (), &m))
    xi->xh->sendv (x.iov (), x.iovcnt (), addr);
  else
    warn ("asrv_accepterr: xdr_replymsg failed\n");
}

void
svccb::reject (auth_stat stat)
{
  trace (3, "reject (auth_stat %d) %s:%s x=%x\n",
	 stat, srv->rpcprog->name, srv->tbl[msg.rm_call.cb_proc].name,
	 xidswap (msg.rm_xid));

  if (!srv->xi->ateof ())
    asrv_auth_reject (srv->xi, addr, xid (), stat);
  srv->sendreply (this, NULL, true);
}

void
svccb::reject (accept_stat stat)
{
  trace (3, "reject (accept_stat %d) %s:%s x=%x\n", stat,
      	 srv->rpcprog->name, srv->tbl[msg.rm_call.cb_proc].name,
	 xidswap (msg.rm_xid));

  if (!srv->xi->ateof ())
    asrv_accepterr (srv->xi, addr, stat, &msg);
  srv->sendreply (this, NULL, true);
}

void
svccb::ignore ()
{
  // Drop a request on the floor
  srv->sendreply (this, NULL, true);
}

asrv::asrv (ref<xhinfo> xi, const rpc_program &pr, asrv_cb::ptr cb)
  : rpcprog (&pr), tbl (pr.tbl), nproc (pr.nproc), cb (cb), recv_hook (NULL),
    xi (xi), pv (pr.progno, pr.versno)
{
  start ();
}

void
asrv::start ()
{
  if (xi->stab[progvers (rpcprog->progno, rpcprog->versno)])
    panic ("attempt to reregister %s on same transport\n", rpcprog->name);
  xi->stab.insert (this);
}

void
asrv::stop ()
{
  if (xi->stab [progvers (rpcprog->progno, rpcprog->versno)] == this)
    xi->stab.remove (this);
}

asrv::~asrv ()
{
  stop ();
}

const ref<axprt> &
asrv::xprt () const
{
  return xi->xh;
}

ptr<asrv>
asrv::alloc (ref<axprt> x, const rpc_program &pr, asrv_cb::ptr cb)
{
  ptr<xhinfo> xi = xhinfo::lookup (x);
  if (!xi)
    return NULL;
  if (x->reliable)
    return New refcounted<asrv> (xi, pr, cb);
  else
    return New refcounted<asrv_unreliable> (xi, pr, cb);
}

void
asrv::seteof (ref<xhinfo> xi, const sockaddr *src, bool force)
{
  if (force || xi->xh->connected) {
    asrv *s;
    ptr<asrv> sp;
    for (s = xi->stab.first (); s; s = xi->stab.next (s)) {
      sp = mkref (s);
      if (s->cb)
	(*s->cb) (NULL);
    }
  }
}

void
asrv::sendreply (svccb *sbp, xdrsuio *x, bool)
{
  if (!xi->ateof () && x)
    xi->xh->sendv (x->iov (), x->iovcnt (), sbp->addr);
  /* If x contains a marshaled version of sbp->template getres<...> (),
   * we need to clear the uio first (since deleting sbp will delete
   * the object getres returned). */
  if (sbp->resdat)
    xsuio (x)->clear ();
  dec_svccb_count ();
  delete sbp;
}

void
asrv::setcb (asrv_cb::ptr c)
{
  cb = c;
  if (cb && xi->ateof ())
    (*cb) (NULL);
}

void
asrv::dispatch (ref<xhinfo> xi, const char *msg, ssize_t len,
		const sockaddr *src)
{
  if (!msg || len < 8 || getint (msg + 4) != CALL) {
    seteof (xi, src, len < 0);
    return;
  }
 
  xdrmem x (msg, len, XDR_DECODE);
  auto_ptr<svccb> sbp (New svccb);
  rpc_msg *m = &sbp->msg;

  if (!xdr_callmsg (x.xdrp (), m)) {
    trace (1) << "asrv::dispatch: xdr_callmsg failed\n";
    seteof (xi, src);
    return;
  }
  if (m->rm_call.cb_rpcvers != RPC_MSG_VERSION) {
    trace (1) << "asrv::dispatch: bad RPC message version\n";
    asrv_rpc_mismatch (xi, src, m->rm_xid);
    return;
  }

  asrv *s = xi->stab[progvers (sbp->prog (), sbp->vers ())];
  if (!s || !s->cb) {
    if (asrvtrace >= 1) {
      if (s)
	warn ("asrv::dispatch: no callback for %s (proc = %u)\n",
	      s->rpcprog->name, sbp->proc ());
      else
	warn ("asrv::dispatch: invalid prog/vers %u/%u (proc = %u)\n",
	      (u_int) sbp->prog (), (u_int) sbp->vers (),
	      (u_int) sbp->proc ());
    }
    asrv_accepterr (xi, src, PROG_UNAVAIL, m);
    return;
  }

  if (s->recv_hook)
    s->recv_hook ();

  sbp->init (s, src);

  if (sbp->proc () >= s->nproc) {
    if (asrvtrace >= 1)
      warn ("asrv::dispatch: invalid procno %s:%u\n",
	    s->rpcprog->name, (u_int) sbp->proc ());
    asrv_accepterr (xi, src, PROC_UNAVAIL, m);
    return;
  }

  if (s->isreplay (sbp.get ())) {
    trace (4, "replay %s:%s x=%x",
           s->rpcprog->name, s->tbl[m->rm_call.cb_proc].name,
           xidswap (m->rm_xid)) << sock2str (src) << "\n";
    return;
  }

  const rpcgen_table *rtp = &s->tbl[sbp->proc ()];
  sbp->arg = s->tbl[sbp->proc ()].alloc_arg ();
  if (!rtp->xdr_arg (x.xdrp (), sbp->arg)) {
    if (asrvtrace >= 1)
      warn ("asrv::dispatch: bad message %s:%s x=%x", s->rpcprog->name,
	    rtp->name, xidswap (m->rm_xid))
	      << sock2str (src) << "\n";
    asrv_accepterr (xi, src, GARBAGE_ARGS, m);
    s->inc_svccb_count ();
    s->sendreply (sbp.release (), NULL, true);
    return;
  }

  if (asrvtrace >= 2) {
    if (const authunix_parms *aup = sbp->getaup ())
      trace (2, "serve %s:%s x=%x u=%u g=%u",
	     s->rpcprog->name, rtp->name, xidswap (m->rm_xid),
             aup->aup_uid, aup->aup_gid)
	       << sock2str (src) << "\n";
    else if (u_int32_t i = sbp->getaui ())
      trace (2, "serve %s:%s x=%x i=%u",
	     s->rpcprog->name, rtp->name, xidswap (m->rm_xid), i)
	       << sock2str (src) << "\n";
    else
      trace (2, "serve %s:%s x=%x",
	     s->rpcprog->name, rtp->name, xidswap (m->rm_xid))
	       << sock2str (src) << "\n";
  }
  if (asrvtrace >= 5 && rtp->print_arg)
    rtp->print_arg (sbp->arg, NULL, asrvtrace - 4, "ARGS", "");

  s->inc_svccb_count ();
  (*s->cb) (sbp.release ());
}


/* asrv_replay */

svccb *
asrv_replay::lookup (svccb *sbp)
{
  svccb *osbp = rtab[*sbp];
  if (!osbp)
    rtab.insert (sbp);
  return osbp;
}

void
asrv_replay::delsbp (svccb *sbp)
{
  // trace (4, "EVICT x=%x o=%lu\n", xidswap (sbp->xid ()),
  //                                (unsigned long) sbp->offset);
  rtab.remove (sbp);
  rq.remove (sbp);
  delete sbp;
}

asrv_replay::~asrv_replay ()
{
  rtab.traverse (wrap (this, &asrv_replay::delsbp));
}

void
asrv_replay::sendreply (svccb *sbp, xdrsuio *x, bool nocache)
{
  
  if (!x) {
    rtab.remove (sbp);
    delete sbp;
    return;
  }

  if (sbp->arg) {
    xdr_delete (tbl[sbp->proc ()].xdr_arg, sbp->arg);
    sbp->arg = NULL;
  }

  sbp->reslen = x->uio ()->resid ();
  sbp->res = suio_flatten (x->uio ());
  x->uio ()->clear ();
  if (!xi->ateof ())
    xi->xh->send (sbp->res, sbp->reslen, sbp->addr);
  if (sbp->resdat) {
    xdr_delete (tbl[sbp->proc ()].xdr_res, sbp->resdat);
    sbp->resdat = NULL;
  }

  if (nocache) {
    rtab.remove (sbp);
    delete sbp;
    return;
  }
}


/* asrv_unreliable */

bool
asrv_unreliable::isreplay (svccb *sbp)
{
  svccb *osbp = lookup (sbp);
  if (!osbp)
    return false;

  if (osbp->res) {
    trace (4, "reply to replay x=%x\n", xidswap (osbp->xid ()));
    xi->xh->send (osbp->res, osbp->reslen, osbp->addr);
  }
  // else still waiting for sendreply

  return true;
}

void
asrv_unreliable::sendreply (svccb *sbp, xdrsuio *x, bool nocache)
{
  asrv_replay::sendreply (sbp, x, nocache);
  if (!x || nocache)
    return;

  ref<asrv> hold = sbp->srv;	// Don't let this be freed
  sbp->srv = NULL;		// Decrement reference count on this
  rsize++;
  rq.insert_tail (sbp);

  while (rsize > maxrsize) {
    delsbp (rq.first);
    rsize--;
  }
}


/* asrv_resumable */

/* We keep track of the byte offset of each reply so that cached
 * replies can be evicted when we known they have been received by the
 * client.
 *
 * When we resume, all calls cached in the reply queue (rq) are given
 * a zero byte offset.  If such a call is replayed by the client, it
 * is moved to the end of the reply queue and given the byte offset of
 * its most recent reply message.  But when we receive the first
 * non-replay call after resumption, we assume all calls left over
 * from the previous connection have been dealt with, and evict them
 * from the reply queue.
 */

bool
asrv_resumable::isreplay (svccb *sbp)
{
  svccb *osbp = lookup (sbp);
  if (!osbp) {
    // clear out any calls from the last connection
    while ((sbp = rq.first) && !sbp->offset)
      delsbp (sbp);
    return false;
  }

  if (osbp->res) {
    xi->xh->send (osbp->res, osbp->reslen, osbp->addr);
    osbp->offset = xi->xh->get_raw_bytes_sent ();
    rq.remove (osbp);
    rq.insert_tail (osbp);
  }
  // else still waiting for sendreply

  return true;
}

void
asrv_resumable::sendreply (svccb *sbp, xdrsuio *x, bool nocache)
{
  assert (!(x && nocache));
  asrv_replay::sendreply (sbp, x, nocache);
  if (!x)
    return;

  sbp->offset = xi->xh->get_raw_bytes_sent ();

  ref<asrv> hold = sbp->srv;	// Don't let this be freed
  sbp->srv = NULL;		// Decrement reference count on this
  rq.insert_tail (sbp);

  u_int64_t bytes_sent = xi->xh->get_raw_bytes_sent ();
  int sndbufsz = xi->xh->sndbufsize ();
  u_int64_t known_received = sndbufsz > 0
                             ? (bytes_sent > (unsigned) sndbufsz
                                ? bytes_sent - sndbufsz : 0)
                             : 0;
  known_received = max (known_received, xi->max_acked_offset);
  while ((sbp = rq.first) && sbp->offset && sbp->offset < known_received)
    delsbp (sbp);
}

bool
asrv_resumable::resume (ref<axprt> newxprt)
{
  if (!newxprt->reliable)
    panic ("resumable asrv on unreliable transport: unimplemented\n");
  ptr<xhinfo> newxi = xhinfo::lookup (newxprt);
  if (!newxi)
    return false;

  stop ();
  xi = newxi;
  start ();

  svccb *sbp;
  for (sbp = rtab.first (); (sbp); sbp = rtab.next (sbp)) {
    sbp->offset = 0;
    xi->svcadd ();
  }
  return true;
}

ptr<asrv_resumable>
asrv_resumable::alloc (ref<axprt> x, const rpc_program &pr, asrv_cb::ptr cb)
{
  ptr<xhinfo> xi = xhinfo::lookup (x);
  if (!xi)
    return NULL;
  if (!x->reliable)
    panic ("resumable asrv on unreliable transport unimplemented\n");
  return New refcounted<asrv_resumable> (xi, pr, cb);
}

ptr<asrv>
asrv_alloc (ref<axprt> x, const rpc_program &pr,
    	    callback<void, svccb *>::ptr cb, bool resumable)
{
  if (resumable)
    return asrv_resumable::alloc (x, pr, cb);
  else
    return asrv::alloc (x, pr, cb);
}

asrv_delayed_eof::asrv_delayed_eof (ref<xhinfo> xi, const rpc_program &pr, 
				    asrv_cb::ptr scb, cbv::ptr eofcb)
  : asrv (xi, pr, NULL),
    _count (0), 
    _eof (false), 
    _eofcb (eofcb) 
{
  asrv_delayed_eof::setcb (scb);
}

void
asrv_delayed_eof::setcb (asrv_cb::ptr cb)
{
  bool isset = _asrv_cb;
  _asrv_cb = cb;
  if (cb && !isset) {
    this->asrv::setcb (wrap (this, &asrv_delayed_eof::dispatch));
  } else if (!cb && isset) {
    this->asrv::setcb (NULL);
  }
}



ptr<asrv_delayed_eof>
asrv_delayed_eof::alloc (ref<axprt> x, const rpc_program &pr, 
			 asrv_cb::ptr cb, cbv::ptr eofcb)
{
  ptr<xhinfo> xi = xhinfo::lookup (x);
  if (!xi || !x->reliable)
    return NULL;
  return New refcounted<asrv_delayed_eof> (xi, pr, cb, eofcb);
}

void
asrv_delayed_eof::dispatch (svccb *sbp)
{
  if (sbp == NULL) {
    _eof = true;
    cbv::ptr c(_eofcb);
    _eofcb = NULL;
    if (_count == 0) {
      (*_asrv_cb) (NULL);
    } else if (c) {
      (*c) ();
    }
  } else {
    (*_asrv_cb) (sbp);
  }
}

void
asrv_delayed_eof::dec_svccb_count ()
{
  assert (--_count >= 0);
  if (_count == 0 && _eof) {
    (*_asrv_cb) (NULL);
  }
}

void 
asrv_delayed_eof::sendreply (svccb *s, xdrsuio *x, bool nocache)
{
  if (_eof) {
    warn << "Swallowing RPC reply due to EOF on TCP socket.\n";
    dec_svccb_count ();
  } else if (xprt ()->getwritefd () < 0) {
    warn << "Swallowing RPC reply due to unexpected EOF/error on socket.\n";
    dec_svccb_count ();
  } else {
    // decref already happens in sendreply(), so no need to do it a 
    // second time
    asrv::sendreply (s, x, nocache);
  }
}
