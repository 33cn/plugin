#ifndef _inode_h
#define _inode_h 1

#include "nfsd.h"
#include "fs_interf.h"

#define INODE_INLINE_DATA 256

struct inode {
  fattr attr;       // Attributes.
  union {
    char data[INODE_INLINE_DATA];   /* Data specific to each inode type.*/
    void *next;
  } iu;
};


/* Macros to increment and decrement inode reference count given
   inode* */
#define incr_refcnt(i) (++((i)->attr.nlink))
#define decr_refcnt(i) (--((i)->attr.nlink))


/* Macros to increment and access inode generation number given inode* */
#define gen(i) ((i)->attr.fsid)
#define incr_gen(i) ((i)->attr.fsid++)


#endif /* _inode_h */
