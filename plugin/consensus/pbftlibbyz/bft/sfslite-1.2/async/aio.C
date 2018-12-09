/* $Id: aio.C 3758 2008-11-13 00:36:00Z max $ */

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

#include "aiod.h"

aiobuf::aiobuf (aiod *d, size_t p, size_t l)
  : buf (d->shmbuf + p), len (l), iod (d), pos (p)
{
#ifdef DMALLOC
  if (len) {
    memset (base (), 0xc5, len);
    memset (base () + len, 0xd1, (1 << log2c (len)) - len);
  }
#endif /* DMALLOC */
  iod->addref ();
}

aiobuf::~aiobuf ()
{
  if (len) {
#ifdef DMALLOC
    memset (base (), 0xc5, len);
    for (char *p = base () + len, *e = base () + (1 << log2c (len));
	 p < e; p++)
      if (static_cast<u_char> (*p) != 0xd1)
	panic ("aiobuf: buffer was overrun\n");
#endif /* DMALLOC */
    iod->bb.dealloc (pos, len);
    if (!iod->bbwaitq.empty ())
      iod->bufwake ();
  }
  iod->delref ();
}

void
aiod::writeq::output ()
{
  char buf[maxwrite];
  size_t wsize = min ((size_t) maxwrite, wbuf.resid ());
  assert (wsize);

  wbuf.copyout (buf, wsize);
  ssize_t n = write (wfd, buf, wsize);
  if (n < 0)
    fatal ("write to aiod failed (%m)\n"); // XXX - should make aiod fail
  wbuf.rembytes (n);
  if (!wbuf.resid ())
    fdcb (wfd, selwrite, NULL);
}

void
aiod::writeq::sendmsg (aiomsg_t msg)
{
  static timeval ztv = { 0, 0 };
  bool wasempty = !wbuf.resid ();
  if (!wasempty || fdwait (wfd, selwrite, &ztv) < 1) {
    wbuf.copy (&msg, sizeof (msg));
    if (wasempty)
      fdcb (wfd, selwrite, wrap (this, &aiod::writeq::output));
  }
  else {
    ssize_t n = write (wfd, &msg, sizeof (msg));
    if (n < 0)
      fatal ("write to aiod failed (%m)\n");
    if (n != sizeof (msg)) {
      // Writes less than PIPE_BUF were supposed to be atomic
      warn ("aiod::writeq::sendmsg: partial write (%d bytes)\n", (int) n);
      wbuf.copy (reinterpret_cast<char *> (&msg) + n, sizeof (msg) - n);
      fdcb (wfd, selwrite, wrap (this, &aiod::writeq::output));
    }
  }
}

bool
aiod::daemon::launch (str path, int shmfd, int commonfd)
{
  assert (pid == -1);		// Otherwise, already launched!

  int fds[2];
  if (socketpair (AF_UNIX, SOCK_STREAM, 0, fds) < 0)
    fatal ("aiod::daemon::launch: socketpair failed: %m\n");
  wq.wfd = fd = fds[0];
  close_on_exec (fd);

  str shmfdarg (strbuf ("%d", shmfd));
  str rfdarg (strbuf ("%d", commonfd));
  str rwfdarg (strbuf ("%d", fds[1]));

  char *av[5] = { const_cast<char *> (path.cstr ()),
		  const_cast<char *> (shmfdarg.cstr ()),
		  const_cast<char *> (rfdarg.cstr ()),
		  const_cast<char *> (rwfdarg.cstr ()),
		  NULL };

  pid = spawn (path, av);
  close (fds[1]);
  if (pid < 0) {
    warn << path << ": " << strerror (errno) << "\n";
    return false;
  }
  return true;
}

