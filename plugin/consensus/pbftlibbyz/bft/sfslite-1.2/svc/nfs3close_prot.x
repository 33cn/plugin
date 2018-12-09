/* $Id: nfs3close_prot.x 1754 2006-05-19 20:59:19Z max $ */

%#include "nfs3_prot.h"

program cl_NFS_PROGRAM {
	version cl_NFS_V3 {
		void
		cl_NFSPROC3_NULL (void) = 0;
		
		getattr3res
		cl_NFSPROC3_GETATTR (nfs_fh3) = 1;
		
		wccstat3
		cl_NFSPROC3_SETATTR (setattr3args) = 2;
		
		lookup3res
		cl_NFSPROC3_LOOKUP (diropargs3) = 3;
		
		access3res
		cl_NFSPROC3_ACCESS (access3args) = 4;
		
		readlink3res
		cl_NFSPROC3_READLINK (nfs_fh3) = 5;
		
		read3res
		cl_NFSPROC3_READ (read3args) = 6;
		
		write3res
		cl_NFSPROC3_WRITE (write3args) = 7;
		
		diropres3
		cl_NFSPROC3_CREATE (create3args) = 8;
		
		diropres3
		cl_NFSPROC3_MKDIR (mkdir3args) = 9;
		
		diropres3
		cl_NFSPROC3_SYMLINK (symlink3args) = 10;
		
		diropres3
		cl_NFSPROC3_MKNOD (mknod3args) = 11;
		
		wccstat3
		cl_NFSPROC3_REMOVE (diropargs3) = 12;
		
		wccstat3
		cl_NFSPROC3_RMDIR (diropargs3) = 13;
		
		rename3res
		cl_NFSPROC3_RENAME (rename3args) = 14;
		
		link3res
		cl_NFSPROC3_LINK (link3args) = 15;
		
		readdir3res
		cl_NFSPROC3_READDIR (readdir3args) = 16;
		
		readdirplus3res
		cl_NFSPROC3_READDIRPLUS (readdirplus3args) = 17;
		
		fsstat3res
		cl_NFSPROC3_FSSTAT (nfs_fh3) = 18;
		
		fsinfo3res
		cl_NFSPROC3_FSINFO (nfs_fh3) = 19;
		
		pathconf3res
		cl_NFSPROC3_PATHCONF (nfs_fh3) = 20;
		
		commit3res
		cl_NFSPROC3_COMMIT (commit3args) = 21;

		commit3res
		cl_NFSPROC3_CLOSE (nfs_fh3) = 25;
	} = 3;
} = 100003;
