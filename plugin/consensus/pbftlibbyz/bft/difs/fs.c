#include <fcntl.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <assert.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <rpc/rpc.h>
#include <errno.h>

#include "fs_interf.h"
#include "inode.h"
#include "fs.h"
#include "libbyz.h"
#include "fh_hash.h"
#include "fileid_hash.h"

#include "stats.h"

#ifdef STATS

int num_get_file = 0;
int num_get_dir = 0;
int num_get_free = 0;

int num_put = 0;
int num_put_file = 0;
int num_put_dir = 0;
int num_put_free = 0;

int num_recov = 0;

#endif

/*
 * Global variables:
 */

/* Gobal file system pointer */
struct FS *fs;

/* Cannot be a global if we allow concurrency */
struct timeval cur_time;

/* Cannot be globals if we allow concurrency */
uid_t cred_uid;
gid_t cred_gid;

/* Cannot be a global if we allow concurrency */
char nfs_iobuf[NFS_MAXDATA];

#define min(a,b) (a>b ? (b) : (a))
extern void perform_RPC_call(int function, char *arg, char *res);
extern void free_RPC_res(int function, char *res);

int cmp_dir_entries(const void *a, const void *b);
int cmp_fhandles(fhandle *fh1, fhandle *fh2);

void modified_fs_info();

/* 
 * File system initialization:
 */

static struct {
        enum nfsstat stat;
        int nfs_errno;
} nfs_errtbl[] = {
        { NFS_OK,               0               },
        { NFSERR_PERM,          EPERM           },
        { NFSERR_NOENT,         ENOENT          },
        { NFSERR_IO,            EIO             },
        { NFSERR_NXIO,          ENXIO           },
        { NFSERR_ACCES,         EACCES          },
        { NFSERR_EXIST,         EEXIST          },
        { NFSERR_NODEV,         ENODEV          },
        { NFSERR_NOTDIR,        ENOTDIR         },
        { NFSERR_ISDIR,         EISDIR          },
#ifdef NFSERR_INVAL
        { NFSERR_INVAL,         EINVAL          },      /* that Sun forgot */
#endif
        { NFSERR_FBIG,          EFBIG           },
        { NFSERR_NOSPC,         ENOSPC          },
        { NFSERR_ROFS,          EROFS           },
        { NFSERR_NAMETOOLONG,   ENAMETOOLONG    },
        { NFSERR_NOTEMPTY,      ENOTEMPTY       },
        { NFSERR_DQUOT,         EDQUOT          },
        { NFSERR_STALE,         ESTALE          },
#ifdef EWFLUSH
        { NFSERR_WFLUSH,        EWFLUSH         },
#endif
        { -1,                   EIO             }
};

static char *nfs_strerror(int stat)
{
        int i;
        static char buf[256];

        for (i = 0; nfs_errtbl[i].stat != -1; i++) {
                if (nfs_errtbl[i].stat == stat)
                        return strerror(nfs_errtbl[i].nfs_errno);
        }
        sprintf(buf, "unknown nfs status return value: %d", stat);
        return buf;
}

void add_inode_to_free_list(int inum) {
  modified_fs_info();
  assert(fs->logical_inodes[inum].type != FREE_INODE_TYPE);
  fs->logical_inodes[inum].type = FREE_INODE_TYPE;
  fs->logical_inodes[inum].free_inode.next_free = fs->first_free_inode;
  fs->first_free_inode = inum;
}
  
int FS_fhandle_to_NFS_handle(fhandle *fh)
{
  struct inode_id iid;
  
  FS_client_fhandle_to_inode_id(fh, &iid);

  if (iid.inode_num < 0
   || iid.inode_num >= NUM_LOGICAL_INODES
   || fs->logical_inodes[iid.inode_num].generation_number != iid.generation_num
   || fs->logical_inodes[iid.inode_num].type == FREE_INODE_TYPE) 
    return -1;

  memcpy(fh, &fs->logical_inodes[iid.inode_num].entry_inode.NFS_handle,
	 sizeof(fhandle));
  
  return iid.inode_num;

}

int FS_NFS_handle_to_client_handle(fhandle *fh) {
  int generation_num;
  int inode_num = find(fh);

  if (inode_num == -1) {
    fprintf(stderr, "Could not perform NFS handle lookup in hash.\n");
    return -1;
  }

  assert(fs->logical_inodes[inode_num].type != FREE_INODE_TYPE);

  generation_num = fs->logical_inodes[inode_num].generation_number;
  NFS_inode_id_to_client_fhandle(inode_num, generation_num, fh);
  
  return inode_num;
}


void FS_client_fhandle_to_inode_id(fhandle *fh, struct inode_id *id) {  
  int *log_inode = (int *) fh;
  id->inode_num = *log_inode;
  id->generation_num = *(++log_inode);
}

void NFS_inode_id_to_client_fhandle(int logical_inode_num, int generation_num,
				    fhandle *fh) {
  int *ptr = (int *) fh;

  memset(fh, 0, sizeof(fhandle));
  *ptr = logical_inode_num;
  ptr++;
  *ptr = generation_num;
}


int FS_attr_NFS_to_client(int inum, fattr *attr) {
  
  if (inum < 0 || inum > NUM_LOGICAL_INODES ||
      fs->logical_inodes[inum].type == FREE_INODE_TYPE)
    return -1;

  if (attr->type == NFDIR) {
    attr->size = Block_size;
    attr->blocks = 1;
  }
  else
    attr->blocks = (attr->size >> 12) + 1;

  attr->rdev = 0;
  attr->fileid = inum;
  attr->blocksize = Block_size;
  attr->fsid = BYZ_FSID;
#ifdef OMMIT_UID_GID
  attr->uid = 0;
  attr->gid = 0;
#endif
  attr->atime.seconds = fs->logical_inodes[inum].entry_inode.atime.seconds;
  attr->mtime.seconds = fs->logical_inodes[inum].entry_inode.mtime.seconds;
  attr->ctime.seconds = fs->logical_inodes[inum].entry_inode.ctime.seconds;
  attr->atime.useconds = fs->logical_inodes[inum].entry_inode.atime.useconds;
  attr->mtime.useconds = fs->logical_inodes[inum].entry_inode.mtime.useconds;
  attr->ctime.useconds = fs->logical_inodes[inum].entry_inode.ctime.useconds;
  return 0;
}

