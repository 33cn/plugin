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

#define RQCRED_SIZE     400             /* this size is excessive */

/*
 *  nfsproc handler routines.  
 */
void nfsproc_null(nfsproc_argument *, nfsproc_result *);
void nfsproc_getattr(nfsproc_argument *, nfsproc_result *);
void nfsproc_setattr(nfsproc_argument *, nfsproc_result *);
void nfsproc_root(nfsproc_argument *, nfsproc_result *);
void nfsproc_lookup(nfsproc_argument *, nfsproc_result *);
void nfsproc_readlink(nfsproc_argument *, nfsproc_result *);
void nfsproc_read(nfsproc_argument *, nfsproc_result *);
void nfsproc_writecache(nfsproc_argument *, nfsproc_result *);
void nfsproc_write(nfsproc_argument *, nfsproc_result *);
void nfsproc_create(nfsproc_argument *, nfsproc_result *);
void nfsproc_remove(nfsproc_argument *, nfsproc_result *);
void nfsproc_rename(nfsproc_argument *, nfsproc_result *);
void nfsproc_link(nfsproc_argument *, nfsproc_result *);
void nfsproc_symlink(nfsproc_argument *, nfsproc_result *);
void nfsproc_mkdir(nfsproc_argument *, nfsproc_result *);
void nfsproc_rmdir(nfsproc_argument *, nfsproc_result *);
void nfsproc_readdir(nfsproc_argument *, nfsproc_result *);
void nfsproc_statfs(nfsproc_argument *, nfsproc_result *);

/*
 *  All the information necessary to handle any NFS request.
 */
struct nfsproc {
    void (*handler)(nfsproc_argument *, nfsproc_result *);
    xdrproc_t xdr_arg_type;
    xdrproc_t xdr_res_type;
};

#define X xdrproc_t

static struct nfsproc const nfsproc_table[] = {
    { nfsproc_null,       (X) xdr_void,       (X) xdr_void },
    { nfsproc_getattr,    (X) xdr_fhandle,    (X) xdr_attrstat },
    { nfsproc_setattr,    (X) xdr_sattrargs,  (X) xdr_attrstat },
    { nfsproc_root,       (X) xdr_void,       (X) xdr_void },
    { nfsproc_lookup,     (X) xdr_diropargs,  (X) xdr_diropres },
    { nfsproc_readlink,   (X) xdr_fhandle,    (X) xdr_readlinkres },
    { nfsproc_read,       (X) xdr_readargs,   (X) xdr_readres },
    { nfsproc_writecache, (X) xdr_void,       (X) xdr_void },
    { nfsproc_write,      (X) xdr_writeargs,  (X) xdr_attrstat },
    { nfsproc_create,     (X) xdr_createargs, (X) xdr_diropres },
    { nfsproc_remove,     (X) xdr_diropargs,  (X) xdr_nfsstat },
    { nfsproc_rename,     (X) xdr_renameargs, (X) xdr_nfsstat },
    { nfsproc_link,       (X) xdr_linkargs,   (X) xdr_nfsstat },
    { nfsproc_symlink,    (X) xdr_symlinkargs,(X) xdr_nfsstat },
    { nfsproc_mkdir,      (X) xdr_createargs, (X) xdr_diropres },
    { nfsproc_rmdir,      (X) xdr_diropargs,  (X) xdr_nfsstat },
    { nfsproc_readdir,    (X) xdr_readdirargs,(X) xdr_readdirres },
    { nfsproc_statfs,     (X) xdr_fhandle,    (X) xdr_statfsres },
};

#undef X

const int num_nfsprocs = sizeof (nfsproc_table) / sizeof (struct nfsproc);

static void set_access(struct svc_req *r);


/* 

When not using replication must sync modifications to disk before
replying to the client. By calling sync_fs to guarantee NFS V2
semantics.  This is not being done to allow a more conservative
comparison.  

*/

#ifdef NO_REPLICATION
#define SYNC()
#define Byz_modify2(a, b)
#define Byz_modify1(a)
#else
#define SYNC()
#endif


/*
 *   nfsd_dispatch -- This function is called for each NFS request.
 *		    It dispatches the NFS request to a handler function.
 */

/* TODO: Cannot be globals if we allow concurrency */
extern struct timeval cur_time;
static SVCXPRT *byz_svc = 0;

