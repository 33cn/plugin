/* $Id: agentmisc.C 3773 2008-11-13 20:50:37Z max $ */

/*
 *
 * Copyright (C) 1998 David Mazieres (dm@uun.org)
 * Copyright (C) 1999 Michael Kaminsky (kaminsky@lcs.mit.edu)
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

#include "agentmisc.h"
#include "agentconn.h"
#include "crypt.h"
#include "rxx.h"
#include <dirent.h>

str dotsfs; 
str agentsock;
str userkeysdir;

inline bool
agent_userdir_ok (const char *file, u_int32_t uid)
{
  struct stat sb;
  return !lstat (file, &sb)
    && S_ISDIR (sb.st_mode) && sb.st_uid == uid
    && (sb.st_mode & 0777) == 0700;
}


static str
agent_userdir_search (const char *tmpdir, u_int32_t uid, bool create)
{
  DIR *dp = opendir (".");
  if (!dp) {
    warn ("%s: %m\n", tmpdir);
    return NULL;
  }

  str rx (strbuf ("sfs-%d-(0|[1-9]\\d{0,6})", uid));
  rxx filter (rx);

  u_int best = (u_int) -1;
  while (dirent *dep = readdir (dp)) {
    if (filter.match (dep->d_name)) {
      u_int num = atoi (filter[1]);
      str file (strbuf ("sfs-%d-%u", uid, num));
      assert (file == dep->d_name);
      if (num < best && agent_userdir_ok (file, uid))
	best = num;
    }
  }
  closedir (dp);

  if (best != (u_int) -1)
    return strbuf ("sfs-%d-%u", uid, best);
  else if (!create)
    return NULL;

  for (u_int n = 0; n < 1999999; n++) {
    str path (strbuf ("sfs-%d-%u", uid, n));
    if (!mkdir (path, 0700))
      return path;
    else if (errno != EEXIST)
      return NULL;
    /* utimes because we want to avoid having cron jobs delete the old
     * directory we are now using.  There is a tiny race condition in
     * that our utimes and agent_userdir_ok could both execute
     * between, say, lstat and exec calls in find. */
    else if (!utimes (path, NULL) && agent_userdir_ok (path, uid))
      return path;
  }
  return NULL;
}

str
agent_userdir (u_int32_t uid, bool create)
{
  int fd = open (".", O_RDONLY);
  if (fd < 0) {
    warn ("current working directory (.): %m\n");
    return NULL;
  }

  const char *tmpdir = safegetenv ("TMPDIR");
  if (!tmpdir || tmpdir[0] != '/')
    tmpdir = "/tmp";
  
  str ret;
  struct stat sb;
  uid_t myuid = getuid ();

  if (chdir (tmpdir) < 0)
    warn ("%s: %m\n", tmpdir);
  else if (stat (".", &sb) < 0)
    warn ("%s: %m\n", tmpdir);
  else if ((sb.st_mode & 022) && !(sb.st_mode & 01000))
    warn ("bad permissions on %s; chmod +t or set TMPDIR elsewhere", tmpdir);
  else if (myuid == uid)
    ret = agent_userdir_search (tmpdir, uid, create);
  else if (!myuid) {
    int fds[2];
    if (pipe (fds) < 0)
      warn ("pipe: %m\n");
    else if (pid_t pid = afork ()) {
      close (fds[1]);
      strbuf sb;
      while (sb.tosuio ()->input (fds[0]) > 0)
	;
      close (fds[0]);
      int status = 1;
      if (!waitpid (pid, &status, 0) && !status)
	ret = sb;
    }
    else {
      close (fds[0]);
      _exit (setuid (uid)
	     || !(ret = agent_userdir_search (tmpdir, uid, create))
	     || (write (fds[1], ret, ret.len ())
	         != implicit_cast<ssize_t> (ret.len ())));
    }
  }

  rc_ignore (fchdir (fd));
  close (fd);
  return ret ? str (strbuf ("%s/", tmpdir) << ret) : str (NULL);
}

