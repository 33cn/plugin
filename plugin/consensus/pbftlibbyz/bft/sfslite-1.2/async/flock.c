/* $Id: flock.c 1117 2005-11-01 16:20:39Z max $ */

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

#include "sysconf.h"

#ifndef HAVE_FLOCK

/* This is a substitute definition for flock in terms of the POSIX
 * fcntl file locking.  Note that the semantics are not the same.  In
 * particular, POSIX fcntl locking has very stupid semantics where
 * closing any file descriptor for a file relinquishes all locks.
 * Thus, unrelated parts of your program can unintentionally nuke your
 * locks.
 */
int
flock (int fd, int options)
{
  static const char optok[16] = {
    0, 1, 1, 0,
    0, 1, 1, 0,
    1, 0, 0, 0,
    0, 0, 0, 0,
  };
  int ret;
  struct flock lf;

  if ((u_int) options > 8 || !optok[options]) {
    errno = EINVAL;
    return -1;
  }

  bzero (&lf, sizeof (lf));
  if (options & LOCK_UN) {
    lf.l_type = F_UNLCK;
    ret = fcntl (fd, F_SETLK, &lf);
  }
  else {
    int cmd = (options & LOCK_NB) ? F_SETLK : F_SETLKW;
    lf.l_type = (options & LOCK_EX) ? F_WRLCK : F_RDLCK;
    ret = fcntl (fd, cmd, &lf);
  }

  return -(ret == -1);
}

#endif /* !HAVE_FLOCK */
