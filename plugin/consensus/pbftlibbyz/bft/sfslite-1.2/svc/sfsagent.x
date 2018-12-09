/* $Id: sfsagent.x 1754 2006-05-19 20:59:19Z max $ */

/*
 * This file was written by David Mazieres and Michael Kaminsky.  Its
 * contents is uncopyrighted and in the public domain.  Of course,
 * standards of academic honesty nonetheless prevent anyone in
 * research from falsely claiming credit for this work.
 */

%#include "sfs_prot.h"

typedef string sfs_filename<255>;

struct sfsagent_authinit_arg_old {
  int ntries;
  string requestor<>;
  sfs_authinfo authinfo;
  sfs_seqno seqno;
};

struct sfsagent_authinit_arg {
  int ntries;
  string requestor<>;
  sfs_authinfo authinfo;
  sfs_seqno seqno;
  sfs_idname user;
  unsigned server_release;
};

struct sfsagent_authmore_arg {
  sfs_authinfo authinfo;
  sfs_seqno seqno;
  bool checkserver;		/* To request mutual authentication at end */
  opaque more<>;
};

union sfsagent_auth_res switch (bool authenticate) {
case TRUE:
  opaque certificate<>;
case FALSE:
  void;
};

typedef string sfsagent_path<1024>;
enum sfsagent_lookup_type {
  LOOKUP_NOOP = 0,
  LOOKUP_MAKELINK = 1,
  LOOKUP_MAKEDIR  = 2
};
union sfsagent_lookup_res switch (sfsagent_lookup_type type) {
 case LOOKUP_NOOP:
   void;
 case LOOKUP_MAKELINK:
   sfsagent_path path;
 case LOOKUP_MAKEDIR:
   sfsagent_path dir;
};

enum sfs_revocation_type {
  REVOCATION_NONE = 0,
  REVOCATION_BLOCK = 1,
  REVOCATION_CERT = 2
};
union sfsagent_revoked_res switch (sfs_revocation_type type) {
 case REVOCATION_NONE:
   void;
 case REVOCATION_BLOCK:
   void;
 case REVOCATION_CERT:
   sfs_pathrevoke cert;
};
	
struct sfsagent_symlink_arg {
  sfs_filename name;
  sfsagent_path contents;
};

typedef opaque sfsagent_seed[48];
const sfs_badgid = -1;


union sfs_privkey2_clear switch (sfs_keytype type) {
 case SFS_RABIN:
   sfs_rabin_priv_xdr rabin;
 case SFS_1SCHNORR:
   sfs_1schnorr_priv_xdr schnorr1;
 case SFS_2SCHNORR:
   sfs_2schnorr_priv_xdr schnorr2;
 case SFS_ESIGN:
   sfs_esign_priv_xdr esign;
};

typedef string sfsagent_comment<1023>;
struct sfs_addkey_arg {
  sfs_privkey2_clear privkey;
  int key_version;
  sfs_time expire;
  sfsagent_comment name;
};

struct sfsagent_addextauth_arg {
  sfs_time expire;
  int pid;
  sfsagent_comment name;
};
struct sfsextauth_init {
  sfsagent_comment name;
  sfsagent_authinit_arg autharg;
};
struct sfsextauth_more {
  sfsagent_comment name;
  sfsagent_authmore_arg autharg;
};

enum sfs_remauth_type {
  SFS_REM_PUBKEY,
  SFS_REM_NAME,
  SFS_REM_PID
};
union sfs_remauth_arg switch (sfs_remauth_type type) {
 case SFS_REM_PUBKEY:
   sfs_pubkey2 pubkey;
 case SFS_REM_NAME:
   sfsagent_comment name;
 case SFS_REM_PID:
   int pid;
};

struct sfs_keylistelm {
  string desc<>;                /* Long string description */
  sfs_time expire;
  sfsagent_comment name;
  sfs_keylistelm *next;
};
typedef sfs_keylistelm *sfs_keylist;

typedef string sfsagent_progarg<>;
typedef sfsagent_progarg sfsagent_cmd<>;

struct sfsagent_certprog {
  string prefix<>;		/* Prefix path to match against path name */
  string filter<>;		/* Regular expression filter on path name */
  string exclude<>;		/* Regular expression filter on path name */
  sfsagent_cmd av;		/* External program to run */
};

typedef sfsagent_certprog sfsagent_certprogs<>;

struct sfsagent_blockfilter {
  string filter<>;		/* Regular expression filter on hostname */
  string exclude<>;		/* Regular expression filter on hostname */
};
struct sfsagent_revokeprog {
  sfsagent_blockfilter *block;	/* Block hostid even without revocation cert */
  sfsagent_cmd av;		/* External program to run */
};

