/* $Id: suidgetfd.C 3170 2008-01-29 21:30:21Z max $ */

/*
 *
 * Copyright (C) 1999 David Mazieres (dm@uun.org)
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

#include "sfsmisc.h"
#include "sfsagent.h"
#include <grp.h>

static int
recvfd (int fd)
{
  int nfd = -1;
  char c;
  readfd (fd, &c, 1, &nfd);
  return nfd;
}

int
suidgetfd (str prog)
{
  int fds[2];
  if (socketpair (AF_UNIX, SOCK_STREAM, 0, fds) < 0)
    return -1;
  close_on_exec (fds[0]);

  str path = fix_exec_path ("suidconnect");
  const char *av[] = { "suidconnect", "-q", prog.cstr (), NULL };
  if (spawn (path, av, fds[1]) == -1) {
    close (fds[0]);
    close (fds[1]);
    return -1;
  }
  close (fds[1]);
  
  int fd = recvfd (fds[0]);
  close (fds[0]);
  return fd;
}

int
suidgetfd_required (str prog)
{
  int fds[2];
  if (socketpair (AF_UNIX, SOCK_STREAM, 0, fds) < 0)
    fatal ("socketpair: %m\n");
  close_on_exec (fds[0]);

  str path = fix_exec_path ("suidconnect");
  const char *av[] = { "suidconnect", prog.cstr (), NULL };
  if (spawn (path, av, fds[1]) == -1)
    fatal << path << ": " << strerror (errno) << "\n";
  close (fds[1]);
  
  int fd = recvfd (fds[0]);
  if (fd < 0) {
    struct stat sb;
    if (!runinplace && !stat (path, &sb)
	&& (sb.st_gid != sfs_gid || !(sb.st_mode & 02000))) {
      if (struct group *gr = getgrgid (sfs_gid))
	warn << path << " should be setgid to group " << gr->gr_name << "\n";
      else
	warn << path << " should be setgid to group " << sfs_gid << "\n";
    }
    else {
      warn << "have you launched the appropriate daemon (sfscd or sfssd)?\n";
      warn << "have subsidiary daemons died (check your system logs)?\n";
    }
    fatal ("could not suidconnect for %s\n", prog.cstr ());
  }
  close (fds[0]);
  return fd;
}
