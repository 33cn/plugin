/* $Id: axprt_unix.C 2531 2007-02-11 14:40:18Z max $ */

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

#include "arpc.h"
#include "vec.h"

#ifdef MAINTAINER
bool axprt_unix_spawn_connected;
#endif /* MAINTAINER */
pid_t axprt_unix_spawn_pid;

void
axprt_unix::clone (ref<axprt_clone> x)
{
  assert (pktsize >= x->pktsize); // sic - allow no possibility of failure
  assert (!x->ateof ());

  sendfd (x->takefd ());
  assert (x->pktlen >= 4);
  send (x->pktbuf + 4, x->pktlen - 4, NULL);
}

axprt_unix::~axprt_unix ()
{
  while (!fdrecvq.empty ())
    close (fdrecvq.pop_front ());
}

void
axprt_unix::recvbreak ()
{
  if (!allow_recvfd) {
    axprt_stream::recvbreak ();
    return;
  }
}

int
axprt_unix::recvfd ()
{
  if (fdrecvq.empty ())
    return -1;
  return fdrecvq.pop_front ();
}

void
axprt_unix::sendfd (int sfd, bool closeit)
{
  fdsendq.push_back (fdtosend (sfd, closeit));
  sendbreak (NULL);
}

ssize_t
axprt_unix::doread (void *buf, size_t maxlen)
{
  if (!allow_recvfd)
    return read (fdread, buf, maxlen);
  int nfd = -1;
  ssize_t n = readfd (fdread, buf, maxlen, &nfd);
  if (nfd >= 0) {
    if (fdrecvq.size () >= 4) {
      close (nfd);
      warn ("axprt_unix: too many unclaimed file descriptors\n");
    }
    else
      fdrecvq.push_back (nfd);
  }
  return n;
}

int
axprt_unix::dowritev (int cnt)
{
  if (fdsendq.empty ())
    return out->output (fdwrite, cnt);

  static timeval ztv;
  if (!fdwait (fdwrite, selwrite, &ztv))
    return 0;

  if (cnt < 0)
    cnt = out->iovcnt ();
  if (cnt > UIO_MAXIOV)
    cnt = UIO_MAXIOV;
  ssize_t n = writevfd (fdwrite, out->iov (), cnt, fdsendq.front ().fd);
  if (n < 0)
    return errno == EAGAIN ? 0 : -1;
  fdsendq.pop_front ();
  out->rembytes (n);
  return 1;
}

ref<axprt_unix>
axprt_unix::alloc (int f, size_t ps)
{
  ref<axprt_unix> x = New refcounted<axprt_unix> (f, ps);

  if (!isunixsocket (f)) {
    warn ("axprt_unix::alloc(%d): not unix domain socket\n", f);
    x->fail ();
  }
  return x;
}

#ifdef MAINTAINER
static ptr<axprt_unix>
tryconnect (str path, const char *arg0, u_int ps)
{
  const char *prog = strrchr (path, '/');
  if (!prog)
    panic ("tryconnect: path '%s' has no '/'\n", path.cstr ());
  prog++;

  if (builddir) {
    const char *a0;
    if (arg0) {
      if ((a0 = strrchr (arg0, '/')))
	a0++;
      else
	a0 = arg0;
    }
    else
      a0 = prog;

    str np = strbuf ("%s/.%s",
		     buildtmpdir ? buildtmpdir.cstr () : builddir.cstr (),
		     a0);
    return axprt_unix_connect (np, ps);
  }
  return NULL;
}
#endif /* MAINTAINER */

static ptr<axprt_unix>
axprt_unix_dospawnv (str path, const vec<str> &avs,
		     size_t ps, cbv::ptr postforkcb, bool async,
		     char *const *env)
{
  axprt_unix_spawn_pid = -1;
  vec<const char *> av;
  if (!ps)
    ps = axprt_stream::defps;

  // path = fix_exec_path (path);
#ifdef MAINTAINER
  if (ptr<axprt_unix> x = tryconnect (path, avs[0], ps)) {
    axprt_unix_spawn_connected = true;
    return x;
  }
  axprt_unix_spawn_connected = false;
#endif /* MAINTAINER */

  for (const str *s = avs.base (), *e = avs.lim (); s < e; s++)
    av.push_back (s->cstr ());
  av.push_back (NULL);

  int fds[2];
  if (socketpair (AF_UNIX, SOCK_STREAM, 0, fds) < 0) {
    warn ("socketpair: %m\n");
    return NULL;
  }
  close_on_exec (fds[0]);
  pid_t pid;
  if (async)
    pid = aspawn (path, av.base (), fds[1], fds[1], 2, postforkcb, env);
  else
    pid = spawn (path, av.base (), fds[1], fds[1], 2, postforkcb, env);
  axprt_unix_spawn_pid = pid;
  close (fds[1]);
  if (pid < 0) {
    close (fds[0]);
    return NULL;
  }

  // return axprt_unix::alloc (fds[0], ps);
  // XXX - extra complexity to work around g++ bug
  ref<axprt_unix> x = axprt_unix::alloc (fds[0], ps);
  return x;
}

