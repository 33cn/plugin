#include <assert.h>
#include <rpc/rpc.h>

#define in_stream(x) ((XDR*)((x)->xp_p1))
#define out_stream(x) ((XDR*)((x)->xp_p2))

bool_t byz_recv(SVCXPRT *xpt, struct rpc_msg *m) {  
  if (!xdr_callmsg(in_stream(xpt), m))
    return FALSE;
  
  /* Save xid to put in reply */
  xpt->xp_sock = m->rm_xid;  
  return TRUE;
}

enum xprt_stat byz_stat(SVCXPRT *xpt) {                 
  return (XPRT_IDLE); 
}

bool_t byz_getargs(SVCXPRT *xpt, xdrproc_t p, caddr_t b) {
  return (*p)(in_stream(xpt), (caddr_t*)b);
}

bool_t  byz_reply(SVCXPRT *xpt, struct rpc_msg *m) { 
  m->rm_xid = xpt->xp_sock;
  return xdr_replymsg(out_stream(xpt), m) ;
}

bool_t byz_freeargs(SVCXPRT *xpt, xdrproc_t p, caddr_t b) {
  bool_t ret;
  in_stream(xpt)->x_op = XDR_FREE;
  ret = (*p)(in_stream(xpt), (caddr_t*)b);
  in_stream(xpt)->x_op = XDR_DECODE;
  return ret;
}

void  byz_destroy(SVCXPRT *xpt) {
  free(xpt->xp_p1);
  free(xpt->xp_p2);
  free((struct xp_ops*) (xpt->xp_ops));
}

SVCXPRT *
svcbyz_create(char* in, int in_size, char *out, int out_size) {
  register SVCXPRT *xprt;
  struct xp_ops* xops;

  xprt = (SVCXPRT *)malloc(sizeof(SVCXPRT));
  if (xprt == NULL) {
    (void)fprintf(stderr, "svcbyz_create: out of memory\n");
    return (NULL);
  }

  xprt->xp_sock = -1;
  xprt->xp_port = -1;
  xops = (struct xp_ops*) malloc(sizeof(struct xp_ops));
  xops->xp_recv = byz_recv;
  xops->xp_stat = byz_stat;
  xops->xp_getargs = byz_getargs;
  xops->xp_reply = byz_reply;
  xops->xp_freeargs = byz_freeargs;
  xops->xp_destroy = byz_destroy;
  xprt->xp_ops = xops;
  xprt->xp_p1 = malloc(sizeof(XDR));     /* in XDR */
  xprt->xp_p2 = malloc(sizeof(XDR));     /* out XDR */
  xprt->xp_verf.oa_flavor = AUTH_NULL;
  xprt->xp_verf.oa_length = 0;

  xdrmem_create(in_stream(xprt), in, in_size, XDR_DECODE);
  xdrmem_create(out_stream(xprt), out, out_size, XDR_ENCODE);

  return xprt;
}

void 
svcbyz_recycle(SVCXPRT *xpt, char* in, int in_size, char *out, int out_size) {  
  assert(xpt != 0);

  xdrmem_create(in_stream(xpt), in, in_size, XDR_DECODE);
  xdrmem_create(out_stream(xpt), out, out_size, XDR_ENCODE);
}
  
int 
svcbyz_reply_bytes(SVCXPRT *xpt) {
  return xdr_getpos(out_stream(xpt));
}


enum auth_stat
svcbyz_auth_unix(struct svc_req *rqst, struct rpc_msg *msg) {
  register enum auth_stat stat;
  XDR xdrs;
  register struct authunix_parms *aup;
  register int *buf;
  struct area {
    struct authunix_parms area_aup;
    char area_machname[MAX_MACHINE_NAME+1];
    unsigned int area_gids[NGRPS]; // JC: changed to unsigned
  } *area;
  u_int auth_len;
  int str_len, gid_len;
  register int i;

  area = (struct area *) rqst->rq_clntcred;
  aup = &area->area_aup;
  aup->aup_machname = area->area_machname;
  aup->aup_gids = area->area_gids;
  auth_len = (unsigned int)msg->rm_call.cb_cred.oa_length;
  xdrmem_create(&xdrs, msg->rm_call.cb_cred.oa_base, auth_len,XDR_DECODE);
  buf = XDR_INLINE(&xdrs, auth_len);
  if (buf != NULL) {
    aup->aup_time = IXDR_GET_LONG(buf);
    str_len = IXDR_GET_U_LONG(buf);
    if (str_len > MAX_MACHINE_NAME) {
      stat = AUTH_BADCRED;
      goto done;
    }
    bcopy((caddr_t)buf, aup->aup_machname, (u_int)str_len);
    aup->aup_machname[str_len] = 0;
    str_len = RNDUP(str_len);
    buf += str_len / sizeof(int);
    aup->aup_uid = IXDR_GET_LONG(buf);
    aup->aup_gid = IXDR_GET_LONG(buf);
    gid_len = IXDR_GET_U_LONG(buf);
    if (gid_len > NGRPS) {
      stat = AUTH_BADCRED;
      goto done;
    }
    aup->aup_len = gid_len;
    for (i = 0; i < gid_len; i++) {
      aup->aup_gids[i] = IXDR_GET_LONG(buf);
    }
    /*
     * five is the smallest unix credentials structure -
     * timestamp, hostname len (0), uid, gid, and gids len (0).
     */
    if ((5 + gid_len) * BYTES_PER_XDR_UNIT + str_len > auth_len) {
      (void) printf("bad auth_len gid %d str %d auth %d\n",
		    gid_len, str_len, auth_len);
      stat = AUTH_BADCRED;
      goto done;
    }
  } else if (! xdr_authunix_parms(&xdrs, aup)) {
    xdrs.x_op = XDR_FREE;
    (void)xdr_authunix_parms(&xdrs, aup);
    stat = AUTH_BADCRED;
    goto done;
  }
  rqst->rq_xprt->xp_verf.oa_flavor = AUTH_NULL;
  rqst->rq_xprt->xp_verf.oa_length = 0;
  stat = AUTH_OK;
done:
  XDR_DESTROY(&xdrs);
  return (stat);
}

enum
auth_stat svcbyz_authenticate(struct svc_req *rqst, struct rpc_msg *msg) {
  register int cred_flavor;

  rqst->rq_cred = msg->rm_call.cb_cred;
  rqst->rq_xprt->xp_verf.oa_flavor = AUTH_NULL;
  rqst->rq_xprt->xp_verf.oa_length = 0;
  cred_flavor = rqst->rq_cred.oa_flavor;
  if (cred_flavor == AUTH_UNIX) {
    return svcbyz_auth_unix(rqst, msg);
  }

  return (AUTH_REJECTEDCRED);
}


