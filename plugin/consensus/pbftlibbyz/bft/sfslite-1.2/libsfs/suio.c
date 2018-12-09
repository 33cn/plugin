/* $Id: suio.c 3769 2008-11-13 20:21:34Z max $ */

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

#include "suio.h"

#ifdef DMALLOC

const char *suiocheck_file;
int suiocheck_line;

struct memsum {
  struct iovec iov;
  u_int16_t sum;
  int line;
  const char *file;
};

/* Simple, IP-like checksum */
u_int16_t
cksum (void *_data, int len)
{
  u_char *data = (u_char *) _data;
  union {
    u_char b[2];
    u_int16_t w;
  } bwu;
  u_int32_t sum;
  u_char *end;

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

static void
memsum_check (void *_ms)
{
  struct memsum *ms = (struct memsum *) _ms;
  if (cksum (ms->iov.iov_base, ms->iov.iov_len) != ms->sum) {
    fprintf (stderr, "%s:%d: data in uio subesquently changed!\n",
	     ms->file, ms->line);
    abort ();
  }
  xfree (ms);
}

void
__suio_check (const char *file, int line, struct suio *uio,
	      const void *base, size_t len)
{
  if (uio->uio_iovcnt > 0) {
    struct memsum *ms = (struct memsum *) xmalloc (sizeof (*ms));
    ms->iov.iov_base = (iovbase_t) base;
    ms->iov.iov_len = len;
    ms->sum = cksum (ms->iov.iov_base, ms->iov.iov_len);
    if (suiocheck_file) {
      ms->file = suiocheck_file;
      ms->line = suiocheck_line;
    }
    else {
      ms->file = file;
      ms->line = line;
    }
    uio->uio_endp = NULL;
    suio_callback (uio, memsum_check, ms);
  }
}

#define xxmalloc xmalloc

#else /* !DMALLOC */

void *
xxmalloc (size_t n)
{
  void *r = malloc (n);
  if (!r) {
    const char msg[] = "malloc failed\n";
    int i = write (2, msg, sizeof (msg) - 1);
    i ++;
    abort ();
  }
  return r;
}

#endif /* DMALLOC */

static inline void
__suio_init (struct suio *uio)
{
  bzero (uio, offsetof (struct suio, uio_defbuf.b_data));
  uio->uio_mem = uio->uio_iov = uio->uio_defiov;
  uio->uio_buf1 = uio->uio_bufp = &uio->uio_defbuf;
  uio->uio_dp = uio->uio_defbuf.b_data;
  uio->uio_dspace = SBUFSIZ;
  uio->uio_iovmax = NDEFIOV;
  TAILQ_INIT (&uio->uio_cb);
}

void
suio_construct (struct suio *uio)
{
  __suio_init (uio);
}

static inline void
__suio_freemem (struct suio *uio)
{
  struct s_buf *b, *n;
  struct iovcb *cb1, *cb2;
  for (cb1 = uio->uio_cb.tqh_first; cb1; cb1 = cb2) {
    cb2 = cb1->cb_link.tqe_next;
    cb1->cb_fn (cb1->cb_arg);
    xfree (cb1);
  }
  for (b = uio->uio_buf1; b; b = n) {
    n = b->b_next;
    if (b != &uio->uio_defbuf)
      xfree (b);
  }
  if (uio->uio_memalloced)
    xfree (uio->uio_mem);
}

void
suio_destruct (struct suio *uio)
{
  __suio_freemem (uio);
}

static inline void
__suio_bfree (struct suio *uio, struct s_buf *b)
{
    if (b == &uio->uio_defbuf)
      uio->uio_defbfree = 1;
    else
      xfree (b);
}

struct suio *
#ifndef DMALLOC
suio_alloc (void)
#else /* DMALLOC */
__suio_alloc (const char *file, int line)
#endif /* DMALLOC */
{
  struct suio *uio;
#ifndef DMALLOC
  uio = (struct suio *) xxmalloc (sizeof (*uio));
#else /* DMALLOC */
  uio = (struct suio *) _xmalloc_leap (file, line, sizeof (*uio));
#endif /* DMALLOC */
  __suio_init (uio);
  return (uio);
}

void
suio_free (struct suio *uio)
{
  __suio_freemem (uio);
  xfree (uio);
}

void
suio_rembytes (struct suio *uio, u_int nbytes)
{
  u_int64_t niov, iovn;
  struct s_buf *b;
  struct iovcb *cb1;

  assert (nbytes <= uio->uio_resid);

#if 0
  /* Blow away everything if we are asked to remove all bytes */
  if (nbytes == uio->uio_resid) {
    niov = uio->uio_nremiov + uio->uio_iovcnt;
    __suio_freemem (uio);
    __suio_init (uio);
    uio->uio_nremiov = niov;
    return;
  }

  /* Actually, the above is potentially race-prone of a bad idea in
   * the presence of arbirary callbacks. */
#endif

  /* Adjust byte count */
  uio->uio_resid -= nbytes;

  if (uio->uio_resid) {
    /* Count IOV's that are consumed in their entirety, with nbytes of
     * last IOV left over */
    for (niov = 0; nbytes >= uio->uio_iov[niov].iov_len; niov++)
      nbytes -= uio->uio_iov[niov].iov_len;

    /* Adjust IOV pointers and counts in last IOV */
    uio->uio_nremiov += niov;
    uio->uio_iov += niov;
    uio->uio_iovcnt -= niov;
    uio->uio_iovmax -= niov;
    if (nbytes) {
      uio->uio_iov[0].iov_len -= nbytes;
      uio->uio_iov[0].iov_base =
	(iovbase_t) ((char *) uio->uio_iov[0].iov_base + nbytes);
    }
  }
  else {
    /* All IOV's have been consumed, so we might as well start over at
     * uio_mem. */
    niov = uio->uio_iovcnt;
    uio->uio_nremiov += niov;
    uio->uio_iovcnt = 0;
    uio->uio_iovmax += uio->uio_mem - uio->uio_iov;
    uio->uio_iov = uio->uio_mem;
    /* No more last IOV to append onto. */
    uio->uio_endp = NULL;
  }

  /* Free any dynamically allocated s_buf's no longer needed */
  b = uio->uio_buf1;
  while (b->b_next && (u_int) b->b_uiono <= niov) {
    struct s_buf *n;
    niov -= b->b_uiono;
    n = b->b_next;
    __suio_bfree (uio, b);
    b = n;
  }
  uio->uio_buf1 = b;

  /* make any necessary callbaks */
  iovn = uio->uio_nremiov;
  while ((cb1 = uio->uio_cb.tqh_first) && iovn >= cb1->cb_niov) {
    TAILQ_REMOVE (&uio->uio_cb, cb1, cb_link);
    cb1->cb_fn (cb1->cb_arg);
    xfree (cb1);
  }
}

#ifndef DMALLOC
char *
suio_flatten (const struct suio *uio)
#else /* DMALLOC */
char *
__suio_flatten (const struct suio *uio, const char *file, int line)
#endif /* DMALLOC */
{
  struct iovec *iovp = uio->uio_iov;
  struct iovec *lastiov = uio->uio_iov + uio->uio_iovcnt;
#ifndef DMALLOC
  char *buf = (char *) xxmalloc (uio->uio_resid);
#else /* DMALLOC */
  char *buf = (char *) _xmalloc_leap (file, line, uio->uio_resid);
#endif /* DMALLOC */
  char *cp = buf;

  while (iovp < lastiov) {
    int n = iovp->iov_len;
    memcpy (cp, iovp->iov_base, n);
    cp += n;
    iovp++;
  }
  return buf;
}

char *
__suio_newbuf (struct suio *uio)
{
  struct s_buf *b;

  if (uio->uio_defbfree) {
    b = &uio->uio_defbuf;
    uio->uio_defbfree = 0;
  }
  else
    b = (struct s_buf *) xxmalloc (sizeof (*b));
  uio->uio_bufp->b_next = b;
  uio->uio_bufp = b;
  uio->uio_dp = b->b_data;
  uio->uio_dspace = SBUFSIZ;
  b->b_next = NULL;
  b->b_uiono = 0;
  return (b->b_data);
}

void
__suio_newiov (struct suio *uio)
{
  u_int n;
  struct iovec *newiov;

  n = uio->uio_iov - uio->uio_mem;  /* Wasted slots below uio_iov */
  if (n >= uio->uio_iovmax) {
    memmove (uio->uio_mem, uio->uio_iov,
	     uio->uio_iovcnt * sizeof (struct iovec));
    uio->uio_iov = uio->uio_mem;
    uio->uio_iovmax += n;
    return;
  }
  n = 2 * (n + uio->uio_iovmax);  /* Allocate twice the number of old slots */
  newiov = (struct iovec *) xxmalloc (n * sizeof (struct iovec));
  memcpy (newiov, uio->uio_iov, uio->uio_iovcnt * sizeof (struct iovec));
  if (uio->uio_memalloced)
    xfree (uio->uio_mem);
  uio->uio_iov = uio->uio_mem = newiov;
  uio->uio_memalloced = 1;
  uio->uio_iovmax = n;
}

void
__suio_copy (struct suio *uio, const char *data, u_int len)
{
  char *dp;
  u_int n;

  if (len > SBUFSIZ) {
    dp = (char *) xxmalloc (len);
    memcpy (dp, data, len);
    __suio_addiov (uio, dp, len);
    suio_callfree (uio, dp);
    return;
  }

  dp = uio->uio_dp;
  n = len < (u_int) uio->uio_dspace ? len : (u_int) uio->uio_dspace;
  memcpy (dp, data, n);
  __suio_addiov (uio, dp, n);
 
  while (len - n >= SBUFSIZ) {
    dp = __suio_newbuf (uio);
    memcpy (dp, data + n, SBUFSIZ);
    __suio_addiov (uio, dp, SBUFSIZ);
    n += SBUFSIZ;
  }

  if (len > n) {
    dp = __suio_newbuf (uio);
    memcpy (dp, data + n, len - n);
    __suio_addiov (uio, dp, len - n);
  }

  uio->uio_dp = uio->uio_endp;
  uio->uio_dspace = SBUFSIZ - (uio->uio_dp - uio->uio_bufp->b_data);
}

void
__suio_fill (struct suio *uio, char c, u_int n)
{
  char *base;

  base = __suio_newbuf (uio);
  memset (base, c, SBUFSIZ);
  while (n > SBUFSIZ) {
    suio_print (uio, base, SBUFSIZ);
    n -= SBUFSIZ;
  }
  if (n > 0)
    suio_print (uio, base, n);
}

void
suio_cat (struct suio *dst, const struct suio *src)
{
  u_int i;

  for (i = 0; i < src->uio_iovcnt; i++) {
    if (dst->uio_iovcnt >= dst->uio_iovmax)
      __suio_newiov (dst);
    dst->uio_resid += src->uio_iov[i].iov_len;
    dst->uio_iov[dst->uio_iovcnt++] = src->uio_iov[i];
  }
}

void
suio_move (struct suio *dst, struct suio *src)
{
  struct iovcb *cb, *ocb;

  cb = src->uio_cb.tqh_first;
  TAILQ_INIT (&src->uio_cb);

  while (src->uio_iovcnt) {
    if ((char *) src->uio_iov->iov_base >= src->uio_buf1->b_data
	&& (char *) src->uio_iov->iov_base < src->uio_buf1->b_data + SBUFSIZ)
      suio_copy (dst, src->uio_iov->iov_base, src->uio_iov->iov_len);
    else
      __suio_addiov (dst, src->uio_iov->iov_base, src->uio_iov->iov_len);
    suio_rembytes (src, src->uio_iov->iov_len);

    while (cb && src->uio_nremiov >= cb->cb_niov) {
      ocb = cb;
      cb = cb->cb_link.tqe_next;
      ocb->cb_niov = dst->uio_nremiov + dst->uio_iovcnt;
      TAILQ_INSERT_TAIL (&dst->uio_cb, ocb, cb_link);
    }
  }

  assert (!cb);
}

void
suio_copyv (struct suio *uio, const struct iovec *iov, int cnt, u_int skip)
{
  u_int size = iovsize (iov, cnt);
  u_int n;
  const struct iovec *iovp;
  char *buf;
  char *cp;

  assert (skip <= size);
  size -= skip;

  n = size;
  iovp = iov + cnt;

  if (!size)
    return;
  buf = (char *) xxmalloc (size);
  cp = buf + size;

  while (iovp-- > iov && iovp->iov_len <= n) {
    n -= iovp->iov_len;
    cp -= iovp->iov_len;
    memcpy (cp, iovp->iov_base, iovp->iov_len);
  }
  assert (cp - buf == (int) n && n <= size);
  if (n > 0) {
    assert (iovp > iov);
    memcpy (buf, (char *) iovp->iov_base + iovp->iov_len - n, n);
  }

#if 0
  {
    size_t nc = 0;
    for (int i = 0; i < cnt; i++) {
      u_char *p = (u_char *) iov[i].iov_base;
      u_char *e = p + iov[i].iov_len;
      while (p < e) {
	if (nc > size + skip)
	  panic ("suio_copyv: calculated size was wrong\n");
	if (nc >= skip && buf[nc - skip] != *p)
	  panic ("suio_copyv: wrong byte\n");
	nc++;
	p++;
      }
    }
  }
#endif

  suio_print (uio, buf, size);
  suio_callfree (uio, buf);
}

#undef MIN3
#undef MIN2
#define MIN3(a, b, c) \
  ((a) < (b) ? ((a) < (c) ? (a) : (c)) : ((b) < (c) ? (b) : (c)))
#define MIN2(a, b) ((a) < (b) ? (a) : (b))

#ifndef _KERNEL

int
suio_output (struct suio *uio, int fd, int cnt)
{
  ssize_t n = 0;
  if (cnt < 0)
    while (uio->uio_resid
	   && (n = writev (fd, uio->uio_iov,
			   MIN2 (uio->uio_iovcnt, UIO_MAXIOV))) > 0)
      suio_rembytes (uio, n);
  else {
    u_int64_t maxiovno = uio->uio_nremiov + cnt;
    while (uio->uio_resid && uio->uio_nremiov < maxiovno
	   && (n = writev (fd, uio->uio_iov,
			   MIN2((int) (maxiovno - uio->uio_nremiov),
				UIO_MAXIOV))) > 0)
      suio_rembytes (uio, n);
  }
  if (n > 0)
    return 1;
  if (n == 0 || errno == EAGAIN)
    return 0;
  return -1;
}

#endif /* !_KERNEL */

size_t
iovsize (const struct iovec *iov, int cnt)
{
  const struct iovec *end;
  size_t size = 0;
  for (end = iov + cnt; iov < end; iov++)
    size += iov->iov_len;
  return size;
}

#ifdef __cplusplus

static inline void
__makecbv (void *_cb)
{
  callback<void>::ref &cb = *static_cast<callback<void>::ref *> (_cb);
  (*cb) ();
  delete &cb;
}

#endif /* __cplusplus */
