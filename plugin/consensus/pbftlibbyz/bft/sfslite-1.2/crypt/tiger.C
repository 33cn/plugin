/* $Id: tiger.C 1117 2005-11-01 16:20:39Z max $ */

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

/* This source code is based on the implementation in the paper
 * "Tiger: A Fast New Hash Function" by Ross Anderson and Eli Biham.
 */


#include "tiger.h"

void
tiger::newstate (u_int64_t state[hashquads])
{
  state[0] = INT64 (0x0123456789abcdef);
  state[1] = INT64 (0xfedcba9876543210);
  state[2] = INT64 (0xf096a5b4c3b2e187);
}

void
tiger::transform (u_int64_t state[hashquads], const u_char block[blocksize])
{
#define getquad(cp)				\
  ((u_int64_t) (cp)[0]				\
   | (u_int64_t) (cp)[1] << 8			\
   | (u_int64_t) (cp)[2] << 16			\
   | (u_int64_t) (cp)[3] << 24			\
   | (u_int64_t) (cp)[4] << 32			\
   | (u_int64_t) (cp)[5] << 40			\
   | (u_int64_t) (cp)[6] << 48			\
   | (u_int64_t) (cp)[7] << 56)

#if 1
#define round(a, b, c, x, mul)					\
  c ^= x;							\
  a -= (t1[((c)>>(0*8))&0xff] ^ t2[((c)>>(2*8))&0xff]		\
	^ t3[((c)>>(4*8))&0xff] ^ t4[((c)>>(6*8))&0xff]);	\
  b += (t4[((c)>>(1*8))&0xff] ^ t3[((c)>>(3*8))&0xff]		\
	^ t2[((c)>>(5*8))&0xff] ^ t1[((c)>>(7*8))&0xff]);	\
  b *= mul;
#else
#define round(a, b, c, x, mul)				\
  c ^= x;						\
  a -= (t1[u_char (c)] ^ t2[u_char ((c)>>16)]		\
	^ t3[u_char ((c)>>32)] ^ t4[u_char ((c)>>48)]);	\
  b += (t4[u_char ((c)>>8)] ^ t3[u_char ((c)>>24)]	\
	^ t2[u_char ((c)>>40)] ^ t1[u_char ((c)>>56)]);	\
  b *= mul;
#endif

#define pass(a, b, c, mul)			\
  round (a, b, c, x0, mul);			\
  round (b, c, a, x1, mul);			\
  round (c, a, b, x2, mul);			\
  round (a, b, c, x3, mul);			\
  round (b, c, a, x4, mul);			\
  round (c, a, b, x5, mul);			\
  round (a, b, c, x6, mul);			\
  round (b, c, a, x7, mul);

#define key_schedule()				\
  x0 -= x7 ^ INT64(0xa5a5a5a5a5a5a5a5);		\
  x1 ^= x0;					\
  x2 += x1;					\
  x3 -= x2 ^ ((~x1)<<19);			\
  x4 ^= x3;					\
  x5 += x4;					\
  x6 -= x5 ^ ((~x4)>>23);			\
  x7 ^= x6;					\
  x0 += x7;					\
  x1 -= x0 ^ ((~x7)<<19);			\
  x2 ^= x1;					\
  x3 += x2;					\
  x4 -= x3 ^ ((~x2)>>23);			\
  x5 ^= x4;					\
  x6 += x5;					\
  x7 -= x6 ^ INT64(0x0123456789abcdef);

  u_int64_t a = state[0], b = state[1], c = state[2];

  u_int64_t x0 = getquad (block);
  u_int64_t x1 = getquad (block + 8);
  u_int64_t x2 = getquad (block + 16);
  u_int64_t x3 = getquad (block + 24);
  u_int64_t x4 = getquad (block + 32);
  u_int64_t x5 = getquad (block + 40);
  u_int64_t x6 = getquad (block + 48);
  u_int64_t x7 = getquad (block + 56);

  u_int64_t aa = a, bb = b, cc = c;

#if 0
  pass (a, b, c, 5);
  key_schedule ();
  pass (c, a, b, 7);
  key_schedule ();
  pass (b, c, a, 9);
#else
  for (int i = 0;; i++) {
    pass (a, b, c, ((i == 0) ? 5 : (i == 1) ? 7 : 9));
    u_int64_t t = a;
    a = c;
    c = b;
    b = t;
    if (i == 2)
      break;
    key_schedule ();
  }
#endif

  a ^= aa;
  b -= bb;
  c += cc;

  state[0] = a;
  state[1] = b;
  state[2] = c;
}

void
tiger::state2bytes (void *_out, const u_int64_t state[hashquads])
{
  u_char *out = static_cast<u_char *> (_out);
#define putquad_be(cp, val)			\
  (cp)[0] = val >> 56;				\
  (cp)[1] = val >> 48;				\
  (cp)[2] = val >> 40;				\
  (cp)[3] = val >> 32;				\
  (cp)[4] = val >> 24;				\
  (cp)[5] = val >> 16;				\
  (cp)[6] = val >> 8;				\
  (cp)[7] = val;

  putquad_be (out, state[0]);
  putquad_be (out + 8, state[1]);
  putquad_be (out + 16, state[2]);
}
