// -*-c++-*-
/* $Id: str2file.C 3769 2008-11-13 20:21:34Z max $ */

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

int
mktmp_atomic (str tmpfile, int perm)
{
  int fd;
  struct stat sb1, sb2;

  for (;;) {
    for (;;) {
      if ((fd = open (tmpfile, O_CREAT|O_EXCL|O_WRONLY, perm)) >= 0
	  || errno != EEXIST)
	return fd;
      if (lstat (tmpfile, &sb1) >= 0) {
	if (!S_ISREG (sb1.st_mode)) {
	  errno = EEXIST;
	  return -1;
	}
	break;
      }
      if (errno != ENOENT)
	return -1;
    }

    for (;;) {
      if ((fd = open (tmpfile, O_CREAT|O_EXCL|O_WRONLY, perm)) >= 0
	  || errno != EEXIST)
	return fd;
      sleep (1);
      if (lstat (tmpfile, &sb2) < 0) {
	if (errno != ENOENT)
	  return -1;
	continue;
      }
      if (!S_ISREG (sb2.st_mode)) {
	errno = EEXIST;
	return -1;
      }
      if (sb1.st_dev != sb2.st_dev || sb1.st_ino != sb2.st_ino
	  || sb1.st_size != sb2.st_size)
	sb1 = sb2;
      else if (unlink (tmpfile) >= 0) {
	sleep (1);
	break;
      }
    }
  }
}

bool
str2file (str file, str s, int perm, bool excl, struct stat *sbp, bool binary)
{
  if (!file.len () || file.len () < strlen (file)) {
    errno = EINVAL;
    return false;
  }
  if (file[file.len () - 1] == '/') {
    errno = EISDIR;
    return false;
  }

  str tmpfile = file << "~";
  unlink (tmpfile);
  int fd;
  if (excl)
    fd = open (tmpfile, O_CREAT|O_EXCL|O_WRONLY, perm);
  else
    fd = mktmp_atomic (tmpfile, perm);
  if (fd < 0)
    return false;

  if (write (fd, s.cstr (), s.len ()) != int (s.len ())) {
    close (fd);
    unlink (tmpfile);
    return false;
  }
  if (s.len () && s[s.len () - 1] != '\n' && !binary)
    v_write (fd, "\n", 1);
  int err = fsync (fd);
  if (sbp && !err)
    err = fstat (fd, sbp);
  if (close (fd) < 0 || err < 0 || (excl && link (tmpfile, file) < 0)
      || (!excl && rename (tmpfile, file) < 0)) {
    unlink (tmpfile);
    return false;
  }
  if (excl)
    unlink (tmpfile);
  return true;
}

str
file2str (str file)
{
  int fd = open (file, O_RDONLY, 0);
  if (fd < 0)
    return NULL;

  struct stat sb;
  if (fstat (fd, &sb) < 0) {
    close (fd);
    return NULL;
  }
  if (!S_ISREG (sb.st_mode)) {
    warn << file << ": not a regular file\n";
    close (fd);
    errno = EINVAL;
    return NULL;
  }
  mstr m (sb.st_size);
  errno = EAGAIN;
  ssize_t n = read (fd, m, sb.st_size);
  int saved_errno = errno;
  close (fd);
  errno = saved_errno;
  if (n < 0)
    return NULL;
  m.setlen (n);
  return m;
}
