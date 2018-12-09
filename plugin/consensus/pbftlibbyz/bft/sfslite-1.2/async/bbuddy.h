// -*-c++-*-
/* $Id: bbuddy.h 3148 2007-12-21 03:52:27Z max $ */

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

#ifndef _BBUDDY_H_
#define _BBUDDY_H_ 1

#include "sysconf.h"
#include "opnew.h"

extern "C" const char bytepop[];

class bbfree;

class bbuddy {
  typedef off_t totsize_t;

  totsize_t totsize;
  const u_int log2minalloc;
  const u_int log2maxalloc;
  bbfree *const freemaps;
  totsize_t spaceleft;

  bbfree &fm (size_t sn);
  bool _check_pos (u_int sn, size_t pos, bool set);
public:
  bbuddy (totsize_t totsize, size_t minalloc, size_t maxalloc);
  ~bbuddy ();

  off_t alloc (size_t);
  void dealloc (off_t, size_t);
  totsize_t space () const { return spaceleft; }
  totsize_t gettotsize () const { return totsize; }
  void settotsize (totsize_t totsize);

  size_t maxalloc () const { return 1 << log2maxalloc; }
  void _check ();
};

#endif /* _BBUDDY_H_ */
