// -*-c++-*-
/* $Id: ex5.T 3276 2008-05-17 20:28:51Z max $ */

#include "tame.h"
#include "arpc.h"
#include "parseopt.h"
#include "ex_prot.h"

//
// ex5.T 
//
//   Functionally equivalent to ex4.T, but also tests support for
//   taming private class methods.
//
class remote_host_t {
public:
  remote_host_t (const str &n) : _name (n), _some_class_member (5) {}
  void dostuff (int port, cbb cb) 
  { dostuff_private (port, cb); }
private:
  void dostuff_private (int port, cbb cv, CLOSURE);
  str _name;
public:
  size_t _some_class_member;
};

// TAME_OFF
typedef enum { BLOCK_FOO = 1, NONBLOCK_FOO = 2, JOIN_FOO = 3, 
	       TAME_FOO = 5 } my_enum_t;
/* TAME_ON */

static void foo_func (my_enum_t e)
{
  /* TAME_OFF */
  switch (e) {
  case BLOCK_FOO:
    warn << "case 1\n";
    break;
  case NONBLOCK_FOO:
    warn << "case 2\n";
    break;
  case JOIN_FOO:
    warn << "case 3\n";
    break;
  case TAME_FOO:
    warn << "case 5\n";
    break;
  }
  // TAME_ON
}


tamed void 
remote_host_t::dostuff_private (int port, cbb cv)
{
  // declare all of your "stack" variables here
  tvars {
    int fd;
    ptr<axprt_stream> x;
    ptr<aclnt> cli;
    vec<int> res;
    vec<clnt_stat> errs;
    int n_tot (40), window_sz (5), sent (0), recv (0);
    bool ok (true);
    rendezvous_t<int> rv;
    int return_id;
    size_t test (_self->_some_class_member + 2);
  }

  // Test isn't useful, just to test that tvars{..} can reference
  // members of the underlying class.
  test ++;

  // Call tcpconnect, and block until it returns; when it does return,
  // assign the local variable 'fd' to the result.
  twait { tcpconnect (_name, port, mkevent(fd)); }

  if (fd < 0) {
    warn ("%s:%d: connection failed: %m\n", _name.cstr(), port);
    ok = false;
  } else {
    res.setsize (n_tot);
    errs.setsize (n_tot);
    x = axprt_stream::alloc (fd);
    cli = aclnt::alloc (x, ex_prog_1);

    while ((ok && recv < n_tot) || (!ok && recv < sent)) {
      
      if (sent < n_tot && sent - recv < window_sz && ok) {
	RPC::ex_prog_1::ex_random (cli, &res[sent], 
				   mkevent(rv,sent,errs[sent]));
	sent++;
      } else {
	// wait until a call has returned; when it has, get
	// the index variable associated with the call, and assign
	// it to 'return_id' to which finished.
	twait (rv, return_id);
	if (errs[return_id]) {
	  ok = false;
	  warn << "RPC error: " << errs[return_id] << "\n";
	} else {
	  warn << "Success " << return_id << ": " << res[return_id] << "\n";
	}
	recv++;
      }
    }
  }
  warn << "All done...\n";
  TRIGGER (cv, ok);
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

  remote_host_t h (argv[1]);
  h.dostuff (port, wrap (finish));

  // TAME_OFF
  foo_func (TAME_FOO);
  // TAME_ON
  
  amain ();
}
