// -*-c++-*-
/* $Id: rsa.h 1117 2005-11-01 16:20:39Z max $ */

/*
 *
 * Copyright (C) 2005 Kevin Fu (fubob@mit.edu)
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

#ifndef _RSA_H_
#define _RSA_H_ 1

#include "bigint.h"

class rsa_pub {
public:
  const bigint n;		/* Modulus */
  const bigint e;               /* Encryption exponent */

protected:
  const int nbits;

public:
  rsa_pub (const bigint &nn)
    : n (nn), e (3), nbits (max ((int) n.nbits () - 1, 0)) 
  {
    // warn << "e " << e.getstr () << "\n";
    // warn << "n " << n.getstr () << "\n";
  }
  const bigint &modulus () const { return n; }
  const bigint &exponent () const { return e; }

  bigint encrypt (const str &msg) const {
    bigint m = pre_encrypt (msg, nbits);
    if (!m)
      return 0;
    m = encrypt (m);
    return m;
  }

  bigint encrypt (const bigint &m) const {
    if (!m || static_cast<int>(m.nbits ()) > nbits)
      return 0;
    bigint c = powm (m, e, n);
    return c;
  }
};

class rsa_priv : public rsa_pub {
public:
  const bigint p;		/* Smaller prime */
  const bigint q;		/* Larger prime */
  bigint phin;            /* phi (n) */
  bigint d;               /* Decryption exponent */
  bigint dp, dq;          /* Cache d mod (p-1) and d mod (q-1) for CRT */
  bigint pinvq;           /* p^-1 mod q for CRT */

protected:
  void init ();

public:
  rsa_priv (const bigint &, const bigint &);
  static ptr<rsa_priv> make (const bigint &n1, const bigint &n2);

  str decrypt (const bigint &msg, size_t msglen) const {
    bigint m = decrypt (msg);
    return post_decrypt (m, msglen, nbits);
  }

  bigint decrypt (const bigint &c) const {
    bigint m;
    bigint mp = powm (c, dp, p);
    bigint mq = powm (c, dq, q);
    bigint v = mod ((mq - mp) * pinvq, q);
    m = mp + p*v;
    // m = powm (c, d, n); 
    return m;
  }
};

rsa_priv rsa_keygen (size_t nbits);

/*
 * Serialized format of a rsa private key:
 *
 * The private key itself is stored using the following XDR data structures:
 *
 * struct keyverf {
 *   asckeytype type;  // SK_RSA_EKSBF
 *   bigint pubkey;    // The modulus of the public key
 * };
 *
 * struct privkey {
 *   bigint p;         // Smaller prime of secret key
 *   bigint q;         // Larger prime of secret key
 *   sfs_hash verf;    // SHA-1 hash of keyverf structure
 * };
 *
 * Option 1:  The secret key is stored without a passphrase
 *
 * "SK" SK_RSA_EKSBF ",," privkey "," pubkey "," comment
 *
 *   SK_RSA_EKSBF - the number 1 in ascii decimal
 *          privkey - a struct privkey, XDR and armor64 encoded
 *           pubkey - the public key modulus in hex starting "0x"
 *          comment - an arbitrary and possibly empty string
 *
 * Option 2:  There is a passphrase
 *
 * "SK" SK_RSA_EKSBF "," rounds "$" salt "$" ptext "," seckey "," pubkey
 *   "," comment
 *
 *           rounds - the cost paraeter of eksblowfish in ascii decimal
 *             salt - a 16 byte random salt for eksblowfish, armor64 encoded
 *            ptext - arbitrary length string of the user's choice
 *        secretkey - A privkey struct XDR encoded, with 4 null-bytes
 *                    appended if neccesary to make the size a multiple
 *                    of 8 bytes, encrypted once with eksblowfish,
 *                    then armor64 encoded
 *       
 */

class rsa_priv;

const size_t SK_RSA_SALTBITS = 1024;

#endif /* !_RSA_H_  */
