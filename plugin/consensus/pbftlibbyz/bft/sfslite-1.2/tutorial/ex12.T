// -*-c++-*-
/* $Id: ex12.T 2548 2007-03-05 02:02:48Z max $ */

#include "tame.h"
#include "parseopt.h"
#include "ex_prot.h"

//
// A test case to make sure that biggest callbacks work.
//

typedef ptr<callback<void,int,bool,str> > cb3_t ;

tamed static void callee3 (cb3_t cv)
{
  str s = "hello";
  TRIGGER(cv, 10, true, s);
}


tamed static void main_T (int argc, char **argv)
{
  tvars {
    int i;
    bool b;
    str s;
  }
  twait { callee3 (mkevent(i,b,s)); }
  exit (0);
}

int
main (int argc, char *argv[])
{
  main_T (argc, argv);
  amain ();
}
