/* $Id: bbuddy.C 1117 2005-11-01 16:20:39Z max $ */

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

#include "bbuddy.h"
#include "bitvec.h"
#include "msb.h"

#undef setbit
#undef clrbit

/* Population count (number of bits set) */
const char bytepop[0x100] = {
  0, 1, 1, 2, 1, 2, 2, 3, 1, 2, 2, 3, 2, 3, 3, 4, 1, 2, 2, 3, 2, 3, 3, 4,
  2, 3, 3, 4, 3, 4, 4, 5, 1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5,
  2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6, 1, 2, 2, 3, 2, 3, 3, 4,
  2, 3, 3, 4, 3, 4, 4, 5, 2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6,
  2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6, 3, 4, 4, 5, 4, 5, 5, 6,
  4, 5, 5, 6, 5, 6, 6, 7, 1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5,
  2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6, 2, 3, 3, 4, 3, 4, 4, 5,
  3, 4, 4, 5, 4, 5, 5, 6, 3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7,
  2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6, 3, 4, 4, 5, 4, 5, 5, 6,
  4, 5, 5, 6, 5, 6, 6, 7, 3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7,
  4, 5, 5, 6, 5, 6, 6, 7, 5, 6, 6, 7, 6, 7, 7, 8,
};

class bbfree : private bitvec {
  size_t hint;			// Possibly the index of a non-zero map_t
  size_t cnt;			// # of 1 bits
  size_t nmaps;

public:
  enum { mapbits = bitvec::mapbits };
  explicit bbfree (size_t nb = 0) : hint (0), cnt (0) { init (nb); }
  void init (size_t nb) {
    assert (nb >= nbits);
    size_t obits = nbits;
    setsize (nb);
    nmaps = (nb + mapbits - 1) / mapbits;
    range_clr (obits, nmaps * mapbits);
  }
  size_t getsize () const { return nbits; }

  bool getbit (u_long pos) const { return at (pos); }
  void setbit (u_long pos) {
    const size_t mi = pos / mapbits;
    const map_t mask = map_t (1) << pos % mapbits;
#ifdef CHECK_BOUNDS
    assert (pos < nbits);
    assert (!(map[mi] & mask));
#endif /* CHECK_BOUNDS */
    map[mi] |= mask;
    if (!map[hint])
      hint = mi;
    cnt++;
  }
  void clrbit (u_long pos) {
    const size_t mi = pos / mapbits;
    const map_t mask = map_t (1) << pos % mapbits;
#ifdef CHECK_BOUNDS
    assert (pos < nbits);
    assert (map[mi] & mask);
#endif /* CHECK_BOUNDS */
    map[mi] &= ~mask;
    cnt--;
  }

  bool findbit (size_t *posp);
  void _check () const;
};

bool
bbfree::findbit (size_t *posp)
{
  if (!cnt || !nbits)
    return false;
  if (map_t v = map[hint]) {
    *posp = hint * mapbits + ffs (v) - 1;
    return true;
  }
  for (size_t i = 0; i < nmaps; i++)
    if (map_t v = map[i]) {
      hint = i;
      *posp = i * mapbits + ffs (v) - 1;
      return true;
    }
  panic ("bbfree::findbit: cnt was wrong!\n");
}

void
bbfree::_check () const
{
  size_t sum = 0;
  for (u_char *cp = reinterpret_cast<u_char *> (map),
	 *end = cp + nbytes (nbits);
       cp < end; cp++)
    sum += bytepop[*cp];
  assert (sum == cnt);
}

inline bbfree &
bbuddy::fm (size_t sn)
{
  assert (sn >= log2minalloc && sn <= log2maxalloc);
  return freemaps[sn - log2minalloc];
}

bbuddy::bbuddy (totsize_t ts, size_t minalloc, size_t maxalloc)
  : totsize (0),
    log2minalloc (log2c (minalloc)), log2maxalloc (log2c (maxalloc)),
    freemaps (New bbfree[1 + log2maxalloc - log2minalloc]),
    spaceleft (0)
{
  assert (log2maxalloc >= log2minalloc);
  settotsize (ts);
}

void
bbuddy::settotsize (totsize_t ts)
{
  const size_t maxinc = 1 << log2maxalloc;
  ts = ts & ~(maxinc - 1);
  assert (ts >= totsize);

  totsize_t o = ts >> log2minalloc;
  for (u_int sn = log2minalloc; sn <= log2maxalloc; sn++) {
    fm (sn).init (o);
    o >>= 1;
  }

  o = totsize;
  if (o < ts) {
    for (;;) {
      size_t n = ffs (o) - 1;
      if (n >= log2maxalloc)
	break;
      size_t inc = 1 << n;
      dealloc (o, inc);
      o += inc;
    }
    while (o < ts) {
      dealloc (o, maxinc);
      o += maxinc;
    }
  }

  totsize = ts;
}

bbuddy::~bbuddy ()
{
  delete[] freemaps;
}

off_t
bbuddy::alloc (size_t n)
{
  u_int sn = log2c (n);
  if (sn < log2minalloc)
    sn = log2minalloc;
  if (sn > log2maxalloc)
    return -1;

  size_t pos;

  if (fm (sn).findbit (&pos)) {
    fm (sn).clrbit (pos);
    spaceleft -= 1 << sn;
    return (off_t) (pos) << sn;
  }

  u_int sni = sn;
  do {
    if (++sni > log2maxalloc)
      return -1;
  } while (!fm (sni).findbit (&pos));
  fm (sni).clrbit (pos);
  while (--sni >= sn) {
    pos <<= 1;
    fm (sni).setbit (pos + 1);
  }
  spaceleft -= 1 << sn;
  return (off_t) (pos) << sn;
}

void
bbuddy::dealloc (off_t off, size_t len)
{
  u_int sn = log2c (len);
  if (sn < log2minalloc)
    sn = log2minalloc;
  if (sn > log2maxalloc)
    panic ("bbuddy::dealloc: invalid len %lu\n", (u_long) len);

  size_t pos = off >> sn;
  assert (off == (off_t) (pos) << sn);
  spaceleft += 1 << sn;

  u_int sni;
  for (sni = sn; sni < log2maxalloc; sni++, pos >>= 1) {
    bbfree &bf = fm (sni);
    if (bf.getbit (pos ^ 1))
      bf.clrbit (pos ^ 1);
    else {
      bf.setbit (pos);
      return;
    }
  }
  fm (log2maxalloc).setbit (pos);
}

bool
bbuddy::_check_pos (u_int sn, size_t pos, bool set)
{
  bool ret = fm (sn).getbit (pos);
  if (ret) {
    if (set)
      panic ("bbuddy::_check_pos: bit should not be set!\n");
    set = true;
  }
  if (sn > log2minalloc) {
    sn--;
    pos <<= 1;
    bool ret1 = _check_pos (sn, pos, set);
    bool ret2 = _check_pos (sn, pos + 1, set);
    assert (!ret1 || !ret2);
  }
  return ret;
}

void
bbuddy::_check ()
{
  for (u_int sn = log2minalloc; sn <= log2maxalloc; sn++)
    fm (sn)._check ();
  for (size_t pos = 0, lim = fm (log2maxalloc).getsize (); pos < lim; pos++)
    _check_pos (log2maxalloc, pos, false);
}
