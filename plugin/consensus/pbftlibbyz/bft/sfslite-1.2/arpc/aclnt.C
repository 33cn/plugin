/* $Id: aclnt.C 2508 2007-01-12 23:39:52Z yipal $ */

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
#include "list.h"
#include "backoff.h"
#include "xdr_suio.h"

#ifdef MAINTAINER
int aclnttrace (getenv ("ACLNT_TRACE")
		? atoi (getenv ("ACLNT_TRACE")) : 0);
bool aclnttime (getenv ("ACLNT_TIME"));
#else /* !MAINTAINER */
enum { aclnttrace = 0, aclnttime = 0 };
#endif /* !MAINTAINER */

#define trace (traceobj (aclnttrace, "ACLNT_TRACE: ", aclnttime))


AUTH *auth_none;
static tmoq<rpccb_unreliable, &rpccb_unreliable::tlink> rpctoq;

static void
ignore_clnt_stat (clnt_stat)
{
}

aclnt_cb aclnt_cb_null (gwrap (ignore_clnt_stat));

INITFN (aclnt_init);

static void
aclnt_init ()
{
  auth_none = authnone_create ();
}

callbase::callbase (ref<aclnt> c, u_int32_t xid, const sockaddr *d)
  : c (c), dest (d), tmo (NULL), xid (xid), offset (0)
{
  c->calls.insert_tail (this);
  c->xi->xidtab.insert (this);
}

callbase::~callbase ()
{
  c->calls.remove (this);
  if (tmo)
    timecb_remove (tmo);
  if (c->xi->xidtab[xid] == this)
    c->xi->xidtab.remove (this);
  tmo = reinterpret_cast<timecb_t *> (0xc5c5c5c5); // XXX - debugging
}

bool
callbase::checksrc (const sockaddr *src) const
{
  if (c->xi->xh->connected)
    return true;
  return addreq (src, dest, c->xi->xh->socksize);
}

void
callbase::timeout (time_t sec, long nsec)
{
  assert (!tmo);
  if (timecb_t *t = delaycb (sec, nsec, wrap (this, &callbase::expire)))
    tmo = t;
}

u_int32_t (*next_xid) () = arandom;
static u_int32_t
genxid (xhinfo *xi)
{
  u_int32_t xid;
  while (xi->xidtab[xid = (*next_xid) ()] || !xid)
    ;
  return xid;
}

rpccb::rpccb (ref<aclnt> c, u_int32_t xid, aclnt_cb cb,
	      void *out, sfs::xdrproc_t outproc, const sockaddr *d)
  : callbase (c, xid, d), cb (cb), outmem (out), outxdr (outproc)
{
}

rpccb::rpccb (ref<aclnt> c, xdrsuio &x, aclnt_cb cb,
	      void *out, sfs::xdrproc_t outproc, const sockaddr *d)
  : callbase (c, getxid (c, x), d), cb (cb), outmem (out), outxdr (outproc)
{
}

callbase *
rpccb::init (xdrsuio &x)
{
  ref<aclnt> cc (c);
  cc->xprt ()->sendv (x.iov (), x.iovcnt (), dest);
  if (cc->xi_xh_ateof_fail ())
    return NULL;
  offset = c->xprt ()->get_raw_bytes_sent ();
  return this;
#if 0
  if (tmo && !c->xprt ()->reliable) {
    panic ("transport callback added timeout (%p)\n"
	   "   xprt type: %s\n  rpccb type: %s\n\n"
	   "*** PLEASE REPORT THIS PROBLEM TO sfs-dev@pdos.lcs.mit.edu ***\n",
	   tmo, typeid (*c->xprt ()).name (),
	   typeid (*this).name ());
  }
#endif
}

void
rpccb::finish (clnt_stat stat)
{
  aclnt_cb c (cb);
  delete this;
  (*c) (stat);
}

u_int32_t
rpccb::getxid (ref<aclnt> c, xdrsuio &x)
{
  assert (x.iovcnt () > 0);
  assert (x.iov ()[0].iov_len >= 4);
  u_int32_t *xidp = reinterpret_cast<u_int32_t *> (x.iov ()[0].iov_base);
  if (!*xidp)
    *xidp = genxid (c->xi);
  return *xidp;
}

