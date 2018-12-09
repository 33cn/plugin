/* $Id: spawn.C 3769 2008-11-13 20:21:34Z max $ */

/*
 *
 * Copyright (C) 1998 David Mazieres (dm@uun.org)
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

#include "amisc.h"
#include "rxx.h"
#include <dirent.h>

str execdir (EXECDIR);
#ifdef MAINTAINER
str builddir;
str buildtmpdir;
#endif /* MAINTAINER */

static void
nop (int)
{
}

#ifdef MAINTAINER
bool afork_debug = safegetenv ("AFORK_DEBUG");
#endif /* MAINTAINER */

pid_t
afork ()
{
  if (pid_t pid = fork ())
    return pid;

  fatal_no_destruct = true;
  err_reset ();

  /* If we close/dup2 stderr before an exec, but the exec fails, we
   * still want warn/fatal to work so as to report the error. */
  if (errfd == 2) {
    int fd = dup (errfd);
    if (fd < 3)
      close (fd);
    else {
      close_on_exec (fd);
      errfd = fd;
    }
  }

  /* If we exec, we want the child to get SIGPIPEs again.  The signal
   * mask and SIG_IGN signals are preserved by execve, but not
   * handlers for specific functions.  Thus we change the handler from
   * the probably more efficient SIG_IGN to an empty function. */
  struct sigaction sa;
  bzero (&sa, sizeof (sa));
  sa.sa_handler = nop;
  sigaction (SIGPIPE, &sa, NULL);

#ifdef MAINTAINER
  if (afork_debug) {
    warn ("AFORK_DEBUG: child process pid %d\n", getpid ());
    sleep (7);
  }
#endif /* MAINTAINER */

  return 0;
}

inline bool
execok (const char *path)
{
  struct stat sb;
  return !stat (path, &sb) && S_ISREG (sb.st_mode) && (sb.st_mode & 0111);
}

#ifdef MAINTAINER
static str
searchdir (str dir, str prog)
{
  DIR *dirp = opendir (dir);
  if (!dirp)
    return NULL;
  while (dirent *dp = readdir (dirp)) {
    struct stat sb;
    str file = dir << "/" << dp->d_name;
    str np;
    if (!stat (file, &sb) && S_ISDIR (sb.st_mode)
	&& execok (np = file << prog)) {
      closedir (dirp);
      return np;
    }
  }
  closedir (dirp);
  return NULL;
}
#endif /* MAINTAINER */

str
fix_exec_path (str path, str dir)
{
  const char *prog = strrchr (path, '/');
  if (prog)
    return path;

  if (!dir)
    dir = execdir;
  path = dir << "/" << path;
  prog = strrchr (path, '/');
  if (!prog)
    panic ("No EXECDIR for unqualified path %s.\n", path.cstr ());

#ifdef MAINTAINER
  if (builddir && dir == execdir) {
    str np;
    np = builddir << prog;
    if (execok (np))
      return np;
    np = builddir << prog << prog;
    if (execok (np))
      return np;
    if (np = searchdir (builddir, prog))
      return np;
    if (np = searchdir (builddir << "/lib", prog))
      return np;
  }
#endif /* MAINTAINER */

  return path;
}

str
find_program (const char *program)
{
  static rxx colonplus (":+");
  str r;
  if (strchr (program, '/')) {
    r = program;
    if (execok (r))
      return r;
    return NULL;
  }

#ifdef MAINTAINER
  if (builddir) {
    r = fix_exec_path (program);
    if (execok (r))
      return r;
  }
#endif /* MAINTAINER */
  if (progdir) {
    r = progdir << program;
    if (execok (r))
      return r;
  }

  const char *p = getenv ("PATH");
  if (!p)
    return NULL;
  vec<str> vs;
  split (&vs, colonplus, p);
  for (str *sp = vs.base (); sp < vs.lim (); sp++) {
    if (!*sp || !sp->len ())
      continue;
    r = *sp << "/" << program;
    if (execok (r))
      return r;
  }
  return NULL;
}

str
find_program_plus_libsfs (const char *program)
{
  str s = fix_exec_path (program);
  if (!s || !execok (s))
    s = find_program (program);
  return s;
}


void
setstdfds (int in, int out, int err)
{
  if (in != 0) {
    dup2 (in, 0);
    if (in > 2 && in != out && in != err)
      close (in);
  }
  if (out != 1) {
    dup2 (out, 1);
    if (out > 2 && out != err)
      close (out);
  }
  if (err != 2) {
    dup2 (err, 2);
    if (err > 2)
      close (err);
  }
}

extern bool amain_panic;
pid_t
spawn (const char *path, char *const *argv,
       int in, int out, int err, cbv::ptr postforkcb, char *const *env)
{
  int fds[2];

  if (pipe (fds) < 0)
    return -1;

  close_on_exec (fds[0]);
  close_on_exec (fds[1]);

  pid_t pid = afork ();
  if (pid < 0)
    return pid;

  if (!pid) {
    amain_panic = true;
    close (fds[0]);
    setstdfds (in, out, err);
    if (postforkcb)
      (*postforkcb) ();
    if (env)
      execve (path, argv, env);
    else
      execv (path, argv);

    int saved_err = errno;
    v_write (fds[1], &saved_err, sizeof (saved_err));
    close (fds[1]);
    // Since we return a useful errno, there is no need to print
    // anything in the child process.  Just exit.
    _exit (1);
  }

  close (fds[1]);
  int chld_err;
  int n = read (fds[0], &chld_err, sizeof (chld_err));
  close (fds[0]);

  if (n == 0)
    return pid;
  errno = chld_err;
  return -1;
}

pid_t
aspawn (const char *path, char *const *argv,
	int in, int out, int err, cbv::ptr postforkcb, char *const *env)
{
  pid_t pid = afork ();
  if (pid < 0)
    return pid;

  if (pid)
    return pid;

  setstdfds (in, out, err);

  if (postforkcb)
    (*postforkcb) ();
  if (env)
    execve (path, argv, env);
  else
    execv (path, argv);

  warn ("%s: %m\n", path);
  _exit (1);
}

