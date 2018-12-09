// -*-c++-*-
/* $Id: tame_event.C 2264 2006-10-11 17:42:40Z max $ */

#include "tame_recycle.h"

//-----------------------------------------------------------------------
// recycle bin for ref flags, used in both callback.h, and also
// tame.h/.C

static recycle_bin_t<obj_flag_t> *rfrb;
recycle_bin_t<obj_flag_t> * obj_flag_t::get_recycle_bin () { return rfrb; }

void
obj_flag_t::recycle (obj_flag_t *p)
{
  get_recycle_bin ()->recycle (p);
}

ptr<obj_flag_t>
obj_flag_t::alloc (const obj_state_t &b)
{
  ptr<obj_flag_t> ret = get_recycle_bin ()->get ();
  if (ret) {
    ret->set (b);
  } else {
    ret = New refcounted<obj_flag_t> (b);
  }
  assert (ret);
  return ret;
}


//
//-----------------------------------------------------------------------


int recycle_init::count;

void
recycle_init::start ()
{
  static bool initialized;
  if (initialized)
	  panic ("ref_flag_init::start called twice");
  initialized = true;
  rfrb = New recycle_bin_t<obj_flag_t> ();
}

void
recycle_init::stop () {}

