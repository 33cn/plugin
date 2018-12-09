// -*-c++-*-
/* $Id: list.h 1117 2005-11-01 16:20:39Z max $ */

#ifndef _QTAILQ_H_INCLUDED_
#define _QTAILQ_H_INCLUDED_ 1

#include "list.h"

/*
 * qtailq
 *
 *   A "quick" form of tailq, in which the object is stored non-intrusively
 *   in the tailq, via a simple wrapper slot object.
 *
 */

template<class T> 
class qtailq_slot_t {
public:
  qtailq_slot_t (const T &t) : _obj (t) {}
  T obj () const { return _obj; }
  const T _obj;
  tailq_entry<qtailq_slot_t<T> > _lnk;
};

template<class T> 
class qtailq_t {
public:
  qtailq_t () {}

  typedef qtailq_slot_t<T> slot_t;

  ~qtailq_t ()
  {
    slot_t *slot;
    while ((slot = _q.first)) {
      rmslot (slot);
    }
  }

  bool pop_front (T *out) 
  {
    slot_t *slot;
    if ((slot = _q.first)) {
      *out = slot->obj ();
      rmslot (slot);
      return true;
    }
    return false;
  }

  void push_front (const T &obj)
  {
    slot_t *s = New slot_t (obj);
    _q.insert_head (s);
  }

  void push_back (const T &obj)
  {
    slot_t *s = New slot_t (obj);
    _q.insert_tail (s);
  }
  
  void rmslot (slot_t *slot)
  {
    _q.remove (slot);
    delete (slot);
  }

private:
  tailq<slot_t, &slot_t::_lnk> _q;
};

#endif /* _QTAILQ_INCLUDED_H_ */
