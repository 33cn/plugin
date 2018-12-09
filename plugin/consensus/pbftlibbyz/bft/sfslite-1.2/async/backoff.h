// -*-c++-*-
/* $Id: backoff.h 1117 2005-11-01 16:20:39Z max $ */

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


#ifndef _BACKOFF_H_
#define _BACKOFF_H_ 1

template<class T>
#ifndef NO_TEMPLATE_FRIENDS
class
#else /* NO_TEMPLATE_FRIENDS */
struct
#endif /* NO_TEMPLATE_FRIENDS */
tmoq_entry {
  u_int qno;
  time_t tm;
  T *next;
  T **pprev;

public:
  tmoq_entry () : qno ((u_int) -1) {}

#ifndef NO_TEMPLATE_FRIENDS
  template<class U, tmoq_entry<U> U::*field,
    u_int minto, u_int maxsend> friend class tmoq;
#endif /* NO_TEMPLATE_FRIENDS */
};

template<class T, tmoq_entry<T> T::*field, u_int minto = 2, u_int maxsend = 5>
class tmoq {
  struct head {
    T *first;
    T **plast;
    head () {first = NULL; plast = &first;}
  };

  head queue[maxsend];
  bool pending[maxsend];

  static void tcb (tmoq *tq, u_int qn) {
    tq->pending[qn] = false;
    tq->runq (qn);
    tq->schedq (qn);
  }

  void schedq (u_int qn) {
    T *p;
    if (!pending[qn] && (p = queue[qn].first)) {
      pending[qn] = true;
      timecb ((p->*field).tm, wrap (tcb, this, qn));
    }
  }

  void runq (u_int qn) {
    time_t now = time (NULL);
    T *p;

    while ((p = queue[qn].first) && (p->*field).tm <= now) {
      remove (p);
      if (qn < maxsend - 1)
	insert (p, qn + 1, now);
      else {
	(p->*field).qno = maxsend;
	p->timeout ();
      }
    }
  }

  void insert (T *p, u_int qn, time_t now = 0) {
    (p->*field).qno = qn;
    (p->*field).tm = (now ? now : time (NULL)) + (minto << qn);
    (p->*field).next = NULL;
    (p->*field).pprev = queue[qn].plast;
    *queue[qn].plast = p;
    queue[qn].plast = &(p->*field).next;
    schedq (qn);
    p->xmit (qn);
  }

public:
  tmoq () {
    for (size_t i = 0; i < maxsend; i++)
      pending[i] = false;
  }

  void start (T *p) { insert (p, 0); }

  void keeptrying (T *p) {
    assert ((p->*field).qno >= maxsend);
    insert (p, maxsend - 1);
  }

  void remove (T *p) {
    if ((p->*field).qno < maxsend) {
      if ((p->*field).next)
	((p->*field).next->*field).pprev = (p->*field).pprev;
      else
	queue[(p->*field).qno].plast = (p->*field).pprev;
      *(p->*field).pprev = (p->*field).next;
    }
  }
};

#endif /* !_BACKOFF_H_ */