int FS_create_entry(fhandle *new_fh, fattr *attrs, struct timeval *curr_time,
		    int entry_type, int parent_inum)
{
  /* Grab a free logical inode to the new entry */
  int new_inum = fs->first_free_inode;

  assert(new_inum != -1);

  //  fprintf(stderr, "Creating fileid %d inum %d\n", attrs->fileid, new_inum);
  assert(fs->logical_inodes[parent_inum].type == DIR_INODE_TYPE);
  assert(fs->logical_inodes[new_inum].type == FREE_INODE_TYPE);
  assert(entry_type == FILE_INODE_TYPE || entry_type == DIR_INODE_TYPE || entry_type == SYMLINK_INODE_TYPE);
  modified_fs_info();
  fs->first_free_inode = fs->logical_inodes[new_inum].free_inode.next_free;

  /* fill in new inode info */
  modified_inode(new_inum);
  fs->logical_inodes[new_inum].type = entry_type;
  fs->logical_inodes[new_inum].generation_number++;
  fs->logical_inodes[new_inum].entry_inode.parent = parent_inum;
  memcpy(&fs->logical_inodes[new_inum].entry_inode.NFS_handle, new_fh,
	 sizeof(fhandle));
  fs->logical_inodes[new_inum].entry_inode.atime.seconds = curr_time->tv_sec;
  fs->logical_inodes[new_inum].entry_inode.atime.useconds = curr_time->tv_usec;
  fs->logical_inodes[new_inum].entry_inode.mtime.seconds = curr_time->tv_sec;
  fs->logical_inodes[new_inum].entry_inode.mtime.useconds = curr_time->tv_usec;
  fs->logical_inodes[new_inum].entry_inode.ctime.seconds = curr_time->tv_sec;
  fs->logical_inodes[new_inum].entry_inode.ctime.useconds = curr_time->tv_usec;

  /* add mapping between NFS handle and logical inode num */
  add(&fs->logical_inodes[new_inum].entry_inode.NFS_handle, new_inum);

  /*  Add recovery info  */
  fs->fileid[new_inum] = attrs->fileid;

  add_fileid(attrs->fileid, new_inum);

  return new_inum;

}

int FS_set_attr(int file_inum, sattr *set_attr)
{
  
  if (file_inum < 0 || file_inum > NUM_LOGICAL_INODES ||
      fs->logical_inodes[file_inum].type == FREE_INODE_TYPE)
    return -1;

  if (set_attr->atime.seconds != -1)
    fs->logical_inodes[file_inum].entry_inode.atime = set_attr->atime;
  if (set_attr->mtime.seconds != -1)
    fs->logical_inodes[file_inum].entry_inode.mtime = set_attr->mtime;

  return 0;
}


int FS_remove_entry(int file_inum)
{
  assert(file_inum > 0 && file_inum < NUM_LOGICAL_INODES && (fs->logical_inodes[file_inum].type == DIR_INODE_TYPE || fs->logical_inodes[file_inum].type == FILE_INODE_TYPE || fs->logical_inodes[file_inum].type == SYMLINK_INODE_TYPE));

  //  fprintf(stderr, "Removing fileid %d inum %d\n", fs->fileid[file_inum], file_inum);

  remove_fh(&fs->logical_inodes[file_inum].entry_inode.NFS_handle);
  remove_fileid(fs->fileid[file_inum]);

  add_inode_to_free_list(file_inum);

  fs->fileid[file_inum] = 0;

  return 0;
}

int FS_update_file_info(int file_inum, fhandle *new_handle, int new_parent_inum, int fileid, int fsid)
{
  if (fs->logical_inodes[file_inum].type == FREE_INODE_TYPE)
    return -1;
  if (!cmp_fhandles(&fs->logical_inodes[file_inum].entry_inode.NFS_handle,
		    new_handle)) {
    if (remove_fh(&fs->logical_inodes[file_inum].entry_inode.NFS_handle) < 0)
      return -1;
    memcpy(&fs->logical_inodes[file_inum].entry_inode.NFS_handle, new_handle,
	   sizeof(fhandle));
    add(&fs->logical_inodes[file_inum].entry_inode.NFS_handle, file_inum);
  }

  fs->logical_inodes[file_inum].entry_inode.parent = new_parent_inum;
  if (fs->fileid[file_inum] != fileid) {
    remove_fileid(fs->fileid[file_inum]);
    add_fileid(fileid, file_inum);
    fs->fileid[file_inum] = fileid;
  }
  return 0;
}

int FS_update_time_modified(int inum, struct timeval *curr_time)
{
  if (inum < 0 || inum > NUM_LOGICAL_INODES ||
      fs->logical_inodes[inum].type == FREE_INODE_TYPE)
    return -1;
  fs->logical_inodes[inum].entry_inode.atime.seconds = curr_time->tv_sec;
  fs->logical_inodes[inum].entry_inode.atime.useconds = curr_time->tv_usec;
  fs->logical_inodes[inum].entry_inode.mtime.seconds = curr_time->tv_sec;
  fs->logical_inodes[inum].entry_inode.mtime.useconds = curr_time->tv_usec;
  return 0;
}

int FS_init(char *hostname, char *dirname) {

  int i;
  struct timeval t = {10, 0};
  struct fhstatus status;
  enum clnt_stat repval;

  CLIENT *clnt = clnt_create(hostname, MOUNTPROG, MOUNTVERS, "udp");

  fs = (struct FS*)malloc(sizeof(struct FS));

  if (!clnt) {
    fprintf(stderr, "clnt_create failed");
    return -1;
  }
  
  repval = clnt_call(clnt, PROCMNT,
		     (xdrproc_t)xdr_dirpath, (char *)&dirname,
		     (xdrproc_t)xdr_fhstatus, (char *)&status,
		     t);
  
  if (repval != RPC_SUCCESS) {
    perror("rpc mount");
    return -1;
  }
  
  if (status.status != 0) {
    fprintf(stderr, "mount: %s:%s failed, reason given by server: %s\n",
	    hostname, dirname, nfs_strerror(status.status));
    return -1;
  }

  // initialize root stuff:

  // root inode
  fs->logical_inodes[0].type = DIR_INODE_TYPE;
  fs->logical_inodes[0].generation_number = 0;
  memcpy(&fs->logical_inodes[0].entry_inode.NFS_handle, &status.directory,
	 sizeof(fhandle));

  // root entry in the hash table in fh_hash.h
  add(&fs->logical_inodes[0].entry_inode.NFS_handle, 0);

  fs->logical_inodes[0].entry_inode.parent = -1;
  // canonical root time (avoids nondeterminism in this time):
  fs->logical_inodes[0].entry_inode.atime.seconds = ROOT_CREAT_TIME;
  fs->logical_inodes[0].entry_inode.ctime.seconds = ROOT_CREAT_TIME;
  fs->logical_inodes[0].entry_inode.mtime.seconds = ROOT_CREAT_TIME;
  fs->logical_inodes[0].entry_inode.atime.useconds = 0;
  fs->logical_inodes[0].entry_inode.ctime.useconds = 0;
  fs->logical_inodes[0].entry_inode.mtime.useconds = 0;
  //  fs->fsid[0] = 0;
  fs->fileid[0] = 0;


  // initialize remaining logical inodes
  for(i=1; i<NUM_LOGICAL_INODES; i++) {
    fs->logical_inodes[i].type = FREE_INODE_TYPE;
    fs->logical_inodes[i].generation_number = 0;
    *((long *)&fs->logical_inodes[i].entry_inode.NFS_handle) = 1;
    fs->logical_inodes[i].free_inode.next_free = i+1;
    fs->fileid[i] = 0;
  }

  fs->logical_inodes[NUM_LOGICAL_INODES-1].free_inode.next_free = -1;
  fs->first_free_inode = 1;

  clnt_freeres(clnt, (xdrproc_t)xdr_fhstatus, (char *)&status);
  clnt_destroy(clnt);

#ifdef STATS
  init_stats();
#endif

  return NUM_LOGICAL_INODES + 1;
}

