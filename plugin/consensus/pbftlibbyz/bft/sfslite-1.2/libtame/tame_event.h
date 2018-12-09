
// -*-c++-*-
/* $Id: tame_event.h 3758 2008-11-13 00:36:00Z max $ */

#ifndef _LIBTAME_TAME_EVENT_H_
#define _LIBTAME_TAME_EVENT_H_

#include "refcnt.h"
#include "vec.h"
#include "init.h"
#include "async.h"
#include "list.h"
#include "tame_slotset.h"
#include "tame_run.h"

// Specify 1 extra argument, that way we can do template specialization
// elsewhere.  We should never have an instatiated event class with
// 4 templated types, though.
template<class T1=void, class T2=void, class T3=void, class T4=void> 
class _event;

class _event_cancel_base : public virtual refcount {
public:
  _event_cancel_base (const char *loc) : 
    _loc (loc), 
    _cancelled (false),
    _cleared (false),
    _reuse (false),
    _performing (false)
  { g_stats->did_mkevent (); }

  ~_event_cancel_base () {}

  void reinit (const char *loc)
  {
    _loc = loc;
    _cancelled = false;
    _performing = false;
    _cleared = false;
    _reuse = false;
  }

  void set_cancel_notifier (ptr<_event<> > e) { _cancel_notifier = e; }
  void cancel ();
  const char *loc () const { return _loc; }
  bool cancelled () const { return _cancelled; }

  operator bool() const { return is_alive (); }

  bool is_alive () const 
  { return (!_cancelled && (_reuse || !_performing) && !_cleared); }

  void set_reuse (bool b) { _reuse = b; }
  bool get_reuse () const { return _reuse; }

  bool can_trigger ()
  {
    bool ret = false;
    if (_cancelled) {
      if (tame_strict_mode ()) 
	tame_error (_loc, "event triggered after it was cancelled");
    } else if (_performing && !_reuse) {
      tame_error (_loc, "event triggered recursively");
    } else if (_cleared) {
      tame_error (_loc, "event triggered after it was cleared");
    } else {
      ret = true;
    }
    return ret;
  }

  void trigger_no_assign ()
  {
    if (can_trigger ()) {
      ptr<_event_cancel_base> hold (mkref (this));
      _performing = true;
      if (perform_action (this, _loc, _reuse)) {
	_cleared = true;
      }
      _performing = false;
    }
  }

  void finish () { clear (); }

  list_entry<_event_cancel_base> _lnk;

protected:
  virtual bool perform_action (_event_cancel_base *e, const char *loc,
			       bool reuse) = 0;
  virtual void clear_action () = 0;
  void clear ();

  const char *_loc;
  bool _cancelled;
  bool _cleared;
  bool _reuse;
  bool _performing;
  ptr<_event<> > _cancel_notifier;

};

typedef ptr<_event_cancel_base> _event_hold_t;

typedef list<_event_cancel_base, &_event_cancel_base::_lnk> 
event_cancel_list_t;

void report_leaks (event_cancel_list_t *lst);

template<class A, class T1=void, class T2=void, class T3=void, class T4=void> 
class _event_impl;

template<class T1=void, class T2=void, class T3=void>
class event {
public:
  typedef struct ref<_event<T1,T2,T3> > ref;
  typedef struct ptr<_event<T1,T2,T3> > ptr;
};


#ifdef WRAP_DEBUG
# define CALLBACK_ARGS(x) "???", x, x
# else
# define CALLBACK_ARGS(x)
#endif

#define mkevent(...) _mkevent (__cls_g, __FL__, ##__VA_ARGS__)
#define mkevent_rs(...) _mkevent_rs (__cls_g, __FL__, ##__VA_ARGS__)

#endif /* _LIBTAME_TAME_EVENT_H_ */
