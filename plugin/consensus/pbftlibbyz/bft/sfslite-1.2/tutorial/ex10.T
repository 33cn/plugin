// -*-c++-*-
/* $Id: ex10.T 2572 2007-03-20 16:25:23Z max $ */


#include "tame.h"
#include "parseopt.h"
#include "ex_prot.h"
#include "arpc.h"

template<class T, size_t X>
class my_host_t {
public:
  my_host_t (str h, int p) : h (h), port (p), XX (X) {}
  
  void try_it (cbb cceoc, CLOSURE);
  
  T foo;
  str h;
  int port;
  size_t XX;
};
 
tamed template<class T, size_t X> void
my_host_t<T,X>::try_it (cbb cv)
{
  tvars {
    bool ret (false);
    int fd, n (3);
    ptr<axprt_stream> x;
    ptr<aclnt> cli;
    vec<int> rv;
    vec<clnt_stat> ev;
    int i;

    /*
     * these work but are not used so therefore generate compiler warnings
     *
     T foo;
     size_t XXX (X);
    */
  }

  twait { tcpconnect (h, port, mkevent(fd)); }
 
  if (fd < 0) {
    warn ("%s:%d: connection failed: %m\n", h.cstr(), port);
  } else {
    x = axprt_stream::alloc (fd);
    cli = aclnt::alloc (x, ex_prog_1);
 
    ev.setsize (n);
    rv.setsize (n);
 
    twait {
      for (i = 0; i < n; i++) {
        cli->call (EX_RANDOM, NULL, &rv[i], mkevent(ev[i]) );
      }
    }
    ret = true;
    // check for n-fold success as usual
    for (int i = 0 ; i < n; i++) {
      if (ev[i]) {
        warn << "A failure: " << ev[i] << "\n";
        ret = false;
      } else {
         warn << "Result " << i << ": " << rv[i] << "\n";
      }
    }
  }
  TRIGGER(cv,ret);
  fd ++;
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

  my_host_t<int, 5500> *h = New my_host_t<int, 5500> (argv[1], port);
  
  h->try_it (wrap (finish));
  amain ();
}
