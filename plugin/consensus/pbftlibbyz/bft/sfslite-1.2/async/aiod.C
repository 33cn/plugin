/* $Id: aiod.C 3773 2008-11-13 20:50:37Z max $ */

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

/*
 * Aiod is an asynchronous I/O daemon implementing the protocol in
 * aiod_prot.h.  Aiod communicates to its client through shared
 * memory, a pipe, and a unix domain socket.  It is invoked with three
 * file descriptor numbers as arguments:
 *
 *    aiod shmfd rfd rwfd
 *
 * The shmfd file descriptor corresponds to a file that will be used
 * for shared memory between aiod and its client.  Aiod mmaps this
 * file descriptor on startup.
 *
 * Shmfd is also used for flock or fcntl locking when the client
 * employs multiple aiods for greater concurrency.  Because flock
 * locks are inherited with file descriptors, every aiod must inherit
 * a separately opened shmfd.  The parent is responsible for reopening
 * the shared memory file (with the open system call) before launching
 * each of several aiod's.
 *
 * [The open system call is the only way to create a new descriptor
 * for a file in Unix.  Dup, dup2, F_DUPFD, file descriptor passing
 * over Unix domain sockets and fork only make new copies of existing
 * file descriptors.]
 *
 * Rfd is a pipe from which aiod reads commands.  Each command is a
 * single integer of type aiomsg_t.  The integer corresponds to the
 * offset in the shared memory region at which an aio request
 * (starting with aiod_reqhdr) resides.  Rfd is intended to be shared
 * by multiple aiod's.
 *
 * [Writes of less than PIPE_BUF bytes to pipes are atomic, and POSIX
 * requires a PIPE_BUF of at least 512, much larger than an aiomsg_t.
 * There is no danger of two aiod's each reading part of an aiomsg_t.]
 *
 * Rwfd is one end of a unix domain stream socket.  This socket must
 * not be shared by any other aiod processes.  When an aiod finishes
 * processing a request, it writes the position of the request (the
 * exact same aiomsg_t it initially read) back to rwfd.  Aiod also
 * reads requests from rwfd.  Thus, for instance, a client can
 * broadcast an AIOD_CLOSE request to all aiod's by writing the
 * request's address to the other end of each aiod's rwfd.  (When
 * writing to the other end of rfd, there is no telling which aiod
 * will get the request.)
 *
 * You will notice that this code does not call select to wait for
 * data on rfd and rwfd.  If many aiod's selected on the same rfd, any
 * write to the other end of that pipe would cause all aiod's to wake
 * up.  This would be far more expensive than a simple shared memory
 * asynchronous I/O request should be.  In fact, even if all the
 * aiod's simply blocked reading rfd, internally the kernel would end
 * up waking every aiod's kernel thread.  Thus, we synchronize access
 * to rfd with flock on shmfd (which should be more efficient).  We
 * also set rwfd to generate SIGIO on data arrival, so that we will be
 * woken up from flock to process requests from rwfd as necessary.
 *
 * Note that aiod completely trusts the process it communicates with.
 * While it may sometimes minimally checks on request data structures,
 * this is only for debugging purposes.  The structures are never
 * copied from shared memory, and a bad client could easily tamper
 * with structures after they are checked.
 */

#include "amisc.h"
#include "aiod_prot.h"
#include "parseopt.h"
#include "list.h"
#include "qhash.h"
#include "dirent.h"

class aiosrv;

static sigset_t sigio_mask;
static int sigio_received;
static int32_t shmfd, rfd, rwfd;
static aiosrv *srv;

class fhtab {
  struct fh {
    const int fd;
    const dev_t dev;
    const ino_t ino;
    const char *const path;

    fh (int f, dev_t d, ino_t i, const char *p)
      : fd (f), dev (d), ino (i), path (xstrdup (p)) {}
    ~fh () { ::close (fd); xfree (const_cast<char *> (path)); }
  };

  qhash<int, ref<fh> > tab;

