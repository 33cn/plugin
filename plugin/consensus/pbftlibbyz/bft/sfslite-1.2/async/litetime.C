
#include "litetime.h"
#include "async.h"
#include <sys/time.h>
#include <stdlib.h>
#include <sys/mman.h>
#include <unistd.h>
#include <fcntl.h>
#include <stdio.h>
#include "parseopt.h"

//-----------------------------------------------------------------------
// Begin Global Clock State
//


struct mmap_clock_t {
  mmap_clock_t (const str &fn) 
    : mmp (NULL), nbad (0), fd (-1), file (fn), 
    mmp_sz (sizeof (struct timespec) * 2)
  {
    ::clock_gettime (CLOCK_REALTIME, &last);
  }
  ~mmap_clock_t () ;
  bool init ();
  int clock_gettime (struct timespec *ts);

  enum { stale_threshhold = 50000 // number of stale values before we give up
  };

  struct timespec *mmp; // mmap'ed pointer
  int nbad;             // number of calls since diff
  struct timespec last; // last returned value
  int fd;               // fd for the file
  const str file;       // file name of the mmaped clock file
  const size_t mmp_sz;  // size of the mmaped region
  
};

struct sfs_clock_state_t {
  sfs_clock_state_t () {}

  void clear ();
  void init_from_env ();
  void set (sfs_clock_t t, const str &arg, bool lzy);

  bool enable_mmap_clock (const str &arg);
  void disable_mmap_clock ();
  bool enable_timer ();
  bool disable_timer ();
  void mmap_clock_fail ();
  int my_clock_gettime (struct timespec *tp);
  void get_tsnow (struct timespec *ts, bool frc);
  time_t get_timenow (bool frc);
  void set_timestamp ();
  void refresh_timestamp ();

  inline void left_sel_loop () { _left_sel_loop = true; }

  bool _timer_enabled;
  sfs_clock_t _type;
  bool _lazy_clock;
  str _mmap_clock_loc;
  mmap_clock_t *_mmap_clock;
  int _timer_res;
  bool _need_refresh;
  bool _left_sel_loop;

  struct timespec _tsnow;
};

sfs_clock_state_t g_clockstate;

//
// End Global Clock State
//-----------------------------------------------------------------------


//-----------------------------------------------------------------------
// Mmap Clock Type
//

bool
sfs_clock_state_t::enable_mmap_clock (const str &arg)
{
  if (_mmap_clock) 
    return true;

  _mmap_clock = New mmap_clock_t (arg);
  return _mmap_clock->init ();
}

void
sfs_clock_state_t::disable_mmap_clock ()
{
  if (_mmap_clock) {
    delete _mmap_clock;
    _mmap_clock = NULL;
  }
}

void
sfs_clock_state_t::mmap_clock_fail ()
{
  warn << "*mmap clock failed: reverting to stable clock\n";
  disable_mmap_clock ();
  _type =  SFS_CLOCK_GETTIME;
}


mmap_clock_t::~mmap_clock_t ()
{
  if (mmp) 
    munmap (mmp, mmp_sz);
  if (fd >= 0)
    close (fd);
}

bool
mmap_clock_t::init ()
{
  void *tmp;
  struct stat sb;

  if ((fd = open (file.cstr (), O_RDONLY)) < 0) {
    warn ("%s: mmap clock file open failed: %m\n", file.cstr ());
    return false;
  }

  if (fstat (fd, &sb) < 0) {
    warn ("%s: cannot fstat file: %m\n", file.cstr ());
    return false;
  }

  // on Linux, sb.st_size is signed
  if (int (sb.st_size) < int (mmp_sz)) {
    warn << file << ": short file; aborting\n";
    return false; 
  }

  u_int opts = MAP_SHARED;
#ifdef HAVE_MAP_NOSYNC
  opts |= MAP_NOSYNC;
#endif

  tmp = mmap (NULL, mmp_sz, PROT_READ, opts , fd, 0); 

  if (tmp == MAP_FAILED) {
    warn ("%s: mmap clock mmap failed: %m\n", file.cstr ());
    return false;
  }
  mmp = (struct timespec *)tmp;
  warn << "*unstable: mmap clock initialized\n";
  return true;
}

