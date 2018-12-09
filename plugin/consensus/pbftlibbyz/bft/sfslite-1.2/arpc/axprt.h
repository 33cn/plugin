// -*-c++-*-
/* $Id: axprt.h 2881 2007-05-18 19:06:04Z max $ */

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

class axprt : public virtual refcount {
  friend class xhinfo;
  axprt (const axprt &);
  axprt &operator= (const axprt &);

protected:
  xhinfo *xhip;
  ptr<axprt> x; // contained axprt
  axprt (bool r, bool c, size_t ss = 0)
    : xhip (NULL), x (NULL), reliable (r), connected (c), socksize (ss) {}

public:
  enum { defps = 0x10400 };

  const bool reliable;
  const bool connected;
  const size_t socksize;

  typedef callback<void, const char *, ssize_t,
    const sockaddr *>::ptr recvcb_t;

  virtual void sendv (const iovec *, int, const sockaddr *) = 0;
  virtual void setwcb (cbv cb) { (*cb) (); }
  virtual void setrcb (recvcb_t) = 0;
  virtual bool ateof () { return false; }
  virtual u_int64_t get_raw_bytes_sent () const { return 0; }
  virtual int sndbufsize () const { panic ("unimplemented"); return 0; }
  virtual void poll () = 0;
  virtual int getreadfd () = 0;
  virtual int getwritefd () = 0;

  void send (const void *data, size_t len, const sockaddr *dest) {
    iovec iov = {(char *) data, len};
    sendv (&iov, 1, dest);
  }
};

class axprt_dgram : public axprt {
  const size_t pktsize;
  int fd;

  recvcb_t cb;
  sockaddr *sabuf;
  char *pktbuf;

  static bool isconnected (int fd);
  void input ();

protected:
  axprt_dgram (int, bool, size_t, size_t);
  virtual ~axprt_dgram ();

public:
  int getreadfd () { return fd; }
  int getwritefd () { return fd; }
  void sendv (const iovec *, int, const sockaddr *);
  void setrcb (recvcb_t);
  void poll ();

  static ref<axprt_dgram> alloc (int f, size_t ss = sizeof (sockaddr),
				 size_t ps = defps)
    { return New refcounted<axprt_dgram> (f, isconnected (f), ss, ps); }
};


class axprt_pipe : public axprt {
  bool destroyed;
  bool ingetpkt;
  vec<u_int64_t> syncpts;
  int sndbufsz;

protected:
  const size_t pktsize;
  const size_t bufsize;

  int fdread;
  int fdwrite;

  recvcb_t cb;
  u_int32_t pktlen;
  char *pktbuf;

  struct suio *out;
  bool wcbset;
  
  u_int64_t raw_bytes_sent;

  void wrsync ();
  void sendbreak (cbv::ptr);
  bool checklen (int32_t *len);
  virtual ssize_t doread (void *buf, size_t maxlen);
  virtual int dowritev (int iovcnt) { return out->output (fdwrite, iovcnt); }
  virtual void recvbreak ();
  virtual bool getpkt (char **, char *);

  void _sockcheck(int fd);
  void fail ();
  void input ();
  void callgetpkt ();
  void output ();
  
  axprt_pipe (int rfd, int wfd, size_t ps, size_t bufsize = 0);
  virtual ~axprt_pipe ();

public:
  int outlen() { return (out->resid()); }
  void ungetpkt (const void *pkt, size_t len);
  void reclaim (int *rfd, int *wfd);
  int getreadfd () { return fdread; }
  int getwritefd () { return fdwrite; }
  void sockcheck ();
  void poll ();

  bool ateof () { return fdread < 0; }
  virtual void sendv (const iovec *, int, const sockaddr * = NULL);
  void setrcb (recvcb_t);
  void setwcb (cbv);

  u_int64_t get_raw_bytes_sent () const { return raw_bytes_sent; }
  int sndbufsize () const { return sndbufsz; }

  static ref<axprt_pipe> alloc (int rfd, int wfd, size_t ps = defps)
  { return New refcounted<axprt_pipe> (rfd, wfd, ps); }

  unsigned long bytes_sent;
  unsigned long bytes_recv;
};


class axprt_stream : public axprt_pipe {

protected:
  axprt_stream (int fd, size_t ps, size_t bufsize = 0);

public:
  int getfd () { return fdread; }
  int reclaim ();

  static ref<axprt_stream> alloc (int f, size_t ps = defps)
    { return New refcounted<axprt_stream> (f, ps); }
};


/* Clonesrv reads only one packet at a time from the kernel.  Its file
 * descriptor can therefore be passed off to another process without
 * fear of loosing buffered data in the sender. */
class axprt_clone : public axprt_stream {
  friend class axprt_unix;
  int takefd ();
protected:
  axprt_clone (int f, size_t ps) : axprt_stream (f, ps) {}
  virtual ssize_t doread (void *buf, size_t maxsz);
public:
  void extract (int *fdp, str *datap);
  static ref<axprt_clone> alloc (int f, size_t ps = defps)
    { return New refcounted<axprt_clone> (f, ps); }
};

class axprt_unix : public axprt_stream {
  struct fdtosend {
    const int fd;
    mutable bool closeit;
    fdtosend (int f, bool c) : fd (f), closeit (c) {}
    ~fdtosend () { if (closeit) close (fd); }
    fdtosend (const fdtosend &f) : fd (f.fd), closeit (f.closeit)
      { f.closeit = false; }
  };

  vec<fdtosend> fdsendq;
  vec<int> fdrecvq;
  //void sendit (int, bool);

protected:
  axprt_unix (int f, size_t ps, size_t bs = 0)
    : axprt_stream (f, ps, bs), allow_recvfd (true) {}
  virtual ~axprt_unix ();
  virtual void recvbreak ();
  virtual ssize_t doread (void *buf, size_t maxsz);
  virtual int dowritev (int iovcnt);

public:
  bool allow_recvfd;
  static ref<axprt_unix> alloc (int, size_t = axprt_stream::defps);

  void sendfd (int fd, bool closeit = true);
  void sendfd (ref<axprt_unix> x) { sendfd (x->fdwrite, false); }
  void clone (ref<axprt_clone> x);
  int recvfd ();
};

extern pid_t axprt_unix_spawn_pid;
#ifdef MAINTAINER
extern bool axprt_unix_spawn_connected;
#else /* !MAINTAINER */
enum { axprt_unix_spawn_connected = 0 };
#endif /* !MAINTAINER */
ptr<axprt_unix> axprt_unix_spawnv (str path, const vec<str> &av,
				   size_t = 0, cbv::ptr postforkcb = NULL,
				   char *const *env = NULL);
ptr<axprt_unix> axprt_unix_aspawnv (str path, const vec<str> &av,
				    size_t = 0, cbv::ptr postforkcb = NULL,
				    char *const *env = NULL);
ptr<axprt_unix> axprt_unix_spawn (str path, size_t = 0,
				  char *arg0 = NULL, ...);
ptr<axprt_unix> axprt_unix_connect (const char *path,
				    size_t ps = axprt_stream::defps);
ptr<axprt_unix> axprt_unix_stdin (size_t ps = axprt_stream::defps);

typedef callback<ptr<axprt_stream>, int>::ref cloneserv_cb;
bool cloneserv (int fd, cloneserv_cb cb, size_t ps = axprt_stream::defps);

#if 0
template<>
struct hashfn<const ref<axprt> > {
  hashfn () {}
  hash_t operator () (axprt *p) const { return (u_int) p; }
};
#endif
