// -*-c++-*-
/* $Id: test.T 2468 2007-01-02 22:07:27Z max $ */

#include "async.h"
#include "tame.h"
#include "tame_connectors.h"
#include "qhash.h"

int short_to = 3;
int long_to = 6;

static void usage ()
{
  warnx << "usage: " << progname << " [-t|-f]\n";
  exit (-1);
}

tamed static void
turtle (int to, evv_t cb)
{
  tvars {
    outcome_t oc;
  }
  warn << "turtle: about to wait " << to << " seconds...\n";
  twait { delaycb (to, 0, connector::cnc (mkevent (), cb, &oc)); }
  warn << "turtle: Outcome flag: " << int (oc) << "\n";
  cb->trigger ();
}


tamed static void
expriment (bool turtlewins, bool fix, evv_t ev)
{
  tvars {
    rendezvous_t<bool> rv (__FILE__, __LINE__);
    bool timeout;
    int tto, dto;
  }
  if (turtlewins) {
    tto = short_to;
    dto = long_to;
  } else {
    tto = long_to;
    dto = short_to;
  }

  warn << "expriment: delay is waiting " << dto << " seconds...\n";
  delaycb (dto, 0, mkevent (rv, true));
  turtle (tto, mkevent (rv, false));
  twait (rv, timeout);
  warn << "experiment: finished experiment with result: " << timeout << "\n";
  ev->trigger ();
  if (fix) {
    warn << "experiment: calling cancel...\n";
    rv.cancel ();
  }
}

tamed static void 
main2 (int argc, char **argv)
{
  tvars {
    int ch;
    bool fix (false);
    bool turtlewins (false);
  }

  while ((ch = getopt (argc, argv, "tf")) != -1) {
    switch (ch) {
    case 'f':
      fix = true;
      break;
    case 't':
      turtlewins = true;
      break;
    default:
      usage ();
      break;
    }
  }

  twait { expriment (turtlewins, fix, mkevent ()); }
  warn << "Experiment completed....\n";
  twait { delaycb (short_to + 2, 0, mkevent ()); }
  warn << "Exiting....\n";
  exit (0);
}

int main (int argc, char *argv[])
{
  setprogname (argv[0]);
  main2 (argc, argv);
  amain ();
}
