/* $Id: suio++.C 1117 2005-11-01 16:20:39Z max $ */

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

#include "suio++.h"

#ifdef DMALLOC

/* Simple, IP-like checksum */
static u_int16_t
cksum (const void *_data, int len)
{
  const u_char *data = static_cast<const u_char *> (_data);
  union {
    u_char b[2];
    u_int16_t w;
  } bwu;
  u_int32_t sum;
  const u_char *end;

  if (!len)
    return 0;

  bwu.w = 0;
  if ((u_long) data & 1) {
    bwu.b[0] = *data;
    sum = cksum ((char *) data + 1, len - 1);
    sum = ((sum >> 8) & 0xff) | ((sum & 0xff) << 8);
    sum += bwu.w;
    len = 1;
  }
  else
    sum = 0;

  end = data + (len & ~1);
  while (data < end) {
    sum += *(u_int16_t*) data;
    data += sizeof (u_int16_t);
  }
  if (len & 1) {
    bwu.b[0] = *data;
    sum += bwu.w;
  }

  return ~sum;
}

struct suio_check_dat {
  const void *buf;
  size_t len;
  u_int16_t sum;
  const char *line;
};

static void
suio_docheck (suio_check_dat d)
{
  if (cksum (d.buf, d.len) != d.sum)
    panic ("%s: data in suio subsquently changed!\n", d.line);
}

void
__suio_check (const char *line, suio *uio, const void *buf, size_t len)
{
  suio_check_dat d = { buf, len, cksum (buf, len), line };
  uio->iovcb (wrap (suio_docheck, d));
}

void
__suio_printcheck (const char *line, suio *uio, const void *buf, size_t len)
{
  static bool do_check = dmalloc_debug_current () & 0x200000; // free-blank
  uio->print (buf, len);
  if (do_check) {
    suio_check_dat d = { buf, len, cksum (buf, len), line };
    uio->iovcb (wrap (suio_docheck, d));
  }
}

#endif /* DMALLOC */

size_t
iovsize (const iovec *iov, int cnt)
{
  const iovec *end = iov + cnt;
  size_t size = 0;
  while (iov < end)
    size += iov++->iov_len;
  return size;
}

iovmgr::iovmgr (const iovec *v, int iovcnt)
  : iov (v), lim (iov + iovcnt)
{
  if (iov < lim)
    cur = *iov++;
  else {
    iov = lim = NULL;
    cur.iov_base = NULL;
    cur.iov_len = 0;
  }
}

void
iovmgr::skip (size_t n)
{
  if (n < implicit_cast<size_t> (cur.iov_len)) {
    cur.iov_base = (char *) cur.iov_base + n;
    cur.iov_len -= n;
    return;
  }
  n -= cur.iov_len;
  while (iov < lim && n >= implicit_cast<size_t> (iov->iov_len))
    n -= iov++->iov_len;
  if (n) {
    if (iov == lim || n > implicit_cast<size_t> (iov->iov_len))
      panic ("iovmgr: skip value larger than iovsize\n");
    cur.iov_base = (char *) iov->iov_base + n;
    cur.iov_len = iov->iov_len - n;
    iov++;
  }
  else {
    cur.iov_base = NULL;
    cur.iov_len = 0;
  }
}

size_t
iovmgr::copyout (char *buf, size_t len)
{
  if (len < implicit_cast<size_t> (cur.iov_len)) {
    memcpy (buf, cur.iov_base, len);
    cur.iov_base = (char *) cur.iov_base + len;
    cur.iov_len -= len;
    return len;
  }
  memcpy (buf, cur.iov_base, cur.iov_len);
  char *cp = buf + cur.iov_len;
  char *eom = buf + len;

  while (iov < lim
	 && implicit_cast<size_t> (iov->iov_len) <= (size_t) (eom - cp)) {
    memcpy (cp, iov->iov_base, iov->iov_len);
    cp += iov++->iov_len;
  }

  if (iov == lim) {
    cur.iov_base = NULL;
    cur.iov_len = 0;
  }
  else if (cp < eom) {
    size_t n = eom - cp;
    memcpy (cp, iov->iov_base, n);
    cp += n;
    cur.iov_base = (char *) iov->iov_base + n;
    cur.iov_len = iov->iov_len - n;
    iov++;
  }
  else
    cur = *iov++;
  return cp - buf;
}

