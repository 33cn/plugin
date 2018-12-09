/* $Id: srpc.c 2506 2007-01-12 20:13:01Z yipal $ */

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

#include "sfs-internal.h"

#ifdef HAVE___SETERR_REPLY
#define seterr_reply __seterr_reply
#else /* !HAVE___SETERR_REPLY */
#define seterr_reply _seterr_reply
#endif /* !HAVE___SETERR_REPLY */

#define MAXMSG 0x2400

static inline bool_t
xdr_putint (XDR *xdrs, xdrlong_t val)
{
  return xdr_putlong (xdrs, &val);
}

static inline bool_t
xdr_getint (XDR *xdrs, u_int32_t *valp)
{
  xdrlong_t val;
  if (xdr_getlong (xdrs, &val)) {
    *valp = val;
    return TRUE;
  }
  return FALSE;
}

static bool_t
xdr_auth (XDR *xdrs, struct opaque_auth *oap)
{
  u_int32_t val;
  if (!xdr_getint (xdrs, &val))
    return FALSE;
  oap->oa_flavor = val;
  if (!xdr_getint (xdrs, &val))
    return FALSE;
  oap->oa_length = val;

  if (oap->oa_length) {
    if (oap->oa_length > 400)
      return FALSE;
    xfree (oap->oa_base);
    oap->oa_base = malloc (oap->oa_length);
    if (!oap->oa_base) {
      oap->oa_length = 0;
      return FALSE;
    }
    if (!xdr_opaque (xdrs, oap->oa_base, oap->oa_length))
      return FALSE;
  }
  else if (oap->oa_base) {
    xfree (oap->oa_base);
    oap->oa_base = NULL;
  }
  return TRUE;
}

enum clnt_stat
srpc_marshal_call (XDR *xdrs, u_int32_t xid,
		   u_int32_t progno, u_int32_t versno, u_int32_t procno,
		   AUTH *auth, xdrproc_t inproc, void *in)
{
  u_int32_t *dp = (u_int32_t *) xdr_inline (xdrs, 6*4);
  if (dp) {
    *dp++ = xid;
    *dp++ = htonl (CALL);
    *dp++ = htonl (RPC_MSG_VERSION);
    *dp++ = htonl (progno);
    *dp++ = htonl (versno);
    *dp++ = htonl (procno);
  }
  else if (!xdr_putint (xdrs, xid)
	   || !xdr_putint (xdrs, CALL)
	   || !xdr_putint (xdrs, RPC_MSG_VERSION)
	   || !xdr_putint (xdrs, progno)
	   || !xdr_putint (xdrs, versno)
	   || !xdr_putint (xdrs, procno))
    return RPC_CANTENCODEARGS;

  if (!auth) {
    dp = (u_int32_t *) xdr_inline (xdrs, 4*4);
    if (dp)
      bzero (dp, 4 * 4);
    else if (!xdr_putint (xdrs, 0)
	     || !xdr_putint (xdrs, 0)
	     || !xdr_putint (xdrs, 0)
	     || !xdr_putint (xdrs, 0))
      return RPC_CANTENCODEARGS;
  }
  else if (!auth_marshall (auth, xdrs))
    return RPC_CANTENCODEARGS;

  if (!inproc (xdrs, (void *) in))
    return RPC_CANTENCODEARGS;

  return RPC_SUCCESS;
}

