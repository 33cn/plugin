
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

#ifndef _LIBTAME_CLOSURE_H_
#define _LIBTAME_CLOSURE_H_

#include "tame_event.h"
#include "tame_run.h"
#include "tame_weakref.h"


// All closures are numbered serially so that our accounting does not
// get confused.
extern u_int64_t closure_serial_number;

class rendezvous_base_t;

class closure_t : public virtual refcount {
public:
  closure_t (const char *filename, const char *fun) ;

  virtual ~closure_t () {}

  // manage function reentry
  inline void set_jumpto (int i) { _jumpto = i; }
  inline u_int jumpto () const { return _jumpto; }

  inline u_int64_t id () { return _id; }

  // given a line number of the end of scope, perform sanity
  // checks on scoping, etc.
  void end_of_scope_checks (int line);

  // Initialize a block environment with the ID of this block
  // within the given function.  Also reset any internal counters.
  void init_block (int blockid, int lineno);

  u_int _jumpto;

  // Display full location with filename, line number and also
  // function name.
  str loc (int lineno) const;
  void error (int lineno, const char *msg);

  // Decremenet the block count; return TRUE if it goes down to 0, signifying
  // contuination inside the function.
  bool block_dec_count (const char *loc);
  inline void block_inc_count () { _block._count++; }

  //
  // Rendezvous can't statically type what kind of closure they
  // need to jump back into; therefore, we need a virtual reentry
  // function, as given here.  It won't be called for implicit rvs.
  //
  virtual void v_reenter () = 0;

  //
  // only called if tame_check_leaks is on
  //
  virtual bool is_onstack (const void *p) const = 0;

  void collect_rendezvous ();

protected:

  u_int64_t _id;
  
  const char *_filename;              // filename for the function
  const char *_funcname;         

public:
  // Variables involved with managing BLOCK blocks. Note that only one
  // can be active at any given time.
  struct block_t { 
    block_t () : _id (0), _count (0), _lineno (0) {}
    int _id, _count, _lineno;
  };
  block_t _block;
  
  vec<weakref<rendezvous_base_t> > _rvs;
};

typedef ptr<closure_t> closure_ptr_t;

template<class C>
class closure_action {
public:
  closure_action (ptr<C> c) : _closure (c) {}

  //
  // One would like to assert (!_closure) in the destructor,
  // but we cannot, unfortunately, since there are short-lived
  // copies of a closure_action that don't ever use their
  // _closure field aside from just passing it onto other 
  // closure_actions
  //
  ~closure_action () {}

  bool perform (_event_cancel_base *event, const char *loc, bool _reuse)
  {
    bool ret = false;
    if (!_closure) {
      tame_error (loc, "event reused after deallocation");
    } else {
      // Maybe reenter does clear() so no need to call it separately
      maybe_reenter (loc);
      ret = true;
    }
    return ret;
  }

  void clear (_event_cancel_base *e) 
  {
    if (_closure) {
      _closure = NULL;
    }
  }

private:

  void maybe_reenter (const char *loc)
  {
    ptr<C> c = _closure;
    _closure = NULL;
    if (c->block_dec_count (loc)) {
      if (tame_always_virtual ()) {
	c->v_reenter ();
      } else {
	c->reenter ();
      }
    }
    c = NULL;
  }
  
  ptr<C> _closure;
};

/*
 * Need to disambiguate between explicit rendezvous and implicit rendezvous
 * cases.  In the implicit case, we wrap a ptr<C> (C = some closure) in
 * a closure_wrapper, so that we pick up the right templated version
 * of _mkevent.
 */
template<class C>
class closure_wrapper {
public:
  closure_wrapper (const ptr<C> &c) : _closure (c) {}
  const ptr<C> &closure () const { return _closure; }
private:
  const ptr<C> &_closure;
};

template<class C, class T1, class T2, class T3>
typename event<T1,T2,T3>::ptr
_mkevent_implicit_rv (const ptr<C> &c, 
		      const char *loc,
		      const _tame_slot_set<T1,T2,T3> &rs)
{
  ptr<_event_impl<closure_action<C>,T1,T2,T3> > ret =
    New refcounted<_event_impl<closure_action<C>,T1,T2,T3> > 
    (closure_action<C> (c), rs, loc);

  c->block_inc_count ();
  g_stats->mkevent_impl_rv_alloc (loc);
  return ret;
}

template<class T> void use_reference (T &i) {}

void start_rendezvous_collection ();
void collect_rendezvous (weakref<rendezvous_base_t> r);

extern ptr<closure_t> __cls_g;
extern ptr<closure_t> null_closure;
#define CLOSURE              ptr<closure_t> __frame = NULL

#endif /* _LIBTAME_CLOSURE_H_ */