size_t
iovmgr::size () const
{
  size_t n = cur.iov_len;
  for (const iovec *v = iov; v < lim; v++)
    n += v->iov_len;
  return n;
}

void
suio::makeuiocbs ()
{
  callback<void>::ptr cb;
  while (!uiocbs.empty () && uiocbs.front ().nbytes <= nrembytes) {
    // it is safer to pop first, and then call. 
    cb = uiocbs.pop_front ().cb;
    (*cb) ();
  }
}

suio::suio ()
  : uiobytes (0), nrembytes (0), nremiov (0),
    lastiovend (NULL),
    scratch_buf (defbuf), scratch_pos (defbuf),
    scratch_lim (defbuf + sizeof (defbuf)),
    allocator (default_allocator), deallocator (default_deallocator)
{
}

suio::~suio ()
{
  clear ();
}

void
suio::clear ()
{
  rembytes (resid ());
  /* XXX - GCC BUG:  To help derived classes work around compiler
   * bugs, we actually must clear all memory associated with the suio
   * structure.  Gcc often fails to call the destructor, so calling
   * clear should be a viable workaround. */
  if (scratch_buf != defbuf) {
    deallocator (scratch_buf, scratch_lim - scratch_buf);
    scratch_buf = defbuf;
    scratch_lim = defbuf + sizeof (defbuf);
  }
  scratch_pos = defbuf;
  iovs.clear ();
  uiocbs.clear ();
}

char *
suio::morescratch (size_t size)
{
  size = ((size + MALLOCRESV + (blocksize - 1))
	  & ~(blocksize - 1)) - MALLOCRESV;
  if (scratch_buf != defbuf)
    iovcb (wrap (deallocator, scratch_buf, scratch_lim - scratch_buf));
  scratch_buf = scratch_pos = static_cast<char *> (allocator (size));
  scratch_lim = scratch_buf + size;
  return scratch_buf;
}

void
suio::slowfill (char c, size_t len)
{
  size_t n = scratch_lim - scratch_pos;
  if (len <= n) {
    memset (scratch_pos, c, len);
    pushiov (scratch_pos, len);
  }
  else {
    if (n >= smallbufsize || scratch_pos == lastiovend) {
      memset (scratch_pos, c, n);
      pushiov (scratch_pos, n);
      len -= n;
    }
    morescratch (len);
    memset (scratch_pos, c, len);
    pushiov (scratch_pos, len);
  }
}

void
suio::slowcopy (const void *_buf, size_t len)
{
  const char *buf = static_cast<const char *> (_buf);
  size_t n = scratch_lim - scratch_pos;
  if (len <= n) {
    memcpy (scratch_pos, buf, len);
    pushiov (scratch_pos, len);
  }
  else {
    if (n >= smallbufsize || scratch_pos == lastiovend) {
      memcpy (scratch_pos, buf, n);
      pushiov (scratch_pos, n);
      buf += n;
      len -= n;
    }
    morescratch (len);
    memcpy (scratch_pos, buf, len);
    pushiov (scratch_pos, len);
  }
}

void
suio::copyv (const iovec *iov, size_t cnt, size_t skip)
{
  iovmgr iom (iov, cnt);
  iom.skip (skip);

  size_t n = scratch_lim - scratch_pos;
  if (scratch_pos == lastiovend || n >= smallbufsize) {
    size_t m = iom.copyout (scratch_pos, scratch_lim - scratch_pos);
    if (m > 0)
      pushiov (scratch_pos, m);
  }

  n = iom.size ();
  if (n > 0) {
    morescratch (n);
    iom.copyout (scratch_pos, n);
    pushiov (scratch_pos, n);
  }
}

void
suio::take (suio *uio)
{
  int64_t bdiff = nrembytes + uiobytes - uio->nrembytes;

  uio->nrembytes += uio->uiobytes;
  uio->nremiov += uio->iovs.size ();
  uio->uiobytes = 0;
  for (iovec *v = uio->iovs.base (), *e = uio->iovs.lim (); v < e; v++)
    if (v->iov_base >= uio->defbuf
	&& v->iov_base < uio->defbuf + sizeof (uio->defbuf))
      copy (v->iov_base, v->iov_len);
    else
      pushiov (v->iov_base, v->iov_len);
  uio->iovs.clear ();

  for (uiocb *c = uio->uiocbs.base (), *e = uio->uiocbs.lim (); c < e; c++)
    uiocbs.push_back (uiocb (c->nbytes + bdiff, c->cb));
  uio->uiocbs.clear ();

  uio->scratch_buf = uio->scratch_pos = uio->defbuf;
  uio->scratch_lim = uio->defbuf + sizeof (uio->defbuf);
}

