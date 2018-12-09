// -*-c++-*-
/* $Id: test_bitvec.C 3256 2008-05-14 04:02:23Z max $ */

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

#include "bitvec.h"
#include "async.h"
#include "arc4.h"

#if 0 /* DMALLOC */
void
dmprt (void *p)
{
  size_t size;
  char *file;
  u_int line;
  void *ra;

  if (!dmalloc_examine (p, &size, &file, &line, &ra))
    warnx << file << ":" << line << ": " << size << " bytes\n";
  else
    warnx << "error\n";
}
#endif /* DMALLOC */

static void
test (bitvec &bv, int n)
{
  arc4 rs;
  rs.setkey (&n, sizeof (n));
  rs.setkey ("hello", 5);

  size_t size = rs.getbyte ();
  bv.setsize (size);
  for (u_int i = 0; i < size; i++)
    bv[i] = rs.getbyte () & 1;

  bitvec bv2 (bv);

  size_t e = size ? rs.getbyte () % size : 0;
  size_t s = e ? rs.getbyte () % e : 0;
  bool v = rs.getbyte () & 1;

  bv.setrange (s, e, v);

  for (u_int i = 0; i < size; i++) {
    if (bv[i] != ((i < s || i >= e) ? bv2[i] : v))
      panic << "i = " << i << ", s = " << s
	    << ", e = " << e << ", v = " << int (v) << "\n"
	    << bv << "\n" << bv2 << "\n";
  }
}

int
main (int argc, char **argv)
{
  bitvec tv;

  for (int i = 0; i < 1024; i++)
    test (tv, i);

  setprogname (argv[0]);
  return 0;
}
