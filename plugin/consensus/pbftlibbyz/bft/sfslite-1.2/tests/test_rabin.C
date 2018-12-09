/* $Id: test_rabin.C 2612 2007-03-23 19:16:50Z max $ */

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
#include "bench.h"

u_int64_t vtime;
u_int64_t signtime;


void
test_key_encrypt (rabin_priv &sk)
{
  u_int64_t tmp1, tmp2, tmp3;
  for (int i = 0; i < 50; i++) {
    size_t len = rnd.getword () % (mpz_sizeinbase2 (&sk.n) / 8 - 34) + 1;
    wmstr wmsg (len);
    rnd.getbytes (wmsg, len);
    str msg1 = wmsg;
    bigint m = sk.encrypt (msg1);
    str msg2 = sk.decrypt (m, len);
    if (!msg2 || msg1 != msg2)
      panic << "Decryption failed\n"
	    << "  p = " << sk.p << "\n"
	    << "  q = " << sk.q << "\n"
	    << "msg = " << hexdump (msg1.cstr (), msg1.len ()) << "\n"
	    << " cr = " << m << "\n";

    tmp1 = get_time ();
    m = sk.sign (msg1);
    tmp2 = get_time ();
    if (!sk.verify (msg1, m))
      panic << "Verify failed\n"
	    << "  p = " << sk.p << "\n"
	    << "  q = " << sk.q << "\n"
	    << "msg = " << hexdump (msg1.cstr (), msg1.len ()) << "\n"
	    << "sig = " << m << "\n";
    tmp3 = get_time ();

    vtime += (tmp3 - tmp2);
    signtime += (tmp2 - tmp1);

    int bitno = rnd.getword () % mpz_sizeinbase2 (&m);
    m.setbit (bitno, !m.getbit (bitno));
    if (sk.verify (msg1, m))
      panic << "Verify should have failed\n"
	    << "  p = " << sk.p << "\n"
	    << "  q = " << sk.q << "\n"
	    << "msg = " << hexdump (msg1.cstr (), msg1.len ()) << "\n"
	    << "sig = " << m << "\n";

    if (len > mpz_sizeinbase2 (&sk.n) / 8 - 38) {
      len = rnd.getword () % (mpz_sizeinbase2 (&sk.n) / 8 - 38) + 1;
      msg1 = substr (msg1, 0, len);
    }

    m = sk.sign_r (msg1);
    msg2 = sk.verify_r (m, len);
    if (!msg2 || msg1 != msg2)
      panic << "Verify_r failed\n"
	    << "  p = " << sk.p << "\n"
	    << "  q = " << sk.q << "\n"
	    << "msg = " << hexdump (msg1.cstr (), msg1.len ()) << "\n"
	    << " cr = " << m << "\n";
    bitno = rnd.getword () % mpz_sizeinbase2 (&m);
    m.setbit (bitno, !m.getbit (bitno));
    if (sk.verify_r (m, len))
      panic << "Verify_r should have failed\n"
	    << "  p = " << sk.p << "\n"
	    << "  q = " << sk.q << "\n"
	    << "msg = " << hexdump (msg1.cstr (), msg1.len ()) << "\n"
	    << "sig = " << m << "\n";
  }
}

int
main (int argc, char **argv)
{
  vtime = signtime = 0;
  bool opt_v = false;
  int vsz = 1280;
  if (argc > 1 && !strcmp (argv[1], "-v")) {
    opt_v = true;
  }

  setprogname (argv[0]);
  random_update ();
  for (int i = 0; i < 10; i++) {
    rabin_priv sk = rabin_keygen (opt_v ? vsz : 424 + rnd.getword () % 256);
    test_key_encrypt (sk);
  }
  if (opt_v) {
    warn ("Signed 500 messages with %d bit key in %" U64F "u " 
	  TIME_LABEL " per signature\n", vsz, signtime / 500);
    warn ("Verified 500 messages with %d bit key in %" U64F "u " 
	  TIME_LABEL " per verify\n", vsz, vtime / 500);
  }
  return 0;
  return 0;
}
