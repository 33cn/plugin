
// -*-c++-*-
/* $Id: core.C 2654 2007-03-31 05:42:21Z max $ */

#include "tame_run.h"
#include "tame_event.h"
#include "tame_closure.h"
#include "tame_rendezvous.h"
#include "tame_thread.h"

int tame_options;

int tame_global_int;
u_int64_t closure_serial_number;
ptr<closure_t> __cls_g;
ptr<closure_t> null_closure;

int tame_init::count;

void
tame_init::start ()
{
  static bool initialized;
  if (initialized)
    panic ("tame_init called twice\n");
  initialized = true;

#ifdef HAVE_TAME_PTH
  pth_init ();
#endif

  tame_options = 0;
  closure_serial_number = 0;
  tame_collect_rv_flag = false;
  __cls_g = NULL;
  null_closure = NULL;
  g_stats = New tame_stats_t ();

  tame_thread_init ();

  tame_options = 0;

  char *e = safegetenv (TAME_OPTIONS);
  for (char *cp = e; cp && *cp; cp++) {
    switch (*cp) {
    case 'V':
      tame_options |= TAME_ALWAYS_VIRTUAL;
      break;
    case 'R':
      tame_options |= TAME_RECYCLE_EVENTS;
      break;
    case 'S':
      tame_options |= TAME_STRICT;
      break;
    case 'Q':
      tame_options |= TAME_ERROR_SILENT;
      break;
    case 'A':
      tame_options |= TAME_ERROR_FATAL;
      break;
    case 'L':
      tame_options |= TAME_CHECK_LEAKS;
      break;
    case 'O':
      tame_options |= TAME_OPTIMIZE;
      break;
    case 's':
      g_stats->enable ();
      break;
    }
  }
}

void
tame_init::stop ()
{
  g_stats->dump ();
}
