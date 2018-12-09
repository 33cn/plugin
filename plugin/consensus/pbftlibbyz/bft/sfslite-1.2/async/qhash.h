// -*-c++-*-
/* $Id: qhash.h 3807 2008-11-19 17:49:56Z max $ */

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

#ifndef _QHASH_H_
#define _QHASH_H_ 1

#include "ihash.h"

template<class T> struct qhash_lookup_return {
  typedef T *type;
  typedef const T *const_type;
  static type ret (T *v) { return v; }
  static const_type const_ret (T *v) { return v; }
};
template<class T> struct qhash_lookup_return<ref<T> > {
  typedef ptr<T> type;
  typedef ptr<const T> const_type;
  static type ret (ref<T> *v) { if (v) return *v; return NULL; }
  static const_type const_ret (ref<T> *v) { if (v) return *v; return NULL; }
};
template<class T> struct qhash_lookup_return<T &> {
  typedef T *type;
  typedef const T *const_type;
  static type ret (T *v) { return v; }
  static const_type const_ret (T *v) { return v; }
};

template<class K, class V> struct qhash_slot {
  ihash_entry<qhash_slot> link;
  const K key;
  V value;
  qhash_slot (const K &k, typename CREF (V) v) : key (k), value (v) {}
  qhash_slot (const K &k, typename NCREF (V) v) : key (k), value (v) {}
};

template<class K, class V, class H = hashfn<K>, class E = equals<K>,
  // XXX - We need to kludge this for both g++ and KCC
  class R = qhash_lookup_return<V>,
  ihash_entry<qhash_slot<K, V> > qhash_slot<K,V>::*kludge
	= &qhash_slot<K,V>::link>
class qhash
  : public ihash_core<qhash_slot<K, V>, kludge> {
public:
  typedef qhash_slot<K, V> slot;
  typedef ihash_core<slot, kludge> core;

private:
  const E eq;
  const H hash;

  slot *getslot (const K &k) const {
    slot *s;
    for (s = lookup_val (hash (k)); s && !eq (s->key, k); s = next_val (s))
      ;
    return s;
  }
  void delslot (slot *s) { core::remove (s); delete s; }

  void copyslot (const slot &s) { insert (s.key, s.value); }

  static void mkcb (ref<callback<void, K, typename R::type> > cb, slot *s)
    { (*cb) (s->key, R::ret (&s->value)); }
  static void mkcbr (ref<callback<void, const K &, typename R::type> > cb,
		     slot *s)
    { (*cb) (s->key, R::ret (&s->value)); }

public:
  qhash () : eq (E ()), hash (H ()) {}
  qhash (const qhash<K,V,H,E,R> &in)
  {
    in.core::traverse (wrap (this, &qhash::copyslot));
  }
  void clear () {
    core::traverse (wrap (this, &qhash::delslot));
    core::clear ();
  }
  ~qhash () { clear (); }

  qhash<K,V,H,E,R> &operator= (const qhash<K,V,H,E,R> &in)
  {
    clear ();
    in.core::traverse (wrap (this, &qhash::copyslot));
    return *this;
  }

#if 0
  void traverse (callback<void, K, typename R::type>::ref cb)
    { core::traverse (wrap (mkcb, cb)); }
#endif
  void traverse (ref<callback<void, const K &, typename R::type> > cb)
    { core::traverse (wrap (mkcbr, cb)); }

  void insert (const K &k) {
    if (slot *s = getslot (k))
      s->value = V ();
    else
      core::insert_val (New slot (k, V ()), hash (k));
  }
  void insert (const K &k, typename CREF (V) v) {
    if (slot *s = getslot (k))
      s->value = v;
    else
      core::insert_val (New slot (k, v), hash (k));
  }
  void insert (const K &k, typename NCREF (V) v) {
    if (slot *s = getslot (k))
      s->value = v;
    else
      core::insert_val (New slot (k, v), hash (k));
  }
  void remove (const K &k) {
    if (slot *s = getslot (k))
      delslot (s);
  }
  typename R::type operator[] (const K &k) {
    if (slot *s = getslot (k))
      return R::ret (&s->value);
    else
      return R::ret (NULL);
  }

  bool remove (const K &k, V *v) {
    if (slot *s = getslot (k)) {
      *v = s->value;
      delslot (s);
      return true;
    } else {
      return false;
    }
  }

  typename R::const_type operator[] (const K &k) const {
    if (slot *s = getslot (k))
      return R::const_ret (&s->value);
    else
      return R::const_ret (NULL);
  }
};

