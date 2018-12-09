// -*-c++-*-
/* $Id: ex1.T 2670 2007-04-02 20:23:42Z max $ */

#include "tame.h"
#include "parseopt.h"

static void
foo (evi_t ev)
{
  ev->trigger (10);
}

tamed static void 
main2()
{
  tvars {
    int i;
  }

  warn << "Cb1\n";
  twait { delaycb (1, 0, mkevent ()); }
  warn << "Cb2\n";
  twait { delaycb (1, 0, mkevent ()); }
  warn << "Cb3\n";


  for (i = 0; i < 1000; i++) {
    twait { delaycb (0, 0, mkevent ()); }
  }

  twait { foo (mkevent (i)); }

  exit (0);
}

int main (int argc, char *argv[])
{
  main2 ();
  amain ();
}
