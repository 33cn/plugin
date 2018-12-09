
// -*-c++-*-
/* $Id: ex1.T 2236 2006-09-29 00:00:22Z max $ */

#include "tame.h"

tamed static void
sleeper_ev (cbi cb)
{
  warn << "++ sleeper_ev: to sleep!\n";
  twait { delaycb (3, 0, mkevent ()); }
  warn << "-- sleeper_ev: woke up!\n";
  TRIGGER (cb, 40);
}

static int
sleeper_thr ()
{
  warn << "++ sleeper_thr: to sleep!\n"; 
  sleep (4);
  warn << "-- sleeper_thr: woke up!\n";
  return 10;
}

static void
run_thr ()
{
  int r1,r2;
  warn << "+ run_thr\n";
  twait {
    sleeper_ev (mkevent (r1));
    tfork (r2, wrap (sleeper_thr));
  }
  warn << "- run_thr (" << r1 << "," << r2 << ")\n";

}

tamed static void
run_ev (cbv cb)
{
  tvars {
    int r1, r2;
  }
  warn << "+ run_ev\n";
  twait {
    sleeper_ev (mkevent (r1));
    tfork (r2, wrap (sleeper_thr));
  }
  warn << "- run_ev (" << r1 << "," << r2 << ")\n";
  TRIGGER (cb);
}

tamed static void
run ()
{
  twait { run_ev (mkevent ()); }
  twait { tfork (wrap (run_thr)); }
  twait { run_ev (mkevent ()); }
  twait { tfork (wrap (run_thr)); }
  exit (0);
}

int main (int argc, char *argv[])
{
  run ();
  amain ();
}
