#ifndef _FS_interf_H
#define _FS_interf_H 1

#include "nfsd.h"

int FS_init(char *, char *);
/* Effects: Initializes the file system (conformance rep's structures).  */

struct inode;
struct inode_id;

int FS_fhandle_to_NFS_handle(fhandle *fh);
/* Requires - fh points to a pre-alloc'd file handle
   Effects  - if there is a correct mapping between a client fhandle given
              by the original value of *fh and an NFS handle used by
	      the undelying NFS server, set the contents of the handle
	      pointed by fh to that NFS handle and returns the logical
	      inode number for that file.
	      Otherwise, returns -1                                      */

int FS_NFS_handle_to_client_handle(fhandle *fh);
/* Requires - fh points to a pre-alloc'd file handle
   Effects  - if there is a correct mapping between an NFS fhandle given
              by the original value of *fh and a client fhandle,
	      set the contents of the handle
	      pointed by fh to that client handle and returns the inode
	      logical number for that file.
	      Otherwise, returns -1                                      */

void FS_client_fhandle_to_inode_id(fhandle *fh, struct inode_id *id);
/* Requires - fh, id are pointer to pre-alloc'd structures. fh points to
              a valid Client fhandle (returned by the replica to be
	      compliant with the client fhandle format)
   Effects  - the value pointed by id is set to the
              {logical inode num, generation num} corresponding to the
	      client fhandle pointed by fh.                              */

void NFS_inode_id_to_client_fhandle(int logical_inode_num, int generation_num,
				    fhandle *fh);
/* Requires - fh is a pointer to a pre-alloced structure.
   Effects  - the value pointed by fh is set to the
              client file handle obtained from the other arguments       */



int FS_attr_NFS_to_client(int logicalinum, fattr *attr);
/* Requires - "attr" points to a pre-alloc'd fattr structure
   Effects  - if logicalinum is the inode number of a file or dir in the
              system sets the attr structure to its canonical (seen by
              the client) vaules and returns 0. Otherwise, returns -1    */

int FS_create_entry(fhandle *new_fh, fattr *attrs, struct timeval *curr_time,
		    int entry_type, int parent_inum);
/* Requires - new_fh, attrs and curr_time are pre-allocated
   Effects  - informs the conformance wrapper that an entry of type
              'entry_type' whose parent dir has inode num
	      'parent_inum' has been created
	      and NFS returned the fhandle 'new_fh' and attributes
	      'attrs'. The conf wrapper will modify its
	      rep to reflect this.                                       */

int FS_set_attr(int file_inum, sattr *set_attr);
/* Requires - set_attr is pre-allocated
   Effects  - if 'file_inum' is the inode number of a non-null entry,
              sets the times in the conformance rep to the corresponding
              values in the set_attr structure (or does not change them
	      if these contain the value '-1') and returns 0.
	      Otherwise returns -1.                                      */

int FS_remove_entry(int file_inum);
/* Effects  - informs the file system that the file 
              with inum 'file_inum' has been removed.
	      The file system will modify its internal
	      state to reflect this (set the corresponding
	      entry to a null entry).                                    */

int FS_update_file_info(int file_inum, fhandle *new_handle,
			int new_parent_inum, int fileid, int fsid);
/* Requires - new_handle is a pointer to a pre-alloc'd structure.
   Effects  - if file_inum is valid logical inode number for a non-null
              entry in the file system, updates the conformance rep to
	      reflect the new entry's file handle, parent inode number,
	      fileid and fsid; and returns 0. Otherwise returns -1.      */

int FS_update_time_modified(int inum, struct timeval *curr_time);
/* Requires - curr_time is pre-allocated
   Effects  - if inum is the inode number of a non-null entry, sets
              the 'atime' and 'mtime' in the conformance rep to the value
	      in 'curr_time'.                                            */
              

void modified_inode(int inum);
/* Effects - Informs the BASE library that the object with inode number
             'inum' is going to be changed                               */
             

/* Use read-only optimization for lookup and read operations
   (time last accessed is not set when these operations execute) */
/* #define NO_READ_ONLY_OPT 1 */

#endif /*_FS_interf_H*/


