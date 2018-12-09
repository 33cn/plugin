// -*-c++-*-
/* $Id: rabinpoly.h,v 1.2 2002/04/12 06:20:35 lpcox Exp $ */

/*
 *
 * Copyright (C) 2000 David Mazieres (dm@uun.org)
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

#ifndef _RABINPOLY_H_
#define _RABINPOLY_H_ 1

#include <sys/types.h>
#include <string.h>

u_int64_t polymod (u_int64_t nh, u_int64_t nl, u_int64_t d);
u_int64_t polygcd (u_int64_t x, u_int64_t y);
void polymult (u_int64_t *php, u_int64_t *plp, u_int64_t x, u_int64_t y);
u_int64_t polymmult (u_int64_t x, u_int64_t y, u_int64_t d);
bool polyirreducible (u_int64_t f);
u_int64_t polygen (u_int degree);

class rabinpoly {
  int shift;
  u_int64_t T[256];		// Lookup table for mod
  void calcT ();
public:
  const u_int64_t poly;		// Actual polynomial

  explicit rabinpoly (u_int64_t poly);
  u_int64_t append8 (u_int64_t p, u_char m) const
    { return ((p << 8) | m) ^ T[p >> shift]; }
};

class window : public rabinpoly {
public:
  enum {size = 48};
  //enum {size = 24};
private:
  u_int64_t fingerprint;
  int bufpos;
  u_int64_t U[256];
  u_char buf[size];

public:
  window (u_int64_t poly);
  u_int64_t slide8 (u_char m) {
    if (++bufpos >= size)
      bufpos = 0;
    u_char om = buf[bufpos];
    buf[bufpos] = m;
    return fingerprint = append8 (fingerprint ^ U[om], m);
  }
  void reset () { 
    fingerprint = 0; 
    bzero ((char*) buf, sizeof (buf));
  }
};

#endif /* !_RABINPOLY_H_ */
