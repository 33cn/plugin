// -*- c++ -*-
/* $Id: corebench.h 1648 2006-03-25 08:59:48Z max $ */

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

#ifndef _COREBENCH_H_INCLUDED_
#define _COREBENCH_H_INCLUDED_ 1

#if defined (__i386__)

static __inline unsigned long long
i386_rdtsc(void)
{
  unsigned long long rv;
  __asm __volatile(".byte 0x0f, 0x31" : "=A" (rv));
  return (rv);
}

# define corebench_get_time() i386_rdtsc()
# define BENCH_TICK_TYPE "cycles"

#else

inline u_int64_t
get_time ()
{
  timeval tv;
  gettimeofday (&tv, NULL);
  return (u_int64_t) tv.tv_sec * 1000000 + tv.tv_usec;
}

# define corebench_get_time() get_time()
# define BENCH_TICK_TYPE "usec"

#endif /* __I386__ */

#define START_ACHECK_TIMER() \
do { if (do_corebench) tia_tmp = corebench_get_time (); } while(0)

#define STOP_ACHECK_TIMER()  \
do {                         \
  if (do_corebench) {        \
    unsigned long long x = corebench_get_time ();  \
    assert (x > tia_tmp);                          \
    time_in_acheck += (x - tia_tmp);               \
  }                                                \
} while(0)

extern bool do_corebench;
inline void toggle_corebench (bool f) { do_corebench = f; }
extern unsigned long long tia_tmp, time_in_acheck;

#endif /* !_COREBENCH_H_INCLUDED_ */
