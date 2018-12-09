// -*-c++-*-
/* $Id: ex1.T 3807 2008-11-19 17:49:56Z max $ */

#include "tame.h"
#include "parseopt.h"
#include "qhash.h"
 
tamed static 
void try_connect (str h, int port)
{
  tvars { int fd (-1); }
  twait { tcpconnect (h, port, mkevent (fd)); }
  if (fd >= 0) {
    warn << "Connection succeeded!\n";
    exit (0);
  } else {
    warn << "Connection failed!\n";
    exit (1);
  }
}

static
void foo()
{
  bhash<int> bh;
  bh.insert (1);
  bh.insert (2);
  bh.insert (5);

  bhash_const_iterator_t<int> it (bh);
  const int *k;
  while ((k = it.next ())) {
    warn << "it: " << *k << "\n";
  }

}
 
int main (int argc, char *argv[])
{
  int port;
  if (argc != 3 || !convertint (argv[2], &port))
    fatal << "usage: ex2 <hostname> <port>\n";

  str s;
  warn << "What happens when you warn a NULL string? " << s << "\n";

  foo ();
  
  try_connect (argv[1], port);
  amain ();
}
