#include <assert.h>
#include <sys/types.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <rpc/rpc.h>
#include <rpc/auth.h>

#include "libbyz.h"
#include "nfs.h"
#include "nfsd.h"
#include "fs_interf.h"
#include "inode.h"
#include "svc.h"

#include "stats.h"

#ifdef STATS
int numcalls[NUM_CALLS];
#endif

#define RQCRED_SIZE     400             /* this size is excessive */

#define MAX_ENTRIES_READDIR 600

nfscookie cookies[MAX_ENTRIES_READDIR];
int prev_fid = 0;

/*
 *  argument conversion routines.  
 */
int nfsargconv_noconv(nfsproc_argument *);
int nfsargconv_fhandle(nfsproc_argument *);
int nfsargconv_readirargs(nfsproc_argument *);
int nfsargconv_diropargs(nfsproc_argument *);
int nfsargconv_createargs(nfsproc_argument *);
int nfsargconv_writeargs(nfsproc_argument *);
int nfsargconv_readargs(nfsproc_argument *);
int nfsargconv_sattrargs(nfsproc_argument *);
int nfsargconv_renameargs(nfsproc_argument *);
int nfsargconv_symlinkargs(nfsproc_argument *);
int nfsargconv_linkargs(nfsproc_argument *);

/*
 * state update (before NFS RPC) routines
 */
void nfs_pre_update_void(nfsproc_argument *);
void nfs_pre_update_rm(nfsproc_argument *);
void nfs_pre_update_last(nfsproc_argument *);
void nfs_pre_update_rename(nfsproc_argument *);

/*
 * state update (after NFS RPC) routines
 */
void nfs_pos_update_sattr(nfsproc_argument *, nfsproc_result *);
void nfs_pos_update_create(nfsproc_argument *, nfsproc_result *);
void nfs_pos_update_write(nfsproc_argument *, nfsproc_result *);
void nfs_pos_update_read(nfsproc_argument *, nfsproc_result *);
void nfs_pos_update_void(nfsproc_argument *, nfsproc_result *);
void nfs_pos_update_lkup(nfsproc_argument *, nfsproc_result *);
void nfs_pos_update_rm(nfsproc_argument *, nfsproc_result *);
void nfs_pos_update_mkdir(nfsproc_argument *, nfsproc_result *);
void nfs_pos_update_rename(nfsproc_argument *, nfsproc_result *);
void nfs_pos_update_symlink(nfsproc_argument *, nfsproc_result *);

/*
 *  result conversion routines.  
 */
void nfsres_void(nfsproc_result *);
void nfsresconv_attrstat(nfsproc_result *);
void nfsresconv_readdirres(nfsproc_result *);
void nfsresconv_diropres(nfsproc_result *);
void nfsresconv_readres(nfsproc_result *);


/*
 *  All the information necessary to handle any NFS request.
 */
struct nfsproc {
    int (*argument_conv)(nfsproc_argument *);
    void (*pre_update_replica_state)(nfsproc_argument *);
    void (*result_conv)(nfsproc_result *);
    void (*pos_update_replica_result)(nfsproc_argument *, nfsproc_result *);
    xdrproc_t xdr_arg_type;
    xdrproc_t xdr_res_type;
};

#define X xdrproc_t

