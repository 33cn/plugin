
#include "tame_thread.h"
#include "async.h"

#ifdef HAVE_TAME_PTH
# include <pth.h>
#endif /* HAVE_TAME_PTH */

int threads_out;

void
tame_thread_spawn (const char *loc, void * (*fn) (void *), void *arg)
{
#ifdef HAVE_TAME_PTH

  //warn << "thread spawn ....\n";
  pth_attr_t attr = pth_attr_new ();
  pth_attr_set (attr, PTH_ATTR_NAME, loc); 
  pth_attr_set (attr, PTH_ATTR_STACK_SIZE, 0x10000);

  // Must not be joinable ; no one will join....
  pth_attr_set (attr, PTH_ATTR_JOINABLE, FALSE);

  threads_out ++;
    
  pth_spawn (attr, fn, arg);

  pth_attr_destroy (attr);

#else 
  panic ("no PTH package available\n");
#endif /* HAVE_TAME_PTH */
  
}

void
tame_thread_exit ()
{
#ifdef HAVE_TAME_PTH
  --threads_out;
  pth_exit (NULL);
#else 
  panic ("no PTH package available\n");
#endif /* HAVE_TAME_PTH */
}

void
tame_thread_init ()
{
#ifdef HAVE_TAME_PTH
  pth_init ();
  threads_out = 0;
#endif /* HAVE_TAME_PTH */
}

