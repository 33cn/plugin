// -*-c++-*-
/* $Id: rabin.h 1117 2005-11-01 16:20:39Z max $ */

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

#ifndef _RABIN_H_
#define _RABIN_H_ 1

#include "bigint.h"
#include "sha1.h"
#include "blowfish.h"

bigint pre_encrypt (str msg, size_t nbits);
str post_decrypt (const bigint &m, size_t msglen, size_t nbits);
bigint pre_sign (sha1ctx *sc, size_t nbits);
bool post_verify (sha1ctx *sc, const bigint &s, size_t nbits);
bigint pre_sign_r (str msg, size_t nbits);
str post_verify_r (const bigint &s, size_t msglen, size_t nbits);

class rabin_pub {
public:
  const bigint n;		/* Modulus */

protected:
  const int nbits;

  bool E1 (bigint &, const bigint &) const;
  void E2 (bigint &, const bigint &) const;
  void D1 (bigint &, const bigint &) const;

public:
  rabin_pub (const bigint &nn)
    : n (nn), nbits (max ((int) n.nbits () - 5, 0)) {}
  const bigint &modulus () const { return n; }

  bigint encrypt (const str &msg) const {
    bigint m = pre_encrypt (msg, nbits);
    if (!m || !E1 (m, m))
      return 0;
    E2 (m, m);
    return m;
  }
  bool verify (const str &msg, const bigint &s) const {
    bigint m;
    E2 (m, s);
    D1 (m, m);
    sha1ctx sc;
    sc.update (msg.cstr (), msg.len ());
    return post_verify (&sc, m, nbits);
  }
  str verify_r (const bigint &s, size_t msglen) const {
    bigint m;
    E2 (m, s);
    D1 (m, m);
    return post_verify_r (m, msglen, nbits);
  }
};

class rabin_priv : public rabin_pub {
public:
  const bigint p;		/* Smaller prime */
  const bigint q;		/* Larger prime */

protected:
  bigint u;			/* q^(-1) mod p */
  bigint kp;			/* (((p-1)(q-1)+4)/8) % p-1 */
  bigint kq;			/* (((p-1)(q-1)+4)/8) % q-1 */

  void init ();

  void D2 (bigint &, const bigint &, int rsel = 0) const;

public:
  rabin_priv (const bigint &, const bigint &);
  static ptr<rabin_priv> make (const bigint &n1, const bigint &n2);

  str decrypt (const bigint &msg, size_t msglen) const {
    bigint m;
    D2 (m, msg);
    D1 (m, m);
    return post_decrypt (m, msglen, nbits);
  }
  bigint sign (const str &msg) const {
    sha1ctx sc;
    sc.update (msg.cstr (), msg.len ());
    bigint m = pre_sign (&sc, nbits);
    E1 (m, m);
    D2 (m, m, rnd.getword ());
    return m;
  }
  bigint sign_r (const str &msg) const {
    bigint m = pre_sign_r (msg, nbits);
    E1 (m, m);
    D2 (m, m, rnd.getword ());
    return m;
  }
};

rabin_priv rabin_keygen (size_t nbits, u_int iter = 32);

/*
 * Serialized format of a rabin private key:
 *
 * The private key itself is stured using the following XDR data structures:
 *
 * struct keyverf {
 *   asckeytype type;  // SK_RABIN_EKSBF
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
 * "SK" SK_RABIN_EKSBF ",," privkey "," pubkey "," comment
 *
 *   SK_RABIN_EKSBF - the number 1 in ascii decimal
 *          privkey - a struct privkey, XDR and armor64 encoded
 *           pubkey - the public key modulus in hex starting "0x"
 *          comment - an arbitrary and possibly empty string
 *
 * Option 2:  There is a passphrase
 *
 * "SK" SK_RABIN_EKSBF "," rounds "$" salt "$" ptext "," seckey "," pubkey
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

class rabin_priv;

enum asckeytype {
  SK_ERROR = 0,			// Keyfile corrupt
  SK_RABIN_EKSBF = 1,		// Rabin secret key encrypted with eksblowfish
};

const size_t SK_RABIN_SALTBITS = 1024;

inline str
file2wstr (str path)
{
  return str2wstr (file2str (path));
}

#endif /* !_RABIN_H_  */
