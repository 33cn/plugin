/* $Id: armor.C 3414 2008-06-14 15:50:24Z max $ */

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

#include "serial.h"

static const char b2a32[32] = {
  'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
  'i', 'j', 'k', 'm', 'n', 'p', 'q', 'r',
  's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
  '2', '3', '4', '5', '6', '7', '8', '9',
};

static const signed char a2b32[256] = {
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, 24, 25, 26, 27, 28, 29, 30, 31, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, -1, 11, 12, -1,
  13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
};

static const int b2a32rem[5] = {0, 2, 4, 5, 7};
static const int a2b32rem[8] = {0, -1, 1, -1, 2, 3, -1, 4};

    /*   0 1 2 3 4 5 6 7
     * 0 5 3
     * 1   2 5 1
     * 2       4 4
     * 3         1 5 2
     * 4             3 5
     */

str
armor32 (const void *dp, size_t dl)
{
  const u_char *p = static_cast<const u_char *> (dp);
  int rem = dl % 5;
  const u_char *e = p + (dl - rem);
  mstr res ((dl / 5) * 8 + b2a32rem[rem]);
  char *d = res;

  while (p < e) {
    d[0] = b2a32[p[0] >> 3];
    d[1] = b2a32[(p[0] & 0x7) << 2 | p[1] >> 6];
    d[2] = b2a32[p[1] >> 1 & 0x1f];
    d[3] = b2a32[(p[1] & 0x1) << 4 | p[2] >> 4];
    d[4] = b2a32[(p[2] & 0xf) << 1 | p[3] >> 7];
    d[5] = b2a32[p[3] >> 2 & 0x1f];
    d[6] = b2a32[(p[3] & 0x3) << 3 | p[4] >> 5];
    d[7] = b2a32[p[4] & 0x1f];
    p += 5;
    d += 8;
  }

  switch (rem) {
  case 4:
    d[6] = b2a32[(p[3] & 0x3) << 3];
    d[5] = b2a32[p[3] >> 2 & 0x1f];
    d[4] = b2a32[(p[2] & 0xf) << 1 | p[3] >> 7];
    d[3] = b2a32[(p[1] & 0x1) << 4 | p[2] >> 4];
    d[2] = b2a32[p[1] >> 1 & 0x1f];
    d[1] = b2a32[(p[0] & 0x7) << 2 | p[1] >> 6];
    d[0] = b2a32[p[0] >> 3];
    d += 7;
    break;
  case 3:
    d[4] = b2a32[(p[2] & 0xf) << 1];
    d[3] = b2a32[(p[1] & 0x1) << 4 | p[2] >> 4];
    d[2] = b2a32[p[1] >> 1 & 0x1f];
    d[1] = b2a32[(p[0] & 0x7) << 2 | p[1] >> 6];
    d[0] = b2a32[p[0] >> 3];
    d += 5;
    break;
  case 2:
    d[3] = b2a32[(p[1] & 0x1) << 4];
    d[2] = b2a32[p[1] >> 1 & 0x1f];
    d[1] = b2a32[(p[0] & 0x7) << 2 | p[1] >> 6];
    d[0] = b2a32[p[0] >> 3];
    d += 4;
    break;
  case 1:
    d[1] = b2a32[(p[0] & 0x7) << 2];
    d[0] = b2a32[p[0] >> 3];
    d += 2;
    break;
  }

  assert (d == res + res.len ());
  return res;
}

size_t
armor32len (const u_char *s)
{
  const u_char *p = s;
  while (a2b32[*p] >= 0)
    p++;
  return p - s;
}

