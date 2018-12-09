/* $Id: suio.h 435 2004-06-02 15:46:36Z max $ */

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


/* Handy-dandy all-purpose maintainer of iovec structures.  Handles
 * partial writev's.  Allocates memory for small writes.  Can even be
 * asked to free memory for you at some later point.
 */

#ifndef _ASYNC_SUIO_H_
#define _ASYNC_SUIO_H_ 1

#include "sfs-internal.h"
#include "queue.h"

#ifndef MALLOCRESV
#define MALLOCRESV 16
#endif /* !MALLOCRESV */

#ifdef SMALL_LIMITS
/* Small limits to help find reallocation bugs */
#define SBUFSIZ 32U
#define NDEFIOV 2
#define MINVEC 4
#else /* !SMALL_LIMITS */
#define SBUFSIZ (256 - sizeof (void *) - sizeof (int) - MALLOCRESV)
#define NDEFIOV ((512 - sizeof (struct s_buf) - 6 * sizeof (int) - 8 \
                 - 8 * sizeof (void *) - MALLOCRESV)/sizeof (struct iovec))
#define MINVEC 64
#endif /* !SMALL_LIMITS */

struct s_buf {
  struct s_buf *b_next;
  int b_uiono;
  char b_data[SBUFSIZ];
};

struct iovcb {
  u_int64_t cb_niov;		/* Call when uio_nremiov >= cb_niov */
  void (*cb_fn) (void *);	/* Function to call */
  void *cb_arg;			/* Argument */
  TAILQ_ENTRY (iovcb) cb_link;	/* Linked list */
};

struct suio {
  struct iovec *uio_mem;	/* iovec memory to be freed */
  struct iovec *uio_iov;	/* an iovec array suitable for writev */
  int uio_memalloced;		/* Set when uio_mem must be freed */
  u_int uio_iovcnt;		/* Number of used iovec structures */
  u_int uio_iovmax;		/* usable iov before calling newiov */
  u_int uio_resid;		/* Number of bytes in iovec structure */
  u_int64_t uio_nremiov;	/* Number of iov's removed by rembytes */
  TAILQ_HEAD (, iovcb) uio_cb;	/* Memory that needs to be freed */
  struct s_buf *uio_buf1;	/* Pointer to oldest used buffer */
  struct s_buf *uio_bufp;	/* Pointer to current buffer */
  char *uio_dp;			/* Pointer to scratch data in buffers */
  char *uio_endp;		/* End of last iov data */
  int uio_dspace;		/* Space in current scratch buffer */
  int uio_defbfree;		/* Set when uio_defbuf is free */
  /* Some initial space to malloc with a suio */
  struct s_buf uio_defbuf;	/* First buffer for converted numbers */
  struct iovec uio_defiov[NDEFIOV]; /* If fewer than NDEFIOV iovecs */
};
typedef struct suio suio;

extern size_t iovsize (const struct iovec *, int);

extern void suio_free (struct suio *);
extern void suio_rembytes (struct suio *, u_int);
extern void suio_construct (struct suio *);
extern void suio_destruct (struct suio *);
#ifndef DMALLOC
extern struct suio *suio_alloc (void);
extern char *suio_flatten (const struct suio *);
#define __suio_check(f, l, u)
#define __suio_printcheck(f, l, u, d, s) suio_print (u, d, s)
#define suio_printcheck suio_print
#define __suio_setcheckinfo(f, l) ((void) 0)
#else /* DMALLOC */
extern const char *suiocheck_file;
extern int suiocheck_line;
extern struct suio *__suio_alloc (const char *, int);
#define suio_alloc() __suio_alloc (__FILE__, __LINE__)
extern char *__suio_flatten (const struct suio *, const char *, int);
#define suio_flatten(uio) \
	__suio_flatten(uio, __FILE__, __LINE__)
extern void __suio_check (const char *, int, struct suio *,
			  const void *, size_t);
#define __suio_printcheck(f, l, u, d, s)			\
do {								\
  if (s > 0) {							\
    suio_print (u, d, s);					\
    if ((u)->uio_iov[(u)->uio_iovcnt - 1].iov_base == (d))	\
      __suio_check (f, l, u, d, s);				\
  }								\
  suiocheck_file = NULL;					\
} while (0)
#define suio_printcheck(u, d, s) \
	__suio_printcheck (__FILE__, __LINE__, u, d, s)
static inline void
__suio_setcheckinfo (const char *file, int line)
{
  if (!suiocheck_file) {
    suiocheck_file = file;
    suiocheck_line = line;
  }
}
#endif /* DMALLOC */
#ifndef DSPRINTF_DEBUG
/* The uprintf functions build up uio structures based on format
 * strings.  String ("%s") arguments are NOT copied, so you must not
 * modify any strings passed in.  Also, a format string of '%m'
 * doesn't convert any arguments but is equivalent to '%s' with an
 * argument of strerror (errno).  (This is like syslog.)  */
extern void suio_vuprintf (struct suio *, const char *, va_list);
extern void suio_uprintf (struct suio *, const char *, ...)
     __attribute__ ((format (printf, 2, 3)));
