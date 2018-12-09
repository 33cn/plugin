#ifndef _NFSD_H
#define _NFSD_H 1

/*  From rfc1094.h -- Define constants and data structures defined by rfc1094.
 *  This file is shared between mountd and nfsd.  It should not contain
 *  any prototypes, structures, or constants that are specific to this
 * implementation of nfsd.  Everything here is taken directly from the RFC.
 */

#define NFS_MAXDATA 8192
#define NFS_MAXPATHLEN 1024
#define NFS_MAXNAMLEN 255
#define COOKIESIZE 4
#define FHSIZE 32

enum nfsstat {
	NFS_OK = 0,
	NFSERR_PERM = 1,
	NFSERR_NOENT = 2,
	NFSERR_IO = 5,
	NFSERR_NXIO = 6,
	NFSERR_ACCES = 13,
	NFSERR_EXIST = 17,
	NFSERR_NODEV = 19,
	NFSERR_NOTDIR = 20,
	NFSERR_ISDIR = 21,
	NFSERR_FBIG = 27,
	NFSERR_NOSPC = 28,
	NFSERR_ROFS = 30,
	NFSERR_NAMETOOLONG = 63,
	NFSERR_NOTEMPTY = 66,
	NFSERR_DQUOT = 69,
	NFSERR_STALE = 70,
	NFSERR_WFLUSH = 99,
};
typedef enum nfsstat nfsstat;
bool_t xdr_nfsstat(XDR *xdrs, nfsstat *objp);

enum ftype {
	NFNON = 0,
	NFREG = 1,
	NFDIR = 2,
	NFBLK = 3,
	NFCHR = 4,
	NFLNK = 5,
	NFSOCK = 6,
	NFBAD = 7,
	NFFIFO = 8
};
typedef enum ftype ftype;
bool_t xdr_ftype(XDR *xdrs, ftype *objp);


struct fhandle {
	char data[FHSIZE];
};
typedef struct fhandle fhandle;
bool_t xdr_fhandle(XDR *xdrs, fhandle *objp);


struct nfstimeval {
	u_int seconds;
	u_int useconds;
};
typedef struct nfstimeval nfstimeval;
bool_t xdr_nfstimeval(XDR *xdrs, nfstimeval *objp);


struct fattr {
	ftype type;
	u_int mode;
	u_int nlink;
	u_int uid;
	u_int gid;
	u_int size;
	u_int blocksize;
	u_int rdev;
	u_int blocks;
	u_int fsid;
	u_int fileid;
	nfstimeval atime;
	nfstimeval mtime;
	nfstimeval ctime;
};
typedef struct fattr fattr;
bool_t xdr_fattr(XDR *xdrs, fattr *objp);


struct sattr {
	u_int mode;
	u_int uid;
	u_int gid;
	u_int size;
	nfstimeval atime;
	nfstimeval mtime;
};
typedef struct sattr sattr;
bool_t xdr_sattr(XDR *xdrs, sattr *objp);


typedef char *filename;
bool_t xdr_filename(XDR *xdrs, filename *objp);


typedef char *path;
bool_t xdr_path(XDR *xdrs, path *objp);


struct attrstat {
	nfsstat status;
	fattr attributes;
};
typedef struct attrstat attrstat;
bool_t xdr_attrstat(XDR *xdrs, attrstat *objp);


struct diropargs {
	fhandle dir;
	filename name;
};
typedef struct diropargs diropargs;
bool_t xdr_diropargs(XDR *xdrs, diropargs *objp);


struct diropokres {
	fhandle file;
	fattr attributes;
};
typedef struct diropokres diropokres;
bool_t xdr_diropokres(XDR *xdrs, diropokres *objp);


struct diropres {
	nfsstat status;
	diropokres diropok;
};
typedef struct diropres diropres;
bool_t xdr_diropres(XDR *xdrs, diropres *objp);


struct sattrargs {
	fhandle file;
	sattr attributes;
};
typedef struct sattrargs sattrargs;
bool_t xdr_sattrargs(XDR *xdrs, sattrargs *objp);


struct readlinkres {
	nfsstat status;
	path data;
};
typedef struct readlinkres readlinkres;
bool_t xdr_readlinkres(XDR *xdrs, readlinkres *objp);


struct readargs {
	fhandle file;
	u_int offset;
	u_int count;
	u_int totalcount;
};
typedef struct readargs readargs;
bool_t xdr_readargs(XDR *xdrs, readargs *objp);


struct readokres {
	fattr attributes;
	struct {
		u_int data_len;
		char *data_val;
	} data;
};
typedef struct readokres readokres;
bool_t xdr_readokres(XDR *xdrs, readokres *objp);


struct readres {
	nfsstat status;
	readokres reply;
};
typedef struct readres readres;
bool_t xdr_readres(XDR *xdrs, readres *objp);


struct writeargs {
	fhandle file;
	u_int beginoffset;
	u_int offset;
	u_int totalcount;
	struct {
		u_int data_len;
		char *data_val;
	} data;
};
typedef struct writeargs writeargs;
bool_t xdr_writeargs(XDR *xdrs, writeargs *objp);


struct createargs {
	diropargs where;
	sattr attributes;
};
typedef struct createargs createargs;
bool_t xdr_createargs(XDR *xdrs, createargs *objp);


