#ifndef _FS_H
#define _FS_H 1

#include "fs_interf.h"
#include "inode.h"

#define MOUNTPROG 100005
#define MOUNTVERS 1
#define PROCMNT 1

/* File system constants: */

#define BYZ_FSID 2055

#define ROOT_CREAT_TIME 973264515

#define NUM_LOGICAL_INODES 50000
//16384
//85000

struct block;

struct FS {

  /* FS inodes, which include a free list of inodes */
  struct inode logical_inodes[NUM_LOGICAL_INODES];
  int first_free_inode;

  /* Recovery info */
  u_int fileid[NUM_LOGICAL_INODES];

};

extern struct FS *fs;

#endif