#else /* DSPRINTF_DEBUG */
extern void __suio_vuprintf (const char *, int,
			     struct suio *, const char *, va_list);
#define suio_vuprintf(uio, fmt, ap) \
	__suio_vuprintf (__FILE__, __LINE__, uio, fmt, ap)
extern void __suio_uprintf (const char *, int,
			    struct suio *, const char *, ...)
     __attribute__ ((format (printf, 4, 5)));
#define suio_uprintf(uio, fmt, args...) \
     __suio_uprintf (__FILE__, __LINE__, uio, fmt , ## args)
#endif /* DSPRINTF_DEBUG */
extern char *__suio_newbuf (struct suio *);
extern void __suio_newiov (struct suio *);
extern void __suio_fill (struct suio *, char, u_int);
extern void __suio_copy (struct suio *, const char *, u_int);
extern void suio_cat (struct suio *, const struct suio *);
extern void suio_move (struct suio *, struct suio *);
extern void suio_copyv (struct suio *, const struct iovec *, int, u_int);
extern int suio_output (struct suio *, int, int);
#ifndef DMALLOC
extern void *xxmalloc (size_t);
#else /* DMALLOC */
# define xxmalloc xmalloc
#endif /* DMALLOC */

static inline void
__suio_addiov (struct suio *uio, const void *data, u_int len)
{
  if (len == 0)
    return;
  if ((const char *) data == uio->uio_endp)
    uio->uio_iov[uio->uio_iovcnt-1].iov_len += len;
  else {
    struct iovec *iov;
    if (uio->uio_iovcnt >= uio->uio_iovmax)
      __suio_newiov (uio);
    iov = &uio->uio_iov[uio->uio_iovcnt++];
    iov->iov_base = (iovbase_t) data;
    iov->iov_len = len;
    uio->uio_bufp->b_uiono++;
  }
  uio->uio_resid += len;
  uio->uio_endp = (char *) data + len;
}

static inline char *
suio_getspace (struct suio *uio, int space)
{
  if (space <= uio->uio_dspace)
    return (uio->uio_dp);
  assert ((unsigned) space <= SBUFSIZ);
  return (__suio_newbuf (uio));
}

static inline char *
suio_getspace_aligned (struct suio *uio, int space)
{
  int pad = (4UL - (u_long) uio->uio_dp) & 3;
  if (uio->uio_dspace < pad + space)
    __suio_newbuf (uio);
  else {
    uio->uio_dspace -= pad;
    uio->uio_dp += pad;
  }
  return suio_getspace (uio, space);
}

static inline void
suio_copy (struct suio *uio, const void *data, int len)
{
  char *dp;

  if (len > uio->uio_dspace) {
    __suio_copy (uio, (const char *) data, len);
    return;
  }
  dp = uio->uio_dp;
  memcpy (dp, data, len);
  __suio_addiov (uio, dp, len);
  uio->uio_dspace -= len;
  uio->uio_dp = dp + len;
}

static inline void
suio_print (struct suio *uio, const void *data, int len)
{
  if (len < MINVEC && uio->uio_dp != data)
    suio_copy (uio, data, len);
  else {
    __suio_addiov (uio, data, len);
    if (uio->uio_dp == (const char *) data) {
      uio->uio_dspace -= len;
      uio->uio_dp += len;
    }
  }
}

static inline void
suio_fill (struct suio *uio, char c, int n)
{
  if (n <= uio->uio_dspace) {
    if (n > 0) {
      memset (uio->uio_dp, c, n);
      suio_print (uio, uio->uio_dp, n);
    }
  }
  else
    __suio_fill (uio, c, n);
}

static inline void
suio_clear (struct suio *uio)
{
  suio_rembytes (uio, uio->uio_resid);
}

static inline void
suio_callback (struct suio *uio, void (*fn) (void *), void *arg)
{
  struct iovcb *cb = (struct iovcb *) xxmalloc (sizeof (*cb));
  cb->cb_niov = uio->uio_nremiov + uio->uio_iovcnt;
  cb->cb_fn = fn;
  cb->cb_arg = arg;
  TAILQ_INSERT_TAIL (&uio->uio_cb, cb, cb_link);
}

/* Ask suio_rembytes to call free at some later point */
static inline void
suio_callfree (struct suio *uio, void *ptr)
{
  suio_callback (uio, xfree, ptr);
}

static inline void
suio_putc (struct suio *uio, char c)
{
  suio_copy (uio, &c, 1);
}

static inline void
suio_strcpy (struct suio *uio, const char *str)
{
  suio_copy (uio, str, strlen (str));
}

#ifdef __cplusplus
#include "callback.h"

extern "C" void __makecbv (void *);

inline void
suio_callback (suio *uio, callback<void>::ref cb)
{
  suio_callback (uio, __makecbv, New callback<void>::ref (cb));
}

struct Suio : suio {
  Suio () { suio_construct (this); }
  ~Suio () { suio_destruct (this); }
};
#endif /* __cplusplus */

#endif /*! _ASYNC_SUIO_H_ */
