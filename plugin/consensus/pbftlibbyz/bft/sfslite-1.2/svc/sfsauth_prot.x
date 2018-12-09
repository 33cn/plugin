/* $Id: sfsauth_prot.x 1754 2006-05-19 20:59:19Z max $ */

/*
 * This file was written by David Mazieres.  Its contents is
 * uncopyrighted and in the public domain.  Of course, standards of
 * academic honesty nonetheless prevent anyone in research from
 * falsely claiming credit for this work.
 */

%#include "sfs_prot.h"

typedef string sfsauth_errmsg<>;
const SPRIVK_HISTORY_LEN = 2;


/*
 * sfs_authreq2 -- Contents of the login certificate in sfs_loginarg
 */
struct sfs_authreq2 {
  sfs_msgtype type;		/* = SFS_SIGNED_AUTHREQ(_NOCRED)? */
  sfs_hash authid;		/* SHA-1 (sfs_authinfo) */
  sfs_seqno seqno;		/* Counter, value unique per authid */
  sfs_idname user;		/* User name, can be "" for sigauth */
};
struct sfs_sigauth {
  sfs_authreq2 req;
  sfs_pubkey2 key;
  sfs_sig2 sig;
};
struct sfs_unixpwauth {
  sfs_authreq2 req;
  string password<>;
};
struct sfs_unixpwauth_res {	/* this struct is resmore for unixpw reply */
  string prompt<>;
  bool echo;
};
struct sfs_srpauth {
  sfs_authreq2 req;
  opaque msg<>;
};
union sfs_autharg2 switch (sfs_authtype type) {
  case SFS_NOAUTH:
   void;
 case SFS_AUTHREQ:
   sfs_authreq authreq1;
 case SFS_AUTHREQ2:
   sfs_sigauth sigauth;
 case SFS_UNIXPWAUTH:
   sfs_unixpwauth pwauth;
 case SFS_SRPAUTH:
   sfs_srpauth srpauth;
};

#if 0 /* from sfs_prot.x: */
struct sfs_loginarg {
  sfs_seqno seqno;
  opaque certificate<>;		/* marshalled sfs_autharg2 */
};
#endif
struct sfsauth2_loginarg {
  sfs_loginarg arg;
  sfs_hash authid;
  string source<>;		/* Source of request, for audit trail */
};

struct sfsauth2_loginokres {
  sfsauth_cred creds<>;
  opaque resmore<>;		/* If necessary, for mutual authentication */
  string hello<>;		/* To be printed on user's terminal */
};
union sfsauth2_loginres switch (sfs_loginstat status) {
 case SFSLOGIN_OK:
   sfsauth2_loginokres resok;
 case SFSLOGIN_MORE:
   opaque resmore<>;
 case SFSLOGIN_BAD:
   sfsauth_errmsg errmsg;
 default:
   void;
};

enum sfsauth_keyhalf_type {
  SFSAUTH_KEYHALF_NONE = 0,
  SFSAUTH_KEYHALF_PRIV = 1,
  SFSAUTH_KEYHALF_DELTA = 2,
  SFSAUTH_KEYHALF_FLAG = 3
};

union sfsauth_keyhalf switch (sfsauth_keyhalf_type type) {
  case SFSAUTH_KEYHALF_NONE:
    void;
  case SFSAUTH_KEYHALF_PRIV:
    sfs_2schnorr_priv_xdr priv<SPRIVK_HISTORY_LEN>;
  case SFSAUTH_KEYHALF_DELTA:
    bigint delta;
 case SFSAUTH_KEYHALF_FLAG:
   void;
};

/*
 * Auth server database types
 */
typedef sfs_groupmember sfs_groupmembers<>;

%const u_int32_t sfsauth_noid = 0xffffffff;
struct sfsauth_userinfo {
  sfs_idname name;
  unsigned id;
  unsigned vers;
  unsigned gid;
  sfs_idname *owner;
  sfs_pubkey2 pubkey;
  string privs<>;
  string pwauth<>;		/* Never returned, only set */
  sfs_privkey2 privkey;         /* Only returned after SRP authentication */
  sfsauth_keyhalf srvprivkey;	/* Never returned, only set */
  string audit<>;
};
struct sfsauth_groupinfo {
  sfs_idname name;
  unsigned id;
  unsigned vers;
  sfs_groupmembers owners;
  sfs_groupmembers members;
  string properties<>;
  string audit<>;
};
struct sfsauth_ids {
  sfs_idname user;
  unsigned uid;
  unsigned gid;
  unsigned gidlist<>;
};
struct sfsauth_srpparms {
  unsigned pwcost;
  string parms<>;
};
struct sfsauth_cacheentry {
  sfs_groupmember key;
  sfs_groupmembers values;
  unsigned vers;
  sfs_time refresh;
  sfs_time timeout;
  sfs_time last_update;
};
struct sfsauth_logentry {
  unsigned vers;
  sfs_groupmembers members;
  bool more;
  sfs_time refresh;
  sfs_time timeout;
  string audit<>;
};
struct sfsauth_revinfo {
  unsigned hyper dbrev;
  opaque dbid[16];
};
enum sfsauth_dbtype {
  SFSAUTH_ERROR = 0,
  SFSAUTH_USER = 1,
  SFSAUTH_GROUP = 2,
  SFSAUTH_IDS = 3,
  SFSAUTH_SRPPARMS = 5,
  SFSAUTH_CERTINFO = 6,
  SFSAUTH_CACHEENTRY = 7,
  SFSAUTH_EXPANDEDGROUP = 8,
  SFSAUTH_LOGENTRY = 9,
  SFSAUTH_REVINFO = 10,
  SFSAUTH_DELUSER = 11,
  SFSAUTH_DELGROUP = 12,
  SFSAUTH_NEXT = 13
};
union sfsauth_dbrec switch (sfsauth_dbtype type) {
 case SFSAUTH_ERROR:
   sfsauth_errmsg errmsg;
 case SFSAUTH_USER:
   sfsauth_userinfo userinfo;
 case SFSAUTH_GROUP:
 case SFSAUTH_EXPANDEDGROUP:
   sfsauth_groupinfo groupinfo;
 case SFSAUTH_IDS:
   sfsauth_ids ids;
 case SFSAUTH_SRPPARMS:
   sfsauth_srpparms srpparms;
 case SFSAUTH_CERTINFO:
   sfsauth_certinfores certinfo;
 case SFSAUTH_CACHEENTRY:
   sfsauth_cacheentry cacheentry;
 case SFSAUTH_LOGENTRY:
   sfsauth_logentry logentry;
 case SFSAUTH_REVINFO:
   sfsauth_revinfo revinfo;
 case SFSAUTH_DELUSER:
 case SFSAUTH_DELGROUP:
   sfs_idname deleted;
};

