// -*-c++-*-
/* $Id: ex11.T 2669 2007-04-02 19:16:53Z max $ */

#include "tame.h"
#include "arpc.h"
#include "parseopt.h"
#include "ex_prot.h"

struct rpc_t {
  rpc_t (u_int32_t i) : input (i) {}
  u_int32_t input;
  u_int32_t output;
  clnt_stat err;
};

static void 
launch_call (rendezvous_t<ptr<rpc_t> > *rv, u_int i, ptr<aclnt> cli)
{
  ptr<rpc_t> r = New refcounted<rpc_t> (i);
  cli->call (EX_RANDOM2, &r->input, &r->output, mkevent(*rv, r, r->err));
}

tamed static 
bool dostuff (str h, int port, cbb cb)
{
  // declare all of your "stack" variables here
  tvars {
    int fd;
    ptr<axprt_stream> x;
    ptr<aclnt> cli;
    int n_tot (100), window_sz (20), i;
    bool err_occurred (false);
    rendezvous_t<ptr<rpc_t> > window;
    bool ret (true);
    vec<ptr<rpc_t> > bad_rpcs;
    ptr<rpc_t> return_rpc;
  }

  DEFAULT_RETURN { return ret; } 

  // Call tcpconnect, and block until it returns; when it does return,
  // assign the local variable 'fd' to the result.
  twait { tcpconnect (h, port, mkevent(fd)); }

  if (fd < 0) {
    warn ("%s:%d: connection failed: %m\n", h.cstr(), port);
    err_occurred = true;
  } else {
    x = axprt_stream::alloc (fd);
    cli = aclnt::alloc (x, ex_prog_1);

    for (i = 0; i < window_sz && i < n_tot; i++) {
      launch_call (&window, i, cli);
    }

    // a wait is needed for every call that returns
    while (window.need_wait ()) {

      // Block until a call has returned; when it has, get
      // the index variable associated with the call, and assign
      // it to 'return_id' for the scope of the JOIN block.
      twait (window, return_rpc);
      if (return_rpc->err) {
	warn << "RPC error " << return_rpc->err 
	     << " occurred on input " << return_rpc->input << "\n";
	bad_rpcs.push_back (return_rpc);
      } else {
	warn << "Success: " << return_rpc->input << " -> "
	     << return_rpc->output << "\n";
      }
      if (i < n_tot) 
	launch_call (&window, i++, cli);
    }
    warn << "All done...\n";
    warn << bad_rpcs.size () << " bad RPCs, total\n";
  }
  TRIGGER(cb, !err_occurred);
}

static void finish (bool rc)
{
  exit (rc ? 0 : -1);
}

int
main (int argc, char *argv[])
{
  int port;
  if (argc != 3 || !convertint (argv[2], &port))
    fatal << "usage: ex2 <hostname> <port>\n";
  
  dostuff (argv[1], port, wrap (finish));
  amain ();
}
