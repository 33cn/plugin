/* $Id: aerr.C 1117 2005-11-01 16:20:39Z max $ */

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
#include <dirent.h>

static suio *erruio = New suio;

static void _err_output_async (suio *, int);

static void
_err_reset_async ()
{
  erruio->clear ();
  fdcb (errfd, selwrite, NULL);
}

void
err_init ()
{
  int flags;

  erruio->clear ();
  if ((flags = fcntl (errfd, F_GETFL, 0)) != -1)
    fcntl (errfd, F_SETFL, flags|O_APPEND);
  _err_output = _err_output_async;
  _err_reset_hook = _err_reset_async;
}

void
err_flush ()
{
  if (_err_output != _err_output_async)
    return;
  make_sync (errfd);
  erruio->output (errfd);
}

static void
err_wcb ()
{
  int n;
  int cnt;

  if (!erruio->resid () || _err_output != _err_output_async) {
    fdcb (errfd, selwrite, NULL);
    return;
  }

  /* Try to write whole lines at a time. */
  for (cnt = min (erruio->iovcnt (), (size_t) UIO_MAXIOV);
       cnt > 0 && (erruio->iov ()[cnt-1].iov_len == 0
		   || *((char *) erruio->iov ()[cnt-1].iov_base
			+ erruio->iov ()[cnt-1].iov_len - 1) != '\n');
       cnt--)
    ;
  if (!cnt) {
    if (erruio->iovcnt () < UIO_MAXIOV) {
      /* Wait for a carriage return */
      fdcb (errfd, selwrite, NULL);
      return;
    }
    else
      cnt = -1;
  }

  /* Write asynchronously, but keep stderr synchronous in case of
   * emergency (e.g. maybe assert wants to fprintf to stderr). */
  if (globaldestruction)
    n = erruio->output (errfd, cnt);
  else {
    _make_async (errfd);
    n = erruio->output (errfd, cnt);
    make_sync (errfd);
  }

  if (n < 0)
    err_reset ();

  if (erruio->resid () && !globaldestruction)
    fdcb (errfd, selwrite, wrap (err_wcb)); 
  else
    fdcb (errfd, selwrite, NULL);
}

void
_err_output_async (suio *uio, int flags)
{
  int saved_errno = errno;
  
  if (flags & warnobj::panicflag) {
    erruio->copyu (uio);
    make_sync (errfd);
    erruio->output (errfd);
    myabort ();
  }

  /* Start new iovecs after newlines so as to output entire lines when
   * possible. */
  if (erruio->resid ()) {
    const iovec *iovp = erruio->iov () + erruio->iovcnt () - 1;
    if (((char *) iovp->iov_base)[iovp->iov_len - 1] == '\n')
      erruio->breakiov ();
  }
  erruio->copyu (uio);

  if (flags & warnobj::fatalflag) {
    err_flush ();
    exit (1);
  }

  err_wcb ();
  errno = saved_errno;
}

EXITFN (exitflush);
static void
exitflush ()
{
  if (_err_output != _err_output_async) {
    err_flush ();
    err_reset ();
  }
}