str
dearmor32 (const char *_s, ssize_t len)
{
  const u_char *s = reinterpret_cast<const u_char *> (_s);

  if (len < 0)
    len = armor32len (s);
  int rem = a2b32rem[len & 7];
  if (rem < 0)
    return NULL;
  if (!len)
    return "";

  mstr bin ((len >> 3) * 5 + rem);
  char *d = bin;
  int c0, c1, c2, c3, c4, c5, c6, c7;

  for (const u_char *e = s + (len & ~7); s < e; s += 8, d += 5) {
    c0 = a2b32[s[0]];
    c1 = a2b32[s[1]];
    d[0] = c0 << 3 | c1 >> 2;
    c2 = a2b32[s[2]];
    c3 = a2b32[s[3]];
    d[1] = c1 << 6 | c2 << 1 | c3 >> 4;
    c4 = a2b32[s[4]];
    d[2] = c3 << 4 | c4 >> 1;
    c5 = a2b32[s[5]];
    c6 = a2b32[s[6]];
    d[3] = c4 << 7 | c5 << 2 | c6 >> 3;
    c7 = a2b32[s[7]];
    d[4] = c6 << 5 | c7;
  }

  if (rem >= 1) {
    c0 = a2b32[s[0]];
    c1 = a2b32[s[1]];
    *d++ = c0 << 3 | c1 >> 2;
    if (rem >= 2) {
      c2 = a2b32[s[2]];
      c3 = a2b32[s[3]];
      *d++ = c1 << 6 | c2 << 1 | c3 >> 4;
      if (rem >= 3) {
	c4 = a2b32[s[4]];
	*d++ = c3 << 4 | c4 >> 1;
	if (rem >= 4) {
	  c5 = a2b32[s[5]];
	  c6 = a2b32[s[6]];
	  *d++ = c4 << 7 | c5 << 2 | c6 >> 3;
	}
      }
    }
  }

  assert (d == bin + bin.len ());
  return bin;
}

static const char b2a64[64] = {
  'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H',
  'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P',
  'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X',
  'Y', 'Z', 'a', 'b', 'c', 'd', 'e', 'f',
  'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n',
  'o', 'p', 'q', 'r', 's', 't', 'u', 'v',
  'w', 'x', 'y', 'z', '0', '1', '2', '3',
  '4', '5', '6', '7', '8', '9', '+', '/',
};

static const signed char a2b64[256] = {
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, 62, -1, -1, -1, 63,
  52, 53, 54, 55, 56, 57, 58, 59, 60, 61, -1, -1, -1, -1, -1, -1,
  -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14,
  15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, -1, -1, -1, -1, -1,
  -1, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40,
  41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
};

inline str
_armor64 (const char *b2a, bool endpad, const void *dp, size_t len)
{
  const u_char *p = static_cast<const u_char *> (dp);
  int rem = len % 3;
  const u_char *e = p + (len - rem);
  mstr res (((len + 2) / 3) * 4);
  char *d = res;

  while (p < e) {
    d[0] = b2a[p[0] >> 2];
    d[1] = b2a[(p[0] & 0x3) << 4 | p[1] >> 4];
    d[2] = b2a[(p[1] & 0xf) << 2 | p[2] >> 6];
    d[3] = b2a[p[2] & 0x3f];
    d += 4;
    p += 3;
  }

  switch (rem) {
  case 1:
    d[0] = b2a[p[0] >> 2];
    d[1] = b2a[(p[0] & 0x3) << 4];
    d[2] = d[3] = '=';
    d += 4;
    break;
  case 2:
    d[0] = b2a[p[0] >> 2];
    d[1] = b2a[(p[0] & 0x3) << 4 | p[1] >> 4];
    d[2] = b2a[(p[1] & 0xf) << 2];
    d[3] = '=';
    d += 4;
    break;
  }

  assert (d == res + res.len ());
  if (!endpad && rem)
    res.setlen (res.len () + rem - 3);

  return res;
}

str
armor64 (const void *s, size_t len)
{
  return _armor64 (b2a64, true, s, len);
}

