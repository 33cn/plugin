
// -*-c++-*-

#ifndef _ASYNC_RCTREE_H_INCLUDED_
#define _ASYNC_RCTREE_H_INCLUDED_ 1

#include "itree.h"


/*
 * rctree_t
 *
 *  This class allows a itree.h-like interface to reference-counted objects,
 *  so that as long as the object is in the list, its refcount is at least
 *  1.
 *
 */

template<class T>
struct rctree_entry_t {
public:
  rctree_entry_t () : _node (NULL) {}
  ~rctree_entry_t () { assert (!_node); }
    
  void put_in_list (void *v)
  {
    /* don't double-insert this object */
    assert (!_node); 
    _node = v;
  }
  void *v_node () { return _node; }
  
  void remove_from_tree ()
  {
    assert (_node);
    _node = NULL;
  }

private:
  void *_node;
};

template<class K, class V, rctree_entry_t<V> V::*field>
class rctree_node_t {
public:
  rctree_node_t (const K &k, ptr<V> v) : 
    _key (k), _elm (v) { (v->*field).put_in_list (this); }

  const K _key;
  itree_entry<rctree_node_t<K,V,field> > _lnk;
  ptr<V> elem () { return _elm; }
  ptr<const V> elem () const { return _elm; }
private:
  ptr<V> _elm;
};

template<class K, class V, rctree_entry_t<V> V::*field, class C = compare<K> >
class rctree_t {
private:
  rctree_t (const rctree_t<K, V, field, C> &r);
  rctree_t &operator = (const rctree_t<K, V, field, C> &r);


public:

  rctree_t () {}
  ~rctree_t () { clear (); }
  
  typedef rctree_node_t<K, V, field> mynode_t;
  typedef itree<const K, mynode_t, &mynode_t::_key, &mynode_t::_lnk> mytree_t;

private:
  
  static mynode_t *get_node (ptr<V> e) 
  { return static_cast<mynode_t *> ( (e->*field).v_node () ); }
  static const mynode_t *get_node (ptr<const V> e) 
  { return static_cast<const mynode_t *> ( (e->*field).v_node () ); }
  
  int search_adapter_cb (typename callback<int, ptr<V> >::ref fn, const V *el)
  {
    return (*fn) (el->elem ());
  }

public:

  ptr<V> root ()
  {
    ptr<V> ret; 
    mynode_t *r = _tree.root ();
    if (r) ret = r->elem (); 
    return ret;
  }
  
#define WALKFN(typ,dir)                                               \
  static ptr<typ V> dir (ptr<typ V> e)                                \
  {                                                                   \
    ptr<typ V> ret;                                                   \
    typ mynode_t *p, *n;                                              \
    if (e && (p = get_node (e)) && (n = mytree_t::dir(p)))            \
      ret = n->elem ();                                               \
    return ret;                                                       \
  }

  WALKFN(,up)
  WALKFN(const,up)
  WALKFN(,right)
  WALKFN(const,right)
  WALKFN(,left)
  WALKFN(const,left)
#undef WALKFN

  void insert (const K &k, ptr<V> n) { _tree.insert (New mynode_t (k, n)); }
  
  void remove (ptr<V> p) {
    mynode_t *n = get_node (p);
    _tree.remove (n);
    (p->*field).remove_from_tree ();
    delete n;
  }

#define POWALKFN(typ,dir)                                                \
  static ptr<typ V> dir (ptr<typ V> in)                                  \
  {                                                                      \
    ptr<typ V> ret;                                                      \
    typ mynode_t *n = NULL;                                              \
    if (in && (n = get_node (in)) && (n = mytree_t::dir (n)))            \
      ret = n->elem ();                                                  \
    return ret;                                                          \
  }

  POWALKFN(, min_postorder);
  POWALKFN(, next_postorder);
  POWALKFN(const, next_postorder);

#undef POWALKFN

  void clear ()
  {
    mynode_t *n, *nn;
    for (n = _tree.min_postorder (_tree.root ()); n; n = nn) {
      nn = _tree.next_postorder (n);
      remove (n->elem ());
    }
    _tree.clear ();
  }

  ptr<V> operator[] (const K &k) {
    ptr<V> ret;
    mynode_t *n = _tree[k];
    if (n) ret = n->elem ();
    return ret;
  }

  ptr<const V> operator[] (const K &k) const {
    ptr<const V> ret;
    const mynode_t *n = _tree[k];
    if (n) ret = n->elem ();
    return ret;
  }
  
  ptr<V> search (typename callback<int, ptr<V> >::ref fn) const 
  {
    ptr<V> ret;
    mynode_t *n = _tree.search (wrap (search_adapter_cb, fn));
    if (n) {
      ret = n->elem ();
    }
    return ret;
  }
  
private:
  mytree_t _tree;
};


#endif /* _ASYNC_RCTREE_H_INCLUDED_ */
