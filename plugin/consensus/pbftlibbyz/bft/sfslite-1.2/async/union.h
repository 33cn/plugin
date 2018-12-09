// -*-c++-*-
/* $Id: union.h 1117 2005-11-01 16:20:39Z max $ */

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


#ifndef _ARPC_TAGUNION_INCLUDED_
#define _ARPC_TAGUNION_INCLUDED_ 1

#include "opnew.h"
#include "array.h"
#include <typeinfo>

#ifndef TUNION_DEBUG
# if defined (CHECK_BOUNDS) || defined (DMALLOC)
#  define TUNION_DEBUG 1
# endif /* CHECK_BOUNDS || DMALLOC */
#endif /* !TUNION_DEBUG */

#if 0
/* For a contant n, C++ defines sizeof (T[n]) to be n * sizeof (T).
 * Assuming 8 is the maximum alignment requirement, this means any
 * object T's maximum alignment requirment is sizeof (T) % 8.
 * AlignmentHack (T) can be used to declare a primitive type with the
 * same alignment requirements.  */
template<size_t n> struct __alignment_hack;
template<> struct __alignment_hack<0> { typedef double type; };
template<> struct __alignment_hack<1> { typedef char type; };
template<> struct __alignment_hack<2> { typedef u_int16_t type; };
template<> struct __alignment_hack<3> { typedef char type; };
template<> struct __alignment_hack<4> { typedef u_int32_t type; };
template<> struct __alignment_hack<5> { typedef char type; };
template<> struct __alignment_hack<6> { typedef u_int16_t type; };
template<> struct __alignment_hack<7> { typedef char type; };
#define AlignmentHack(T) typename ::__alignment_hack<(sizeof(T)%8)>::type
#else
#define AlignmentHack(T) u_int64_t
#endif

class union_entry_base {
protected:
  struct vtbl {
    const std::type_info *type;
    void (*destructor) (union_entry_base *);
    void (*assignop) (union_entry_base *, const union_entry_base *);
    size_t size;
  };

  const vtbl *vptr;

public:
  void init () { vptr = NULL; }
  void init (const union_entry_base &b) { init (); assign (b); }
  void destroy () { if (vptr) vptr->destructor (this); init (); }
  void assign (const union_entry_base &b) {
    if (b.vptr)
      b.vptr->assignop (this, &b);
    else 
      destroy ();
  }
  const std::type_info *type () const { return vptr ? vptr->type : NULL; }
};

template<class T> class union_entry : public union_entry_base {
  union {
    AlignmentHack(T) _hack;
    char m[sizeof (T)];
  };

  static const vtbl *getvptr () {
    static const vtbl vt = { &typeid (T), &destructor, &assignop, sizeof (T) };
    return &vt;
  }
  static void destructor (union_entry_base *tb) {
    union_entry *te = static_cast<union_entry *> (tb);
    te->_addr ()->~T ();
  }
  static void assignop (union_entry_base *dstb,
			const union_entry_base *srcb) {
    union_entry *dst = static_cast<union_entry *> (dstb);
    const union_entry *src = static_cast<const union_entry *> (srcb);
    dst->select ();
    (void) (*dst->_addr () = *src->_addr ()); // XXX - cast required by egcs
  }

  T *_addr () { return reinterpret_cast<T *> (m); }
  const T *_addr () const { return reinterpret_cast<const T *> (m); }

  void verify () const {
#if TUNION_DEBUG
    if (!vptr || *vptr->type != typeid (T))
      panic ("union_entry<%s>::verify: accessed when %s selected\n",
	     typeid (T).name (), vptr ? vptr->type->name () : "NULL");
#endif /* TUNION_DEBUG */
  }

public:
  T *addr () { verify (); return _addr (); }
  const T *addr () const { verify (); return _addr (); }
  T *operator-> () { return addr (); }
  const T *operator-> () const { return addr (); }
  T &operator *() { return *addr (); }
  const T &operator *() const { return *addr (); }
  operator T *() { return addr (); }
  operator const T *() const { return addr (); }

  void select () {
    if (!vptr || *vptr->type != typeid (T)) {
      destroy ();
      vptr = getvptr ();
      new (static_cast<void *> (m)) T;
    }
  }
  void Xstompcast () {
#if TUNION_DEBUG
    if (!vptr || vptr->size != sizeof (T))
      panic ("union_entry<%s>::Xstompcast: cast changes object size\n",
	     typeid (T).name ());
#endif /* TUNION_DEBUG */
    vptr = getvptr ();
  }
};

template<> class union_entry<void> : public union_entry_base {
  static const vtbl *getvptr () {
    static const vtbl vt = { &typeid (void), &destructor, &assignop, 0 };
    return &vt;
  }
  static void destructor (union_entry_base *tb) {}
  static void assignop (union_entry_base *dstb,
			const union_entry_base *srcb) {
    union_entry *dst = static_cast<union_entry *> (dstb);
    dst->select ();
  }

public:
  void select () {
    if (!vptr || *vptr->type != typeid (void)) {
      destroy ();
      vptr = getvptr ();
    }
  }
  /* It's okay to cast anything to void. */
  void Xstompcast () { select (); }
};

template<class T, size_t n> class union_entry<T[n]>
  : public union_entry<typename toarray (T[n])> {
  // typedef toarray (T[n]) array_t;
  typename toarray (T) &operator[] (ptrdiff_t i) { return (**this)[i]; }
  const typename toarray (T) &operator[] (ptrdiff_t i) const
    { return (**this)[i]; }
};

#endif /* !_ARPC_TAGUNION_INCLUDED_ */