  fh *alloc (aiod_file *af, int *errp, bool create = false, mode_t mode = 0);
public:
  int lookup (aiod_file *af, int *errp);
  int create (aiod_file *af, mode_t mode, int *errp);
  int close (aiod_file *af, int *errp);
};

fhtab::fh *
fhtab::alloc (aiod_file *af, int *errp, bool create, mode_t mode)
{
  int oflags = af->oflags;
  if (!create)
    oflags &= ~(O_CREAT|O_TRUNC|O_EXCL);
  int fd = open (af->path, oflags, mode);
  if (fd < 0) {
    *errp = errno;
    tab.remove (af->handle);
    return NULL;
  }
  struct stat sb;
  if (fstat (fd, &sb) < 0) {
    *errp = errno;
    ::close (fd);
    tab.remove (af->handle);
    return NULL;
  }

  if (create) {
    af->dev = sb.st_dev;
    af->ino = sb.st_ino;
  }
  else if (af->dev != sb.st_dev || af->ino != sb.st_ino) {
    *errp = ESTALE;
    ::close (fd);
    tab.remove (af->handle);
    return NULL;
  }

  ref<fh> h = New refcounted<fh> (fd, sb.st_dev, sb.st_ino, af->path);
  tab.insert (af->handle, h);
  return h;
}

int
fhtab::lookup (aiod_file *af, int *errp)
{
  fh *h = tab[af->handle];
  if (!h) {
    h = alloc (af, errp);
    if (!h)
      return -1;
  }
  else if (af->dev != h->dev || af->ino != h->ino) {
    /* This shouldn't happen, unless someone closes a file descriptor
     * with outstanding requests (in which case the close could get
     * reordered).  However such usage is really a bug in the calling
     * program. */
    warn ("stale handle on already open file\n");
    tab.remove (af->handle);
    h = alloc (af, errp);
    if (!h)
      return -1;
  }
  return h->fd;
}

int
fhtab::create (aiod_file *af, mode_t mode, int *errp)
{
  fh *h = alloc (af, errp, true, mode);
  return h ? h->fd : -1;
}

int
fhtab::close (aiod_file *af, int *errp)
{
  errno = 0;
  tab.remove (af->handle);
  if (errno) {
    *errp = errno;
    return -1;
  }
  return 0;
}

class dhtab {
  struct fh {
    DIR *fd;
    const dev_t dev;
    const ino_t ino;
    const char *const path;

    fh (DIR *f, dev_t d, ino_t i, const char *p)
      : fd (f), dev (d), ino (i), path (xstrdup (p)) {}
    ~fh () { ::closedir (fd); xfree (const_cast<char *> (path)); }
  };

  qhash<int, ref<fh> > tab;

  fh *alloc (aiod_file *af, int *errp);
public:
  DIR *lookup (aiod_file *af, int *errp);
  int create (aiod_file *af, int *errp);
  int close (aiod_file *af, int *errp);
};

dhtab::fh *
dhtab::alloc (aiod_file *af, int *errp)
{
  DIR *fd = opendir (af->path);
  if (fd == NULL) {
    *errp = errno;
    tab.remove (af->handle);
    return NULL;
  }
  
  ref<fh> h = New refcounted<fh> (fd, 0, 0, af->path);
  tab.insert (af->handle, h);
  return h;
}

DIR *
dhtab::lookup (aiod_file *af, int *errp)
{
  fh *h = tab[af->handle];
  if (!h) {
    h = alloc (af, errp);
    if (!h)
      return NULL;
  }
  else if (strcmp(af->path, h->path) != 0) { 
    /* This shouldn't happen, unless someone closes a file descriptor
     * with outstanding requests (in which case the close could get
     * reordered).  However such usage is really a bug in the calling
     * program. */
    warn ("stale handle on already open file\n");
    tab.remove (af->handle);
    h = alloc (af, errp);
    if (!h)
      return NULL;
  }
  return h->fd;
}

int
dhtab::create (aiod_file *af, int *errp)
{
  fh *h = alloc (af, errp);
  return h ? 1 : -1;
}

