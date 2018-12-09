// -*-c++-*-
/* $Id: bitvec.h 3373 2008-06-04 14:31:12Z max $ */

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

/* XXX - this is suboptimal */

#ifndef _ASYNC_BITVEC_H_
#define _ASYNC_BITVEC_H_ 1

#include "str.h"

class bitvec {
protected:
  friend void swap (bitvec &a, bitvec &b);
  typedef unsigned long map_t;
  enum { mapbits = 8 * sizeof (map_t) };

  map_t *map;
  size_t nbits;

  void init () { map = NULL; nbits = 0; }
  static size_t nbytes (size_t nb)
    { return ((nb + mapbits - 1) / mapbits) * sizeof (map_t); }
  void datalloc (size_t nb) {
    if (nb)
      map = static_cast<map_t *> (xrealloc (map, nbytes (nb)));
    else {
      /* Jump through hoops for dmalloc */
      xfree (map);
      map = NULL;
    }
  }

#ifdef CHECK_BOUNDS
#define bcheck(n) assert ((size_t) (n) < nbits)
#define rcheck(s, e) assert (s <= e && e <= nbits)
#else /* !CHECK_BOUNDS */
#define bcheck(n)
#define rcheck(s, e)
#endif /* !CHECK_BOUNDS */

  class wbit {
    friend class bitvec;

    map_t *const intp;
    const u_int bitpos;

    wbit (map_t *i, u_int b) : intp (i), bitpos (b) {}

  public:
    operator bool () const { return *intp & map_t (1) << bitpos; }
    wbit &operator= (bool v) {
      if (v)
        *intp |= map_t (1) << bitpos;
      else
        *intp &= ~(map_t (1) << bitpos);
      return *this;
    }
    wbit &operator^= (bool v) { *intp ^= map_t (v) << bitpos; return *this; }
  };

  void range_set (size_t s, size_t e) {
    size_t sp = s / mapbits, ep = e / mapbits;
    int sb = s % mapbits, eb = e % mapbits;
    if (sp == ep) {
      if (eb)
	map[sp] |= (map_t (-1) << sb) & ~(map_t (-1) << eb);
    }
    else {
      map[sp] |= (map_t (-1) << sb);
      if (eb)
	map[ep] |= ~(map_t (-1) << eb);
      memset (&map[sp+1], 0xff, (ep - sp - 1) * sizeof (map_t));
    }
  }

  void range_clr (size_t s, size_t e) {
    size_t sp = s / mapbits, ep = e / mapbits;
    int sb = s % mapbits, eb = e % mapbits;
    if (sp == ep) {
      if (eb)
	map[sp] &= ~(map_t (-1) << sb) | (map_t (-1) << eb);
    }
    else {
      map[sp] &= ~(map_t (-1) << sb);
      if (eb)
	map[ep] &= (map_t (-1) << eb);
      bzero (&map[sp+1], (ep - sp - 1) * sizeof (map_t));
    }
  }

public:
  bitvec () { init (); }
  bitvec (size_t n) { init (); zsetsize (n); }
  bitvec (const bitvec &v) { init (); *this = v; }
  ~bitvec () { xfree (map); }

  void setsize (size_t n) { datalloc (nbits = n); }
  void zsetsize (size_t n) {
    datalloc (n);
    if (n > nbits)
      range_clr (nbits, n);
    nbits = n;
  }
  void osetsize (size_t n) {
    datalloc (n);
    if (n > nbits)
      range_set (nbits, n);
    nbits = n;
  }
  size_t size () const { return nbits; }

  bool at (ptrdiff_t i) const {
    bcheck (i);
    return map[(size_t) i / mapbits] & map_t (1) << ((size_t) i % mapbits);
  }
  void (setbit) (ptrdiff_t i, bool val) {
    bcheck (i);
    if (val)
      map[(size_t) i / mapbits] |= map_t (1) << ((size_t) i % mapbits);
    else
      map[(size_t) i / mapbits] &= ~(map_t (1) << ((size_t) i % mapbits));
  }
  void setrange (size_t s, size_t e, bool v) {
    rcheck (s, e);
    if (v)
      range_set (s, e);
    else
      range_clr (s, e);
  }

  bool operator[] (ptrdiff_t i) const {
    bcheck (i);
    return map[(size_t) i / mapbits] & map_t (1) << ((size_t) i % mapbits);
  }
  wbit operator[] (ptrdiff_t i) {
    bcheck (i);
    return wbit (map + (size_t) i / mapbits, (size_t) i % mapbits);
  }

  bitvec &operator= (const bitvec &v) {
    setsize (v.nbits);
    memcpy (map, v.map, nbytes (nbits));
    return *this;
  }

  // return the index of the first unset bit, or -1 if none found
  int first_unset_bit () const {
    int ret = -1;
    map_t all = 0;
    map_t tmp;
    all = ~all;
    size_t slots = nbits / mapbits + 1;
    ptrdiff_t bitindex = 0;
    for (size_t i = 0; ret < 0 && i < slots; i++) {
      if ((tmp = (map[i] ^ all))) {
	for (size_t j = 0; ret < 0 && j < mapbits; j++) {
	  if (tmp & 1) {
	    ret = bitindex;
	  } else {
	    tmp = (tmp >> 1);
	    bitindex++;
	  }
	}
	assert (ret >= 0);
      }
      bitindex += mapbits;
    }
    if (ret >= 0) bcheck (ret);
    return ret;
  }

#undef bcheck
#undef rcheck
};

inline void
swap (bitvec &a, bitvec &b)
{
  bitvec::map_t *map = a.map;
  a.map = b.map;
  b.map = map;

  size_t nbits = a.nbits;
  a.nbits = b.nbits;
  b.nbits = nbits;
}

inline const strbuf &
strbuf_cat (const strbuf &sb, const bitvec &v)
{
  char *p = sb.tosuio ()->getspace (v.size ());
  for (size_t i = 0; i < v.size (); i++)
    p[i] = v[i] ? '1' : '0';
  sb.tosuio ()->print (p, v.size ());
  return sb;
}

#endif /* !_ASYNC_BITVEC_H_ */
