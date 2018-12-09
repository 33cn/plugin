// -*-c++-*-
/* $Id: stllike.h 1117 2005-11-01 16:20:39Z max $ */

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

#ifndef _STL_LIKE_H_
#define _STL_LIKE_H_

#ifndef USE_STL

/* misc stuff */
#undef min
template <class T>
inline const T &min (const T &a, const T &b)
{
  return b < a ? b : a;
}
#undef max
template <class T>
inline const T &max (const T &a, const T &b)
{
  return a < b ? b : a;
}

#else /* USE_STL */
#include <algorithm>
#endif /* USE_STL */

template<class R> inline R
implicit_cast (R r)
{
  return r;
}

/*
 * The remaining code is additionally covered by this copyright:
 *
 * Copyright (c) 1997
 * Silicon Graphics Computer Systems, Inc.
 *
 * Permission to use, copy, modify, distribute and sell this software
 * and its documentation for any purpose is hereby granted without fee,
 * provided that the above copyright notice appear in all copies and
 * that both that copyright notice and this permission notice appear
 * in supporting documentation.  Silicon Graphics makes no
 * representations about the suitability of this software for any
 * purpose.  It is provided "as is" without express or implied warranty.
 *
 */

template <class T> class auto_ptr {
private:
  T* M_ptr;

public:
  typedef T element_type;
  explicit auto_ptr (T *p = 0) : M_ptr (p) {}
  auto_ptr (auto_ptr &a) : M_ptr (a.release ()) {}
  template <class U> auto_ptr (auto_ptr <U> &a) : M_ptr (a.release ()) {}

  auto_ptr &operator= (auto_ptr& a) {
    if (&a != this) {
      delete M_ptr;
      M_ptr = a.release();
    }
    return *this;
  }
  template <class U> auto_ptr &operator= (auto_ptr<U> &a) {
    if (a.get() != this->get()) {
      delete M_ptr;
      M_ptr = a.release();
    }
    return *this;
  }

  ~auto_ptr () { delete M_ptr; }

  T &operator* () const
    { return *M_ptr; }
  T *operator-> () const
    { return M_ptr; }
  T *get () const
    { return M_ptr; }
  T *release () {
    T *tmp = M_ptr;
    M_ptr = 0;
    return tmp;
  }
  void reset (T *p = 0) {
    delete M_ptr;
    M_ptr = p;
  }

#if 0
  // I don't understand the following

  // According to the C++ standard, these conversions are required.  Most
  // present-day compilers, however, do not enforce that requirement---and, 
  // in fact, most present-day compilers do not support the language 
  // features that these conversions rely on.
private:
  template<class U> struct auto_ptr_ref {
    U *M_ptr;
    auto_ptr_ref (U *p) : M_ptr (p) {}
  };

public:
  auto_ptr (auto_ptr_ref<T> ref) : M_ptr (ref.M_ptr) {}
  template <class U> operator auto_ptr_ref<U> ()
    { return auto_ptr_ref<T> (this.release ()); }
  template <class U> operator auto_ptr<U> ()
    { return auto_ptr<U> (this->release ()) }

#endif /* 0 */
};


#endif /* !_STL_LIKE_H_ */
