
// -*-c++-*-
/* $Id: tame_core.h 2225 2006-09-28 15:41:28Z max $ */

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

#ifndef _LIBTAME_TAME_CFORK_H_
#define _LIBTAME_TAME_CFORK_H_

#include "tame_event.h"
#include "tame_closure.h"
#include "tame_rendezvous.h"
#include "tame_thread.h"
#include "tame_typedefs.h"
#include "async.h"
#include "tame_event_ag.h"

template<class R>
class cthread_t {
public:
  cthread_t (evv_t e, R &r, typename callback<R, void>::ref a) 
    : _event (e), _result (r), _action (a) {}

  static void * run (void *me) 
  { (reinterpret_cast<cthread_t<R> *> (me))->_run (); return NULL; }

  void _run ()
  {
    _result = (*_action) ();
    _event->trigger ();
    delete this;
    tame_thread_exit ();
  }
  
private:
  evv_t _event;
  R &_result;
  typename callback<R, void>::ref _action;
};

class thread_implicit_rendezvous_t : public rendezvous_t<>
{
public:
  thread_implicit_rendezvous_t (const ptr<closure_t> &cl, const char *l) 
    : rendezvous_t<> (l),
      _cls (cl) {}
  const ptr<closure_t> &closure () { return _cls; }

  ~thread_implicit_rendezvous_t() { waitall (); }
private:
  ptr<closure_t> _cls;
};

template<class R>
void __tfork (const char *loc, evv_t e, R &r, 
	      typename callback<R,void>::ref a)
{
  cthread_t<R> *t = New cthread_t<R> (e, r, a);
  tame_thread_spawn (loc, cthread_t<R>::run, static_cast<void *> (t)); 
}

template<>
class cthread_t<void> {
public:
  cthread_t (evv_t e, cbv a) : _event (e), _action (a) {}
  
  static void * run (void *me)
  { (reinterpret_cast<cthread_t<void> *> (me))->_run (); return NULL; }

  void _run ()
  {
    (*_action) ();
    _event->trigger ();
    delete this;
    tame_thread_exit ();
  }
private:
  evv_t _event;
  cbv _action;
};

void __tfork (const char *loc, evv_t e, cbv a);

/*
 * This _tfork is for a situation such as:
 *    
 *   twait { 
 *     tfork (wrap (foo)); 
 *     tfork (wrap (bar));
 *   }
 *
 * in a **TAMED** function.
 */
template<class C>
void _tfork (const closure_wrapper<C> &c, const char *loc, cbv a)
{
  __tfork (loc, _mkevent (c, loc), a);
}

/*
 * This _tfork is for a situation such as:
 *
 *   twait {  
 *     tfork (wrap (foo)); 
 *     bar (mkevent ());
 *   }
 * 
 * in a non-tamed (i.e., THREADED) function.  In this case, we'll
 * allocate an actual rendezvous in the twait{} block, and wait on
 * all threads to finish before leaving the block (control flow
 * stops in the thread_implicit_rendezvous_t destructor).
 */
void _tfork (thread_implicit_rendezvous_t *rv, const char *loc, cbv a);

/*
 * ...and a version of the above for a function that returns an 
 * argument
 */
template<class R>
void _tfork (thread_implicit_rendezvous_t *rv, const char *loc,
	     R &r, typename callback<R,void>::ref a)
{
  __tfork (loc, 
	   _mkevent (rv->closure (), loc,
		     *static_cast<rendezvous_t<> *> (rv)),
	   r, a);
}

/*
 * This _fork is for any situation with an explicit rendezvous, be it
 * in a threaded or TAMED function.
 */
void _tfork (ptr<closure_t> c, const char *loc, rendezvous_t<> &rv, cbv a);

template<class C, class R>
void _tfork (const closure_wrapper<C> &c, const char *loc, R &r, 
	     typename callback<R,void>::ref a)
{
  __tfork (loc, _mkevent (c, loc), r, a);
}

#include "tame_tfork_ag.h"

#define tfork(...) _tfork (__cls_g, __FL__, ##__VA_ARGS__)


event<>::ref
_mkevent (thread_implicit_rendezvous_t *r, const char *loc);

template<class T1>
typename event<T1>::ref
_mkevent (thread_implicit_rendezvous_t *r, const char *loc, T1 &t1)
{
  return _mkevent (r->closure (), loc, *r, t1);
}

template<class T1, class T2>
typename event<T1,T2>::ref
_mkevent (thread_implicit_rendezvous_t *r, const char *loc, T1 &t1, T2 &t2)
{
  return _mkevent (r->closure (), loc, *r, t1, t2);
}

template<class T1, class T2, class T3>
typename event<T1,T2,T3>::ref
_mkevent (thread_implicit_rendezvous_t *r, const char *loc, 
	  T1 &t1, T2 &t2, T3 &t3)
{
  return _mkevent (r->closure (), loc, *r, t1, t2, t3);
}


#endif /* _LIBTAME_TAME_CFORK_H_ */
