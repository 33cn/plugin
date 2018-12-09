
// -*-c++-*-
/* $Id: ex12.T 2548 2007-03-05 02:02:48Z max $ */

#include "tame.h"
#include "parseopt.h"

class foo_t {
public:
  foo_t (int i, evv_t ev) : _i (i), _j (0)
  {
    ev->trigger ();
  }
  void foo () { warn << "I=" << _i << "; J=" << _j++ << "\n"; }
private:
  int _i;
  int _j;
};

tamed static void
main2 ()
{
  tvars {
    int i;
    ptr<foo_t> foo;
  }

  for (i = 0; i < 100; i++) {
    twait {
      foo = New refcounted<foo_t> (i, mkevent ()); 
      foo->foo ();
    }
    foo->foo ();
  }
  exit (0);
}

int main(int argc, char *argv[])
{
  main2 ();
  amain ();
}
