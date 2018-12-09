// -*-c++-*-
/* $Id: password.h 1117 2005-11-01 16:20:39Z max $ */

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

#ifndef _SFSCRYPT_PASSWORD_H_
#define _SFSCRYPT_PASSWORD_H_ 1

#include "str.h"

str pw_armorsalt (u_int cost, str bsalt, str ptext = "");
bool pw_dearmorsalt (u_int *costp, str *bsaltp, str *ptextp, str armor);
str pw_dorawcrypt (str ptext, size_t outsize, eksblowfish *eksb);
str pw_rawcrypt (u_int cost, str pwd, str bsalt, str ptext = "",
		 size_t outsize = 0, eksblowfish *eksb = NULL);

str pw_gensalt (u_int cost, str ptext = "");
str pw_getptext (str salt);
str pw_crypt (str pwd, str salt, size_t outsize = 0, eksblowfish *eksb = NULL);
bigint pw_getint (str pwd, str salt, size_t nbits, eksblowfish *eksb = NULL);

#endif /* !_SFSCRYPT_PASSWORD_H_ */