int cmp_fhandles(fhandle *fh1, fhandle *fh2)
{
  int i;
  for (i=0; i<FHSIZE; i++)
    if (fh1->data[i]!=fh2->data[i])
      return 0;
  return 1;
}


/* Logical state mappings */

#define MIN_INODES_PAGE 0
#define MAX_INODES_PAGE (MIN_INODES_PAGE + NUM_LOGICAL_INODES - 1)

#define FS_ROOT_PAGE (MAX_INODES_PAGE + 1)

#define MAX_STATE FS_ROOT_PAGE

#define UNLNK_DIR_NAME "unlinked-dir"
#define UNLNK_FILE_NAME "unlinked-file-"

typedef struct {
  int type;
  int gen;
  int next_free;
} free_obj;

typedef struct {
  int type;
  int gen;
  int parent;
  fattr attr;
} entry_obj;

void fill_attr(fattr *to, fattr *from, int inum)
{
  memcpy(to, from, sizeof(fattr));
  to->rdev = 0;
  to->fileid = inum;
  to->blocksize = Block_size;
  to->blocks = (from->size >> 12) + 1;
  to->fsid = BYZ_FSID;
#ifdef OMMIT_UID_GID
  to->uid = 0;
  to->gid = 0;
#endif
  to->atime = fs->logical_inodes[inum].entry_inode.atime;
  to->ctime = fs->logical_inodes[inum].entry_inode.ctime;
  to->mtime = fs->logical_inodes[inum].entry_inode.mtime;
}

int recover_entry(int n) /* XXX - called recursively: should detect cyles */
{
  readdirargs rda;
  readdirres rdr;
  diropargs da;
  diropres dr;
  entry *ptr;
  int innum;
  nfscookie *cookie = NULL;

  if (n<0 || n>=NUM_LOGICAL_INODES || fs->logical_inodes[n].type == FREE_INODE_TYPE)
    return -1;

  if (*((long *)&fs->logical_inodes[n].entry_inode.NFS_handle) == 0) {
    if (fs->logical_inodes[n].type != DIR_INODE_TYPE)
      return recover_entry(fs->logical_inodes[n].entry_inode.parent);
    else {
      if (recover_entry(fs->logical_inodes[n].entry_inode.parent) < 0)
	return -1;
    }
  }

  assert(fs->logical_inodes[n].type == DIR_INODE_TYPE);

  rdr.readdirok.eof = FALSE;
  while (!rdr.readdirok.eof) {
    memset(&rda, 0, sizeof(rda));
    memset(&rdr, 0, sizeof(rdr));
    memcpy(&rda.dir, &fs->logical_inodes[n].entry_inode.NFS_handle, sizeof(fhandle));
    rda.count = NFS_MAXDATA;
    if (cookie) {
      memcpy(&rda.cookie, cookie, sizeof(nfscookie));
      free(cookie);
    }
    perform_RPC_call(NFSPROC_READDIR, (char *)&rda, (char *)&rdr);
    if (rdr.status != NFS_OK) {
      fprintf(stderr, "Error %d in getpage read dir\n", rdr.status);
      free_RPC_res(NFSPROC_READDIR, (char *)&rdr);
      return -1;
    }
    ptr = rdr.readdirok.entries;
    while (ptr) {
      if (strcmp(ptr->name, ".") && strcmp(ptr->name, "..")) {
	memset(&da, 0, sizeof(da));
	memset(&dr, 0, sizeof(dr));
	da.name = ptr->name;
	memcpy(&da.dir, &fs->logical_inodes[n].entry_inode.NFS_handle, sizeof(fhandle));
	perform_RPC_call(NFSPROC_LOOKUP, (char*) &da, (char *)&dr);
	if (dr.status != NFS_OK) {
	  fprintf(stderr, "Error %d in getpage dir lookup\n", rdr.status);
	  free_RPC_res(NFSPROC_READDIR, (char *)&rdr);
	  free_RPC_res(NFSPROC_LOOKUP, (char *)&dr);
	  return -1;
	}
	/*	for(innum=0; innum<NUM_LOGICAL_INODES; innum++) {
	  if (dr.diropok.attributes.fileid == fs->fileid[innum] &&
	      dr.diropok.attributes.fsid == fs->fsid[innum])
	    break;
	}
	if (innum == NUM_LOGICAL_INODES) {
	  fprintf(stderr, "Could not find fileid during %d recovery\n", n);
	  return -1;
	  } */

	if ((innum = find_fileid(dr.diropok.attributes.fileid)) == -1) {
	  fprintf(stderr, "Could not find inum for %s (%d) during %d recovery\n", ptr->name, ptr->fileid, n);
	  return -1;
	}

	/* Reconstruct conformance wrapper's volatile state */
	if (*((long *)&fs->logical_inodes[innum].entry_inode.NFS_handle) == 0) {
	  //	  fprintf(stderr, "RCNSTR> %s id %d inum %d \t", da.name, fs->fileid[innum], innum);
	  memcpy(&fs->logical_inodes[innum].entry_inode.NFS_handle, &dr.diropok.file, sizeof(fhandle));
	  add(&fs->logical_inodes[innum].entry_inode.NFS_handle, innum);
	  fs->fileid[innum] = dr.diropok.attributes.fileid;
	}
	free_RPC_res(NFSPROC_LOOKUP, (char *)&dr);
      }
      if (!rdr.readdirok.eof && !ptr->nextentry) {
	cookie = (nfscookie *)malloc(sizeof(nfscookie));
	memcpy(cookie, &ptr->cookie, sizeof(nfscookie));
      }
      ptr = ptr->nextentry;
    }
  }
  free_RPC_res(NFSPROC_READDIR, (char *)&rdr);
  return 0;
}

