// -*-c++-*-
/* $Id: poly.h 1121 2005-11-01 16:30:49Z max $ */

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

#ifndef _POLYNOMIAL_H_
#define _POLYNOMIAL_H_ 1

#include "vec.h"
#include "bigint.h"

// Polynomial coefficients c[0..(n-1)] stored in ascending order:
//   c0 + c1 x^1 + c2 x^2 + ... + c(n-1) x^{n-1}
class polynomial {
 private:
  vec<bigint> coeffs;

 public:
  polynomial () {}

  void interpolate_coeffs (const vec<bigint> &x, const vec<bigint> &y);  
  void generate_coeffs    (const vec<bigint> &roots);
  void generate_coeffs    (const vec<bigint> &roots, const bigint &modulus);
  void evaluate (bigint &y, const bigint &x) const;
  void evaluate (bigint &y, const bigint &x, const bigint &modulus) const;
  
  const vec<bigint> coefficients () const { return coeffs; }
};

const strbuf & strbuf_cat (const strbuf &sb, const polynomial &P);


#endif /* _POLYNOMIAL_H_ */

