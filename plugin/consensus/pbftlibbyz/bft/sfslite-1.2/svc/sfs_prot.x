/* $id: sfs_prot.x,v 1.73.4.2 2002/08/01 17:13:24 max Exp $ */

/*
 * This file was written by David Mazieres.  Its contents is
 * uncopyrighted and in the public domain.  Of course, standards of
 * academic honesty nonetheless prevent anyone in research from
 * falsely claiming credit for this work.
 */

%#include "bigint.h"

#ifdef SFSSVC
%#include "nfs3exp_prot.h"
#else /* !SFSSVC */
%#include "nfs3_prot.h"
const ex_NFS_PROGRAM = 344444;
const ex_NFS_V3 = 3;
#endif

#ifdef RPCC
# ifndef UNION_ONLY_DEFAULT
#  define UNION_ONLY_DEFAULT 1
# endif /* UNION_ONLY_DEFAULT */
#endif

#ifndef FSINFO
#define FSINFO sfs_fsinfo
#endif /* !FSINFO */

const SFS_PORT = 4;
const SFS_RELEASE = 8;		/* 100 * release no. where protocol changed */

enum sfsstat {
  SFS_OK = 0,
  SFS_BADLOGIN = 1,
  SFS_NOSUCHHOST = 2,
  SFS_NOTSUPP = 10004,
  SFS_TEMPERR = 10008,
  SFS_REDIRECT = 10020
};

/* Types of hashed or signed messages */
enum sfs_msgtype {
  SFS_NULL = 0,
  SFS_HOSTINFO = 1,
  SFS_KSC = 2,
  SFS_KCS = 3,
  SFS_SESSINFO = 4,
  SFS_AUTHINFO = 5,
  SFS_SIGNED_AUTHREQ = 6,
  SFS_AUTHREGISTER = 7,
  SFS_AUTHUPDATE = 8,
  SFS_PATHREVOKE = 9,
  SFS_KEYCERT = 10,
  SFS_ROFSINFO = 11,
  SFS_UPDATEREQ = 12,
  SFS_PUBKEY2_HASH = 13,
  SFS_SIGNED_AUTHREQ_NOCRED = 14,
  SFS_SESSINFO_SECRETID = 15
};

/* Type of service requested by clients */
enum sfs_service {
  SFS_SFS = 1,			/* File system service */
  SFS_AUTHSERV = 2,		/* Crypto key server */
  SFS_REX = 3			/* Remote execution */
};

typedef string sfs_extension<>;
typedef string sfs_hostname<222>;
typedef string sfs_idname<32>;
typedef string sfs_groupmember<256>;
typedef opaque sfs_hash[20];
typedef opaque sfs_secret[16];
typedef unsigned hyper sfs_seqno;
typedef unsigned hyper sfs_time;

typedef bigint sfs_pubkey;
typedef bigint sfs_ctext;
typedef bigint sfs_sig;

enum sfs_keytype {
  SFS_NOKEY = 0,
  SFS_RABIN = 1,
  SFS_2SCHNORR = 2,          /* proactive 2-Schnorr -- private */
  SFS_SCHNORR = 3,           /* either *Schnorr -- public */
  SFS_1SCHNORR = 4,          /* standard 1-Schnorr -- private */
  SFS_ESIGN = 5
};

struct sfs_1schnorr_priv_xdr {
  bigint p;
  bigint q;
  bigint g;
  bigint y;
  bigint x;
};

struct sfs_2schnorr_priv_xdr {
  bigint p;
  bigint q;
  bigint g;
  bigint y;
  bigint x;
  sfs_hostname hostname;     /* signing authority */
  sfs_idname uname;          /* username on signing server */
  string audit<>;            /* when generated or last updated */
};
struct sfs_2schnorr_priv_export {
  sfs_keytype kt;            /* == SFS_2SCHNORR */
  sfs_hash cksum;
  sfs_2schnorr_priv_xdr privkeys<2>;
};

struct sfs_1schnorr_priv_export {
  sfs_keytype kt;            /* == SFS_1SCHNORR */
  sfs_hash cksum;
  sfs_1schnorr_priv_xdr privkey;
};

struct sfs_esign_priv_xdr {
  bigint p;
  bigint q;
  unsigned k;
};

struct sfs_esign_priv_export {
  sfs_keytype kt;            /* == SFS_ESIGN */
  sfs_hash cksum;
  sfs_esign_priv_xdr privkey;
};

struct sfs_schnorr_pub_xdr {
  bigint p;
  bigint q;
  bigint g;
  bigint y;
};
struct sfs_schnorr_sig {
  bigint r;
  bigint s;
};
struct sfs_2schnorr_presig {
  bigint r;
};

