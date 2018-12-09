// -*-c++-*-
/* $Id: ex9.T 2669 2007-04-02 19:16:53Z max $ */

#include "tame.h"
#include "arpc.h"
#include "parseopt.h"
#include "ex_prot.h"

struct state_t {
  state_t (str h, int p) : host (h), port (p), n_tot (40), window_sz (5) {}

  str host;
  int port;
  rendezvous_t<int> window;
  vec<int> res;
  vec<clnt_stat> errs;
  ptr<axprt_stream> x;
  ptr<aclnt> cli;
  int n_tot, window_sz;
  int i;

  void launch (cbb cv, CLOSURE);
  void relaunch (cbb cv, CLOSURE);
};

tamed void 
state_t::launch (cbb cv)
{
  // declare all of your "stack" variables here
  tvars {
    int fd;
    bool ret (true);
  }

  // Call tcpconnect, and block until it returns; when it does return,
  // assign the local variable 'fd' to the result.
  twait { tcpconnect (host, port, mkevent(fd)); }

  if (fd < 0) {
    warn ("%s:%d: connection failed: %m\n", host.cstr(), port);
    ret = false;
  } else {
    res.setsize (n_tot);
    errs.setsize (n_tot);
    x = axprt_stream::alloc (fd);
    cli = aclnt::alloc (x, ex_prog_1);

    for (i = 0; i < window_sz && i < n_tot; i++) {

      // launch an asychronous function call, but don't
      // block on its return. Associate the call with the 
      // join group 'window' and the index value 'i'
      cli->call (EX_RANDOM, NULL, &res[i], mkevent(window,i,errs[i]) );
    }
  }
  TRIGGER(cv,ret);
}

tamed void 
state_t::relaunch (cbb cv)
{
  tvars {
    bool err_occurred (false);
    int return_id;
  }
  // a wait is needed for every call that returns
  while (window.need_wait ()) {
    
    // Block until a call has returned; when it has, get
    // the index variable associated with the call, and assign
    // it to 'return_id' for the scope of the JOIN block.
    twait (window, return_id);
    if (errs[return_id]) {
      err_occurred = true;
      warn << "RPC error: " << errs[return_id] << "\n";
    } else {
      warn << "Success " << return_id << ": " << res[return_id] << "\n";
      
      // in the common case, for every call that returns,
      // another should be launched. but don't bother launching
      // more if an error occurred.
      if (i < n_tot && !err_occurred) {
	cli->call (EX_RANDOM, NULL, &res[i], mkevent (window,i, errs[i]) );
	i++;
      }
    }
  }
  warn << "All done...\n";
  TRIGGER(cv, !err_occurred);
}

tamed static 
void go (str h, int p)
{
  tvars {
    state_t s (h, p);
    bool rc;
  }
  twait { s.launch (mkevent(rc)); }
  if (rc) {
    twait { s.relaunch (mkevent(rc)); }
  }
  delaycb (0, 0, wrap (exit, rc ? 0 : -1));
}

int
main (int argc, char *argv[])
{
  int port;
  if (argc != 3 || !convertint (argv[2], &port))
    fatal << "usage: ex2 <hostname> <port>\n";

  go (argv[1], port);
  amain ();
}