u_int32_t
rpccb::getxid (ref<aclnt> c, char *buf, size_t len)
{
  assert (len >= 4);
  u_int32_t *xidp = reinterpret_cast<u_int32_t *> (buf);
  if (!*xidp)
    *xidp = genxid (c->xi);
  return *xidp;
}

clnt_stat
rpccb::decodemsg (const char *msg, size_t len)
{
  char *m = const_cast<char *> (msg);

  xdrmem x (m, len, XDR_DECODE);
  rpc_msg rm;
  bzero (&rm, sizeof (rm));
  rm.acpted_rply.ar_verf = _null_auth; 
  rm.acpted_rply.ar_results.where = (char *) outmem;
  rm.acpted_rply.ar_results.proc = reinterpret_cast<sun_xdrproc_t> (outxdr);
  bool ok = xdr_replymsg (x.xdrp (), &rm);

  /* We don't support any auths with meaningful reply verfs */
  if (rm.rm_direction == REPLY && rm.rm_reply.rp_stat == MSG_ACCEPTED
      && rm.acpted_rply.ar_verf.oa_base)
    free (rm.acpted_rply.ar_verf.oa_base);

  if (!ok)
    return RPC_CANTDECODERES;
  rpc_err re;
  seterr_reply (&rm, &re);
  return re.re_status;
}

void
rpccb_msgbuf::xmit (int retry)
{
  if (!c->xi->ateof ()) {
    if (retry > 0)
      trace (2, "retransmit #%d x=%x\n", retry,
                *reinterpret_cast<u_int32_t *> (msgbuf));
    c->xprt ()->send (msgbuf, msglen, dest);
  }
}

rpccb_unreliable::rpccb_unreliable (ref<aclnt> c, xdrsuio &x,
				    aclnt_cb cb,
				    void *out, sfs::xdrproc_t outproc,
				    const sockaddr *d)
  : rpccb_msgbuf (c, x, cb, out, outproc, d)
{
}

rpccb_unreliable::rpccb_unreliable (ref<aclnt> c,
				    char *buf, size_t len,
				    aclnt_cb cb,
				    void *out, sfs::xdrproc_t outproc,
				    const sockaddr *d)
  : rpccb_msgbuf (c, buf, len, cb, out, outproc, d)
{
}

rpccb_unreliable::~rpccb_unreliable ()
{
  rpctoq.remove (this);
}

callbase *
rpccb_unreliable::init (xdrsuio &x)
{
  assert (!tmo);
  rpctoq.start (this);
  assert (!tmo);
  return this;
}

aclnt::aclnt (const ref<xhinfo> &x, const rpc_program &p)
  : xi (x), rp (p), eofcb (NULL), dest (NULL), stopped (true),
    send_hook (NULL), recv_hook (NULL)
{
  start ();
}

aclnt::~aclnt ()
{
  assert (!calls.first);
  stop ();
  if (dest)
    xfree (dest);
}

void
aclnt::start ()
{
  if (stopped) {
    stopped = false;

    xi->clist.insert_head (this);

    rpccb_msgbuf *rb;
    for (rb = static_cast<rpccb_msgbuf *> (calls.first); rb;
         rb = static_cast<rpccb_msgbuf *> (calls.next (rb))) {
      assert (!xi->xidtab[rb->xid]);
      xi->xidtab.insert (rb);
    }
  }
}

void
aclnt::stop ()
{
  if (!stopped) {
    stopped = true;

    aclnt *XXX_gcc296_bug __attribute__ ((unused)) = xi->clist.remove (this);
    // also needed for gcc 3.0.2 on RedHat 7.3

    rpccb_msgbuf *rb;
    for (rb = static_cast<rpccb_msgbuf *> (calls.first); rb;
         rb = static_cast<rpccb_msgbuf *> (calls.next (rb))) {
      assert (xi->xidtab[rb->xid] == rb);
      xi->xidtab.remove (rb);
    }
  }
}

const ref<axprt> &
aclnt::xprt () const
{
  return xi->xh;
}

