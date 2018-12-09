#include <sys/mman.h>
#include <fcntl.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <assert.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <rpc/rpc.h>

#include "fs_interf.h"
#include "inode.h"
#include "fs.h"
#include "libbyz.h"


#ifdef NO_REPLICATION
#define Byz_modify2(a, b)
#define Byz_modify1(a)
#define Byz_modify(a, b)
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

/* 
 * Generic inode and block manipulation.
 */ 
struct inode *FS_alloc_inode() {
  struct inode *ret;
  if (fs->free_inodes) {
    ret = fs->free_inodes;

    Byz_modify1(&(fs->free_inodes));
    fs->free_inodes = (struct inode *)(ret->iu.next);
    fs->finodes--; /* Byz_modify1 for free_inodes curently covers finodes */

    return ret;
  }
   
  printf("File system full: no free inodes \n");
  return 0;
}


struct block *FS_alloc_block() {
  struct block *ret;
  if (fs->free_blocks) {
    ret = fs->free_blocks;

    Byz_modify1(&(fs->free_blocks));
    fs->free_blocks = (struct block *)(ret->bu.next);
    fs->fblocks--; /* Byz_modify1 for free_blocks curently covers fblocks */

    return ret;
  }
   
  printf("File system full: no free blocks \n");
  return 0;
}


void FS_free_inode(struct inode *i) {
  Byz_modify1(&(i->iu.next));
  i->iu.next = (void*) fs->free_inodes;
  Byz_modify1(&(fs->free_inodes));
  fs->free_inodes = i;
  fs->finodes++; /* Byz_modify1 for free_inodes curently covers finodes */
  /* Increment inode generation to detect accesses to old inode */
  Byz_modify1(&gen(i));
  incr_gen(i);
}


void FS_free_block(struct block *b) {
  Byz_modify1(&(b->bu.next));
  b->bu.next = (void*)fs->free_blocks;
  Byz_modify1(&(fs->free_blocks));
  fs->free_blocks = b;
  fs->fblocks++; /* Byz_modify1 for free_blocks curently covers fblocks */
}

#define fetch_inode(inum) (fs->first_inode+(inum))
#define fetch_block(bnum) (fs->first_block+(bnum))
#define inum_inode(i) ((i)-fs->first_inode)
#define bnum_block(b) ((b)-fs->first_block)


/*
 * Directories and directory manipulation 
 */

/* Number of directory entries inlined in inode */
#define NUM_INLINE_DIR_ENTRIES (INODE_INLINE_DATA/sizeof(struct dir_entry))
#define NUM_INDIRECT_DIR_ENTRIES (Page_size/sizeof(struct dir_entry))
#define MAX_DIR_ENTRIES (NUM_INLINE_DIR_ENTRIES+NUM_INDIRECT_DIR_ENTRIES*2)

/* use rdev and blocksize to store block nums of indirect blocks (if any) */
#define dir_block(d)  ((d)->attr.rdev)
#define dir_block1(d) ((d)->attr.blocksize)

/* Inline inode data format for directories */
struct dir_data {
  struct dir_entry entries[NUM_INLINE_DIR_ENTRIES];
};

int FS_lookup_internal(struct inode *dir, char *name) {
  struct dir_data *d;
  int num_entries;
  int i;

  assert(dir->attr.type == NFDIR);
  num_entries = dir->attr.size;

#ifndef READ_ONLY_OPT
  /* Update time of last accesss */
  Byz_modify2(&(dir->attr.atime), sizeof(dir->attr.atime));
  dir->attr.atime.seconds = cur_time.tv_sec;
  dir->attr.atime.useconds = cur_time.tv_usec;
#endif

  /* First search inline entries */
  d = (struct dir_data*)(dir->iu.data);  
  for (i = 0; i < NUM_INLINE_DIR_ENTRIES; i++) {
    if (i >= num_entries) return -1; 
    if (!strcmp(name, d->entries[i].name)) return i;
  }

  /* Search non-inline blocks */
  d = (struct dir_data*)fetch_block(dir_block(dir));
  num_entries -= NUM_INLINE_DIR_ENTRIES;
  for (i = 0; i < NUM_INDIRECT_DIR_ENTRIES; i++) {
    if (i >= num_entries) return -1;
    if (!strcmp(name, d->entries[i].name)) 
      return i+NUM_INLINE_DIR_ENTRIES;
  }

  d = (struct dir_data*)fetch_block(dir_block1(dir));
  num_entries -= NUM_INDIRECT_DIR_ENTRIES;
  for (i = 0; i < num_entries; i++) {
    if (!strcmp(name, d->entries[i].name)) 
      return i+NUM_INLINE_DIR_ENTRIES+NUM_INDIRECT_DIR_ENTRIES;
  }

  return -1;
}


