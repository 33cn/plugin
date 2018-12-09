#ifndef _FS_H
#define _FS_H 1

#include "fs_interf.h"

/* 
 * File system is simply a memory mapped file created with a certain size
 * that has inodes and file blocks.
 */

struct inode;
struct block;

struct FS {
  int num_pages;      /* Number of Page_size pages in file system. */
  struct inode *first_inode; /* Array of inodes. */
  int num_inodes;     /* Number of inodes in array. */
  int finodes;        /* Number of inodes in free list */
  struct inode *free_inodes; /* Free list for inodes. */
  struct block *first_block; /* Array of blocks */
  int num_blocks;     /* Number of blocks in array. */
  int fblocks;        /* Number of blocks in free list */
  struct block *free_blocks; /* Free list of blocks. */
  struct block *zero_block; /* Pointer to a special zero block that 
                               is always full of zeros and is never modified */
};

extern struct FS *fs;

#endif