struct sfs_rabin_priv_export {
  bigint p;
  bigint q;
  sfs_hash cksum;
};

struct sfs_esign_pub_xdr {
  bigint n;
  unsigned k;
};

struct sfs_rabin_priv_xdr {
  bigint p;
  bigint q;
};

typedef opaque sfs_privkey2<>;

union sfs_pubkey2 switch (sfs_keytype type) {
 case SFS_RABIN:
   bigint rabin;
 case SFS_SCHNORR:
   sfs_schnorr_pub_xdr schnorr;
 case SFS_ESIGN:
   sfs_esign_pub_xdr esign;
 default:
   void;
};

union sfs_ctext2 switch (sfs_keytype type) {
 case SFS_RABIN:
   bigint rabin;
};

union sfs_sig2 switch (sfs_keytype type) {
 case SFS_RABIN:
   bigint rabin;
 case SFS_SCHNORR:
   sfs_schnorr_sig schnorr;
 case SFS_ESIGN:
   bigint esign;
 default:
   void;
};

struct sfs_hashcharge {
  unsigned int bitcost;
  sfs_hash target;
};
typedef opaque sfs_hashpay[64];


/*
 * Hashed structures
 */

/* Two, identical copies of of the sfs_hostinfo structure are
 * concatenated and then hashed with SHA-1 to form the hostid. */
struct sfs_hostinfo {
  sfs_msgtype type;		/* = SFS_HOSTINFO */
  sfs_hostname hostname;
  sfs_pubkey pubkey;
};

struct sfs_pubkey2_hash {
  sfs_msgtype type;             /* = SFS_PUBKEY2_HASH */
  sfs_pubkey2 pubkey;
};

/* The HostID is computed as SHA-1 (SHA-1 (sfs_hostinfo2), sfs_hostinfo2),
 * where sfs_hostinfo2 is the XDR-marshaled value of this structure */
struct sfs_hostinfo2 {
  sfs_msgtype type;             /* = SFS_HOSTINFO XXX -- do we need this? */
  sfs_hostname hostname;
  unsigned port;
  sfs_pubkey2 pubkey;
};

struct sfs_connectinfo_4 {
  sfs_service service;
  sfs_hostname name;		/* Server hostname */
  sfs_hash hostid;		/* = SHA1 (sfs_hostinfo, sfs_hostinfo) */
  sfs_extension extensions<>;
};

struct sfs_connectinfo_5 {
  unsigned release;		/* 100 times release when protocol changed */
  sfs_service service;
  filename3 sname;		/* Self-certifying name */
  sfs_extension extensions<>;
};

union sfs_connectinfo switch (unsigned civers) {
 case 4:
   sfs_connectinfo_4 ci4;
 case 5:
   sfs_connectinfo_5 ci5;
};

struct sfs_servinfo_5 {
  sfs_hostinfo host;		/* Server hostinfo */
  unsigned prog;
  unsigned vers;
};

struct sfs_servinfo_7 {
  unsigned release;
  sfs_hostinfo2 host;
  unsigned prog;
  unsigned vers;
};

union sfs_servinfo switch (unsigned sivers) {
 case 5:
 case 6:
   sfs_servinfo_5 cr5;
 case 7:
   sfs_servinfo_7 cr7;
};

/* The two shared session keys, ksc and kcs, are the SHA-1 hashes of
 * sfs_sesskeydat with type = SFS_KCS or SFS_KSC.  */
struct sfs_sesskeydat {
  sfs_msgtype type;		/* = SFS_KSC or SFS_KCS */
  sfs_servinfo si;
  sfs_secret sshare;		/* Server's share of session key */
  sfs_connectinfo ci;
  sfs_pubkey kc;
  sfs_secret cshare;		/* Client's share of session key */
};

/* The sessinfo structure is hashed to produce a session ID--a
 * structure both the client and server know to be fresh, but which,
 * unlike the session keys, can safely be divulged to 3rd parties
 * during user authentication.  */
struct sfs_sessinfo {
  sfs_msgtype type;		/* = SFS_SESSINFO */
  opaque ksc<>;			/* = SHA-1 ({SFS_KSC, ...}) */
  opaque kcs<>;			/* = SHA-1 ({SFS_KCS, ...}) */
};

/* The authinfo structure is hashed to produce an authentication ID.
 * The authentication ID can be computed by an untrusted party (such
 * as a user's unprivileged authentication agent), but allows that
 * third party to verify or log the hostname and hostid to which
 * authentication is taking place.  */
