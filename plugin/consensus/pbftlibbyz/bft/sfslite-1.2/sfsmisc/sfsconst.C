/* $Id: sfsconst.C 3773 2008-11-13 20:50:37Z max $ */

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

#include "sfsmisc.h"
#include "amisc.h"
#include "parseopt.h"
#include <pwd.h>
#include <grp.h>

#ifndef SFSUSER
# define SFSUSER "sfs"
#endif /* !SFSUSER */

u_int32_t sfs_release = SFS_RELEASE;
u_int16_t sfs_defport;
uid_t sfs_uid;
gid_t sfs_gid;
uid_t nobody_uid = (uid_t) -1;
gid_t nobody_gid = (gid_t) -1;
u_int32_t sfs_resvgid_start;	// First reserved gid
u_int sfs_resvgid_count;	// Number of reserved gid's
#ifdef MAINTAINER
bool runinplace;
#endif /* MAINTAINER */
str sfsdir = SFSDIR;
str sfssockdir = SFSDIR "/sockets";
str sfsdevdb;
const char *etc1dir = ETCDIR;
const char *etc2dir = DATADIR;
const char *etc3dir = NULL;
const char *sfsroot = "/sfs";
const char *sfs_authd_syslog_priority = "local4.info";
u_int sfs_dlogsize = 1024;
u_int sfs_mindlogsize = 768;
u_int sfs_maxdlogsize = 24576;
u_int sfs_rsasize = 1280;
u_int sfs_minrsasize = 768;
u_int sfs_maxrsasize = 24576;
u_int sfs_pwdcost = 12;
u_int sfs_hashcost = 0;
u_int sfs_maxhashcost = 22;

static bool const_set;

#ifdef MAINTAINER
# define fwarn warn
#else /* !MAINTAINER */
# define fwarn fatal
#endif /* !MAINTAINER */

static void
idlookup (str uid, str gid)
{
  if (!uid)
    uid = "sfs";
  if (!gid)
    gid = uid;

  bool uidok = convertint (uid, &sfs_uid);
  struct passwd *pw = uidok ? getpwuid (sfs_uid) : getpwnam (uid.cstr ());
  bool gidok = convertint (gid, &sfs_gid);
  struct group *gr = gidok ? getgrgid (sfs_gid) : getgrnam (gid.cstr ());

  if (!uidok) {
    if (!pw)
      fatal << "Could not find user " << uid << "\n";
    sfs_uid = pw->pw_uid;
  }
  if (!gidok) {
    if (!gr)
      fatal << "Could not find group " << gid << "\n";
    sfs_gid = gr->gr_gid;
  }
  if (gr && gr->gr_mem[0])
    fwarn << "Group " << gid << " must not have any members\n";
  if (pw && gr && (gid_t) pw->pw_gid != (gid_t) gr->gr_gid)
    fwarn << "User " << uid << " must have login group " << gid << ".\n";

  endpwent ();
  endgrent ();
}

static void
resvgidset (str lowstr, str highstr)
{
  if (!lowstr || !highstr)
    return;

  u_int32_t end;

  if (!convertint (lowstr, &sfs_resvgid_start))
    fatal << "Could not interpret resvgid value " 
	  << lowstr << " as a number.\n";
  if (!convertint (highstr, &end))
    fatal << "Could not interpret resvgid value " 
	  << highstr << " as a number.\n";

  if (sfs_resvgid_start > end)
    fatal << "Starting value of resvgid range is greater than end value.\n";

  sfs_resvgid_count = end - sfs_resvgid_start + 1;
}

static void
got_sfsdir (bool *setp, vec<str> s, str loc, bool *errp)
{
  if (*setp) {
    *errp = true;
    warn << loc << ": duplicate sfsdir directive\n";
  } else if (s.size () != 2) {
    *errp = true;
    warn << loc << ": usage: sfsdir <directory>\n";
  } else if (!runinplace) {
    sfsdir = s[1];
    sfssockdir = sfsdir << "/sockets";
  }
  *setp = true;
}