bool
aclnt::marshal_call (xdrsuio &x, AUTH *auth,
		     u_int32_t progno, u_int32_t versno, u_int32_t procno,
		     sfs::xdrproc_t inproc, const void *in)
{
  u_int32_t *dp = (u_int32_t *) XDR_INLINE (x.xdrp (), 6*4);
#if 0
  if (dp) {
#endif
    *dp++ = 0;
    *dp++ = htonl (CALL);
    *dp++ = htonl (RPC_MSG_VERSION);
    *dp++ = htonl (progno);
    *dp++ = htonl (versno);
    *dp++ = htonl (procno);
#if 0
  }
  else {
    xdr_putint (x.xdrp (), 0);
    xdr_putint (x.xdrp (), CALL);
    xdr_putint (x.xdrp (), RPC_MSG_VERSION);
    xdr_putint (x.xdrp (), progno);
    xdr_putint (x.xdrp (), versno);
    xdr_putint (x.xdrp (), procno);
  }
#endif
  if (!AUTH_MARSHALL (auth ? auth : auth_none, x.xdrp ())) {
    warn ("failed to marshal auth crap\n");
    return false;
  }
  if (!inproc (x.xdrp (), const_cast<void *> (in))) {
    warn ("arg marshaling failed (prog %d, vers %d, proc %d)\n",
	  progno, versno, procno);
    return false;
  }
  return true;
}

static void
printreply (aclnt_cb cb, str name, void *res,
	    void (*print_res) (const void *, const strbuf *, int,
			       const char *, const char *),
	    clnt_stat err)
{
  if (err)
    trace (3) << "reply " << name << ": " << err << "\n";
  else {
    trace (4) << "reply " << name << "\n";
    if (aclnttrace >= 5 && print_res)
      print_res (res, NULL, aclnttrace - 4, "REPLY", "");
  }
  (*cb) (err);
}

bool
aclnt::init_call (xdrsuio &x,
	          u_int32_t procno, const void *in, void *out,
	          aclnt_cb &cb, AUTH *auth,
	          sfs::xdrproc_t inproc, sfs::xdrproc_t outproc,
	          u_int32_t progno, u_int32_t versno)
{
  if (xi_ateof_fail ()) {
    (*cb) (RPC_CANTSEND);
    return false;
  }
  if (!auth)
    auth = auth_none;
  if (!progno) {
    progno = rp.progno;
    assert (procno < rp.nproc);
    if (!inproc)
      inproc = rp.tbl[procno].xdr_arg;
    if (!outproc)
      outproc = rp.tbl[procno].xdr_res;
    if (!versno)
      versno = rp.versno;
  }
  assert (inproc);
  assert (outproc);
  assert (progno);
  assert (versno);

  if (!marshal_call (x, auth, progno, versno, procno, inproc, in)) {
    (*cb) (RPC_CANTENCODEARGS);
    return false;
  }
  assert (x.iov ()[0].iov_len >= 4);
  u_int32_t &xid = *reinterpret_cast<u_int32_t *> (x.iov ()[0].iov_base);
  if (!forget_call (cb))
    xid = genxid (xi);

  if (aclnttrace >= 2) {
    str name;
    const rpcgen_table *rtp;
    if (progno == rp.progno && versno == rp.versno && procno < rp.nproc) {
      rtp = &rp.tbl[procno];
      name = strbuf ("%s:%s x=%x", rp.name, rtp->name, xid);
    }
    else {
      rtp = NULL;
      name = strbuf ("prog %d vers %d proc %d x=%x",
		     progno, versno, procno, xid);
    }
    trace () << "call " << name << "\n";
    if (aclnttrace >= 5 && rtp && rtp->xdr_arg == inproc && rtp->print_arg)
      rtp->print_arg (in, NULL, aclnttrace - 4, "ARGS", "");
    if (aclnttrace >= 3 && cb != aclnt_cb_null)
      cb = wrap (printreply, cb, name, out,
		 (rtp && rtp->xdr_res == outproc ? rtp->print_res : NULL));
  }

  return true;
}