ptr<axprt_unix>
axprt_unix_spawnv (str path, const vec<str> &avs,
		   size_t ps, cbv::ptr postforkcb, char *const *env)
{
  ptr<axprt_unix> x = axprt_unix_dospawnv (path, avs, ps, postforkcb, false,
					   env);
  return x;
}
ptr<axprt_unix>
axprt_unix_aspawnv (str path, const vec<str> &avs,
		    size_t ps, cbv::ptr postforkcb, char *const *env)
{
  ptr<axprt_unix> x = axprt_unix_dospawnv (path, avs, ps, postforkcb, true,
					   env);
  return x;
}

ptr<axprt_unix>
axprt_unix_spawn (str path, size_t ps, char *arg0, ...)
{
  axprt_unix_spawn_pid = -1;
  vec<char *> av;
  if (!ps)
    ps = axprt_stream::defps;

  if (arg0) {
    va_list ap;
    va_start (ap, arg0);
    av.push_back (arg0);
    while (av.push_back (va_arg (ap, char *)))
      ;
    va_end (ap);
  }
  else {
    av.push_back (const_cast<char *> (path.cstr ()));
    av.push_back (NULL);
  }

  // path = fix_exec_path (path);
#ifdef MAINTAINER
  if (ptr<axprt_unix> x = tryconnect (path, arg0, ps)) {
    axprt_unix_spawn_connected = true;
    return x;
  }
  axprt_unix_spawn_connected = false;
#endif /* MAINTAINER */

  int fds[2];
  if (socketpair (AF_UNIX, SOCK_STREAM, 0, fds) < 0) {
    warn ("socketpair: %m\n");
    return NULL;
  }
  close_on_exec (fds[0]);
  pid_t pid = spawn (path, av.base (), fds[1]);
  axprt_unix_spawn_pid = pid;
  close (fds[1]);
  if (pid < 0) {
    close (fds[0]);
    return NULL;
  }

  // return axprt_unix::alloc (fds[0], ps);
  // XXX - extra complexity to work around g++ bug
  ref<axprt_unix> x = axprt_unix::alloc (fds[0], ps);
  return x;
}

ptr<axprt_unix>
axprt_unix_connect (const char *path, size_t ps)
{
  int fd = unixsocket_connect (path);
  if (fd < 0)
    return NULL;
  return axprt_unix::alloc (fd, ps);
}

#ifdef MAINTAINER
ptr<axprt_unix>
axprt_unix_accept (const char *path, size_t ps)
{
  mode_t m = umask (0);
  int fd = unixsocket (path);
  if (fd < 0) {
    warn ("unixsocket: %m\n");
    umask (m);
    return NULL;
  }
  umask (m);

  sockaddr_un sun;
  socklen_t len = sizeof (sun);
  bzero (&sun, sizeof (sun));
  int afd = -1;
  if (!listen (fd, 1))
    afd = accept (fd, (sockaddr *) &sun, &len);
  unlink (path);
  close (fd);

  if (afd < 0) {
    warn ("%s: %m\n", path);
    return NULL;
  }
  return axprt_unix::alloc (afd, ps);
}
#endif /* MAINTAINER */

ptr<axprt_unix>
axprt_unix_stdin (size_t ps)
{
  ptr<axprt_unix> x = axprt_unix::alloc (0);
#ifdef MAINTAINER
  if (x->ateof () && builddir) {
    str np = strbuf ("%s/.%s",
		     buildtmpdir ? buildtmpdir.cstr () : builddir.cstr (),
		     progname.cstr ());
    x = axprt_unix_accept (np, ps);
  }
#endif /* MAINTAINER */
  if (x && !x->ateof ())
    return x;
  warn ("axprt_unix_stdin: %m\n");
  return NULL;
}
