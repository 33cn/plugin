// -*-c++-*-
/* $Id: crypt.h 1117 2005-11-01 16:20:39Z max $ */

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

#ifndef _SFS_CRYPT_H_
#define _SFS_CRYPT_H_ 1

#include "arpc.h"
#include "wmstr.h"
#include "bigint.h"
#include "serial.h"

#include "prng.h"
extern prng rnd;

#include "rabin.h"
#include "rsa.h"
#include "blowfish.h"
#include "arc4.h"
#include "axprt_crypt.h"

#undef setbit

extern sha1oracle rnd_input;
void random_start ();		// Start init so later random_init is fast
void random_update ();
void random_set_seedfile (str path);
void random_init ();		// Call this or next function before using rnd
void random_init_file (str path);

u_int32_t random_getword ();

bigint random_zn (const bigint &n);
bigint random_bigint (size_t bits);

void mp_setscrub ();
void mp_clearscrub ();

#endif /* _SFS_CRYPT_H_ */
