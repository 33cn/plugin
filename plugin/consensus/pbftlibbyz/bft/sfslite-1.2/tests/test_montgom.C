/* $Id: test_montgom.C 2 2003-09-24 14:35:33Z max $ */

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


#include "crypt.h"
#include "modalg.h"

int
main (int argc, char **argv)
{
  random_update ();

  bigint m, m2, r, r2, ri, s1, s2;
  montgom b;

  for (int i = 120; i < 162; i++) {
    int res = 0;
    m = random_bigint (i);
    m.setbit (0, 1);
    b.set (m);
    m2 = m * b.getr ();
    for (int j = i - 33; j <= 2 * i; j++) {
      r = random_zn (m2);
      r.trunc (j);
      s1 = mod (r * b.getri (), m);
      //s2 = b.mreduce (r);
      b.mpz_mreduce (&s2, &r);
      if (s1 != s2) {
	res |= 1;
	panic << "mreduce failed\n"
	      << " m = " << m << "\n"
	      << " r = " << r << "\n"
	      << "     " << s1 << "\n  != " << s2 << "\n"
	      << "    ["
	      << strbuf ("%*s", mpz_sizeinbase (&s1, 16),
			 bigint (abs (s1 - s2)).cstr ())
	      << "]\n";
      }
    }

    // r = s1;
    r = random_zn (m);
    r2 = random_zn (m);
    assert (r < m && r2 < m);

    s1 = mod (r * r2 * b.getri (), m);
    b.mpz_mmul (&s2, &r, &r2);
    if (s1 != s2) {
      res |= 2;
      panic << "mmul failed\n"
	    << " m = " << m << "\n"
	    << " r = " << r << "\n"
	    << "     " << s1 << "\n  != " << s2 << "\n"
	    << "    ["
	    << strbuf ("%*s", mpz_sizeinbase (&s1, 16),
		       bigint (abs (s1 - s2)).cstr ())
	    << "]\n";
    }

    s1 = powm (r, r2, m);
    b.mpz_powm (&s2, &r, &r2);
    if (s1 != s2) {
      res |= 4;
      panic << "powm failed\n"
	    << " m = " << m << "\n"
	    << " r = " << r << "\n"
	    << "     " << s1 << "\n  != " << s2 << "\n"
	    << "    ["
	    << strbuf ("%*s", mpz_sizeinbase (&s1, 16),
		       bigint (abs (s1 - s2)).cstr ())
	    << "]\n";
    }

#if 0
    warn ("%s mreduce.. %d\n", (res&1) ? "fail" : "ok", i);
    warn ("%s mmul.. %d\n", (res&2) ? "fail" : "ok", i);
    warn ("%s powm.. %d\n", (res&4) ? "fail" : "ok", i);
#endif
  }

  return 0;
}
