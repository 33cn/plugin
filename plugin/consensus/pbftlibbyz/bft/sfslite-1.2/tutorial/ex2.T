// -*-c++-*-
/* $Id: ex2.T 3476 2008-07-25 02:13:37Z max $ */

#include "tame.h"
#include "parseopt.h"
#include "arpc.h"
#include "ex_prot.h"
#include "tame_rpc.h"

#if 0
static callbase *
foo (ptr<aclnt> c, u_int32_t rpc, const void *a, void *r, aclnt_cb cb)
{
  return c->call (rpc, a, r, cb);
}
#endif

tamed static void 
try_rpc (str h, int port, evb_t cb)
{
  tvars {
    bool ret (false);
    int32_t fd, r1;
    ptr<axprt_stream> x;
    ptr<aclnt> cli;
 
    ex_str_t r2, a2;
    ex_struct_t r3;
    clnt_stat e1, e2, e3;
    callbase *cbase;
  }

  twait { tcpconnect (h, port, mkevent(fd)); }
 
  if (fd < 0) {
    warn ("%s:%d: connection failed: %m\n", h.cstr(), port);
  } else {
    x = axprt_stream::alloc (fd);
    cli = aclnt::alloc (x, ex_prog_1);
    a2 = "go hang a salami i'm a lasagna hog";
    
    twait {
      RPC::ex_prog_1::ex_random (cli, &r1, mkevent (e1));
      RPC::ex_prog_1::ex_reverse (cli, a2, &r2, mkevent(e2));

      // Calling can be more general than just (void)call() methods.
      // Need to specify callbase* unfortunately; can't be inferred
      // from the type of tame_rpc...
      cbase = RPC::ex_prog_1::w_ex_struct<callbase *> 
	(wrap (tame_rpc::rcall, cli), &r3, mkevent(e3));
    }
    if (e1 || e2 || e3) {
      warn << "at least 1 RPC failed!\n";
    } else {
      warn << "the results are in:\n"
	   << "\trandom # = " << r1 << "\n"
	   << "\treversed string = " << r2 << "\n"
	   << "\tstupid stuct = { s = " << r3.s << "; u = " << r3.u << " }\n";
      ret = true;
    }
  }
  cb->trigger (true);
}
 
static void finish (bool rc)
{
  exit (rc ? 0 : -1);
}

tamed static void
main2 (int argc, char **argv)
{
  tvars {
    bool rc;
    int port;
  }
  if (argc != 3 || !convertint (argv[2], &port))
    fatal << "usage: ex2 <hostname> <port>\n";

  twait { try_rpc (argv[1], port, mkevent (rc)); }
  finish (rc);
}
 
int main (int argc, char *argv[])
{
  main2 (argc, argv);
  amain ();
  return 0;
}
