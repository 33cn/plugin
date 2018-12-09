
// -*-c++-*-
/* $Id: ex1.T 2236 2006-09-29 00:00:22Z max $ */

#include "tame.h"

tamed static void
sleeper_ev (cbi cb)
{
  warn << "++ sleeper_ev: to sleep!\n";
  twait { delaycb (3, 0, mkevent ()); }
  warn << "-- sleeper_ev: woke up!\n";
  TRIGGER (cb, 33);
}

static int
sleeper_thr ()
{
  warn << "++ sleeper_thr: to sleep!\n"; 
  sleep (8);
  warn << "-- sleeper_thr: woke up!\n";
  return 12;
}

static void
run_thr ()
{
  rendezvous_t<int> rv;
  int i = 0;
  int r1, r2;
  warn << "+ run_thr\n";

  sleeper_ev (mkevent (rv, 0, r1));
  tfork (rv, 1, r2, wrap (sleeper_thr));

  rv.wait (i);
  warn << "oo event return: " << i << "\n";
  rv.wait (i);
  warn << "oo event return: " << i << "\n";

  warn << "- run_thr (" << r1 << "," << r2 << ")\n";
}

tamed static void
run_ev (cbv cb)
{
  tvars {
    rendezvous_t<int> rv;
    int i;
    int r1, r2;
  }
  warn << "+ run_ev\n";

  sleeper_ev (mkevent (rv, 0, r1));
  tfork (rv, 1, r2, wrap (sleeper_thr));

  twait (rv, i);
  warn << "oo event return: " << i << "\n";
  twait (rv, i);
  warn << "oo event return: " << i << "\n";

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
