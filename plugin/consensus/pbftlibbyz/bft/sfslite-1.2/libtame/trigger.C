
#include "tame_trigger.h"

void
dtrigger (event<>::ref cb)
{
  delaycb (0, 0, wrap (cb, &_event<>::trigger));
}

