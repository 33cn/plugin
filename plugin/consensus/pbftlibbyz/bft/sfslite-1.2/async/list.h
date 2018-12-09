// -*-c++-*-
/* $Id: list.h 1117 2005-11-01 16:20:39Z max $ */

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

#ifndef _LIST_H_INCLUDED_
#define _LIST_H_INCLUDED_ 1

template<class T>
struct list_entry {
  T *next;
  T **pprev;
};

template<class T, list_entry<T> T::*field>
class list {
  // forbid copying
  list (const list &);
  list &operator= (const list &);
public:
  T *first;

  list () { first = NULL; }

  void insert_head (T *elm) {
    if (((elm->*field).next = first))
      (first->*field).pprev = &(elm->*field).next;
    first = elm;
    (elm->*field).pprev = &first;
  }

  static T *remove (T *elm) {
    if ((elm->*field).next)
      ((elm->*field).next->*field).pprev = (elm->*field).pprev;
    *(elm->*field).pprev = (elm->*field).next;
    return elm;
  }

  static T *next (const T *elm) {
    return (elm->*field).next;
  }

  void traverse (typename callback<void, T*>::ref cb) const {
    T *p, *np;
    for (p = first; p; p = np) {
      np = (p->*field).next;
      (*cb) (p);
    }
  }

  void swap (list &b) {
    T *tp = first;
    if ((first = b.first))
      (first->*field).pprev = &first;
    if ((b.first = tp))
      (tp->*field).pprev = &b.first;
  }
};

#if 0
template<class T> inline void
list_remove (T *elm, list_entry<T> T::*field)
{
  list<T, field>::remove (elm);
}
#endif

template<class T>
struct tailq_entry {
  T *next;
  T **pprev;
};

template<class T, tailq_entry<T> T::*field>
struct tailq {
  T *first;
  T **plast;

  tailq () {first = NULL; plast = &first;}

  void insert_head (T *elm) {
    if (((elm->*field).next = first))
      (first->*field).pprev = &(elm->*field).next;
    else
      plast = &(elm->*field).next;
    first = elm;
    (elm->*field).pprev = &first;
  }

  void insert_tail (T *elm) {
    (elm->*field).next = NULL;
    (elm->*field).pprev = plast;
    *plast = elm;
    plast = &(elm->*field).next;
  }

  T *remove (T *elm) {
    if ((elm->*field).next)
      ((elm->*field).next->*field).pprev = (elm->*field).pprev;
    else
      plast = (elm->*field).pprev;
    *(elm->*field).pprev = (elm->*field).next;
    return elm;
  }

  static T *next (T *elm) {
    return (elm->*field).next;
  }

  void traverse (typename callback<void, T *>::ref cb) const {
    T *p, *np;
    for (p = first; p; p = np) {
      np = (p->*field).next;
      (*cb) (p);
    }
  }
};

#endif /* !_LIST_H_INCLUDED_ */