static struct nfsproc const nfsproc_table[] = {
    { nfsargconv_noconv,       nfs_pre_update_void,       nfsres_void,
      nfs_pos_update_void,     (X) xdr_void,       (X) xdr_void },
    { nfsargconv_fhandle,      nfs_pre_update_void,    nfsresconv_attrstat,
      nfs_pos_update_void,     (X) xdr_fhandle,    (X) xdr_attrstat },
    { nfsargconv_sattrargs,    nfs_pre_update_last,     nfsresconv_attrstat,
      nfs_pos_update_sattr,    (X) xdr_sattrargs,  (X) xdr_attrstat },
    { nfsargconv_noconv,       nfs_pre_update_void,       nfsres_void,
      nfs_pos_update_void,     (X) xdr_void,       (X) xdr_void },
    { nfsargconv_diropargs,    nfs_pre_update_void,     nfsresconv_diropres,
      nfs_pos_update_lkup,     (X) xdr_diropargs,  (X) xdr_diropres },
    { nfsargconv_fhandle,      nfs_pre_update_void,   nfsres_void,
      nfs_pos_update_void,     (X) xdr_fhandle,    (X) xdr_readlinkres },
    { nfsargconv_readargs,     nfs_pre_update_void,       nfsresconv_readres,
      nfs_pos_update_read,     (X) xdr_readargs,   (X) xdr_readres },
    { nfsargconv_noconv,       nfs_pre_update_void, nfsres_void,
      nfs_pos_update_void,     (X) xdr_void,       (X) xdr_void },
    { nfsargconv_writeargs,    nfs_pre_update_last,       nfsresconv_attrstat,
      nfs_pos_update_write,    (X) xdr_writeargs,  (X) xdr_attrstat },
    { nfsargconv_createargs,   nfs_pre_update_last,      nfsresconv_diropres,
      nfs_pos_update_create,   (X) xdr_createargs, (X) xdr_diropres },
    { nfsargconv_diropargs,    nfs_pre_update_rm,       nfsres_void,
      nfs_pos_update_rm,       (X) xdr_diropargs,  (X) xdr_nfsstat },
    { nfsargconv_renameargs,   nfs_pre_update_rename,     nfsres_void,
      nfs_pos_update_rename,   (X) xdr_renameargs, (X) xdr_nfsstat },
    { nfsargconv_linkargs,     nfs_pre_update_void,       nfsres_void,
      nfs_pos_update_void,     (X) xdr_linkargs,   (X) xdr_nfsstat },
    { nfsargconv_symlinkargs,  nfs_pre_update_last,     nfsres_void,
      nfs_pos_update_symlink,  (X) xdr_symlinkargs,(X) xdr_nfsstat },
    { nfsargconv_createargs ,  nfs_pre_update_last,       nfsresconv_diropres, 
      nfs_pos_update_mkdir,    (X) xdr_createargs, (X) xdr_diropres },
    { nfsargconv_diropargs,    nfs_pre_update_rm,      nfsres_void,
      nfs_pos_update_rm,       (X) xdr_diropargs,  (X) xdr_nfsstat },
    { nfsargconv_readirargs,   nfs_pre_update_void, nfsresconv_readdirres,
      nfs_pos_update_void,     (X) xdr_readdirargs,(X) xdr_readdirres },
/* RAF * changed nfsargconv_noconv to _fhandle */
    { nfsargconv_fhandle,       nfs_pre_update_void,     nfsres_void, /* RAF */
      nfs_pos_update_void,     (X) xdr_fhandle,    (X) xdr_statfsres },
};

#undef X

const int num_nfsprocs = sizeof (nfsproc_table) / sizeof (struct nfsproc);

static void set_access(struct svc_req *r);

/*
 *   nfsd_dispatch -- This function is called for each NFS request.
 *		    It dispatches the NFS request to a handler function.
 */

/* TODO: Cannot be globals if we allow concurrency */
extern struct timeval cur_time;
extern char hostname[];
static SVCXPRT *byz_svc = 0;
CLIENT *nfs_clnt = 0;
int last_inum, last_inum_from, last_inum_to, inum_clobbered;

#ifndef MAX_MACHINE_NAME
#define MAX_MACHINE_NAME 256
#endif

void perform_RPC_call(int function, char *arg, char *res)
{
  enum clnt_stat repval;
  struct nfsproc const *np = nfsproc_table + function;
  struct timeval t = {100, 0};
  char *err;

  if (!nfs_clnt) {
    nfs_clnt = clnt_create(hostname, NFS_PROGRAM, NFS_VERSION, "udp");
    if (!nfs_clnt) {
      printf("Couldn't create client\n");
      return;
    }
    else {
      nfs_clnt->cl_auth = authunix_create_default();
      if (!nfs_clnt->cl_auth) {
	printf("Error in authunix create\n");
	return;
      }

    }
  }

  repval = clnt_call(nfs_clnt, function,
		     np->xdr_arg_type, arg,
		     np->xdr_res_type, res,
		     t);
  if (repval != RPC_SUCCESS) {
    clnt_perror(nfs_clnt, err);
    fprintf(stderr, "Error %d in RPC call. %s\n", repval, err);
  }
}

void free_RPC_res(int function, char *res)
{
  if (!clnt_freeres(nfs_clnt, nfsproc_table[function].xdr_res_type,
		    res))
    fprintf(stderr, "clnt_freeres failed\n");
}