struct sfs_authinfo {
  sfs_msgtype type;		/* = SFS_AUTHINFO */
  sfs_service service;
  sfs_hostname name;
  sfs_hash hostid;		/* = as described for sfs_hostinfo2 */
  sfs_hash sessid;		/* = SHA-1 (sfs_sessinfo) */
};

/*
 * Public key ciphertexts
 */

struct sfs_kmsg {
  sfs_secret kcs_share;
  sfs_secret ksc_share;
};

/*
 * Signed messages
 */

struct sfs_keycert_msg {
  sfs_msgtype type;		/* = SFS_KEYCERT */
  unsigned duration;		/* Lifetime of certificate */
  sfs_time start;		/* Time of signature */
  sfs_pubkey key;		/* Temporary public key */
};

struct sfs_keycert {
  sfs_keycert_msg msg;
  sfs_sig sig;
};

struct sfs_signed_authreq {
  sfs_msgtype type;		/* = SFS_SIGNED_AUTHREQ */
  sfs_hash authid;		/* SHA-1 (sfs_authinfo) */
  sfs_seqno seqno;		/* Counter, value unique per authid */
  opaque usrinfo[16];		/* All 0s, or <= 15 character logname */
};

struct sfs_redirect {
  sfs_time serial;
  sfs_time expire;
  sfs_hostinfo2 hostinfo;
};

/* Note: an sfs_signed_pathrevoke with a NULL redirect (i.e. a
 * revocation certificate) always takes precedence over one with a
 * non-NULL redirect (a forwarding pointer). */
struct sfs_pathrevoke_msg {
  sfs_msgtype type;		/* = SFS_PATHREVOKE */
  sfs_hostinfo2 path;		/* Hostinfo of old self-certifying pathname */
  sfs_redirect *redirect;	/* Optional forwarding pointer */
};

struct sfs_pathrevoke {
  sfs_pathrevoke_msg msg;
  sfs_sig2 sig;
};

/*
 * RPC arguments and results
 */

typedef sfs_connectinfo sfs_connectarg;

struct sfs_connectok {
  sfs_servinfo servinfo;
  sfs_hashcharge charge;
};

union sfs_connectres switch (sfsstat status) {
 case SFS_OK:
   sfs_connectok reply;
 case SFS_REDIRECT:
   sfs_pathrevoke revoke;
 default:
   void;
};

struct sfs_encryptarg {
  sfs_hashpay payment;
  sfs_ctext kmsg;
  sfs_pubkey pubkey;
};

struct sfs_encryptarg2 {
  sfs_hashpay payment;
  sfs_ctext2 kmsg;
  sfs_pubkey2 pubkey;
};

typedef sfs_ctext sfs_encryptres;
typedef sfs_ctext2 sfs_encryptres2;

struct sfs_nfs3_subfs {
  nfspath3 path;
  nfs_fh3 fh;
};
struct sfs_nfs3_fsinfo {
  nfs_fh3 root;
  sfs_nfs3_subfs subfs<>;
};

union sfs_nfs_fsinfo switch (int vers) {
 case ex_NFS_V3:
   sfs_nfs3_fsinfo v3;
};

union sfs_fsinfo switch (int prog) {
 case ex_NFS_PROGRAM:
   sfs_nfs_fsinfo nfs;
 case 344446:   /* SFSRO_PROGRAM */
   void;
 default:
   void;
};



union sfs_opt_idname switch (bool present) {
 case TRUE:
   sfs_idname name;
 case FALSE:
   void;
};

struct sfs_idnums {
  int uid;
  int gid;
};

struct sfs_idnames {
  sfs_opt_idname uidname;
  sfs_opt_idname gidname;
};

enum sfs_loginstat {
  SFSLOGIN_OK = 0,		/* Login succeeded */
  SFSLOGIN_MORE = 1,		/* More communication with client needed */
  SFSLOGIN_BAD = 2,		/* Invalid login */
  SFSLOGIN_ALLBAD = 3,		/* Invalid login don't try again */
  SFSLOGIN_AGAIN = 4		/* Repeat request & seqno, server will
                                 * proceed differently */
};
struct sfs_loginokres {
  unsigned authno;
  opaque resmore<>;		/* If necessary, for mutual authentication */
  string hello<>;		/* To be printed on user's terminal */
};
union sfs_loginres switch (sfs_loginstat status) {
 case SFSLOGIN_OK:
   sfs_loginokres resok;
 case SFSLOGIN_MORE:
   opaque resmore<>;
 case SFSLOGIN_BAD:
   string errmsg<>;
 case SFSLOGIN_ALLBAD:
 case SFSLOGIN_AGAIN:
   void;
};
union sfs_loginres_old switch (sfs_loginstat status) {
 case SFSLOGIN_OK:
   unsigned authno;
 case SFSLOGIN_MORE:
   opaque resmore<>;
 case SFSLOGIN_BAD:
 case SFSLOGIN_ALLBAD:
 case SFSLOGIN_AGAIN:
   void;
};


