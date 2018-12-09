// -*-c++-*-
/* $Id: tame_core.h 2654 2007-03-31 05:42:21Z max $ */

/*
 *
 * Copyright (C) 2005 Max Krohn (max@okws.org)
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU General Public License as
 * published by the Free Software Foundation; either version 2, or (at
 * your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 59 Temple Place, Suite 330, Boston, MA 02111-1307
 * USA
 *
 */

#ifndef _LIBTAME_RENDEZVOUS_H_
#define _LIBTAME_RENDEZVOUS_H_

#include "tame_event.h"
#include "tame_closure.h"

#ifdef HAVE_TAME_PTH
# include <pth.h>

#endif /* HAVE_TAME_PTH */

extern nil_t g_nil;

template<class T1=nil_t, class T2=nil_t, class T3=nil_t, class T4=nil_t>
struct value_set_t {

  typedef value_set_t<T1,T2,T3,T4> my_type_t;

  value_set_t () {}
  value_set_t (const T1 &v1) : v1 (v1) {}
  value_set_t (const T1 &v1, const T2 &v2) : v1 (v1), v2 (v2) {}
  value_set_t (const T1 &v1, const T2 &v2, const T3 &v3) 
    : v1 (v1), v2 (v2), v3 (v3) {}
  value_set_t (const T1 &v1, const T2 &v2, const T3 &v3, const T4 &v4) 
    : v1 (v1), v2 (v2), v3 (v3), v4 (v4) {}

  T1 v1;
  T2 v2;
  T3 v3;
  T4 v4;
};

typedef enum { JOIN_NONE = 0,
	       JOIN_EVENTS = 1,
	       JOIN_THREADS = 2 } join_method_t;

/*
 * An "action" is a type of activity to take internally to a "trigger".
 * In the case of rendezvous, this means joining with the rendezvous 
 * and potentially reentering the function (if twait(); was called).
 * The "action" concept is a workaround virtual method calls, so by
 * convention, "actions" such as the one below must implement the
 * clear() and perform() methods.
 *
 * Template parameters: 
 *   R={type of rendezvous}
 *   V={type of wait value set}
 */
template<class R, class V>
class rendezvous_action {
public:
  rendezvous_action (R *rv,
		     ptr<closure_t> c,
		     const V &vs)
    : _rv (mkweakref (rv)),
      _cls (c),
      _value_set (vs),
      _cleared (false)
  {
  }

  //
  // Perform the "action" required by this trigger. In this case, attempt
  // to rejoin the rendezvous. If the event is not reusable, then clear
  // the event in the process, and return 'true.'  If the event is
  // reusable, then don't clear, and return 'false.'
  //
  bool perform (_event_cancel_base *event, const char *loc, bool _reuse)
  {
    R *rp;
    bool ret = false;
    if (_cleared) {
      tame_error (loc, "event reused after deallocation");
    } else if ((rp = _rv.pointer ())) {
      rp->_ti_join (_value_set, event, !_reuse);
      if (!_reuse) {
	clear ();
	ret = true;
      }
    } else if (!_rv.flag ()->is_cancelled ()) {
      tame_error (loc, "event triggered after rendezvous was deallocated");
    }
    return ret;
  }

  void clear (_event_cancel_base *e)
  {
    if (!_cleared) {
      if (R *rp = _rv.pointer ()) {
	rp->remove (e);
      }
      clear ();
    }
  }

private:
  void clear () 
  {
    _cls = NULL;
    _cleared = true;
  }

  weakref<R> _rv;
  ptr<closure_t> _cls;
  V _value_set;
  bool _cleared;
};

class rendezvous_base_t : public weakrefcount {
public:
  rendezvous_base_t (const char *loc)
    : _loc (loc ? loc : "(unknown)") 
  { collect_self (); }

  rendezvous_base_t (const char *loc, int line)
    : _loc_s (strbuf ("%s:%d", loc, line)),
      _loc (_loc_s.cstr ()) 
  { collect_self (); }

  virtual ~rendezvous_base_t () {}

  inline const char *loc () const { return _loc; }
  virtual u_int n_triggers_left () const = 0;

  inline void collect_self ()
  {
    if (tame_check_leaks ()) {
      collect_rendezvous (mkweakref (this));
    }
  }

private:
  str _loc_s;
  const char *_loc;
};

