// -*-c++-*-
/* $Id: arc4.h 3758 2008-11-13 00:36:00Z max $ */

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


#ifndef _ARC4_H_
#define _ARC4_H_

#include "sysconf.h"

/* Arcfour random stream generator.  This code is derived from section
 * 17.1 of Applied Cryptography, second edition, which describes a
 * stream cipher allegedly compatible with RSA Labs "RC4" cipher (the
 * actual description of which is a trade secret).  The same algorithm
 * is used as a stream cipher called "arcfour" in Tatu Ylonen's ssh
 * package.
 *
 * RC4 is a registered trademark of RSA Laboratories.
 */

class arc4 {
  u_char i;
  u_char j;
  u_char s[256];

  void _setkey (const u_char *key, size_t len);

public:
  arc4 () { reset (); }
  ~arc4 () { i = j = 0; bzero (s, sizeof (s)); }

  void reset ();
  void setkey (const void *key, size_t len);

  u_char getbyte () {
    i = (i + 1) & 0xff;
    u_char si = s[i];
    j = (j + si) & 0xff;
    u_char sj = s[j];
    s[i] = sj;
    s[j] = si;
    return s[(si + sj) & 0xff];
  }
};

#endif /* _ARC4_H_ */