struct dir_entry *Directory_entry(struct inode *d, int num) {
  struct dir_data *ddata;

  assert(num < d->attr.size);

  ddata = (struct dir_data*)(d->iu.data);
  if (num < NUM_INLINE_DIR_ENTRIES) {
    return ddata->entries+num;
  } else { 
    num -= NUM_INLINE_DIR_ENTRIES;
    if (num < NUM_INDIRECT_DIR_ENTRIES) {
      ddata = (struct dir_data*)fetch_block(dir_block(d));
      return ddata->entries+num;
    }

    num -= NUM_INDIRECT_DIR_ENTRIES;
    ddata = (struct dir_data*)fetch_block(dir_block1(d));
    return ddata->entries+num;
  }
}


struct inode *FS_lookup(struct inode *dir, char *name) {
  struct dir_entry *de;
  int ret = FS_lookup_internal(dir, name);
  if (ret == -1) return 0;

  de = Directory_entry(dir, ret);
  return fetch_inode(de->inum);
}


nfsstat FS_link(struct inode *dir, char *name, struct inode *child) {
  int num_entries;
  struct dir_data *d;

  assert(dir->attr.type == NFDIR);


  if (strlen(name) > MAX_NAME_LEN) {
    return NFSERR_NAMETOOLONG;
  }
  
  if (FS_lookup(dir, name) != 0) {
    return NFSERR_EXIST;
  }

#if 0
  /* Need to check acces control */
  if (~write_access ()) {
    return (NFSERR_PERM);
  }
#endif

  num_entries = dir->attr.size;
  if (num_entries == MAX_DIR_ENTRIES) {
    return NFSERR_FBIG;
  }

  /* Associate "child" with "name" in "dir" */
  if (num_entries < NUM_INLINE_DIR_ENTRIES) {
    d = (struct dir_data*)(dir->iu.data);
    Byz_modify2(d->entries+num_entries, strlen(name)+1+sizeof(int));
    d->entries[num_entries].inum = inum_inode(child);
    strcpy(d->entries[num_entries].name, name);
  } else { 
    num_entries -= NUM_INLINE_DIR_ENTRIES;
    if (num_entries < NUM_INDIRECT_DIR_ENTRIES) {
      if (num_entries == 0) {
	/* Need to allocate first block */
	struct block *b = FS_alloc_block();
	if (b == 0) return NFSERR_NOSPC;
	Byz_modify2(&(dir->attr), sizeof(dir->attr));
	dir_block(dir) = bnum_block(b);
	dir->attr.blocks = 1;
      }
      d = (struct dir_data*)fetch_block(dir_block(dir));
      Byz_modify2(d->entries+num_entries, strlen(name)+1+sizeof(int));
      d->entries[num_entries].inum = inum_inode(child);
      strcpy(d->entries[num_entries].name, name);
    } else {
      num_entries -= NUM_INDIRECT_DIR_ENTRIES;
      if (num_entries == 0) {
	/* Need to allocate second block */
	struct block *b = FS_alloc_block();
	if (b == 0) return NFSERR_NOSPC;
	Byz_modify2(&(dir->attr), sizeof(dir->attr));
	dir_block1(dir) = bnum_block(b);
	dir->attr.blocks = 2;
      }
      d = (struct dir_data*)fetch_block(dir_block1(dir));
      Byz_modify2(d->entries+num_entries, strlen(name)+1+sizeof(int));
      d->entries[num_entries].inum = inum_inode(child);
      strcpy(d->entries[num_entries].name, name);
    }
  }

  Byz_modify2(&(dir->attr), sizeof(dir->attr));
  dir->attr.size++;
  /* Update time of last modification and last status change for dir */
  dir->attr.mtime.seconds = cur_time.tv_sec;
  dir->attr.mtime.useconds = cur_time.tv_usec;
  dir->attr.ctime = dir->attr.mtime;
  
  Byz_modify2(&(child->attr), sizeof(child->attr));
  incr_refcnt(child);
  /* Update time of last status change for child */
  child->attr.ctime = dir->attr.ctime;

  return NFS_OK; 
}


