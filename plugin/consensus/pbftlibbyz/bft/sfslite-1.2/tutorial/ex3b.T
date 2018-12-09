// -*-c++-*-
/* $Id: ex3b.T 2272 2006-10-27 02:25:59Z max $ */

#include "tame.h"
#include "parseopt.h"
#include "ex_prot.h"
#include "arpc.h"
 
tamed static 
void try_rpc (str h, int port, cbb cb)
{
  tvars {
    bool ret (false);
    int fd, n (5), i;
    ptr<axprt_stream> x;
    ptr<aclnt> cli;
    vec<int> rv;
    vec<clnt_stat> ev;
    rendezvous_t<> rzv;
  }
  
  twait { tcpconnect (h, port, mkevent(fd)); }
 
  if (fd < 0) {
    warn ("%s:%d: connection failed: %m\n", h.cstr(), port);
  } else {
    x = axprt_stream::alloc (fd);
    cli = aclnt::alloc (x, ex_prog_1);
 
    ev.setsize (n);
    rv.setsize (n);
 
    
    for (i = 0; i < n; i++) {
      cli->call (EX_RANDOM, NULL, &rv[i], mkevent(rzv,ev[i]) );
    }
    while (rzv.n_triggers_left ()) {
      twait (rzv);
    }

    ret = true;
    // check for n-fold success as usual
    for (i = 0 ; i < n; i++) {
      if (ev[i]) {
        warn << "A failure: " << ev[i] << "\n";
        ret = false;
      } else {
         warn << "Result " << i << ": " << rv[i] << "\n";
      }
    }
  }
  TRIGGER (cb, ret);
}
 
static void finish (bool rc)
{
  exit (rc ? 0 : -1);
}
 
int main (int argc, char *argv[])
{
  int port;
  if (argc != 3 || !convertint (argv[2], &port))
    fatal << "usage: ex3 <hostname> <port>\n";
  
  try_rpc (argv[1], port, wrap (finish));
  amain ();
}