int
dhtab::close (aiod_file *af, int *errp)
{
  errno = 0;
  tab.remove (af->handle);
  if (errno) {
    *errp = errno;
    return -1;
  }
  return 0;
}

class shmbuf {
  char *const buf;
  const size_t len;

protected:
  shmbuf (int f, void *buf, size_t len)
    : buf (static_cast<char *> (buf)), len (len), fd (f) {}
  ~shmbuf ();

public:
  const int fd;
  template<class T> T *getptr (aiomsg_t pos) {
#ifdef CHECK_BOUNDS
    assert (pos >= 0 && pos + sizeof (T) <= len);
#endif /* CHECK_BOUNDS */
    return reinterpret_cast<T *> (buf + pos);
  }
  aiod_op getop (aiomsg_t pos) {
#ifdef CHECK_BOUNDS
    assert (pos >= 0 && pos + sizeof (aiod_op) <= len);
#endif /* CHECK_BOUNDS */
    return *reinterpret_cast<aiod_op *> (buf + pos);
  }
  void *getbuf (aiod_iobuf *bp) {
#ifdef CHECK_BOUNDS
    assert (bp->buf >= 0 && bp->buf + (size_t) bp->len <= len);
#endif /* CHECK_BOUNDS */
    return buf + bp->buf;
  }
  static ptr<shmbuf> alloc (int fd);
};

template<> inline void *
shmbuf::getptr<void> (aiomsg_t pos)
{
#ifdef CHECK_BOUNDS
    assert (pos >= 0 && pos <= len);
#endif /* CHECK_BOUNDS */
    return buf + pos;
}

shmbuf::~shmbuf ()
{
  if (munmap (buf, len) < 0)
    warn ("munmap: %m\n");
}

ptr<shmbuf>
shmbuf::alloc (int fd)
{
  struct stat sb;
  if (fstat (fd, &sb) < 0) {
    warn ("stat shared mem file: %m\n");
    return NULL;
  }
  void *buf = mmap (NULL, (size_t) sb.st_size, PROT_READ|PROT_WRITE,
		    MAP_FILE|MAP_SHARED, fd, 0);
  if (buf == reinterpret_cast<char *> (MAP_FAILED)) {
    warn ("mmap: %m\n");
    return NULL;
  }
  return New refcounted<shmbuf> (fd, buf, sb.st_size);
}

class aiosrv {
  fhtab fht;
  dhtab dht;
  const int fd;
  const ref<shmbuf> buf;

  void nop (aiomsg_t);
  void pathop (aiomsg_t);
  void fhop (aiomsg_t);
  void fstat (aiomsg_t);
  void dhop (aiomsg_t);
public:
  aiosrv (int f, ref<shmbuf> b) : fd (f), buf (b) { make_sync (fd); }
  void getmsg (aiomsg_t msg);
};

void
aiosrv::pathop (aiomsg_t msg)
{
  static int fd = -1;
  aiod_pathop *rq = buf->Xtmpl getptr<aiod_pathop> (msg);
  errno = 0;
  switch (rq->op) {
  case AIOD_UNLINK:
    unlink (rq->path1 ());
    break;
  case AIOD_LINK:
    rc_ignore (link (rq->path1 (), rq->path2 ()));
    break;
  case AIOD_SYMLINK:
    rc_ignore (symlink (rq->path1 (), rq->path2 ()));
    break;
  case AIOD_RENAME:
    rename (rq->path1 (), rq->path2 ());
    break;
  case AIOD_READLINK:
    rq->bufsize = readlink (rq->path1 (), rq->pathbuf, rq->bufsize);
    break;
  case AIOD_GETCWD:
    // XXX - shouldn't need to chdir... just write our own getcwd-like func.
    if ((fd >= 0 || (fd = open (".", O_RDONLY)) >= 0)
	&& chdir (rq->path1 ()) >= 0) {
      if (getcwd (rq->pathbuf, rq->bufsize))
	errno = 0;
      else if (!errno)
	errno = EINVAL;
      if (fchdir (fd))
	warn ("fchdir: %m\n");
    }
    break;
  case AIOD_STAT:
    stat (rq->path1 (), rq->statbuf ());
    break;
  case AIOD_LSTAT:
    lstat (rq->path1 (), rq->statbuf ());
    break;
  default:
    panic ("aiosrv::pathop: bad op %d\n", rq->op);
    break;
  }
  if (errno)
    rq->err = errno;
}

