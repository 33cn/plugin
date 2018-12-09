// -*-c++-*-
/* $Id: err.h 3161 2008-01-14 16:55:36Z max $ */

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

#ifndef _ASYNC_ERR_H_
#define _ASYNC_ERR_H_ 1

#include "str.h"

extern bssstr progname;
extern str progdir;
extern void (*fatalhook) ();

extern int errfd;
extern bool fatal_no_destruct;
extern void (*_err_output) (suio *, int);
extern void (*_err_reset_hook) ();
void err_reset ();
void _err_output_sync (suio *, int);

void setprogname (char *argv0);
void setprogpid (int p);

/* Old-style C functions for compatibility */
extern "C" {
  void sfs_warn (const char *fmt, ...) __attribute__ ((format (printf, 1, 2)));
  void sfs_warnx (const char *fmt, ...)
    __attribute__ ((format (printf, 1, 2)));
  void sfs_vwarn (const char *fmt, va_list ap);
  void sfs_vwarnx (const char *fmt, va_list ap);
  void fatal (const char *fmt, ...)
    __attribute__ ((noreturn, format (printf, 1, 2)));
  void panic (const char *fmt, ...)
    __attribute__ ((noreturn, format (printf, 1, 2)));
}

class warnobj : public strbuf {
  const int flags;
public:
  enum { xflag = 1, fatalflag = 2, panicflag = 4, timeflag = 8 };

  explicit warnobj (int);
  ~warnobj ();
  const warnobj &operator() (const char *fmt, ...) const
    __attribute__ ((format (printf, 2, 3)));
};
#define warn warnobj (0)
#define vwarn warn.vfmt
#define warnx warnobj (int (::warnobj::xflag))
#define warnt warnobj (int (::warnobj::timeflag))
#define vwarnx warnx.vfmt

#ifndef __attribute__
/* Fatalobj is just a warnobj with a noreturn destructor. */
class fatalobj : public warnobj {
public:
  explicit fatalobj (int f) : warnobj (f) {}
  ~fatalobj () __attribute__ ((noreturn));
};
#else /* __attribute__ */
# define fatalobj warnobj
#endif /* __attribute__ */
#define fatal fatalobj (int (::warnobj::fatalflag))
#define panic fatalobj (int (::warnobj::panicflag)) ("%s\n", __BACKTRACE__)

struct traceobj : public strbuf {
  int current_level;
  const char *prefix;
  const bool dotime;
  bool doprint;

  traceobj (int current_level, const char *prefix = "", bool dotime = false)
    : current_level (current_level), prefix (prefix), dotime (dotime) {}
  ~traceobj ();
  void init ();

  const traceobj &operator() (int threshold = 0);
  const traceobj &operator() (int threshold, const char *fmt, ...)
    __attribute__ ((format (printf, 3, 4)));
};

template<class T> inline const traceobj &
operator<< (const traceobj &sb, const T &a)
{
  if (sb.doprint)
    strbuf_cat (sb, a);
  return sb;
}
inline const traceobj &
operator<< (const traceobj &sb, const str &s)
{
  if (sb.doprint)
    suio_print (sb.tosuio (), s);
  return sb;
}

template<class T> inline const warnobj &
operator<< (const warnobj &sb, const T &a)
{
  strbuf_cat (sb, a);
  return sb;
}

inline const warnobj &
operator<< (const warnobj &sb, const str &s)
{
  if (s)
    suio_print (sb.tosuio (), s);
  else {
    sb.cat ("(null)", true);
  }
  return sb;
}


#undef assert
#define assert(e)						\
  do {								\
    if (!(e))							\
      panic ("assertion \"%s\" failed at %s\n", #e, __FL__);	\
  } while (0)

#ifdef DMALLOC
inline void myabort () __attribute__ ((noreturn));
inline void
myabort ()
{
  static bool shutdown_called;
  if (!shutdown_called) {
    shutdown_called = true;
    dmalloc_shutdown ();
  }
  abort ();
}
#else /* !DMALLOC */
# define myabort abort
#endif /* !DMALLOC */

#endif /* !_ASYNC_ERR_H_ */