struct sfs_loginarg {
  sfs_seqno seqno;
  opaque certificate<>;		/* marshalled sfs_autharg */
};


/*
 * User-authentication structures
 */

enum sfsauth_stat {
  SFSAUTH_OK = 0,
  SFSAUTH_LOGINMORE = 1,	/* More communication with client needed */
  SFSAUTH_FAILED = 2,
  SFSAUTH_LOGINALLBAD = 3,	/* Invalid login don't try again */
  SFSAUTH_NOTSOCK = 4,
  SFSAUTH_BADUSERNAME = 5,
  SFSAUTH_WRONGUID = 6,
  SFSAUTH_DENYROOT = 7,
  SFSAUTH_BADSHELL = 8,
  SFSAUTH_DENYFILE = 9,
  SFSAUTH_BADPASSWORD = 10,
  SFSAUTH_USEREXISTS = 11,
  SFSAUTH_NOCHANGES = 12,
  SFSAUTH_NOSRP = 13,
  SFSAUTH_BADSIGNATURE = 14,
  SFSAUTH_PROTOERR = 15,
  SFSAUTH_NOTTHERE = 16,
  SFSAUTH_BADAUTHID = 17,
  SFSAUTH_KEYEXISTS = 18,
  SFSAUTH_BADKEYNAME = 19
};

enum sfs_authtype {
  SFS_NOAUTH = 0,
  SFS_AUTHREQ = 1,
  SFS_AUTHREQ2 = 2,
  SFS_SRPAUTH = 3,
  SFS_UNIXPWAUTH = 4
};

struct sfs_authreq {
  sfs_pubkey usrkey;		/* Key with which signed_req signed */
  sfs_sig signed_req;		/* Recoveraby signed sfs_signed_authreq */
};

union sfs_autharg switch (sfs_authtype type) {
 case SFS_NOAUTH:
   void;
 case SFS_AUTHREQ:
   sfs_authreq req;
};

enum sfs_credtype {
  SFS_NOCRED = 0,
  SFS_UNIXCRED = 1,
  SFS_PKCRED = 2,
  SFS_GROUPSCRED = 3
};

struct sfs_unixcred {
  string username<>;
  string homedir<>;
  string shell<>;
  unsigned uid;
  unsigned gid;
  unsigned groups<>;
};

union sfsauth_cred switch (sfs_credtype type) {
 case SFS_NOCRED:
   void;
 case SFS_UNIXCRED:
   sfs_unixcred unixcred;
 case SFS_PKCRED:
   string pkhash<>;
 case SFS_GROUPSCRED:
   sfs_idname groups<>;
};

struct sfsauth_loginokres {
  sfsauth_cred cred;
  sfs_hash authid;
  sfs_seqno seqno;
};


union sfsauth_loginres switch (sfs_loginstat status) {
 case SFSLOGIN_OK:
   sfsauth_loginokres resok;
 case SFSLOGIN_MORE:
   opaque resmore<>;
 default:
   void;
};


/*
 * Secure Remote Password (SRP) protocol
 */

struct sfssrp_parms {
  bigint N;			/* Prime */
  bigint g;			/* Generator */
};

union sfsauth_srpparmsres switch (sfsauth_stat status) {
 case SFSAUTH_OK:
   sfssrp_parms parms;
 default:
   void;
};

typedef opaque sfssrp_bytes<>;
struct sfssrp_init_arg {
  string username<>;
  sfssrp_bytes msg;
};

union sfsauth_srpres switch (sfsauth_stat status) {
 case SFSAUTH_OK:
   sfssrp_bytes msg;
 default:
   void;
};

struct sfsauth_fetchresok {
  string privkey<>;
  sfs_hash hostid;
};

union sfsauth_fetchres switch (sfsauth_stat status) {
 case SFSAUTH_OK:
   sfsauth_fetchresok resok;
 default:
   void;
};

typedef string sfsauth_realm<>;
typedef string sfsauth_certpath<>;
typedef sfsauth_certpath sfsauth_certpaths<>;

enum sfs_certstat {
  SFSAUTH_CERT_SELF = 0,	/* Single machine authentication */
  SFSAUTH_CERT_REALM = 1	/* Realm-based authentication */
};