aiod::aiod (u_int nproc, ssize_t shmsize, size_t mb, bool sp,
	    str path, str tmpdir)
  : closed (false), finalized (false), growlock (false),
    bufwakereq (false), bufwakelock (false), shmpin (sp),
    refcnt (0), shmmax ((shmsize + mb - 1) & ~(mb - 1)), shmlen (0),
    bb (shmlen, minbuf, mb), ndaemons (nproc), fhno_ctr (1), maxbuf (mb)
{
  assert (shmsize > 0);
  static const char *const templates[] = {
    "/var/tmp/aioshmXXXXXXXX",
    "/usr/tmp/aioshmXXXXXXXX",
    "/tmp/aioshmXXXXXXXX",
    NULL
  };

  str tmpfile;
  mode_t m = umask (077);

  if (!tmpdir)
    tmpdir = safegetenv ("TMPDIR");
  if (tmpdir && tmpdir.len ()) {
    if (tmpdir[tmpdir.len () - 1] == '/')
      tmpdir = strbuf () << tmpdir << "aioshmXXXXXXXX";
    else
      tmpdir = strbuf () << tmpdir << "/aioshmXXXXXXXX";
    char *temp = xstrdup (tmpdir);
    shmfd = mkstemp (temp);
    if (shmfd > 0)
      tmpfile = temp;
    xfree (temp);
  }
  else
    for (const char *const *p = templates; *p && !tmpfile; p++) {
      char *temp = xstrdup (*p);
      shmfd = mkstemp (temp);
      if (shmfd > 0)
	tmpfile = temp;
      xfree (temp);
    }
  if (!tmpfile)
    fatal ("aiod: could not create temporary file: %m\n");
  umask (m);
  close_on_exec (shmfd);
  if (ftruncate (shmfd, shmmax) < 0)
    fatal ("aiod: could not grow shared mem file (%m)\n");
  struct stat sb;
  if (fstat (shmfd, &sb) < 0)
    fatal ("fstat (%s): %m\n", tmpfile.cstr ());

  shmbuf = static_cast<char *>
    (mmap (NULL, (size_t) shmmax, PROT_READ|PROT_WRITE,
	   MAP_FILE|MAP_SHARED, shmfd, 0));
  if (shmbuf == (char *) MAP_FAILED)
    fatal ("aiod: could not mmap shared mem file (%m)\n");

  int fds[2];
  if (pipe (fds) < 0)
    fatal ("aiod: pipe syscall failed: %m\n");
  wq.wfd = fds[1];
  close_on_exec (wq.wfd);
  int rfd = fds[0];
  shutdown (rfd, 1);

  if (!path)
    path = "aiod";
  str aiod_path = fix_exec_path (path);
  dv = New daemon[ndaemons];
  for (u_int i = 0; i < ndaemons; i++) {
    /* We have to reopen the temporary file for each daemon, because
     * the daemons use flock for synchronization.  If everyone tried
     * to flock the same file descriptor, there would be no
     * synchronization.  (The same is not true of fcntl locking, but
     * wherever possible we use flock as it is faster.)  */
    int fd = ::open (tmpfile, O_RDWR);
    if (fd < 0)
      fatal ("cannot reopen %s: %m\n", tmpfile.cstr ());
    struct stat sb2;
    if (fstat (fd, &sb2) < 0)
      fatal ("fstat (%s): %m\n", tmpfile.cstr ());
    if (sb.st_dev != sb2.st_dev || sb.st_ino != sb2.st_ino)
      fatal ("aiod: somone tampered with %s\n", tmpfile.cstr ());

    bool res = dv[i].launch (aiod_path, fd, rfd);
    close (fd);
    if (!res) {
      fail ();
      break;
    }
    fdcb (dv[i].fd, selread, wrap (this, &aiod::input, i));
  }
  close (rfd);

  /* Right now it's a sparse file, so it's not the end of the world if
   * we die and leave it sitting around.  However, we unlink it before
   * consuming disk space to make sure it gets garbage collected
   * properly.  */
  if (::unlink (tmpfile) < 0)
    fatal ("aiod: unlink (%s): %m\n", tmpfile.cstr ());
}

aiod::~aiod ()
{
  fail ();
  if (munmap (shmbuf, shmlen) < 0)
    warn ("~aiod could not unmap shared mem: %m\n");
  close (shmfd);
  delete[] dv;
}

void
aiod::delreq (aiod::request *r)
{
  while (!r->cbvec.empty ())
    (*r->cbvec.pop_front ()) (NULL);
  rqtab.remove (r);
  delete r;
}

void
aiod::fail ()
{
  closed = true;
  wq.close ();
  for (size_t i = 0; i < ndaemons; i++)
    dv[i].wq.close ();
  rqtab.traverse (wrap (this, &aiod::delreq));
  for (int i = 0, n = bbwaitq.size (); i < n && !bbwaitq.empty (); i++)
    (*bbwaitq.pop_front ()) ();
  /* If we still have entries in bbwaitq, someone was ignoring the
   * closed flag.  This is a bug. */
  assert (bbwaitq.empty ());
}

