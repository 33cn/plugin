#ifndef _inode_h
#define _inode_h 1

#include "nfsd.h"

#define FILE_INODE_TYPE 0
#define DIR_INODE_TYPE  1
#define FREE_INODE_TYPE 2
#define SYMLINK_INODE_TYPE 3

#define MAX_NAME_LEN 59

/* Define the following to order the directory entries */
#define ORDER_DIR_ENTRIES

/* Define the following to ommit uid / gid info */
#define OMMIT_UID_GID

/* Should be the same as Block_size in "libbyz.h" */

struct inode_id {
  int inode_num;
  int generation_num;
};

struct direntry {
  int fileid;
  char name[MAX_NAME_LEN+1]; /* RR-TODO - should I change this to the
			       bigger "NFS_MAXNAMELEN" ?  */
};

struct inode {
  int type;
  int generation_number;
  union {
    struct {
      int parent;
      fhandle NFS_handle;
      nfstimeval atime;
      nfstimeval mtime;
      nfstimeval ctime;
    } entry_inode_t;
    struct {
      int next_free;
    } free_inode_t;
  } iu;
};

#define entry_inode     iu.entry_inode_t
#define free_inode      iu.free_inode_t

#endif /* _inode_h */