// debug - RR-TODO XXX erase
int lookup_cached = 0, read_cached = 0;

int 
nfsd_dispatch(Byz_req *inb, Byz_rep *outb, Byz_buffer *non_det, int client, int ro) {
  struct svc_req r;
  struct rpc_msg m;
  char cred_area[2*MAX_AUTH_BYTES + RQCRED_SIZE]; 
  nfsproc_argument argument;
  nfsproc_result result;
  struct nfsproc const *np;
  enum auth_stat astat;
  char *dir; int i;  //debug - XXX erase
  struct direntry *de;

#ifdef STATS
  start_counter(INIT_REQ);
#endif

  if (!ro) {
    /* Read in current time as chosen by primary into a global */
    bcopy(non_det->contents, (char*)&cur_time, sizeof(struct timeval));
  }

  if (byz_svc == 0) {
    byz_svc = svcbyz_create(inb->contents, inb->size, outb->contents, outb->size);
  } else {
    svcbyz_recycle(byz_svc, inb->contents, inb->size, outb->contents, outb->size);
  }

  /* Decode rpc_msg after initializing credential pointers into cred_area */
  m.rm_call.cb_cred.oa_base = cred_area;
  m.rm_call.cb_verf.oa_base = &(cred_area[MAX_AUTH_BYTES]);
  r.rq_clntcred = &(cred_area[2*MAX_AUTH_BYTES]);
  if (!SVC_RECV(byz_svc, &m)) {
    svcerr_noprog(byz_svc);
    outb->size = svcbyz_reply_bytes(byz_svc);
    return 0;
  }

  /* Fill in request */
  r.rq_xprt = byz_svc;
  r.rq_prog = m.rm_call.cb_prog;
  r.rq_vers = m.rm_call.cb_vers;
  r.rq_proc = m.rm_call.cb_proc;

#ifdef STATS
  stop_counter(INIT_REQ);
  numcalls[r.rq_proc]++;
  start_counter(MIN_CALL_STATS + r.rq_proc * NUM_STATS_PER_CALL + 0);
#endif

  /* Initialize request credentials to cooked form */
  if ((astat = svcbyz_authenticate(&r, &m)) != AUTH_OK) {
    svcerr_auth(byz_svc, astat);
    outb->size = svcbyz_reply_bytes(byz_svc);
    return 0;
  }
 
  // Note that we check if the request is truly read only here and
  // disallow execution if it is not with a noproc error.

#ifndef NO_READ_ONLY_OPT
  if (r.rq_proc < 0 || r.rq_proc >= num_nfsprocs 
      || (r.rq_proc != 1 && r.rq_proc != 4 && r.rq_proc != 6 && ro)) {
#else
  if (r.rq_proc < 0 || r.rq_proc >= num_nfsprocs 
      || (r.rq_proc != 1 && ro)) {
#endif
    svcerr_noproc(byz_svc);
    outb->size = svcbyz_reply_bytes(byz_svc); 
    return 0;
  }
 
  /* Get credential out of RPC message into a global */
  set_access(&r);

 /* Compute dispatch table pointer, and decode arguments */
  np = nfsproc_table + r.rq_proc;
  memset(&argument, 0, sizeof (argument));
  if (!svc_getargs(byz_svc, np->xdr_arg_type, (void *) &argument)) {
    svcerr_decode(byz_svc);
    outb->size = svcbyz_reply_bytes(byz_svc);
    return 0;
  }

  memset (&result, 0, sizeof (result));

#ifdef STATS
  stop_counter(MIN_CALL_STATS + r.rq_proc * NUM_STATS_PER_CALL + 0);
  start_counter(MIN_CALL_STATS + r.rq_proc * NUM_STATS_PER_CALL + 1);
#endif

  /* Convert arguments */
  if (((*np->argument_conv)(&argument)) < 0) {
    /* could not convert arguments */
    //    fprintf(stderr, "oi ");
    *((nfsstat *)(&result)) = NFSERR_STALE;
    if (!svc_sendreply(byz_svc, np->xdr_res_type, (void *) &result))
      svcerr_systemerr(byz_svc);
    outb->size = svcbyz_reply_bytes(byz_svc);
    svc_freeargs(byz_svc, np->xdr_arg_type, (void *) &argument);
#ifdef STATS
    stop_counter(MIN_CALL_STATS + r.rq_proc * NUM_STATS_PER_CALL + 1);
#endif
    return 0;
  }

#ifdef STATS
  stop_counter(MIN_CALL_STATS + r.rq_proc * NUM_STATS_PER_CALL + 1);
  start_counter(MIN_CALL_STATS + r.rq_proc * NUM_STATS_PER_CALL + 2);
#endif

  /* Update local state */
  (*np->pre_update_replica_state)(&argument);

#ifdef STATS
  stop_counter(MIN_CALL_STATS + r.rq_proc * NUM_STATS_PER_CALL + 2);
  start_counter(MIN_CALL_STATS + r.rq_proc * NUM_STATS_PER_CALL + 3);
#endif

  /* Perform RPC call */
  perform_RPC_call(r.rq_proc, (char *)&argument, (char *)&result);

#ifdef STATS
  stop_counter(MIN_CALL_STATS + r.rq_proc * NUM_STATS_PER_CALL + 3);
  start_counter(MIN_CALL_STATS + r.rq_proc * NUM_STATS_PER_CALL + 4);
#endif

  /* Update local state with the result of the NFS call */
  (*np->pos_update_replica_result)(&argument, &result);

#ifdef STATS
  stop_counter(MIN_CALL_STATS + r.rq_proc * NUM_STATS_PER_CALL + 4);
  start_counter(MIN_CALL_STATS + r.rq_proc * NUM_STATS_PER_CALL + 5);
#endif

  /* Convert result */
  (*np->result_conv)(&result);

#ifdef STATS
  stop_counter(MIN_CALL_STATS + r.rq_proc * NUM_STATS_PER_CALL + 5);
  start_counter(MIN_CALL_STATS + r.rq_proc * NUM_STATS_PER_CALL + 6);
#endif

  if (!svc_sendreply(byz_svc, np->xdr_res_type, (void *) &result))
    svcerr_systemerr(byz_svc);

  outb->size = svcbyz_reply_bytes(byz_svc);

#if 0
  fprintf(stderr, "Size %d\n", outb->size);
  for (i=0; i<outb->size; i+=4)
    fprintf(stderr, "%10x", *((int*)&outb->contents[i]));
  fprintf(stderr, "\n");
#endif

  /* Free any data allocated by XDR library */
  svc_freeargs(byz_svc, np->xdr_arg_type, (void *) &argument);

  free_RPC_res(r.rq_proc, (char *)&result);

#ifdef STATS
  stop_counter(MIN_CALL_STATS + r.rq_proc * NUM_STATS_PER_CALL + 6);
#endif
  /* return number of bytes in out_stream */  
  return 0;
}


/* Cannot be globals if we allow concurrency */
extern uid_t cred_uid;
extern gid_t cred_gid;

#define NOBODY_UID 65534
#define NOBODY_GID 65534

static void set_access(struct svc_req *r) {
  if (r->rq_cred.oa_flavor == AUTH_UNIX) {
    struct authunix_parms *unix_cred;
    
    unix_cred = (struct authunix_parms *) r->rq_clntcred;
    cred_uid = unix_cred->aup_uid;
    cred_gid = unix_cred->aup_gid;
    /* TODO: check other groups */
  } else {
    cred_uid = NOBODY_UID;
    cred_gid = NOBODY_GID;
  }
}


/*
 * state update (before NFS RPC) routines
 */
void nfs_pre_update_void(nfsproc_argument *ap) {}

void nfs_pre_update_rm(nfsproc_argument *ap) 
{
  diropargs *dargs = (diropargs *) ap;
  diropres dres;

  modified_inode(last_inum); // going to change dir

  perform_RPC_call(NFSPROC_LOOKUP, (char *)dargs, (char *)&dres);

  if (dres.status != NFS_OK) {
    fprintf(stderr, "Error in update NFS RPC call\n");
    last_inum = -1;
  }
  else
    if ((dres.diropok.attributes.type != NFDIR && dres.diropok.attributes.nlink == 1) || (dres.diropok.attributes.type == NFDIR && dres.diropok.attributes.nlink == 2)) {
      last_inum = FS_NFS_handle_to_client_handle(&dres.diropok.file);
      modified_inode(last_inum);  // going to change entry
    }
    else
      last_inum = 0; // Entry still exists: don't delete!
  free_RPC_res(NFSPROC_LOOKUP, (char *)&dres);
}

void nfs_pre_update_last(nfsproc_argument *ap) 
{
  modified_inode(last_inum);
}

void nfs_pre_update_rename(nfsproc_argument *ap)
{
  modified_inode(last_inum); // going to change file
  modified_inode(last_inum_from); // going to change from dir
  modified_inode(last_inum_to); // going to change to dir
  if (inum_clobbered)
    modified_inode(inum_clobbered);
}


/*
 *  argument conversion routines.  
 */
int nfsargconv_noconv(nfsproc_argument *ap) { return 0; }

int nfsargconv_fhandle(nfsproc_argument *ap) {
  fhandle *argp = (fhandle *) ap;

  if ((last_inum = FS_fhandle_to_NFS_handle(argp)) < 0) {
    return -1;
}
  return 0;
}

int nfsargconv_readirargs(nfsproc_argument *ap) {
  readdirargs *argp = (readdirargs *) ap;
  if (nfsargconv_fhandle((nfsproc_argument *)&(argp->dir)) < 0)
    return -1;
  prev_fid = *((int *)&argp->cookie);
  if (prev_fid) {
    memcpy(&argp->cookie, &cookies[prev_fid], sizeof(nfscookie));
  }
  return 0;
}
  
int nfsargconv_diropargs(nfsproc_argument *ap)
{
  diropargs *argp = (diropargs *) ap;
  return nfsargconv_fhandle((nfsproc_argument *)&(argp->dir));
}

int nfsargconv_createargs(nfsproc_argument *ap)
{
  createargs *argp = (createargs *) ap;
  return nfsargconv_fhandle((nfsproc_argument *)&(argp->where.dir));
}

int nfsargconv_writeargs(nfsproc_argument *ap)
{
  writeargs *argp = (writeargs *) ap;
  return nfsargconv_fhandle((nfsproc_argument *)&(argp->file));
}

int nfsargconv_readargs(nfsproc_argument *ap)
{
  readargs *argp = (readargs *) ap;
  return nfsargconv_fhandle((nfsproc_argument *)&(argp->file));
}

int nfsargconv_sattrargs(nfsproc_argument *ap)
{
  sattrargs *argp = (sattrargs *) ap;
#ifdef OMMIT_UID_GID
  argp->attributes.uid = -1;
  argp->attributes.gid = -1;
#endif
  return nfsargconv_fhandle((nfsproc_argument *)&(argp->file));
}

int nfsargconv_renameargs(nfsproc_argument *ap)
{
  renameargs *argp = (renameargs *) ap;
  diropargs *dargs = &argp->from;
  diropres dres;

  if ((last_inum_from = FS_fhandle_to_NFS_handle(&(argp->from.dir))) < 0)
    return -1;
  if ((last_inum_to = FS_fhandle_to_NFS_handle(&(argp->to.dir))) < 0)
    return -1;

  perform_RPC_call(NFSPROC_LOOKUP,(char *)dargs, (char *)&dres);

  if (dres.status != NFS_OK) {
    fprintf(stderr, "Error in update NFS RPC call\n");
    last_inum = -1;
  }
  else 
    last_inum = FS_NFS_handle_to_client_handle(&dres.diropok.file);
  free_RPC_res(NFSPROC_LOOKUP, (char *)&dres);

  // Are we going to overwrite a file? Let's check:
  dargs = &argp->to;

  perform_RPC_call(NFSPROC_LOOKUP,(char *)dargs, (char *)&dres);

  if (dres.status != NFS_OK) {
    inum_clobbered = 0;
  }
  else
    if ((dres.diropok.attributes.type != NFDIR && dres.diropok.attributes.nlink == 1) || (dres.diropok.attributes.type == NFDIR && dres.diropok.attributes.nlink == 2))
    inum_clobbered = FS_NFS_handle_to_client_handle(&dres.diropok.file);
  free_RPC_res(NFSPROC_LOOKUP, (char *)&dres);

    return 0;
}

int nfsargconv_symlinkargs(nfsproc_argument *ap)
{
  symlinkargs *argp = (symlinkargs *) ap;
  return nfsargconv_fhandle((nfsproc_argument *)&(argp->from.dir));
}

int nfsargconv_linkargs(nfsproc_argument *ap)
{
  linkargs *argp = (linkargs *) ap;
  if ((last_inum_from = FS_fhandle_to_NFS_handle(&argp->from)) < 0)
    return -1;
  if ((last_inum_to = FS_fhandle_to_NFS_handle(&argp->to.dir)) < 0)
    return -1;
  return 0;
}

/*
 *  result conversion routines.  
 */
void nfsres_void(nfsproc_result *rp) {}

void nfsresconv_attrstat(nfsproc_result *rp) {
  attrstat *resp = (attrstat *) rp;

  if (resp->status != NFS_OK) {
    fprintf(stderr,"ERR %d ", resp->status);
    return;
  }

  if (FS_attr_NFS_to_client(last_inum, &(resp->attributes)) < 0) 
    fprintf(stderr, "Conversion was not possible. Screwed up!\n");
}

void bubbleSort(entry *list)
{
  entry *pFirst, *pSecond;
  char *tmp;

  for (pFirst = list; pFirst; pFirst = pFirst->nextentry) {
    for (pSecond = pFirst; pSecond; pSecond = pSecond->nextentry) {

      // if we find a node out of place with the following node

      if (strcmp(pFirst->name, pSecond->name) > 0) {
        // swap em... this will sort increasing values
        tmp = pFirst->name;
        pFirst->name = pSecond->name;
        pSecond->name = tmp;
      }
    }
  }
}

void nfsresconv_readdirres(nfsproc_result *rp)
{
  int i;

  entry *ptr;
  readdirres *res = (readdirres *) rp;
  if (res->status != NFS_OK) {
    fprintf(stderr,"ERR %d ", res->status);
    return;
  }
  ptr = res->readdirok.entries;
  while (ptr) {
    ptr->fileid = ++prev_fid; /* XXX - This is not according to the NFS spec */
    memcpy(&cookies[prev_fid], &ptr->cookie, sizeof(nfscookie));
    memset(&ptr->cookie, 0, sizeof(nfscookie));
    *((int *)&ptr->cookie) = prev_fid;
    ptr = ptr->nextentry;
  }

#ifdef ORDER_DIR_ENTRIES
  /* Bubble sort this list XXX Inefficient - change? */
  bubbleSort(res->readdirok.entries);
#endif
}

void nfsresconv_diropres(nfsproc_result *rp)
{
  diropres *resp = (diropres *) rp;

  if (resp->status != NFS_OK) {
    if (resp->status != 2)
      fprintf(stderr,"ERR %d ", resp->status);
    return;
  }

  /* convert file handle */
  if ((last_inum = FS_NFS_handle_to_client_handle(&(resp->diropok.file))) < 0)
    fprintf(stderr, "fh Conversion was not possible. Screwed up!\n");

  /* convert attributes */
  if (FS_attr_NFS_to_client(last_inum, &(resp->diropok.attributes)) < 0) 
    fprintf(stderr, "attr Conversion was not possible. Screwed up!\n");
}

void nfsresconv_readres(nfsproc_result *rp)
{
  attrstat *resp = (attrstat *) rp;

  if (resp->status != NFS_OK) {
    fprintf(stderr,"ERR %d ", resp->status);
    return;
  }

  if (FS_attr_NFS_to_client(last_inum, &(resp->attributes)) < 0) 
    fprintf(stderr, "Conversion was not possible. Screwed up!\n");
}


/*
 * state update (after NFS RPC) routines
 */
void nfs_pos_update_void(nfsproc_argument *ap, nfsproc_result *rp) { }

void nfs_pos_update_create(nfsproc_argument *ap, nfsproc_result *rp)
{
  //  createargs *argp = (createargs *) ap;
  diropres *resp = (diropres *) rp;
  if (resp->status != NFS_OK) {
    fprintf(stderr,"ERR %d ", resp->status);
    return;
  }
  if ((last_inum = FS_create_entry(&(resp->diropok.file), &(resp->diropok.attributes), &cur_time, FILE_INODE_TYPE, last_inum)) < 0)
    fprintf(stderr, "Create file update (after RPC) returned error!\n");
}

void nfs_pos_update_write(nfsproc_argument *ap, nfsproc_result *rp)
{
  /*  writeargs *argp = (writeargs *) ap; */
  attrstat *resp = (attrstat *) rp;
  if (resp->status != NFS_OK) {
    fprintf(stderr,"ERR %d ", resp->status);
    return;
  }
  if (FS_update_time_modified(last_inum, &cur_time) < 0)
    fprintf(stderr, "Problem updating local state after write");
}

void nfs_pos_update_sattr(nfsproc_argument *ap, nfsproc_result *rp)
{
  sattrargs *argp = (sattrargs *) ap;
  attrstat *resp = (attrstat *) rp;

  if (resp->status != NFS_OK) {
    fprintf(stderr,"SATTR ERR %d ", resp->status);
    return;
  }
  if (FS_set_attr(last_inum, &(argp->attributes)) < 0)
    fprintf(stderr, "Problem updating local attr after sattr");
}

void nfs_pos_update_read(nfsproc_argument *ap, nfsproc_result *rp)
{
#ifdef NO_READ_ONLY_OPT
  /* RR-TODO: XXX update atime here */
#endif
}

void nfs_pos_update_lkup(nfsproc_argument *ap, nfsproc_result *rp)
{
#ifdef NO_READ_ONLY_OPT
  /* RR-TODO: XXX update atime here */
#endif
}

void nfs_pos_update_rm(nfsproc_argument *ap, nfsproc_result *rp)
{
  //  diropargs *argp = (diropargs *) ap;
  nfsstat *resp = (nfsstat *) rp;

  if (*resp != NFS_OK) {
    fprintf(stderr, "NFS remove returned error %d\n", *resp);
    return;
  }

  if (last_inum > 0)
    if (FS_remove_entry(last_inum) < 0)
      fprintf(stderr, "Error updating local state after remove\n");
  
}

void nfs_pos_update_rename(nfsproc_argument *ap, nfsproc_result *rp)
{
  renameargs *argp = (renameargs *) ap;
  nfsstat *resp = (nfsstat *) rp;

  diropargs *dargs;
  diropres dres;

  if (*resp != NFS_OK) {
    fprintf(stderr, "NFS rename returned error %d\n", *resp);
    return;
  }

  dargs = &(argp->to);
  perform_RPC_call(NFSPROC_LOOKUP, (char *)dargs, (char *)&dres);

  if (dres.status != NFS_OK)
    fprintf(stderr, "Error in update NFS RPC call\n");
  else {

    if (inum_clobbered > 0)
      if (FS_remove_entry(inum_clobbered) < 0)
	fprintf(stderr, "Error updating local state after overwrite rename\n");

    if (FS_update_file_info(last_inum, &dres.diropok.file, last_inum_to, dres.diropok.attributes.fileid, dres.diropok.attributes.fsid) < 0)
      fprintf(stderr, "Could not update file info after rename\n");
  }
  free_RPC_res(NFSPROC_LOOKUP, (char *)&dres);
}

void nfs_pos_update_mkdir(nfsproc_argument *ap, nfsproc_result *rp)
{
  diropres *resp = (diropres *) rp;
  if (resp->status != NFS_OK) {
    fprintf(stderr,"ERR %d ", resp->status);
    return;
  }
  if ((last_inum = FS_create_entry(&(resp->diropok.file), &(resp->diropok.attributes), &cur_time, DIR_INODE_TYPE, last_inum)) < 0)
    fprintf(stderr, "Create file update (after RPC) returned error!\n");    
}

void nfs_pos_update_symlink(nfsproc_argument *ap, nfsproc_result *rp)
{
  symlinkargs *argp = (symlinkargs *) ap;
  nfsstat *resp = (nfsstat *) rp;

  diropargs *dargs;
  diropres dres;

  if (*resp != NFS_OK) {
    fprintf(stderr, "NFS create symlink returned error %d\n", *resp);
    return;
  }

  dargs = &(argp->from);
  perform_RPC_call(NFSPROC_LOOKUP, (char *)dargs, (char *)&dres);

  if (dres.status != NFS_OK)
    fprintf(stderr, "Error in update NFS RPC call\n");
  else
    if ((last_inum = FS_create_entry(&dres.diropok.file, &dres.diropok.attributes, &cur_time, SYMLINK_INODE_TYPE, last_inum)) < 0)
      fprintf(stderr, "could not update state after creating symlink\n");
  free_RPC_res(NFSPROC_LOOKUP, (char *)&dres);
}