enum clnt_stat
srpc_marshal_reply (XDR *xdrs, struct rpc_msg *m)
{
  u_int32_t buf[3];
  u_int32_t *dp = (u_int32_t *) xdr_inline (xdrs, 3*4);
  struct rpc_err re;

#define GETINT(r)				\
  do {						\
    u_int32_t val;				\
    if (!xdr_getint (xdrs, &val))		\
      return RPC_CANTDECODERES;			\
    (r) = val;					\
  } while (0)

  if (!dp) {
    if (!xdr_getbytes (xdrs, (char *) buf, sizeof (buf)))
      return RPC_CANTDECODERES;
    dp = buf;
  }

  m->rm_xid = dp[0];
  m->rm_direction = ntohl (dp[1]);
  if (m->rm_direction != REPLY)
    return RPC_CANTDECODERES;
  m->rm_reply.rp_stat = ntohl (dp[2]);

  switch (m->rm_reply.rp_stat) {
  case MSG_ACCEPTED:
    if (!xdr_auth (xdrs, &m->acpted_rply.ar_verf))
      return RPC_CANTDECODERES;
    GETINT (m->acpted_rply.ar_stat);
    switch ((int) m->acpted_rply.ar_stat) {
    case SUCCESS:
      if (!m->acpted_rply.ar_results.proc (xdrs,
					   m->acpted_rply.ar_results.where))
	return RPC_CANTDECODERES;
      break;
    case PROG_MISMATCH:
      GETINT (m->acpted_rply.ar_vers.low);
      GETINT (m->acpted_rply.ar_vers.high);
      break;
    }
    break;

  case MSG_DENIED:
    GETINT (m->rjcted_rply.rj_stat);
    switch (m->rjcted_rply.rj_stat) {
    case RPC_MISMATCH:
      GETINT (m->rjcted_rply.rj_vers.low);
      GETINT (m->rjcted_rply.rj_vers.high);
      break;
    case AUTH_ERROR:
      GETINT (m->rjcted_rply.rj_why);
      break;
    default:
      return RPC_CANTDECODERES;
    }
    break;

  default:
    return RPC_CANTDECODERES;
  }

  seterr_reply (m, &re);
  return re.re_status;

#undef GETINT
}

static ssize_t
readpkt (int fd, void *buf, size_t bufsize)
{
  size_t pos = 0;

  do {
    ssize_t n = read (fd, buf + pos, bufsize - pos);
    if (n <= 0) {
      perror ("srpc");
      return -1;
    }
    pos += n;
    if (pos - n < 4 && pos >= 4) {
      u_char *bp = buf;
      u_int32_t msgsize = bp[0] << 24 | bp[1] << 16 | bp[2] << 8 | bp[3];
      if (!(msgsize & 0x80000000))
	return -1;
      msgsize &= 0x7fffffff;
      if (msgsize + 4 > bufsize || pos > msgsize + 4)
	return -1;
      bufsize = msgsize + 4;
    }
  } while (pos < bufsize);

  return bufsize - 4;
}

static int xidctr;
enum clnt_stat
srpc_callraw (int fd, u_int32_t prog, u_int32_t vers, u_int32_t proc,
	      xdrproc_t inproc, void *in, xdrproc_t outproc, void *out,
	      AUTH *auth)
{
  u_int32_t buf[(MAXMSG+7)/4];
  XDR x;
  u_int32_t xid = ++xidctr;
  enum clnt_stat err;
  size_t msgsize;
  ssize_t n;
  struct rpc_msg msg;

  xdrmem_create (&x, (char *) &buf[1], MAXMSG, XDR_ENCODE);
  err = srpc_marshal_call (&x, xid, prog, vers, proc, auth, inproc, in);
  xdr_destroy (&x);
  if (err)
    return err;

  msgsize = xdr_getpos (&x);
  buf[0] = htonl (0x80000000|msgsize);
  if ((size_t) write (fd, buf, msgsize + 4) != msgsize + 4)
    return RPC_CANTSEND;

  n = readpkt (fd, buf, sizeof (buf));
  if (n <= 0)
    return RPC_CANTRECV;
  msgsize = n;

  xdrmem_create (&x, (char *) &buf[1], msgsize, XDR_DECODE);
  bzero (&msg, sizeof (msg));
  msg.acpted_rply.ar_results.where = (char *) out;
  msg.acpted_rply.ar_results.proc = outproc;
  err = srpc_marshal_reply (&x, &msg);
  if (msg.rm_direction == REPLY && msg.rm_reply.rp_stat == MSG_ACCEPTED
      && msg.acpted_rply.ar_verf.oa_base)
    free (msg.acpted_rply.ar_verf.oa_base);
  xdr_destroy (&x);

  return err;
}

enum clnt_stat
srpc_call (const struct rpc_program_tc *rpp, int fd, u_int32_t proc,
	   void *in, void *out)
{
  assert (proc < rpp->nproc);
  return srpc_callraw (fd, rpp->progno, rpp->versno, proc,
		       rpp->tbl[proc].xdr_arg, in,
		       rpp->tbl[proc].xdr_res, out, NULL);
}