void
aiod::input (int i)
{
  aiomsg_t buf[maxwrite/sizeof (aiomsg_t)];

  ssize_t n = read (dv[i].fd, buf, sizeof (buf));
  if (n <= 0) {
    if (n < 0)
      warn ("aiod: read: %m\n");
    else
      warn ("aiod: EOF\n");
    fail ();
    return;
  }
  if (n % sizeof (aiomsg_t)) {
    warn ("aiod: invalid read of %d bytes\n", (int) n);
    fail ();
    return;
  }

  addref ();
  assert (!bufwakelock);
  bufwakelock = true;

  for (aiomsg_t *op = buf, *ep = buf + (n / sizeof (aiomsg_t));
       op < ep; op++) {
    request *r = rqtab[*op];
    if (!r) {
      warn ("aiod: got invalid response 0x%lx\n", (u_long) *op);
      fail ();
      bufwakelock = false;
      return;
    }
    (*r->cbvec.pop_front ()) (r->buf);
    if (r->cbvec.empty ())
      delreq (r);
  }

  bufwakelock = false;
  if (bufwakereq)
    bufwake ();
  delref ();
}

void
aiod::sendmsg (ref<aiobuf> buf, cbb cb, int dst)
{
  if (closed) {
    (*cb) (NULL);
    return;
  }

  request *r = rqtab[buf->pos];
  if (!r) {
    r = New request (buf);
    rqtab.insert (r);
  }

  r->cbvec.push_back (cb);
  if (dst == -1)
    wq.sendmsg (buf->pos);
  else {
    assert (dst >= 0 && (u_int) dst < ndaemons);
    dv[dst].wq.sendmsg (buf->pos);
  }
}

void
aiod::bufwake ()
{
  if (bufwakelock) {
    bufwakereq = true;
    return;
  }
  bufwakelock = true;
  do {
    bufwakereq = false;
    vec<cbv> nq;
    swap (nq, bbwaitq);
    while (!nq.empty ())
      (*nq.pop_front ()) ();
  } while (bufwakereq);
  bufwakelock = false;
}

void
aiod::cbi_cb (cbi cb, ptr<aiobuf> buf)
{
  (*cb) (buf ? buf2hdr (buf)->err : EIO);
}

void
aiod::cbstat_cb (cbstat cb, ptr<aiobuf> buf)
{
  if (!buf)
    (*cb) (NULL, EIO);
  else {
    aiod_pathop *rq = buf2pathop (buf);
    /* Be slightly careful about "bad" aiod's clobbering shared
     * memory.  For instance, the callback may rely on its struct stat
     * argument not being NULL if the error is 0.  So we avoid code
     * like:
     *     if (rq->err)
     *       (*cb) (NULL, rq->err);
     */
    if (int err = rq->err)
      (*cb) (NULL, err);
    else
      (*cb) (rq->statbuf (), 0);
  }
}

void
aiod::pathret_cb (cbsi cb, ptr<aiobuf> buf)
{
  if (!buf)
    (*cb) (NULL, EIO);
  else {
    aiod_pathop *rq = buf2pathop (buf);
    if (int err = rq->err)
      (*cb) (NULL, err);
    else {
      /* We can't trust the aiod's not to set bufsize to something
       * weird. */
      size_t size = rq->bufsize;
      if (aiod_pathop::totsize (size) > buf->size ())
	(*cb) (NULL, EIO);
      else
	(*cb) (str (rq->pathbuf, size), 0);
    }
  }
}

void
aiod::open_cb (ref<aiofh> fh, cbopen cb, ptr<aiobuf> buf)
{
  if (!buf)
    (*cb) (NULL, EIO);
  else {
    aiod_fhop *rq = buf2fhop (buf);
    if (int err = rq->err)
      (*cb) (NULL, err);
    else
      (*cb) (fh, 0);
  }
}

ptr<aiobuf>
aiod::bufalloc (size_t len)
{
  assert (len > 0);
  assert (len <= bb.maxalloc ());
  ssize_t pos = bb.alloc (len);
  if (pos != -1)
    return New refcounted<aiobuf> (this, pos, len);
  if (!growlock && shmlen + maxbuf <= shmmax) {
    // XXX - inc must be multiple of maxbuf
    size_t inc = min (shmmax - shmlen, max<size_t> (maxbuf, shmlen >> 2));
    // XXX - can't allocate buf without tweaking bbuddy
    ref<aiobuf> buf (New refcounted<aiobuf> (this, shmlen, 0));
    aiod_nop *rq = buf2nop (buf);
    assert (!rq->op);		// Sparse file data must be 0
    growlock = true;
    sendmsg (buf, wrap (this, &aiod::bufalloc_cb1, inc));
  }
  return NULL;
}

void
aiod::bufalloc_cb1 (size_t inc, ptr<aiobuf> buf)
{
  if (buf && buf2nop (buf)->nopsize) {
    buf2nop (buf)->nopsize = inc;
    sendmsg (buf, wrap (this, &aiod::bufalloc_cb2, inc));
  }
  else
    growlock = false;
}

