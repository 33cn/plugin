#ifndef _FS_interf_H
#define _FS_interf_H 1

#include "nfsd.h"

void FS_init(int sz, int usz);
/* Effects: Initializes the file system from "file". */

#define MAX_NAME_LEN 59
#define Page_size 4096

struct inode;
struct block {
  union {
    char data[Page_size];
    void *next;
  } bu;
};

struct dir_entry {
  int inum;
  char name[MAX_NAME_LEN+1]; 
};


struct inode *FS_alloc_inode();
void FS_free_inode(struct inode *i);
struct block *FS_alloc_block();
void FS_free_block(struct block *);
struct inode *FS_create_dir(struct inode *parent);
struct inode *FS_lookup(struct inode *dir, char *name);
nfsstat FS_link(struct inode *dir, char *name, struct inode *child);
struct inode *FS_fhandle_to_inode(fhandle *fh);
void FS_inode_to_fhandle(struct inode *arg, fhandle *res);
int FS_read(struct inode *file, int offset, int count, char **data);
nfsstat FS_write(struct inode *file, int offset, int count, char *data);
nfsstat FS_unlink(struct inode *dir, char *name, int remove_dir);
struct inode *FS_create_dir(struct inode *parent);
struct inode *FS_create_symlink(char *text);
struct inode *FS_create_file(struct inode *parent);
struct dir_entry *Directory_entry(struct inode *d, int num);
nfsstat FS_truncate(struct inode *file, int new_size);
void FS_free_blocks(int *tot_blocks, int *free_blocks);
void sync_fs();


/* Use read-only optimization for lookup and read operations
   (time last accessed is not set when these operations execute) */
#define READ_ONLY_OPT 1

#ifndef NO_REPLICATION
/* Print Byz library stats when statfs is called */
//#define PRINT_BSTATS  
#endif


#endif /*_FS_interf_H*/