inline str
_dearmor64 (const signed char *a2b, const u_char *s, ssize_t len)
{
  if (!len)
    return "";

  mstr bin (len - (len>>2));
  char *d = bin;
  int c0, c1, c2, c3;

  for (const u_char *e = s + len - 4; s < e; s += 4, d += 3) {
    c0 = a2b[s[0]];
    c1 = a2b[s[1]];
    d[0] = c0 << 2 | c1 >> 4;
    c2 = a2b[s[2]];
    d[1] = c1 << 4 | c2 >> 2;
    c3 = a2b[s[3]];
    d[2] = c2 << 6 | c3;
  }

  c0 = a2b[s[0]];
  c1 = a2b[s[1]];
  *d++ = c0 << 2 | c1 >> 4;
  if ((c2 = a2b[s[2]]) >= 0) {
    *d++ = c1 << 4 | c2 >> 2;
    if ((c3 = a2b[s[3]]) >= 0)
      *d++ = c2 << 6 | c3;
  }

  bin.setlen (d - bin);
  return bin;
}

static size_t
_armor64len (const signed char *a2b, bool pad, const u_char *s)
{
  const u_char *p = s;
  while (a2b[*p] >= 0)
    p++;
  if (pad) {
    if (*p == '=')
      p++;
    if (*p == '=')
      p++;
  }
  return p - s;
}

size_t armor64len (const u_char *s) { return _armor64len (a2b64, true, s); }

str
dearmor64 (const char *_s, ssize_t len)
{
  const u_char *s = reinterpret_cast<const u_char *> (_s);
  if (len < 0)
    len = armor64len (s);
  if (len & 3)
    return NULL;
  return _dearmor64 (a2b64, s, len);
}

static const char b2a64A[64] = {
  'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H',
  'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P',
  'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X',
  'Y', 'Z', 'a', 'b', 'c', 'd', 'e', 'f',
  'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n',
  'o', 'p', 'q', 'r', 's', 't', 'u', 'v',
  'w', 'x', 'y', 'z', '0', '1', '2', '3',
  '4', '5', '6', '7', '8', '9', '+', '-',
};

static const signed char a2b64A[256] = {
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, 62, -1, 63, -1, -1,
  52, 53, 54, 55, 56, 57, 58, 59, 60, 61, -1, -1, -1, -1, -1, -1,
  -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14,
  15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, -1, -1, -1, -1, -1,
  -1, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40,
  41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
};

static char b2a64X[64] = {
  'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H',
  'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P',
  'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X',
  'Y', 'Z', 'a', 'b', 'c', 'd', 'e', 'f',
  'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n',
  'o', 'p', 'q', 'r', 's', 't', 'u', 'v',
  'w', 'x', 'y', 'z', '0', '1', '2', '3',
  '4', '5', '6', '7', '8', '9', '@', '_',
};

static const signed char a2b64X[256] = {
    -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
    -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
    -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
    52, 53, 54, 55, 56, 57, 58, 59, 60, 61, -1, -1, -1, -1, -1, -1,
    62, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14,
    15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, -1, -1, -1, -1, 63,
    -1, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40,
    41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, -1, -1, -1, -1, -1,
    -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
    -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
    -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
    -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
    -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
    -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
    -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
    -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
};

str
armor64A (const void *s, size_t len)
{
  return _armor64 (b2a64A, false, s, len);
}

str
armor64X(const void* s, size_t len)
{
    return _armor64(b2a64X, true, s, len);
}

size_t armor64Alen (const u_char *s) { return _armor64len (a2b64A, false, s); }
size_t armor64Xlen (const u_char *s) { return _armor64len (a2b64X, true, s); }

str
dearmor64A (const char *_s, ssize_t len)
{
  const u_char *s = reinterpret_cast<const u_char *> (_s);
  if (len < 0)
    len = armor64Alen (s);
  return _dearmor64 (a2b64A, s, len);
}

str
dearmor64X (const char *_s, ssize_t len)
{
  const u_char *s = reinterpret_cast<const u_char *> (_s);
  if (len < 0)
    len = armor64Xlen (s);
  if (len & 3)
    return NULL;
  return _dearmor64 (a2b64X, s, len);
}