struct renameargs {
	diropargs from;
	diropargs to;
};
typedef struct renameargs renameargs;
bool_t xdr_renameargs(XDR *xdrs, renameargs *objp);


struct linkargs {
	fhandle from;
	diropargs to;
};
typedef struct linkargs linkargs;
bool_t xdr_linkargs(XDR *xdrs, linkargs *objp);


struct symlinkargs {
	diropargs from;
	path to;
	sattr attributes;
};
typedef struct symlinkargs symlinkargs;
bool_t xdr_symlinkargs(XDR *xdrs, symlinkargs *objp);


struct nfscookie {
	char data[COOKIESIZE];
};
typedef struct nfscookie nfscookie;
bool_t xdr_nfscookie(XDR *xdrs, nfscookie *objp);


struct readdirargs {
	fhandle dir;
	nfscookie cookie;
	u_int count;
};
typedef struct readdirargs readdirargs;
bool_t xdr_readdirargs(XDR *xdrs, readdirargs *objp);


struct entry {
	u_int fileid;
	filename name;
	nfscookie cookie;
	struct entry *nextentry;
};
typedef struct entry entry;
bool_t xdr_entry(XDR *xdrs, entry *objp);


struct dirlist {
	entry *entries;
	bool_t eof;
};
typedef struct dirlist dirlist;
bool_t xdr_dirlist(XDR *xdrs, dirlist *objp);


struct readdirres {
	nfsstat status;
	dirlist readdirok;
};
typedef struct readdirres readdirres;
bool_t xdr_readdirres(XDR *xdrs, readdirres *objp);


struct statfsokres {
	u_int tsize;
	u_int bsize;
	u_int blocks;
	u_int bfree;
	u_int bavail;
};
typedef struct statfsokres statfsokres;
bool_t xdr_statfsokres(XDR *xdrs, statfsokres *objp);


struct statfsres {
	nfsstat status;
	statfsokres info;
};
typedef struct statfsres statfsres;
bool_t xdr_statfsres(XDR *xdrs, statfsres *objp);



#define NFSPROC_NULL		0
#define NFSPROC_GETATTR		1
#define NFSPROC_SETATTR		2
#define NFSPROC_ROOT		3
#define NFSPROC_LOOKUP		4
#define NFSPROC_READLINK	5
#define NFSPROC_READ		6
#define NFSPROC_WRITECACHE	7
#define NFSPROC_WRITE		8
#define NFSPROC_CREATE		9
#define NFSPROC_REMOVE		10
#define NFSPROC_RENAME		11
#define NFSPROC_LINK		12
#define NFSPROC_SYMLINK		13
#define NFSPROC_MKDIR		14
#define NFSPROC_RMDIR		15
#define NFSPROC_READDIR		16
#define NFSPROC_STATFS		17

#define MNTPATHLEN 1024
#define MNTNAMLEN 255

struct fhstatus {
	u_int status;
	fhandle directory;
};
typedef struct fhstatus fhstatus;
bool_t xdr_fhstatus(XDR *xdrs, fhstatus *objp);


typedef char *dirpath;
bool_t xdr_dirpath(XDR *xdrs, dirpath *objp);


typedef char *name;
bool_t xdr_name(XDR *xdrs, name *objp);


typedef struct mountnode *mountlist;
bool_t xdr_mountlist(XDR *xdrs, mountlist *objp);


struct mountnode {
	name hostname;
	dirpath directory;
	mountlist nextentry;
};
typedef struct mountnode mountnode;
bool_t xdr_mountnode(XDR *xdrs, mountnode *objp);


typedef struct groupnode *groups;
bool_t xdr_groups(XDR *xdrs, groups *objp);


struct groupnode {
	name grname;
	groups grnext;
};
typedef struct groupnode groupnode;
bool_t xdr_groupnode(XDR *xdrs, groupnode *objp);


typedef struct exportnode *exportlist;
bool_t xdr_exportlist(XDR *xdrs, exportlist *objp);


struct exportnode {
	dirpath filesys;
	groups _groups;
	exportlist next;
};
typedef struct exportnode exportnode;
bool_t xdr_exportnode(XDR *xdrs, exportnode *objp);


#define MOUNTPROG		100005
#define MOUNTVERS		1
#define MOUNTPROC_NULL		0
#define MOUNTPROC_MNT		1
#define MOUNTPROC_DUMP		2
#define MOUNTPROC_UMNT		3
#define MOUNTPROC_UMNTALL	4
#define MOUNTPROC_EXPORT	5
#define MOUNTPROC_EXPORTALL	6


/*
 *  A block of memory large enough to hold any nfsproc argument.
 */
typedef union {
    fhandle     _fhandle;
    sattrargs   _sattrargs;
    diropargs   _diropargs;
    readargs    _readargs;
    writeargs   _writeargs;
    createargs  _createargs;
    renameargs  _renameargs;
    linkargs    _linkargs;
    symlinkargs _symlinkargs;
    readdirargs _readdirargs;
} nfsproc_argument;

/*
 *  A block of memory large enough to hold any nfsproc result.
 */
typedef union {
    attrstat    _attrstat;
    diropres    _diropres;
    readlinkres _readlinkres;
    readres     _readres;
    nfsstat	_nfsstat;
    readdirres  _readdirres;
    statfsres   _statfsres;
} nfsproc_result;

extern enum nfsstat nfs_errno();

#endif 