int get_obj(int n, char **obj)
{
  attrstat attrrep;
  readargs ra;
  readres rr;
  readdirargs rda;
  readdirres rdr;
  diropargs da;
  diropres dr;
  readlinkres rlr;
  entry *ptr, *prev, *entry_list = NULL;
  int read = 0, entries = 0, i = 0, innum;
  struct direntry *de;
  nfscookie *cookie = NULL;

  if (n < 0 || n > MAX_STATE)
    return -1;
  if (n >= MIN_INODES_PAGE && n <= MAX_INODES_PAGE) { 

    if (fs->logical_inodes[n].type != FREE_INODE_TYPE && *((long *)&fs->logical_inodes[n].entry_inode.NFS_handle) == 0) {
#ifdef STATS
      start_counter(RECOVER_WRAPPER_STATE + num_recov - 1);
#endif
      recover_entry(n);
#ifdef STATS
      stop_counter(RECOVER_WRAPPER_STATE + num_recov - 1);
#endif
    }

    switch (fs->logical_inodes[n].type) {
    case FREE_INODE_TYPE:
#ifdef STATS
      num_get_free++;
      start_counter(GET_FREE);
#endif
      *obj = (char *)malloc(sizeof(free_obj));
      ((free_obj*)(*obj))->type = FREE_INODE_TYPE;
      ((free_obj*)(*obj))->gen = fs->logical_inodes[n].generation_number;
      ((free_obj*)(*obj))->next_free = fs->logical_inodes[n].free_inode.next_free;
#ifdef STATS
      stop_counter(GET_FREE);
#endif
      return sizeof(free_obj);

    case FILE_INODE_TYPE:
#ifdef STATS
      num_get_file++;
      start_counter(GET_FILE_GET_ATTR);
#endif
      memset(&attrrep, 0, sizeof(attrrep));
      perform_RPC_call(NFSPROC_GETATTR, (char *)&fs->logical_inodes[n].entry_inode.NFS_handle, (char *)&attrrep);
#ifdef STATS
      stop_counter(GET_FILE_GET_ATTR);
#endif
      if (attrrep.status != NFS_OK) {
	fprintf(stderr, "Error %d in getpage file get attr\n", attrrep.status);
	free_RPC_res(NFSPROC_GETATTR, (char *)&attrrep);
	return 0;
      }
#ifdef STATS
      start_counter(GET_FILE_RESULT_OBJ);
#endif
      assert(attrrep.attributes.type == NFREG);
      *obj = (char *)malloc(sizeof(entry_obj) + attrrep.attributes.size);
      ((entry_obj*)(*obj))->type = FILE_INODE_TYPE;
      ((entry_obj*)(*obj))->gen = fs->logical_inodes[n].generation_number;
      fill_attr(&((entry_obj*)(*obj))->attr, &attrrep.attributes, n);
      ((entry_obj*)(*obj))->parent = fs->logical_inodes[n].entry_inode.parent;
#ifdef STATS
      stop_counter(GET_FILE_RESULT_OBJ);
      start_counter(GET_FILE_READ_CONTENTS);
#endif

      /* Read file's contents and copy it to the object */
      while (read < attrrep.attributes.size) {
      
	memset(&ra, 0, sizeof(ra));
	memset(&rr, 0, sizeof(rr));
	memcpy(&ra.file, &fs->logical_inodes[n].entry_inode.NFS_handle, sizeof(fhandle));
	ra.offset = read;
	ra.count = min(NFS_MAXDATA, attrrep.attributes.size - read);
	
	perform_RPC_call(NFSPROC_READ, (char *)&ra, (char *)&rr);
	if (rr.status != NFS_OK) {
	  fprintf(stderr, "Error %d in getpage file read\n", rr.status);
	  free_RPC_res(NFSPROC_GETATTR, (char *)&attrrep);
	  free_RPC_res(NFSPROC_READ, (char *)&rr);
	  return 0;
	}
	
	memcpy(*obj + sizeof(entry_obj) + read, rr.reply.data.data_val, rr.reply.data.data_len);
	read += rr.reply.data.data_len;
	free_RPC_res(NFSPROC_READ, (char *)&rr);
      }
      read = attrrep.attributes.size;
      free_RPC_res(NFSPROC_GETATTR, (char *)&attrrep);
#ifdef STATS
      stop_counter(GET_FILE_READ_CONTENTS);
#endif
      return (sizeof(entry_obj) + read);

    case DIR_INODE_TYPE:
#ifdef STATS
      num_get_dir++;
      start_counter(GET_DIR_GET_ATTR);
#endif
      memset(&attrrep, 0, sizeof(attrrep));
      perform_RPC_call(NFSPROC_GETATTR, (char *)&fs->logical_inodes[n].entry_inode.NFS_handle, (char *)&attrrep);
#ifdef STATS
      stop_counter(GET_DIR_GET_ATTR);
#endif
      if (attrrep.status != NFS_OK) {
	fprintf(stderr, "Error %d in getpage file get attr\n", attrrep.status);
	free_RPC_res(NFSPROC_GETATTR, (char *)&attrrep);
	return 0;
      }
      assert(attrrep.attributes.type == NFDIR);

#ifdef STATS
      start_counter(GET_DIR_READ_DIR);
#endif
      rdr.readdirok.eof = FALSE;
      while (!rdr.readdirok.eof) {
	memset(&rda, 0, sizeof(rda));
	memset(&rdr, 0, sizeof(rdr));
	memcpy(&rda.dir, &fs->logical_inodes[n].entry_inode.NFS_handle, sizeof(fhandle));
	rda.count = NFS_MAXDATA;
	if (cookie) {
	  memcpy(&rda.cookie, cookie, sizeof(nfscookie));
	  free(cookie);
	}
	perform_RPC_call(NFSPROC_READDIR, (char *)&rda, (char *)&rdr);
	if (rdr.status != NFS_OK) {
	  fprintf(stderr, "Error %d in getpage read dir\n", rdr.status);
	  free_RPC_res(NFSPROC_GETATTR, (char *)&attrrep);
	  free_RPC_res(NFSPROC_READDIR, (char *)&rdr);
	  return 0;
	}
	if (!entry_list) {
	  ptr = entry_list = rdr.readdirok.entries;
	}
	else {
	  prev->nextentry = ptr = rdr.readdirok.entries;
	}
	while (ptr) {
	  if (strcmp(ptr->name, ".") && strcmp(ptr->name, ".."))
	    entries++;
	  prev = ptr;
	  if (!rdr.readdirok.eof && !ptr->nextentry) {
	    cookie = (nfscookie *)malloc(sizeof(nfscookie));
	    memcpy(cookie, &ptr->cookie, sizeof(nfscookie));
	  }
	  ptr = ptr->nextentry;
	}
      }
#ifdef STATS
      stop_counter(GET_DIR_READ_DIR);
      start_counter(GET_DIR_RESULT_OBJ);
#endif

      *obj = (char *)malloc(sizeof(entry_obj) + entries * sizeof(struct direntry));
      ((entry_obj*)(*obj))->type = DIR_INODE_TYPE;
      ((entry_obj*)(*obj))->gen = fs->logical_inodes[n].generation_number;
      fill_attr(&((entry_obj*)(*obj))->attr, &attrrep.attributes, n);
      ((entry_obj*)(*obj))->parent = fs->logical_inodes[n].entry_inode.parent;
      de = (struct direntry *) ((*obj) + sizeof(entry_obj));
      ptr = entry_list;
      while (ptr) {
	if (strcmp(ptr->name, ".") && strcmp(ptr->name, "..")) {
	  memset(de[i].name, 0, MAX_NAME_LEN+1);
	  strncpy(de[i].name, ptr->name, MAX_NAME_LEN+1);
	  de[i].fileid = find_fileid(ptr->fileid);
	  if (de[i].fileid == -1) {
	    fprintf(stderr, "Can't find fhandle. NON-EMPTY FS AT INIT?\n");
	  }
	  
	  i++;
	}
	ptr = ptr->nextentry;
      }

#ifdef ORDER_DIR_ENTRIES
      /* Sorting the entries - XXX this could bw slow */
      qsort(de, entries, sizeof(struct direntry), cmp_dir_entries); 
#endif

      assert(i == entries);
      ((entry_obj*)(*obj))->attr.size = entries;
      ((entry_obj*)(*obj))->attr.blocks = (entries >> 4) + 1;
#ifdef STATS
      stop_counter(GET_DIR_RESULT_OBJ);
#endif
      free_RPC_res(NFSPROC_GETATTR, (char *)&attrrep);
      free_RPC_res(NFSPROC_READDIR, (char *)&rdr);
      return (sizeof(entry_obj) + entries * sizeof(struct direntry));

    case SYMLINK_INODE_TYPE:
      memset(&attrrep, 0, sizeof(attrrep));
      perform_RPC_call(NFSPROC_GETATTR, (char *)&fs->logical_inodes[n].entry_inode.NFS_handle, (char *)&attrrep);
      if (attrrep.status != NFS_OK) {
	fprintf(stderr, "Error %d in getpage file get attr\n", attrrep.status);
	free_RPC_res(NFSPROC_GETATTR, (char *)&attrrep);
	return 0;
      }
      assert(attrrep.attributes.type == NFLNK);
      *obj = (char *)malloc(sizeof(entry_obj) + MAX_NAME_LEN + 1);
      ((entry_obj*)(*obj))->type = SYMLINK_INODE_TYPE;
      ((entry_obj*)(*obj))->gen = fs->logical_inodes[n].generation_number;
      fill_attr(&((entry_obj*)(*obj))->attr, &attrrep.attributes, n);
      ((entry_obj*)(*obj))->parent = fs->logical_inodes[n].entry_inode.parent;

      /* Read SYMBOLIC LINK's contents and copy it to the object */
      memset(&rlr, 0, sizeof(rlr));
      perform_RPC_call(NFSPROC_READLINK, (char *)&fs->logical_inodes[n].entry_inode.NFS_handle, (char *)&rlr);
      if (rlr.status != NFS_OK) {
	fprintf(stderr, "Error %d in getpage read symlink\n", rlr.status);
	free_RPC_res(NFSPROC_GETATTR, (char *)&attrrep);
	free_RPC_res(NFSPROC_READLINK, (char *)&rlr);
	return 0;
      }
      memset(*obj + sizeof(entry_obj), 0, MAX_NAME_LEN + 1);
      strncpy(*obj + sizeof(entry_obj), rlr.data, MAX_NAME_LEN + 1);
      free_RPC_res(NFSPROC_GETATTR, (char *)&attrrep);
      free_RPC_res(NFSPROC_READLINK, (char *)&rlr);
      return (sizeof(entry_obj) + MAX_NAME_LEN + 1);
    }
    fprintf(stderr, "UH - OH. Unrecognized obj type\n");
  }
  if (n == FS_ROOT_PAGE) { 
    *obj = (char *)malloc(sizeof(int));
    *((int *)(*obj)) = fs->first_free_inode;
    return sizeof(int);
  }
  fprintf(stderr, "UH - OH. Unrecognized obj number\n");
  return -1;
}