nfsstat FS_unlink(struct inode *dir, char *name, int remove_dir) {
  struct inode *child;
  int child_num;
  int child_inum;
  struct dir_entry *d1, *d2;
  int refcnt;

  assert(dir->attr.type == NFDIR);
 
  if (strlen(name) > MAX_NAME_LEN) {
    return NFSERR_NAMETOOLONG;
  }
  
  child_num = FS_lookup_internal(dir, name);
  if (child_num == -1) {
    return NFSERR_NOENT;
  }

  child_inum = Directory_entry(dir, child_num)->inum;
  child = fetch_inode(child_inum);
  if (!remove_dir && child->attr.type == NFDIR) {
    return NFSERR_ISDIR;
  }

  if (remove_dir && child->attr.type != NFDIR) {
    return NFSERR_NOTDIR;
  }

  if (!strcmp(name, ".") || !strcmp(name, "..")) {
    return NFSERR_ACCES;
  }

#if 0
  /* TODO: Need to check acces control */
  if (~write_access ()) {
    return (NFSERR_PERM);
  }
#endif

  if (remove_dir && child->attr.size != 2 && child->attr.nlink == 2) {
    return NFSERR_NOTEMPTY;
  }

  d1 = Directory_entry(dir, child_num);
  d2 = Directory_entry(dir, dir->attr.size-1);
  if (d1 != d2) {
    Byz_modify2(d1, sizeof(struct dir_entry));
    bcopy((char*)d2, (char*)d1, sizeof(struct dir_entry));
  } 

  Byz_modify2(&(dir->attr), sizeof(dir->attr));
  dir->attr.size--;
  if (dir->attr.blocks > 0) {
    if (dir->attr.size <= NUM_INLINE_DIR_ENTRIES) {
      /* Can free first indirect block */
      FS_free_block(fetch_block(dir_block(dir)));
      dir_block(dir) = -1;
      dir->attr.blocks = 0;
    } else {
      if (dir->attr.blocks > 1 && 
	  dir->attr.size <= NUM_INLINE_DIR_ENTRIES+NUM_INDIRECT_DIR_ENTRIES) {
	/* Can free second indirect block */
	FS_free_block(fetch_block(dir_block1(dir)));
	dir_block1(dir) = -1;
	dir->attr.blocks = 1;
      }
    }
  }

  dir->attr.ctime.seconds = cur_time.tv_sec;
  dir->attr.ctime.useconds = cur_time.tv_usec;
  dir->attr.mtime = dir->attr.ctime;

  Byz_modify2(&(child->attr), sizeof(child->attr));
  refcnt = decr_refcnt(child);
  if (refcnt == 0) {
    /* Free child and its associated storage */
    if (child->attr.type == NFREG)
      FS_truncate(child, 0);
    FS_free_inode(child);
    return NFS_OK;
  } else {
    /* Update time of last status change for child */
    child->attr.ctime = dir->attr.ctime;
  }
  
  /* If it is a directory "." points to self so refcnt == 1 means it
     can be removed. We checked for emptyness before. */
  if (refcnt == 1 && child->attr.type == NFDIR) {
    /* Decrement reference count of parent */
    struct inode *parent = FS_lookup(child, "..");
    if (parent != dir) {
      Byz_modify1(&(parent->attr.nlink));
    }
    decr_refcnt(parent);
    FS_free_inode(child);
  }
  return NFS_OK; 
}