void
aiosrv::dhop (aiomsg_t msg)
{
  dirent *dp;
  aiod_fhop *rq = buf->Xtmpl getptr<aiod_fhop> (msg);
  aiod_file *af = buf->Xtmpl getptr<aiod_file> (rq->fh);

  if (rq->op == AIOD_OPENDIR) {
    dht.create (af, &rq->err);
    return;
  }
  if (rq->op == AIOD_CLOSEDIR) {
    dht.close (af, &rq->err);
    return;
  }

  DIR *fd = dht.lookup (af, &rq->err);
  if (fd == NULL)
    return;

  errno = 0;
  switch (rq->op) {
  case AIOD_READDIR:
    dp = readdir (fd);
    if (dp != NULL) {
      rq->iobuf.len = sizeof(dirent);
      memcpy(buf->getbuf (&rq->iobuf), dp, sizeof(dirent));
    }
    else {
      rq->iobuf.len = 0;
    }
    break;
  default:
    panic ("aiosrv::dhop: bad op %d\n", rq->op);
    break;
  }
  if (errno)
    rq->err = errno;
}

void
aiosrv::fhop (aiomsg_t msg)
{
  aiod_fhop *rq = buf->Xtmpl getptr<aiod_fhop> (msg);
  aiod_file *af = buf->Xtmpl getptr<aiod_file> (rq->fh);

  if (rq->op == AIOD_OPEN) {
    fht.create (af, rq->mode, &rq->err);
    return;
  }
  if (rq->op == AIOD_CLOSE) {
    fht.close (af, &rq->err);
    return;
  }

  int fd = fht.lookup (af, &rq->err);
  if (fd < 0)
    return;

  errno = 0;
  switch (rq->op) {
  case AIOD_FSYNC:
    fsync (fd);
    break;
  case AIOD_FTRUNC:
    rc_ignore (ftruncate (fd, rq->length));
    break;
  case AIOD_READ:
#ifdef HAVE_PREAD
    if (rq->iobuf.pos == -1)
      rq->iobuf.len = read (fd, buf->getbuf (&rq->iobuf), rq->iobuf.len);
    else
      rq->iobuf.len = pread (fd, buf->getbuf (&rq->iobuf), rq->iobuf.len,
			     rq->iobuf.pos);
#else /* !HAVE_PREAD */
    if (rq->iobuf.pos == -1 || lseek (fd, rq->iobuf.pos, SEEK_SET) != -1)
      rq->iobuf.len = read (fd, buf->getbuf (&rq->iobuf), rq->iobuf.len);
    else
      rq->iobuf.len = -1;
#endif /* !HAVE_PREAD */
    break;
  case AIOD_WRITE:
#ifdef HAVE_PWRITE
    if (rq->iobuf.pos == -1)
      rq->iobuf.len = write (fd, buf->getbuf (&rq->iobuf), rq->iobuf.len);
    else
      rq->iobuf.len = pwrite (fd, buf->getbuf (&rq->iobuf), rq->iobuf.len,
			      rq->iobuf.pos);
#else /* !HAVE_PWRITE */
    if (rq->iobuf.pos == -1 || lseek (fd, rq->iobuf.pos, SEEK_SET) != -1)
      rq->iobuf.len = write (fd, buf->getbuf (&rq->iobuf), rq->iobuf.len);
    else
      rq->iobuf.len = -1;
#endif /* !HAVE_PWRITE */
    break;
  default:
    panic ("aiosrv::fhop: bad op %d\n", rq->op);
    break;
  }
  if (errno)
    rq->err = errno;
}

