// -*-c++-*-
/* $Id: paillier.h 2330 2006-11-19 18:18:00Z max $ */

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

#ifndef _PAILLIER_H_
#define _PAILLIER_H_ 1

#include "homoenc.h"

// Use Chinese Remaindering for fast decryption
#define _PAILLIER_CRT_ 1

class paillier_pub : public virtual homoenc_pub {
public:
  const bigint n;		/* Modulus */
  const bigint g;		/* Basis g */
  const size_t nbits;

protected:
  bool fast;
  bigint nsq;	   	        /* Modulus^2 */
  // _FAST_PAILLIER_  
  bigint gn;                    /* g^N % nsq */

  void init ();

public:
  paillier_pub (const bigint &nn);
  paillier_pub (const bigint &nn, const bigint &gg);
  virtual ~paillier_pub () {}

  virtual const size_t &mod_size      () const { return nbits; }
  virtual const bigint &ptext_modulus () const { return n;     }
  virtual const bigint &ctext_modulus () const { return nsq;   }

  virtual bool encrypt (crypt_ctext *c, const bigint &msg, bool recover) const;
  virtual bool encrypt (crypt_ctext *c, const str &msg,
			bool recover = true) const
  { return homoenc_pub::encrypt (c, msg, recover); }

  virtual crypt_keytype ctext_type () const { return CRYPT_PAILLIER; }

  // Resulting ciphertext is sum of msgs' corresponding plaintexts
  // ctext = ctext1*ctext2 ==> ptext = ptext1 + ptext2
  virtual void add (crypt_ctext *c, const crypt_ctext &msg1, 
		    const crypt_ctext &msg2) const;

  // Resulting ciphertext is msg's corresponding plaintext * constant
  // ctext = ctext^const ==> ptext = const * ptext 
  virtual void mult (crypt_ctext *c, const crypt_ctext &msg, 
		     const bigint &cons) const;
};


class paillier_priv : public paillier_pub, public virtual homoenc_priv {
public:
  const bigint p;		/* Smaller prime */
  const bigint q;	        /* Larger prime  */
  const bigint a;               /* For fast decryption */

protected:
  bigint p1;		        /* p-1             */
  bigint q1;		        /* q-1             */

  bigint k;	       	        /* lcm (p-1)(q-1)  */
  bigint psq;                   /* p^2             */
  bigint qsq;                   /* q^2             */

#if _PAILLIER_CRT_
  // Pre-computations
  bigint rp;                    /* q^{-1} % p */
  bigint rq;                    /* p^{-1} % q */

  bigint two_p;			/* 2^|p| */
  bigint two_q;			/* 2^|q| */

  bigint lp;			/* p^{-1} % 2^|p| */
  bigint lq;			/* q^{-1} % 2^|q| */

  bigint hp;			/* Lp (g^a % p^2) ^ {-1} % p */
  bigint hq;			/* Lq (g^a % q^2) ^ {-1} % q */
#else
  bigint two_n;                 /* 2^|n| */
  bigint ln;                    /* n^{-1} % 2^|n| */
  bigint hn;                    /* Ln (g^k % n^2) ^ {-1} % n */
#endif

  void init ();

  void CRT (bigint &, bigint &) const;
  void D   (bigint &, const bigint &) const;

public:
  paillier_priv (const bigint &pp, const bigint &qq, const bigint *nn = NULL);
  paillier_priv (const bigint &pp, const bigint &qq, 
		 const bigint &aa, const bigint &gg, 
		 const bigint &kk, const bigint *nn = NULL);
  virtual ~paillier_priv () {}

  // Use the slower version, yet more standard cryptographic assumptions
  static ptr<paillier_priv> make (const bigint &p, const bigint &q);
  // Use the fast decryption version
  static ptr<paillier_priv> make (const bigint &p, const bigint &q,
				  const bigint &a);

  virtual str decrypt (const crypt_ctext &msg, size_t msglen,
		       bool recover = true) const;
};


// Paillier without fast decryption
paillier_priv paillier_skeygen  (size_t nbits, u_int iter = 32);

// Paillier with fast decryption
// abits gives length of subgroup [e.g., 160 bits for 1024-bit keys]
paillier_priv paillier_keygen  (size_t nbits, size_t abits, u_int iter = 32);

#endif /* !_PAILLIER_H_  */
