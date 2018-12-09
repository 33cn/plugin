// -*-c++-*-
/* $Id: event.C 2738 2007-04-16 19:35:40Z max $ */

#include "tame_event.h"
#include "tame_closure.h"
#include "tame_event_ag.h"
#include "async.h"

nil_t g_nil;

void 
_event_cancel_base::cancel ()
{
  _cancelled = true;
  clear ();
  if (_cancel_notifier) {
    ptr<_event_cancel_base> hold (mkref (this));
    if (!_cancel_notifier->cancelled ()) {
      _cancel_notifier->trigger ();
    }
    _cancel_notifier = NULL;
  }
}

void
_event_cancel_base::clear ()
{
  if (!_cleared) {
    clear_action ();
    _cleared = true;
  }
}