typedef sfsagent_revokeprog sfsagent_revokeprogs<>;

typedef sfs_hash sfsagent_norevoke_list<>;

typedef string sfsagent_srpname<>;
struct sfsagent_srpname_pair {
  sfsagent_srpname srpname;	/* user@name.domain */
  sfs_hostname sfsname;		/* self-certifying hostname */
};

typedef sfsagent_srpname_pair sfsagent_srpname_pairs<>;

union sfsagent_srpname_res switch (bool status) {
 case TRUE:
   sfs_hostname sfsname;
 case FALSE:
   void;
};

typedef sfsagent_cmd sfsagent_confprog;
typedef sfsagent_cmd sfsagent_srpcacheprog;

struct sfsagent_rex_arg {
  sfs_hostname dest;            /* destination user named on cmdline */
  sfs_hostname schost;          /* corresponding self-certifying hostname */
  bool forwardagent;
  bool blockactive; 
  bool resumable;
};

struct sfsagent_rex_resok {
  sfs_hash sessid;
  sfs_seqno seqno;
  sfs_hash newsessid;
  opaque kcs<>;
  opaque ksc<>;
};

union sfsagent_rex_res switch (bool status) {
 case TRUE:
   sfsagent_rex_resok resok;
 case FALSE:
   void;
};

struct rex_sessentry {
  sfs_hostname dest;            /* destination user named on cmdline */
  sfs_hostname schost;          /* corresponding self-certifying hostname */
  sfs_hostname created_from;
  bool agentforwarded;
};

typedef rex_sessentry rex_sessvec<>;

struct sfsctl_getfh_arg {
  filename3 filesys;
  u_int64_t fileid;
};

union sfsctl_getfh_res switch (nfsstat3 status) {
 case NFS3_OK:
   nfs_fh3 fh;
 default:
   void;
};

struct sfsctl_getidnames_arg {
  filename3 filesys;
  sfs_idnums nums;
};

union sfsctl_getidnames_res switch (nfsstat3 status) {
 case NFS3_OK:
   sfs_idnames names;
 default:
   void;
};

struct sfsctl_getidnums_arg {
  filename3 filesys;
  sfs_idnames names;
};

union sfsctl_getidnums_res switch (nfsstat3 status) {
 case NFS3_OK:
   sfs_idnums nums;
 default:
   void;
};

union sfsctl_getcred_res switch (nfsstat3 status) {
 case NFS3_OK:
   sfsauth_cred cred;
 default:
   void;
};

struct sfsctl_lookup_arg {
  filename3 filesys;
  diropargs3 arg;
};

struct sfsctl_getacl_arg {
  filename3 filesys;
  diropargs3 arg;
};

struct sfsctl_setacl_arg {
  filename3 filesys;
  setaclargs arg;
};

program AGENTCTL_PROG {
	version AGENTCTL_VERS {
		void
		AGENTCTL_NULL (void) = 0;

		bool
		AGENTCTL_ADDKEY (sfs_addkey_arg) = 1;

		bool
		AGENTCTL_REMAUTH (sfs_remauth_arg) = 2;

		void
		AGENTCTL_REMALLKEYS (void) = 3;

		sfs_keylist
		AGENTCTL_DUMPKEYS (void) = 4;

		void
		AGENTCTL_CLRCERTPROGS (void) = 5;

		bool
		AGENTCTL_ADDCERTPROG (sfsagent_certprog) = 6;

		sfsagent_certprogs
		AGENTCTL_DUMPCERTPROGS (void) = 7;

		void
		AGENTCTL_CLRREVOKEPROGS (void) = 8;

		bool
		AGENTCTL_ADDREVOKEPROG (sfsagent_revokeprog) = 9;

		sfsagent_revokeprogs
		AGENTCTL_DUMPREVOKEPROGS (void) = 10;

		void
		AGENTCTL_SETNOREVOKE (sfsagent_norevoke_list) = 11;

		sfsagent_norevoke_list
		AGENTCTL_GETNOREVOKE (void) = 12;

		void
		AGENTCTL_SYMLINK (sfsagent_symlink_arg) = 13;

		void
		AGENTCTL_RESET (void) = 14;

		int
		AGENTCTL_FORWARD (sfs_hostname) = 15;

		void
		AGENTCTL_RNDSEED (sfsagent_seed) = 16;

		sfsagent_rex_res
		AGENTCTL_REX (sfsagent_rex_arg) = 17;

		rex_sessvec
		AGENTCTL_LISTSESS (void) = 18;

		bool
		AGENTCTL_KILLSESS (sfs_hostname) = 19;

		bool
		AGENTCTL_CLRCERTPROG_BYREALM (sfsauth_realm) = 20;

		bool
		AGENTCTL_ADDEXTAUTH (sfsagent_addextauth_arg) = 21;

		void
		AGENTCTL_CLRSRPNAMES (void) = 22;

		bool
		AGENTCTL_ADDSRPNAME (sfsagent_srpname_pair) = 23;

		sfsagent_srpname_pairs
		AGENTCTL_DUMPSRPNAMES (void) = 24;

		sfsagent_srpname_res
		AGENTCTL_LOOKUPSRPNAME (sfsagent_srpname) = 25;

                void
                AGENTCTL_CLRCONFIRMPROG (void) = 26;

                bool
                AGENTCTL_ADDCONFIRMPROG (sfsagent_confprog) = 27;

                sfsagent_confprog
                AGENTCTL_DUMPCONFIRMPROG (void) = 28;

                void
                AGENTCTL_CLRSRPCACHEPROG (void) = 29;

                bool
                AGENTCTL_ADDSRPCACHEPROG (sfsagent_srpcacheprog) = 30;

                sfsagent_srpcacheprog
                AGENTCTL_DUMPSRPCACHEPROG (void) = 31;

                bool
                AGENTCTL_KEEPALIVE (sfs_hostname) = 32;

		void
		AGENTCTL_KILL (void) = 33;
	} = 1;
} = 344428;