struct inode *FS_create_dir(struct inode *parent) {
  struct inode *dir = FS_alloc_inode();
  if (!dir) return 0;


  /* Make sure we got the right inum for the root, i.e., 0 */
  assert(parent != 0 || inum_inode(dir) == 0);

  /* Initialize directory */
  Byz_modify2(&(dir->attr), sizeof(dir->attr));
  dir->attr.type = NFDIR;
  dir->attr.mode = (parent) ? parent->attr.mode : 16895;
  dir->attr.nlink = 0;
  dir->attr.uid = cred_uid;
  dir->attr.gid = cred_gid;
  dir->attr.size = 0;        /* number of directory entries */
  dir->attr.blocksize = -1;
  dir->attr.rdev = -1;
  dir->attr.blocks = 0;
  dir->attr.fileid = inum_inode(dir);
  dir->attr.atime.seconds = cur_time.tv_sec;
  dir->attr.atime.useconds = cur_time.tv_usec;
  dir->attr.mtime = dir->attr.atime;
  dir->attr.ctime = dir->attr.atime;

  if (!parent) {
    /* root directory */
    parent = dir;
    dir->attr.nlink = 1;
  }

  /* Initially directory has no non-inline entries */
  dir_block(dir) = -1;
  dir_block1(dir) = -1;

  /* Create "." and ".." entries */
  FS_link(dir, ".", dir);
  FS_link(dir, "..", parent);

  return dir;
}


struct inode *FS_create_symlink(char *text) {
  struct inode *lnk = FS_alloc_inode();
  if (!lnk) return 0;

  /* Initialize symbolic link */
  Byz_modify2(lnk, sizeof(*lnk));
  lnk->attr.type = NFLNK;
  lnk->attr.mode = 0777; 
  lnk->attr.nlink = 0;
  lnk->attr.uid = cred_uid;
  lnk->attr.gid = cred_gid;
  lnk->attr.size = strlen(text);
  lnk->attr.blocksize = Page_size;
  lnk->attr.rdev = -1;
  lnk->attr.blocks = 0;
  lnk->attr.fileid = inum_inode(lnk);
  lnk->attr.atime.seconds = cur_time.tv_sec;
  lnk->attr.atime.useconds = cur_time.tv_usec;
  lnk->attr.mtime = lnk->attr.atime;
  lnk->attr.ctime = lnk->attr.atime;

  strcpy(lnk->iu.data, text);

  return lnk;
}


/*
 *   Files and file manipulation:
 */

/* Number of directory entries inlined in inode */
#define NUM_INLINE_BLOCKS (INODE_INLINE_DATA/sizeof(int)-1)
#define MAX_FILE_SIZE ((NUM_INLINE_BLOCKS+Page_size/sizeof(int))*Page_size)

/* Inline inode data format for regular files */
struct file_data {
  int in_blocks[NUM_INLINE_BLOCKS];
  int out_block;
};

struct inode *FS_create_file(struct inode *parent) {
  struct file_data *fdata;
  struct inode *file = FS_alloc_inode();
  if (!file) return 0;

  assert(parent != 0);

  /* Initialize file */
  Byz_modify2(&(file->attr), sizeof(file->attr));
  file->attr.type = NFREG;
  file->attr.mode = parent->attr.mode;
  file->attr.nlink = 0;
  file->attr.uid = cred_uid;
  file->attr.gid = cred_gid;
  file->attr.size = 0;
  file->attr.blocksize = Page_size;
  file->attr.rdev = 0;
  file->attr.blocks = 0;
  file->attr.fileid = inum_inode(file);
  file->attr.atime.seconds = cur_time.tv_sec;
  file->attr.atime.useconds = cur_time.tv_usec;
  file->attr.mtime = file->attr.atime;
  file->attr.ctime = file->attr.atime;

