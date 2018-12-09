/* $Id: arc4.C 3758 2008-11-13 00:36:00Z max $ */

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


#include "arc4.h"

void
arc4::reset ()
{
  i = 0xff;
  j = 0;
  for (int n = 0; n < 0x100; n++)
    s[n] = n;
}

void
arc4::_setkey (const u_char *key, size_t keylen)
{
  for (u_int n = 0, keypos = 0; n < 256; n++, keypos++) {
    if (keypos >= keylen)
      keypos = 0;
    i = (i + 1) & 0xff;
    u_char si = s[i];
	 j = (j + si + key[keypos]) & 0xff;
    s[i] = s[j];
    s[j] = si;
  }
}

void
arc4::setkey (const void *_key, size_t len)
{
  const u_char *key = static_cast<const u_char *> (_key);
  while (len > 128) {
    len -= 128;
    key += 128;
    _setkey (key, 128);
  }
  if (len > 0)
    _setkey (key, len);
  j = i;
}
