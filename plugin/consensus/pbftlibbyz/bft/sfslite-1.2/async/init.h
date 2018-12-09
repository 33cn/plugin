// -*-c++-*-
/* $Id: init.h 1117 2005-11-01 16:20:39Z max $ */

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

#ifndef _ASYNC_INIT_H_
#define _ASYNC_INIT_H_ 1

#if __GNUC__ >= 3
# define __init_attribute__(x)
#else /* gcc < 3 */
# define __init_attribute__(x) __attribute__ (x)
#endif /* gcc < 3 */

#define INIT(name)				\
static class name {				\
  static int count;				\
  static int &cnt () { return count; }		\
  static void start ();				\
  static void stop ();				\
public:						\
  name () {if (!cnt ()++) start ();}		\
  ~name () {if (!--cnt ()) stop ();}		\
} init_ ## name __init_attribute__ ((unused))

class initfn {
  initfn ();
public:
  initfn (void (*fn) ()) { fn (); }
};
#define INITFN(fn)				\
static void fn ();				\
static initfn init_ ## fn (fn) __init_attribute__ ((unused))

class exitfn {
  void (*const fn) ();
public:
  exitfn (void (*fn) ()) : fn (fn) {}
  ~exitfn () { fn (); }
};
#define EXITFN(fn)					\
static void fn ();					\
static exitfn exit_ ## fn (fn) __init_attribute__ ((unused))

#endif /* !_ASYNC_INIT_H_ */
