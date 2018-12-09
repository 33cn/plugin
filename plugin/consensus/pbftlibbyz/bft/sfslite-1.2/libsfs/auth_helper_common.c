/* $Id: auth_helper_common.c 2091 2006-07-17 19:29:32Z max $ */

/*
 *
 * Copyright (C) 2004 David Mazieres (dm@uun.org)
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU General Public License as
 * published by the Free Software Foundation; either version 2, or (at
 * your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 59 Temple Place, Suite 330, Boston, MA 02111-1307
 * USA
 *
 */

#include "sfs-internal.h"
#include "auth_helper.h"
#include <sys/types.h>
#include <unistd.h>
#include <stdio.h>

static int use_rpc;		/* For debugging */
char *progname;
char *service;
char *user;

#define PW_BUF_SIZE 256

static char *
authhelp_input_tty (char *prompt, int echo)
{
  printf ("%s", prompt);
  fflush (stdout);
  if (!echo)
    return strdup (getpass (""));
  else {
    char *buf = malloc (PW_BUF_SIZE);
    int n = read (0, buf, PW_BUF_SIZE);
    if (n <= 0) {
      fprintf (stderr, "%s(%s,%s): EOF\n", progname,
	       user ? user : "unknown user", service);
      exit (1);
    }
    if (n >= PW_BUF_SIZE) {
      fprintf (stderr, "%s(%s,%s): input too long\n",
	       progname, user ? user : "unknown user", service);
      exit (1);
    }
    buf[n] = '\0';
    while (n > 0 && (buf[n-1] == '\r' || buf[n-1] == '\n'))
      buf[--n] = '\0';
    return buf;
  }
}

static void
authhelp_approve_tty (char *u, char *hello)
{
  if (hello)
    fprintf (stderr,"%s\n", hello);
  fprintf (stderr, "%s(%s,%s): authenticated %s\n",
	   progname, user, service, u);
}

static char *
authhelp_input_rpc (char *prompt, int echo)
{
  authhelp_getpass_arg arg;
  authhelp_getpass_res res;
  enum clnt_stat err;

  bzero (&arg, sizeof (arg));
  bzero (&res, sizeof (res));
  arg.prompt = prompt;
  arg.echo = echo;

  err = srpc_call (&authhelp_prog_1, 0, AUTHHELPPROG_GETPASS, &arg, &res);
  if (err) {
    fprintf (stderr, "%s(%s,%s): %s\n",
	     progname, user ? user : "unknown user",
	     service, clnt_sperrno (err));
    exit (1);
  }

  return res.response;
}

static void
authhelp_approve_rpc (char *u, char *hello)
{
  authhelp_succeed_arg arg;
  bzero (&arg, sizeof (arg));
  arg.user = u;
  arg.hello = hello ? hello : "";
  srpc_call (&authhelp_prog_1, 0, AUTHHELPPROG_SUCCEED, &arg, NULL);
}

char *
authhelp_input (char *prompt, int echo)
{
  if (use_rpc)
    return authhelp_input_rpc (prompt, echo);
  else
    return authhelp_input_tty (prompt, echo);
}

void
authhelp_approve (char *user, char *hello)
{
  if (use_rpc)
    authhelp_approve_rpc (user, hello);
  else
    authhelp_approve_tty (user, hello);
}

static void usage (void) __attribute__ ((noreturn));
static void
usage (void)
{
  fprintf (stderr, "usage: %s [-r] <service> [<user>]\n", progname);
  exit (1);
}

int
main (int argc, char **argv)
{
  if ((progname = strrchr (argv[0], '/')))
    progname++;
  else
    progname = argv[0];

  /* N.B.:  We explicitly do not want to use getopt here, because
   * Linux getopt sometimes does some very weird and un-posix things,
   * which could cause late arguments (after the service) to be
   * interpreted as command line arguments.  We definitely want to
   * avoid any kind of "login -froot" type bugs here.
   */
  optind = 1;
  if (optind < argc && !strcmp (argv[optind], "--"))
    optind++;
  else if (optind < argc && !strcmp (argv[optind], "-r")) {
    use_rpc = 1;
    optind++;
  }
  else if (optind < argc && argv[optind][0] == '-')
    usage ();

  if (optind + 2 == argc)
    user = argv[optind+1];
  else if (optind + 1 != argc)
    usage ();

  service = argv[optind];

  if (!user)
    user = authhelp_input ("Username: ", 1);

  exit (authhelp_go () != 0);
}