program SFSEXTAUTH_PROG {
	version SFSEXTAUTH_VERS {
		void
		SFSEXTAUTH_NULL (void) = 0;

		sfsagent_auth_res
		SFSEXTAUTH_AUTHINIT (sfsextauth_init) = 1;

		sfsagent_auth_res
		SFSEXTAUTH_AUTHMORE (sfsextauth_more) = 2;
	} = 1;
} = 344429;

program SETUID_PROG {
	version SETUID_VERS {
		/* Note:  SETUIDPROC_SETUID requires an authunix AUTH. */
		int SETUIDPROC_SETUID (void) = 0;
	} = 1;
} = 344430;

program AGENT_PROG {
	version AGENT_VERS {
		void
		AGENT_NULL (void) = 0;

		int
		AGENT_START (void) = 1;

		int
		AGENT_KILL (void) = 2;

		int
		AGENT_KILLSTART (void) = 3;

		void
		AGENT_SYMLINK (sfsagent_symlink_arg) = 4;

		void
		AGENT_FLUSHNAME (sfs_filename) = 5;

		void
		AGENT_FLUSHNEG (void) = 6;

		void
		AGENT_REVOKE (sfs_pathrevoke) = 7;

		sfsagent_seed
		AGENT_RNDSEED (void) = 8;

		unsigned
		AGENT_AIDALLOC (void) = 9;

		int
		AGENT_GETAGENT (void) = 10;
	} = 1;
} = 344432;

program AGENTCB_PROG {
	version AGENTCB_VERS {
		void
		AGENTCB_NULL (void) = 0;

		sfsagent_auth_res
		AGENTCB_AUTHINIT (sfsagent_authinit_arg) = 1;

		sfsagent_auth_res
		AGENTCB_AUTHMORE (sfsagent_authmore_arg) = 2;

		sfsagent_lookup_res
		AGENTCB_LOOKUP (sfs_filename) = 3;

		sfsagent_revoked_res
		AGENTCB_REVOKED (filename3) = 4;

		void
		AGENTCB_CLONE (void) = 5;

	} = 1;
} = 344433;

program SFSCTL_PROG {
	version SFSCTL_VERS {
		void
		SFSCTL_NULL (void) = 0;

		void
		SFSCTL_SETPID (int) = 1;

		sfsctl_getfh_res
		SFSCTL_GETFH (sfsctl_getfh_arg) = 2;

		sfsctl_getidnames_res
		SFSCTL_GETIDNAMES (sfsctl_getidnames_arg) = 3;

		sfsctl_getidnums_res
		SFSCTL_GETIDNUMS (sfsctl_getidnums_arg) = 4;

		sfsctl_getcred_res
		SFSCTL_GETCRED (filename3) = 5;

		lookup3res
		SFSCTL_LOOKUP (sfsctl_lookup_arg) = 6;

		read3res
		SFSCTL_GETACL (sfsctl_getacl_arg) = 7;

		write3res
		SFSCTL_SETACL (sfsctl_setacl_arg) = 8;
	} = 1;
} = 344434;

