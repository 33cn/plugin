
// -*-c++-*-

#ifndef _ASYNC_RCLIST_H_INCLUDED_
#define _ASYNC_RCLIST_H_INCLUDED_ 1

#include "list.h"


/*
 * rclist_t
 *
 *  This class allows a list.h-like interface to reference-counted objects,
 *  so that as long as the object is in the list, its refcount is at least
 *  1.
 *
 */

template<class T>
struct rclist_entry_t {
public:
  rclist_entry_t () : _node (NULL) {}
  ~rclist_entry_t () { assert (!_node); }
    
  void put_in_list (void *v)
  {
    /* don't double-insert this object */
    assert (!_node); 
    _node = v;
  }
  void *v_node () { return _node; }
  
  void remove_from_list ()
  {
    assert (_node);
    _node = NULL;
  }
  
private:
  void *_node;
  bool _in_list;
};

template<class T, rclist_entry_t<T> T::*field>
class rclist_node_t {
public:
  rclist_node_t (ptr<T> e) : _elm (e) { (e->*field).put_in_list (this); }
  list_entry<rclist_node_t<T,field> > _lnk;
  ptr<T> elem () { return _elm; }
private:
  ptr<T> _elm;
};

template<class T, rclist_entry_t<T> T::*field>
class rclist_t {
public:

  rclist_t () {}
  
  typedef list<rclist_node_t<T,field>, 
	       &rclist_node_t<T,field>::_lnk > mylist_t;

  typedef rclist_node_t<T, field> mynode_t;
  
  ptr<T> first ()
  {
    ptr<T> ret;
    if (_lst.first) {
      ret = _lst.first->elem ();
    }
    return ret;
  }
  
  void insert_head (ptr<T> p) {
    _lst.insert_head (New mynode_t (p));
  }
  
  void remove (ptr<T> p) {
    void *v = (p->*field).v_node ();
    mynode_t *x = static_cast<mynode_t *> (v);
    _lst.remove (x);
    (p->*field).remove_from_list ();
    delete x;
  }
  
  void delete_all ()
  {
    while (_lst.first) {
      remove (_lst.first->elem ());
    }
  }
  
  static ptr<T> next (ptr<T> p) {
    void *v = (p->*field).v_node ();
    mynode_t *x = static_cast<mynode_t *> (v);
    x = mylist_t::next (x);
    
    ptr<T> ret;
    
    if (x) {
      ret = x->elem ();
    }
    
    return ret;
  }
  
  void traverse (typename callback<void, ptr<T> >::ref cb) const {
    mynode_t *p, *np;
    for (p = _lst.first; p; p = np) {
      np = _lst.next (p);
      (*cb) (p->elem ());
    }
  }
  
private:
  mylist_t _lst;
};

  
#if 0

template<class T, list_entry<T> T::*field>
class rclist_t : private list<T,field> {
  rclist_t (const rclist_t &);
  rclist_t &operator= (const rclist_t &);
public:
  rclist_t () : list<T,field> () {}

  ptr<T> get_first () { return mkref (this->first); }

  static ptr<T> next (const T *elm) {
    return mkref (list<T, field>::next (elm));
  }

  void insert_head (T *elm) { 
    elm->refcount_inc ();
    list<T, field>::insert_head (elm);
  }

  static ptr<T> remove (T *elm) {
    ptr<T> x = mkref (list<T, field>::remove (elm));
    elm->refcount_dec ();
    return x;
  }

  void traverse (typename callback<void, ptr<T> >::ref cb) const {
    T *p, *np;
    for (p = this->first; p; p = np) {
      np = (p->*field).next;
      (*cb) (mkref (p));
    }
  }

  void delete_all () {
    T *x;
    while ((x = this->first)) {
      remove (x);
    }
  }

};

#endif

#endif /* _ASYNC_RCLIST_H_INCLUDED_ */
