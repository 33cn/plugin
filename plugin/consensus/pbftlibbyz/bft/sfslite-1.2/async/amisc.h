// -*-c++-*-
/* $Id: amisc.h 3773 2008-11-13 20:50:37Z max $ */

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

#ifndef _ASYNC_AMISC_H_
#define _ASYNC_AMISC_H_ 1

#include "sysconf.h"
#include "err.h"
#include "callback.h"
#include "serial.h"

/* getopt declarations */
extern char *optarg;
extern int optind;

/* Common callback types */
typedef callback<void>::ref cbv;
typedef callback<void, int>::ref cbi;
typedef callback<void, str>::ref cbs;
typedef callback<void, bool>::ref cbb;
extern cbs cbs_null;
extern cbb cbb_null;
extern cbv cbv_null;
extern cbi cbi_null;

/* arandom.c */
extern "C" {
  extern u_int32_t (*arandom_fn) ();
  u_int32_t arandom ();
}

/* straux.C */
char *mempbrk (char *, const char *, int);
char *xstrsep (char **, const char *);
char *strnnsep (char **, const char *);

/* socket.C */
extern in_addr inet_bindaddr;
int inetsocket (int, u_int16_t = 0, u_int32_t = INADDR_ANY);
int inetsocket_resvport (int, u_int32_t = INADDR_ANY);
int unixsocket (const char *);
int unixsocket_connect (const char *);
bool isunixsocket (int);
void close_on_exec (int, bool = true);
int _make_async (int);
void make_async (int);
void make_sync (int);
void tcp_nodelay (int);
void tcp_abort (int);
bool addreq (const sockaddr *, const sockaddr *, socklen_t);

/* sigio.C */
int sigio_set (int fd);
int sigio_clear (int fd);

/* passfd.c */
#include "rwfd.h"

/* myname.C */
str myname ();
str mydomain ();

/* myaddrs.C */
bool myipaddrs (vec<in_addr> *res);

/* fdwait.C */
enum selop { selread = 0, selwrite = 1 };
int fdwait (int fd, bool r, bool w, timeval *tvp);
int fdwait (int rfd, int wfd, bool r, bool w, timeval *tvp);
int fdwait (int fd, selop op, timeval *tvp = NULL);

/* spawn.C */
extern str execdir;
#ifdef MAINTAINER
extern str builddir;		// For running programs in place
extern str buildtmpdir;		// For creating files (e.g. prog.pid)
#endif /* MAINTAINER */
pid_t afork ();
str fix_exec_path (str path, str dir = NULL);
str find_program (const char *program);
str find_program_plus_libsfs (const char *program);
pid_t spawn (const char *, char *const *,
	     int in = 0, int out = 1, int err = 2,
	     cbv::ptr postforkcb = NULL, char *const *env = NULL);
inline pid_t
spawn (const char *path, const char *const *av,
       int in = 0, int out = 1, int err = 2, cbv::ptr postforkcb = NULL,
       char *const *env = NULL)
{
  return spawn (path, const_cast<char *const *> (av),
		in, out, err, postforkcb, env);
}
pid_t aspawn (const char *, char *const *,
	      int in = 0, int out = 1, int err = 2,
	      cbv::ptr postforkcb = NULL, char *const *env = NULL);
inline pid_t
aspawn (const char *path, const char *const *av,
	int in = 0, int out = 1, int err = 2, cbv::ptr postforkcb = NULL,
	char *const *env = NULL)
{
  return aspawn (path, const_cast<char *const *> (av),
		 in, out, err, postforkcb, env);
}

/* lockfile.C */
bool stat_unchanged (const struct stat *sb1, const struct stat *sb2);
class lockfile {
protected:
  bool islocked;
  int fd;

  ~lockfile ();
  bool openit ();
  void closeit ();
  bool fdok ();

public:
  const str path;

  lockfile (const str &p) : islocked (false), fd (-1), path (p) {}
  bool locked () const { return islocked; }
  int getfd () const { return fd; }
  bool acquire (bool wait = false);
  void release ();
  bool ok () { return islocked && fdok (); }
  static ptr<lockfile> alloc (const str &path, bool wait = false);
};

/* daemonize.C */
extern str syslog_priority;
void daemonize (const str &name = NULL);
void start_logger ();
int start_logger (const str &pri, const str &tag, const str &line, 
		  const str &logfile, int flags, mode_t mode);

/* Random usefull operators */
#include "keyfunc.h"

inline bool
operator== (const in_addr &a, const in_addr &b)
{
  return a.s_addr == b.s_addr;
}
inline bool
operator!= (const in_addr &a, const in_addr &b)
{
  return a.s_addr != b.s_addr;
}
template<> struct hashfn<in_addr> {
  hashfn () {}
  hash_t operator() (const in_addr a) const
    { return a.s_addr; }
};

inline bool
operator== (const sockaddr_in &a, const sockaddr_in &b)
{
  return a.sin_port == b.sin_port && a.sin_addr.s_addr == b.sin_addr.s_addr;
}
inline bool
operator!= (const sockaddr_in &a, const sockaddr_in &b)
{
  return a.sin_port != b.sin_port || a.sin_addr.s_addr != b.sin_addr.s_addr;
}
template<> struct hashfn<sockaddr_in> {
  hashfn () {}
  hash_t operator() (const sockaddr_in &a) const
    { return ntohs (a.sin_port) << 16 ^ htonl (a.sin_addr.s_addr); }
};

inline bool
operator== (const timespec &a, const timespec &b)
{
  return a.tv_nsec == b.tv_nsec && a.tv_sec == b.tv_sec;
}
inline bool
operator!= (const timespec &a, const timespec &b)
{
  return a.tv_nsec != b.tv_nsec || a.tv_sec != b.tv_sec;
}

inline int
tscmp (const timespec &a, const timespec &b)
{
  if (a.tv_sec < b.tv_sec)
    return -1;
  if (b.tv_sec < a.tv_sec)
    return 1;
  if (a.tv_nsec < b.tv_nsec)
    return -1;
  return b.tv_nsec < a.tv_nsec;
};
inline bool
operator< (const timespec &a, const timespec &b)
{
  return tscmp (a, b) < 0;
}
inline bool
operator<= (const timespec &a, const timespec &b)
{
  return tscmp (a, b) <= 0;
}
inline bool
operator> (const timespec &a, const timespec &b)
{
  return tscmp (a, b) > 0;
}
inline bool
operator>= (const timespec &a, const timespec &b)
{
  return tscmp (a, b) >= 0;
}

inline timespec
operator+ (const timespec &a, const timespec &b)
{
  timespec ts;
  ts.tv_sec = a.tv_sec + b.tv_sec;
  if ((ts.tv_nsec = a.tv_nsec + b.tv_nsec) > 1000000000) {
    ts.tv_nsec -= 1000000000;
    ++ts.tv_sec;
  }
  return ts;
}

inline timespec
operator- (const timespec &a, const timespec &b)
{
  timespec ts;
  ts.tv_sec = a.tv_sec - b.tv_sec;
  if (a.tv_nsec > b.tv_nsec)
    ts.tv_nsec = a.tv_nsec - b.tv_nsec;
  else {
    ts.tv_nsec = a.tv_nsec + 1000000000 - b.tv_nsec;
    --ts.tv_sec;
  }
  return ts;
}

template<class T> void rc_ignore(const T &in) {}

inline void v_write (int fd, const void *buf, size_t sz) 
{ rc_ignore (write (fd, buf, sz)); }


#endif /* !_ASYNC_AMISC_H_ */
