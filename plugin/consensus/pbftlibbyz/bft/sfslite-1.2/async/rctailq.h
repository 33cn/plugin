
// -*-c++-*-

#ifndef _ASYNC_RCTAILQ_H_INCLUDED_
#define _ASYNC_RCTAILQ_H_INCLUDED_ 1

#include "list.h"


/*
 * rctailq_t
 *
 *  This class allows a list.h-like interface to reference-counted objects,
 *  so that as long as the object is in the list, its refcount is at least
 *  1.
 *
 */

template<class T>
struct rctailq_entry_t {
public:
  rctailq_entry_t () : _node (NULL) {}
  ~rctailq_entry_t () { assert (!_node); }
    
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

template<class T, rctailq_entry_t<T> T::*field>
class rctailq_node_t {
public:
  rctailq_node_t (ptr<T> e) : _elm (e) { (e->*field).put_in_list (this); }
  tailq_entry<rctailq_node_t<T,field> > _lnk;
  ptr<T> elem () { return _elm; }
private:
  ptr<T> _elm;
};

template<class T, rctailq_entry_t<T> T::*field>
class rctailq_t {
public:

  rctailq_t () {}
  
  typedef tailq<rctailq_node_t<T,field>, 
	       &rctailq_node_t<T,field>::_lnk > mylist_t;

  typedef rctailq_node_t<T, field> mynode_t;
  
  ptr<T> first ()
  {
    ptr<T> ret;
    if (_lst.first) {
      ret = _lst.first->elem ();
    }
    return ret;
  }

  ptr<T> last ()
  {
    ptr<T> ret;
    if (_lst.plast && *_lst.plast) {
      ret = (*(_lst.plast))->elem ();
    }
    return ret;
  }
  
  void insert_head (ptr<T> p) {
    _lst.insert_head (New mynode_t (p));
  }

  void insert_tail (ptr<T> p) {
    _lst.insert_tail (New mynode_t (p));
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

#endif /* _ASYNC_RCTAILQ_H_INCLUDED_ */
