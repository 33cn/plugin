
// -*-c++-*-
#include "tame_lock.h"
#include "tame_trigger.h"

namespace tame {

lock_t::waiter_t *
lock_t::acquire (lock_t::mode_t m, evv_t cb)
{
  lock_t::waiter_t *ret = NULL;
  bool wait = false;

  assert (m != OPEN);

  if (m == SHARED) {
    if (_mode == OPEN) {
      _mode = SHARED;
      assert (_sharers == 0);
      _sharers = 1;
    } else if (_mode == SHARED) {
      assert (_sharers > 0);
      _sharers ++;
    } else {
      wait = true;
    }
  } else {
    assert (m == EXCLUSIVE);
    if (_mode == OPEN) {
      _mode = EXCLUSIVE;
      assert (_sharers == 0);
    } else {
      wait = true;
    }
  }
  
  /*
   * DEBUG
   warn ("LL acquire %p mode=%d sharers=%d wait=%d\n", 
   this, m, _sharers, wait);
  */
    
  if (wait) {
    ret = New lock_t::waiter_t (m, cb);
    _waiters.insert_tail (ret);
  }

  if (!ret) cb->trigger ();
  return ret;
}

void
lock_t::call (waiter_t *w, bool delay)
{
  _waiters.remove (w);
  evv_t cb (w->_cb);
  delete w;
  if (delay)
    dtrigger (cb);
  else
    cb->trigger ();
}

void
lock_t::cancel (waiter_t *w)
{
  /*
   * DEBUG
   warn ("LL cancel %p\n", this);
  */
  _waiters.remove (w);
  delete w;
}

lock_t::lock_t (mode_t m)
  : _mode (m), _sharers (m == SHARED ? 1 : 0) {}

void
lock_t::release ()
{
  mode_t old_mode = _mode;

  /*
   * DEBUG
  warn ("LL release %p mode=%d sharers=%d\n", this, _mode, _sharers);
  */

  assert (_mode != OPEN);

  if (_mode == SHARED) {
    assert (_sharers > 0);
    if ( -- _sharers == 0) {
      _mode = OPEN;
    }
  } else {
    assert (_sharers == 0);
    _mode = OPEN;
  }

  if (_mode == OPEN) {

    waiter_t *w = _waiters.first;
    if (w) {
      if (w->_mode == SHARED) {

	// In other words, assert that we've plucked all sharers
	// off the list below.  If we haven't, we're in shaky
	// recursive release waters (see below).
	assert (old_mode != SHARED);

	_mode = SHARED;

	waiter_t *n, *p;
	for (p = _waiters.first; p; p = n) {
	  n = _waiters.next (p);
	  if (p->_mode == SHARED) {
	    _sharers ++;

	    // Walter Mundt 8/5/08 points out that calling p's callback has
	    // to take place from the main loop.  If not, then calling
	    // p can release right away, which will recursively call into
	    // this method, failing the above assertion (old_mode != SHARED)
	    // and also fouling our traversal through the list
	    call (p, true);

	  }
	}
      } else {
	assert (w->_mode == EXCLUSIVE);
	_mode = EXCLUSIVE;
	call (w);
      }
    }
  }

}

tamed void
lock_t::timed_acquire (lock_t::mode_t m, u_int s, u_int ms, evb_t cb)
{
  tvars {
    rendezvous_t<bool> rv  (__FILE__, __LINE__);
    lock_t::waiter_t *w;
    bool ok;
    timecb_t *tcb;
  }
  tcb = delaycb (s, ms, mkevent (rv,false));
  w = acquire (m, mkevent (rv,true) );
  twait (rv, ok);
  if (!ok) cancel (w);
  else timecb_remove (tcb);
  cb->trigger (ok);
}

};
