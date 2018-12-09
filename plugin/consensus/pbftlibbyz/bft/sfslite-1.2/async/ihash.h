// -*-c++-*-
/* $Id: ihash.h 1693 2006-04-28 23:17:35Z max $ */

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

/* Template implementation of intrusive hash table.  Note that while
 * externally this code is type safe, interally it makes use of void *
 * pointers and byte offsets of field structures.  This permits a
 * single implementation of the _ihash_grow() function. */

#ifndef _IHASH_H_
#define _IHASH_H_ 1

#ifdef DMALLOC
# ifndef IHASH_DEBUG
#  define IHASH_DEBUG 1
# endif /* !IHASH_DEBUG */
# if IHASH_DEBUG
   /* Check hash table consistency if check-funcs token set */
#  define ihash_do_debug() (dmalloc_debug_current () & 0x4000)
#  include "opnew.h"
# endif /* IHASH_DEBUG */
# include <typeinfo>
#endif /* DMALLOC */

#include "callback.h"
#include "keyfunc.h"

template<class T> class ihash_entry;
template<class T, ihash_entry<T> T::*field> class ihash_core;

struct _ihash_table;
extern void _ihash_grow (_ihash_table *, const size_t);

struct _ihash_entry {
  void *next;
  void **pprev;
  hash_t val;
};

/* The template is for consistency accross all interfaces.  We don't
 * actually need it. */
template<class T>
#ifndef NO_TEMPLATE_FRIENDS
class
#else /* NO_TEMPLATE_FRIENDS */
struct
#endif /* NO_TEMPLATE_FRIENDS */
ihash_entry : _ihash_entry {
#ifndef NO_TEMPLATE_FRIENDS
  template<class U, ihash_entry<U> U::*field> friend class ihash_core;
  // template<ihash_entry<T> T::*field> friend class ihash_core<T, field>;
#endif /* NO_TEMPLATE_FRIENDS */
};

struct _ihash_table {
  size_t buckets;
  size_t entries;
  void **tab;
};

template<class T, ihash_entry<T> T::*field>
class ihash_core {
  _ihash_table t;

protected:
  void init () {
    t.buckets = t.entries = 0;
    t.tab = NULL;
    _ihash_grow (&t, (size_t) &(((T *) 0)->*field));
  }
  ihash_core () { init (); }

  bool present (T *elm) {
    for (T *e = lookup_val ((elm->*field).val); e; e = next_val (e))
      if (e == elm)
	return true;
    return false;
  }
  void _check () {
#if IHASH_DEBUG
    if (ihash_do_debug ()) {
      size_t s = 0;
      for (size_t n = 0; n < t.buckets; n++)
	for (T *e = (T *) t.tab[n], *ne; e; e = ne) {
	  ne = (T *) (e->*field).next;
	  assert (n == (e->*field).val % t.buckets);
	  assert (ne != e);
	  s++;
	}
      assert (s == t.entries);
    }
#endif /* IHASH_DEBUG */
  }

  void insert_val (T *elm, hash_t hval) {
#if IHASH_DEBUG
    if (ihash_do_debug () && present (elm))
      panic ("ihash_core(%s)::insert_val: element already in hash table\n",
	     typeid (ihash_core).name ());
#endif /* IHASH_DEBUG */
    _check ();
    if (++t.entries >= t.buckets)
      _ihash_grow (&t, (size_t) (_ihash_entry *) &(((T *) 0)->*field));
    (elm->*field).val = hval;

    size_t bn = hval % t.buckets;
    T *np;
    if ((np = (T *) t.tab[bn]))
      (np->*field).pprev = &(elm->*field).next;
    (elm->*field).next = np;
    (elm->*field).pprev = &t.tab[bn];
    t.tab[bn] = elm;
    _check ();
  }

  T *lookup_val (hash_t hval) const {
    T *elm;
    for (elm = (T *) t.tab[hval % t.buckets];
	 elm && (elm->*field).val != hval;
	 elm = (T *) (elm->*field).next)
      ;
    return elm;
  }

  static T *next_val (T *elm) {
    hash_t hval = (elm->*field).val;
    while ((elm = (T *) (elm->*field).next) && (elm->*field).val != hval)
      ;
    return elm;
  }

public:
  void clear () { delete[] t.tab; init (); }
  ~ihash_core () { delete[] t.tab; }
  void deleteall () {
    for (size_t i = 0; i < t.buckets; i++)
      for (T *n = (T *) t.tab[i], *nn; n; n = nn) {
	nn = (T *) (n->*field).next;
	delete n;
      }
    clear ();
  }
  size_t size () const { return t.entries; }
  bool constructed () const { return t.buckets > 0; }

  void remove (T *elm) {
#if IHASH_DEBUG
    if (ihash_do_debug () && !present (elm))
      panic ("ihash_core(%s)::remove: element not in hash table\n",
	     typeid (ihash_core).name ());
#endif /* IHASH_DEBUG */
    _check ();
    t.entries--;
    if ((elm->*field).next)
      (((T *) (elm->*field).next)->*field).pprev = (elm->*field).pprev;
    *(elm->*field).pprev = (elm->*field).next;
  }

