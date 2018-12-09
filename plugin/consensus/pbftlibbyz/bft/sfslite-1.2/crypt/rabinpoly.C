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

#include "rabinpoly.h"
#include "msb.h"
#ifndef INT64
# define INT64(n) n##LL
#endif
#define MSB64 INT64(0x8000000000000000)
                    
u_int64_t
polymod (u_int64_t nh, u_int64_t nl, u_int64_t d)
{
  //  assert (d);
  int k = fls64 (d) - 1;
  d <<= 63 - k;

  if (nh) {
    if (nh & MSB64)
      nh ^= d;
    for (int i = 62; i >= 0; i--)
      if (nh & ((u_int64_t) 1) << i) {
	nh ^= d >> (63 - i);
	nl ^= d << (i + 1);
      }
  }
  for (int i = 63; i >= k; i--)
  {  
    if (nl & INT64 (1) << i)
      nl ^= d >> (63 - i);
  }
  
  return nl;
}

u_int64_t
polygcd (u_int64_t x, u_int64_t y)
{
  for (;;) {
    if (!y)
      return x;
    x = polymod (0, x, y);
    if (!x)
      return y;
    y = polymod (0, y, x);
  }
}

void
polymult (u_int64_t *php, u_int64_t *plp, u_int64_t x, u_int64_t y)
{
  u_int64_t ph = 0, pl = 0;
  if (x & 1)
    pl = y;
  for (int i = 1; i < 64; i++)
    if (x & (INT64 (1) << i)) {
      ph ^= y >> (64 - i);
      pl ^= y << i;
    }
  if (php)
    *php = ph;
  if (plp)
    *plp = pl;
}

u_int64_t
polymmult (u_int64_t x, u_int64_t y, u_int64_t d)
{
  u_int64_t h, l;
  polymult (&h, &l, x, y);
  return polymod (h, l, d);
}

bool
polyirreducible (u_int64_t f)
{
  u_int64_t u = 2;
  int m = (fls64 (f) - 1) >> 1;
  for (int i = 0; i < m; i++) {
    u = polymmult (u, u, f);
    if (polygcd (f, u ^ 2) != 1)
      return false;
  }
  return true;
}

void
rabinpoly::calcT ()
{
  //  assert (poly >= 0x100);
  int xshift = fls64 (poly) - 1;
  shift = xshift - 8;
  u_int64_t T1 = polymod (0, INT64 (1) << xshift, poly);
  for (int j = 0; j < 256; j++)
  {
    T[j] = polymmult (j, T1, poly) | ((u_int64_t) j << xshift);
  }
}

rabinpoly::rabinpoly (u_int64_t p)
  : poly (p)
{
  calcT ();
}

window::window (u_int64_t poly)
  : rabinpoly (poly), fingerprint (0), bufpos (-1)
{
  u_int64_t sizeshift = 1;
  for (int i = 1; i < size; i++)
    sizeshift = append8 (sizeshift, 0);
  for (int i = 0; i < 256; i++)
    U[i] = polymmult (i, sizeshift, poly);
  bzero ((char*) buf, sizeof (buf));
}
