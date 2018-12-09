// -*-c++-*-
/* $Id: ex1.T 2670 2007-04-02 20:23:42Z max $ */

#include "tame.h"
#include "parseopt.h"
#include "tame_typedefs.h"
#include "tame_pc.h"

tamed static void
delayed_produce (tame::pc_t<int> *pc, int v)
{
  tvars {
    int t (4);
  }
  warn << "+ delay for " << t << "sec\n";
  twait { delaycb (t, 0, mkevent ()); }
  warn << "- end delay!\n";
  pc->produce (v);
}
 
tamed static void 
test_pc (evv_t ev)
{
  tvars {
    tame::pc_t<int> pc;
    int v;
  }

  pc.produce (10);
  pc.produce (20);
  twait { pc.consume (mkevent (v)); }
  twait { pc.consume (mkevent (v)); }

  delayed_produce (&pc, 30);

  twait { pc.consume (mkevent (v)); }
  warn << "consumed: " << v << "\n";
  twait { delaycb (0, 0, mkevent ()); }
  
  ev->trigger ();
}

tamed static void
main2 ()
{
  twait { test_pc (mkevent ()); }
  exit (0);
}
 
int main (int argc, char *argv[])
{
  main2 ();
  amain ();
}