int set_attrs(int inum, fattr *attrs)
{
  sattrargs sa;
  attrstat attr;

  fs->logical_inodes[inum].entry_inode.atime = attrs->atime;
  fs->logical_inodes[inum].entry_inode.mtime = attrs->mtime;
  fs->logical_inodes[inum].entry_inode.ctime = attrs->ctime;

  memset(&sa, 0, sizeof(sa));
  memset(&attr, 0, sizeof(attr));
  memcpy(&sa.file, &fs->logical_inodes[inum].entry_inode.NFS_handle, sizeof(fhandle));
  sa.attributes.mode = attrs->mode;
#ifdef OMMIT_UID_GID
  sa.attributes.uid = -1;
  sa.attributes.gid = -1;
#else
  sa.attributes.uid = attrs->uid;
  sa.attributes.gid = attrs->gid;
#endif
  sa.attributes.size = attrs->size;
  sa.attributes.atime.seconds = -1;
  sa.attributes.mtime.seconds = -1;
  sa.attributes.atime.useconds = -1;
  sa.attributes.mtime.useconds = -1;
  perform_RPC_call(NFSPROC_SETATTR, (char *)&sa, (char *)&attr);
  if (attr.status == NFS_OK) {
    //    fs->fileid[inum] = attr.attributes.fileid;
    free_RPC_res(NFSPROC_SETATTR, (char *)&attr);
    return 0;
  }
  free_RPC_res(NFSPROC_SETATTR, (char *)&attr);
  fprintf(stderr, "Error %d in setattr for file %d\n", attr.status, inum);
  return -1;
}

void write_file_contents(int inum, int size, char *contents)
{
  writeargs wa;
  attrstat as;
  int written = 0;
  assert(fs->logical_inodes[inum].type == FILE_INODE_TYPE);

  while (written < size) {
    memset(&wa, 0, sizeof(wa));
    memset(&as, 0, sizeof(as));
    memcpy(&wa.file, &fs->logical_inodes[inum].entry_inode.NFS_handle, sizeof(fhandle));
    wa.offset = written;
    wa.data.data_len = min(NFS_MAXDATA, size - written);
    wa.data.data_val = contents + written;
    perform_RPC_call(NFSPROC_WRITE, (char *)&wa, (char *)&as);
    if (as.status != NFS_OK) {
      fprintf(stderr, "Error %d in put's NFSPROC_WRITE\n", as.status);
      free_RPC_res(NFSPROC_WRITE, (char *)&as);
      return;
    }
    written += min(NFS_MAXDATA, size - written);
    free_RPC_res(NFSPROC_WRITE, (char *)&as);
  }
}

int create_file_or_dir(fhandle *parent_fh, fattr *attrs, char *name, fhandle *new_fh, unsigned int *new_fileid, int type)
{
  createargs ca;
  diropres dr;

  memset(&ca, 0, sizeof(ca));
  memset(&dr, 0, sizeof(dr));
  memcpy(&ca.where.dir, parent_fh, sizeof(fhandle));
  ca.where.name = name;
  ca.attributes.mode = attrs->mode;
  ca.attributes.uid = attrs->uid;
  ca.attributes.gid = attrs->gid;
  ca.attributes.size = attrs->size;
  ca.attributes.atime = attrs->atime;
  ca.attributes.mtime = attrs->mtime;
  if (type == DIR_INODE_TYPE)
    perform_RPC_call(NFSPROC_MKDIR, (char *)&ca, (char *)&dr);
  else if (type == FILE_INODE_TYPE)
    perform_RPC_call(NFSPROC_CREATE, (char *)&ca, (char *)&dr);
  else
    return -1;
  if (dr.status != NFS_OK) {
    fprintf(stderr, "Error in creating file/dir %s\n", name);
    free_RPC_res(NFSPROC_CREATE, (char *)&dr);
    return -1;
  }
  memcpy(new_fh, &dr.diropok.file, sizeof(fhandle));
  *new_fileid = dr.diropok.attributes.fileid;
  free_RPC_res(NFSPROC_CREATE, (char *)&dr);
  return 0;
}

int is_element(int inum, int *vector, int total)
{
  int i;
  for (i=0; i<total; i++)
    if (vector[i]==inum)
      return i;
  return -1;
}

void remove_NFS_file(fhandle *dir, char *name)
{
  diropargs da;
  nfsstat res;

  memset(&res, 0, sizeof(res));
  memcpy(&da.dir, dir, sizeof(fhandle));
  da.name = name;
  perform_RPC_call(NFSPROC_REMOVE, (char *)&da, (char *)&res);
  if (res!=NFS_OK)
    fprintf(stderr, "Error %d removing %s\n", res, name);
  free_RPC_res(NFSPROC_REMOVE, (char *)&res);
}

