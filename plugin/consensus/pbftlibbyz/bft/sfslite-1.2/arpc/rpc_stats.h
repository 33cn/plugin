
// -*-c++-*-
/* $Id: rpc_stats.h 742 2005-04-15 06:19:56Z max $ */

/*
 *
 * Copyright (C) 2008 Tom Quisel (tom@okcupid.com)
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
 *  Authors:
 *      Tom Quisel (tom@okcupid.com)  [primary]
 *      Max Krohn  (max@okcupid.com)  [slight stylistic changes]
 *
 */


#ifndef RPC_STAT_COLLECTOR_H
#define RPC_STAT_COLLECTOR_H

#include <time.h>
#include "qhash.h"
#include "str.h"

class svccb;

namespace rpc_stats {

  struct rpc_stats_t {
    void init(u_int64_t time_delta); 
    
    u_int32_t count;
    u_int64_t time_sum;
    u_int64_t time_squared_sum;
    u_int64_t min_time;
    u_int64_t max_time;
  };

  struct rpc_proc_t {

    bool operator== (const rpc_proc_t &s) const
    { return s.prog == prog && s.vers == vers && s.proc == proc; }

    u_int32_t prog;
    u_int32_t vers;
    u_int32_t proc;
  };
  
}

typedef rpc_stats::rpc_proc_t rpc_stat_proc_t;

/** lets us use rpc_proc_t as a key in a qhash */
template<> struct hashfn<rpc_stat_proc_t> {
    hashfn () {}
    hash_t operator() (const rpc_stat_proc_t &s) const {
      u_int64_t tmp = (((s.vers << 4) + s.proc) << 16) + s.prog;
      return tmp;
    }
};

template<> struct equals<rpc_stat_proc_t> {
    equals () {}
    bool operator() (const rpc_stat_proc_t &s1, const rpc_stat_proc_t &s2) const
    { return s1 == s2; }
};

namespace rpc_stats {

/** Collects time to process and call frequency for RPC handling code. The 
 * data is printed out periodically in a machine parseable format. The point is
 * to identify RPCs which take too long or are called to often. */
  class rpc_stat_collector_t
  {
  public:
    rpc_stat_collector_t ();
    
    void print_info();

    /** Reset stats and m_last_print */
    void reset();
    
    rpc_stat_collector_t &set_active(bool active) 
    { m_active = active; return (*this); }

    rpc_stat_collector_t &set_interval(u_int32_t secs) 
    { m_interval = secs; return (*this); }

    rpc_stat_collector_t &set_n_per_line (size_t n)
    { m_n_per_line = n; return (*this); }
    
    /** Call this at the end of an RPC handler */
    void end_call(svccb *call_obj, const timespec &strt);

  protected:
    bool m_active;
    u_int32_t m_interval;
    timespec m_last_print;
    size_t m_n_per_line;
    qhash<rpc_proc_t, rpc_stats_t> m_stats;

    void output_line (size_t i, const strbuf &p, strbuf &l, bool frc);
  };
  
} // namespace rpc_stats

// Access the singleton stats object
rpc_stats::rpc_stat_collector_t & get_rpc_stats ();

#endif // RPC_STAT_COLLECTOR_H
