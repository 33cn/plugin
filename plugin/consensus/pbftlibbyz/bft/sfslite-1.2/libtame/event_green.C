
#include "tame_recycle.h"
#include "tame_event_green.h"

static recycle_bin_t<green_event_t<void> > *_vrb;

//-----------------------------------------------------------------------
// optimized events

namespace green_event {

  recycle_bin_t<green_event_t<void> > *
  vrb () 
  {
    if (!_vrb) {
      _vrb = New recycle_bin_t<green_event_t<void> > (0x100000);
    }
    return _vrb;
  }
}

RECYCLE_EVENT_C(int,int);
RECYCLE_EVENT_C(bool,bool);

//
//-----------------------------------------------------------------------
