
// -*-c++-*-
/* $Id: litetime.h 2679 2007-04-04 20:53:20Z max $ */

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

#ifndef _ASYNC_LITETIME_H_
#define _ASYNC_LITETIME_H_ 1

#include "amisc.h"
#include "init.h"

//
// public
//
#define HAVE_SFS_CLOCK_T 1
typedef enum { SFS_CLOCK_GETTIME = 0, 
	       SFS_CLOCK_MMAP = 1, 
	       SFS_CLOCK_TIMER = 2 } sfs_clock_t;

INIT(litetime_init);


#define HAVE_MY_CLOCK_GETTIME 1

#define TIMESPEC_INC(ts)                  \
  if (++ (ts)->tv_nsec == 1000000000L)  { \
    (ts)->tv_sec ++;                      \
    (ts)->tv_nsec = 0L;                   \
  }                                    

#define TIMESPEC_LT(ts1, ts2)              \
  (((ts1).tv_sec < (ts2).tv_sec) ||       \
   ((ts1).tv_sec == (ts2).tv_sec && (ts1).tv_sec < (ts2).tv_sec))

#define TIMESPEC_EQ(ts1, ts2) \
  ((ts1).tv_sec == (ts2).tv_sec && (ts1).tv_nsec == (ts2).tv_nsec)


//
// Public interface to clock.
//
void sfs_set_clock (sfs_clock_t typ, const str &s = NULL, bool lzy = false);
struct timespec sfs_get_tsnow (bool force = false);
void sfs_get_tsnow (struct timespec *ts, bool force = false);
void sfs_set_global_timestamp ();
void sfs_leave_sel_loop ();
time_t sfs_get_timenow(bool force = false);
//
//

#endif