callbase *
aclnt::call (u_int32_t procno, const void *in, void *out,
	     aclnt_cb cb,
	     AUTH *auth,
	     sfs::xdrproc_t inproc, sfs::xdrproc_t outproc,
	     u_int32_t progno, u_int32_t versno,
	     sockaddr *d)
{
  xdrsuio x (XDR_ENCODE);
  if (!init_call (x, procno, in, out, cb, auth, inproc,
		  outproc, progno, versno))
    return NULL;
  if (!outproc)
    outproc = rp.tbl[procno].xdr_res;
  if (!d)
    d = dest;

  if (send_hook)
    (*send_hook) ();

  if (forget_call (cb)) {
    if (!xi->ateof ())
      xi->xh->sendv (x.iov (), x.iovcnt (), d);
    return NULL;
  }
  else
    return (*rpccb_alloc) (mkref (this), x, cb, out, outproc, d);
}

bool aclnt::xi_ateof_fail () { return xi->ateof (); }
bool aclnt::xi_xh_ateof_fail () { return xi->xh->ateof (); }

bool
aclnt::forget_call (aclnt_cb cb)
{
  /* If we don't care about the reply, then don't bother keeping
   * state around.  We use the reserved XID 0 for these calls, so
   * that we don't accidentally recycle the XID of a call whose
   * state we threw away. */
  return xi->xh->reliable && cb == aclnt_cb_null;
}

callbase *
aclnt::timedcall (time_t sec, long nsec,
		  u_int32_t procno, const void *in, void *out,
		  aclnt_cb cb,
		  AUTH *auth,
		  sfs::xdrproc_t inproc, sfs::xdrproc_t outproc,
		  u_int32_t progno, u_int32_t versno,
		  sockaddr *d)
{
  callbase *cbase = call (procno, in, out, cb, auth, inproc,
			  outproc, progno, versno, d);
  if (cbase)
    cbase->timeout (sec, nsec);
  return cbase;
}

static void
scall_cb (clnt_stat *errp, bool *donep, clnt_stat err)
{
  *errp = err;
  *donep = true;
}

clnt_stat
aclnt::scall (u_int32_t procno, const void *in, void *out,
	      AUTH *auth,
	      sfs::xdrproc_t inproc, sfs::xdrproc_t outproc,
	      u_int32_t progno, u_int32_t versno,
	      sockaddr *d, time_t duration)
{
  bool done = false;
  clnt_stat err;
  callbase *cbase = call (procno, in, out, wrap (scall_cb, &err, &done),
			  auth, inproc, outproc, progno, versno, d);
  if (cbase && duration)
    cbase->timeout (duration);
  while (!done)
    xprt ()->poll ();
  return err;
}

class rawcall : public callbase {
  aclntraw_cb::ptr cb;
  u_int32_t oldxid;

  rawcall ();

  PRIVDEST virtual ~rawcall () {}

public:
  rawcall (ref<aclnt> c, const char *msg, size_t len,
	   aclntraw_cb cb, sockaddr *d)
    : callbase (c, genxid (c->xi), d), cb (cb) {
    assert (len >= 4);
    assert (c->xprt ()->reliable);
    memcpy (&oldxid, msg, 4);
    iovec iov[2] = {
      { iovbase_t (&xid), 4 },
      { iovbase_t (msg + 4), len - 4 },
    };
    c->xprt ()->sendv (iov, 2, d);
  }

  clnt_stat decodemsg (const char *msg, size_t len) {
    memcpy (const_cast<char *> (msg), &oldxid, 4);
    (*cb) (RPC_SUCCESS, msg, len);
    cb = NULL;
    return RPC_SUCCESS;
  }
  void finish (clnt_stat stat) {
    if (cb)
      (*cb) (stat, NULL, -1);
    delete this;
  }
};

callbase *
aclnt::rawcall (const char *msg, size_t len, aclntraw_cb cb, sockaddr *dest)
{
  return New ::rawcall (mkref (this), msg, len, cb, dest);
}

void
aclnt::seteofcb (cbv::ptr e)
{
  eofcb = e;
  if (xi->ateof ()) {
    eofcb = NULL;
    if (e)
      (*e) ();
  }
}

inline ptr<aclnt>
aclnt_mkptr (aclnt *c)
{
  if (c)
    return mkref (c);
  else
    return NULL;
}

