/* $Id: test_mpz_raw.C 2 2003-09-24 14:35:33Z max $ */

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

#include "bigint.h"
#include "async.h"

static inline void
memchk (const void *_cp, int c, size_t len)
{
  char *cp = (char *) _cp;
  char *ep = cp + len;
  while (cp < ep)
    if (*cp++ != (char) c)
      panic ("memchk: 0x%x should be 0x%x\n", cp[-1], c);
}

static void
getrnd (bigint &r, int limbs = 3)
{
  mpz_random (&r, limbs);
  mpz_umod_2exp (&r, &r, ((limbs - 1) * sizeof (mp_limb_t)
			  + (rand () % sizeof (mp_limb_t))) * 8);
}

static void
getrnd2 (bigint &r, int limbs = 3)
{
  mpz_random2 (&r, limbs);
  mpz_umod_2exp (&r, &r, ((limbs - 1) * sizeof (mp_limb_t)
			  + (rand () % sizeof (mp_limb_t))) * 8);
}

static void
test (const bigint &a)
{
  size_t size = mpz_rawsize (&a);
  char *p = New char[size + sizeof (mp_limb_t) + 1];
  bigint b;

  for (size_t i = 0; i < size; i++) {
    memset (p, 0xc9, size + sizeof (mp_limb_t) + 1);
    mpz_get_raw (p, i, &a);
    memchk (p + i, 0xc9, size + sizeof (mp_limb_t) - i);
  }
  for (size_t i = 0; i < sizeof (mp_limb_t) + 1; i++) {
    memset (p, 0xc9, size + sizeof (mp_limb_t) + 1);
    mpz_get_raw (p, size + i, &a);
    memchk (p + size + i, 0xc9, sizeof (mp_limb_t) - i);
    assert (p[size+i] == char (0xc9));
    mpz_set_raw (&b, p, size + i);
    _mpz_assert (&b);
    if (a != b) {
      strbuf x;
      for (size_t i = 0; i < sizeof (mp_limb_t) + 1; i++)
	x.fmt ("%02x", (u_char) p[i]);
      panic << a << " != " << b << "\n"
	    << "(raw: " << x << ")\n";
    }
  }

  if (a < 0)
    goto leave;

  for (size_t i = 0; i < size; i++) {
    memset (p, 0xc9, size + sizeof (mp_limb_t) + 1);
    mpz_get_rawmag_le (p, i, &a);
    memchk (p + i, 0xc9, size + sizeof (mp_limb_t) - i);
  }
  for (size_t i = 0; i < sizeof (mp_limb_t) + 1; i++) {
    memset (p, 0xc9, size + sizeof (mp_limb_t) + 1);
    mpz_get_rawmag_le (p, size + i, &a);
    memchk (p + size + i, 0xc9, sizeof (mp_limb_t) - i);
    assert (p[size+i] == char (0xc9));
    mpz_set_rawmag_le (&b, p, size + i);
    _mpz_assert (&b);
    if (a != b) {
      strbuf x;
      for (size_t i = 0; i < sizeof (mp_limb_t) + 1; i++)
	x.fmt ("%02x", (u_char) p[i]);
      panic << a << " != " << b << "\n"
	    << "(raw: " << x << ")\n";
    }
  }

 leave:
  delete[] p;
}

int
main (int argc, char **argv)
{
  int i;
  bigint n;

  setprogname (argv[0]);

  for (i = 0; i < 50; i++) {
    getrnd (n, 1); test (n); test (-n);
    getrnd (n, 2); test (n); test (-n);
    getrnd (n, 3); test (n); test (-n);
    getrnd2 (n); test (n); test (-n);
  }

  return 0;
}
