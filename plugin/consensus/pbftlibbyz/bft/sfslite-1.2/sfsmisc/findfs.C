/* $Id: findfs.C 3769 2008-11-13 20:21:34Z max $ */

/*
 *
 * Copyright (C) 2001 David Mazieres (dm@uun.org)
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
#include "getfh3.h"
#include "rxx.h"
#include "parseopt.h"

const u_int64_t nodev (static_cast<u_int64_t> (-1));
#define NFSDB ".nfsdb"
static rxx nfsdbrx ("^(0x[0-9a-f]+)\\s+([\\w\\.:\\-]+)\\s+([\\w\\-]+)"
		    "\\s+([0-9a-f]+)\\s+(\\w+):(\\d+)");

static void
nfsdbcat (int out)
{
  str nfsdbpath = strbuf ("%s/%s", sfsroot, NFSDB);
  str nfsdb;
  do {
    nfsdb = file2str (nfsdbpath);
  } while (!nfsdb && errno == ESTALE);
  if (nfsdb)
    v_write (out, nfsdb, nfsdb.len ());
}

void
nfsdbfetch (cbs cb)
{
  chldrun (wrap (nfsdbcat), cb);
}

template<size_t max> inline bool
a2bytes (rpc_bytes<max> &b, str a)
{
  int i = 0;
  if (a.len () & 1)
    return false;
  b.setsize (a.len () / 2);
  for (const char *p = a.cstr (), *e = p + a.len (); p < e; p += 2) {
    int h, l;
    if (!hexconv (h, p[0]) || !hexconv (l, p[1]))
      return false;
    b[i++] = h << 4 | l;
  }
  return true;
}

static void
getsfsnfs_2 (str sspath, str path, ref<aclnt> c,
	     callback<void, ptr<aclnt>, const nfs_fh3 *>::ref cb,
	     const nfs_fh3 *fhp, const fattr3exp *, str err)
{
  if (err) {
    warn << sspath << "/" << path << ": " << err << "\n";
    (*cb) (NULL, NULL);
  }
  else
    (*cb) (c, fhp);
}

static void
getsfsnfs_1 (u_int64_t dev, str path, const authunix_parms *aup,
	     callback<void, ptr<aclnt>, const nfs_fh3 *>::ref cb,
	     str nfsdb)
{
  sockaddr_in sin;
  nfs_fh3 fh;
  u_int64_t d2;
  strbuf sb;

  if (!nfsdb) {
    (*cb) (NULL, NULL);
    return;
  }

  bzero (&sin, sizeof (&sin));
  sb << nfsdb;

  while (str line = suio_getline (sb.tosuio ()))
    if (nfsdbrx.match (line) && convertint (nfsdbrx[1], &d2) && d2 == dev
	&& a2bytes (fh.data, nfsdbrx[4])
	&& nfsdbrx[5] == "UDP" && convertint (nfsdbrx[6], &sin.sin_port)) {
      sin.sin_family = AF_INET;
      sin.sin_port = htons (sin.sin_port);
      sin.sin_addr.s_addr = htonl (INADDR_LOOPBACK);
      ref<aclnt> c = aclnt::alloc (udpxprt(), nfs_program_3, (sockaddr *) &sin);
      lookupfh3 (c, fh, path, wrap (getsfsnfs_2, nfsdbrx[2], path, c, cb),
		 aup);
      return;
    }
  (*cb) (NULL, NULL);
}

static void
getsfsnfs (u_int64_t dev, str path, const authunix_parms *aup,
	   callback<void, ptr<aclnt>, const nfs_fh3 *>::ref cb)
{
  nfsdbfetch (wrap (getsfsnfs_1, dev, path, aup, cb));
}

static void
aupsetgroups (const authunix_parms *aup)
{
  int ng = min<u_int> (NGROUPS_MAX, aup->aup_len);
  GETGROUPS_T gids[NGROUPS_MAX];
  for (int i = 0; i < ng; i++)
    gids[i] = aup->aup_gids[i];
  if (setgroups (ng, (gid_t*)gids) < 0) {
    warn ("setgroups: %m\n");
    _exit (1);
  }
  if (setgid (aup->aup_gid) < 0) {
    warn ("setgid: %m\n");
    _exit (1);
  }
}

static void
pathinfofetch_cb (callback<void, u_int64_t, str, str>::ref cb, str s)
{
  static rxx pirx ("^(0x[0-9a-f]+)\n(.*)\n([\\x00-\\xff]*)\n$");
  u_int64_t dev = 0;
  if (!s || !pirx.search (s) || !convertint (pirx[1], &dev))
    (*cb) (nodev, NULL, NULL);
  else
    (*cb) (dev, pirx[2], pirx[3]);
}

void
pathinfofetch (const authunix_parms *aup, str path,
	       callback<void, u_int64_t, str, str>::ref cb, int *fdp)
{
  static str pathinfo;
  if (!pathinfo)
    pathinfo = fix_exec_path ("pathinfo");
  int fds[2];
  if (socketpair (AF_UNIX, SOCK_STREAM, 0, fds) < 0) {
    warn ("socketpair: %m\n");
    (*cb) (nodev, NULL, NULL);
    return;
  }
  close_on_exec (fds[0]);

  str uid (strbuf () << (aup ? aup->aup_uid : getuid ()));
  const char *av[] = {
    pathinfo.cstr (), "-u", uid.cstr (), path.cstr (), NULL
  };
  if (aup)
    aspawn (pathinfo, const_cast<char **> (av), 0, fds[1], 2,
	   wrap (&aupsetgroups, aup));
  else
    aspawn (pathinfo, const_cast<char **> (av), 0, fds[1]);
  close (fds[1]);
  if (fdp)
    *fdp = -1;
  pipe2str (fds[0], wrap (pathinfofetch_cb, cb), fdp);
}

class fsfetch {
  const int flags;
  const authunix_parms *aup;
  str path;
  findfscb_t cb;
  str rpath;
  str devname;
  str hostname;
  sockaddr_in hostaddr;
  nfs_fh3 fh;
  ptr<aclnt> nfsc;
  int pathfd;
  u_int64_t rdev;

  void fail (str err) { close (pathfd); (*cb) (NULL, err); delete this; }
  void finish (const nfs_fh3 *fhp);
  void gotpi (u_int64_t dev, str devname, str rp);
  void gotsfsi (ptr<aclnt> c, const nfs_fh3 *fhp);
  void gotnfsi (const nfs_fh3 *fhp, str);
  void gotnfsc (ptr<aclnt> c, clnt_stat stat);
  void lookupcb (const nfs_fh3 *fhp, const fattr3exp *, str err);

public:
  fsfetch (const authunix_parms *aup, str path, findfscb_t cb, int flags);
};

fsfetch::fsfetch (const authunix_parms *aup, str path, findfscb_t cb, int fl)
  : flags (fl), aup (aup), path (path), cb (cb), pathfd (-1)
{ 
  pathinfofetch (aup, path, wrap (this, &fsfetch::gotpi), &pathfd);
}

void
fsfetch::finish (const nfs_fh3 *fhp)
{
  ref<nfsinfo> info = New refcounted<nfsinfo> (nfsc, hostname, *fhp, rdev);
  info->fd = pathfd;
  (*cb) (info, NULL);
  delete this;
}

void
fsfetch::gotpi (u_int64_t dev, str dn, str rp)
{
  if (!dn || !rp) {
    fail (path << ": no information on path");
    return;
  }
  rpath = rp;
  devname = dn;
  rdev = dev;
  if (flags & FINDFS_NOSFS)
    gotsfsi (NULL, NULL);
  else
    getsfsnfs (rdev, rpath, aup, wrap (this, &fsfetch::gotsfsi));
}

void
fsfetch::gotsfsi (ptr<aclnt> c, const nfs_fh3 *fhp)
{
  if (c && fhp) {
    nfsc = c;
    finish (fhp);
    return;
  }
  static rxx nfspath ("^([a-zA-Z0-9.\\-]+):(/.*)$");
  if (nfspath.match (devname)) {
    hostname = nfspath[1];
    getfh3 (hostname, nfspath[2], wrap (this, &fsfetch::gotnfsi));
  }
  else if (!(flags & FINDFS_NOLOCAL)) {
    hostname = "127.0.0.1";
    devname = hostname << ":" << path;
    rpath = "";
    getfh3 (hostname, path, wrap (this, &fsfetch::gotnfsi));
  }
  else {
    fail (path << ": not NFS or SFS file system");
    return;
  }
}

void
fsfetch::gotnfsi (const nfs_fh3 *fhp, str err)
{
  if (err && !(flags & FINDFS_NOLOCAL) && hostname == "127.0.0.1") {
    vec<in_addr> addrs;
    myipaddrs (&addrs);
    while (!addrs.empty () && ntohl (addrs[0].s_addr) == INADDR_LOOPBACK)
      addrs.pop_front ();
    if (!addrs.empty ()) {
      hostname = inet_ntoa (addrs[0]);
      // devname = hostname << ":" << path;
      getfh3 (hostname, path, wrap (this, &fsfetch::gotnfsi));
      return;
    }
  }
  if (err) {
    fail (devname << ": NFS mount: " << err);
    return;
  }
  fh = *fhp;
  aclntudp_create (hostname, 0, nfs_program_3,
		   wrap (this, &fsfetch::gotnfsc));
}

void
fsfetch::gotnfsc (ptr<aclnt> c, clnt_stat stat)
{
  if (stat) {
    fail (hostname << ": NFS server: " << stat);
    return;
  }
  nfsc = c;

  if (rpath == "/" || rpath == "")
    lookupcb (&fh, NULL, NULL);
  else
    lookupfh3 (nfsc, fh, rpath, wrap (this, &fsfetch::lookupcb), aup);
}

void
fsfetch::lookupcb (const nfs_fh3 *fhp, const fattr3exp *, str err)
{
  if (fhp)
    finish (fhp);
  else
    fail (devname << rpath << ": " << err);
}

void
findfs (const authunix_parms *aup, str path, findfscb_t cb, int flags)
{
  vNew fsfetch (aup, path, cb, flags);
}
