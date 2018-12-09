// -*-c++-*-
/* $Id: serial.h 3615 2008-09-19 06:26:22Z max $ */

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

#ifndef _ASYNC_SERIAL_H_
#define _ASYNC_SERIAL_H_ 1

#include "str.h"

bool str2file (str file, str s, int perm = 0666, bool excl = false,
	       struct stat *sbp = NULL, bool binary = false);
str file2str (str file);

static inline void
putshort (void *_dp, u_int16_t val)
{
  u_char *dp = static_cast<u_char *> (_dp);
  dp[0] = val >> 8;
  dp[1] = val;
}

static inline u_int16_t
getshort (const void *_dp)
{
  const u_char *dp = static_cast<const u_char *> (_dp);
  return dp[0] << 8 | dp[1];
}

static inline void
putint (void *_dp, u_int32_t val)
{
  u_char *dp = static_cast<u_char *> (_dp);
  dp[0] = val >> 24;
  dp[1] = val >> 16;
  dp[2] = val >> 8;
  dp[3] = val;
}

static inline u_int32_t
getint (const void *_dp)
{
  const u_char *dp = static_cast<const u_char *> (_dp);
  return dp[0] << 24 | dp[1] << 16 | dp[2] << 8 | dp[3];
}

static inline void
puthyper (void *_dp, u_int64_t val)
{
  u_char *dp = static_cast<u_char *> (_dp);
  dp[0] = val >> 56;
  dp[1] = val >> 48;
  dp[2] = val >> 40;
  dp[3] = val >> 32;
  dp[4] = val >> 24;
  dp[5] = val >> 16;
  dp[6] = val >> 8;
  dp[7] = val;
}

static inline u_int64_t
gethyper (const void *_dp)
{
  const u_char *dp = static_cast<const u_char *> (_dp);
  return (u_int64_t) dp[0] << 56 | (u_int64_t) dp[1] << 48
    | (u_int64_t) dp[2] << 40 | (u_int64_t) dp[3] << 32
    | getint (dp + 4);
}

str armor64 (const void *, size_t);
size_t armor64len (const u_char *);
str dearmor64 (const char *, ssize_t len = -1);


/*
 * RFC 1521 Base-64 encoding
 */

inline str
armor64 (str bin)
{
  return armor64 (bin.cstr (), bin.len ());
}

inline str
dearmor64 (str asc)
{
  if (armor64len (reinterpret_cast<const u_char *> (asc.cstr ()))
      != asc.len ())
    return NULL;
  return dearmor64 (asc.cstr (), asc.len ());
}

str armor32 (const void *, size_t);
size_t armor32len (const u_char *s);
str dearmor32 (const char *, ssize_t len = -1);


/*
 * Alternate Base-64 encoding, suitable for file names.  Uses '-'
 * instead of '/' and does not append any '=' chars.
 */

str armor64A (const void *s, size_t len);
size_t armor64Alen (const u_char *s);
str dearmor64A (const char *s, ssize_t len);

inline str
armor64A (str bin)
{
  return armor64A (bin.cstr (), bin.len ());
}

inline str
dearmor64A (str asc)
{
  if (armor64Alen (reinterpret_cast<const u_char *> (asc.cstr ()))
      != asc.len ())
    return NULL;
  return dearmor64A (asc.cstr (), asc.len ());
}

/*
 * PHP and Python-complaint Base-64 encoding.
 */
str armor64X (const void *s, size_t len);
size_t armor64Xlen (const u_char *s);
str dearmor64X (const char *s, ssize_t len);

inline str
armor64X (str bin)
{
  return armor64X (bin.cstr (), bin.len ());
}

inline str
dearmor64X (str asc)
{
  if (armor64Xlen (reinterpret_cast<const u_char *> (asc.cstr ()))
      != asc.len ())
    return NULL;
  return dearmor64X (asc.cstr (), asc.len ());
}


/*
 * Base-32 encoding, using '2'-'9','a'-'k','m','n','p'-'z'
 * (i.e. digits and lower-case letters except 0,1,o,l).
 */

inline str
armor32 (str bin)
{
  return armor32 (bin.cstr (), bin.len ());
}

inline str
dearmor32 (str asc)
{
  if (armor32len (reinterpret_cast<const u_char *> (asc.cstr ()))
      != asc.len ())
    return NULL;
  return dearmor32 (asc.cstr (), asc.len ());
}

/*
 * Base-16 encoding
 */

inline bool
hexconv (int &out, const char in)
{
  if (in >= '0' && in <= '9')
    out = in - '0';
  else if (in >= 'a' && in <= 'f')
    out = in - ('a' - 10);
  else if (in >= 'A' && in <= 'F')
    out = in - ('A' - 10);
  else
    return false;
  return true;
}

inline str
hex2bytes (str asc)
{
  if (asc.len () & 1)
    return NULL;
  mstr b (asc.len () / 2);
  int i = 0;
  for (const char *p = asc.cstr (), *e = p + asc.len (); p < e; p += 2) {
    int h, l;
    if (!hexconv (h, p[0]) || !hexconv (l, p[1]))
      return NULL;
    b[i++] = h << 4 | l;
  }
  return b;
}

// Note: see hexdump in str.h to print data in base-16

#endif /* !_ASYNC_SERIAL_H_ */
