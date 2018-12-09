/* $Id: sfscd_prot.x 1754 2006-05-19 20:59:19Z max $ */

/*
 * This file was written by David Mazieres.  Its contents is
 * uncopyrighted and in the public domain.  Of course, standards of
 * academic honesty nonetheless prevent anyone in research from
 * falsely claiming credit for this work.
 */

%#include "sfs_prot.h"
%#include "nfsmounter.h"
%#include "sfsagent.h"

struct sfscd_initarg {
  string name<>;
};

/* If cres is present, a mount call must be preceeded by a sendfd */
struct sfscd_mountarg {
  sfs_connectarg carg;
  sfs_connectok *cres;
  string hostname<>;		/* Host to which TCP connection established */
};

/* A successful mountres must be preceeded by a sendfd */
struct sfscd_mountok {
  int mntflags;
  nfsmnt_handle fh;
};
union sfscd_mountres switch (int err) {
 case 0:
   sfscd_mountok reply;
 default:
   void;
};

union sfscd_authreq switch (int type) {
 case AGENTCB_AUTHINIT:
   sfsagent_authinit_arg init;
 case AGENTCB_AUTHMORE:
   sfsagent_authmore_arg more;
};

typedef unsigned hyper sfs_aid;

struct sfscd_agentreq_arg {
  sfs_aid aid;
  sfscd_authreq agentreq;
};

program SFSCD_PROGRAM {
	version SFSCD_VERSION {
		void
		SFSCDPROC_NULL (void) = 0;

		void
		SFSCDPROC_INIT (sfscd_initarg) = 1;

		sfscd_mountres	
		SFSCDPROC_MOUNT (sfscd_mountarg) = 2;

		void
		SFSCDPROC_UNMOUNT (nfspath3) = 3;

		void
		SFSCDPROC_FLUSHAUTH (sfs_aid) = 4;

		void
		SFSCDPROC_CONDEMN (nfspath3) = 5;
	} = 1;
} = 344438;

program SFSCDCB_PROGRAM {
	version SFSCDCB_VERSION {
		void
		SFSCDCBPROC_NULL (void) = 0;
		
		sfsagent_auth_res
		SFSCDCBPROC_AGENTREQ (sfscd_agentreq_arg) = 1;

		void
		SFSCDCBPROC_IDLE (nfspath3) = 2;

		void
		SFSCDCBPROC_DELFS (nfspath3) = 3;

		void
		SFSCDCBPROC_HIDEFS (nfspath3) = 4;

		void
		SFSCDCBPROC_SHOWFS (nfspath3) = 5;
	} = 1;
} = 344439;
