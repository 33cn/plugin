/* $Id: test_mpz_square.C 2 2003-09-24 14:35:33Z max $ */

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


#include "crypt.h"

int
main (int argc, char **argv)
{
  random_update ();

  bigint r, s1, s2;

  for (int i = 0; i < 1024; i++) {
    r = random_bigint (rnd.getword () % 2048);
    s1 = r * r;
    mpz_square (&s2, &r);
    if (s1 != s2)
      panic << "r = " << r << "\n"
	    << "     " << s1 << "\n  != " << s2 << "\n"
	    << "    ["
	    << strbuf ("%*s", mpz_sizeinbase (&s1, 16),
		       bigint (abs (s1 - s2)).cstr ())
	    << "]\n";
  }

  return 0;
}
