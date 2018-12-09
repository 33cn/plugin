// -*-c++-*-
/* $Id: tame_event.h 2666 2007-04-02 03:15:22Z max $ */

#ifndef _LIBTAME_TAME_WEAKREF_H_
#define _LIBTAME_TAME_WEAKREF_H_

#include "refcnt.h"
#include "tame_recycle.h"

class weakrefcount {
public:
  weakrefcount () : _flag (obj_flag_t::alloc (OBJ_ALIVE)) {}
  obj_flag_ptr_t flag () { return _flag; }
  ~weakrefcount () { _flag->set_dead (); }
private:
  obj_flag_ptr_t _flag;
};

template<class T>
class weakref {
public:
  weakref (T *p, obj_flag_ptr_t f) : _pointer (p), _flag (f) {}
  inline T *pointer () { return _flag->is_alive () ? _pointer : NULL; }
  inline const T *pointer() const 
  { return _flag->is_alive () ? _pointer : NULL; }

  obj_flag_ptr_t flag () { return _flag; }
private:
  T *_pointer;
  obj_flag_ptr_t _flag;
};

template<class T> weakref<T>
mkweakref (T *p)
{
  return weakref<T> (p, p->flag ());
}

#endif // _LIBTAME_TAME_WEAKREF_
