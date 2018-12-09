// -*-c++-*-
/* $Id: aiod.h 2881 2007-05-18 19:06:04Z max $ */

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

#ifndef _ASYNC_AIOD_H_
#define _ASYNC_AIOD_H_ 1

#include "async.h"
#include "bbuddy.h"
#include "ihash.h"
#include "aiod_prot.h"

struct aiod_req;

// gcc 4.1 fixes
class aiod;
class aiofh;

class aiobuf {
  friend class aiod;
  friend class aiofh;

  char *const buf;
  const size_t len;
  aiod *const iod;
  const size_t pos;

  aiobuf (const aiobuf &);
  aiobuf &operator= (const aiobuf &);
  
protected:
  aiobuf (aiod *d, size_t pos, size_t len);
  ~aiobuf ();

public:
  inline char *base ();
  inline const char *base () const;
  size_t size () const { return len; }
  inline char *lim ();
  inline const char *lim () const;
};

class aiod {
  friend class aiobuf;
  friend class aiofh;

  typedef callback<void, ptr<aiobuf> >::ref cbb;
  typedef callback<void, str, int>::ref cbsi;
  typedef callback<void, struct stat *, int>::ref cbstat;
  typedef callback<void, ptr<aiofh>, int>::ref cbopen;

  enum { maxwrite = (PIPE_BUF / sizeof (off_t)) * sizeof (off_t) };

  class writeq {
    static void checkmaxwrite () { switch (0) case maxwrite: case 0:; }
    suio wbuf;
    void output ();
  public:
    int wfd;
    writeq () : wfd (-1) {}
    ~writeq () { close (); }
    void close () {
      if (wfd >= 0) {
	fdcb (wfd, selread, NULL);
	fdcb (wfd, selwrite, NULL);
	::close (wfd);
	wfd = -1;
      }
    }
    void sendmsg (aiomsg_t msg);
  };

  struct daemon {
    pid_t pid;
    int fd;
    writeq wq;

    daemon () : pid (-1), fd (-1) {}
    bool launch (str path, int shmfd, int commonfd);
  };

  struct request {
    ref<aiobuf> buf;
    const size_t pos;
    vec<cbb, 1> cbvec;
    ihash_entry<request> hlink;
    request (ref<aiobuf> b);
  };
  friend class request;
  
  bool closed;
  bool finalized;
  bool growlock;
  bool bufwakereq;
  bool bufwakelock;
  bool shmpin;
  size_t refcnt;

  int shmfd;
  size_t shmmax;
  size_t shmlen;
  char *shmbuf;
  bbuddy bb;
  vec<cbv> bbwaitq;

  writeq wq;

  const size_t ndaemons;
  daemon *dv;

  int fhno_ctr;
  vec<int> fhno_avail;

  ihash<const size_t, request, &request::pos, &request::hlink> rqtab;

  void delreq (request *r);
  void fail ();
  void input (int);
  void bufwake ();
  void bufalloc_cb1 (size_t inc, ptr<aiobuf> buf);
  void bufalloc_cb2 (size_t inc, ptr<aiobuf> buf);
  void sendmsg (ref<aiobuf> buf, cbb cb, int dst = -1);

  int fhno_alloc () { return fhno_avail.empty () ? fhno_ctr++
			: fhno_avail.pop_back (); }
  void fhno_free (int n) { fhno_avail.push_back (n); }

  static size_t buf2pos (aiobuf *buf) { return buf->pos; }
  static aiod_reqhdr *buf2hdr (aiobuf *buf)
    { return reinterpret_cast<aiod_reqhdr *> (buf->base ()); }
  static aiod_pathop *buf2pathop (aiobuf *buf)
    { return reinterpret_cast<aiod_pathop *> (buf->base ()); }
  static aiod_fhop *buf2fhop (aiobuf *buf)
    { return reinterpret_cast<aiod_fhop *> (buf->base ()); }
  static aiod_fstat *buf2fstat (aiobuf *buf)
    { return reinterpret_cast<aiod_fstat *> (buf->base ()); }
  static aiod_file *buf2file (aiobuf *buf)
    { return reinterpret_cast<aiod_file *> (buf->base ()); }
  static aiod_nop *buf2nop (aiobuf *buf)
    { return reinterpret_cast<aiod_nop *> (buf->base ()); }

  static void cbi_cb (cbi cb, ptr<aiobuf> buf);
  static void cbstat_cb (cbstat cb, ptr<aiobuf> buf);
  static void pathret_cb (cbsi cb, ptr<aiobuf> buf);
  static void open_cb (ref<aiofh> fh, cbopen cb, ptr<aiobuf> buf);
  
  void pathop (aiod_op op, str p1, str p2, cbb cb, size_t bufsize = 0);
  void statop (aiod_op op, const str &path, const cbstat &cb)
    { pathop (op, path, NULL, wrap (cbstat_cb, cb), sizeof (struct stat)); }