static bool
parseconfig (const char *dir, const char *file)
{
  str cf;
  if (!file)
    return false;
  if (file[0] == '/')
    cf = file;
  else if (!dir)
    return false;
  else
    cf = strbuf ("%s/%s", dir, file);

  if (access (cf, F_OK) < 0) {
    if (errno != ENOENT)
      warn << cf << ": " << strerror (errno) << "\n";
    return false;
  }

  conftab ct;

  bool dirset = false;
  ct.add ("RSASize", &sfs_rsasize, sfs_minrsasize, sfs_maxrsasize)
    .add ("DLogSize", &sfs_dlogsize, sfs_mindlogsize, sfs_maxdlogsize)
    .add ("PwdCost", &sfs_pwdcost, 0, 32)
    .add ("LogPriority", &syslog_priority)
    .add ("sfsdir", wrap (&got_sfsdir, &dirset));

  str uid, gid, nuid, ngid, resvgidlow, resvgidhigh;

  bool errors = false;
  parseargs pa (cf);
  int line;
  vec<str> av;
  while (pa.getline (&av, &line)) {
    if (!strcasecmp (av[0], "sfsuser")) {
      if (uid) {
	errors = true;
	warn << cf << ":" << line << ": Duplicate sfsuser directive\n";
      }
      else if (av.size () == 2)
	uid = gid = av[1];
      else if (av.size () == 3) {
	uid = av[1];
	gid = av[2];
      }
      else {
	errors = true;
	warn << cf << ":" << line << ": usage: sfsuser user [group]\n";
      }
    }
    else if (!strcasecmp (av[0], "anonuser")) {
      if (nuid) {
	errors = true;
	warn << cf << ":" << line << ": Duplicate anonuser directive\n";
      }
      else if (av.size () == 2)
	nuid = av[1];
      else if (av.size () == 3) {
	nuid = av[1];
	ngid = av[2];
      }
      else {
	errors = true;
	warn << cf << ":" << line << ": usage: anonuser user [group]\n";
	continue;
      }

      gid_t g;
      if (ngid) {
	if (!convertint (ngid, &g)) {
	  if (struct group *gr = getgrnam (ngid))
	    g = gr->gr_gid;
	  else {
	    errors = true;
	    warn << cf << ":" << line << ": no group " << ngid << "\n";
	    ngid = NULL;
	  }
	}
      }

      uid_t u;
      if (!convertint (nuid, &u)) {
	struct passwd *pw = getpwnam (nuid);
	if (!pw) {
	  errors = true;
	  warn << cf << ":" << line << ": no user " << nuid << "\n";
	  continue;
	}
	nobody_uid = pw->pw_uid;
	if (ngid)
	  nobody_gid = g;
	else
	  nobody_gid = pw->pw_gid;
      }
      else if (ngid) {
	nobody_uid = u;
	nobody_gid = g;
      }
      else {
	errors = true;
	warn << cf << ":" << line << ": Must specify gid with numeric uid";
	continue;
      }
    }
    else if (!strcasecmp (av[0], "resvgids")) {
      if (resvgidhigh) {
	errors = true;
	warn << cf << ":" << line << ": Duplicate resvgids directive\n";
      }
      else if (av.size () != 3) {
	errors = true;
	warn << cf << ":" << line << ": usage: resvgids lower upper\n";
      }
      else {
	resvgidlow = av[1];
	resvgidhigh = av[2];
      }
    }
    else if (!ct.match (av, cf, line, &errors)) {
      warn << cf << ":" << line << ": Unknown directive '"
	   << av[0] << "'\n";
    }
  }    

  if (errors)
    warn << "errors in " << cf << "\n";

  if (uid)
    idlookup (uid, gid);
  resvgidset (resvgidlow, resvgidhigh);
  return true;
}

/* If the directory specified by path does not exist, create it with
 * the given mode. If we fail for any reason, terminate with error.  */
void
mksfsdir (str path, mode_t mode, struct stat *sbp, uid_t uid)
{
  assert (path[0] == '/');

  mode_t m = umask (0);
  struct stat sb;
  if (stat (path, &sb) < 0) {
    if (errno != ENOENT || (mkdir (path, mode) < 0 && errno != EEXIST))
      fatal ("%s: %m\n", path.cstr ());
    if (chown (path, uid, sfs_gid) < 0) {
      int saved_errno = errno;
      rmdir (path);
      fatal ("chown (%s): %s\n", path.cstr (), strerror (saved_errno));
    }
    if (stat (path, &sb) < 0)
      fatal ("stat (%s): %m\n", path.cstr ());
  }
  umask (m);

  if (!S_ISDIR (sb.st_mode))
    fatal ("%s: not a directory\n", path.cstr ());
  if (sb.st_uid != uid)
    fwarn << path << ": owned by uid " << sb.st_uid
	  << ", should be uid " << uid << "\n";
  if (sb.st_gid != sfs_gid)
    fwarn << path << ": has gid " << sb.st_gid
	  << ", should be gid " << sfs_gid << "\n";
  if (sb.st_mode & 07777 & ~mode)
    fwarn ("%s: mode 0%o, should be 0%o\n",
	   path.cstr (), int (sb.st_mode & 07777), int (mode));

  if (sbp)
    *sbp = sb;
}

