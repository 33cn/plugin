/* $Id: pipe2str.C 1117 2005-11-01 16:20:39Z max $ */

/*
 *
 * Copyright (C) 2002 David Mazieres (dm@uun.org)
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

void
pipe2str (int fd, cbs cb, int *fdp, strbuf *sb)
{
  if (!sb) {
    sb = New strbuf ();
    make_async (fd);
    fdcb (fd, selread, wrap (pipe2str, fd, cb, fdp, sb));
  }

  int n;
  if (fdp && *fdp < 0) {
    char *buf = sb->tosuio ()->getspace (8192);
    n = readfd (fd, buf, 8192, fdp);
    if (n > 0)
      sb->tosuio ()->print (buf, n);
  }
  else
    n = sb->tosuio ()->input (fd);

  if (!n)
    (*cb) (*sb);
  else if (n < 0 && errno != EAGAIN)
    (*cb) (NULL);
  else
    return;
  fdcb (fd, selread, NULL);
  close (fd);
  delete sb;
}

void
chldrun (cbi chld, cbs cb)
{
  int fds[2];
  if (pipe (fds) < 0)
    (*cb) (NULL);
  switch (afork ()) {
  case -1:
    (*cb) (NULL);
    return;
  case 0:
    close (fds[0]);
    (*chld) (fds[1]);
    _exit (0);
  }
  close (fds[1]);
  pipe2str (fds[0], cb);
}

