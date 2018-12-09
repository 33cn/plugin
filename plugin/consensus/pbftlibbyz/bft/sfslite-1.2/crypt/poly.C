// -*-c++-*-
/* $Id: poly.C 1121 2005-11-01 16:30:49Z max $ */

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

#include "poly.h"

static const bigint zero   = 0;
static const bigint one    = 1;
static const bigint negone = -1;

// Polynomial interpolation for coefficients for random polynomial with supplied roots
void
polynomial::interpolate_coeffs (const vec<bigint> &x, const vec<bigint> &y)
{
  // UNTESTED

  // degree-n polynomial has n+1 coefficients
  u_int deg = x.size (), deg1 = deg - 1;
  assert (y.size () == deg);

  vec<bigint> t;
  t.setsize (deg);
  coeffs.setsize (deg);

  u_int i,j;

  for (i=0; i < deg; i++) {
    coeffs[i] = zero;
    t[i]  = zero;
  }

  for (i=0; i < deg; i++) {
    for (j=deg1-i; j < deg1; j++)
      t[j] -= (x[i] * t[j+1]);
    t[deg1] -= x[i];
  }

  bigint deriv, rderiv, accum;
  for (i=0; i < deg; i++) {
    deriv = (int) deg;
    for (j=deg1; j; j--) {
      // deriv = (j * t[j]) + (x[i] * deriv);
      deriv *= x[i];
      deriv += (j * t[j]);
    }

    // rderiv = (y[i] / deriv);
    rderiv  = y[i];
    rderiv /= deriv;

    accum   = one;
    for (int k=(int) deg1; k >= 0; k--) {
      coeffs[k] += (accum * rderiv);
      // accum = t[k] + (x[i] * accum);
      accum *= x[i];
      accum += t[k];
    }
  }
}

void
polynomial::generate_coeffs (const vec<bigint> &roots)
{
  // degree-n polynomial has n+1 coefficients, with coeffs[0] as constant
  size_t deg = roots.size () + 1;

  coeffs.clear ();
  coeffs.setsize (deg);

  u_int i,j;

  coeffs[0] = one;
  for (i=1; i < deg; i++)
    coeffs[i] = zero;

  for (i=1; i < deg; i++) {
    coeffs[i] = coeffs[i-1];

    for (j=i-1; j; j--) {
      // coeffs[j] = coeffs[j-1] - (coeffs[j] * roots[i-1]);
      coeffs[j] *= roots[i-1];
      coeffs[j] *= negone;
      coeffs[j] += coeffs[j-1];
    }
    coeffs[0] *= roots[i-1];
    coeffs[0] *= negone;
  }
}

void
polynomial::generate_coeffs (const vec<bigint> &roots, const bigint &modulus)
{
  // degree-n polynomial has n+1 coefficients, with coeffs[0] as constant
  int deg1 = roots.size ();
  int deg  = deg1 + 1;

  coeffs.clear ();
  coeffs.setsize (deg);

  int i,j;

  coeffs[0] = one;
  for (i=1; i < deg; i++)
    coeffs[i] = zero;

  for (i=1; i < deg; i++) {
    coeffs[i] = coeffs[i-1];

    for (j=i-1; j; j--) {
      // coeffs[j] = coeffs[j-1] - (coeffs[j] * roots[i-1]);
      bigint &coeff = coeffs[j];
      coeff *= roots[i-1];
      coeff %= modulus;
      coeff *= negone;
      coeff += coeffs[j-1];
      coeff %= modulus;
    }

    coeffs[0] *= roots[i-1];
    coeffs[0] *= negone;
    coeffs[0] %= modulus;
  }
}

// Use Horner's method to compute polynomial evaluation, i.e., 
//    p = c[0] + x (c[1] + x (...x (c[n-1] + x (c[n]))...))
void
polynomial::evaluate (bigint &y, const bigint &x) const
{
  size_t deg = coeffs.size ();

  y = coeffs[deg];
  for (int i=(int) deg-1; i >= 0; i--) {
    y *= x;
    y += coeffs[i];
  }
}


// Use Horner's method to compute polynomial evaluation, i.e., 
//    p = c[0] + x (c[1] + x (...x (c[n-1] + x (c[n]))...))
void
polynomial::evaluate (bigint &y, const bigint &x, const bigint &modulus) const
{
  size_t deg = coeffs.size ();

  y = coeffs[deg];
  for (int i=(int) deg-1; i >= 0; i--) {
    y *= x;
    y %= modulus;
    y += coeffs[i];
  }
  y %= modulus;
}


const strbuf & 
strbuf_cat (const strbuf &sb, const polynomial &P)
{
  const vec<bigint> coeffs = P.coefficients ();
  size_t len = coeffs.size ();
  if (!len)
    return sb;

  for (size_t i=0; i < len-1; i++) {
    strbuf_cat (sb, coeffs[i]);
    strbuf_cat (sb, ",");
  }
  return strbuf_cat (sb, coeffs[len-1]);
}
