/* $Id: getsysnoise.C 3773 2008-11-13 20:50:37Z max $ */

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

#include "async.h"
#include "prng.h"
#include "sha1.h"
#include <sys/resource.h>

const char *const noiseprogs[][5] = {
  { PATH_PS, "laxwww" },
  { PATH_PS, "-al" },
  { PATH_LS, "-nfail", "/tmp/." },
#ifdef PATH_NETSTAT
  { PATH_NETSTAT, "-s" },
  { PATH_NETSTAT, "-an" },
  { PATH_NETSTAT, "-in" },
#endif /* PATH_NETSTAT */
#ifdef PATH_NTPQ
  { PATH_NTPQ, "-np" },
#endif /* PATH_NTPQ */
#ifdef PATH_W
  { PATH_W },
#endif /* PATH_W */
#ifdef PATH_NFSSTAT
  { PATH_NFSSTAT },
#endif /* PATH_NFSSTAT */
#ifdef PATH_VNSTAT
  { PATH_VMSTAT },
  { PATH_VMSTAT, "-i" },
  { PATH_VMSTAT, "-s" },
#endif /* PATH_VNSTAT */
#ifdef PATH_IOSTAT
#if defined (__linux__) || defined (__osf__)
  { PATH_IOSTAT },
#else /* not linux or osf */
  { PATH_IOSTAT, "-I" },
#endif /* not linux or osf */
#endif /* PATH_IOSTAT */
#ifdef PATH_LSOF
  { PATH_LSOF, "-bwn",
# ifdef LSOF_DEVCACHE
    "-Di"
# endif /* LSOF_DEVCACHE */
  },
#else /* no lsof */
# ifdef PATH_FSTAT
  { PATH_FSTAT },
# endif /* PATH_FSTAT */
# ifdef PATH_PSTAT
  { PATH_PSTAT, "-f" },
# endif /* PATH_PSTAT */
#endif /* no lsof */
#ifdef PATH_PSTAT
  { PATH_PSTAT, "-t" },
# if defined (__OpenBSD__) || defined (__NetBSD__) || defined (__FreeBSD__)
  { PATH_PSTAT, "-v" },
# endif /* open/net/freebsd */
#endif /* PATH_PSTAT */
#ifdef PATH_NFSSTAT
  { PATH_NFSSTAT },
#endif /* PATH_NFSSTAT */
#if 0
  { PATH_RUP },
  { PATH_RUSERS, "-l" },
#endif
  { NULL }
};

class noise_from_fd {
  friend void getfdnoise (datasink *, int, cbv, size_t);

  datasink *const dst;
  const int fd;
  size_t bytes;
  cbv cb;

  noise_from_fd (datasink *dst, int fd, cbv cb, size_t maxbytes);
  ~noise_from_fd ();
  void doread ();
};

class noise_from_prog {
  pid_t pid;
  timecb_t *to;
  cbv cb;

  PRIVDEST ~noise_from_prog () { if (to) timecb_remove (to); }
  int execprog (const char *const *);
  void done ();
  void timeout ();
public:
  noise_from_prog (datasink *dst, int fd, pid_t pid, cbv cb);
  noise_from_prog (datasink *dst, const char *const *av, cbv cb);
};

void
getfdnoise (datasink *dst, int fd, cbv cb, size_t maxbytes)
{
  vNew noise_from_fd (dst, fd, cb, maxbytes);
}

void
getprognoise (datasink *dst, const char *const *av, cbv cb)
{
  vNew noise_from_prog (dst, av, cb);
}

void
getprognoise (datasink *dst, int fd, pid_t pid, cbv cb)
{
  vNew noise_from_prog (dst, fd, pid, cb);
}

void
getfilenoise (datasink *dst, const char *path, cbv cb, size_t maxbytes)
{
  int fds[2];
  if (pipe (fds) < 0)
    fatal ("pipe: %m\n");
  pid_t pid = afork ();

  if (pid == -1)
    (*cb) ();
  else if (!pid) {
    close (fds[0]);
    int fd = open (path, O_RDONLY|O_NONBLOCK, 0);
    if (fd < 0) {
      fatal ("%s: %m\n", path);
      _exit (1);
      return;
    }
    for (;;) {
      char buf[1024];
      size_t n = read (fd, buf, min (maxbytes, sizeof (buf)));
      if (n <= 0)
	_exit (0);
      v_write (fds[1], buf, n);
      maxbytes -= n;
      if (!maxbytes)
	_exit (0);
    }
  }
  else {
    close (fds[1]);
    close_on_exec (fds[0]);
    getprognoise (dst, fds[0], pid, cb);
  }
}

