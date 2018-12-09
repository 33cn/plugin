
// -*-c++-*-
/* $Id: tame.h 2077 2006-07-07 18:24:23Z max $ */

#ifndef _LIBTAME_PIPELINE_H_
#define _LIBTAME_PIPELINE_H_

#include "async.h"
#include "tame.h"

//
// A class to automate pipelined operations.
//
namespace tame {

class pipeliner_t {
public:
  pipeliner_t (size_t w);
  virtual ~pipeliner_t () {}

  void run (evv_t done, CLOSURE);
  void cancel () { _cancelled = true; }

protected:
  virtual void pipeline_op (size_t i, evv_t done, CLOSURE) = 0;
  virtual bool keep_going (size_t i) const = 0;

  size_t _wsz;
  rendezvous_t<> _rv;
  bool _cancelled;

private:
  void wait_n (size_t n, evv_t done, CLOSURE);
  void launch (size_t i, evv_t done, CLOSURE);
};

typedef callback<void, size_t, cbb, ptr<closure_t> >::ref pipeline_op_t;

void do_pipeline (size_t w, size_t n, pipeline_op_t op, evv_t done, CLOSURE);

};


#endif /* _LIBTAME_PIPELINE_H_ */