  /* Initially file has no non-inline entries */
  fdata = (struct file_data*)(file->iu.data);
  Byz_modify1(&(fdata->out_block));
  fdata->out_block = -1;
  return file;
}


struct block *File_fetch_block(struct inode *f, int bnum) {
  struct file_data *fdata;
  fdata = (struct file_data*)(f->iu.data);
  if (bnum < NUM_INLINE_BLOCKS) {
    return fetch_block(fdata->in_blocks[bnum]);
  } 

  assert(fdata->out_block >= 0);
  fdata = (struct file_data*)(fetch_block(fdata->out_block));
  return fetch_block(fdata->in_blocks[bnum-NUM_INLINE_BLOCKS]);
}

void File_store_block(struct inode *f, int bnum, struct block *b) {
  struct file_data *fdata;
  fdata = (struct file_data*)(f->iu.data);
  if (bnum < NUM_INLINE_BLOCKS) {
    Byz_modify1(&(fdata->in_blocks[bnum]));
    fdata->in_blocks[bnum] = bnum_block(b);
    return;
  } 
  
  assert(fdata->out_block >= 0);
  fdata = (struct file_data*)(fetch_block(fdata->out_block));
  Byz_modify1(&(fdata->in_blocks[bnum-NUM_INLINE_BLOCKS]));
  fdata->in_blocks[bnum-NUM_INLINE_BLOCKS] = bnum_block(b);
}


int FS_read(struct inode *file, int offset, int count, char **data) {
  int max;
  int bnum;
  int to_read;
  struct block *b;

  assert(file->attr.type == NFREG);

#ifndef READ_ONLY_OPT  
  /* Set access time */
  Byz_modify2(&(file->attr.atime), sizeof(file->attr.atime));
  file->attr.atime.seconds = cur_time.tv_sec;
  file->attr.atime.useconds = cur_time.tv_usec;
#endif
  
  /* Compute maximum number of bytes that can be read from this file
     starting at "offset" */
  max = file->attr.size-offset;
  if (count > max) count = max;
  if (count <= 0) return 0;
 
  bnum = offset/Page_size;
  if (bnum == (offset+count-1)/Page_size) {
    /* Common case: Reading a single block */
    struct block *b = File_fetch_block(file, bnum);
    *data = b->bu.data+(offset%Page_size);
    return count;
  }

  /* General case: reading accross blocks, since max transfer is 4KB and pages are 4KB 
     reads can span at most two blocks */
  if (count > Page_size)
    return 0;

  b = File_fetch_block(file, bnum);
  to_read = Page_size-(offset%Page_size);
  bcopy(b->bu.data+(offset%Page_size), nfs_iobuf, to_read);
  b = File_fetch_block(file, bnum+1);
  bcopy(b->bu.data, nfs_iobuf+to_read, count-to_read);
  *data = nfs_iobuf;
  return count;
}


nfsstat FS_append_zeros(struct inode *file, int new_size) {
  struct file_data *fdata;
  int bnum, last_bnum;
  if (new_size <= file->attr.size) {
    /* Nothing to be done */
    return NFS_OK;
  }

  fdata = (struct file_data*)(file->iu.data);
  last_bnum = new_size/Page_size;
  if (last_bnum >= NUM_INLINE_BLOCKS && fdata->out_block < 0) {
    /* Need to allocate an indirect block */
    struct block *b = FS_alloc_block();
    if (b == 0) {
      return NFSERR_NOSPC;
    }
    Byz_modify1(&(fdata->out_block));
    fdata->out_block = bnum_block(b); 
  }

  for (bnum = file->attr.blocks; bnum <= last_bnum; bnum++) {
    File_store_block(file, bnum, fs->zero_block);
  }

  /* Update attributes */
  Byz_modify2(&(file->attr), sizeof(file->attr));
  file->attr.size = new_size;
  file->attr.blocks = last_bnum+1;
  file->attr.ctime.seconds = cur_time.tv_sec;
  file->attr.ctime.useconds = cur_time.tv_usec;

  return NFS_OK;
}