void
suio::rembytes (size_t n)
{
  assert (n <= uiobytes);	// error to remove more than we have

  uiobytes -= n;
  nrembytes += n;

  iovec *iov = iovs.base (), *end = iovs.lim ();
  while (iov < end && n >= implicit_cast<size_t> (iov->iov_len))
    n -= iov++->iov_len;
  if (n > 0) {
    assert (iov < end);		// else uiobytes was incorrect
    iov->iov_base = static_cast<char *> (iov->iov_base) + n;
    iov->iov_len -= n;
  }

  size_t niov = iov - iovs.base ();
  iovs.popn_front (niov);
  nremiov += niov;
  if (iovs.empty ()) {
    scratch_pos = scratch_buf;
    lastiovend = NULL;
  }
  makeuiocbs ();
}

int
suio::output (int fd, int cnt)
{
  u_int64_t startpos = nrembytes;
  ssize_t n = 0;
  if (cnt < 0)
    while (!iovs.empty ()
	   && (n = writev (fd, const_cast<iovec *> (iov ()),
			   min (iovcnt (), (size_t) UIO_MAXIOV))) > 0)
      rembytes (n);
  else {
    assert ((size_t) cnt <= iovs.size ());
    u_int64_t maxiovno =  nremiov + cnt;
    while (nremiov < maxiovno
	   && (n = writev (fd, const_cast<iovec *> (iov ()),
			   min(maxiovno - nremiov,
			       (u_int64_t) UIO_MAXIOV))) > 0)
      rembytes (n);
  }
  if (n < 0 && errno != EAGAIN)
    return -1;
  return nrembytes > startpos;
}

size_t
suio::copyout (void *_buf, size_t len) const
{
  char *buf = static_cast<char *> (_buf);
  char *cp = buf;
  for (const iovec *v = iov (), *e = iovlim (); v < e; v++) {
    if (len >= implicit_cast<size_t> (v->iov_len)) {
      memcpy (cp, v->iov_base, v->iov_len);
      cp += v->iov_len;
      len -= v->iov_len;
    }
    else {
      memcpy (cp, v->iov_base, len);
      return cp - buf + len;
    }
  }
  return cp - buf;
}

int
suio::input (int fd, size_t len)
{
  size_t space = scratch_lim - scratch_pos;

  if (len <= space || !space) {
    void *buf = getspace (len);
    ssize_t n = read (fd, buf, len);
    if (n > 0)
      pushiov (buf, n);
    return n;
  }

  size_t size = ((len - space + MALLOCRESV + (blocksize - 1))
		 & ~(blocksize - 1)) - MALLOCRESV;
  char *buf = static_cast<char *> (allocator (size));

  iovec iov[2];
  iov[0].iov_base = scratch_pos;
  iov[0].iov_len = space;
  iov[1].iov_base = buf;
  iov[1].iov_len = len - space;

  ssize_t n = readv (fd, iov, 2);
  if (n > 0 && (size_t) n > space) {
    pushiov (iov[0].iov_base, iov[0].iov_len);
    assert (scratch_pos == scratch_lim);
    if (scratch_buf != defbuf)
      iovcb (wrap (deallocator, scratch_buf, scratch_lim - scratch_buf));
    scratch_pos = scratch_buf = buf;
    scratch_lim = buf + size;
    pushiov (scratch_pos, n - space);
  }
  else {
    if (n > 0)
      pushiov (iov[0].iov_base, n);
    deallocator (buf, size);
  }
  return n;
}

#ifndef DMALLOC
char *
suio_flatten (const struct suio *uio)
#else /* DMALLOC */
char *
__suio_flatten (const struct suio *uio, const char *file, int line)
#endif /* DMALLOC */
{
#ifndef DMALLOC
  char *buf = (char *) xmalloc (uio->resid ());
#else /* DMALLOC */
  char *buf = (char *) _xmalloc_leap (file, line, uio->resid ());
#endif /* DMALLOC */
  uio->copyout (buf);
  return buf;
}
