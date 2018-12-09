/* $Id: test_mpz_xor.C 2 2003-09-24 14:35:33Z max $ */

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

void
dumb_xor (MP_INT *r, const MP_INT *a, const MP_INT *b)
{
  bigint t1;
  bigint t2;

  mpz_ior (&t1, a, b);
  mpz_and (&t2, a, b);
  mpz_com (&t2, &t2);
  mpz_and (r, &t1, &t2);
}

void
_test (const bigint &a, const bigint &b)
{
  bigint smart;
  bigint dumb;

  dumb_xor (&dumb, &a, &b);
  smart = a ^ b;
  _mpz_assert (&smart);
  if (smart != dumb)
    panic ("(r = a ^ b) %s ^ %s = %s\n\t(should be %s)\n",
	   a.cstr (), b.cstr (), smart.cstr (), dumb.cstr ());

  smart = a;
  smart ^= b;
  _mpz_assert (&smart);
  if (smart != dumb)
    panic ("(r ^= b) %s ^ %s = %s\n\t(should be %s)\n",
	   a.cstr (), b.cstr (), smart.cstr (), dumb.cstr ());

  smart = b;
  mpz_xor (&smart, &a, &smart);
  _mpz_assert (&smart);
  if (smart != dumb)
    panic ("(r = a ^ r) %s ^ %s = %s\n\t(should be %s)\n",
	   a.cstr (), b.cstr (), smart.cstr (), dumb.cstr ());
}

void
test (const bigint &a, const bigint &b)
{
  _test (a, b);
  _test (b, a);

  bigint aa = -a;
  bigint bb = -b;

  _test (aa, bb);
  _test (bb, aa);
  _test (a, bb);
  _test (bb, a);
  _test (aa, b);
  _test (b, aa);
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


int
main ()
{
  bigint a, b;

  test (bigint ("5"), bigint ("-1"));
  test (bigint ("0x9999999999999999"),
	~bigint (bigint ("0x6666666666666666")));
  test (bigint ("-0x5555555555555555"), bigint ("-0x1111111111111111"));
  test (bigint ("0x9999999999999999"),
	~bigint (bigint ("0x9999999999999999")));
  test (bigint ("0x55555555555555555"), bigint ("-0x1111111111111111"));
  test (bigint ("-0x1111111111111111"), bigint ("0x55555555555555555"));
  test (bigint ("0x55555555555555555"), bigint ("0x1111111111111111"));

  for (int i = 0; i < 50; i++) {
    getrnd (a); getrnd (b); test (a, b);
    getrnd (a, rand () % 10); getrnd (b, rand () % 10); test (a, b);
    getrnd2 (a); getrnd (b); test (a, b);
    getrnd2 (a, rand () % 10); getrnd (b, rand () % 10); test (a, b);
    getrnd2 (a); getrnd2 (b); test (a, b);
    getrnd2 (a, rand () % 10); getrnd2 (b, rand () % 10); test (a, b);
  }

  return 0;
}