template<class W1=nil_t, class W2=nil_t, class W3=nil_t, class W4=nil_t>
class rendezvous_t : public rendezvous_base_t {

  typedef rendezvous_t<W1,W2,W3,W4> my_type_t;
  typedef value_set_t<W1,W2,W3,W3> my_value_set_t;
  friend class rendezvous_action<my_type_t, my_value_set_t>;
  typedef rendezvous_action<my_type_t, my_value_set_t> my_action_t;

public:
  rendezvous_t (const char *loc, int line)
    : rendezvous_base_t (loc, line),
      _join_method (JOIN_NONE),
      _n_events (0),
      _is_cancelling (false)
  {
    pth_init ();
  }

  rendezvous_t (const char *loc = NULL)
    : rendezvous_base_t (loc),
      _join_method (JOIN_NONE),
      _n_events (0),
      _is_cancelling (false)
  {
    pth_init ();
  }
  
  ~rendezvous_t () { cleanup(); }

  //=======================================================================
  // Public Interface to Rendezvous Class

  void cancel ()
  {
    // setting this flag prevents more events from being added on
    // to this rendezvous.
    _is_cancelling = true;

    cancel_all_events ();

    // Must set flag after cancelling all events, since events cannot
    // access this object if it's cancelled (due to particulars of
    // weak refcount implementation -- nothing "deep").
    this->flag()->set_cancelled ();
  }

  u_int n_pending () const { return _pending_values.size (); }
  u_int n_events_out () const { return _n_events; }
  u_int n_triggers_left () const { return n_events_out () + n_pending (); }
  bool need_wait () const { return n_triggers_left () > 0; }

  //-----------------------------------------------------------------------
  // Threaded interface

  void wait (W1 &r1 = g_nil, W2 &r2 = g_nil, W3 &r3 = g_nil, W4 &r4 = g_nil)
  { 
    bool rls = thread_lock_acquire ();
    while (!_ti_next_trigger (r1, r2, r3, r4)) 
      threadwait (); 
    thread_lock_release (rls);
  }

  void waitall () 
  { 
    bool rls = thread_lock_acquire ();
    while (n_events_out () > 0) {
      while (!_ti_next_trigger ()) {
	threadwait ();
      }
    }
    thread_lock_release (rls);
  }
  // End threaded interface
  //-----------------------------------------------------------------------

  // End Public Interface
  //=======================================================================

  //=======================================================================
  // Public Interface, but internal to tame (hence _ti_*)

  void _ti_set_join_method (join_method_t jm)
  {
    assert (_join_method == JOIN_NONE);
    _join_method = jm;
  }

  void _ti_clear_join_method () 
  { 
    if (_join_method == JOIN_EVENTS) {
      _join_cls = NULL;
    }
    _join_method = JOIN_NONE; 
  }
  
  void _ti_set_join_cls (ptr<closure_t> c)
  { 
    _ti_set_join_method (JOIN_EVENTS); 
    _join_cls = c;
  }

  template<class T1, class T2, class T3>
  typename event<T1,T2,T3>::ptr
  _ti_mkevent (ptr<closure_t> cls, 
	       const char *eloc,
	       const my_value_set_t &vs,
	       const _tame_slot_set<T1,T2,T3> &rs)
  {
    ptr<_event_impl<my_action_t,T1,T2,T3> > ret;
    if (!this->flag ()->is_alive () || _is_cancelling) {
      strbuf b;
      b.fmt ("Attempted to add an event to a rendezvous (allocated %s) "
	     "this is no longer active", loc ());
      str s = b;
      tame_error (eloc, s.cstr ());
    } else {
      ret = New refcounted<_event_impl<my_action_t,T1,T2,T3> > 
	(my_action_t (this, cls, vs), rs, eloc);
      _n_events ++;
      _events.insert_head (ret);
    }
    return ret;
  }
  
