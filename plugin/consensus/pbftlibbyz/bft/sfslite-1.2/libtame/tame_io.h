
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

#ifndef _LIBTAME_TAME_IO_H_
#define _LIBTAME_TAME_IO_H_

#include <sys/types.h>
#include <sys/socket.h>

//
// Tame library functions: wrappers around typical Unix I/O.
// All library functions are in the tame namespace.
//
namespace tame {

  void clearread (int fd);
  void clearwrite (int fd);
  void waitread (int fd, evv_t cb);
  void waitwrite (int fd, evv_t cb);
  
  void fdcb1(int fd, selop which, evv_t cb, CLOSURE);
  void sigcb1 (int sig, evv_t cb, CLOSURE);
  
  class iofd_t {
  public:
    iofd_t (int fd, selop op) : _fd (fd), _op (op), _on (false) {}
    ~iofd_t () { off (); }
    void on (evv_t cb, CLOSURE);
    void off (bool check = true);
    int fd () const { return _fd; }
  private:
    const int _fd;
    const selop _op;
    bool _on;
  }; 
  
  class iofd_sticky_t {
  public:
    iofd_sticky_t (int fd, selop op) : _fd (fd), _op (op), _on (false) {}
    ~iofd_sticky_t () { finish (); }
    void setev (evv_t ev) { _ev = ev; ev->set_reuse (true); }
    void on ();
    void off ();
    void finish ();
    int fd () const { return _fd; }
  private:
    const int _fd;
    const selop _op;
    bool _on;
    evv_t::ptr _ev;
  };

  void read (int fd, char *buf, size_t sz, evi_t ev, CLOSURE);
  void write (int fd, const char *buf, size_t sz, evi_t ev, CLOSURE);
  void accept (int sockfd, struct sockaddr *addr, socklen_t *addrlen, 
	       evi_t ev, CLOSURE);

  //-----------------------------------------------------------------------

  class proxy_t : public virtual refcount {
  public:
    proxy_t (const str &d = NULL) : 
      _debug_name (d), 
      _debug_level (0), 
      _eof (false) {}

    virtual ~proxy_t () {}

    void go (int infd, int outfd, evv_t ev, CLOSURE);
    bool poke ();
    void set_debug_level (int d) { _debug_level = d; }

  protected:
    virtual bool is_readable () const = 0;
    virtual bool is_writable () const = 0;
    virtual int v_read (int fd) = 0;
    virtual int v_write (int fd) = 0;

    virtual bool is_eof () const { return _eof; }
    virtual void set_eof () { _eof = true; }

    // Can signal errors other than via errno
    virtual bool read_error (str *s) { return false; }
    virtual bool write_error (str *s) { return false; }

    void do_debug (const str &msg) const;

    evv_t::ptr _poke_ev;
    const str _debug_name;
    int _debug_level;
    bool _eof;

  };

  class std_proxy_t : public proxy_t {
  public:
    std_proxy_t (const str &d = NULL, ssize_t sz = -1);
    virtual ~std_proxy_t ();
   
  protected:
    virtual bool is_readable () const;
    virtual bool is_writable () const;
    virtual int v_read (int fd);
    virtual int v_write (int fd);

    size_t room_left () const { return _sz - _buf.resid (); }

    size_t _sz;
    suio _buf;
  };

  void proxy (int in, int out, evv_t cb, CLOSURE);

  //-----------------------------------------------------------------------

};

#endif /* _LIBTAME_TAME_IO_H_ */
