/* $Id: core.C 3810 2008-11-21 16:04:13Z max $ */

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

#include "async.h"
#include "fdlim.h"
#include "ihash.h"
#include "itree.h"
#include "list.h"
#include "corebench.h"

#include <typeinfo>

#include "litetime.h"
#include "sfs_select.h"
#include <stdio.h>

bool amain_panic;

/* Global variables used for configuring the core select behavior */

static timeval selwait;

namespace sfs_core {
  static bool g_busywait;
  static bool g_zombie_collect;
  static selector_t *selector;

  void set_busywait (bool b) { g_busywait = b; }
  void set_zombie_collect (bool b) { g_zombie_collect = b; }
};


/*
 * returns: 0 if no change, -1 if changed to an unavailable policy,
 * and 1 if the changes was successful.
 */
int
sfs_core::set_select_policy (select_policy_t p)
{

  int ret = 1;
  if (p == selector->typ ()) {
    ret = 0;
  } else {
    selector_t *ns = NULL;
    switch (p) {
    case SELECT_EPOLL:
#ifdef HAVE_EPOLL
      ns = New epoll_selector_t (selector);
#endif
      break;
    case SELECT_KQUEUE:
#ifdef HAVE_KQUEUE
      ns = New kqueue_selector_t (selector);
#endif
      break;
    case SELECT_STD:
      ns = New std_selector_t (selector);
      break;
    default:
      break;
    }
    if (ns) {
      delete selector;
      selector = ns;
      ret = 1;
    } else {
      ret = -1;
    }
  }
  return ret;
}

#ifdef WRAP_DEBUG
#define CBTR_FD    0x0001
#define CBTR_TIME  0x0002
#define CBTR_SIG   0x0004
#define CBTR_CHLD  0x0008
#define CBTR_LAZY  0x0010
static int callback_trace;
static bool callback_time;

static inline const char *
timestring ()
{

  if (!callback_time)
    return "";

  struct timespec ts;

  sfs_get_tsnow (&ts, true);

  static str buf;
  buf = strbuf ("%d.%06d ", int (ts.tv_sec), int (ts.tv_nsec/1000));
  return buf;
}

#endif /* WRAP_DEBUG */



struct child {
  pid_t pid;
  cbi cb;
  ihash_entry<child> link;
  child (pid_t p, cbi c) : pid (p), cb (c) {}
};
static ihash<pid_t, child, &child::pid, &child::link> chldcbs;

struct zombie_t {
  pid_t _pid;
  int _status;
  ihash_entry<zombie_t> _link;
  zombie_t (pid_t p, int s) : _pid (p), _status (s) {}
};
static ihash<pid_t, zombie_t, &zombie_t::_pid, &zombie_t::_link> zombies;

struct timecb_t {
  timespec ts;
  const cbv cb;
  itree_entry<timecb_t> link;
  timecb_t (const timespec &t, const cbv &c) : ts (t), cb (c) {}
};
static itree<timespec, timecb_t, &timecb_t::ts, &timecb_t::link> timecbs;
static bool timecbs_altered;

struct lazycb_t {
  const time_t interval;
  time_t next;
  const cbv cb;
  list_entry<lazycb_t> link;

  lazycb_t (time_t interval, cbv cb);
  ~lazycb_t ();
};
static bool lazycb_removed;
static list<lazycb_t, &lazycb_t::link> *lazylist;

#ifdef HAVE_EPOLL

#else /* !HAVE_EPOLL */


#endif /* HAVE_EPOLL */

static int sigpipes[2] = { -1, -1 };
#ifdef NSIG
const int nsig = NSIG;
#else /* !NSIG */
const int nsig = 32;
#endif /* !NSIG */
/* Note: sigdocheck and sigcaught intentionally ints rather than
 * bools.  The hope is that an int can safely be written without
 * affecting surrounding memory.  (This is certainly not the case on
 * some architectures if bool is a char.  Consider getting signal 2
 * right after signal 3 on an alpha, for instance.  You might end up
 * clearing sigcaught[2] when you finish setting sigcaught[3].) */
static volatile int sigdocheck;
static volatile int sigcaught[nsig];
static bssptr<cbv::type> sighandler[nsig];

void sigcb_check ();

void
chldcb (pid_t pid, cbi::ptr cb)
{
  if (child *c = chldcbs[pid]) {
    chldcbs.remove (c);
    delete c;
  }

  zombie_t *z;
  if ((z = zombies[pid])) {
    int s = z->_status;
    zombies.remove (z);
    delete z;
    if (cb) 
      (*cb) (s);
  } else if (cb) {
    chldcbs.insert (New child (pid, cb));
  }
}