struct sfs_namevers {
  sfs_idname name;
  unsigned vers;
};
enum sfsauth_dbkeytype {
  SFSAUTH_DBKEY_NULL = 0,
  SFSAUTH_DBKEY_NAME = 1,
  SFSAUTH_DBKEY_ID = 2,
  SFSAUTH_DBKEY_PUBKEY = 3,
  SFSAUTH_DBKEY_NAMEVERS = 4,
  SFSAUTH_DBKEY_REVINFO = 5
};
union sfsauth_dbkey switch (sfsauth_dbkeytype type) {
 case SFSAUTH_DBKEY_NULL:
   void;
 case SFSAUTH_DBKEY_NAME:
   sfs_idname name;
 case SFSAUTH_DBKEY_ID:
   unsigned id;
 case SFSAUTH_DBKEY_PUBKEY:
   sfs_pubkey2 key;
 case SFSAUTH_DBKEY_NAMEVERS:
   sfs_namevers namevers;
 case SFSAUTH_DBKEY_REVINFO:
   sfsauth_revinfo revinfo;
};

/* arg must be accompanied by an authuint to retrieve certain fields. */
struct sfsauth2_query_arg {
  sfsauth_dbtype type;
  sfsauth_dbkey key;
};
typedef sfsauth_dbrec sfsauth2_query_res;
typedef unsigned hyper sfs_update_opts;

/* SFS Update Options Bit mask values */
const SFSUP_KPSRP = 0x1;	/* Don't overwrite SRP information */
const SFSUP_KPESK = 0x2;        /* Don't overwrite secret key information */
const SFSUP_KPPK = 0x4;		/* Don't overwrite public key */
const SFSUP_CLROKH = 0x8;	/* Clear old server keyhalf */
const SFSUP_CLRNKH = 0x10;	/* Clear new server keyhalf */

/* 
 * Signed message required for an update
 */
struct sfs_updatereq {
  sfs_msgtype type;		/* = SFS_UPDATEREQ */
  sfs_hash authid;		/* SHA-1 (sfs_authinfo) */
  sfsauth_dbrec rec;		/* USER or GROUP only */
  sfs_update_opts opts;         /* Bit Mask with Update Options */
};

/* Arg must be accompanied by authuint. */
struct sfsauth2_update_arg {
  sfs_updatereq req;
  sfs_sig2 *newsig;		/* sig by req.rec.userinfo->pubkey if
				 * non admin user updating key. */
  sfs_sig2 *authsig;		/* Signature for key corresponding to
				 * authuint.  Can be empty when
				 * registering. */
};
union sfsauth2_update_res switch (bool ok) {
 case false:
   sfsauth_errmsg errmsg;
 case true:
   void;
};

union sfsauth2_presig switch (sfs_keytype type) {
 case SFS_2SCHNORR:
   sfs_2schnorr_presig schnorr; 
 default:
   void;
};

union sfsauth2_sigreq switch (sfs_msgtype type) {
 case SFS_NULL:  
   sfs_hash rnd;
 case SFS_SIGNED_AUTHREQ:
   sfs_authreq2 authreq;
 case SFS_UPDATEREQ:
   sfs_updatereq updatereq;
};

struct sfsauth2_sign_arg {
  sfsauth2_sigreq req;
  sfs_authinfo authinfo;
  sfsauth2_presig presig;       /* used to pass a partial signature */
  sfs_idname user;              /* if no authno, look up user name */
  sfs_hash pubkeyhash;          /* hash of the pubkey client signed with */
};

union sfsauth2_sign_res switch (bool ok) {
 case false:
   sfsauth_errmsg errmsg;
 case true:
   sfs_sig2 sig;
};
  
program SFSAUTH_PROG {
	version SFSAUTH_V2 {
		void
		SFSAUTH2_NULL (void) = 0;

		sfsauth2_loginres
		SFSAUTH2_LOGIN (sfsauth2_loginarg) = 1;

		sfsauth2_query_res
		SFSAUTH2_QUERY (sfsauth2_query_arg) = 2;

		sfsauth2_update_res
		SFSAUTH2_UPDATE (sfsauth2_update_arg) = 3;

		sfsauth2_sign_res
		SFSAUTH2_SIGN (sfsauth2_sign_arg) = 4;
	} = 2;
} = 344442;
