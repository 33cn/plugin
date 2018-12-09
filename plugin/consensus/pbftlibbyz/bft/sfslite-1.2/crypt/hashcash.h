/*
 *
 * Copyright (C) 1999 Frans Kaashoek and David Mazieres (kaashoek@lcs.mit.edu)
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

#ifndef _HASHCASH_H_
#define _HASHCASH_H_

#include "sha1.h"

u_long hashcash_pay (char payment[sha1::blocksize],
		     const char inithash[sha1::hashsize], 
	             const char target[sha1::hashsize], unsigned int bitcost);

bool hashcash_check (const char payment[sha1::blocksize],
		     const char inithash[sha1::hashsize], 
		     const char target[sha1::hashsize], unsigned int bitcost);


#endif /* !_HASHCASH__ */