  void _ti_join (const my_value_set_t &v, _event_cancel_base *e, bool clear)
  {

#ifdef HAVE_TAME_PTH
    bool rls = thread_lock_acquire ();
#endif 

    _pending_values.push_back (v);

    if (clear) 
      remove (e);

    if (_join_method == JOIN_EVENTS) {
      assert (_join_cls);
      ptr<closure_t> c = _join_cls;
      _join_cls = NULL;
      _join_method = JOIN_NONE;
      c->v_reenter ();
    } else if (_join_method == JOIN_THREADS) {
#ifdef HAVE_TAME_PTH
      pth_cond_notify (&_cond, 0);
#else
      panic ("no PTH available\n");
#endif
    } else {
      /* called join before a waiter; we can just queue */
    }

#ifdef HAVE_TAME_PTH
    thread_lock_release (rls);
#endif

  }

  bool _ti_next_trigger (W1 &r1 = g_nil, W2 &r2 = g_nil, 
			 W3 &r3 = g_nil, W4 &r4 = g_nil)
  {
    bool ret = true;

#ifdef HAVE_TAME_PTH
    bool rls = thread_lock_acquire ();
#endif

    value_set_t<W1,W2,W3,W4> *v;
    if (pending (&v)) {
      r1 = v->v1;
      r2 = v->v2;
      r3 = v->v3;
      r4 = v->v4;
      consume ();
    } else
      ret = false;

#ifdef HAVE_TAME_PTH
    thread_lock_release (rls);
#endif

    return ret;
  }

  // End tame-internal public interface
  //=======================================================================

private:

  inline bool thread_lock_acquire ()
  {
    bool rls = false;
#ifdef HAVE_TAME_PTH
    if (!_has_lock) {
      pth_mutex_acquire (&_mutex, 0, NULL);
      _has_lock = true;
      rls = true;
    }
#else /* ! HAVE_TAME_PTH */
    /* noop */
#endif /* HAVE_TAME_PTH */
    return rls;
  }

  inline void thread_lock_release (bool rls)
  {
#ifdef HAVE_TAME_PTH
    if (rls) {
      assert (_has_lock);
      _has_lock = false;
      pth_mutex_release (&_mutex);
    }
#else /* ! HAVE_TAME_PTH */
    /* noop */
#endif /* HAVE_TAME_PTH */
  }

  void threadwait ()
  {
#ifdef HAVE_TAME_PTH
    assert (_has_lock);
    _ti_set_join_method (JOIN_THREADS);
    pth_cond_await (&_cond, &_mutex, NULL);
    _ti_clear_join_method ();
#else /* ! HAVE_TAME_PTH */
    panic ("no PTH available...\n");
#endif
  }

  inline void pth_init ()
  {
#ifdef HAVE_TAME_PTH
    pth_mutex_init (&_mutex);
    pth_cond_init (&_cond);
    _has_lock = false;
#endif /* HAVE_TAME_PTH */
  }

  void cleanup ()
  {
    if (need_wait () && !this->flag ()->is_cancelled ()) {
      strbuf b;
      b.fmt ("rendezvous went out of scope when expecting %u trigger%s",
	     n_triggers_left (),
	     n_triggers_left () > 1 ? "s" : "");
      str s = b;
      tame_error (loc(), s.cstr ());
    } 
    this->flag()->set_dead ();
    report_leaks (&_events);
    cancel_all_events ();
  }

private:

  bool pending (my_value_set_t **p = NULL)
  {
    bool ret = false;
    if (_pending_values.size ()) {
      if (p) *p = &_pending_values[0];
      ret = true;
    }
    return ret;
  }

  void consume ()
  {
    assert (_pending_values.size ());
    _pending_values.pop_front ();
  }

  void remove (_event_cancel_base *e) 
  { 
    _n_events --;
    _events.remove (e); 
  }
  
  void cancel_all_events ()
  {
    _event_cancel_base *b;
    while ((b = _events.first)) {
      remove (b);
      b->cancel ();
    }
  }

private:
  // Disallow copying
  rendezvous_t (const my_type_t &rv) {}

  event_cancel_list_t _events;
  vec<my_value_set_t> _pending_values;
  ptr<closure_t> _join_cls;
  join_method_t _join_method;
  u_int _n_events;

#ifdef HAVE_TAME_PTH
  pth_cond_t _cond;
  pth_mutex_t _mutex;
  bool _has_lock;
#endif /* HAVE_TAME_PTH */

  bool _is_cancelling;

};


#endif /* _LIBTAME_RENDEZVOUS_H_ */