void remove_NFS_dir(fhandle *parent_dir, char *name, fhandle *dir)
{
  readdirargs rda;
  readdirres rdr;
  diropargs da;
  diropres dr;
  nfsstat stat;
  entry *entry_ptr;
  nfscookie *cookie = NULL;

  /* 1. Read dir contents */
  rdr.readdirok.eof = FALSE;
  while (!rdr.readdirok.eof) {
    memset(&rda, 0, sizeof(readdirargs));
    memset(&rdr, 0, sizeof(readdirres));
    rda.count = NFS_MAXDATA;
    memcpy(&rda.dir, dir, sizeof(fhandle));
    if (cookie) {
      memcpy(&rda.cookie, cookie, sizeof(nfscookie));
      free(cookie);
    }
    perform_RPC_call(NFSPROC_READDIR, (char *)&rda, (char *)&rdr);
    if (rdr.status != NFS_OK) {
      fprintf(stderr, "Error in put while reading dir. Error %d\n", rdr.status);
      free_RPC_res(NFSPROC_READDIR, (char *)&rdr);
      return;
    }
    /* 2. Clean up dir */
    entry_ptr = rdr.readdirok.entries;
    while (entry_ptr) {
      memset(&da, 0, sizeof(diropargs));
      memset(&dr, 0, sizeof(diropres));
      memcpy(&da.dir, dir, sizeof(fhandle));
      da.name = entry_ptr->name;
      if (strcmp(entry_ptr->name, ".") && strcmp(entry_ptr->name, "..")) {
	perform_RPC_call(NFSPROC_LOOKUP, (char *)&da, (char *)&dr);
	if (dr.status != NFS_OK) {
	  fprintf(stderr, "Error %d in lkup file %s in dir %s\n", dr.status, entry_ptr->name, name);
	  free_RPC_res(NFSPROC_READDIR, (char *)&rdr);
	  free_RPC_res(NFSPROC_LOOKUP, (char *)&dr);
	  return;
	}
	if (dr.diropok.attributes.type == NFREG || dr.diropok.attributes.type == NFLNK)
	  remove_NFS_file(dir, entry_ptr->name);
	else if (dr.diropok.attributes.type == NFDIR)
	  remove_NFS_dir(dir, entry_ptr->name, &dr.diropok.file);
	else
	  fprintf(stderr, "Warning: unknown entry type");
	free_RPC_res(NFSPROC_LOOKUP, (char *)&dr);
	remove_fh(&dr.diropok.file);
	remove_fileid(dr.diropok.attributes.fileid);
      }
      if (!rdr.readdirok.eof && !entry_ptr->nextentry) {
	cookie = (nfscookie *)malloc(sizeof(nfscookie));
	memcpy(cookie, &entry_ptr->cookie, sizeof(nfscookie));
	free_RPC_res(NFSPROC_READDIR, (char *)&rdr);
	entry_ptr = NULL;
      }
      else
	entry_ptr = entry_ptr->nextentry;
    }
  }
  /* 3. Remove dir */
  memset(&da, 0, sizeof(da));
  memcpy(&da.dir, parent_dir, sizeof(fhandle));
  da.name = name;
  perform_RPC_call(NFSPROC_RMDIR, (char *)&da, (char *)&stat);
  if (stat!=NFS_OK)
    fprintf(stderr, "Error %d removing directory %s\n", stat, name);  
  free_RPC_res(NFSPROC_READDIR, (char *)&rdr);
  free_RPC_res(NFSPROC_RMDIR, (char *)&stat);
}

void create_symlink()
{
  fprintf(stderr, "Not implemented yet!\n");
}


void check_entry_in_dir(int dir_inum, int pos, char *entry_name, int num_objs, int *sizes, int *indices, char **objs)
{
  diropargs da;
  diropres dr;
  struct direntry * de = (struct direntry *)(objs[pos] + sizeof(entry_obj));
  int entrypos, i, num_entries = (sizes[pos] - sizeof(entry_obj)) / sizeof(struct direntry);

  memset(&da, 0, sizeof(diropargs));
  memset(&dr, 0, sizeof(diropres));
  memcpy(&da.dir, &fs->logical_inodes[dir_inum].entry_inode.NFS_handle, sizeof(fhandle));
  da.name = entry_name;
  perform_RPC_call(NFSPROC_LOOKUP, (char *)&da, (char *)&dr);
  if (dr.status != NFS_OK) {
    fprintf(stderr, "Error %d in lkup file %s in dir %d\n", dr.status, entry_name, dir_inum);
    free_RPC_res(NFSPROC_LOOKUP, (char *)&dr);
    return;
  }

  for (i=0; i<num_entries; i++) {
    if (!strncmp(entry_name, de[i].name, MAX_NAME_LEN+1)) {
      /* Matches, check if inum and type are ok, if so return */

      if (de[i].fileid != find(&dr.diropok.file)) {
	// must update inum <-> fhandle

	remove_fh(&dr.diropok.file);
	remove_fileid(dr.diropok.attributes.fileid);
	add_fileid(dr.diropok.attributes.fileid, de[i].fileid);
	fs->fileid[de[i].fileid] = dr.diropok.attributes.fileid;
	memcpy(&fs->logical_inodes[de[i].fileid].entry_inode.NFS_handle, &dr.diropok.file, sizeof(fhandle)); 
	add(&fs->logical_inodes[de[i].fileid].entry_inode.NFS_handle, de[i].fileid);
      }

      if ((entrypos = is_element(de[i].fileid, indices, num_objs)) < 0) {
	free_RPC_res(NFSPROC_LOOKUP, (char *)&dr);
	return;
      }
      if ( (((entry_obj *)objs[entrypos])->type == FILE_INODE_TYPE &&
	    dr.diropok.attributes.type == NFREG) ||
	   (((entry_obj *)objs[entrypos])->type == DIR_INODE_TYPE &&
	    dr.diropok.attributes.type == NFDIR) ||
	   (((entry_obj *)objs[entrypos])->type == SYMLINK_INODE_TYPE &&
	    dr.diropok.attributes.type == NFLNK) ) {
	free_RPC_res(NFSPROC_LOOKUP, (char *)&dr);
	return;
      }
      /* Type mismatch: remove entry and create correct one */
      if (dr.diropok.attributes.type == NFDIR)
	remove_NFS_dir(&fs->logical_inodes[dir_inum].entry_inode.NFS_handle, entry_name, &dr.diropok.file);
      else
	remove_NFS_file(&fs->logical_inodes[dir_inum].entry_inode.NFS_handle, entry_name);
      remove_fh(&fs->logical_inodes[de[i].fileid].entry_inode.NFS_handle);
      remove_fileid(fs->fileid[de[i].fileid]);
      switch(((entry_obj *)objs[entrypos])->type) {
      case FILE_INODE_TYPE:
	create_file_or_dir(&fs->logical_inodes[dir_inum].entry_inode.NFS_handle, &((entry_obj *)objs[entrypos])->attr, entry_name, &fs->logical_inodes[de[i].fileid].entry_inode.NFS_handle, &fs->fileid[de[i].fileid], FILE_INODE_TYPE);
	add_fileid(fs->fileid[de[i].fileid], de[i].fileid);
	add(&fs->logical_inodes[de[i].fileid].entry_inode.NFS_handle, de[i].fileid);
	break;
      case DIR_INODE_TYPE:
	create_file_or_dir(&fs->logical_inodes[dir_inum].entry_inode.NFS_handle, &((entry_obj *)objs[entrypos])->attr, entry_name, &fs->logical_inodes[de[i].fileid].entry_inode.NFS_handle, &fs->fileid[de[i].fileid], DIR_INODE_TYPE);
	add_fileid(fs->fileid[de[i].fileid], de[i].fileid);
	add(&fs->logical_inodes[de[i].fileid].entry_inode.NFS_handle, de[i].fileid);
	break;
      case SYMLINK_INODE_TYPE:
	create_symlink( );
	break;
      }
      free_RPC_res(NFSPROC_LOOKUP, (char *)&dr);
      return;
    }
  }

  /* Entry not found in correct logical state, must erase it.
     If it is a dir, erase it recursively. */
  if (dr.diropok.attributes.type == NFDIR)
    remove_NFS_dir(&fs->logical_inodes[dir_inum].entry_inode.NFS_handle, entry_name, &dr.diropok.file);
  else
    remove_NFS_file(&fs->logical_inodes[dir_inum].entry_inode.NFS_handle, entry_name); /* XXX Don't forget to clean up mappings in hash table */
  remove_fh(&dr.diropok.file);
  remove_fileid(dr.diropok.attributes.fileid);
  free_RPC_res(NFSPROC_LOOKUP, (char *)&dr);
}