void
chldcb_check ()
{
  for (;;) {
    int status;
    pid_t pid = waitpid (-1, &status, WNOHANG);
    if (pid == 0 || pid == -1)
      break;
    if (child *c = chldcbs[pid]) {
      chldcbs.remove (c);
#ifdef WRAP_DEBUG
      if (callback_trace & CBTR_CHLD)
	warn ("CALLBACK_TRACE: %schild pid %d (status %d) %s <- %s\n",
	      timestring (), pid, status, c->cb->dest, c->cb->line);
#endif /* WRAP_DEBUG */
      STOP_ACHECK_TIMER ();
      sfs_leave_sel_loop ();
      (*c->cb) (status);
      START_ACHECK_TIMER ();
      delete c;
    } else if (sfs_core::g_zombie_collect) {
      zombie_t *z = zombies[pid];
      if (z) {
	z->_status = status;
      } else {
	zombies.insert (New zombie_t (pid, status));
      }
    }
  }
}

timecb_t *
timecb (const timespec &ts, cbv cb)
{
  sfs_add_new_cb ();
  timecb_t *to = New timecb_t (ts, cb);
  timecbs.insert (to);
  return to;
}

timecb_t *
delaycb (time_t sec, u_int32_t nsec, cbv cb)
{
  timespec ts;
  if (sec == 0 && nsec == 0) {
    ts.tv_sec = 0;
    ts.tv_nsec = 0;
  } else {
    sfs_get_tsnow (&ts, true);

    ts.tv_sec += sec;
    ts.tv_nsec += nsec;
    if (ts.tv_nsec >= 1000000000) {
      ts.tv_nsec -= 1000000000;
      ts.tv_sec++;
    }

  }
  return timecb (ts, cb);
}

void
timecb_remove (timecb_t *to)
{
  if (!to)
    return;

  for (timecb_t *tp = timecbs[to->ts]; tp != to; tp = timecbs.next (tp))
    if (!tp || tp->ts != to->ts)
      panic ("timecb_remove: invalid timecb_t\n");
  timecbs_altered = true;
  timecbs.remove (to);
  delete to;
}

void
timecb_check ()
{
  struct timespec my_ts;

  timecb_t *tp, *ntp;

  if (timecbs.first ()) {
    sfs_set_global_timestamp ();
    my_ts = sfs_get_tsnow ();

    for (tp = timecbs.first (); tp && tp->ts <= my_ts;
	 tp = timecbs_altered ? timecbs.first () : ntp) {
      ntp = timecbs.next (tp);
      timecbs.remove (tp);
      timecbs_altered = false;
#ifdef WRAP_DEBUG
      if (callback_trace & CBTR_TIME)
	warn ("CALLBACK_TRACE: %stimecb %s <- %s\n", timestring (),
	      tp->cb->dest, tp->cb->line);
#endif /* WRAP_DEBUG */
      STOP_ACHECK_TIMER ();
      sfs_leave_sel_loop ();
      (*tp->cb) ();
      START_ACHECK_TIMER ();
      delete tp;
    }
  }

  selwait.tv_usec = 0;
  selwait.tv_sec = 0;
  if (!sfs_core::g_busywait && !sigdocheck) {
    if (!(tp = timecbs.first ()))
      selwait.tv_sec = 86400;
    else {
      if (tp->ts.tv_sec == 0) {
	selwait.tv_sec = 0;
      } else {
	sfs_set_global_timestamp ();
	my_ts = sfs_get_tsnow ();
	if (tp->ts < my_ts)
	  selwait.tv_sec = 0;
	else if (tp->ts.tv_nsec >= my_ts.tv_nsec) {
	  selwait.tv_sec = tp->ts.tv_sec - my_ts.tv_sec;
	  selwait.tv_usec = (tp->ts.tv_nsec - my_ts.tv_nsec) / 1000;
	}
	else {
	  selwait.tv_sec = tp->ts.tv_sec - my_ts.tv_sec - 1;
	  selwait.tv_usec = (1000000000 + tp->ts.tv_nsec - 
			     my_ts.tv_nsec) / 1000;
	}
      }
    }
  }
}

void fdcb_check () { sfs_core::selector->fdcb_check (&selwait); }

void _fdcb (int fd, selop op, cbv::ptr cb, const char *file, int line) 
{ sfs_core::selector->_fdcb (fd, op, cb, file, line); }

static void
sigcatch (int sig)
{
  sigdocheck = 1;
  sigcaught[sig] = 1;
  selwait.tv_sec = selwait.tv_usec = 0;
  /* On some operating systems, select is not a system call but is
   * implemented inside libc.  This may cause a race condition in
   * which select ends up being called with the original (non-zero)
   * value of selwait.  We avoid the problem by writing to a pipe that
   * will wake up the select. */
  v_write (sigpipes[1], "", 1);
}

