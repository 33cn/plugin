
// -*-c++-*-
/* $Id: ex4m.C 1668 2006-04-15 16:08:00Z max $ */

#include "tame.h"
#include "arpc.h"
#include "parseopt.h"
#include "ex_prot.h"

template<class P1, class W1> static void 
__nonblock_cb_1_1 (ptr<closure_t> hold, ptr<joiner_t<W1> > j,
		 refset_t<P1> rs, value_set_t<W1> w, P1 v)
{
  rs.assign (v);

  // always return to the main loop to avoid funny race conditions.
  j->join (w);
}




static  void 
dostuff( str h,  int port,  cbb cb, ptr<closure_t> __cls_g = NULL);

class dostuff__closure_t : public closure_t {
public:
  dostuff__closure_t ( str h,  int port,  cbb cb) : _stack (h, port, cb), _args (h, port, cb), _block1 (0) {}
  ~dostuff__closure_t () { warn << "deleting closure!!\n"; }

  void reenter () {
    dostuff (_args.h, _args.port, _args.cb, mkref (this));
  }

  void block_cb_switch (int i) {}
  
  struct stack_t {
    stack_t ( str h,  int port,  cbb cb) : n_tot (40), window_sz (5), n_out (0), i (0), err (false)  {}
    int fd;
    ptr< axprt_stream > x;
    ptr< aclnt > cli;
    vec< int > res;
    vec< clnt_stat > errs;
    int n_tot;
    int window_sz;
    int n_out;
    int i;
    int cid;
    bool err;
    join_group_t<int> RPC;
  };

  struct args_t {
    args_t ( str h,  int port,  cbb cb) : h (h), port (port), cb (cb) {}
     str h;
     int port;
     cbb cb;
  };

  template<class T>
  void cb1 (refset_t<T> rs, T v1)
  {
    rs.assign (v1);
    if (!--_block1)
      delaycb (0, 0, wrap (mkref (this), &dostuff__closure_t::reenter));
  }

  int _bottom_fencepost;
  stack_t _stack;
  int _top_fencepost;

  args_t _args;

  int _block1;

  bool is_onstack (const void *p) const
  {
    return (static_cast<const void *> (&_bottom_fencepost) < p &&
            static_cast<const void *> (&_top_fencepost) > p);
  }

};

void 
dostuff( str __tame_h,  int __tame_port,  cbb __tame_cb, ptr<closure_t> __cls_g)
{
  ptr<dostuff__closure_t> __cls_r;
  dostuff__closure_t *__cls; // speed up to not use smart pointer ?
  if (!__cls_g) {
    __cls_r = New refcounted<dostuff__closure_t > (__tame_h, __tame_port, __tame_cb);
    __cls = __cls_r;
  } else {
    __cls = reinterpret_cast<dostuff__closure_t *> 
      (static_cast<closure_t *> (__cls_g));
    __cls_r = mkref (__cls);
  }
  
  int &fd = __cls->_stack.fd;
  ptr< axprt_stream > &x = __cls->_stack.x;
  ptr< aclnt > &cli = __cls->_stack.cli;
  vec< int > &res = __cls->_stack.res;
  vec< clnt_stat > &errs = __cls->_stack.errs;
  int &n_tot = __cls->_stack.n_tot;
  int &window_sz = __cls->_stack.window_sz;
  int &n_out = __cls->_stack.n_out;
  int &i = __cls->_stack.i;
  int &cid = __cls->_stack.cid;
  bool &err = __cls->_stack.err;
  join_group_t<int> &RPC = __cls->_stack.RPC;

  str &h = __cls->_args.h;
  int &port = __cls->_args.port;
  cbb &cb = __cls->_args.cb;
  
  switch (__cls->jumpto ()) {
  case 1:
    goto dostuff__label1;
    break;
  case 2:
    goto dostuff__label2;
    break;
  default:
    break;
  }
  
  {
    __cls->set_jumpto (1);

    // in the case that all calls finish immediately, we still want to 
    // hold one reference, so that the reference count doesn't go down
    // to 0 prematurely.
    __cls->_block1 = 1;
    
    tcpconnect (h, port, 
		(++__cls->_block1,
		 wrap (__cls_r, 
		       &dostuff__closure_t::cb1<typeof (fd)>, 
		       refset_t<typeof (fd)> (fd))));

    // if the reference count is 0 here, that means that all calls returned
    // immediately, and there is no need to block; therefore, only block
    // (by returning) if there is at least one call outstanding.
    if (-- __cls->_block1 )
      return;

  }
 dostuff__label1:
  
  
  if (fd < 0) {
    warn ("%s:%d: connection failed: %m\n", h.cstr(), port);
    err = true;
  } else {

    res.setsize (n_tot);
    errs.setsize (n_tot);
    x = axprt_stream::alloc (fd);
    cli = aclnt::alloc (x, ex_prog_1);
    
    while (n_out < window_sz && i < n_tot) {
      n_out ++ ;
      cid = i++;
      cli->call (EX_RANDOM, NULL, &res[cid],
		 (RPC.launch_one (),
		  wrap (__nonblock_cb_1_1<typeof(errs[cid]), typeof (cid)>, 
			__cls_g,
			RPC.make_joiner ("<function location XXX>"), 
			refset_t<typeof(errs[cid])> (errs[cid]),
			value_set_t<typeof(cid)> (cid)
			)));
      
    }
    while (RPC.need_join ()) {
      
      // JOIN (&RPC, cid) { ...
    dostuff__label2:
      typeof (RPC.to_vs ()) v;

      if (RPC.pending (&v)) {
	typeof (v.v1) &cid = v.v1;
	
	--n_out;
	if (errs[cid]) {
	  warn << "RPC error: " << errs[cid] << "\n";
	} else {
	  warn << "Success " << cid << ": " << res[cid] << "\n";
	  if (i != n_tot) {
	    n_out ++;
	    cid = i++;
	    cli->call (EX_RANDOM, NULL, &res[cid],
		       (RPC.launch_one (),
			wrap (__nonblock_cb_1_1< typeof(errs[cid]), 
			      typeof (cid)>,
			      __cls_g,
			      RPC.make_joiner ("<function location YYY>"), 
			      refset_t<typeof(errs[cid])> (errs[cid]),
			      value_set_t<typeof(cid)> (cid)
			      )));
	  }
	}
      } else {
	__cls->_args.h = h;
	__cls->_args.port = port;
	__cls->_args.cb = cb;
	__cls->set_jumpto (2);
	
	RPC.set_join_cb (wrap (__cls_r, &dostuff__closure_t::reenter));
	return;
      }
    }
    warn << "All done...\n";
  }
  (*cb) (!err);
}


static void finish (bool rc)
{
  // delay exit so that we see the closure destroyed!  w00t!
  delaycb (0, 0, wrap (exit, rc ? 0 : -1 ));
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
