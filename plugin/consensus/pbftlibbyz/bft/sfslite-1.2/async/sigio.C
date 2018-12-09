/* $Id: sigio.C 1117 2005-11-01 16:20:39Z max $ */

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

#include "amisc.h"

#ifdef HAVE_SYS_STROPTS_H
# define strbuf streams_strbuf
# include <sys/stropts.h>
# undef strbuf
#endif /* HAVE_SYS_STROPTS_H */

int
sigio_set (int fd)
{
  int flag;
#if defined (O_ASYNC) && defined (F_SETOWN)
  if (fcntl (fd, F_SETOWN, getpid ()) == -1) {
    warn ("sigio_set: F_SETOWN: %m\n");
    return -1;
  }
  flag = fcntl (fd, F_GETFL, 0);
  if (flag == -1) {
    warn ("sigio_set: F_GETFL: %m\n");
    return -1;
  }
  flag |= O_ASYNC;
  if (fcntl (fd, F_SETFL, flag) == -1) {
    warn ("sigio_set: F_SETFL: %m\n");
    return -1;
  }
#elif defined (I_SETSIG)
  flag = S_INPUT | S_OUTPUT;
  if (ioctl (fd, I_SETSIG, flag) < 0) {
    warn ("sigio_set: I_SETSIG: %m\n");
    return -1;
  }
#elif defined (SIOCSPGRP) && defined (FIOASYNC)
  flag = getpgrp ();
  if (ioctl (fd, SIOCSPGRP, &flag) < 0) {
    warn ("sigio_set: SIOCSPGRG: %m\n");
    return -1;
  }
  flag = 1;
  if (ioctl (fd, FIOASYNC, &flag) < 0) {
    warn ("sigio_set: FIOASYNC: %m\n");
    return -1;
  }
#else
# error "Don't know how to enable SIGIO on sockets"
#endif
  return 0;
}

int
sigio_clear (int fd)
{
  int flag;
#if defined (O_ASYNC) && defined (F_SETOWN)
  flag = fcntl (fd, F_GETFL, 0);
  if (flag == -1) {
    warn ("sigio_clear: F_GETFL: %m\n");
    return -1;
  }
  flag &= ~O_ASYNC;
  if (fcntl (fd, F_SETFL, flag) == -1) {
    warn ("sigio_clear: F_SETFL: %m\n");
    return -1;
  }
#elif defined (I_SETSIG)
  flag = 0;
  if (ioctl (fd, I_SETSIG, flag) < 0) {
    warn ("sigio_clear: I_SETSIG: %m\n");
    return -1;
  }
#elif defined (SIOCSPGRP) && defined (FIOASYNC)
  flag = 0;
  if (ioctl (fd, FIOASYNC, &flag) < 0) {
    warn ("sigio_clear: FIOASYNC: %m\n");
    return -1;
  }
#else
# error "Don't know how to enable SIGIO on sockets"
#endif
  return 0;
}