int
mmap_clock_t::clock_gettime (struct timespec *out)
{
  struct timespec tmp;

  *out = mmp[0];
  tmp = mmp[1];

  // either we're unlucky or the guy crashed an a strange state
  // which is unlucky too
  if (!TIMESPEC_EQ (*out, tmp)) {
    
    // debug message
    //warn << "*mmap clock: reverting to clock_gettime\n";

    ::clock_gettime (CLOCK_REALTIME, out);
    last = *out;
    ++nbad;

    //
    // likely case -- timestamp in shared memory is stale.
    //
  } else if (TIMESPEC_LT (*out, last)) {
    TIMESPEC_INC (&last) ;
    *out = last;
    ++nbad;

    //
    // if the two stamps are equal, and they're strictly greater than
    // any timestamp we've previously issued, then it's OK to use it
    // as normal. this should be happening every so often..
    //
  } else {
    last = *out;
    nbad = 0;
  }

  if (nbad > stale_threshhold) 
    // will delete this, so be careful
    g_clockstate.mmap_clock_fail ();

  return 0;
}


//
// End of Mmap clock type
//-----------------------------------------------------------------------




//-----------------------------------------------------------------------
// TIMER clock type
//
static void 
clock_set_timer (int i)
{
  struct itimerval val;
  val.it_value.tv_sec = 0;
  val.it_value.tv_usec = i;
  val.it_interval.tv_sec = 0;
  val.it_interval.tv_usec = i;

  setitimer (ITIMER_REAL, &val, 0);
}

static void
clock_timer_event ()
{
  clock_gettime (CLOCK_REALTIME, &g_clockstate._tsnow);
}

bool
sfs_clock_state_t::enable_timer ()
{
  if (!_timer_enabled) {
    warn << "*unstable: enabling hardware timer\n";
    clock_timer_event ();
    _timer_enabled = true;
    sigcb (SIGALRM, wrap (clock_timer_event));
    clock_set_timer (_timer_res);
  }
  return true;
}

bool
sfs_clock_state_t::disable_timer ()
{
  if (_timer_enabled) {
    warn << "disabling timer\n";
    struct itimerval val;
    memset (&val, 0, sizeof (val));
    setitimer (ITIMER_REAL, &val, 0);
    _timer_enabled = false;
  }
  return true;
}
//
// End TIMER clock type
//-----------------------------------------------------------------------


//
//-----------------------------------------------------------------------
//
// main, publically available functions..
//

//
// set_sfs_clock
//
//    Set the core timing discipline for SFS.
//
void
sfs_set_clock (sfs_clock_t typ, const str &arg, bool lzy)
{
  g_clockstate.set (typ, arg, lzy);
}

//
// sfs_get_timenow
//
//    Get the # of seconds since 1970, and specify a "force" flag if
//    we need to real answer (and not the guestimated version based on
//    our last estimate).
//
time_t 
sfs_get_timenow (bool frc)
{
  return g_clockstate.get_timenow (frc);
}

//
// sfs_get_tsnow
//
//    Same as above, but get the whole timespec.
//
struct timespec 
sfs_get_tsnow (bool frc)
{
  struct timespec ts;
  g_clockstate.get_tsnow (&ts, frc);
  return ts;
}

void
sfs_get_tsnow (struct timespec *ts, bool frc)
{
  g_clockstate.get_tsnow (ts, frc);
}

//
// sfs_set_global_timestamp
//
//   Place a barrier in the code, after which, all calls to sfs_get_tsnow
//   etc are at least as accurate as when this function was called.
//
void
sfs_set_global_timestamp ()
{
  g_clockstate.set_timestamp ();
}

void
sfs_leave_sel_loop ()
{
  g_clockstate.left_sel_loop ();
}


//
// end of the public interface
//-----------------------------------------------------------------------

//-----------------------------------------------------------------------
// Time management

