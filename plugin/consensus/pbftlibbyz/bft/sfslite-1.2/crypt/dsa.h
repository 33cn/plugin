// -*-c++-*-
/* $Id: dsa.h,v 1.2 2006/02/26 02:21:16 kaminsky Exp $ */

/*
 *
 * Copyright (C) 2006 Michael Kaminsky (kaminsky at csail.mit.edu)
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

#ifndef _DSA_H_
#define _DSA_H_ 1

#include "fips186.h"

struct dsa_pub {
  const bigint p;
  const bigint q;
  const bigint g;
  const bigint y;

  dsa_pub (const bigint &pp, const bigint &qq,
	   const bigint &gg, const bigint &yy)
    : p (pp), q (qq), g (gg), y (yy) {}
  virtual ~dsa_pub () {} 

  bool verify (const str &msg, const bigint &r, const bigint &s);

protected:
  bigint msghash_to_bigint (const str &msg);
};

struct dsa_priv : public dsa_pub {
  const bigint x;

  dsa_priv (const bigint &pp, const bigint &qq, const bigint &gg,
	    const bigint &yy, const bigint &xx) 
    : dsa_pub (pp, qq, gg, yy), x (xx) {}

  void sign (bigint *r, bigint *s, const str &msg);
};

struct dsa_gen : public fips186_gen {
  ptr<dsa_priv> sk;

  dsa_gen (u_int pbits, u_int iter) : fips186_gen (pbits, iter) {}
  static ptr<dsa_gen> rgen (u_int pbits, u_int iter = 32) {
    ref<dsa_gen> dg = New refcounted<dsa_gen> (pbits, iter);
    dg->gen (iter);
    return dg;
  }

private:
  void gen (u_int iter);
};

#endif /* !_DSA_H_ */