str
agent_usersock (bool create_dir)
{
  str dir = agent_userdir (myaid () & 0xffffffff, create_dir);
  if (!dir)
    return NULL;
  return strbuf ("%s/agent-%" U64F "d.sock", dir.cstr (), myaid ());
}

void
agent_setsock ()
{
  if (!dotsfs) {
    const char *home = getenv ("HOME");
    if (!home)
      fatal ("No HOME environment variable set\n");
    dotsfs = strbuf ("%s/.sfs", home);
  }
  if (!userkeysdir) 
    userkeysdir = strbuf ("%s/authkeys", dotsfs.cstr ());
      
  if (!agentsock)
    if (char *p = getenv ("SFS_AGENTSOCK"))
      agentsock = p;
}

void
agent_ckdir (bool fail_on_keysdir)
{
  if (!agentsock)
    agent_setsock ();
  if (!agent_ckdir (dotsfs))
    exit (1);
  if (!agent_ckdir (userkeysdir) && fail_on_keysdir)
    exit (1);
}

bool
agent_ckdir (const str &s)
{
  assert (s);
  struct stat sb;
  if (stat (s, &sb) < 0) {
    warn ("%s: %m\n", s.cstr ());
    return false;
  }
  if (!S_ISDIR (sb.st_mode)) {
    warn << s << ": not a directory\n";
    return false;
  }
  if (sb.st_mode & 077) {
    warn ("%s: mode 0%o should be 0700\n", s.cstr (),
	  int (sb.st_mode & 0777));
    return false;
  }
  return true;
}

void
agent_mkdir (const str &dir)
{
  if (!dir)
    return;
  struct stat sb;
  if (stat (dir, &sb) < 0) {
    if (errno != ENOENT)
      fatal ("%s: %m\n", dir.cstr ());
    warn << "creating directory " << dir << "\n";
    if (mkdir (dir, 0700) < 0 || stat (dir, &sb) < 0)
      fatal ("%s: %m\n", dir.cstr ());
  }
  if (!S_ISDIR (sb.st_mode))
    fatal << dir << ": not a directory\n";
  if (sb.st_mode & 077)
    fatal ("%s: mode 0%o should be 0700\n", dir.cstr (),
	   int (sb.st_mode & 0777));
}


void
agent_mkdir ()
{
  if (!agentsock)
    agent_setsock ();
  agent_mkdir (dotsfs);
  agent_mkdir (userkeysdir);
}

str
defkey ()
{
  agent_ckdir ();
  return dotsfs << "/identity";
}

void
rndaskcd ()
{
  static bool done;
  if (done)
    return;
  done = true;

  sfsagent_seed seed;
  ref<agentconn> aconn = New refcounted<agentconn> ();
  ptr<aclnt> c = aconn->ccd (false);
  if (!c || c->scall (AGENT_RNDSEED, NULL, &seed))
    warn ("sfscd not running, limiting sources of entropy\n");
  else
    rnd_input.update (seed.base (), seed.size ());
}

bool
isagentrunning ()
{
  ref <agentconn> aconn = New refcounted<agentconn> ();
  return aconn->isagentrunning ();
}

void
agent_spawn (bool opt_verbose)
{
  // start agent if needed (sfsagent -c)
  if (!isagentrunning ()) {
    if (opt_verbose)
      warn << "No existing agent found; spawning a new one...\n";
    str sa = find_program ("sfsagent");
    vec<const char *> av;

    av.push_back (sa.cstr ());
    av.push_back ("-c");
    av.push_back (NULL);
    pid_t pid = spawn (av[0], av.base ());
    if (waitpid (pid, NULL, 0) < 0)
      fatal << "Could not spawn a new SFS agent: " <<
	strerror (errno) << "\n";
  }
  else {
    if (opt_verbose)
      warn << "Contacting existing agent...\n";
  }
}
