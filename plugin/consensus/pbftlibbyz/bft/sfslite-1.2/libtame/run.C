
// -*-c++-*-
/* $Id: core.C 2654 2007-03-31 05:42:21Z max $ */

#include "tame_run.h"

tame_stats_t *g_stats;


void 
tame_error (const char *loc, const char *msg)
{
  if (!(tame_options & TAME_ERROR_SILENT)) {
    if (loc) {
      warn << loc << ": " << msg << "\n";
    } else 
      warn << msg << "\n";
  }
  if (tame_options & TAME_ERROR_FATAL)
    panic ("abort on tame failure\n");
}

tame_stats_t::tame_stats_t ()
  : _collect (false),
    _n_evv_rec_hit (0),
    _n_evv_rec_miss (0),
    _n_mkevent (0),
    _n_mkclosure (0),
    _n_new_rv (0)
{}

void
tame_stats_t::_mkevent_impl_rv_alloc(const char *loc)
{
  int *c;
  if ((c = _mkevent_impl_rv[loc])) {
    (*c)++;
  } else {
    _mkevent_impl_rv.insert (loc, 1);
  }
}

void
tame_stats_t::dump()
{
  if (!_collect) 
    return;

  warn << "Tame statistics -------------------------------------------\n";
  warn << "  total events allocated: " << _n_mkevent << "\n";
  warn << "  total closures allocated: " << _n_mkclosure << "\n";
  warn << "  total RVs allocated: " << _n_new_rv << "\n";
  warn << "  event<> recyle hits/misses: "
       << _n_evv_rec_hit << "/" << _n_evv_rec_miss << "\n";
  warn << "  event allocations:\n";

  qhash_const_iterator_t<const char *, int> it (_mkevent_impl_rv);
  
  const char *const *k;
  int i;
  while ((k = it.next (&i))) {
    warn << "     " << i << "\t" << *k << "\n";
  }
}
