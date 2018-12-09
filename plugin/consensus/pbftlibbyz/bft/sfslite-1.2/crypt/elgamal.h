// -*-c++-*-
/* $Id: elgamal.h,v 1.1 2006/02/25 02:02:40 mfreed Exp $ */

/*
 *
 * Copyright (C) 2006 Michael J. Freedman (mfreedman at alum.mit.edu)
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

#ifndef _ELGAMAL_H_
#define _ELGAMAL_H_ 1

#include "homoenc.h"

class elgamal_pub : public virtual homoenc_pub {
public:
  const bigint p;		/* Prime modulus: 2q + 1     */
  const bigint q;		/* Prime modulus             */      
  const bigint g;		/* Random generator of Z^*_p */
  const bigint gr;		/* Generator^r               */
  const size_t nbits, abits;    /* Size of p, r              */
  const bigint p1;
  const bigint q1;

  elgamal_pub (const bigint &pp, const bigint &qq, 
	       const bigint &gg, const bigint &ggr, size_t aabits);

  virtual const size_t &mod_size      () const { return nbits; }
  virtual const bigint &ptext_modulus () const { return q;     }
  virtual const bigint &ctext_modulus () const { return p;     }

  virtual bool encrypt (crypt_ctext *c, const bigint &msg, bool recover) const;
  virtual bool encrypt (crypt_ctext *c, const str &msg,    bool recover) const
  { return homoenc_pub::encrypt (c, msg, recover); }

  virtual crypt_keytype ctext_type () const { return CRYPT_ELGAMAL; }

  // Resulting ciphertext is sum of msgs' corresponding plaintexts
  // ctext = ctext1*ctext2 ==> ptext = ptext1 + ptext2
  virtual void add (crypt_ctext *c, const crypt_ctext &msg1, 
		    const crypt_ctext &msg2) const;

  // Resulting ciphertext is msg's corresponding plaintext * constant
  // ctext = ctext1^const ==> ptext = const * ptext1 
  virtual void mult (crypt_ctext *c, const crypt_ctext &msg, 
		     const bigint &cons) const;

};


class elgamal_priv : public elgamal_pub, public virtual homoenc_priv {
public:
  const bigint r;		/* Random interger \in Z^*_p */
  const bigint i2;              /* 2^-1 mod q                */

  elgamal_priv (const bigint &pp, const bigint &qq, 
		const bigint &gg, const bigint &rr);

  static ptr<elgamal_priv> make (const bigint &p, 
				 const bigint &g, 
				 const bigint &r);
    
  str decrypt (const crypt_ctext &msg, size_t msglen, 
	       bool recover = true) const;
};

// abits gives length of random moduli [e.g., 160 bits for 1024-bit keys]
elgamal_priv elgamal_keygen (size_t nbits, size_t abits, u_int iter = 32);

#endif /* !_ELGAMAL_H_  */
