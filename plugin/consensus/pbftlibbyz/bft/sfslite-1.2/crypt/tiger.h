// -*-c++-*-
/* $Id: tiger.h 1117 2005-11-01 16:20:39Z max $ */

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

#ifndef _CRYPT_TIGER_H_INCLUDED_
#define _CRYPT_TIGER_H_INCLUDED_ 1

#include "crypthash.h"

class tiger : public mdblock {
protected:
  static const u_int64_t t1[256];
  static const u_int64_t t2[256];
  static const u_int64_t t3[256];
  static const u_int64_t t4[256];
  
public:
  enum { hashquads = 3 };
  enum { hashsize = hashquads * 8 };

  void finish () { mdblock::finish_le (); }

  static void newstate (u_int64_t[hashquads]);
  static void transform (u_int64_t[hashquads], const u_char[blocksize]);
  static void state2bytes (void *, const u_int64_t[hashquads]);
};

class tigerctx : public tiger {
  u_int64_t state[3];

  void consume (const u_char *p) { transform (state, p); }
public:
  tigerctx () { newstate (state); }
  void reset () { count = 0; newstate (state); }
  void final (void *digest) {
    finish ();
    state2bytes (digest, state);
    bzero (state, sizeof (state));
  }
};

#endif /* !_CRYPT_TIGER_H_INCLUDED_ */