  T *first () const {
    if (t.entries)
      for (size_t i = 0; i < t.buckets; i++)
	if (t.tab[i])
	  return (T *) t.tab[i];
    return NULL;
  }

  T *next (const T *n) const {
    if ((n->*field).next)
      return (T *) (n->*field).next;
    for (size_t i = (n->*field).val % t.buckets; ++i < t.buckets;)
      if (t.tab[i])
	return (T *) t.tab[i];
    return NULL;
  }

  void traverse (typename callback<void, T *>::ref cb) {
    for (size_t i = 0; i < t.buckets; i++)
      for (T *n = (T *) t.tab[i], *nn; n; n = nn) {
	nn = (T *) (n->*field).next;
	(*cb) (n);
      }
  }

  void traverse (typename callback<void, const T &>::ref cb) const {
    for (size_t i = 0; i < t.buckets; i++)
      for (T *n = (T *) t.tab[i], *nn; n; n = nn) {
	nn = (T *) (n->*field).next;
	(*cb) (*n);
      }
  }

  void traverse (void (T::*fn) ()) {
    for (size_t i = 0; i < t.buckets; i++)
      for (T *n = (T *) t.tab[i], *nn; n; n = nn) {
	nn = (T *) (n->*field).next;
	(n->*fn) ();
      }
  }

private:
  /* No copying */
  ihash_core (const ihash_core &);
  ihash_core &operator = (const ihash_core &);
};

template<class K, class V, K V::*key,
  ihash_entry<V> V::*field, class H = hashfn<K>, class E = equals<K> >
class ihash
  : public ihash_core<V, field>
{
  const E eq;
  const H hash;

public:
  ihash () : eq (E ()), hash (H ()) {}
  ihash (const E &e, const H &h) : eq (e), hash (h) {}

  void insert (V *elm) { insert_val (elm, hash (elm->*key)); }

#if 0
  template<class T> V *operator[] (const T &k) const {
#else
  V *operator[] (const K &k) const {
#endif
    V *v;
    for (v = lookup_val (hash (k));
	 v && !eq (k, v->*key);
	 v = next_val (v))
      ;
    return v;
  }

  V *nextkeq (V *v) {
    const K &k = v->*key;
    while ((v = next_val (v)) && !eq (k, v->*key))
      ;
    return v;
  };
};

template<class K1, class K2, class V, K1 V::*key1, K2 V::*key2,
  ihash_entry<V> V::*field, class H = hash2fn<K1, K2>,
  class E1 = equals<K1>, class E2 = equals<K2> >
class ihash2
  : public ihash_core<V, field>
{
  const E1 eq1;
  const E2 eq2;
  const H hash;

public:
  ihash2 () {}
  ihash2 (const E1 &e1, const E2 &e2, const H &h)
    : eq1 (e1), eq2 (e2), hash (h) {}

  void insert (V *elm)
    { insert_val (elm, hash (elm->*key1, elm->*key2)); }

  V *operator() (const K1 &k1, const K2 &k2) const {
    V *v;
    for (v = lookup_val (hash (k1, k2));
	 v && !(eq1 (k1, v->*key1) && eq2 (k2, v->*key2));
	 v = next_val (v))
      ;
    return v;
  }

  V *nextkeq (V *v) {
    const K1 &k1 = v->*key1;
    const K1 &k2 = v->*key2;
    while ((v = next_val (v))
	   && !(eq1 (k1, v->*key1) && eq2 (k2, v->*key2)))
      ;
    return v;
  };
};

template<class V, ihash_entry<V> V::*field,
  class H = hashfn<V>, class E = equals<V> >
class shash
  : public ihash_core<V, field>
{
  const E eq;
  const H hash;

public:
  shash () {}
  shash (const E &e, const H &h) : eq (e), hash (h) {}

  void insert (V *elm) { insert_val (elm, hash (*elm)); }

  V *operator[] (const V &k) const {
    V *v;
    for (v = lookup_val (hash (k));
	 v && !eq (k, *v);
	 v = next_val (v))
      ;
    return v;
  }

  V *nextkeq (V *v) {
    const V &k = *v;
    while ((v = next_val (v)) && !eq (k, *v))
      ;
    return v;
  }
};

template<class V, class K>
struct keyfn {
  keyfn () {}
  const K & operator() (V *v) const { assert (false); return K (); }
};

template<class K, class V, ihash_entry<V> V::*field,
	 class F = keyfn<V, K>, class H = hashfn<K>, class E = equals<K> >
class fhash
  : public ihash_core <V, field>
{
  const E eq;
  const H hash;
  const F keyfn;

public:
  fhash () {}
  fhash (const E &e, const H &h, const F &k) : eq (e), hash (h), keyfn (k) {}

  void insert (V *elm) { insert_val (elm, hash (keyfn (elm))); }

  V *operator[] (const K &k) const {
    V *v;
    for (v = lookup_val (hash (k)); 
	 v && !eq (k, keyfn (v));
	 v = next_val (v)) 
    ;
    return v;
  }
  V *nextkeq (V *v) {
    const K &k = keyfn (v);
    while ((v = next_val (v)) && !eq (k, keyfn (v)))
    ;
    return v;
  }
};


#endif /* !_IHASH_H_ */
