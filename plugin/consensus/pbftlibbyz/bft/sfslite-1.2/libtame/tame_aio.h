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

#ifndef _LIBTAME_TAME_AIO_H_
#define _LIBTAME_TAME_AIO_H_

#include "tame.h"
#include "aiod.h"

/*
 * A library class that puts a wrapper around SFS's aiod interface.
 *
 * In general, create an aiod.  Make a new aiofh_t below, and then
 * can call open/read/lseek/fstat and close on it.  Other features
 * may be added in the future.
 *
 */
namespace tame {

  typedef event<ptr<aiobuf>, ssize_t>::ref aio_read_ev_t;
  typedef event<struct stat *, int>::ref aio_stat_ev_t;

  class aiofh_t {
  public:
    aiofh_t (aiod *a) : _aiod (a), _off (0) {}
    ~aiofh_t ();
    void open (const str &fn, int flg, int mode, evi_t ev, CLOSURE);
    void read (size_t sz, aio_read_ev_t ev, CLOSURE);
    void lseek (off_t o) { _off = o; }
    void fstat (aio_stat_ev_t ev) { _fh->fstat (ev); }
    void close (evi_t::ptr ev = NULL, CLOSURE);

  private:
    aiod *_aiod;
    ptr<aiofh> _fh;
    ptr<aiobuf> _buf;
    size_t _bufsz;
    off_t _off;
    str _fn;
  };

  typedef event<ptr<aiofh_t> >::ref open_ev_t;

};



#endif /* _LIBTAME_TAME_AIO_H_ */