  ~aiod ();
  void addref () { refcnt++; }
  void delref () { if (!--refcnt && finalized) delete this; }

public:
  enum { minbuf = 0x40 };
  const size_t maxbuf;

  aiod (u_int nproc = 1, ssize_t shmsize = 0x200000,
	size_t maxbuf = 0x10000, bool shmpin = false,
	str path = NULL, str tmpdir = NULL);
  void finalize () { finalized = true; addref (); delref (); }

  ptr<aiobuf> bufalloc (size_t len);
  void bufwait (cbv cb) { bbwaitq.push_back (cb); }

  void unlink (str path, cbi cb)
    { pathop (AIOD_UNLINK, path, NULL, wrap (cbi_cb, cb)); }
  void link (str from, str to, cbi cb)
    { pathop (AIOD_LINK, from, to, wrap (cbi_cb, cb)); }
  void symlink (str from, str to, cbi cb)
    { pathop (AIOD_SYMLINK, from, to, wrap (cbi_cb, cb)); }
  void rename (str from, str to, cbi cb)
    { pathop (AIOD_RENAME, from, to, wrap (cbi_cb, cb)); }

  void readlink (str path, cbsi cb)
    { pathop (AIOD_READLINK, path, NULL, wrap (pathret_cb, cb), PATH_MAX); }
  void getcwd (str path, cbsi cb)
    { pathop (AIOD_GETCWD, path, NULL, wrap (pathret_cb, cb), PATH_MAX); }

  void stat (str path, cbstat cb) { statop (AIOD_STAT, path, cb); }
  void lstat (str path, cbstat cb) { statop (AIOD_LSTAT, path, cb); }

  void open (str path, int flags, int mode, cbopen cb);
  void opendir (str path, cbopen cb);
};

inline char *
aiobuf::base ()
{
  return buf;
}

inline const char *
aiobuf::base () const
{
  return buf;
}

inline char *
aiobuf::lim ()
{
  return base () + len;
}

inline const char *
aiobuf::lim () const
{
  return base () + len;
}

inline
aiod::request::request (ref<aiobuf> b)
  : buf (b), pos (buf2pos (buf))
{
}

class aiofh : public virtual refcount {
  friend class aiod;

  typedef callback<void, ptr<aiobuf>, ssize_t, int>::ref cbrw;

  aiod *const iod;
  const ref<aiobuf> fh;
  const int fhno;
  bool isdir;
  bool closed;

  void rw (aiod_op op, off_t pos, ptr<aiobuf> iobuf,
	   u_int iostart, u_int iosize, cbrw cb);
  void simpleop (aiod_op op, aiod::cbb cb, off_t length);
  void sendclose (cbi::ptr cb = NULL);

  void rw_cb (ref<aiobuf> iobuf, cbrw cb, ptr<aiobuf> rqbuf);
  void cbstat_cb (aiod::cbstat cb, ptr<aiobuf> buf);
  static void close_cb (int *ctr, cbi::ptr cb, ptr<aiobuf> buf);
  void cbi_cb (cbi cb, ptr<aiobuf> buf);

protected:
  aiofh (aiod *iod, ref<aiobuf> fh, bool dir = false);
  ~aiofh ();

public:
  void close (cbi cb);
  void closedir (cbi cb);
  void fsync (cbi cb)
    { simpleop (AIOD_FSYNC, wrap (mkref (this), &aiofh::cbi_cb, cb), 
		off_t (0)); }
  void ftrunc (off_t length, cbi cb) {
    simpleop (AIOD_FTRUNC, wrap (mkref (this), &aiofh::cbi_cb, cb), length);
  }
  void read (off_t pos, ptr<aiobuf> buf, cbrw cb)
    { rw (AIOD_READ, pos, buf, 0, buf->size (), cb); }
  void readdir (ptr<aiobuf> buf, cbrw cb)
    { rw (AIOD_READDIR, 0, buf, 0, buf->size (), cb); }
  void sread (off_t pos, ptr<aiobuf> buf, u_int iostart, u_int iosize, cbrw cb)
    { rw (AIOD_READ, pos, buf, iostart, iosize, cb); }
  void sreaddir (off_t pos, ptr<aiobuf> buf, u_int iostart, u_int iosize, cbrw cb)
    { rw (AIOD_READDIR, pos, buf, iostart, iosize, cb); }
  void write (off_t pos, ptr<aiobuf> buf, cbrw cb)
    { rw (AIOD_WRITE, pos, buf, 0, buf->size (), cb); }
  void swrite (off_t pos, ptr<aiobuf> buf, u_int iostart, u_int iosize,
	       cbrw cb) { rw (AIOD_WRITE, pos, buf, iostart, iosize, cb); }
  void fstat (aiod::cbstat cb)
    { simpleop (AIOD_FSTAT, wrap (mkref (this), &aiofh::cbstat_cb, cb), 
		off_t (0)); }
};


#endif /* !_ASYNC_AIOD_H_ */
