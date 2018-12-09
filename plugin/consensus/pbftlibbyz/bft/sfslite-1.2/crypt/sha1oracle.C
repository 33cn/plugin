/* $Id: sha1oracle.C 1117 2005-11-01 16:20:39Z max $ */

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

/*
 * Sha1oracle provides a set of one-way functions that can be used in
 * some places as a substitute for a theoretical random oracle.
 *
 * The constructor takes two arguments:  The result size, and a 64-bit
 * integer designating which one way function to use from the set.
 *
 * Let || denote concatenation.
 * Let <n> designate the 64-bit big-endian representation of number n.
 * Let SHA1_64(x) designate the first 64 bits of SHA1(x).
 *
 * Sha1oracle(size, index), when fed message M, outputs the first size
 * bytes of the infinite sequence:
 *
 *   SHA1_64(<0> || <index> || M) || SHA1_64(<1> || <index> || M)
 *     || SHA1_64(<2> || <index> || M) ...
 */

#include "sha1.h"
#include "opnew.h"

sha1oracle::sha1oracle (size_t nbytes, u_int64_t idx, size_t used)
  : hashused (used), nctx ((nbytes + hashused - 1) / hashused),
    state (New u_int32_t[nctx][hashwords]),
    idx (htonq (idx)), resultsize (nbytes)
{
  reset ();
}

sha1oracle::~sha1oracle ()
{
  bzero (state, nctx * sizeof (*state));
  delete[] state;
}

void
sha1oracle::reset ()
{
  u_int64_t ini[2] = { 0, idx };
  count = 0;
  for (size_t i = 0; i < nctx; i++)
    newstate (state[i]);
  firstblock = true;
  update (ini, sizeof (ini));
}

void
sha1oracle::consume (const u_char *p)
{
  if (!firstblock) {
    for (size_t i = 0; i < nctx; i++)
      transform (state[i], p);
    return;
  }

  firstblock = false;
  assert (p == buffer);
  for (size_t i = 0; i < nctx; i++) {
    *reinterpret_cast<u_int64_t *> (buffer) = htonq (i);
    transform (state[i], p);
  }
}

void
sha1oracle::final (void *_p)
{
  u_char *p = static_cast<u_char *> (_p);
  u_int32_t (*sp)[hashwords] = state;
  u_char *end = p + resultsize;
  u_char buf[hashsize];

  finish ();
  for (; p + hashsize <= end; p += hashused)
    state2bytes (p, *sp++);
  if (p + hashused <= end) {
    state2bytes (buf, *sp++);
    memcpy (p, buf, hashused);
    p += hashused;
  }
  if (p < end) {
    state2bytes (buf, *sp);
    memcpy (p, buf, end - p);
  }
}