void
aiosrv::fstat (aiomsg_t msg)
{
  aiod_fstat *rq = buf->Xtmpl getptr<aiod_fstat> (msg);
  aiod_file *af = buf->Xtmpl getptr<aiod_file> (rq->fh);

  if (rq->op != AIOD_FSTAT)
    panic ("aiosrv::fstat: bad op %d\n", rq->op);

  int fd = fht.lookup (af, &rq->err);
  if (fd < 0)
    return;
  errno = 0;
  ::fstat (fd, &rq->statbuf);
  if (errno)
    rq->err = errno;
}

static char zbuf[0x10000];
void
aiosrv::nop (aiomsg_t msg)
{
  /* If the shmfile is sparse, a nop forces allocation. */
  aiod_nop *rq = buf->Xtmpl getptr<aiod_nop> (msg);
  size_t sz = 0;
  bool touchable = rq->nopsize;
  if (lseek (buf->fd, msg, SEEK_SET) != -1) {
    size_t count = max (rq->nopsize, sizeof (*rq));
    while (sz < count) {
      ssize_t n = write (buf->fd, zbuf, min (count - sz, sizeof (zbuf)));
      if (n <= 0)
	break;
      sz += n;
    }
  }
  if (sz >= sizeof (*rq)) {
    msync (reinterpret_cast<char *> (rq), sz, 0);
    rq->nopsize = sz;
  }
  else if (touchable) {
    rq->err = errno;
    rq->nopsize = 0;
  }
}

#ifdef MAINTAINER
bool aiodtrace = getenv ("AIOD_TRACE");
void
aiod_dump (void *buf)
{
  aiod_reqhdr *rqh = (aiod_reqhdr *) buf;
  switch (rqh->op) {
  case AIOD_NOP:
    warnx ("AIOD_TRACE: NOP, err %d, nopsize %ld\n", rqh->err,
	   long (((aiod_nop *) rqh)->nopsize));
    break;
  default:
    warnx ("AIOD_TRACE: op %d, err %d\n", rqh->op, rqh->err);
    break;
  }
}
#else /* !MAINTAINER */
enum { aiodtrace = false };
#endif /* !MAINTAINER */

void
aiosrv::getmsg (aiomsg_t msg)
{
  aiod_op op = buf->getop (msg);

  if (op == AIOD_NOP)
    nop (msg);
  else if (op >= AIOD_UNLINK && op <= AIOD_LSTAT)
    pathop (msg);
  else if (op >= AIOD_OPEN && op <= AIOD_WRITE)
    fhop (msg);
  else if (op == AIOD_FSTAT)
    fstat (msg);
  else if (op >= AIOD_OPENDIR && op <= AIOD_CLOSEDIR)
    dhop (msg);
  else
    fatal ("bad opcode %d from client\n", op);

  if (aiodtrace)
    aiod_dump (buf->Xtmpl getptr<void> (msg));
  if (write (fd, &msg, sizeof (msg)) != sizeof (msg)) {
    if (errno != EPIPE)
      fatal ("aiosrv::write: %m\n");
    exit (0);
  }
}

static int
fullread (int fd, void *buf, size_t len)
{
  char *bp = static_cast<char *> (buf);
  while (len > 0) {
    ssize_t n = read (fd, bp, len);
    if (n <= 0)
      return n;
    bp += n;
    len -= n;
  }
  return bp - static_cast<char *> (buf);
}

static void
sigio_handler (int sig)
{
  int saved_errno = errno;
  static timeval ztv = { 0, 0 };

  sigio_received = 1;
  flock (shmfd, LOCK_UN);
  while (fdwait (rwfd, selread, &ztv) > 0) {
    /* Note that rwfd is a unix domain socket, not a pipe.  Since unix
     * domain sockets are not specified by POSIX, writes of less than
     * PIPE_BUF bytes could conceivably not be atomic on some systems,
     * and a true select for writing might not even guarantee PIPE_BUF
     * bytes of buffer space.
     *
     * We thereforedo a "fullread" here.  We know the client is trying
     * to write sizeof (msg) bytes to us, so we don't need to worry
     * about blocking. */
    aiomsg_t msg;
    ssize_t n = fullread (rwfd, &msg, sizeof (msg));
    if (n != sizeof (msg)) {
      if (n < 0)
	fatal ("read: %m\n");
      exit (0);
    }
    srv->getmsg (msg);
  }
  errno = saved_errno;
}