template<class K> struct qhash_slot<K, void> {
  ihash_entry<qhash_slot> link;
  const K key;
  qhash_slot (const K &k) : key (k) {}
};

template<class K, class H = hashfn<K>, class E = equals<K>,
  // XXX - We need to kludge this for both g++ and KCC
  ihash_entry<qhash_slot<K, void> > qhash_slot<K, void>::*kludge
	= &qhash_slot<K, void>::link>
class bhash // <K, void, H, E, kludge>
  : public ihash_core<qhash_slot<K, void>, kludge> {
public:
  typedef qhash_slot<K, void> slot;
  typedef ihash_core<slot, kludge> core;

private:
  const E eq;
  const H hash;

  slot *getslot (const K &k) const {
    slot *s;
    for (s = lookup_val (hash (k)); s && !eq (s->key, k); s = next_val (s))
      ;
    return s;
  }
  void delslot (slot *s) { core::remove (s); delete s; }

  void copyslot (const slot &s) { insert (s.key); }

  static void mkcb (ref<callback<void, K> > cb, qhash_slot<K, void> *s)
    { (*cb) (s->key); }
  static void mkcbr (ref<callback<void, const K &> > cb,
		     qhash_slot<K, void> *s)
    { (*cb) (s->key); }

public:
  bhash () {}
  void clear () { this->deleteall (); }
  ~bhash () { clear (); }

  bhash (const bhash<K,H,E> &in)
  {
    in.core::traverse (wrap (this, &bhash::copyslot));
  }

  bhash<K,H,E> &operator= (const bhash<K,H,E> &in)
  {
    clear ();
    in.core::traverse (wrap (this, &bhash::copyslot));
    return *this;
  }

  bool insert (const K &k) {
    if (!getslot (k)) {
      core::insert_val (New slot (k), hash (k));
      return true;
    }
    return false;
  }
  void remove (const K &k) { if (slot *s = getslot (k)) delslot (s); }
  bool operator[] (const K &k) const { return getslot (k); }

#if 0
  void traverse (callback <void, K>::ref cb)
    { core::traverse (wrap (mkcb, cb)); }
#endif
  void traverse (ref<callback <void, const K &> > cb)
    { core::traverse (wrap (mkcbr, cb)); }
};

template<class K, class V, class H = hashfn<K> , class E = equals<K> >
class qhash_const_iterator_t {
public:
  qhash_const_iterator_t (const qhash<K,V,H,E> &q) 
    : _i (q.first ()), _qh (q) {}
  const K *next (V *val = NULL) {
    const K *r = NULL;
    if (_i) {
      if (val) *val = _i->value;
      r = &_i->key;
      _i = _qh.next (_i);
    }
    return r;
  }
private:
  const qhash_slot<K,V> *_i;
  const qhash<K,V,H,E> &_qh;
};

template<class K, class V, class H = hashfn<K>, class E = equals<K> >
class qhash_iterator_t {
public:
  qhash_iterator_t (qhash<K,V,H,E> &q) : _i (q.first ()), _qh (q) {}
  const K *next (V *val = NULL) {
    const K *r = NULL;
    if (_i) {
      if (val) *val = _i->value;
      r = &_i->key;
      _i = _qh.next (_i);
    }
    return r;
  }
private:
  qhash_slot<K,V> *_i;
  qhash<K,V,H,E> &_qh;
};

template<class K, class H = hashfn<K>, class E = equals<K> >
class bhash_const_iterator_t {
public:
  bhash_const_iterator_t (const bhash<K,H,E> &h) 
    : _i (h.first ()), _bh (h) {}

  const K *next () {
    const K *r = NULL;
    if (_i) {
      r = &_i->key;
      _i = _bh.next (_i);
    }
    return r;
  }
  
private:
  const qhash_slot<K,void> *_i;
  const bhash<K,H,E> &_bh;
};

#endif /* !_QHASH_H_ */
