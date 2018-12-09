
// -*-c++-*-
/* $Id: tame.h 2077 2006-07-07 18:24:23Z max $ */

#ifndef _LIBTAME_LOCK_H_
#define _LIBTAME_LOCK_H_

#include "async.h"
#include "list.h"
#include "tame.h"

// 
// A simple lock, to get coarse-grained atomicity across blocking
// operations.
//
namespace tame {

  class lock_t {
  public:
    
    enum mode_t {
      OPEN = 0,
      SHARED = 1,
      EXCLUSIVE = 2
    };
    
    lock_t (mode_t m = OPEN);
    
    struct waiter_t {
      waiter_t (mode_t m, evv_t c) : _mode (m), _cb (c) {}
      mode_t _mode;
      evv_t _cb;
      tailq_entry<waiter_t> _lnk;
    };
    
    waiter_t *acquire (mode_t m, evv_t cb);
    void timed_acquire (mode_t m, u_int s, u_int ms, evb_t cb, CLOSURE);
    void release ();
    void cancel (waiter_t *w);
    mode_t mode () const { return _mode; }
    
  protected:
    void call (waiter_t *w, bool from_main = false);
    
    tailq<waiter_t, &waiter_t::_lnk> _waiters;
    mode_t _mode;
    int _sharers;
  };

};


#endif /* _LIBTAME_LOCK_H_ */