int 
nfsd_dispatch(Byz_req *inb, Byz_rep *outb, Byz_buffer *non_det, int client, int ro) {
  struct svc_req r;
  struct rpc_msg m;
  char cred_area[2*MAX_AUTH_BYTES + RQCRED_SIZE]; 
  nfsproc_argument argument;
  nfsproc_result result;
  struct nfsproc const *np;
  enum auth_stat astat;

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

  /* Initialize request credentials to cooked form */
  if ((astat = svcbyz_authenticate(&r, &m)) != AUTH_OK) {
    svcerr_auth(byz_svc, astat);
    outb->size = svcbyz_reply_bytes(byz_svc);
    return 0;
  }
 
  // Note that we check if the request is truly read only here and
  // disallow execution if it is not with a noproc error.

#ifdef READ_ONLY_OPT
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


  /* Call handler */
  memset (&result, 0, sizeof (result));
  (*np->handler)(&argument, &result);

  if (!svc_sendreply(byz_svc, np->xdr_res_type, (void *) &result))
    svcerr_systemerr(byz_svc);

  outb->size = svcbyz_reply_bytes(byz_svc);

  /* Free any data allocated by XDR library */
  svc_freeargs(byz_svc, np->xdr_arg_type, (void *) &argument);

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
 *  HANDLER FUNCTIONS
 */
void nfsproc_null(nfsproc_argument *ap, nfsproc_result *rp) { }


#define MIN_CLIENT_INUM 2

static void getattr(struct inode *i, fattr *a) {
  *a = i->attr;
  a->rdev = 0;
  a->fsid = 1;
  a->blocksize = Page_size;
  /* Make client happy by not using inums 0, 1 and 2 */
  a->fileid += MIN_CLIENT_INUM;  

  /*
   *  The file type is specified twice.  "This is really a bug in the
   *  protocol and will be fixed in future versions." -- rfc1094
   */
  switch (a->type) {
  case NFDIR:
    a->mode |= 0040000;
    break;
  case NFCHR:
    a->mode |= 0020000;
    break;
  case NFBLK:
    a->mode |= 0060000;
    break;
  case NFREG:
    a->mode |= 0100000;
    break;
  case NFLNK:
    a->mode |= 0120000;
    break;
  case NFNON:
    a->mode |= 0140000;
    break;
  default:
    break;
  }
}

nfsstat setattr(struct inode *i, sattr *a) {
  nfsstat ret = NFS_OK;

  /*
   *  "Notes:  The use of -1 to indicate an unused field in 'attributes'
   *   is changed in the next version of the protocol."  -- rfc1094
   */
  int unused = (unsigned int) (-1);

  Byz_modify2(&(i->attr), sizeof(i->attr));
  if (a->mode != unused)
    i->attr.mode = a->mode;
  if (a->uid != unused)
    i->attr.uid = a->uid;
  if (a->gid != unused)
    i->attr.gid = a->gid;
  if (a->size != unused && i->attr.type == NFREG) {
    ret = FS_truncate(i, a->size);
  }
  if (a->atime.seconds != unused) {
    i->attr.atime.seconds = a->atime.seconds;
    i->attr.atime.useconds = a->atime.useconds;
  }
  if (a->mtime.seconds != unused) {
    i->attr.mtime.seconds = a->mtime.seconds;
    i->attr.mtime.useconds = a->mtime.useconds;
  }

  /* Update time of last status change */
  i->attr.ctime.seconds = cur_time.tv_sec;
  i->attr.ctime.useconds = cur_time.tv_usec;
  return ret;
}

void nfsproc_getattr(nfsproc_argument *ap, nfsproc_result *rp) {
  fhandle *argp = (fhandle *) ap;
  attrstat *resp = (attrstat *) rp;
  
  struct inode *i = FS_fhandle_to_inode(argp);
  if (i == 0) {
    resp->status = NFSERR_STALE;
    return;
  }

  /* TODO: Check permissions. Since we have inode, this will not
     add a significant performance overhead. But it needs to be
     done in the final code. */

  /* Set attributes */
  getattr(i, &(resp->attributes)); 
  resp->status = NFS_OK;
}

void nfsproc_setattr(nfsproc_argument *ap, nfsproc_result *rp) {
  sattrargs *argp = (sattrargs *) ap;
  attrstat *resp = (attrstat *) rp;

  struct inode *i = FS_fhandle_to_inode(&(argp->file));
  if (i == 0) {
    resp->status = NFSERR_STALE;
    return;
  }

  /* TODO: Check permissions. Since we have inode, this will not
     add a significant performance overhead. But it needs to be
     done in the final code. */

  setattr(i, &(argp->attributes));

  /* Fill in attributes */
  getattr(i, &(resp->attributes));
  resp->status = NFS_OK;

  /* Synchronize file system state if we are not using replication */
  SYNC();
}

void nfsproc_root(nfsproc_argument *ap, nfsproc_result *rp) {
    /* "Obsolete." -- rfc1094 */
}

void nfsproc_lookup(nfsproc_argument *ap, nfsproc_result *rp) {
  diropargs *argp = (diropargs *) ap;
  diropres *resp = (diropres *) rp;
  struct inode *res;

  struct inode *dir = FS_fhandle_to_inode(&(argp->dir));
  if (dir == 0) {
    resp->status = NFSERR_STALE;
    return;
  }

  if (dir->attr.type != NFDIR) {
    resp->status = NFSERR_NOTDIR;
    return;
  }

  if (strlen(argp->name) > MAX_NAME_LEN) {
    resp->status = NFSERR_NAMETOOLONG;
    return;
  }

  /* TODO: Check permissions. Since we have inode, this will not
     add a significant performance overhead. But it needs to be
     done in the final code. */

  
  if ((res = FS_lookup(dir, argp->name)) == 0) {
    resp->status = NFSERR_NOENT;
    return;
  }

  FS_inode_to_fhandle(res, &(resp->diropok.file));
  getattr(res, &(resp->diropok.attributes));
  resp->status = NFS_OK;
  

  /* Not synchronizing file system state at this point even though 
     time last accessed of the directory changes */
}


void nfsproc_readlink(nfsproc_argument *ap, nfsproc_result *rp) {
  fhandle *argp = (fhandle *) ap;
  readlinkres *resp = (readlinkres *) rp;
  
  struct inode *i = FS_fhandle_to_inode(argp);
  if (i == 0) {
    resp->status = NFSERR_STALE;
    return;
  }
  
  if (i->attr.type != NFLNK) {
    resp->status = NFSERR_IO;
    return;
  }

  /* TODO: Check permissions. Since we have inode, this will not
     add a significant performance overhead. But it needs to be
     done in the final code. */

  
  resp->status = NFS_OK;
  resp->data = i->iu.data;

  /* Not synchronizing file system state at this point even though 
     time last accessed of the link */  
}


void nfsproc_read(nfsproc_argument *ap, nfsproc_result *rp) {
  readargs *argp = (readargs *) ap;
  readres *resp = (readres *) rp;

  struct inode *i = FS_fhandle_to_inode(&(argp->file));
  if (i == 0) {
    resp->status = NFSERR_STALE;
    return;
  }

  if (i->attr.type != NFREG) {
    resp->status = NFSERR_IO;
    return;
  }

  /* TODO: Check permissions. Since we have inode, this will not
     add a significant performance overhead. But it needs to be
     done in the final code. */

  resp->reply.data.data_len = 
    FS_read(i, argp->offset, argp->count, &(resp->reply.data.data_val));
  getattr(i, &(resp->reply.attributes));
  resp->status = NFS_OK;

  /* Not synchronizing file system state at this point even though 
     time last accessed of the file changed */
}

void nfsproc_writecache(nfsproc_argument *ap, nfsproc_result *rp) {
  /* "To be used in the next protocol revision." -- rfc1094 */
}

void nfsproc_write(nfsproc_argument *ap, nfsproc_result *rp) {
  writeargs *argp = (writeargs *) ap;
  attrstat *resp = (attrstat *) rp;

  struct inode *i = FS_fhandle_to_inode(&(argp->file));
  if (i == 0) {
    resp->status = NFSERR_STALE;
    return;
  }

  if (i->attr.type != NFREG) {
    resp->status = NFSERR_IO;
    return;
  }

  /* TODO: Check permissions. Since we have inode, this will not
     add a significant performance overhead. But it needs to be
     done in the final code. */

  resp->status = 
    FS_write(i, argp->offset, argp->data.data_len, argp->data.data_val);
  if (resp->status == NFS_OK) {
    getattr(i, &(resp->attributes));

    /* Synchronize file system state */
    SYNC();
  }
}
	

void nfsproc_create(nfsproc_argument *ap, nfsproc_result *rp) {
  createargs *argp = (createargs *) ap;
  diropres *resp = (diropres *) rp;
  struct inode *new_file;

  struct inode *dir = FS_fhandle_to_inode(&(argp->where.dir));
  if (dir == 0) {
    resp->status = NFSERR_STALE;
    return;
  }

  if (dir->attr.type != NFDIR) {
    resp->status = NFSERR_NOTDIR;
    return;
  }

  new_file = FS_create_file(dir);
  if (new_file == 0) {
    resp->status = NFSERR_NOSPC;
    return;
  }
  
  resp->status = FS_link(dir, argp->where.name, new_file);
  if (resp->status == NFS_OK) {
    resp->status = setattr(new_file, &(argp->attributes));
    if (resp->status == NFS_OK) {
      getattr(new_file, &(resp->diropok.attributes));
      FS_inode_to_fhandle(new_file, &(resp->diropok.file));

      /* Synchronize file system state */
      SYNC();
      return;
    }
    FS_unlink(dir, argp->where.name, 0);
    return;
  } else {
    FS_free_inode(new_file);
  } 
}


void nfsproc_remove(nfsproc_argument *ap, nfsproc_result *rp) {
  diropargs *argp = (diropargs *) ap;
  nfsstat *resp = (nfsstat *) rp;

  struct inode *dir = FS_fhandle_to_inode(&(argp->dir));
  if (dir == 0) {
    *resp = NFSERR_STALE;
    return;
  }

  *resp = FS_unlink(dir, argp->name, 0);

  /* Synchronize file system state */
  SYNC();
}
    

void nfsproc_rename(nfsproc_argument *ap, nfsproc_result *rp) {
  renameargs *argp = (renameargs *) ap;
  nfsstat *resp = (nfsstat *) rp;
  struct inode *old_dir, *new_dir;
  struct inode *i;
  struct inode *to_i;

  old_dir = FS_fhandle_to_inode(&(argp->from.dir));
  if (old_dir == 0) {
    *resp = NFSERR_STALE;
    return;
  }

  new_dir = FS_fhandle_to_inode(&(argp->to.dir));
  if (new_dir == 0) {
    *resp = NFSERR_STALE;
    return;
  }

  if (new_dir->attr.type != NFDIR || old_dir->attr.type != NFDIR) {
    *resp = NFSERR_NOTDIR;
    return;
  }

  i = FS_lookup(old_dir, argp->from.name);
  if (i == 0) {
    *resp = NFSERR_NOENT;
    return;
  }

  to_i = FS_lookup(new_dir, argp->to.name);
  if (to_i != 0) {
    *resp = FS_unlink(new_dir, argp->to.name, (to_i->attr.type == NFDIR));
    if (*resp != NFS_OK) {
      return;
    }
  }

  *resp = FS_link(new_dir, argp->to.name, i);
  if (*resp != NFS_OK) {
    if (to_i != 0) 
      FS_link(new_dir, argp->to.name, to_i);
    return;
  }

  *resp = FS_unlink(old_dir, argp->from.name, (i->attr.type == NFDIR));
  assert(*resp == NFS_OK);

  /* Synchronize file system state */
  SYNC();
}

void nfsproc_link(nfsproc_argument *ap, nfsproc_result *rp) {
  linkargs *argp = (linkargs *) ap;
  nfsstat *resp = (nfsstat *) rp;
  struct inode *from, *to;

  from = FS_fhandle_to_inode(&(argp->from));
  if (from == 0) {
    *resp = NFSERR_STALE;
    return;
  }
  
  to = FS_fhandle_to_inode(&(argp->to.dir));
  if (to == 0) {
    *resp = NFSERR_STALE;
    return;
  }

  if (to->attr.type != NFDIR) {
    *resp = NFSERR_NOTDIR;
    return;
  }

  *resp = FS_link(to, argp->to.name, from);

  /* Synchronize file system state */
  SYNC();
}

void nfsproc_symlink(nfsproc_argument *ap, nfsproc_result *rp) {
  symlinkargs *argp = (symlinkargs *) ap;
  nfsstat *resp = (nfsstat *) rp;
  struct inode *dir, *new_inode;

  dir = FS_fhandle_to_inode(&(argp->from.dir));
  if (dir == 0) {
    *resp = NFSERR_STALE;
    return;
  }
  
  if (dir->attr.type != NFDIR) {
    *resp = NFSERR_NOTDIR;
    return;
  }
   
  new_inode = FS_create_symlink(argp->to); 
  if (new_inode == 0) {
    *resp = NFSERR_NOSPC;
    return;
  }


  *resp = FS_link(dir, argp->from.name, new_inode);
  if (*resp != NFS_OK) {
    FS_free_inode(new_inode);
  }

  /* Synchronize file system state */
  SYNC();
 
  /*
   * Ignore the attributes, as the RFC1094 says
   * "On UNIX servers the attributes are never used...",
   */
}   

void nfsproc_mkdir(nfsproc_argument *ap, nfsproc_result *rp) {
  createargs *argp = (createargs *) ap;
  diropres *resp = (diropres *) rp;
  struct inode *old_dir, *new_dir;

  old_dir = FS_fhandle_to_inode(&(argp->where.dir));
  if (old_dir == 0) {
    resp->status = NFSERR_STALE;
    return;
  }

  if (old_dir->attr.type != NFDIR) {
    resp->status = NFSERR_NOTDIR;
    return;
  }

  new_dir = FS_create_dir(old_dir);
  if (new_dir == 0) {
    resp->status = NFSERR_NOSPC;
    return;
  }
  
  resp->status = FS_link(old_dir, argp->where.name, new_dir);
  if (resp->status == NFS_OK) {
    resp->status = setattr(new_dir, &(argp->attributes));
    if (resp->status == NFS_OK) {
      getattr(new_dir, &(resp->diropok.attributes));
      FS_inode_to_fhandle(new_dir, &(resp->diropok.file));

      /* Synchronize file system state */
      SYNC();

      return;
    }
    FS_unlink(old_dir, argp->where.name, 1);
  } else {
    /* Delete directory */
    FS_free_inode(new_dir);
  } 
}


void nfsproc_rmdir(nfsproc_argument *ap, nfsproc_result *rp) {
  diropargs *argp = (diropargs *) ap;
  nfsstat *resp = (nfsstat *) rp;

  struct inode *parent_dir = FS_fhandle_to_inode(&(argp->dir));
  if (parent_dir == 0) {
    *resp = NFSERR_STALE;
    return;
  }

  *resp = FS_unlink(parent_dir, argp->name, 1);

  /* Synchronize file system state */
  SYNC();
}


static int dpsize(char *name) {
#define DP_SLOP 16
  return (sizeof(entry) + strlen(name) + DP_SLOP);
}


entry dir_entries[NFS_MAXDATA];
void nfsproc_readdir(nfsproc_argument *ap, nfsproc_result *rp) {
  readdirargs *argp = (readdirargs *) ap;
  readdirres *resp = (readdirres *) rp;
  int res_size, dloc, eof, size;
  entry **e;
  int index = 0;

  struct inode *dir = FS_fhandle_to_inode(&(argp->dir));
  if (dir == 0) {
    resp->status = NFSERR_STALE;
    return;
  }

  res_size = 0;
  dloc = *(int *)&(argp->cookie);
  e = &(resp->readdirok.entries);
  eof = FALSE;
  while (1) {
    struct dir_entry *de;

    if (dloc >= dir->attr.size) {
      eof = TRUE;
      break;
    }

    de = Directory_entry(dir, dloc);
    size = dpsize(de->name);
    if ((res_size + size > argp->count) && (res_size > 0)) {
      /* If the count of bytes is exceeded and we returned at least
         one entry break. */
      break;
    }

    *e = dir_entries+index;
    (*e)->name = de->name;
    /* Make client happy by not using inums 0 and 1*/
    (*e)->fileid = de->inum + MIN_CLIENT_INUM; 
    dloc++;
    index++;

    *((int*)((*e)->cookie.data)) = dloc;
    e = &((*e)->nextentry);
    res_size += size;
  }
    
  *e = NULL;
  resp->readdirok.eof = eof;


  /* Not synchronizing file system state even though time last accessed
     of the directory changes */
}


void nfsproc_statfs(nfsproc_argument *ap, nfsproc_result *rp) {
  fhandle *argp = (fhandle *) ap;
  statfsres *resp = (statfsres *) rp;

  struct inode *i = FS_fhandle_to_inode(argp);

  if (i == 0) {
    resp->status = NFSERR_STALE;
    return;
  }

  resp->status = NFS_OK;
  resp->info.tsize = 4096;
  resp->info.bsize = Page_size;
  FS_free_blocks(&(resp->info.blocks), &(resp->info.bfree));
  resp->info.bavail = resp->info.blocks-resp->info.bfree;

#ifdef PRINT_BSTATS
  Byz_print_stats();
#endif
}