//
// my_clock_gettime
//
//    Function that is called many times through the event loop, and
//    thus, we might want to optimize it for our needes.
//
int
sfs_clock_state_t::my_clock_gettime (struct timespec *tp)
{
  int r = 0;
  switch (_type) {
  case SFS_CLOCK_GETTIME:
    r = clock_gettime (CLOCK_REALTIME, tp);
    break;
  case SFS_CLOCK_TIMER:
    _tsnow.tv_nsec ++;
    *tp = _tsnow;
    break;
  case SFS_CLOCK_MMAP:
    r = _mmap_clock->clock_gettime (tp);
    break;
  default:
    break;
  }
  return r;
}

void
sfs_clock_state_t::set_timestamp ()
{
  _need_refresh = true;
}

void
sfs_clock_state_t::get_tsnow (struct timespec *ts, bool frc)
{
  if (frc || (_need_refresh && _left_sel_loop)) {
    refresh_timestamp ();
  }
  *ts = _tsnow;
}

time_t
sfs_clock_state_t::get_timenow (bool frc)
{
  if (frc || (_need_refresh && _left_sel_loop)) {
    refresh_timestamp ();
  }
  return _tsnow.tv_sec;
}

void
sfs_clock_state_t::refresh_timestamp ()
{
  my_clock_gettime (&_tsnow);
  _need_refresh = false;
  _left_sel_loop = false;
}

// end time management
//-----------------------------------------------------------------------

//-----------------------------------------------------------------------
// Internal state management functions
//

void
sfs_clock_state_t::set (sfs_clock_t typ, const str &arg, bool lzy)
{
  switch (typ) {
  case SFS_CLOCK_TIMER:
    disable_mmap_clock ();
    _type = enable_timer () ? SFS_CLOCK_TIMER : SFS_CLOCK_GETTIME;
    break;
  case SFS_CLOCK_MMAP:
    disable_timer ();
    if (enable_mmap_clock (arg))
      _type = typ;
    else
      mmap_clock_fail ();
    break;
  case SFS_CLOCK_GETTIME:
    disable_timer ();
    disable_mmap_clock ();
    _type = typ;
    break;
  default:
    assert (false);
  }
  _lazy_clock = lzy;
}

//
//-----------------------------------------------------------------------

//-----------------------------------------------------------------------
// Runtime Initialization

int litetime_init::count;

void
litetime_init::start ()
{
  static bool initialized;
  if (initialized)
    panic ("litetime_init called twice\n");
  initialized = true;

  g_clockstate.clear ();
  g_clockstate.init_from_env ();
}

void litetime_init::stop () {}

void
sfs_clock_state_t::clear ()
{
  _timer_enabled = false;
  _type = SFS_CLOCK_GETTIME;
  _lazy_clock = false;
  _mmap_clock = NULL;
  _timer_res = 10000; // 10 ms
  _need_refresh = true;
  _left_sel_loop = true;
}

void
sfs_clock_state_t::init_from_env ()
{
  const char *p = getenv ("SFS_CLOCK_OPTIONS");
  if (p) {
    sfs_clock_t t = SFS_CLOCK_GETTIME;
    bool lzy = false;
    str arg;
    for (const char *c = p; c; c++) {
      switch (*c) {
      case 'T':
      case 't':
	t = SFS_CLOCK_TIMER;
	break;
      case 'l':
      case 'L':
	lzy = true;
	break;
      case 'm':
      case 'M':
	t = SFS_CLOCK_MMAP;
	break;
      default:
	warn ("Unknown SFS_CLOCK_OPTION: '%c'\n", *c);
	break;
      }
    }
    if (t == SFS_CLOCK_MMAP) {
      const char *p = getenv ("SFS_CLOCK_MMAP_FILE");
      if (!p) {
	warn ("Must provide SFS_CLOCK_MMAP_FILE location for mmap clock\n");
	t = SFS_CLOCK_GETTIME;
      } else {
	arg = p;
      }
    }

    if (t == SFS_CLOCK_TIMER) {
      const char *p = getenv ("SFS_CLOCK_TIMER_RESOLUTION");
      int res;
      if (p && convertint (p, &res)) {
	_timer_res = res;
      } else {
	warn ("Bad timer resolution specified.\n");
      }
    }
    set (t, arg, lzy);
  }
}

//
//-----------------------------------------------------------------------
