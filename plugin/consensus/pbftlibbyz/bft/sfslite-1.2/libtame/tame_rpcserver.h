
// -*-c++-*-
/* $Id: tame_io.h 2225 2006-09-28 15:41:28Z max $ */

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

#ifndef _LIBTAME_TAME_RPCSERVER_H_
#define _LIBTAME_TAME_RPCSERVER_H_

#include "async.h"
#include "arpc.h"
#include "tame.h"

//
// Tame library functions: wrappers around typical Unix I/O.
// All library functions are in the tame namespace.
//
namespace tame {

  enum { VERB_NONE = 0,
	 VERB_LOW = 10,
	 VERB_MED = 20,
	 VERB_HIGH = 30 };

  class server_t {
  public:
    server_t (int fd, int v);
    virtual ~server_t () {}
    virtual void dispatch (svccb *svp) = 0;
    virtual const rpc_program &get_prog () const = 0;
    void runloop (CLOSURE);
  private:
    ptr<axprt_stream> _x;
    int _verbosity;
  };

  class server_factory_t {
  public:
    server_factory_t () : _verbosity (VERB_LOW) {}
    virtual ~server_factory_t () {}
    virtual server_t *alloc_server (int fd, int v) = 0;
    void new_connection (int fd);
    void run (const str &port, evb_t done);
    void run (u_int port, evb_t done) { run_T (port, done); }
    void set_verbosity (int i) { _verbosity = i; }
  private:
    void run_T (u_int port, evb_t done, CLOSURE);
    int _verbosity;
  };

};

#endif /* _LIBTAME_TAME_RPCSERVER_H_ */