void
aiod::bufalloc_cb2 (size_t inc, ptr<aiobuf> buf)
{
  growlock = false;
  if (buf && buf2nop (buf)->nopsize == inc) {
    size_t oshmlen = shmlen;
    bb.settotsize (shmlen + inc);
    shmlen = bb.gettotsize ();
    if (shmpin && mlock (shmbuf + oshmlen, shmlen - oshmlen) < 0)
      warn ("could not pin aiod shared memory: %m\n");
    bufwake ();
  }
}


void
aiod::pathop (aiod_op op, str p1, str p2, cbb cb, size_t minsize)
{
  if (closed) {
    (*cb) (NULL);
    return;
  }

  size_t bufsize = p1.len () + 2;
  if (p2)
    bufsize += p2.len ();
  if (minsize > bufsize)
    bufsize = minsize;

  ptr<aiobuf> buf = bufalloc (aiod_pathop::totsize (bufsize));
  if (!buf) {
    bufwait (wrap (this, &aiod::pathop, op, p1, p2, cb, minsize));
    return;
  }

  aiod_pathop *rq = buf2pathop (buf);
  rq->op = op;
  rq->err = 0;
  rq->bufsize = bufsize;
  rq->setpath (p1, p2 ? p2.cstr () : "");
  sendmsg (buf, cb);
}

void
aiod::open (str path, int flags, int mode,
	    callback<void, ptr<aiofh>, int>::ref cb)
{
  if (closed) {
    (*cb) (NULL, NULL);
    return;
  }

  ptr<aiobuf> rqbuf, fhbuf;
  if ((rqbuf = bufalloc (sizeof (aiod_fhop))))
    fhbuf = bufalloc (sizeof (aiod_file) + path.len ());
  if (!rqbuf || !fhbuf) {
    bufwait (wrap (this, &aiod::open, path, flags, mode, cb));
    return;
  }

  aiod_file *af = buf2file (fhbuf);
  bzero (af, sizeof (*af));
  af->oflags = flags;
  strcpy (af->path, path);
  ref<aiofh> fh = New refcounted<aiofh> (this, fhbuf);

  aiod_fhop *rq = buf2fhop (rqbuf);
  rq->op = AIOD_OPEN;
  rq->err = 0;
  rq->fh = fhbuf->pos;
  rq->mode = mode;

  sendmsg (rqbuf, wrap (open_cb, fh, cb));
}

void
aiod::opendir (str path,
	       callback<void, ptr<aiofh>, int>::ref cb)
{
  if (closed) {
    (*cb) (NULL, NULL);
    return;
  }

  ptr<aiobuf> rqbuf, fhbuf;
  if ((rqbuf = bufalloc (sizeof (aiod_fhop))))
    fhbuf = bufalloc (sizeof (aiod_file) + path.len ());
  if (!rqbuf || !fhbuf) {
    bufwait (wrap (this, &aiod::opendir, path, cb));
    return;
  }

  aiod_file *af = buf2file (fhbuf);
  bzero (af, sizeof (*af));
  strcpy (af->path, path);
  ref<aiofh> fh = New refcounted<aiofh> (this, fhbuf, true);

  aiod_fhop *rq = buf2fhop (rqbuf);
  rq->op = AIOD_OPENDIR;
  rq->err = 0;
  rq->fh = fhbuf->pos;
  
  sendmsg (rqbuf, wrap (open_cb, fh, cb));
}

aiofh::aiofh (aiod *d, ref<aiobuf> f, bool dir)
  : iod (d), fh (f), fhno (iod->fhno_alloc ()), isdir (dir), closed (false) 
{
  aiod_file *af = aiod::buf2file (fh);
  af->handle = fhno;
}

aiofh::~aiofh ()
{
  if (!closed)
    sendclose ();
  iod->fhno_free (fhno);
}

void
aiofh::sendclose (cbi::ptr cb)
{
  if (iod->closed) {
    if (cb)
      (*cb) (EBADF);
    return;
  }

  closed = true;

  ptr<aiobuf> buf = iod->bufalloc (sizeof (aiod_fhop));
  if (!buf) {
    iod->bufwait (wrap (mkref (this), &aiofh::sendclose, cb));
    return;
  }
  aiod_fhop *rq = aiod::buf2fhop (buf);

  rq->op = isdir ? AIOD_CLOSEDIR : AIOD_CLOSE;
  rq->err = 0;
  rq->fh = fh->pos;

  int *ctr = New int;
  aiod::cbb ccb (wrap (close_cb, ctr, cb));

  *ctr = iod->ndaemons;
  for (size_t i = 0; i < iod->ndaemons; i++)
    iod->sendmsg (buf, ccb, i);
}