cbv::ptr
sigcb (int sig, cbv::ptr cb, int flags)
{
  sigset_t set;

  sfs_add_new_cb ();
  if (!sigemptyset (&set) && !sigaddset (&set, sig))
    sigprocmask (SIG_UNBLOCK, &set, NULL);

  struct sigaction sa;
  assert (sig > 0 && sig < nsig);
  bzero (&sa, sizeof (sa));
  sa.sa_handler = cb ? sigcatch : SIG_DFL;
  sa.sa_flags = flags;
  if (sigaction (sig, &sa, NULL) < 0) // Must be bad signal, serious bug
    panic ("sigcb: sigaction: %m\n");
  cbv::ptr ocb = sighandler[sig];
  sighandler[sig] = cb;
  return ocb;
}

void
sigcb_check ()
{
  if (sigdocheck) {
    char buf[64];
    while (read (sigpipes[0], buf, sizeof (buf)) > 0)
      ;
    sigdocheck = 0;
    for (int i = 1; i < nsig; i++)
      if (sigcaught[i]) {
	sigcaught[i] = 0;
	if (cbv::ptr cb = sighandler[i]) {
#ifdef WRAP_DEBUG
	  if ((callback_trace & CBTR_SIG) && i != SIGCHLD) {
# ifdef NEED_SYS_SIGNAME_DECL
	    warn ("CALLBACK_TRACE: %ssignal %d %s <- %s\n",
		  timestring (), i, cb->dest, cb->line);
# else /* !NEED_SYS_SIGNAME_DECL */
	    warn ("CALLBACK_TRACE: %sSIG%s %s <- %s\n", timestring (),
		  sys_signame[i], cb->dest, cb->line);
# endif /* !NEED_SYS_SIGNAME_DECL */
	  }
#endif /* WRAP_DEBUG */
	  STOP_ACHECK_TIMER ();
	  sfs_leave_sel_loop ();
	  (*cb) ();
	  START_ACHECK_TIMER ();
	}
      }
  }
}

lazycb_t::lazycb_t (time_t i, cbv c)
  : interval (i), next (sfs_get_timenow(true) + interval), cb (c)
{
  lazylist->insert_head (this);
}

lazycb_t::~lazycb_t ()
{
  lazylist->remove (this);
}

lazycb_t *
lazycb (time_t interval, cbv cb)
{
  return New lazycb_t (interval, cb);
}

void
lazycb_remove (lazycb_t *lazy)
{
  lazycb_removed = true;
  delete lazy;
}

void
lazycb_check ()
{
  time_t my_timenow = 0;

 restart:
  lazycb_removed = false;
  for (lazycb_t *lazy = lazylist->first; lazy; lazy = lazylist->next (lazy)) {

    if (my_timenow == 0) {
      sfs_set_global_timestamp ();
      my_timenow = sfs_get_timenow ();
    }

    if (my_timenow < lazy->next)
      continue;
    lazy->next = my_timenow + lazy->interval;
#ifdef WRAP_DEBUG
    if (callback_trace & CBTR_LAZY)
      warn ("CALLBACK_TRACE: %slazy %s <- %s\n", timestring (),
	    lazy->cb->dest, lazy->cb->line);
#endif /* WRAP_DEBUG */
    STOP_ACHECK_TIMER ();
    sfs_leave_sel_loop ();
    (*lazy->cb) ();
    START_ACHECK_TIMER ();
    if (lazycb_removed)
      goto restart;
  }
}

/*
 * MK 11/21/08  -- Every so often, we get in a weird situation, in which
 * sigpipes[0] is readable, but sigdocheck isn't set, so we're in a tight
 * CPU loop.  This hack should workaround that problem.  It used to be
 * that sigpipes[0] was set selread with cbv_null...
 */
static void sigcb_set_checkbit() { sigdocheck = 1; }

static void
ainit ()
{
  if (sigpipes[0] == -1) {
    if (pipe (sigpipes) < 0)
      fatal ("could not create sigpipes: %m\n");

    _make_async (sigpipes[0]);
    _make_async (sigpipes[1]);
    close_on_exec (sigpipes[0]);
    close_on_exec (sigpipes[1]);
    fdcb (sigpipes[0], selread, wrap (sigcb_set_checkbit));

    /* Set SA_RESTART for SIGCHLD, primarily for the benefit of
     * stdio-using code like lex/flex scanners.  These tend to flip out
     * if read ever returns EINTR. */
    sigcb (SIGCHLD, wrap (chldcb_check), (SA_NOCLDSTOP
#ifdef SA_RESTART
					  | SA_RESTART
#endif /* SA_RESTART */
					  ));
    sigcatch (SIGCHLD);
  }
}

unsigned long long time_in_acheck, tia_tmp, n_wrap_calls;
bool do_corebench = false;

