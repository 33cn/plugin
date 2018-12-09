/* $Id: elgamal.C,v 1.2 2006/03/02 02:15:07 mfreed Exp $ */

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

#include "crypt_prot.h"
#include "prime.h"
#include "elgamal.h"
#include "fips186.h"

#define STRONG_PRIMES 1

INITFN (scrubinit);

static void
scrubinit ()
{
  mp_setscrub ();
}

elgamal_pub::elgamal_pub (const bigint &pp, 
			  const bigint &qq,
			  const bigint &gg, 
			  const bigint &ggr, 
			  size_t aabits) 
  : p  (pp), q (qq), g  (gg), gr (ggr),
    nbits (p.nbits ()),
    abits (min (aabits, nbits-1)),
    p1 (p-1), q1 (q+1)
{ 
  assert (nbits);
}


bool
elgamal_pub::encrypt (crypt_ctext *c, const bigint &msg, bool recover) const 
{
  assert (c);
  assert (c->type == CRYPT_ELGAMAL);

  if (msg >= q) {
    warn << "elgamal_pub::E: input too large [m " << msg.nbits () 
	 << " q " << q.nbits () << "]\n";
    return false;
  }

  elgamal_ctext &ec = *c->elgamal;

  bigint rand;
  do rand = random_bigint (abits);
  while (rand == 0);

  ec.r  = powm (g,  rand, p);
  ec.m  = powm (gr, rand, p);

#if STRONG_PRIMES
  if (recover)
    //X ec.m *= powm (msg + 1, 2, p);
    ec.m *= msg + 1;
  else
    ec.m *= powm (g, msg, p);
#else
  if (recover)
    ec.m *= msg + 1;
  else
    ec.m *= powm (g, msg, p);
#endif

  ec.m %= p;
  return true;
}


void
elgamal_pub::add (crypt_ctext *c, const crypt_ctext &msg1, 
		  const crypt_ctext &msg2) const
{
  assert (c);
  assert (c->type   == CRYPT_ELGAMAL);
  assert (msg1.type == CRYPT_ELGAMAL);
  assert (msg2.type == CRYPT_ELGAMAL);

  elgamal_ctext &ec        = *c->elgamal;
  const elgamal_ctext &ec1 = *msg1.elgamal;
  const elgamal_ctext &ec2 = *msg2.elgamal;
  
  ec.r = ec1.r * ec2.r;
  ec.m = ec1.m * ec2.m;
  
  ec.r %= p;
  ec.m %= p;
}


void
elgamal_pub::mult (crypt_ctext *c, const crypt_ctext &msg, 
		   const bigint &cons) const
{
  assert (c);
  assert (c->type  == CRYPT_ELGAMAL);
  assert (msg.type == CRYPT_ELGAMAL);

  elgamal_ctext &ec        = *c->elgamal;
  const elgamal_ctext &mec = *msg.elgamal;

  ec.r = powm (mec.r, cons, p);
  ec.m = powm (mec.m, cons, p);
};



elgamal_priv::elgamal_priv (const bigint &pp, 
			    const bigint &qq,
			    const bigint &gg, 
			    const bigint &rr) 
  : elgamal_pub (pp, qq, gg, powm (gg, rr, pp), rr.nbits ()), 
    r (rr),
    i2 (invert (2, q))
{
}


str 
elgamal_priv::decrypt (const crypt_ctext &msg, size_t msglen, 
		       bool recover) const
{
  // Only applicable for recoverable-encryption (see encrypt ()).
  // Yet, we can't recover the message otherwise, so let's always
  // assume encrypt() has been called with recover == true and
  // process the message accordingly.

  assert (msg.type == CRYPT_ELGAMAL);

  const elgamal_ctext &ec = *msg.elgamal;

  bigint m;
  m  = powm (ec.r, r, p);
  m  = invert (m, p);
  m *= ec.m;
  m %= p;

#if STRONG_PRIMES
  // Find quadratic residue
  if (recover) {
    //X m = powm (m, i2, p);
    //X  if (m > q) {
    //X   warnx << "> q\n\n";
    //X    m -= q;
    //X }
  }
#endif

  if (recover)
    m -= 1;

  return post_decrypt (m, msglen);
}


ptr<elgamal_priv>
elgamal_priv::make (const bigint &p, const bigint &g, const bigint &r)
{
  bigint q = (p - 1) >> 1;
  
  if (p <= 1     || !p.probab_prime (5)
      || q <= 1  || !q.probab_prime (5)
      || g <= 1  || g < p
      || r < 1   || r > (p-2))
    return NULL;

  return New refcounted<elgamal_priv> (p, q, g, r);
}


struct elgamal_gen : public fips186_gen {
  ptr<elgamal_priv> sk;

  elgamal_gen (u_int pbits, u_int iter) : fips186_gen (pbits, iter) {}

  static ptr<elgamal_priv> rgen (u_int pbits, u_int iter = 32) {
    elgamal_gen dg (pbits, iter);
    dg.gen (iter);
    return dg.sk;
  }

private:
  void gen (u_int iter) {
    bigint q, p, g, r;
    do {
      gen_q (&q);
    } while (!gen_p (&p, q, iter) || !q.probab_prime (iter));
    gen_g (&g, p, q);
    
    do r = random_zn (q);
    while (r == 0);

    sk = New refcounted<elgamal_priv> (p, q, g, r);
  }
};


elgamal_priv
elgamal_keygen (size_t nbits, size_t abits, u_int iter)
{
  assert (nbits > 0);
  assert (abits > 0);
  assert (abits <= nbits);

  random_init ();
  bigint p, q, g, r;

#if STRONG_PRIMES

  do {
    q = random_prime (nbits-1, odd_sieve, 2, iter);
    p = 2 * q + 1;
  } while (p.nbits () != nbits || !p.probab_prime (iter));

  do g = random_zn (p-1);
  while (g == 0 || g == 1);

  g *= g;
  g %= p;
  
  do r = random_bigint (abits);
  while (r == 0);

  return elgamal_priv (p, q, g, r);

#else

  ptr<elgamal_priv> psk = elgamal_gen::rgen (nbits);  
  elgamal_priv sk = *psk;
  return sk;

#endif

}