void
getclocknoise (datasink *dst)
{
  timespec ts; 
  clock_gettime (CLOCK_REALTIME, &ts);
  dst->update (&ts, sizeof (ts));

#if 0
  clock_gettime (CLOCK_VIRTUAL, &ts);
  dst->update (&ts, sizeof (ts));
  clock_gettime (CLOCK_PROF, &ts);
  dst->update (&ts, sizeof (ts));
#endif

  rusage ru;
  getrusage (RUSAGE_CHILDREN, &ru);
  dst->update (&ru, sizeof (ru));
  getrusage (RUSAGE_SELF, &ru);
  dst->update (&ru, sizeof (ru));
}

noise_from_fd::noise_from_fd (datasink *dst, int fd, cbv cb, size_t maxbytes)
  : dst (dst), fd (fd), bytes (maxbytes), cb (cb)
{
  make_async (fd);
  fdcb (fd, selread, wrap (this, &noise_from_fd::doread));
}

noise_from_fd::~noise_from_fd ()
{
  fdcb (fd, selread, NULL);
  close (fd);
  (*cb) ();
}

void
noise_from_fd::doread ()
{
  char buf[8192];
  ssize_t n = read (fd, buf, min (sizeof (buf), bytes));

  getclocknoise (dst);
  if (n > 0) {
    dst->update (buf, n);
    bytes -= n;
    if (!bytes)
      delete this;
  }
  else if (!n || (errno != EAGAIN && errno != EINTR)) {
    if (n < 0)
      warn ("noise_from_fd::doread: %m\n");
    delete this;
  }
}

noise_from_prog::noise_from_prog (datasink *dst, int fd, pid_t p, cbv cb)
  : pid (p), cb (cb)
{
  to = delaycb (30, wrap (this, &noise_from_prog::timeout));
  getfdnoise (dst, fd, wrap (this, &noise_from_prog::done));
}

noise_from_prog::noise_from_prog (datasink *dst, const char *const *av, cbv cb)
  : cb (cb)
{
  int fd = execprog (av);
  to = delaycb (30, wrap (this, &noise_from_prog::timeout));
  getfdnoise (dst, fd, wrap (this, &noise_from_prog::done));
}

int
noise_from_prog::execprog (const char *const *av)
{
  int fds[2];
  if (pipe (fds) < 0)
    fatal ("pipe: %m\n");
  pid = afork ();

  if (!pid) {
    close (fds[0]);
    if (fds[1] != 1)
      dup2 (fds[1], 1);
    if (fds[1] != 2)
      dup2 (fds[1], 2);
    if (fds[1] != 1 && fds[1] != 2)
      close (fds[1]);
    close (0);
    rc_ignore (chdir ("/"));
    open ("/dev/null", O_RDONLY);

    char *env[] = { NULL };
    execve (av[0], const_cast<char *const *> (av), env);
    //warn ("%s: %m\n", av[0]);
    _exit (1);
  }

  close (fds[1]);
  close_on_exec (fds[0]);
  return fds[0];
}

void
noise_from_prog::done ()
{
  (*cb) ();
  delete this;
}

void
noise_from_prog::timeout ()
{
  to = NULL;
  kill (pid, SIGKILL);
}

class noise_getter {
  friend void getsysnoise (datasink *, cbv);

  datasink *dst;
  cbv cb;
  u_int numsources;

  noise_getter (datasink *dst, cbv cb);
  ~noise_getter ();
  void sourcedone ();
};

noise_getter::noise_getter (datasink *dst, cbv cb)
  : dst (dst), cb (cb), numsources (1)
{
  pid_t pid = getpid ();
  dst->update (&pid, sizeof (pid));
  for (int i = 0; *noiseprogs[i]; i++) {
    numsources++;
    getprognoise (dst, noiseprogs[i], wrap (this, &noise_getter::sourcedone));
  }
#ifdef SFS_DEV_RANDOM
  numsources++;
  getfilenoise (dst, SFS_DEV_RANDOM,
		wrap (this, &noise_getter::sourcedone), 16);
#endif /* SFS_DEV_RANDOM */
  sourcedone ();
}

noise_getter::~noise_getter ()
{
  getclocknoise (dst);
  (*cb) ();
}

void
noise_getter::sourcedone ()
{
  if (!--numsources)
    delete this;
}

void
getsysnoise (datasink *dst, cbv cb)
{
  vNew noise_getter (dst, cb);
}
