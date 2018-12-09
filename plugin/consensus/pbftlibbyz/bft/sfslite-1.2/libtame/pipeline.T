
// -*-c++-*-
/* $Id: tame.h 2077 2006-07-07 18:24:23Z max $ */

#include "tame_pipeline.h"


namespace tame {

pipeliner_t::pipeliner_t (size_t w) : 
 _wsz (w), _rv (__FILE__, __LINE__), _cancelled (false)
{ assert (_wsz > 0); }
  
tamed void
pipeliner_t::wait_n (size_t n, evv_t done)
{
  while (_rv.n_triggers_left () > n) {
    twait (_rv);
  }
  done->trigger ();
}

tamed void
pipeliner_t::launch (size_t i, evv_t done)
{
  twait { wait_n (_wsz - 1, mkevent () ); }
  pipeline_op (i, mkevent (_rv) );
  done->trigger ();
}

tamed void
pipeliner_t::run (evv_t done)
{
  tvars {
    size_t i;
  }
  for (i = 0; keep_going (i) && !_cancelled; i++) {
    twait { launch (i, mkevent () ); }
  }
  twait { wait_n (0, mkevent () ); }
  done->trigger ();
}

class pipeliner_cb_t : public pipeliner_t {
public:
  pipeliner_cb_t (size_t w, size_t n, pipeline_op_t o)
    : pipeliner_t (w), _n_calls (n), _op (o) {}

protected:
  void pipeline_op (size_t i, evv_t cb, CLOSURE);
  bool keep_going (size_t i) const { return i < _n_calls; }

private:
  size_t _n_calls;
  pipeline_op_t _op;

};

tamed void
pipeliner_cb_t::pipeline_op (size_t i, evv_t cb)
{
  tvars {
    bool ok;
  }
  twait { (*_op) (i, mkevent (ok), NULL); }
  if (!ok) cancel ();
  cb->trigger ();
}

tamed void
do_pipeline (size_t w, size_t n, pipeline_op_t op, evv_t done)
{
  tvars {
    pipeliner_cb_t ppl (w, n, op);
  }
  twait { ppl.run (mkevent ()); }
  done->trigger ();
}

};