void create_entry_in_dir(char *entry_name, int inum, int dir_inum, int num_objs, int *sizes, int *indices, char **objs)
{
  int entrypos = is_element(inum, indices, num_objs);
  if (entrypos == -1) {
    fprintf(stderr, "Could not find new entry\n");
    return;
  }
  switch(((entry_obj *)objs[entrypos])->type) {
  case FILE_INODE_TYPE:
    create_file_or_dir(&fs->logical_inodes[dir_inum].entry_inode.NFS_handle, &((entry_obj *)objs[entrypos])->attr, entry_name, &fs->logical_inodes[inum].entry_inode.NFS_handle, &fs->fileid[inum], FILE_INODE_TYPE);
    /* Update inode info */
    add_fileid(fs->fileid[inum], inum);
    add(&fs->logical_inodes[inum].entry_inode.NFS_handle, inum);
    // The rest is filled when the file is processed
    break;
  case DIR_INODE_TYPE:
    create_file_or_dir(&fs->logical_inodes[dir_inum].entry_inode.NFS_handle, &((entry_obj *)objs[entrypos])->attr, entry_name, &fs->logical_inodes[inum].entry_inode.NFS_handle, &fs->fileid[inum], DIR_INODE_TYPE);
    /* Update inode info */
    add_fileid(fs->fileid[inum], inum);
    add(&fs->logical_inodes[inum].entry_inode.NFS_handle, inum);
    // The rest is filled when the file is processed
    break;
  case SYMLINK_INODE_TYPE:
    create_symlink( );
    break;
  }
  
}

int entry_in_dirlist(char *name, entry *entry_list)
{
  while (entry_list) {
    if (!strncmp(entry_list->name, name, MAX_NAME_LEN+1))
      return 1;
    entry_list = entry_list->nextentry;
  }
  return 0;
}

void update_dir(int dir_inum, int *uptodate, int *num_uptodate, int num_objs, int *sizes, int *indices, char **objs)
{
  int pos, num_entries, i;
  readdirargs rda;
  readdirres rdr;
  entry *entry_ptr, *prev, *entry_list = NULL;
  struct direntry *de;
  nfscookie *cookie = NULL;

#ifdef STATS
  start_counter(PUT_SCAN_UPTODATE);
#endif
  if ((dir_inum == -1) ||
      (is_element(dir_inum, uptodate, *num_uptodate) != -1) ||
      ((pos = is_element(dir_inum, indices, num_objs)) == -1)) {
#ifdef STATS
    stop_counter(PUT_SCAN_UPTODATE);
#endif
    return;
  }
#ifdef STATS
  stop_counter(PUT_SCAN_UPTODATE);
#endif
  /* First, we update the parent dir */
  update_dir(((entry_obj *)objs[pos])->parent, uptodate, num_uptodate, num_objs, sizes, indices, objs);

#ifdef STATS
  start_counter(PUT_READDIR);
#endif
  num_entries = (sizes[pos] - sizeof(entry_obj)) / sizeof(struct direntry);

  /* Read dir's contents (my own current state) */
  rdr.readdirok.eof = FALSE;
  while (!rdr.readdirok.eof) {
    memset(&rda, 0, sizeof(readdirargs));
    memset(&rdr, 0, sizeof(readdirres));
    rda.count = NFS_MAXDATA;
    memcpy(&rda.dir, &fs->logical_inodes[dir_inum].entry_inode.NFS_handle, sizeof(fhandle));
    if (cookie) {
      memcpy(&rda.cookie, cookie, sizeof(nfscookie));
      free(cookie);
    }
    perform_RPC_call(NFSPROC_READDIR, (char *)&rda, (char *)&rdr);
    if (rdr.status != NFS_OK) {
      free_RPC_res(NFSPROC_READDIR, (char *)&rdr);
      fprintf(stderr, "Error in put while reading dir. Error %d\n", rdr.status);
      return;
    }
    if (!entry_list)
      entry_list = rdr.readdirok.entries;
    else
      prev->nextentry = rdr.readdirok.entries;
    entry_ptr = rdr.readdirok.entries;
    /* Check if existing entries are ok */
    while (entry_ptr) {
      /* check entry */
      if (strcmp(entry_ptr->name, ".") && strcmp(entry_ptr->name, "..")) {
#ifdef STATS
	stop_counter(PUT_READDIR);
	start_counter(PUT_CHECK_EXISTING_ENTRIES);
#endif
	check_entry_in_dir(dir_inum, pos, entry_ptr->name, num_objs, sizes, indices, objs);
#ifdef STATS
	stop_counter(PUT_CHECK_EXISTING_ENTRIES);
	start_counter(PUT_READDIR);
#endif
      }
      if (!rdr.readdirok.eof && !entry_ptr->nextentry) {
	cookie = (nfscookie *)malloc(sizeof(nfscookie));
	memcpy(cookie, &entry_ptr->cookie, sizeof(nfscookie));
	prev = entry_ptr;
      }
      entry_ptr = entry_ptr->nextentry;
    }
  }
#ifdef STATS
  stop_counter(PUT_READDIR);
  start_counter(PUT_CREATE_MISSING_ENTRIES);
#endif
  /* Create missing entries */
  de = (struct direntry *)(objs[pos] + sizeof(entry_obj));
  for (i=0; i<num_entries; i++)
    if (!entry_in_dirlist(de[i].name, entry_list))
      create_entry_in_dir(de[i].name, de[i].fileid, dir_inum, num_objs, sizes, indices, objs);
#ifdef STATS
  stop_counter(PUT_CREATE_MISSING_ENTRIES);
#endif

  /* Add dir to uptodate list */
  uptodate[(*num_uptodate)++] = dir_inum;
  free_RPC_res(NFSPROC_READDIR, (char *)&rdr);

}

