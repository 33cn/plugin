/* $Id: err.C 3758 2008-11-13 00:36:00Z max $ */

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

#include "err.h"

#undef warn
#undef warnx
#undef vwarn
#undef vwarnx
#undef fatal
#undef panic

bssstr progname;
bssstr progpid;
str progdir;
void (*fatalhook) ();

int errfd = 2;
bool fatal_no_destruct;
void (*_err_output) (suio *uio, int flags) = _err_output_sync;
void (*_err_reset_hook) ();

#if DMALLOC_VERSION_MAJOR < 5
#define dmalloc_logpath _dmalloc_logpath
extern "C" char *dmalloc_logpath;
#endif /* DMALLOC_VERSION_MAJOR < 5 */

void
setprogname (char *argv0)
{
  char *cp;
  if ((cp = strrchr (argv0, '/')))
    cp++;
  else
    cp = argv0;
  /* Libtool shell wrappers leave lt- in argv[0] */
  if (cp[0] == 'l' && cp[1] == 't' && cp[2] == '-')
    progname = cp + 3;
  else
    progname = cp;
  if (cp > argv0)
    progdir.setbuf (argv0, cp - argv0);
  else
    progdir = NULL;
#ifdef DMALLOC
  if (dmalloc_logpath) {
    str logname;
    const char *p;
    if (!(p = strrchr (dmalloc_logpath, '/')) || !(p = strrchr (p, '.')))
      p = dmalloc_logpath + strlen (dmalloc_logpath);
    logname = strbuf ("%.*s.%s", int (p - dmalloc_logpath), dmalloc_logpath,
		      progname.cstr ());
    static char *lp;
    if (lp)
      xfree (lp);
    lp = xstrdup (logname);
    dmalloc_logpath = lp;
  }
#endif /* DMALLOC */
}

void
setprogpid (int p)
{
  strbuf b ("%d", p);
  progpid = b;
}

void
err_reset ()
{
  if (_err_reset_hook)
    (*_err_reset_hook) ();
  _err_reset_hook = NULL;
  _err_output = _err_output_sync;
}

void
_err_output_sync (suio *uio, int flags)
{
  int saved_errno = errno;
  uio->output (errfd);
  if (flags & warnobj::panicflag)
    myabort ();
  if (flags & warnobj::fatalflag) {
    if (fatalhook)
      (*fatalhook) ();
    if (fatal_no_destruct)
      _exit (1);
    exit (1);
  }
  errno = saved_errno;
}

static const char *
timestring ()
{
  timespec ts;
  clock_gettime (CLOCK_REALTIME, &ts);
  static str buf;
  buf = strbuf ("%d.%06d", int (ts.tv_sec), int (ts.tv_nsec/1000));
  return buf;
}

/*
 *  warnobj
 */

warnobj::warnobj (int f)
  : flags (f)
{
  if (flags & timeflag)
    cat (timestring ()).cat (" ");
  if (!(flags & xflag) && progname) {
    if (progpid) {
      cat (progname).cat ("[").cat (progpid).cat ("]: ");
    } else {
      cat (progname).cat (": ");
	}
  }
  if (flags & panicflag)
    cat ("PANIC: ");
  else if (flags & fatalflag)
    cat ("fatal: ");
}

const warnobj &
warnobj::operator() (const char *fmt, ...) const
{
  va_list ap;
  va_start (ap, fmt);
  vfmt (fmt, ap);
  va_end (ap);
  return *this;
}

warnobj::~warnobj ()
{
  _err_output (uio, flags);
}

#ifndef fatalobj
fatalobj::~fatalobj ()
{
  /* Of course, gcc won't let this function return, so we have to jump
   * through a few hoops rather than simply implement ~fatalobj as
   * {}. */
  static_cast<warnobj *> (this)->~warnobj ();
  myabort ();
}
#endif /* !fatalobj */

void
traceobj::init ()
{
  if (progname)
    cat (progname).cat (": ");
  cat (prefix);
  if (dotime)
    cat (timestring ()).cat (" ");
}

traceobj::~traceobj ()
{
  if (doprint)
    _err_output (uio, 0);
}

const traceobj &
traceobj::operator() (const int threshold)
{
  doprint = current_level >= threshold;
  if (doprint)
    init ();
  return *this;
}

const traceobj &
traceobj::operator() (const int threshold, const char *fmt, ...)
{
  doprint = current_level >= threshold;
  if (doprint) {
    init ();
    va_list ap;
    va_start (ap, fmt);
    vfmt (fmt, ap);
    va_end (ap);
  }
  return *this;
}

/*
 *  Traditional functions
 */

void
sfs_vwarn (const char *fmt, va_list ap)
{
  suio uio;
  if (progname)
    uio.print (progname.cstr (), progname.len ());
  suio_vuprintf (&uio, fmt, ap);
  _err_output (&uio, 0);
}

void
sfs_vwarnx (const char *fmt, va_list ap)
{
  suio uio;
  suio_vuprintf (&uio, fmt, ap);
  _err_output (&uio, warnobj::xflag);
}

void
sfs_warn (const char *fmt, ...)
{
  va_list ap;
  va_start (ap, fmt);
  sfs_vwarn (fmt, ap);
  va_end (ap);
}

void
sfs_warnx (const char *fmt, ...)
{
  va_list ap;
  va_start (ap, fmt);
  sfs_vwarnx (fmt, ap);
  va_end (ap);
}

void
fatal (const char *fmt, ...)
{
  va_list ap;
  strbuf b;
  if (progname)
    b << progname << ": ";
  b << "fatal: ";

  va_start (ap, fmt);
  b.vfmt (fmt, ap);
  va_end (ap);

  _err_output (b.tosuio (), warnobj::fatalflag);
  exit (1);
}

void
panic (const char *fmt, ...)
{
  va_list ap;
  strbuf b;
  if (progname)
    b << progname << ": ";
  b << "PANIC: " << __BACKTRACE__ << "\n";

  va_start (ap, fmt);
  b.vfmt (fmt, ap);
  va_end (ap);

  _err_output (b.tosuio (), warnobj::panicflag);
  myabort ();
}

GLOBALDESTRUCT;
