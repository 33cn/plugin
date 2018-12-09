
// -*-c++-*-
/* $Id: tame_core.h 2225 2006-09-28 15:41:28Z max $ */

/*
 *
 * Copyright (C) 2005 Max Krohn (max@okws.org)
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

#ifndef _LIBTAME_TAME_TRIGGER_H_
#define _LIBTAME_TAME_TRIGGER_H_

#include "async.h"
#include "tame_event.h"
#include "tame_event_ag.h"

// For backwards-compat, make the trigger method behave like regular
// SFS used to work
#define TRIGGER(cb, ...) (*cb) (__VA_ARGS__)

void dtrigger (event<>::ref cb);

template<class T1> void
dtrigger (typename event<T1>::ref cb, const T1 &t1)
{
    delaycb (0, 0, wrap (cb, &_event<T1>::operator(), t1));
}

template<class T1, class T2> void
dtrigger (typename event<T1,T2>::ref cb, const T1 &t1, const T2 &t2)
{
    delaycb (0, 0, wrap (cb, &_event<T1,T2>::operator(), t1, t2));
}

template<class T1, class T2, class T3> void
dtrigger (typename event<T1,T2,T3>::ref cb, const T1 &t1, 
	  const T2 &t2, const T3 &t3)
{
    delaycb (0, 0, wrap (cb, &_event<T1,T2,T3>::operator(), t1, t2, t3));
}

#endif /* _LIBTAME_TAME_TRIGGER_H_ */
