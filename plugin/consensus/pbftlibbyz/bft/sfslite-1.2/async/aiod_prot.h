// -*-c++-*-
/* $Id: aiod_prot.h 2881 2007-05-18 19:06:04Z max $ */

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

#ifndef _AIOD_AIOD_PROT_H_
#define _AIOD_AIOD_PROT_H_ 1

#include "sysconf.h"

typedef size_t aiomsg_t;

/* File handle type.  Arguments containing file handles contain an
 * aiomsg_t with the position of an aiod_file structure in the shared
 * memory buffer. */
struct aiod_file {
  int handle;
  int oflags;
  dev_t dev;
  ino_t ino;
  char path[1];
};

struct aiod_iobuf {
  off_t pos;			// Position in file
  aiomsg_t buf;			// Position in shared mem of buffer
  ssize_t len;			// Size of buffer/operation
};

enum aiod_op {
  AIOD_NOP = 0,

  /* Requests taking aiod_pathop */
  AIOD_UNLINK = 1,
  AIOD_LINK = 2,
  AIOD_SYMLINK = 3,
  AIOD_RENAME = 4,
  AIOD_READLINK = 5,
  AIOD_STAT = 6,
  AIOD_LSTAT = 7,
  AIOD_GETCWD = 8,

  /* Requests take an aiod_fhop */
  AIOD_OPEN = 9,
  AIOD_CLOSE = 10,
  AIOD_FSYNC = 11,
  AIOD_FTRUNC = 12,
  AIOD_READ = 13,
  AIOD_WRITE = 14,

  /* fstat is special */
  AIOD_FSTAT = 15,

  /*Requests take an aiod_fhop */
  AIOD_OPENDIR = 16,
  AIOD_READDIR = 17,
  AIOD_CLOSEDIR = 18,
};

/* Common header for all requests */
struct aiod_reqhdr {
  aiod_op op;
  int err;
};

struct aiod_nop {
  aiod_op op;
  int err;
  size_t nopsize;
};

/* Argument for operations that take pathnames */
struct aiod_pathop {
  aiod_op op;
  int err;
  size_t bufsize;
  union {
    off_t __alignment_hack;
    char pathbuf[2];		// 1 or 2 null-terminated paths in arg
  };

  char *path1 () { return pathbuf; }
  char *path2 () { return pathbuf + 1 + strlen (pathbuf); }
  struct stat *statbuf () {
#ifdef XXX_CHECK_BOUNDS		// Can't trust aiod's not to clobber bufsize
    assert (sizeof (struct stat) <= bufsize);
#endif /* CHECK_BOUNDS */
    return reinterpret_cast<struct stat *> (pathbuf);
  }
  void setpath (const char *p1, const char *p2 = "") {
    size_t len1 = strlen (p1);
#ifdef XXX_CHECK_BOUNDS
    assert (len1 + 2 + strlen (p2) <= bufsize);
#endif /* CHECK_BOUNDS */
    strcpy (pathbuf, p1);
    strcpy (pathbuf + len1 + 1, p2);
  }
  static size_t totsize (size_t bufsize);
};

inline size_t
aiod_pathop::totsize (size_t bufsize)
{
  return offsetof (aiod_pathop, pathbuf) + bufsize;
}

/* Argument for most operations that take a file handle */
struct aiod_fhop {
  aiod_op op;
  int err;
  aiomsg_t fh;			// Pointer to file handle
  union {
    mode_t mode;		// For open
    off_t length;		// For trunc
    aiod_iobuf iobuf;		// For read, write
  };
};

/* Argument for fstat */
struct aiod_fstat {
  aiod_op op;
  int err;
  aiomsg_t fh;			// Pointer to file handle
  struct stat statbuf;
};

#endif /* !_AIOD_AIOD_PROT_H_ */
