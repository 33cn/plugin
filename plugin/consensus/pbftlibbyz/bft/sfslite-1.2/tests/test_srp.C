/* $Id: test_srp.C 2 2003-09-24 14:35:33Z max $ */

/*
 *
 * Copyright (C) 1999 David Mazieres (dm@uun.org)
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
#include "srp.h"

#define TESTL 0xdeadbabe
#define TESTR 0x31337fac

int
main (int argc, char **argv)
{
  setprogname (argv[0]);
  srp_hash sessid;

  bigint N, g;
#if 1
  //warnx << "Generating SRP parameters...";
  err_flush ();
  srp_base::genparam (512, &N, &g);
  //warnx << "done\n";
#else
  N = "0xb554bc791c15de289b4e46b013f5802933408b3b7c5c6622b91802056a25b436acd645ab35c94718800e7409e77e9237c92fcdcdd88b07c3a5277febb81a764ac0038420b61a4b44cfc058dc34f642b0f8bc13c66da17ea624c4eb808242708a09393e85b50b8b20f59cbd3790caa291f8c7e186c175c4bc7bbf1177f066ec33";
  g = 2;
#endif

  srpmsg m;
  srp_client srpc;
  srp_server srps;

  str V = srpc.create (N, g, "Geheim", "ny.lcs.mit.edu", 5);

  u_int32_t testl = TESTL, testr = TESTR;
  srpc.eksb.encipher (&testl, &testr);

  if (srpc.init (&m, sessid, "dm", "Geheim") != SRP_NEXT)
    panic ("srp_client::init failed\n");
  if (srps.init (&m, &m, sessid, "dm", V) != SRP_NEXT)
    panic ("srp_server::init failed\n");
  if (srpc.next (&m, &m) != SRP_NEXT)
    panic ("srp_client::phase1 failed\n");
  if (srps.next (&m, &m) != SRP_NEXT)
    panic ("srp_server::phase2 failed\n");
  if (srpc.next (&m, &m) != SRP_NEXT)
    panic ("srp_client::phase3 failed\n");
  if (srps.next (&m, &m) != SRP_LAST)
    panic ("srp_server::phase4 failed\n");
  if (srpc.next (&m, &m) != SRP_DONE)
    panic ("srp_client::phase5 failed\n");

  if (srpc.host != "ny.lcs.mit.edu")
    panic ("client got the wrong host name: %s\n", srpc.host.cstr ());

  srpc.eksb.decipher (&testl, &testr);
  if (testl != TESTL || testr != TESTR)
    panic ("could not decrypt message after SRP\n");

  return 0;
}
