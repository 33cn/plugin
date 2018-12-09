// -*- c++ -*-
/* $Id: aes.h 1117 2005-11-01 16:20:39Z max $ */

/*
 *
 * Copyright (C) 2001 David Mazieres (dm@uun.org)
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


#ifndef _CRYPT_AES_H_
#define _CRYPT_AES_H_ 1

#include "sysconf.h"

class aes_e {
protected:
  int nrounds;
  u_int32_t  e_key[60];
  void setkey_e (const char *key, u_int keylen);
public:
  ~aes_e () { nrounds = 0; bzero (e_key, sizeof (e_key)); }
  void setkey (const void *key, u_int keylen);
  void encipher_bytes (void *buf, const void *ibuf) const;
  void encipher_bytes (void *buf) const { encipher_bytes (buf, buf); }
};

class aes : public aes_e {
protected:
  u_int32_t  d_key[60];
  void setkey_d ();
public:
  ~aes () { bzero (d_key, sizeof (d_key)); }
  void setkey (const void *key, u_int keylen);
  void decipher_bytes (void *buf, const void *ibuf) const;
  void decipher_bytes (void *buf) const { decipher_bytes (buf, buf); }
};

#endif /* !_CRYPT_AES_H_ */
