/* $Id: unixserv.C 3773 2008-11-13 20:50:37Z max $ */

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
#include "vec.h"
#include "arpc.h"
#include "sfsagent.h"

static vec<str> sockets;
EXITFN (unixserv_cleanup);

static void
accept_cb (int lfd, cbi cb)
{
  struct sockaddr_un sun;
  socklen_t sunlen = sizeof (sun);
  bzero (&sun, sizeof (sun));
  int fd = accept (lfd, (sockaddr *) &sun, &sunlen);
  if (fd >= 0 || errno != EAGAIN)
    (*cb) (fd);
}

void
sfs_unixserv (str sock, cbi cb, mode_t mode)
{
  /* We set the permissions of the socket because on some OS's this
   * gives an extra measure of protection.  However, the real
   * protection comes from the fact that sfssockdir is not world
   * readable or executable. */
  mode_t m = runinplace ? umask (0) : umask (~(mode & 0777));

  int lfd = unixsocket (sock);
  if (lfd < 0 && errno == EADDRINUSE) {
    /* XXX - This is a slightly race-prone way of cleaning up after a
     * server bails without unlinking the socket.  If we can't connect
     * to the socket, it's dead and should be unlinked and rebound.
     * Two daemons's could do this simultaneously, however. */
    int fd = unixsocket_connect (sock);
    if (fd < 0) {
      unlink (sock);
      lfd = unixsocket (sock);
    }
    else {
      close (fd);
      errno = EADDRINUSE;
    }
  }
  if (lfd < 0)
    fatal ("%s: %m\n", sock.cstr ());

  sockets.push_back (sock);

  umask (m);

  listen (lfd, 5);
  make_async (lfd);
  fdcb (lfd, selread, wrap (accept_cb, lfd, cb));

#if 0  /* A nice idea that unfortunately doesn't work... */
  struct stat sb1, sb2;
  if (stat (sock, &sb1) < 0 || fstat (lfd, &sb2) < 0)
    fatal ("%s: %m\n", sock);
  if (sb1.st_ino != sb2.st_ino || sb1.st_dev != sb2.st_dev)
    fatal ("%s: changed while binding\n", sock);
#endif
}

struct suidaccept {
#ifdef HAVE_GETPEEREID
  uid_t uid;
  gid_t gid;
#endif /* HAVE_GETPEEREID */
  suidservcb::ptr cb;
  ptr<axprt_unix> ax;
  ptr<asrv> as;

  void dispatch (svccb *);
};

/* We have no idea what type aup.aup_gids is (could be gid_t,
 * u_int32_t, etc.)  Rather than autoconf it, just use templates to
 * work around the problem. */
template<class T> inline void
domalloc (T *&tp, size_t glen)
{
  if (glen)
    tp = static_cast<T *> (xmalloc (glen));
  else
    tp = NULL;
}
void
suidaccept::dispatch (svccb *sbp)
{
  if (sbp) {
    const authunix_parms *aup = sbp->getaup ();
    if (!aup) {
      warn ("suidserv: ignoring SETUID message with no auth\n");
      sbp->reject (AUTH_REJECTEDCRED);
    }
#ifdef HAVE_GETPEEREID
    else if (uid && (implicit_cast<u_int32_t> (uid)
		     != implicit_cast<u_int32_t> (aup->aup_uid))) {
      warn ("suidserv: rejected connection from UID %u claiming to be %u\n",
	    uid, aup->aup_uid);
      sbp->reject (AUTH_REJECTEDVERF);
    }
#endif /* HAVE_GETPEEREID */
    else {
      authunix_parms au = *aup;
      if (aup->aup_machname)
	au.aup_machname = xstrdup (aup->aup_machname);
      domalloc (au.aup_gids, aup->aup_len * sizeof (aup->aup_gids[0]));
      memcpy (au.aup_gids, aup->aup_gids,
	      aup->aup_len * sizeof (aup->aup_gids[0]));
      sbp->replyref (0);
      as->setcb (NULL);
      as = NULL;
      (*cb) (ax, &au);
      ax = NULL;
      xfree (au.aup_machname);
      xfree (au.aup_gids);
      delete this;
      return;
    }
  }
  as->setcb (NULL);
  delete this;
}

static void
suidaccept_cb (suidservcb cb, int fd)
{
  if (fd < 0) {
    (*cb) (NULL, NULL);
    return;
  }

  suidaccept *sa = New suidaccept;
#ifdef HAVE_GETPEEREID
  if (getpeereid (fd, &sa->uid, &sa->gid) < 0) {
    warn ("getpeereid: %m\n");
    close (fd);
    delete sa;
    return;
  }
#endif /* HAVE_GETPEEREID */
  sa->cb = cb;
  sa->ax = axprt_unix::alloc (fd);
  assert (!sa->ax->ateof ());
  sa->ax->allow_recvfd = false;
  sa->as = asrv::alloc (sa->ax, setuid_prog_1,
			wrap (sa, &suidaccept::dispatch));
}

void
sfs_suidserv (str prog, suidservcb cb)
{
  str sock = strbuf ("%s/%s.sock", sfssockdir.cstr (), prog.cstr ());
  sfs_unixserv (sock, wrap (suidaccept_cb, cb), 0660);
  struct stat sb;
  if (lstat (sock.cstr (), &sb) < 0 || (sb.st_mode & S_IFMT) != S_IFSOCK)
    fatal << sock << ": bound but cannot stat\n";
  if (!runinplace && sb.st_gid != sfs_gid)
    rc_ignore (chown (sock.cstr (), (uid_t) -1, sfs_gid));
}

static void
unixserv_cleanup ()
{
  while (!sockets.empty ())
    unlink (sockets.pop_front ());
}