nfsstat FS_truncate(struct inode *file, int new_size) {
  int bnum;
  struct block *b;
  struct file_data *fdata = (struct file_data*)(file->iu.data);

  assert(file->attr.type == NFREG);
  if (file->attr.size > new_size) {
    /* Try to free extra blocks */
    for (bnum = (new_size+Page_size-1)/Page_size; bnum < file->attr.blocks; bnum++) {
       b = File_fetch_block(file, bnum);
       if (b != fs->zero_block)
	 FS_free_block(b);
    }

    Byz_modify2(&(file->attr), sizeof(file->attr));
    file->attr.size = new_size;
    file->attr.blocks = (new_size+Page_size-1)/Page_size;
    if (file->attr.blocks <= NUM_INLINE_BLOCKS && fdata->out_block >= 0) {
      /* Free indirect block */
      FS_free_block(fetch_block(fdata->out_block));
      Byz_modify1(&(fdata->out_block));
      fdata->out_block = -1;
    }

    file->attr.ctime.seconds = cur_time.tv_sec;
    file->attr.ctime.useconds = cur_time.tv_usec;
    return NFS_OK;
  }

  return FS_append_zeros(file, new_size);
}

nfsstat FS_write(struct inode *file, int offset, int count, char *data) {
  nfsstat ret;
  int new_size;
  int bnum;
  int start;
  int to_write;
  int allocated = 0;
  struct block *b;
  assert(file->attr.type == NFREG);

  /* If new size is greater than the old size append zeros */
  new_size = offset+count; 
  ret = FS_append_zeros(file, new_size);
  if (ret != NFS_OK) {
    return ret;
  }
  
  /* Set times */
  Byz_modify2(&(file->attr), sizeof(file->attr));
  file->attr.atime.seconds = cur_time.tv_sec;
  file->attr.atime.useconds = cur_time.tv_usec;
  file->attr.mtime = file->attr.atime;


  /* Can write at most two blocks because transfer is limted to 4KB pages are 4KB. */
  if (count > Page_size)
    return NFSERR_NOSPC;

  bnum = offset/Page_size;
  start = offset%Page_size;
  to_write = (Page_size-start > count) ? count : Page_size-start;
  b = File_fetch_block(file, bnum);
  if (b == fs->zero_block) {
    /* Need to allocate a new block */
    b = FS_alloc_block();
    if (b == 0) {
      return NFSERR_NOSPC;
    }
    File_store_block(file, bnum, b);
    allocated = 1;
  }
  
  Byz_modify(b->bu.data+start, to_write);
  bcopy(data, b->bu.data+start, to_write);
  if (allocated) {
    /* If I allocated a new block zero the rest */
    if (start != 0) {
      Byz_modify(b->bu.data, start);
      bzero(b->bu.data, start);    
    }
    Byz_modify(b->bu.data+start+to_write, Page_size-(start+to_write));
    bzero(b->bu.data+start+to_write, Page_size-(start+to_write));
  }

  count = count - to_write;
  if (count == 0) {
    return NFS_OK;
  }

  b = File_fetch_block(file, bnum+1);
  if (b == fs->zero_block) {
    /* Need to allocate a new block */
    b = FS_alloc_block();
    if (b == 0) {
      return NFSERR_NOSPC;
    }
    File_store_block(file, bnum, b);
    allocated = 1;
  }
  Byz_modify(b->bu.data, count);
  bcopy(data+to_write, b->bu.data, count);
  if (allocated) {
    /* If I allocated a new block zero the rest */
    Byz_modify(b->bu.data+count, Page_size-count);
    bzero(b->bu.data+count, Page_size-count);
  }
  return NFS_OK;
}



/*
 *   Conversion between NFS file handles and inodes.
 */

struct nfs_fhandle {
  int inum;
  int generation;
};

struct inode *FS_fhandle_to_inode(fhandle *fh) {
  struct inode *ret;
  struct nfs_fhandle *nfh = (struct nfs_fhandle *)fh;
  