void put_objs(int num_objs, int *sizes, int *indices, char **objs)
{
  int i, obj_type, gen, num_uptodate;

  int *uptodate = (int *)malloc(num_objs*sizeof(int));

#ifdef STATS
  num_put++;
#endif

  for (i=0; i<num_objs; i++)
    uptodate[i] = -1;
  num_uptodate = 0;

  for (i=0; i<num_objs; i++) {
    if (indices[i] >= MIN_INODES_PAGE && indices[i] <= MAX_INODES_PAGE) {
      obj_type = ((entry_obj *)objs[i])->type;
      gen = ((entry_obj *)objs[i])->gen;
      switch (obj_type) {
      case FREE_INODE_TYPE:
#ifdef STATS
	num_put_free++;
#endif
	/* If it's not free it will be removed when parent dir is processed */
	fs->logical_inodes[indices[i]].type = FREE_INODE_TYPE;
	fs->logical_inodes[indices[i]].generation_number = gen;
	fs->logical_inodes[indices[i]].free_inode.next_free = ((free_obj *)objs[i])->next_free;
	fs->fileid[indices[i]] = 0;
	break;
      case FILE_INODE_TYPE:
	/* Update file attrs and contents */
#ifdef STATS
	num_put_file++;
#endif
	update_dir(((entry_obj *)objs[i])->parent, uptodate, &num_uptodate, num_objs, sizes, indices, objs);
#ifdef STATS
	start_counter(PUT_SATTR_WRITE_FILE);
#endif
	fs->logical_inodes[indices[i]].type = FILE_INODE_TYPE;
	fs->logical_inodes[indices[i]].generation_number = ((entry_obj *)objs[i])->gen;
	fs->logical_inodes[indices[i]].entry_inode.parent = ((entry_obj *)objs[i])->parent;
	set_attrs(indices[i], &((entry_obj *)objs[i])->attr);
	write_file_contents(indices[i], ((entry_obj *)objs[i])->attr.size, objs[i] + sizeof(entry_obj));
	stop_counter(PUT_SATTR_WRITE_FILE);
	break;
      case SYMLINK_INODE_TYPE:
	/* Update attrs. XXX If contents mismatch must go to parent dir */
	update_dir(((entry_obj *)objs[i])->parent, uptodate, &num_uptodate, num_objs, sizes, indices, objs);
	fs->logical_inodes[indices[i]].type = SYMLINK_INODE_TYPE;
	fs->logical_inodes[indices[i]].generation_number = ((entry_obj *)objs[i])->gen;
	fs->logical_inodes[indices[i]].entry_inode.parent = ((entry_obj *)objs[i])->parent;
	set_attrs(indices[i], &((entry_obj *)objs[i])->attr);
	break;
      case DIR_INODE_TYPE:
#ifdef STATS
	num_put_dir++;
#endif
	update_dir(indices[i], uptodate, &num_uptodate, num_objs, sizes, indices, objs);
	fs->logical_inodes[indices[i]].type = DIR_INODE_TYPE;
	fs->logical_inodes[indices[i]].generation_number = ((entry_obj *)objs[i])->gen;
	fs->logical_inodes[indices[i]].entry_inode.parent = ((entry_obj *)objs[i])->parent;
	/* Update attrs. */
	fs->logical_inodes[indices[i]].entry_inode.atime = ((entry_obj *)objs[i])->attr.atime;
	fs->logical_inodes[indices[i]].entry_inode.mtime = ((entry_obj *)objs[i])->attr.mtime;
	fs->logical_inodes[indices[i]].entry_inode.ctime = ((entry_obj *)objs[i])->attr.ctime;
	break;
      }
    }
    else if (indices[i] == FS_ROOT_PAGE)
      fs->first_free_inode = *((int *)(objs[i]));
  }
  free(uptodate);
}

void shutdown_state(FILE *o)
{
  int i;
  size_t wb = 0;
  size_t ab = 0;

  // XXX -- TODO: unmount root

  for (i=0; i<NUM_LOGICAL_INODES; i++) {
    wb += fwrite(&fs->logical_inodes[i].type, sizeof(int), 1, o);
    ab++;
    wb += fwrite(&fs->logical_inodes[i].generation_number, sizeof(int), 1, o);
    ab++;
    if (fs->logical_inodes[i].type == FREE_INODE_TYPE) {
      wb += fwrite(&fs->logical_inodes[i].free_inode.next_free, sizeof(int), 1, o);
      ab++;
    }
    else {
      wb += fwrite(&fs->logical_inodes[i].entry_inode.atime, sizeof(nfstimeval), 1, o);
      ab++;
      wb += fwrite(&fs->logical_inodes[i].entry_inode.ctime, sizeof(nfstimeval), 1, o);
      ab++;
      wb += fwrite(&fs->logical_inodes[i].entry_inode.mtime, sizeof(nfstimeval), 1, o);
      ab++;
    }
  }

  wb += fwrite(&fs->first_free_inode, sizeof(int), 1, o);
  ab++;

  save_fileid_map(o);

  if (ab!=wb)
    fprintf(stderr, "Error in writing to disk during shutdown\n");
      
}

void restart_state(FILE *i)
{
  int n;
  size_t rb = 0;
  size_t ab = 0;

  fh_map_clear();
  fileid_map_clear();

  // XXX -- TODO: remount root instead of this:
  add(&fs->logical_inodes[0].entry_inode.NFS_handle,0);

  for (n=0; n<NUM_LOGICAL_INODES; n++) {
    if (n>0)
      *((long *)(&fs->logical_inodes[n].entry_inode.NFS_handle)) = 0;

    rb += fread(&fs->logical_inodes[n].type, sizeof(int), 1, i);
    ab++;
    rb += fread(&fs->logical_inodes[n].generation_number, sizeof(int), 1, i);
    ab++;
    if (fs->logical_inodes[n].type == FREE_INODE_TYPE) {
      rb += fread(&fs->logical_inodes[n].free_inode.next_free, sizeof(int), 1, i);
      ab++;
    }
    else {
      rb += fread(&fs->logical_inodes[n].entry_inode.atime, sizeof(nfstimeval), 1, i);
      ab++;
      rb += fread(&fs->logical_inodes[n].entry_inode.ctime, sizeof(nfstimeval), 1, i);
      ab++;
      rb += fread(&fs->logical_inodes[n].entry_inode.mtime, sizeof(nfstimeval), 1, i);
      ab++;
    }
  }

  rb += fread(&fs->first_free_inode, sizeof(int), 1, i);
  ab++;

  read_fileid_map(i);

  if (ab != rb)
    fprintf(stderr, "Error in reading from disk during restart\n");

#ifdef STATS
  if (num_recov < MAX_NUM_RECOV)
    num_recov++;
#endif
}

void modified_inode(int inum)
{
  int logical_objn = MIN_INODES_PAGE + inum;
  Byz_modify(1, &logical_objn);
}

void modified_fs_info()
{
  int logical_objn = FS_ROOT_PAGE;
  Byz_modify(1, &logical_objn);
}

int cmp_dir_entries(const void *a, const void *b)
{
  return (((struct direntry *)b)->fileid - ((struct direntry *)a)->fileid);
}
