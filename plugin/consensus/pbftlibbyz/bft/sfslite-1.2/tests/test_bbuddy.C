/* $Id: test_bbuddy.C 2 2003-09-24 14:35:33Z max $ */

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

#include "async.h"
#include "arc4.h"
#include "bbuddy.h"

bbuddy bb (0x100000, 0x800, 0x10000);

int
main (int argc, char **argv)
{
  setprogname (argv[0]);

  vec<off_t> ov;
  vec<size_t> lv;
  bb._check ();

  arc4 as;
  char key[] = "will you be my binary buddy?";
  as.setkey (key, sizeof (key));

  for (int i = 0; i < 256; i++) {
    bb.settotsize (bb.gettotsize () + 0x800);
    size_t l = ((as.getbyte () << 8 | as.getbyte ()) % 0x800);
    off_t o = bb.alloc (l);
    if (o < 0)
      panic ("could not allocate 0x%qx bytes with 0x%qx free\n",
	     (u_int64_t) l, (u_int64_t) bb.space ());
    ov.push_back (o);
    lv.push_back (l);
    bb._check ();
  }

  for (int i = 0; i < 256; i++) {
    bb.dealloc (ov.pop_front (), lv.pop_front ());
    bb._check ();
  }

  return 0;
}