void
aiofh::close (cbi cb)
{
  if (closed)
    (*cb) (EBADF);
  else
    sendclose (cb);
}

void
aiofh::closedir (cbi cb)
{
  close (cb);
}

void
aiofh::simpleop (aiod_op op, aiod::cbb cb, off_t length)
{
  if (closed || iod->closed) {
    (*cb) (NULL);
    return;
  }

  const size_t bufsize = ((op == AIOD_FSTAT) ? sizeof (aiod_fstat)
			  : sizeof (aiod_fhop));
  ptr<aiobuf> buf = iod->bufalloc (bufsize);
  if (!buf) {
    iod->bufwait (wrap (mkref (this), &aiofh::simpleop, op, cb, length));
    return;
  }
  aiod_fhop *rq = aiod::buf2fhop (buf);

  rq->op = op;
  rq->err = 0;
  rq->fh = fh->pos;
  rq->length = length;
  iod->sendmsg (buf, cb);
}

void
aiofh::rw (aiod_op op, off_t pos, ptr<aiobuf> iobuf,
	   u_int iostart, u_int iosize, cbrw cb)
{
  assert (iobuf->iod == iod);
  assert (iostart < iobuf->len);
  assert (iosize > 0 && iosize <= iobuf->len - iostart);

  if (closed || iod->closed) {
    (*cb) (NULL, -1, EBADF);
    return;
  }
  ptr<aiobuf> rqbuf = iod->bufalloc (sizeof (aiod_fhop));
  if (!rqbuf) {
#if 0
    // XXX - wrap has limit of 5 arguments
    iod->bufwait (wrap (mkref (this), &aiofh::rw, op, pos,
			iobuf, iostart, iosize, cb));
    return;
#else
    switch (op) {
    case AIOD_READDIR:
      iod->bufwait (wrap (mkref (this), &aiofh::sreaddir, pos,
			  iobuf, iostart, iosize, cb));
      return;
    case AIOD_READ:
      iod->bufwait (wrap (mkref (this), &aiofh::sread, pos,
			  iobuf, iostart, iosize, cb));
      return;
    case AIOD_WRITE:
      iod->bufwait (wrap (mkref (this), &aiofh::swrite, pos,
			  iobuf, iostart, iosize, cb));
      return; 
    default:
      panic ("aiofh::rw: unknown operation %d\n", op);
    }
#endif
  }
  aiod_fhop *rq = aiod::buf2fhop (rqbuf);

  rq->op = op;
  rq->err = 0;
  rq->fh = fh->pos;
  rq->iobuf.pos = pos;
  rq->iobuf.buf = iobuf->pos + iostart;
  rq->iobuf.len = iosize;
  iod->sendmsg (rqbuf, wrap (mkref (this), &aiofh::rw_cb, iobuf, cb));
}

void
aiofh::rw_cb (ref<aiobuf> iobuf, cbrw cb, ptr<aiobuf> rqbuf)
{
  if (!rqbuf)
    (*cb) (NULL, -1, EIO);
  else {
    aiod_fhop *rq = aiod::buf2fhop (rqbuf);
    int err = rq->err;
    ssize_t len = rq->iobuf.len;
    if (!err && (len < 0 || (size_t) len > iobuf->size ()))
      err = EIO;
    if (err)
      (*cb) (NULL, -1, err);
    else
      (*cb) (iobuf, len, 0);
  }
}

void
aiofh::cbstat_cb (aiod::cbstat cb, ptr<aiobuf> buf)
{
  if (!buf)
    (*cb) (NULL, EIO);
  else {
    aiod_fstat *rq = aiod::buf2fstat (buf);
    /* Be slightly careful about "bad" aiod's clobbering shared
     * memory.  For instance, the callback may rely on its struct stat
     * argument not being NULL if the error is 0.  So we avoid code
     * like:
     *     if (rq->err)
     *       (*cb) (NULL, rq->err);
     */
    if (int err = rq->err)
      (*cb) (NULL, err);
    else
      (*cb) (&rq->statbuf, 0);
  }
}

void
aiofh::close_cb (int *ctr, cbi::ptr cb, ptr<aiobuf> buf)
{
  if (!--*ctr) {
    delete ctr;
    if (cb)
      (*cb) (buf ? aiod::buf2fhop (buf)->err : EIO);
  }
}

void
aiofh::cbi_cb (cbi cb, ptr<aiobuf> buf)
{
  (*cb) (buf ? aiod::buf2hdr (buf)->err : EIO);
}

