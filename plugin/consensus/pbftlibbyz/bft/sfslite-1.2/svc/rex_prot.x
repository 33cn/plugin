/* $Id: rex_prot.x 1754 2006-05-19 20:59:19Z max $ */

/*
 * This file was written by David Mazieres.  Its contents is
 * uncopyrighted and in the public domain.  Of course, standards of
 * academic honesty nonetheless prevent anyone in research from
 * falsely claiming credit for this work.
 */

%#include "sfs_prot.h"

%#include "dns.h"
%#ifndef MAXHOSTNAMELEN
%#define MAXHOSTNAMELEN 256
%#endif /* !MAXHOSTNAMELEN */

typedef string ttypath<63>;
typedef string utmphost<MAXHOSTNAMELEN>;

/* Note, a successful PTYD_PTY_ALLOC result is accompanied by a file
 * descriptor for the master side of the pty. */
union pty_alloc_res switch (int err) {
 case 0:
   ttypath path;
 default:
   void;
};

program PTYD_PROG {
	version PTYD_VERS {
		void
		PTYD_NULL (void) = 0;

		pty_alloc_res
		PTYD_PTY_ALLOC (utmphost) = 1;

		int
		PTYD_PTY_FREE (ttypath) = 2;
	} = 1;
} = 344431;

typedef string rex_progarg<>;
typedef rex_progarg rex_cmd<>;
typedef string rex_envvar<>;
typedef rex_envvar rex_env<>;

struct rexd_spawn_arg {
  sfs_kmsg kmsg;
  rex_cmd command;
};

struct rexd_spawn_resok {
  sfs_kmsg kmsg;
};

union rexd_spawn_res switch (sfsstat err) {
 case SFS_OK:
   rexd_spawn_resok resok;
 default:
   void;
};

struct rexd_attach_arg {
  sfs_hash sessid;
  sfs_seqno seqno;
  sfs_hash newsessid;
};

#if 0
struct rexd_attach_resok {
};

union rexd_attach_res switch (sfsstat err) {
 case SFS_OK:
   rexd_attach_resok resok;
 default:
   void;
};
#else
typedef sfsstat rexd_attach_res;
#endif

union rex_setresumable_arg switch (bool resumable) {
  case TRUE:
    opaque secretid<>;
  default:
    void;
};

struct rex_resume_arg {
  sfs_seqno seqno;
  opaque secretid<>;
};

/*
 * To compute session keys, the XDR-marshalled contents of this
 * structure is put through HMAC-SHA1.  The HMAC key is concatenation
 * of the appropriate kmsg fields from rexd_spawn_resok and
 * rexd_spawn_arg (in that order--server shares before client shares).
 */
struct rex_sesskeydat {
  sfs_msgtype type;		/* = SFS_KSC or SFS_KCS */
  sfs_seqno seqno;
};

struct rexctl_connect_arg {
  sfs_seqno seqno;
  sfs_sessinfo si;
};

program REXD_PROG {
	version REXD_VERS {
		void
		REXD_NULL (void) = 0;

		/* Must be accompanied by a previously negotiated authuint */
		rexd_spawn_res
		REXD_SPAWN (rexd_spawn_arg) = 1;

		rexd_attach_res
		REXD_ATTACH (rexd_attach_arg) = 2;
	} = 1;
} = 344424;

program REXCTL_PROG {
	version REXCTL_VERS {
		void
		REXCTL_NULL (void) = 0;

		/* REXCTL_CONNECT is preceeded by a file descriptor */
		void
		REXCTL_CONNECT (rexctl_connect_arg) = 1;
	} = 1;
} = 344426;

struct rex_payload {
  unsigned channel;
  int fd;
  opaque data<>;
};

struct rex_mkchannel_arg {
  int nfds;
  rex_cmd av;
  rex_env env;
};

struct rex_mkchannel_resok {
  unsigned channel;
};

union rex_mkchannel_res switch (sfsstat err) {
 case SFS_OK:
   rex_mkchannel_resok resok;
 default:
   void;
};

struct rex_int_arg {
  unsigned channel;
  int val;
};

struct rex_newfd_arg {
  unsigned channel;
  int fd;
};

union rex_newfd_res switch (bool ok) {
  case TRUE:
    int newfd;
  case FALSE:
    void;
};

struct rexcb_newfd_arg {
  unsigned channel;
  int fd;
  int newfd;
};

struct rex_setenv_arg {
  string name<>;
  string value<>;
};


typedef string rex_unsetenv_arg<>;

typedef string rex_getenv_arg<>;

union rex_getenv_res switch (bool present) {
  case TRUE:
    string value<>;
  case FALSE:
    void;
};

program REX_PROG {
	version REX_VERS {
		void
		REX_NULL (void) = 0;

		bool
		REX_DATA (rex_payload) = 1;

		/* val is fd to close, or -1 to close channel */
		bool
		REX_CLOSE (rex_int_arg) = 2;

		/* val is signal to deliver */
		bool
		REX_KILL (rex_int_arg) = 3;

		rex_mkchannel_res
		REX_MKCHANNEL (rex_mkchannel_arg) = 4;

	        bool
		REX_SETENV (rex_setenv_arg) = 5;

	        void	
		REX_UNSETENV (rex_unsetenv_arg) = 6;

		rex_newfd_res
		REX_NEWFD (rex_newfd_arg) = 7;

		bool
		REX_RESUME (rex_resume_arg) = 8;

                void
                REX_SETRESUMABLE (rex_setresumable_arg) = 9;

                void
                REX_CLIENT_DIED (sfs_seqno) = 10;

	        rex_getenv_res
		REX_GETENV (rex_getenv_arg) = 11;
	} = 1;
} = 344428;

program REXCB_PROG {
	version REXCB_VERS {
		void
		REXCB_NULL (void) = 0;

		bool
		REXCB_DATA (rex_payload) = 1;

		bool
		REXCB_NEWFD (rexcb_newfd_arg) = 2;

		/* val is exit status or -1 for signal */
		void
		REXCB_EXIT (rex_int_arg) = 3;
	} = 1;
} = 344429;
