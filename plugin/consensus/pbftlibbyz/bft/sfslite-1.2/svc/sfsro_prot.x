/*
 * This file was written by Frans Kaashoek and Kevin Fu.  Its contents is
 * uncopyrighted and in the public domain.  Of course, standards of
 * academic honesty nonetheless prevent anyone in research from
 * falsely claiming credit for this work.
 */

%#include "bigint.h"
%#include "sfs_prot.h"

const SFSRO_FHSIZE = 20;
const SFSRO_BLKSIZE = 8192;
const SFSRO_NFH = 256;	       /* Blocks are approx 2KB each */
const SFSRO_NDIR = 7;
const SFSRO_FHDB_KEYS     = 255;  
const SFSRO_FHDB_CHILDREN = 256; /* must be  KEYS+1 */
const SFSRO_FHDB_NFH      = 256; /* FHDB blocks are approx 5KB each */
const SFSRO_MAX_PRIVATE_KEYSIZE = 512;


enum sfsrostat {
  SFSRO_OK = 0,
  SFSRO_ERRNOENT = 1
};

struct sfsro_dataresok {
  opaque data<>;
};

union sfsro_datares switch (sfsrostat status) {
 case SFSRO_OK:
   sfsro_dataresok resok;
 case SFSRO_ERRNOENT: 
   void;
};

enum ftypero {
  SFSROREG      = 1,
  SFSROREG_EXEC = 2,  /* Regular, executable file */
  SFSRODIR      = 3, 
  SFSRODIR_OPAQ = 4,
  SFSROLNK      = 5
};

const SFSRO_IVSIZE = 20;

/* public vs. access controlled content */
enum sfsro_accesstype {
  SFSRO_PUBLIC = 0,
  SFSRO_PRIVATE = 1
};

/* Regression protocol types */
enum sfsro_krtype {
  SFSRO_KRSHA1 = 0,
  SFSRO_KRRSA = 1,
  SFSRO_KRAES = 2
};

enum sfsro_lockboxtype {
  SFSRO_AES = 0,
  SFSRO_PROXY_REENC = 1
};

struct sfsro_sealed {
  unsigned gk_vers;
  sfsro_lockboxtype lt;
  opaque lockbox<SFSRO_MAX_PRIVATE_KEYSIZE>;
  opaque pkcs7<>; /* Encryption of PKCS-7 encoded plaintext */
};

struct sfsro_private {
  filename3 keymgr_sname;
  sfs_hash keymgr_hash;
  unsigned gk_id;
  sfsro_sealed ct;  
};

struct sfsro_public {
  sfs_msgtype type;     /* = SFS_ROFSINFO */
  sfs_time start;       /* seconds since UNIX epoch, GMT */
  unsigned duration;
  opaque iv[SFSRO_IVSIZE];
  sfs_hash fhdb;
  int blocksize;
  sfs_hash rootfh;
};

union sfsro2_signed_fsinfo switch (sfsro_accesstype type) {
 case SFSRO_PUBLIC:
   sfsro_public pub;
 case SFSRO_PRIVATE:
   sfsro_private priv;
};

struct sfsro2_fsinfo {
  sfsro2_signed_fsinfo info;
  sfs_sig2 sig;
};

union sfsro_fsinfo switch (sfsrostat stat) {
 case SFSRO_OK:
   sfsro2_fsinfo v2;
 case SFSRO_ERRNOENT:
   void;	
};

struct sfsro_inode_lnk {
  uint32 nlink;
  nfstime3 mtime;
  nfstime3 ctime;

  nfspath3 dest;
};

struct sfsro_inode_reg {
  uint32 nlink;
  uint64 size;
  uint64 used; 
  nfstime3 mtime;
  nfstime3 ctime;
 
  sfs_hash direct<SFSRO_NDIR>;
  sfs_hash indirect;
  sfs_hash double_indirect;
  sfs_hash triple_indirect;

};

union sfsro_inode switch (ftypero type) {
 case SFSROLNK:
   sfsro_inode_lnk lnk;
 default:
   sfsro_inode_reg reg;
};


struct sfsro_indirect {
  sfs_hash handles<SFSRO_NFH>;
};

struct sfsro_dirent {
  sfs_hash fh;
  string name<>;
  sfsro_dirent *nextentry;
};

struct sfsro_directory {
  sfsro_dirent *entries;
  bool eof;
};


struct sfsro_fhdb_indir {
  /*
     Invariant:
                key[i] < key [j] for all i<j

                keys in GETDATA(child[i]) are 
                   <= key[i+1] <
                keys in GETDATA(child[i+1])
  */
  sfs_hash key<SFSRO_FHDB_KEYS>;     
  sfs_hash child<SFSRO_FHDB_CHILDREN>;
};

/* Handles to direct blocks */
typedef sfs_hash sfsro_fhdb_dir<SFSRO_FHDB_NFH>;

enum dtype {
   SFSRO_INODE      = 0,
   SFSRO_FILEBLK    = 1, /* File data */
   SFSRO_DIRBLK     = 2, /* Directory data */
   SFSRO_INDIR      = 3, /* Indirect data pointer block */
   SFSRO_FHDB_DIR   = 4, /* Direct data pointer block for FH database */
   SFSRO_FHDB_INDIR = 5, /* Indirect data pointer block for FH database */
   SFSRO_SEALED     = 6  /* Sealed with encryption */
};

union sfsro_data switch (dtype type) {
 case SFSRO_INODE:
   sfsro_inode inode;
 case SFSRO_FILEBLK:
   opaque data<>;
 case SFSRO_DIRBLK:
   sfsro_directory dir;
 case SFSRO_INDIR:
   sfsro_indirect indir;
 case SFSRO_FHDB_DIR:
   sfsro_fhdb_dir fhdb_dir;
 case SFSRO_FHDB_INDIR:
   sfsro_fhdb_indir fhdb_indir;
 case SFSRO_SEALED:
   sfsro_sealed ct;
 default:
   void;
};

struct sfsro_getdataargs {
  filename3 sname;	/* Self-certifying name */
  sfs_hash fh;
};

struct sfsro_proxyreenc {
  opaque data<>;
};


struct chefs_stm {
  unsigned vers;
  sfs_hash keystate;
};

enum ktype {
   CHEFS_KRKEY      = 0,
   CHEFS_NOKRKEYS   = 1
};

struct chefs_key {
 unsigned gk_id;
};

struct chefs_keyresenc {
  uint32 msglen;
  bigint c;   /* Encrypted KR member state */
};

union chefs_keyres switch (sfsrostat status) {
 case SFSRO_OK:
  chefs_keyresenc enc;
 case SFSRO_ERRNOENT: 
   void;
};

typedef sfs_hash chefs_keyvec<>;

program SFSRO_PROGRAM {
	version SFSRO_VERSION_V2 {
		void 
		SFSROPROC2_NULL (void) = 0;

		sfsro_datares
		SFSROPROC2_GETDATA (sfsro_getdataargs) = 1;

		sfsro_proxyreenc
		SFSROPROC2_PROXYREENC (sfsro_proxyreenc) = 2;

		chefs_keyres
		SFSROPROC2_GETKEY (chefs_key) = 3;
	} = 2;
} = 344446;
