/* $Id: test_schnorr.C 2 2003-09-24 14:35:33Z max $ */

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


#include "crypt.h"
#include "rabin.h"
#include "schnorr.h"
#include "bench.h"

int _ts;

timeval tv;

inline void startt () { gettimeofday (&tv, NULL); }
inline int stopt () 
{ 
  timeval tv2; 
  gettimeofday (&tv2, NULL); 
  return ( (tv2.tv_sec - tv.tv_sec) * 1000000 +
	   (tv2.tv_usec - tv.tv_usec)) ;
}
int rs, rv, se, sc, sv, n, eg;

void
test_key_encrypt (rabin_priv &sk, schnorr_clnt_priv *scp,
		  schnorr_srv_priv *ssp)
{
  for (int i = 0; i < 10; i++) {
    size_t len = 512;
    wmstr wmsg (len);
    rnd.getbytes (wmsg, len);
    str msg = wmsg;

    startt ();
    ref<ephem_key_pair> ekp = scp->make_ephem_key_pair ();
    eg += stopt ();

    startt ();
    bigint rab = sk.sign (msg);
    rs += stopt ();

    startt();
    if (!sk.verify (msg, rab)) {
      panic << "verify failed.\n";
    }
    rv += stopt ();

    int bitno = rnd.getword () % mpz_sizeinbase2 (&rab);
    rab.setbit (bitno, !rab.getbit (bitno));
    if (sk.verify (msg, rab)) {
      panic << "verify should have failed\n";
    }

    bigint r_srv, s_srv, r, s;
    startt ();
    if (!ssp->endorse_signature (&r_srv, &s_srv, msg, ekp->public_half ())) {
      panic << "cannot endorse\n";
    }
    se += stopt ();
    startt ();
    if (!scp->complete_signature (&r, &s, msg, ekp->public_half (),
				  ekp->private_half (), r_srv, s_srv)) {
      panic << "cannot complete sig\n";
    }
    sc += stopt ();
    
    startt ();
    if (!scp->verify (msg, r, s)) {
      panic << "verify failed\n";
    }
    sv += stopt ();
    bitno = rnd.getword () % mpz_sizeinbase2 (&s);
    s.setbit (bitno, !s.getbit (bitno));
    if (scp->verify (msg, r, s))
      panic << "verify should have failed.\n";

    /*
    warn << "Success: " << i << "\n";
    */
    n++;
  }
}

int
main (int argc, char **argv)
{
  int sg, rg;
  rs = rv = se = sc = sv = n = sg = rg = eg = 0;
  ptr<schnorr_gen> sgt;
  setprogname (argv[0]);
  random_update ();
  int m = 10;
  
  for (int i = 0; i < m; i++) {
    startt ();
    sgt = schnorr_gen::rgen (1024);
    sg += stopt ();
    startt ();
    rabin_priv sk = rabin_keygen (1024);
    rg += stopt ();
    test_key_encrypt (sk, sgt->csk, sgt->ssk);
  }
  /*
  warnx << "n: " << n << "\n"
	<< "Rabin sign:       " << rs / n << "\n"
        << "Rabin verify:     " << rv / n << "\n"
        << "Schnorr Endorse:  " << se / n << "\n"
        << "Schnorr Complete: " << sc / n << "\n"
        << "Schnorr Verify:   " << sv / n << "\n"
	<< "Rabin generate:   " << rg / m << "\n"
	<< "Schnorr generate: " << sg / m << "\n"
	<< "Schnorr Ephem Gn: " << eg / n << "\n";
  */
  return 0;
}