void
aclnt::seteof (ref<xhinfo> xi)
{
  ptr<aclnt> c;
  if (xi->xh->connected)
    for (c = aclnt_mkptr (xi->clist.first); c;
	 c = aclnt_mkptr (xi->clist.next (c)))
      c->fail ();
}

void
aclnt::fail ()
{
  callbase *rb, *nrb;
  for (rb = calls.first; rb; rb = nrb) {
    nrb = calls.next (rb);
    rb->finish (RPC_CANTRECV);
  }
  if (eofcb)
    (*eofcb) ();
}

void
aclnt::dispatch (ref<xhinfo> xi, const char *msg, ssize_t len,
		 const sockaddr *src)
{
  if (!msg || len < 8 || getint (msg + 4) != REPLY) {
    seteof (xi);
    return;
  }

  u_int32_t xid;
  memcpy (&xid, msg, sizeof (xid));
  callbase *rp = xi->xidtab[xid];
  if (!rp || !rp->checksrc (src)) {
    trace (2, "dropping %s x=%x\n",
	   rp ? "reply with bad source address" : "unrecognized reply", xid);
    return;
  }

  clnt_stat err = rp->decodemsg (msg, len);

  if (!err) {
    if (rp->c->recv_hook)
      (*(rp->c->recv_hook)) ();
    xi->max_acked_offset = max (xi->max_acked_offset, rp->offset);
  }

  if (!err || (err && !rp->c->handle_err (err)))
    rp->finish (err);
}

ptr<aclnt>
aclnt::alloc (ref<axprt> x, const rpc_program &pr, const sockaddr *d,
	      aclnt::rpccb_alloc_t ra)
{
  ptr<xhinfo> xi = xhinfo::lookup (x);
  if (!xi)
    return NULL;
  ref<aclnt> c = New refcounted<aclnt> (xi, pr);
  if (!x->connected && d) {
    c->dest = (sockaddr *) xmalloc (x->socksize);
    memcpy (c->dest, d, x->socksize);
  }
  else
    c->dest = NULL;
  if (ra)
    c->rpccb_alloc = ra;
  else if (xi->xh->reliable)
    c->rpccb_alloc = callbase_alloc<rpccb>;
  else
    c->rpccb_alloc = callbase_alloc<rpccb_unreliable>;
  return c;
}


/* aclnt_resumable */

void
aclnt_resumable::fail ()
{
  ref<aclnt> hold = mkref (this);
  if (!(failcb && (*failcb) ()))    // may be called multiple times
    aclnt::fail ();
}

bool
aclnt_resumable::xi_ateof_fail ()
{
  if (xi->ateof ())
    fail ();
  return false;
}

bool
aclnt_resumable::xi_xh_ateof_fail ()
{
  if (xi->xh->ateof ())
    fail ();
  return false;
}

bool
aclnt_resumable::resume (ref<axprt> newxprt)
{
  if (!pre_resume (newxprt))
    return false;
  post_resume ();
  return true;
}

bool
aclnt_resumable::pre_resume (ref<axprt> newxprt)
{
  assert (newxprt->reliable);
  ptr<xhinfo> newxi = xhinfo::lookup (newxprt);
  if (!newxi)
    return false;

  stop ();
  xi = newxi;
  start ();
  return true;
}

void
aclnt_resumable::post_resume ()
{
  rpccb_msgbuf *rb;
  for (rb = static_cast<rpccb_msgbuf *> (calls.first); rb;
       rb = static_cast<rpccb_msgbuf *> (calls.next (rb))) {
    rb->offset = 0;  // we don't know whether the reply will be to the
                     // original call or to this replay
    rb->xmit (1);
  }
}

ptr<aclnt_resumable>
aclnt_resumable::alloc (ref<axprt> x, const rpc_program &pr,
                        callback<bool>::ref failcb)
{
  assert (x->reliable);
  ptr<xhinfo> xi = xhinfo::lookup (x);
  if (!xi)
    return NULL;
  ref<aclnt_resumable> c = New refcounted<aclnt_resumable> (xi, pr, failcb);
  c->rpccb_alloc = callbase_alloc<rpccb_msgbuf_xmit>;
  return c;
}
