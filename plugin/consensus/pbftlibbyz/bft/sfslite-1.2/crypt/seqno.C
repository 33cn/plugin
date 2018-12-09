/* $Id: seqno.C 1117 2005-11-01 16:20:39Z max $ */

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

#include "seqno.h"

seqcheck::seqcheck (size_t n)
  : bottom (0), nbits (n)
{
  bv[0].zsetsize (nbits);
  bv[1].zsetsize (nbits);
}

bool
seqcheck::check (u_int64_t seqno)
{
  if (seqno < bottom)
    return false;
  seqno -= bottom;
  if (seqno >= 3 * nbits) {
    bottom += seqno;
    seqno = 0;
    bv[0].setrange (0, nbits, 0);
    bv[1].setrange (0, nbits, 0);
  }
  else if (seqno >= 2 * nbits) {
    bottom += nbits;
    seqno -= nbits;
    swap (bv[0], bv[1]);
    bv[1].setrange (0, nbits, 0);
  }

  bitvec *bvp;
  if (seqno >= nbits) {
    bvp = &bv[1];
    seqno -= nbits;
  }
  else
    bvp = &bv[0];
  if (bvp->at (seqno))
    return false;
  (*bvp)[seqno] = 1;
  return true;
}
