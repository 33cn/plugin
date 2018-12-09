

// -*-c++-*-
/* $Id: tame_event.h 2487 2007-01-09 14:54:32Z max $ */

#ifndef _LIBTAME_RECYCLE_H_
#define _LIBTAME_RECYCLE_H_

#include "async.h"
#include "refcnt.h"
#include "vec.h"
#include "init.h"
#include "list.h"

template<class C>
class container_t {
public:
  container_t (const C &c) { (void) new (getp_void ()) C (c); }
  inline void *getp_void () { return reinterpret_cast<void *> (_bspace); }
  inline C *getp () { return reinterpret_cast<C *> (getp_void ()); }
  void del () { getp ()->~C (); }
  C &get () { return *getp (); }
  const C &get () const { return *getp (); }
private:
  char _bspace[sizeof(C)];
};

/**
 * A class for collecting recycled objects, and recycling them.  Different
 * from a regular vec in that we never give memory back (via pop_back)
 * which should save some shuffling.
 *
 * Required: that T inherits from refcnt
 */
template<class T>
class recycle_bin_t {
public:
  enum { defsz = 8192 };
  recycle_bin_t (size_t s = defsz) : _capacity (s), _n (0)  {}
  void expand (size_t s) { if (_capacity < s) _capacity = s; }
  void recycle (T *obj)
  {
    if (_n < _capacity) {
      _objects.insert_head (obj);
      _n++;
    } else {
      delete this;
    }
  }

  ptr<T> get () 
  {
    ptr<T> ret;
    if (_objects.first) {
      T *o = _objects.first;
      _objects.remove (o);
      _n--;
      ret = mkref (o);
    }
    return ret;
  }

private:
  list<T, &T::_lnk> _objects;
  size_t       _capacity;
  size_t       _n;
};

INIT(recycle_init);

typedef enum { OBJ_ALIVE = 0, 
	       OBJ_SICK = 0x1, 
	       OBJ_DEAD = 0x2,
	       OBJ_CANCELLED = 0x4 } obj_state_t;

class obj_flag_t : virtual public refcount {
public:
  obj_flag_t (const obj_state_t &b) : _flag (b) {}
  ~obj_flag_t () { }

  static void recycle (obj_flag_t *p);
  static ptr<obj_flag_t> alloc (const obj_state_t &b);
  static recycle_bin_t<obj_flag_t> *get_recycle_bin ();

  void finalize () { recycle (this); }
  inline bool get (obj_state_t b) { return (_flag & b) == b; }

  inline void set (const obj_state_t &b) { _flag = b; }

  inline void set_flag (const obj_state_t &b)
  { _flag = obj_state_t (int (_flag) | b); }

  inline bool is_alive () const { return _flag == OBJ_ALIVE; }
  inline bool is_cancelled () const { return (_flag & OBJ_CANCELLED); }
  inline void set_dead () { set_flag (OBJ_DEAD); }
  inline void set_cancelled () { set_flag (OBJ_CANCELLED); }

  list_entry<obj_flag_t> _lnk;

private:
  obj_state_t _flag;
};

typedef ptr<obj_flag_t> obj_flag_ptr_t;

#endif /* _LIBTAME_RECYCLE_H_ */
