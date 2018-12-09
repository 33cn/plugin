// -*-c++-*-
/* $Id: async.h 2247 2006-09-29 20:52:16Z max $ */

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

#ifndef _ASYNC_SELECT_H_
#define _ASYNC_SELECT_H_ 1

#include "amisc.h"

//-----------------------------------------------------------------------
//
//  Deal with select(2) call, especially if in the case of PTH threads
//  and wanted to use PTH's select loop instead of the standard libc 
//  variety.
//

void sfs_add_new_cb ();

#ifdef HAVE_TAME_PTH

# include <pth.h>
# define SFS_SELECT sfs_pth_core_select

int sfs_pth_core_select (int nfds, fd_set *rfds, fd_set *wfds,
			 fd_set *efds, struct timeval *timeout);

#else /* HAVE_TAME_PTH */

# define SFS_SELECT select

#endif /* HAVE_TAME_PTH */

//
// end select(2) stuff
//
//-----------------------------------------------------------------------

//-----------------------------------------------------------------------
//
//  Now, figure out the specifics of the select-on-FD operations, and
//  whether that will be using select(2) or is using something fancier
//  like epoll or kqueue.
//

namespace sfs_core {

  //-----------------------------------------------------------------------
  // 
  // API available to users:
  //
  typedef enum { SELECT_NONE, 
		 SELECT_STD, 
		 SELECT_EPOLL, 
		 SELECT_KQUEUE } select_policy_t;

  select_policy_t select_policy_from_str  (const str &s);
  select_policy_t select_policy_from_char (char c);

  void set_busywait (bool b);   
  void set_compact_interval (u_int i);
  int  set_select_policy (select_policy_t i);
  void set_zombie_collect (bool b);

  //
  // end public API
  //-----------------------------------------------------------------------
  
  class selector_t {
  public:
    selector_t ();
    selector_t (selector_t *s);
    virtual ~selector_t ();
    virtual void _fdcb (int, selop, cbv::ptr, const char *, int) = 0;
    virtual void fdcb_check (struct timeval *timeout) = 0;
    virtual int set_compact_interval (u_int i) { return -1; }
    virtual int set_busywait (bool b) { return -1; }
    virtual select_policy_t typ () const = 0;

    static int fd_set_bytes;
    static int maxfd;
    static void init (void);

    cbv::ptr **fdcbs () { return _fdcbs; }

    enum { fdsn = 2  };
  protected:
    cbv::ptr *_fdcbs[fdsn];
  };

  // source code locations
  class src_loc_t {
  public:
    src_loc_t () : _file (NULL), _line (0) {}
    void set (const char *f, int l);
    void clear ();
    str to_str () const;
    const char *file () const { return _file; }
    int line () const { return _line; }
  private:
    const char *_file;
    int _line;
  };

  class std_selector_t : public selector_t {
  public:
    std_selector_t ();
    std_selector_t (selector_t *s);
    ~std_selector_t ();
    void _fdcb (int, selop, cbv::ptr, const char *, int);
    void fdcb_check (struct timeval *timeout);
    int set_compact_interval (u_int i) { _compact_interval = i; return 0; }
    int set_busywait (bool b) { _busywait = b; return 0; }
    select_policy_t typ () const { return SELECT_STD; }

  private:
    
    void compact_nselfd ();
    void init_fdsets ();
    void select_failure ();

    u_int _compact_interval;
    u_int _n_fdcb_iter;
    int _nselfd;
    bool _busywait;

    fd_set *_fdsp[fdsn];
    fd_set *_fdspt[fdsn];

    src_loc_t *_src_locs[fdsn];

    int _last_fd, _last_i, _n_repeats;
  };

#ifdef HAVE_EPOLL
# include <sys/epoll.h>

  class epoll_selector_t : public selector_t {
  public:
    epoll_selector_t (selector_t *cur);
    ~epoll_selector_t ();
    void _fdcb (int, selop, cbv::ptr, const char *, int);
    void fdcb_check (struct timeval *timeout);
    select_policy_t typ () const { return SELECT_EPOLL; }

  private:

    int _epfd;
    struct epoll_event *_ret_events;
    int _maxevents;
    struct epoll_state {
      int  user_events; /* holds bits for READ and WRITE */
      bool in_epoll;
    };
    epoll_state *_epoll_states;

    int user_events_to_epoll_events(epoll_state* es);
    int update_epoll_state(epoll_state* es) ;
  };
#endif /* HAVE_EPOLL */

};

#ifdef HAVE_KQUEUE
# include <sys/types.h>
# include <sys/event.h>
# include <sys/time.h>

namespace sfs_core {

  class kqueue_fd_t {
  public:
    kqueue_fd_t ();
    bool toggle (bool on, const char *file, int line);
    void clear ();
    bool odd_flips () const { return (_flips % 2) == 1; }
    bool any_flips () const { return (_flips > 0); }
    bool on () const { return _on; }
    const char *file () const { return _file; }
    int line () const { return _line; }
    void set_removal_bit () { _removal = !_on; }
    bool removal () const { return _removal; }
  private:
    u_int32_t   _flips;
    bool        _on;
    bool        _removal;
    const char *_file;
    int         _line;
  };

  class kqueue_fd_id_t {
  public:
    kqueue_fd_id_t () : _fd (-1), _op (0) {}
    kqueue_fd_id_t (int f, int o) : _fd (f), _op (o) {}
    bool convert (const struct kevent &kev);
    size_t fd () const { assert (_fd >= 0); return _fd; }
    selop op () const { return selop (_op); }
    int      _fd;
    int      _op;
  };

  class kqueue_fd_set_t {
  public:
    void toggle (bool on, int fd, selop op, const char *file, int line);
    void export_to_kernel (vec<struct kevent> *out);
    const kqueue_fd_t *lookup (const struct kevent &kev) const;
    const kqueue_fd_t *lookup (const kqueue_fd_id_t &id) const;
  private:
    vec<kqueue_fd_id_t>  _active;
    vec<kqueue_fd_t>     _fds[selector_t::fdsn];
  };

  class kqueue_selector_t : public selector_t {
  public:
    kqueue_selector_t (selector_t *t);
    ~kqueue_selector_t ();
    void _fdcb (int, selop, cbv::ptr, const char *, int);
    void fdcb_check (struct timeval *timeout);
    select_policy_t typ () const { return SELECT_KQUEUE; }

    enum { MIN_CHANGE_Q_SIZE = 0x1000 };
  private:
    int _kq;
    kqueue_fd_set_t _set;
    vec<struct kevent> _kq_events_out;
    vec<struct kevent> _kq_changes;
  };
};
#endif

//
//-----------------------------------------------------------------------


#endif /* _ASYNC_SELECT_H_ */