  /* Check bounds */
  if (nfh->inum < 0 || nfh->inum >= fs->num_inodes) {
    /* Stale file handle */
    return 0;
  }

  ret = fetch_inode(nfh->inum);
  if (nfh->generation != gen(ret)) {
    /* Stale file handle */
    return 0;
  } 

  return ret;
}

void FS_inode_to_fhandle(struct inode *arg, fhandle *res) {
  struct nfs_fhandle *nfh = (struct nfs_fhandle *)res;
  nfh->inum = inum_inode(arg);
  nfh->generation = gen(arg);
}

/* 
 * File system initialization:
 */
void FS_map(char *file, char **mem, int *sz) {
  /* map file system file */
  int fd;
  int ret;
  struct stat sbuf;

  fd = open(file, O_RDWR, 0);
  if (fd < 0) {
    perror("Could not open file system file");
    exit(-1);
  }

  /* Find size of file */
  ret = fstat(fd, &sbuf);
  if (ret < 0) {
    perror("Could not stat file system file");
    exit(-1);
  }

  // Important to map in same address for pointers to retain their meaning
  fs = (struct FS*)mmap((void*)0x4010c000, sbuf.st_size,  PROT_READ | PROT_WRITE, 
			  MAP_FILE | MAP_SHARED, fd, 0);
  if (fs == (struct FS*)-1) {
    perror("Could not mmap file system file");
    exit(-1);
  }

  printf("mapped at %p\n", fs);

  *mem = (char *)fs;
  *sz = sbuf.st_size;
}

void FS_init(int size, int usize) {
  /* Initialize file system if needed */
  struct inode *root;
  int i;
  int num_inode_pages;
  int max_inodes;

  fs = (struct FS*)((char*)fs+usize);
  size = size - usize;

  if (fs->num_pages == 0) {
    /* File system is empty; initialize file system data structures. */
    fs->num_pages = size/Page_size;

    num_inode_pages = fs->num_pages/32;
    max_inodes = ((num_inode_pages*Page_size)/sizeof(struct inode));

    fs->first_inode = (struct inode *)(((char*)fs)+Page_size);
    fs->num_inodes = max_inodes;
    fs->finodes = max_inodes;
    fs->free_inodes = 0;
    /* Link all but first inode into free_list */
    for (i = 1; i < max_inodes; i++) {
      fs->first_inode[i].iu.next = (void*)fs->free_inodes;
      fs->free_inodes = fs->first_inode+i;
    }
    /* Link first inode into free list for root directory */
    fs->first_inode->iu.next = (void*)fs->free_inodes;
    fs->free_inodes = fs->first_inode;

    /* Initialize special zero block that always contains all zeros */
    fs->zero_block=(struct block*)(((char*)fs)+(num_inode_pages+1)*Page_size);
    bzero((char*)fs->zero_block, Page_size);

    /* Initialize array and free list of blocks */
    fs->first_block = fs->zero_block+1;
    fs->num_blocks = fs->num_pages-num_inode_pages-2;
    fs->fblocks = fs->num_blocks;
    fs->free_blocks = 0;
    /* Link all blocks into free_list */
    for (i = 0; i < fs->num_blocks; i++) {
      fs->first_block[i].bu.next = (void*)fs->free_blocks;
      fs->free_blocks = fs->first_block+i;
    }

    /* Initialize root directory */
    root = FS_create_dir(0);
    if (root == 0) {
      printf("Could not create root\n");
      exit(-1);
    }
  } 
    
  printf("Blocks: used=%d free=%d \nInodes: used=%d free=%d  \n",
	 fs->num_blocks-fs->fblocks, fs->fblocks, fs->num_inodes-fs->finodes, fs->finodes);
}


void FS_free_blocks(int *tot_blocks, int *free_blocks) {
  *tot_blocks = fs->num_blocks;
  *free_blocks = fs->fblocks;
}

void sync_fs() {
  if (msync(fs, fs->num_pages*Page_size, MS_SYNC) != 0) {
    perror("Could not syncronyze file system");
  }
}
