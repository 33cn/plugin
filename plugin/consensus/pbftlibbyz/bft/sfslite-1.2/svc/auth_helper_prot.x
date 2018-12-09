/* $Id: auth_helper_prot.x 1754 2006-05-19 20:59:19Z max $ */

/*
 * This file was written by David Mazieres.  Its contents is
 * uncopyrighted and in the public domain.  Of course, standards of
 * academic honesty nonetheless prevent anyone in research from
 * falsely claiming credit for this work.
 */

struct authhelp_getpass_arg {
  string prompt<>;
  bool echo;
};

struct authhelp_getpass_res {
  string response<>;
};

struct authhelp_succeed_arg {
  string user<>;
  string hello<>;
};

program AUTHHELP_PROG {
	version AUTHHELP_VERS {
		void
		AUTHHELPPROG_NULL (void) = 0;

		authhelp_getpass_res
		AUTHHELPPROG_GETPASS (authhelp_getpass_arg) = 1;

		void
		AUTHHELPPROG_SUCCEED (authhelp_succeed_arg) = 2;
	} = 1;
} = 344446;