static inline void
_acheck ()
{
  
  sfs_leave_sel_loop ();

  START_ACHECK_TIMER();
  // warn << "in acheck...\n";
  if (amain_panic)
    panic ("child process returned from afork ()\n");
  lazycb_check ();
  fdcb_check ();
  sigcb_check ();

  timecb_check ();
  STOP_ACHECK_TIMER ();
}

void
acheck ()
{
  timecb_check ();
  ainit ();
  _acheck ();
}

void
amain ()
{
  static bool amain_called;
  if (amain_called)
    panic ("amain called recursively\n");
  amain_called = true;
  START_ACHECK_TIMER ();

  ainit ();
  err_init ();

  timecb_check ();
  STOP_ACHECK_TIMER ();
  for (;;)
    _acheck ();
}

int async_init::count;

void
async_init::start ()
{
  static bool initialized;
  if (initialized)
    panic ("async_init called twice\n");
  initialized = true;

  /* Ignore SIGPIPE, since we may get a lot of these */
  struct sigaction sa;
  bzero (&sa, sizeof (sa));
  sa.sa_handler = SIG_IGN;
  sigaction (SIGPIPE, &sa, NULL);

  if (!execsafe ()) {
    int fdlim_hard = fdlim_get (1);
    if (char *p = getenv ("FDLIM_HARD")) {
      int n = atoi (p);
      if (n > 0 && n < fdlim_hard) {
	fdlim_hard = n;
	fdlim_set (fdlim_hard, -1);
      }
    }
  }
  if (!getenv ("FDLIM_HARD") || !execsafe ()) {
    str var = strbuf ("FDLIM_HARD=%d", fdlim_get (1));
    xputenv (const_cast<char*>(var.cstr()));
    var = strbuf ("FDLIM_SOFT=%d", fdlim_get (0));
    xputenv (const_cast<char*>(var.cstr()));
  }

  sfs_core::selector_t::init ();
  sfs_core::selector = New sfs_core::std_selector_t ();

  lazylist = New list<lazycb_t, &lazycb_t::link>;

#ifdef WRAP_DEBUG 
  if (char *p = getenv ("CALLBACK_TRACE")) {
    if (strchr (p, 'f'))
      callback_trace |= CBTR_FD;
    if (strchr (p, 't'))
      callback_trace |= CBTR_TIME;
    if (strchr (p, 's'))
      callback_trace |= CBTR_SIG;
    if (strchr (p, 'c'))
      callback_trace |= CBTR_CHLD;
    if (strchr (p, 'l'))
      callback_trace |= CBTR_LAZY;
    if (strchr (p, 'a'))
      callback_trace |= -1;
    if (strchr (p, 'T'))
      callback_time = true;
  }
#endif /* WRAP_DEBUG */

  if (char *p = getenv ("SFS_OPTIONS")) {
    for (const char *cp = p; *cp; cp++) {
      switch (*cp) {
      case 'b':
	sfs_core::set_busywait (true);
	break;
      case 'e':
	if (sfs_core::set_select_policy (sfs_core::SELECT_EPOLL) < 0)
	  warn ("failed to switch select policy to EPOLL\n");
	break;
      case 'k':
	if (sfs_core::set_select_policy (sfs_core::SELECT_KQUEUE) < 0)
	  warn ("failed to switch select policy to KQUEUE\n");
	break;
      case 'z':
	sfs_core::set_zombie_collect (true);
	break;
      default:
	warn ("unknown SFS_OPTION: '%c'\n", *cp);
	break;
      }
    }
  }
}

sfs_core::select_policy_t 
sfs_core::select_policy_from_str (const str &s)
{
  sfs_core::select_policy_t ret = SELECT_NONE;
  if (s && s.len () > 0) {
    char c = s[0];
    ret = select_policy_from_char (c);
  }
  return ret;
}

sfs_core::select_policy_t
sfs_core::select_policy_from_char (char c)
{
  select_policy_t ret = SELECT_NONE;
  switch (c) {
  case 'k':
  case 'K':
    ret = SELECT_KQUEUE;
    break;
  case 'p':
  case 'P':
    ret = SELECT_EPOLL;
    break;
  case 's':
  case 'S':
    ret = SELECT_STD;
    break;
  default:
    break;
  }
  return ret;
}

void
async_init::stop ()
{
  err_flush ();
}

#ifdef WRAP_DEBUG

void callback_trace_fdcb (int i, int fd, cbv::ptr cb)
{
  if (fd != errfd && fd != sigpipes[0] && (callback_trace & CBTR_FD))
    warn ("CALLBACK_TRACE: %sfdcb %d%c %s <- %s\n",
	  timestring (), fd, "rwe"[i],
	  cb->dest, cb->line);
}

#endif /* WRAP_DEBUG */
