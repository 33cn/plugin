// -*-c++-*-
/* $Id: array.h 1117 2005-11-01 16:20:39Z max $ */

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

/* ***** WHY:
 *
 * C arrays (e.g. "int c[64];") vastly complicate some aspects of C++
 * template programming.  Suppose you have a template tmpl with a
 * class parameter T, and T is instantiated with an array:
 *
 *    class elm {
 *      //...
 *    };
 *
 *    template<class T> tmpl {
 *      //...
 *    };
 *
 *    typedef tmpl<elm[64]> problem_t;
 *
 * If, for instance, tmpl generally needs to allocate an object of
 * type T, a function in tmpl might have code like this:
 *
 *    T *objp = new T;
 *
 * However, this won't work when T is elm[64], because the code
 * "new elm[64]" returns an "elm *", not a "(*) elm[64]".
 *
 * Worse yet, any code that uses placement new or calls destructors
 * will not work.  If T is an array, then allocating a "new T" invokes
 * operator new[] rather than operator new, and that generally
 * requires more than sizeof(T) bytes.
 *
 * Finally, a lot of template classes require things like copy
 * constructors or assignment to work, and neither of those does with
 * C arrays.
 *
 *
 * ***** WHAT:
 *
 * The simple solution to all these problems is simply not to use C
 * arrays.  The dirt-simple type "array<type, size>" is simply an
 * array wrapped in a structure.  These arrays can be allocated with
 * the ordinary scalar new, and things like assignment and copy
 * construction will work fine.
 *
 * The macro "toarray" converts a C array type to a template array.
 */

#ifndef _ARRAY_H_WITH_TOARRAY_
#define _ARRAY_H_WITH_TOARRAY_ 1

#include <stddef.h>

template<class T, size_t n> struct array;

template<class T> struct __toarray {
  typedef T type;
};
template<class T, size_t n> struct __toarray<T[n]> {
  typedef array<typename __toarray<T>::type, n> type;
};
#define toarray(T) __toarray<T>::type

template<class T, size_t n> class array {
public:
  typedef typename toarray(T) elm_t;
  enum { nelm = n };

private:
  elm_t a[nelm];

#ifdef CHECK_BOUNDS
  void bcheck (size_t i) const { assert (i < nelm); }
#else /* !CHECK_BOUNDS */
  void bcheck (size_t) const {}
#endif /* !CHECK_BOUNDS */

public:
  static size_t size () { return nelm; }

  elm_t *base () { return a; }
  const elm_t *base () const { return a; }

  elm_t *lim () { return a + nelm; }
  const elm_t *lim () const { return a + nelm; }

  elm_t &operator[] (ptrdiff_t i) { bcheck (i); return a[i]; }
  const elm_t &operator[] (ptrdiff_t i) const { bcheck (i); return a[i]; }
};

#endif /* _ARRAY_H_WITH_TOARRAY_ */
