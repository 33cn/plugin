// -*-c++-*-
/* $Id: prng.h 1117 2005-11-01 16:20:39Z max $ */

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

#ifndef _PRNG_H_
#define _PRNG_H_ 1

#include "async.h"
#include "sha1.h"

template<unsigned int N> struct sumbuf {
  union {
    char chars[N * 4];
    u_char bytes[N * 4];
    u_int32_t words[N];
  };

  sumbuf () {}
  ~sumbuf () { bzero (bytes, sizeof (bytes)); }

  void set (const u_int32_t dp[N]) { memcpy (bytes, dp, sizeof (bytes)); }
  void set (const u_char cp[N * 4]) { memcpy (bytes, cp, sizeof (bytes)); }
  // template<size_t M>
  // inline void add (const sumbuf<M> *, bool carryin = false);
};

// G++ chokes when this is a member function
template<size_t N, size_t M> inline bool
sumbufadd (sumbuf<N> *dst, const sumbuf<M> *src, bool carry = false)
{
  u_int64_t sum = carry;
  size_t i;
  for (i = 0; i < min<size_t> (N, M); i++) {
    dst->words[i] = sum = sum + dst->words[i] + src->words[i];
    sum >>= 32;
  }
  while (i < N && (u_int32_t) (sum)) {
    dst->words[i] = sum = sum + dst->words[i];
    sum >>= 32;
  }
  return sum;
}

class prng : public datasink {
  static const u_int32_t initdat[];

  sumbuf<16> state;
  sumbuf<16> input;

  u_char *inpos;
  u_char *const inlim;

  void transform (sumbuf<5> *);
public:
  prng ();
  virtual ~prng () {}
  void seed (const u_char[64]);
  void seed_oracle (sha1oracle *);
  void update (const void *, size_t);

  void getbytes (void *, size_t);
  u_int32_t getword () {
    u_int32_t ret;
    getbytes (&ret, sizeof (ret));
    return ret;
  }
  u_int64_t gethyper () {
    u_int64_t ret;
    getbytes (&ret, sizeof (ret));
    return ret;
  }
};

#if 0
// XXX - g++ bug: sumbuf::add must be defined after prng.
template<size_t N> template<size_t M> inline void
sumbuf<N>::add (const sumbuf<M> *sb, bool carryin)
{
  u_int64_t sum = carryin;
  size_t i;
  for (i = 0; i < min (N, M); i++) {
    words[i] = sum = sum + words[i] + sb->words[i];
    sum >>= 32;
  }
  while (i < N && (u_int32_t) sum) {
    words[i] = sum = sum + words[i];
    sum >>= 32;
  }
}
#endif

void getclocknoise (datasink *);
void getfdnoise (datasink *, int fd, cbv cb, size_t maxbytes = size_t (-1));
void getfilenoise (datasink *, const char *, cbv, size_t = size_t (-1));
void getprognoise (datasink *, char *const *, cbv);
void getsysnoise (datasink *, cbv);
bool getkbdnoise (size_t, datasink *, cbv);
bool getkbdpwd (str, datasink *, cbs);
bool getkbdline (str, datasink *, cbs, str def = NULL);

#endif /* !_PRNG_H_ */
