/* $Id: daemonize.C 3146 2007-12-20 17:44:02Z max $ */

/*
 *
 * Copyright (C) 1999 David Mazieres (dm@uun.org)
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

#include "async.h"

str syslog_priority ("daemon.notice");

static int
start_log_to_file (const str &line, const str &logfile, int flags, mode_t m)
{
  int fd;
  int n;
  if ((fd = open (logfile.cstr (), flags, m)) < 0) {
    warn ("%s: %m\n", logfile.cstr ());
    fd = errfd;
  } else {
    warn << "Logging via logfile: " << logfile << "\n";
    if (line) {
      n = write (fd, line.cstr (), line.len ());
      if (n < (int (line.len ()))) 
	warn << logfile << ": write to logfile failed\n";
    }
  } 
  return fd;
}

int
start_logger (const str &priority, const str &tag, const str &line, 
	      const str &logfile, int flags, mode_t mode)
{
#ifdef PATH_LOGGER
  const char *av[] = { PATH_LOGGER, "-p", NULL, "-t", NULL, NULL, NULL };
  av[2] = const_cast<char *> (priority.cstr ());
  
  if (line)
    av[5] = const_cast<char *> (line.cstr ());
  else
    av[5] = "log started";

  if (tag)
    av[4] = const_cast<char *> (tag.cstr ());
  else 
    av[4] = "";

  pid_t pid;
  int status;
  if ((pid = spawn (PATH_LOGGER, av, 0, 0, errfd)) < 0) {
    warn ("%s: %m\n", PATH_LOGGER);
    return start_log_to_file (line, logfile, flags, mode);
  } 
  if (waitpid (pid, &status, 0) <= 0 || !WIFEXITED (status) || 
      WEXITSTATUS (status)) 
    return start_log_to_file (line, logfile, flags, mode);

  int fds[2];
  if (socketpair (AF_UNIX, SOCK_STREAM, 0, fds) < 0)
    fatal ("socketpair: %m\n");
  close_on_exec (fds[0]);
  if (fds[1] != 0)
    close_on_exec (fds[1]);
  
  av[5] = NULL;
  if (spawn (PATH_LOGGER, av, fds[1], 0, 0) >= 0) {
    close (fds[1]);
    return fds[0];
  } else 
    warn ("%s: %m\n", PATH_LOGGER);
#endif /* PATH_LOGGER */
  return start_log_to_file (line, logfile, flags, mode);
}

void
start_logger ()
{
#ifdef PATH_LOGGER
  const char *av[] = { PATH_LOGGER, "-p",
		       syslog_priority.cstr (),
		       "-t", "", NULL};
  int fds[2];

  close (0);
  if (int fd = open ("/dev/null", O_RDONLY))
    close (fd);

  if (socketpair (AF_UNIX, SOCK_STREAM, 0, fds) < 0)
    fatal ("socketpair: %m\n");
  close_on_exec (fds[0]);
  if (fds[1] != 0)
    close_on_exec (fds[1]);

  if (spawn (PATH_LOGGER, av, fds[1], 0, 0) >= 0) {
    close (fds[1]);
    if (fds[0] != errfd) {
      err_flush ();		// XXX - we shouldn't depend on aerr.C
      if (dup2 (fds[0], errfd) < 0)
	fatal ("dup2: %m\n");
      close (fds[0]);
    }
    if (errfd != 1)
      dup2 (errfd, 1);
    return;
  }
  else
    warn ("%s: %m\n", PATH_LOGGER);
#endif /* PATH_LOGGER */
  
  /* No logger, at least send chatter to stdout rather than stderr, so
   * that it can be redirected. */
  dup2 (errfd, 1);
}

struct pidfile {
  const str path;
  const struct stat sb;
  pidfile (str p, struct stat s) : path (p), sb (s) {}
};
static vec<pidfile> pidfiles;

EXITFN(pidclean);
static void
pidclean ()
{
  for (; !pidfiles.empty (); pidfiles.pop_front ()) {
    pidfile &pf = pidfiles.front ();
    struct stat sb;
    if (!stat (pf.path, &sb)
	&& sb.st_dev == pf.sb.st_dev
	&& sb.st_ino == pf.sb.st_ino)
      unlink (pf.path);
  }
}

void
daemonize (const str &nm)
{
  str pidfilebase = nm;
  if (!pidfilebase)
    pidfilebase = progname;

  switch (fork ()) {
  default:
    _exit (0);
  case -1:
    fatal ("fork: %m\n");
  case 0:
    break;
  }
    
  if (setsid () == -1)
    fatal ("setsid: %m\n");
  if (!builddir) {
    start_logger ();
    str path = strbuf () << PIDDIR << "/" << pidfilebase << ".pid";
    struct stat sb;
    if (str2file (path, strbuf ("%d\n", int (getpid ())), 0444, false, &sb))
      pidfiles.push_back (pidfile (path, sb));
  }
  else {
    str piddir = buildtmpdir;
    if (!piddir)
      piddir = builddir;
    str path = strbuf () << piddir << "/" << pidfilebase << ".pid";
    struct stat sb;
    if (str2file (path, strbuf ("%d\n", int (getpid ())), 0444, false, &sb))
      pidfiles.push_back (pidfile (path, sb));
  }
}
