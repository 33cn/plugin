// -*-c++-*-
/* $Id: pm.h 2330 2006-11-19 18:18:00Z max $ */

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
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 59 Temple Place, Suite 330, Boston, MA 02111-1307
 * USA
 *
 */

#ifndef _PRIVATE_MATCHING_H_
#define _PRIVATE_MATCHING_H_ 1

#include "vec.h"
#include "qhash.h"
#include "homoenc.h"

struct cpayload {
  crypt_ctext ctxt; // ciphertext
  size_t ptsz;      // plaintext size
};
struct ppayload {
  str ptxt;         // plaintext
};

class pm_client {
 private:
  ref<homoenc_priv> sk;
  vec<crypt_ctext> coeffs;

  bool encrypt_polynomial (vec<crypt_ctext> &ccoeffs) const;

 public:
  pm_client (ref<homoenc_priv> s) : sk (s) {}

  bool set_polynomial (const vec<str> &inputs);
  bool set_polynomial (const vec<bigint> &inputs);
  const vec<crypt_ctext> & get_polynomial () const { return coeffs; }

  void decrypt_intersection (vec<str> &payloads, 
			     const vec<cpayload> &plds) const;
};


class pm_server {
 public:
  qhash<str, ppayload> inputs;
  
 private:
  void evaluate_polynomial (vec<cpayload> *res, 
			    const vec<crypt_ctext> *ccoeffs, 
			    const homoenc_pub *pk,
			    const crypt_ctext *encone,
			    const str &x, ppayload *payload);
 public:
  pm_server () {}

  void evaluate_intersection (vec<cpayload> *res, 
			      const vec<crypt_ctext> *ccoeffs,
			      const homoenc_pub *pk);
};


#endif /* _PRIVATE_MATCHING_H_ */
