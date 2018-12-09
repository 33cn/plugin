
// -*-c++-*-
/* $Id: buggy2.T 2349 2006-12-03 19:12:46Z max $ */

#include "tame.h"
#include "parseopt.h"
#include "arpc.h"

//
// Bug #4: 
//

static void usage ()
{
  fatal << "usage: " << progname << " [-f]\n";
}

tamed static void
run_test (bool fix_bug, evv_t ev)
{
  tvars {
    rendezvous_t<> rv (__FILE__, __LINE__);
  }
  delaycb (1, 0, mkevent (rv));
  delaycb (2, 0, mkevent (rv));
  twait (rv);
  ev->trigger ();
  if (fix_bug)
    twait (rv);
}

tamed static void
main2 (int argc, char **argv)
{
  tvars {
    int ch;
    bool fix_bug (false);
  }
  while ((ch = getopt (argc, argv, "f")) != -1) {
    switch (ch) {
    case 'f':
      fix_bug = true;
      break;
    default:
      usage ();
    }
  }
  twait { run_test (fix_bug, mkevent ()); }
  twait { delaycb (2, 0, mkevent ()); }
  exit (0);
}
 
int 
main (int argc, char *argv[])
{
  setprogname (argv[0]);
  main2 (argc, argv);
  amain ();
}