void
sfsconst_init (bool lite_mode)
{
  if (const_set)
    return;
  const_set = true;

  {
    char *p = safegetenv ("SFS_RELEASE");
    if (!p || !convertint (p, &sfs_release)) {
      str rel (strbuf () << "SFS_RELEASE=" << sfs_release);
      xputenv (const_cast<char*>(rel.cstr()));
    }
  }

#ifdef MAINTAINER
  if (char *p = safegetenv ("SFS_RUNINPLACE")) {
    runinplace = true;
    builddir = p;
    buildtmpdir = builddir << "/runinplace";
  }
  if (char *p = safegetenv ("SFS_ROOT"))
    if (*p == '/')
      sfsroot = p;
#endif /* MAINTAINER */
  sfsdevdb = strbuf ("%s/.devdb", sfsroot);

#ifdef MAINTAINER
  if (runinplace) {
    sfsdir = buildtmpdir;
    sfssockdir = sfsdir;
    etc3dir = etc1dir;
    etc1dir = sfsdir;
    etc2dir = xstrdup (str (builddir << "/etc"));
  }
#endif /* MAINTAINER */
  if (char *ps = safegetenv ("SFS_PORT"))
    if (int pv = atoi (ps))
      sfs_defport = pv;

  str sfs_config = safegetenv ("SFS_CONFIG");
  if (sfs_config && sfs_config[0] == '/') {
    if (!parseconfig (NULL, sfs_config))
      fatal << sfs_config << ": " << strerror (errno) << "\n";
  }
  else {
    if (!parseconfig (etc3dir, sfs_config)) {
      parseconfig (etc3dir, "sfs_config");
      if (!parseconfig (etc2dir, sfs_config)) {
	parseconfig (etc2dir, "sfs_config");
	if (!parseconfig (etc1dir, sfs_config)) 
	  parseconfig (etc1dir, "sfs_config");
      }
    }
  }

  if (!lite_mode) {
    if (!sfs_uid)
      idlookup (NULL, NULL);
  }

  if (char *p = getenv ("SFS_HASHCOST")) {
    sfs_hashcost = strtoi64 (p);
    if (sfs_hashcost > sfs_maxhashcost)
      sfs_hashcost = sfs_maxhashcost;
  }

  if (!getuid () && !runinplace) {
    mksfsdir (sfsdir, 0755);
    mksfsdir (sfssockdir, 0750);
  }
  else if (runinplace && access (sfsdir, 0) < 0) {
    struct stat sb;
    if (!stat (builddir, &sb)) {
      mode_t m = umask (0);
      if (!getuid ()) {
	if (pid_t pid = fork ())
	  waitpid (pid, NULL, 0);
	else {
	  umask (0);
	  setgid (sfs_gid);
	  setuid (sb.st_uid);
	  if (mkdir (sfsdir, 02770) >= 0)
	    rc_ignore (chown (sfsdir, (uid_t) -1, sfs_gid));
	  _exit (0);
	}
      }
      else
	mkdir (sfsdir, 0777);
      umask (m);
    }
  }
}

str
sfsconst_etcfile (const char *name)
{
  const char *path[] = { etc1dir, etc2dir, etc3dir, NULL };
  return sfsconst_etcfile (name, path);
}

str
sfsconst_etcfile_required (const char *name)
{
  const char *path[] = { etc1dir, etc2dir, etc3dir, NULL };
  return sfsconst_etcfile_required (name, path);
}

str
sfsconst_etcfile (const char *name, const char *const *path)
{
  str file;
  if (name[0] == '/') {
    file = name;
    if (!access (file, F_OK))
      return file;
    return NULL;
  }
  for (const char *const *d = path; *d; d++) {
    file = strbuf ("%s/%s", *d, name);
    if (!access (file, F_OK))
      return file;
    if (errno != ENOENT)
      fatal << file << ": " << strerror (errno) << "\n";
  }
  return NULL;
}

str
sfsconst_etcfile_required (const char *name, const char *const *path, bool ftl)
{
  str file = sfsconst_etcfile (name, path);
  if (!file) {
    strbuf msg ("Could not find '%s'. Searched:\n", name);
    for (const char *const *d = path; *d; d++) {
      msg << "  " << *d << "/" << name << "\n";
    }
    str m = msg;
    if (ftl) fatal ("%s", m.cstr ());
    else warn ("%s", m.cstr ());
  }
  return file;
}

str
sfshostname ()
{
  str name = safegetenv ("SFS_HOSTNAME");
  if (!name)
    name = myname ();
  if (!name)
    fatal ("could not figure out host's fully-qualified domain name\n");
  mstr m (name.len ());
  for (u_int i = 0; i < name.len (); i++)
    m[i] = tolower (name[i]);
  return m;
}
