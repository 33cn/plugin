// -*-c++-*-
/* $Id: itree.h 1117 2005-11-01 16:20:39Z max $ */

/*
 *
 * Copyright (C) 1998 David Mazieres (dm@uun.org)
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


#ifndef _ITREE_H_
#define _ITREE_H_ 1

#include "callback.h"
#include "keyfunc.h"

struct __opaquecontainer;
#define oc __opaquecontainer_pointer
typedef __opaquecontainer *oc;

enum itree_color {INVALID, BLACK, RED};

void itree_insert (oc *, oc, const int, int (*) (void *, oc, oc), void *);
void itree_delete (oc *, oc, const int);
oc itree_successor (oc, const int);
oc itree_predecessor (oc, const int);
void __itree_check(oc *, const int, int (*) (void *, oc, oc), void *);
#ifdef ITREE_DEBUG
#define itree_check(r, os, cmpfn) __itree_check(r, os, cmpfn)
#else /* !ITREE_DEBUG */
#define itree_check(r, os, cmpfn)
#endif /* !ITREE_DEBUG */

struct __itree_entry_private {
  oc rbe_up;
  oc rbe_left;
  oc rbe_right;
  enum itree_color rbe_color;
};

template<class T>
#ifndef NO_TEMPLATE_FRIENDS
class
#else /* NO_TEMPLATE_FRIENDS */
struct
#endif /* NO_TEMPLATE_FRIENDS */
itree_entry {
  __itree_entry_private p;
#ifndef NO_TEMPLATE_FRIENDS
  /* Let's get friendly with the implementation... */
  template<class U, itree_entry<U> U::*field, class C>
  friend class itree_core;
#endif /* NO_TEMPLATE_FRIENDS */
public:
  T *up () const { return (T *) p.rbe_up; }
  T *left () const { return (T *) p.rbe_left; }
  T *right () const { return (T *) p.rbe_right; }
};

template<class T, itree_entry<T> T::*field, class C = compare<T> >
class itree_core {
protected:
  oc rb_root;

  const C cmp;

  static int scmp (void *t, oc a, oc b) {
    return (((itree_core<T, field, C> *) t)->cmp) (*(T *) a, *(T *) b);
  }

  /* No copying */
  itree_core (const itree_core &);
  itree_core &operator = (const itree_core &);

#define eos ((ptrdiff_t) &(((T *) 0)->*field).p)
#define cmpfn scmp, (void *) this

  void _deleteall_correct (T *n)
  {
    if (n) {
      _deleteall_correct (left (n));
      _deleteall_correct (right (n));
      delete n;
    }
  }

public:
  itree_core () { clear (); }
  itree_core (const C &c) : cmp (c) { clear (); }

  // MK 7/6/05: deleteall() is fast but broken; accesses freed memory;
  // deleteall_correct () is slow but should be safer.
  // DM: deleteall () possibly fixed
  void deleteall_correct ()
  {
    _deleteall_correct (root ());
    clear ();
  }

  static T *min_postorder (T *n) {
    T *nn;
    if (n)
      while ((nn = left (n)) || (nn = right (n)))
        n = nn;
    return n;
  }
  static T *next_postorder (const T *n) {
    T *nn = up (n), *nnr;
    if (nn && (nnr = right (nn)) && n != nnr)
      return min_postorder (nnr);
    return nn;
  }
  void deleteall () {
    T *n, *nn;
    for (n = min_postorder (root ()); n; n = nn) {
      nn = next_postorder (n);
      delete n;
    }
    clear ();
  }

  void clear () {rb_root = NULL;}

  T *root () const { return (T *) rb_root; }
  static T *up (const T *n) { return (n->*field).up (); }
  static T *left (const T *n) { return (n->*field).left (); }
  static T *right (const T *n) { return (n->*field).right (); }

  T *first () const {
    T *n, *nn;
    for (n = root (); n && (nn = left (n)); n = nn)
      ;
    return n;
  }
  static T *next (const T *n) { return (T *) itree_successor ((oc) n, eos); }
  static T *prev (const T *n) { return (T *) itree_predecessor ((oc) n, eos); }

  void insert (T *n) {
    itree_insert (&rb_root, (oc) n, eos, cmpfn);
    itree_check (&rb_root, eos, cmpfn);
  }
  void remove (T *n) {
    itree_delete (&rb_root, (oc) n, eos);
    itree_check (&rb_root, eos, cmpfn);
  }

  T *search (typename callback<int, const T*>::ref cb) const {
    T *ret = NULL;
    T *n = root ();
    
    while (n) {
      int srchres = (*cb) (n);
      if (srchres < 0)
        n = left (n);
      else if (srchres > 0)
        n = right (n);
      else {
	/* In case there are duplicate keys, keep looking for the first one */
	ret = n;
	n = left (n);
      }
    }
    return ret;
  }

  // XXX - template search to work around egcs 1.2 bug
  template<class A1, class A2>
  T *search (int (*cb) (const A1 *, const A2 *, const T*),
	     const A1 *a1, const A2 *a2) const {
    T *ret = NULL;
    T *n = root ();
    
    while (n) {
      int srchres = (*cb) (a1, a2, n);
      if (srchres < 0)
        n = left (n);
      else if (srchres > 0)
        n = right (n);
      else {
	/* In case there are duplicate keys, keep looking for the first one */
	ret = n;
	n = left (n);
      }
    }
    return ret;
  }

  void traverse (typename callback<void, T *>::ref cb) {
    T *n, *nn;
    for (n = first (); n; n = nn) {
      nn = next (n);
      (*cb) (n);
    }
  }
  void traverse (void (T::*fn) ()) {
    T *n, *nn;
    for (n = first (); n; n = nn) {
      nn = next (n);
      (n->*fn) ();
    }
  }
#undef eos
#undef cmpfn
};
#undef oc

template<class K, class V, K V::*key,
  itree_entry<V> V::*field, class C = compare<K> >
class itree
  : public itree_core<V, field, keyfunc_2<int, V, K, key, C> >
{
  typedef keyfunc_2 <int, V, K, key, C> cmp_t;
  typedef itree_core<V, field, cmp_t> core_t;
  const C kcmp;

#if 0
  template<class T> int kvcmp (const T *k, const V *v)
    { return kcmp (*k, v->*key); }
#else
  int kvcmp (const K *k, const V *v)
    { return kcmp (*k, v->*key); }
#endif
  static int skvcmp (const C *c, const K *k, const V *v)
    { return (*c) (*k, v->*key); }

public:
  itree () {}
  itree (const C &c) : core_t (cmp_t (c)), kcmp (c) {}

#if 0
  template<class T> V *operator[] (const T &k) {
    int (itree::*fn) (const T *, const V *) = &kvcmp<T>;
    return search (wrap (this, fn, &k));
  }
#else
  V *operator[] (const K &k) {
    // return search (wrap (this, &kvcmp, &k));
    return search (skvcmp, &kcmp, &k);
  }
#endif
};

#endif /* !_ITREE_H_ */
