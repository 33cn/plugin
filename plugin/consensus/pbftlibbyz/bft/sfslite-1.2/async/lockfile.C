/* $Id: lockfile.C 1117 2005-11-01 16:20:39Z max $ */

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

#include "amisc.h"

bool
stat_unchanged (const struct stat *sb1, const struct stat *sb2)
{
  return sb1->st_dev == sb2->st_dev && sb1->st_ino == sb2->st_ino
    && sb1->st_mtime == sb2->st_mtime
#ifdef SFS_HAVE_STAT_ST_MTIMESPEC
    && sb1->st_mtimespec.tv_nsec == sb2->st_mtimespec.tv_nsec
#endif /* SFS_HAVE_STAT_ST_MTIMESPEC */
    && sb1->st_size == sb2->st_size;
}

static bool
checkstat (const str &path, const struct stat &sb)
{
  if (!S_ISREG (sb.st_mode))
    warn << path << ": not a regular file -- please delete\n";
  else if (sb.st_nlink > 1)
    warn << path << ": too many links -- please delete\n";
  else if (sb.st_mode & 07177)
    warn ("%s: mode 0%o should be 0600 -- please delete\n",
	  path.cstr (), int (sb.st_mode & 07777));
  else if (sb.st_size)
    warn << path << ": file should be empty -- please delete\n";
  else
    return true;
  return false;
}

lockfile::~lockfile ()
{
  if (fdok () && (islocked || acquire (false)))
    unlink (path);
  closeit ();
}

bool
lockfile::openit ()
{
  struct stat sb;

  if (fd >= 0)
    closeit ();

  errno = 0;
  if (lstat (path, &sb) >= 0 && !checkstat (path, sb))
    return false;
  else if (errno && errno != ENOENT) {
    warn << path << ": " << strerror (errno) << "\n";
    return false;
  }

  /* N.B., 0600 (not 0644) to stop others from wedging us with
   * LOCK_SH.  Anyone with read permission to a file can flock it
   * with LOCK_SH, thereby blocking those with write permission who
   * want to lock it with LOCK_EX. */
  fd = open (path, O_RDWR|O_CREAT, 0600);
  if (fd < 0) {
    warn << path << ": " << strerror (errno) << "\n";
    return false;
  }
  close_on_exec (fd);
  errno = 0;
  if (fstat (fd, &sb) < 0 || !checkstat (path, sb)) {
    if (errno)
      warn << "fstat (" << path << "): " << strerror (errno) << "\n";
    closeit ();
    return false;
  }

  return true;
}

void
lockfile::closeit ()
{
  if (fd >= 0) {
    flock (fd, LOCK_UN);
    close (fd);
    fd = -1;
  }
  islocked = false;
}

bool
lockfile::acquire (bool wait)
{
  for (;;) {
    if (!fdok () && !openit ())
      return false;
    if (islocked)
      return true;

    if (flock (fd, LOCK_EX | (wait ? 0 : LOCK_NB)) < 0) {
      if (wait && errno == EINTR)
	continue;
      return false;
    }

    utimes (path, NULL);
    islocked = true;
  }
}

void
lockfile::release ()
{
  if (islocked) {
    flock (fd, LOCK_UN);
    islocked = false;
  }
}

bool
lockfile::fdok ()
{
  struct stat sb, fsb;
  if (fd < 0 || stat (path, &sb) < 0 || fstat (fd, &fsb) < 0
      || !stat_unchanged (&sb, &fsb)) {
    closeit ();
    return false;
  }
  return true;
}

ptr<lockfile>
lockfile::alloc (const str &path, bool wait)
{
  ref<lockfile> lf = New refcounted<lockfile> (path);
  if (lf->acquire (wait))
    return lf;
  return NULL;
}
