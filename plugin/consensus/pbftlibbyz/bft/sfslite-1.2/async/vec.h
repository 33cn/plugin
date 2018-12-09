// -*- c++ -*-
/* $Id: vec.h 3352 2008-05-30 14:50:22Z max $ */

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

#ifndef _VECTOR_H_INCLUDED_
#define _VECTOR_H_INCLUDED_ 1

#include "opnew.h"
#include "stllike.h"
#include "array.h"
#include "msb.h"

size_t vec_resize_fn (u_int nwanted, u_int nalloc, int objid);
class vec_resizer_t {
public:
   virtual ~vec_resizer_t () {}
   virtual size_t resize (u_int nalloc, u_int nwanted, int objid) = 0;
};
void set_vec_resizer (vec_resizer_t *v);

template<class T>
struct vec_obj_id_t 
{
  vec_obj_id_t () {}
  int operator() (void) const { return 0; }
};

template<class T, size_t N> struct vec_base {
  typedef typename toarray(T) elm_t;
protected:
  elm_t *basep;
  elm_t *firstp; 
  elm_t *lastp; 
  elm_t *limp;

  union {
    double alignment_hack;
    char defbuf_space[N * sizeof (elm_t)];
  };
  elm_t *def_basep () { return reinterpret_cast<elm_t *> (defbuf_space); }
  elm_t *def_limp () { return def_basep () + N; }
  void bfree (T *t) { if ( t != def_basep ()) xfree (t); }
};

template<class T> struct vec_base<T, 0> {
  typedef typename toarray(T) elm_t;
protected:
  elm_t *basep;
  elm_t *firstp; 
  elm_t *lastp; 
  elm_t *limp;

  elm_t *def_basep () { return NULL; }
  elm_t *def_limp () { return NULL; }
  void bfree (T *t) { xfree (t); }

public:
#define doswap(x) tmp = x; x = v.x; v.x = tmp
  void swap (vec_base &v) {
    elm_t *tmp;
    doswap (basep);
    doswap (firstp);
    doswap (lastp);
    doswap (limp);
  };
#undef doswap
};

template<class T, size_t N = 0> class vec : public vec_base<T, N> {
protected:
  typedef vec_base<T, N> base_t;
  typedef typename base_t::elm_t elm_t;
  using base_t::basep;
  using base_t::firstp;
  using base_t::lastp;
  using base_t::limp;
  using base_t::def_basep;
  using base_t::def_limp;

  vec_obj_id_t<T> _vec_obj_id;

  void move (elm_t *dst) {
    if (dst == firstp)
      return;
    assert (dst < firstp || dst >= lastp);
    basep = dst;
    for (elm_t *src = firstp; src < lastp; src++) {
      new ((void *) (dst++)) elm_t (*src);
      src->~elm_t ();
    }

    size_t n_elem = lastp - firstp;
    firstp = basep;
    lastp = firstp + n_elem;
  }

  static elm_t &construct (elm_t &e)
    { return *new (implicit_cast<void *> (&e)) elm_t; }
  static elm_t &cconstruct (elm_t &e, const elm_t &v)
    { return *new (implicit_cast<void *> (&e)) elm_t (v); }
  static void destroy (elm_t &e) { e.~elm_t (); }

  void init () { lastp = firstp = basep = def_basep (); limp = def_limp (); }
  void del () { while (firstp < lastp) firstp++->~elm_t (); bfree (basep); }

#define append(v)						\
do {								\
  reserve (v.size ());						\
  for (const elm_t *s = v.base (), *e = v.lim (); s < e; s++)	\
    cconstruct (*lastp++, *s);					\
} while (0)

#ifdef CHECK_BOUNDS
#define zcheck() assert (lastp > firstp)
#define bcheck(n) assert ((size_t) (n) < (size_t) (lastp - firstp))
#define pcheck(n) assert ((size_t) (n) <= (size_t) (lastp - firstp))
#else /* !CHECK_BOUNDS */
#define zcheck()
#define bcheck(n)
#define pcheck(n)
#endif /* !CHECK_BOUNDS */

public:
  vec () { init (); }
  vec (const vec &v) { init (); append (v); }
  template<size_t NN> vec (const vec<T, NN> &v) { init (); append (v); }
  ~vec () { del (); }
  void clear () { del (); init (); }

  vec &operator= (const vec &v)
    { if (this != &v) { clear (); append (v); } return *this; }
  template<size_t NN> vec &operator= (const vec<T, NN> &v)
    { clear (); append (v); return *this; }
  template<size_t NN> vec &operator+= (const vec<T, NN> &v)
    { append(v); return *this; }

  void reserve (size_t n) {
    if (lastp + n <= limp)
      return;
    size_t nalloc = limp - basep;
    size_t nwanted = lastp - firstp + n;
    if (nwanted > nalloc / 2) {
      nalloc = vec_resize_fn (nalloc, nwanted, _vec_obj_id());
      elm_t *obasep = basep;
      move (static_cast<elm_t *> (txmalloc (nalloc * sizeof (elm_t))));
      limp = basep + nalloc;
      bfree (obasep);
    }
    else
      move (basep);
  }
  void setsize (size_t n) {
    size_t s = size ();
    if (n < s)
      popn_back (s - n);
    else if ((n -= s)) {
      reserve (n);
      elm_t *sp = lastp;
      lastp += n;
      while (sp < lastp)
        construct (*sp++);
    }
  }

  elm_t *base () { return firstp; }
  const elm_t *base () const { return firstp; }
  elm_t *lim () { return lastp; }
  const elm_t *lim () const { return lastp; }
  size_t size () const { return lastp - firstp; }
  bool empty () const { return lastp == firstp; }

  elm_t &front () { zcheck (); return *firstp; }
  const elm_t &front () const { zcheck (); return *firstp; }
  elm_t &back () { zcheck (); return lastp[-1]; }
  const elm_t &back () const { zcheck (); return lastp[-1]; }
  
  elm_t &operator[] (ptrdiff_t i) { bcheck (i); return firstp[i]; }
  const elm_t &operator[] (ptrdiff_t i) const { bcheck (i); return firstp[i]; }

  elm_t &push_back () { reserve (1); return construct (*lastp++); }
  elm_t &push_back (const elm_t &e)
    { reserve (1); return cconstruct (*lastp++, e); }

  elm_t pop_back () { zcheck (); return destroy_return (*--lastp); }
  void popn_back (size_t n) {
    pcheck (n);
    elm_t *sp = lastp;
    lastp -= n;
    while (sp > lastp)
      destroy (*--sp);
  }

  elm_t pop_front () { zcheck (); return destroy_return (*firstp++); }
  void popn_front (size_t n) {
    pcheck (n);
    elm_t *sp = firstp;
    firstp += n;
    while (sp < firstp)
      destroy (*sp++);
  }

#undef zcheck
#undef bcheck
#undef pcheck
#undef append
};

template<class T> void
swap (vec<T> &a, vec<T> &b)
{
  a.swap (b);
}

#endif /* !_VECTOR_H_INCLUDED_ */
