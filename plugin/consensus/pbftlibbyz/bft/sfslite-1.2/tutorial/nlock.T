// -*-c++-*-
/* $Id: test.T 2468 2007-01-02 22:07:27Z max $ */

#include "async.h"
#include "tame.h"
#include "qhash.h"
#include "tame_connectors.h"
#include "parseopt.h"
#include "tame_nlock.h"

tame::lock_table_t<str> tab;

tamed static void 
main2 (int argc, char **argv)
{
  tvars {
    ptr<tame::lock_handle_t<str> > lh;
  }
  twait { tab.acquire (&lh, "foo", tame::lock_t::EXCLUSIVE, mkevent ()); }
  lh = NULL;
}

int main (int argc, char *argv[])
{
  setprogname (argv[0]);
  main2 (argc, argv);
  amain ();
}
