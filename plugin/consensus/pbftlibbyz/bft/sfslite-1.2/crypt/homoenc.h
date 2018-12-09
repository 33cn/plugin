// -*-c++-*-
/* $Id: homoenc.h,v 1.1 2006/02/25 02:02:39 mfreed Exp $ */

/*
 *
 * Copyright (C) 2005 Michael J. Freedman (mfreedman at alum.mit.edu)
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

#ifndef _HOMOENC_H_
#define _HOMOENC_H_ 1

#include "crypt_prot.h"
#include "crypt.h"

class homoenc_pub {
public:
  homoenc_pub () {}
  virtual ~homoenc_pub () {}

  virtual const size_t &mod_size     () const = 0;
  virtual const bigint &ptext_modulus () const = 0;
  virtual const bigint &ctext_modulus () const = 0;

  virtual bool encrypt (crypt_ctext *c, const bigint &msg, 
			bool recover) const = 0;
  virtual bool encrypt (crypt_ctext *c, const str &msg, bool recover) const {
    bigint m = pre_encrypt (msg);
    if (!m) return false;
    return encrypt (c, m, recover);
  }

  virtual crypt_keytype ctext_type () const = 0;

  // Resulting ciphertext is sum of msgs' corresponding plaintexts
  // ctext = ctext1*ctext2 ==> ptext = ptext1 + ptext2
  virtual void add (crypt_ctext *c, const crypt_ctext &msg1, 
		    const crypt_ctext &msg2) const = 0;

  // Resulting ciphertext is msg's corresponding plaintext * constant
  // ctext = ctext^const ==> ptext = const * ptext
  virtual void mult (crypt_ctext *c, const crypt_ctext &msg, 
		     const bigint &cons) const = 0;

  virtual bigint pre_encrypt  (const str &msg) const {
    size_t nbits = mod_size ();
    if (msg.len () > nbits) {
      warn << "pre_encrypt: message too large [len " 
	   << msg.len () << " bits " << nbits << "]\n";
      return 0;
    }
    
    bigint r;
    mpz_set_rawmag_le (&r, msg.cstr (), msg.len ());
    return r;
    
  }

  virtual str post_decrypt (const bigint &msg, size_t msglen) const {
    size_t nbits = mod_size ();
    if (msg.nbits () > nbits || msglen > nbits) {
      warn << "post_decrypt: message too large [len " << msg.nbits ()
	   << " buf " << msglen << " bits " << nbits << "]\n";
      return NULL;
    }
    
    zcbuf zm (nbits);
    mpz_get_rawmag_le (zm, zm.size, &msg);
  
    char *mp = zm;
    wmstr r (msglen);
    memcpy (r, mp, msglen);
    return r;
  }
};


class homoenc_priv : public virtual homoenc_pub {
public:
  homoenc_priv () {}
  virtual ~homoenc_priv () {}
  virtual str decrypt (const crypt_ctext &msg, size_t msglen, 
		       bool recover = true) const = 0;
};


#endif /* !_HOMOENC_H_  */