union sfsauth_certinfo_info switch (sfs_certstat status) {
 case SFSAUTH_CERT_SELF:
   void;
 case SFSAUTH_CERT_REALM:
   sfsauth_certpaths certpaths;
};

struct sfsauth_certinfores {
  sfsauth_realm name;
  sfsauth_certinfo_info info;
};

struct sfsauth_srpinfo {
  string info<>;
  string privkey<>;
};

struct sfsauth_registermsg {
  sfs_msgtype type;		/* = SFS_AUTHREGISTER */
  string username<>;		/* logname */
  string password<>;		/* password for an add */
  sfs_pubkey pubkey;
  sfsauth_srpinfo *srpinfo;
};

struct sfsauth_registerarg {
  sfsauth_registermsg msg;
  sfs_sig sig;
};

enum sfsauth_registerres {
  SFSAUTH_REGISTER_OK = 0,
  SFSAUTH_REGISTER_NOTSOCK = 1,
  SFSAUTH_REGISTER_BADUSERNAME = 2,
  SFSAUTH_REGISTER_WRONGUID = 3,
  SFSAUTH_REGISTER_DENYROOT = 4,
  SFSAUTH_REGISTER_BADSHELL = 5,
  SFSAUTH_REGISTER_DENYFILE = 6,
  SFSAUTH_REGISTER_BADPASSWORD = 7,
  SFSAUTH_REGISTER_USEREXISTS = 8,
  SFSAUTH_REGISTER_FAILED = 9,
  SFSAUTH_REGISTER_NOCHANGES = 10,
  SFSAUTH_REGISTER_NOSRP = 11,
  SFSAUTH_REGISTER_BADSIG = 12
};

struct sfsauth_updatemsg {
  sfs_msgtype type;		/* = SFS_AUTHUPDATE */
  sfs_hash authid;		/* SHA-1 (sfs_authinfo);
				   service is SFS_AUTHSERV */
  sfs_pubkey oldkey;
  sfs_pubkey newkey;
  sfsauth_srpinfo *srpinfo;
  /* maybe username? */
};

struct sfsauth_updatearg {
  sfsauth_updatemsg msg;
  sfs_sig osig;		/* computed with sfsauth_updatereq.oldkey */
  sfs_sig nsig;		/* computed with sfsauth_updatereq.newkey */
};

program SFS_PROGRAM {
	version SFS_VERSION {
		void 
		SFSPROC_NULL (void) = 0;

		sfs_connectres
		SFSPROC_CONNECT (sfs_connectarg) = 1;

		sfs_encryptres
		SFSPROC_ENCRYPT (sfs_encryptarg) = 2;

		FSINFO
		SFSPROC_GETFSINFO (void) = 3;

		sfs_loginres
		SFSPROC_LOGIN (sfs_loginarg) = 4;

		void
		SFSPROC_LOGOUT (unsigned) = 5;

		sfs_idnames
		SFSPROC_IDNAMES (sfs_idnums) = 6;

		sfs_idnums
		SFSPROC_IDNUMS (sfs_idnames) = 7;

		sfsauth_cred
		SFSPROC_GETCRED (void) = 8;

		sfs_encryptres2
		SFSPROC_ENCRYPT2 (sfs_encryptarg2) = 9;
	} = 1;
} = 344440;

program SFSCB_PROGRAM {
	version SFSCB_VERSION {
		void 
		SFSCBPROC_NULL(void) = 0;
	} = 1;
} = 344441;

program SFSAUTH_PROGRAM {
	version SFSAUTH_VERSION {
		void 
		SFSAUTHPROC_NULL (void) = 0;

		sfsauth_loginres
		SFSAUTHPROC_LOGIN (sfs_loginarg) = 1;

		sfsauth_stat
		SFSAUTHPROC_REGISTER (sfsauth_registerarg) = 2;

		sfsauth_stat
		SFSAUTHPROC_UPDATE (sfsauth_updatearg) = 3;

		sfsauth_srpparmsres
		SFSAUTHPROC_SRP_GETPARAMS (void) = 4;

		sfsauth_srpres
		SFSAUTHPROC_SRP_INIT (sfssrp_init_arg) = 5;

		sfsauth_srpres
		SFSAUTHPROC_SRP_MORE (sfssrp_bytes) = 6;

		sfsauth_fetchres
		SFSAUTHPROC_FETCH (void) = 7;

		sfsauth_certinfores
		SFSAUTHPROC_CERTINFO (void) = 8;
	} = 1;
} = 344442;
