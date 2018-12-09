/* $Id: pm.C 2330 2006-11-19 18:18:00Z max $ */

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

/* Polynomial evaluation protocol for private matching, i.e.,
 * privacy-preserving two-party set intersection.  From:
 *
 *    Efficient Private Matching and Set Intersection
 *    Michael J. Freedman, Kobbi Nissim, and Benny Pinkas
 *    Proc. Advances in Cryptology -- EUROCRYPT 2004, May, 2004.
 *
 *
 * Inputs
 * ------
 *
 * Client:
 *   Set X_c[0..kc]
 *   Key pair (K,K') of an additively-homomorphic public-key cryptosystem
 *
 * Server:
 *   Set X_s[0..ks]
 *
 *
 * Protocol
 * --------
 *
 * A. Client:
 *
 *    1. Encode X_c in a degree-kc polynomial P
 *    2. Encrypt P's coefficients using public key K
 *    3. Send kc+1 encrypted coeffs to server
 *
 * B. Server:
 *
 *    1. For all x_i \in X_s,
 *       a. y'_i = Evaluation of (encrypted) P on x_i
 *       b. y_i  = random * y'_i + payload_i
 *    2. Randomly permutes ks+1 evaluations
 *    3. Return these ks+1 evaluations of P to client
 *
 * C. Client:
 *
 *    1. For all y_i in returned evaluations, 
 *       a. decrypt y_i
 *       b. check if result is a well-formed payload
 *    2. Intersection is set of well-formed payloads
 *
 */

#include "pm.h"
#include "poly.h"

static const char   match[4] = { 0xFF, 0xFF, 0xFF, 0xFF };
static const size_t matchlen = sizeof (char) * 4;
static const bigint one = 1;

bool
pm_client::set_polynomial (const vec<str> &inputs)
{
  size_t len = inputs.size ();
  if (!len)
    return false;

  // Convert strings to bigints
  vec<bigint> in;
  in.setsize (len);
  for (size_t i=0; i < len; i++)
    in[i] = sk->pre_encrypt (inputs[i]);
  
  return set_polynomial (in);
}


bool
pm_client::set_polynomial (const vec<bigint> &inputs)
{
  // A.1.
  // Compute coefficients of polynomial with roots = inputs
  polynomial P;
  P.generate_coeffs (inputs, sk->ptext_modulus ());
  const vec<bigint> pcoeffs = P.coefficients ();
  size_t kc = pcoeffs.size ();
  if (!kc)
    return false;

  // Require coefficient c[deg-1] = 1, to ensure that malicious client
  // doesn't send over generate polynomial. See FNP04, 5.1
  assert (pcoeffs[kc-1] == one);

  // A.2.
  // Encrypt polynomial coefficients
  coeffs.clear ();
  for (size_t i=0; i < kc-1; i++) {
    coeffs.push_back (crypt_ctext (sk->ctext_type ()));
    if (!sk->encrypt (&coeffs.back (), pcoeffs[i], false)) {
      coeffs.clear ();
      return false;
    }
  }
  return true;
}


void 
pm_client::decrypt_intersection (vec<str> &payloads,
				 const vec<cpayload> &plds) const
{
  for (size_t i=0, lst=plds.size (); i < lst; i++) {

    // C.1.a
    const cpayload &pld = plds[i];
    str res = sk->decrypt (pld.ctxt, pld.ptsz);

    // C.1.b
    // tests > len so that something left after stripping wellformed
    //X warnx << "dec [" << hexdump (res.cstr (), res.len ()) << "]\n";

    if (!res || res.len () <= matchlen
	|| strncmp (res.cstr (), match, matchlen))
      continue;
    
    str payload (res.cstr () + matchlen, res.len () - matchlen);
    payloads.push_back (payload);
  }
}


void 
pm_server::evaluate_intersection (vec<cpayload> *res, 
				  const vec<crypt_ctext> *ccoeffs,
				  const homoenc_pub *pk)
{
  // B.1
  assert (pk);
  crypt_ctext encone (pk->ctext_type ());
  if (!pk->encrypt (&encone, one, false))
    return;

  vec<cpayload> unshuffled;
  inputs.traverse (wrap (this, &pm_server::evaluate_polynomial, 
			 &unshuffled, ccoeffs, pk, &encone));

  // B.2
  // XXX Do a GOOD random shuffle here...
  size_t usize = unshuffled.size ();
  if (usize) {
    res->reserve (usize);
    for (size_t i=0; i < usize; i++) {
      if (rnd.getword () % 2)
	res->push_back (unshuffled.pop_front ());
      else
	res->push_back (unshuffled.pop_back ());
    }
  }
}


void
pm_server::evaluate_polynomial (vec<cpayload> *res, 
				const vec<crypt_ctext> *pccoeffs, 
				const homoenc_pub *ppk,
				const crypt_ctext *encone,
				const str &x, ppayload *payload)
{
  assert (res && pccoeffs && ppk && encone);
  const vec<crypt_ctext> &ccoeffs = *pccoeffs;
  const homoenc_pub &pk           = *ppk;
  size_t deg = ccoeffs.size ();

  // B.1.a
  // Compute E(P(y))
  bigint px = pk.pre_encrypt (x);
  if (!px)
    return;

  // Require coefficient c[deg-1] = 1, to ensure that malicious client
  // doesn't send over generate polynomial. See FNP04, 5.1
  crypt_ctext cy = *encone;

  // See polynomial::evaluate
  // Coeffs sent over already don't include last element
  while (deg) {
    // y = y * x + coeff[i];
    crypt_ctext tmp (pk.ctext_type ());
    pk.mult (&tmp, cy, px);
    pk.add  (&cy, tmp, ccoeffs[--deg]);
  }

  // B.1.b
  // Compute E(rP(x))
  pk.mult (&cy, cy, random_zn (pk.ptext_modulus ()));

  // Generate payload
  str buf = strbuf () << match << payload->ptxt;
  crypt_ctext cpay (pk.ctext_type ());

  //X warnx << "pay [" << hexdump (buf.cstr (), buf.len ()) << "]\n";

  if (!pk.encrypt (&cpay, buf, true))
    return;

  // Compute E(rP(x) + (match || payload))
  pk.add (&cy, cy, cpay);

  cpayload pay;
  pay.ctxt = cy;
  // if P(x) != 0, resulting plaintext can be > buf.len, but
  // we don't care, because the match check will fail
  pay.ptsz = buf.len ();

  res->push_back (pay);
}