static void
msg_loop ()
{
  sigio_handler (SIGIO);	// Ensure we didn't miss any SIGIO's
  sigprocmask (SIG_UNBLOCK, &sigio_mask, NULL);

  for (;;) {
    sigio_received = 0;

    while (flock (shmfd, LOCK_EX) == -1) {
      if (errno != EINTR) {
	sigprocmask (SIG_BLOCK, &sigio_mask, NULL);
	fatal ("flock: %m\n");
      }
      sigio_received = 0;
    }
    if (sigio_received) {
      flock (shmfd, LOCK_UN);
      continue;
    }

    aiomsg_t msg;
    ssize_t n = read (rfd, &msg, sizeof (msg));
    int saved_errno = errno;
    sigprocmask (SIG_BLOCK, &sigio_mask, NULL);
    flock (shmfd, LOCK_UN);

    if (n == -1 && saved_errno == EINTR && sigio_received) {
      sigprocmask (SIG_UNBLOCK, &sigio_mask, NULL);
      continue;
    }
    if (n != sizeof (msg)) {
      if (n < 0)
	fatal ("read: %m\n");
      if (n > 0)
	fatal ("short read from rfd\n");
      exit (0);
    }

    srv->getmsg (msg);
    sigprocmask (SIG_UNBLOCK, &sigio_mask, NULL);
  }
}

static void
usage ()
{
  warnx << "usage: " << progname << " shmfd rfd rwfd\n";
  exit (1);
}

#ifdef MAINTAINER
void
nop (int sig)
{
}
#endif /* MAINTAINER */

int
main (int argc, char **argv)
{
  setprogname (argv[0]);
  if (argc != 4
      || !convertint (argv[1], &shmfd)
      || !convertint (argv[2], &rfd)
      || !convertint (argv[3], &rwfd))
    usage ();

#ifdef MAINTAINER
  if (getenv ("AIOD_PAUSE")) {
    struct sigaction sa;
    bzero (&sa, sizeof (sa));
    sa.sa_handler = nop;
#ifdef SA_RESETHAND
    sa.sa_flags = SA_RESETHAND;
#endif /* SA_RESETHAND */
    sigaction (SIGCONT, &sa, NULL);
    warn ("pid %d, pausing\n", int (getpid ()));

    timeval tv;
    tv.tv_sec = 60;
    tv.tv_usec = 0;
    select (0, NULL, NULL, NULL, &tv);
  }
#endif /* MAINTAINER */

  umask (0);

  ptr<shmbuf> b = shmbuf::alloc (shmfd);
  if (!b)
    fatal ("could not map shared memory buffer\n");

  srv = New aiosrv (rwfd, b);

  (void) sigemptyset (&sigio_mask);
  sigaddset (&sigio_mask, SIGIO);
  sigprocmask (SIG_BLOCK, &sigio_mask, NULL);

  struct sigaction sa;
  bzero (&sa, sizeof (sa));
  sa.sa_handler = sigio_handler;
  if (sigaction (SIGIO, &sa, NULL) < 0)
    fatal ("sigaction: %m\n");

  /* Since the client code might not necessarily tolerate the death of
   * an aiod process, put ourselves in a new process group.  That way
   * if the parent process catches terminal signals like SIGINT, the
   * aiods will not die.  This allows the parent still to issue aio
   * requests while handling the signal.  Additionally, if sigio_set
   * uses SIOCSPGRP, starting a new process group avoids hitting other
   * processes with our SIGIO signals. */
  setpgid (0, 0);

  if (sigio_set (rwfd) < 0)
    fatal ("could not enable SIGIO\n");

  msg_loop ();

  return 1;
}
