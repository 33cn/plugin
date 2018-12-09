/* $Id: test_esign.C 2612 2007-03-23 19:16:50Z max $ */

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
#include "esign.h"
#include "bench.h"

u_int64_t vtime;
u_int64_t signtime;

void
test_key_sign (esign_priv &sk)
{
  u_int64_t tmp, tmp2, tmp3;
  bool ret;
  for (int i = 0; i < 50; i++) {
    size_t len = rnd.getword () % 256;
    wmstr wmsg (len);
    rnd.getbytes (wmsg, len);
    str msg1 = wmsg;

    tmp = get_time ();
    bigint m = sk.sign (msg1);
    tmp2 = get_time ();
    ret = sk.verify (msg1, m);
    tmp3 = get_time ();

    vtime += (tmp3 - tmp2);
    signtime += (tmp2 - tmp);

    if (!ret)
      panic << "Verify failed\n"
	    << "  p = " << sk.p << "\n"
	    << "  q = " << sk.q << "\n"
	    << "msg = " << hexdump (msg1.cstr (), msg1.len ()) << "\n"
	    << "sig = " << m << "\n";
    int bitno = rnd.getword () % mpz_sizeinbase2 (&m);
    m.setbit (bitno, !m.getbit (bitno));
    if (sk.verify (msg1, m))
      panic << "Verify should have failed\n"
	    << "  p = " << sk.p << "\n"
	    << "  q = " << sk.q << "\n"
	    << "msg = " << hexdump (msg1.cstr (), msg1.len ()) << "\n"
	    << "sig = " << m << "\n";
  }
}

int
main (int argc, char **argv)
{
  setprogname (argv[0]);
  random_update ();

  vtime = signtime = 0;
  int sz = 2048;

  bool opt_v = false;

  if (argc > 1 && !strcmp (argv[1], "-v"))
    opt_v = true;
  if (argc > 2  && !(sz = atoi (argv[2]))) 
    fatal << "bad argument\n";

  for (int i = 0; i < 10; i++) {
    esign_priv sk = esign_keygen (opt_v ? sz : 424 + rnd.getword () % 256);
    test_key_sign (sk);
  }

  if (opt_v) {
    warn ("Signed 500 messages with %d bit key in %" U64F "u " 
	  TIME_LABEL " per signature\n", sz, signtime / 500);
    warn ("Verified 500 messages with %d bit key in %" U64F "u " 
	  TIME_LABEL " per verify\n", sz, vtime / 500);
  }
  return 0;
}

void
dump (bigint *bi)
{
  warn << *bi << "\n";
}
