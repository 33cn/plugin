
#include "sfs_select.h"
#include "fdlim.h"

//-----------------------------------------------------------------------
// 
//  First do PTH-style select policies
//

#ifdef HAVE_TAME_PTH

int sfs_cb_ins;

int sfs_core_select; // XXX bkwds compat

static int
check_cb_ins (void *dummy)
{
  return sfs_cb_ins;
}

int 
sfs_pth_core_select (int nfds, fd_set *rfds, fd_set *wfds,
		     fd_set *efds, struct timeval *timeout)
{
  // Set an arbitrary timeout of 10 seconds
  int pth_timeout = 10;

  // clear the CB insertion flag; it's set to 1 when a new CB is
  // inserted via timecb, sigcb or fdcb
  sfs_cb_ins = 0;

  pth_event_t ev = pth_event (PTH_EVENT_FUNC, check_cb_ins, 0, 
			      pth_time (pth_timeout, 0));
  
  return pth_select_ev(nfds, rfds, wfds, efds, timeout, ev);
		       
}

void sfs_add_new_cb () { sfs_cb_ins = 1; }

#else /* HAVE_PTH */

void sfs_add_new_cb () {}

#endif /* HAVE_PTH */

//
//  end of PTH stuff
//
//-----------------------------------------------------------------------


//-----------------------------------------------------------------------
//
// Different select(2) options for core loop
//

#define FD_SETSIZE_ROUND (sizeof (long))

namespace sfs_core {

  void
  selector_t::init (void)
  {
    maxfd = fdlim_get (0);

#if defined(HAVE_WIDE_SELECT) || defined(HAVE_EPOLL) || defined(HAVE_KQUEUE)
    if (!execsafe () || fdlim_set (FDLIM_MAX, 1) < 0)
      fdlim_set (fdlim_get (1), 0);
    fd_set_bytes = (maxfd+7)/8;
    if (fd_set_bytes % FD_SETSIZE_ROUND)
      fd_set_bytes += FD_SETSIZE_ROUND - (fd_set_bytes % FD_SETSIZE_ROUND);

# else /* !WIDE_SELECT and friends */
    fdlim_set (FD_SETSIZE, execsafe ());
    fd_set_bytes = sizeof (fd_set);

#endif /* !WIDE_SELECT */

  }

  //-----------------------------------------------------------------------

  selector_t::selector_t ()
  {
    for (int i = 0; i < fdsn; i++) {
      _fdcbs[i] = New cbv::ptr[maxfd];
    }
  }

  selector_t::selector_t (selector_t *old)
  {
    for (int i = 0; i < fdsn; i++) {
      _fdcbs[i] = old->fdcbs () [i];
    }
  }

  selector_t::~selector_t () {}

  //-----------------------------------------------------------------------

  int selector_t::fd_set_bytes;
  int selector_t::maxfd;

  //-----------------------------------------------------------------------
  
};

//
//
//-----------------------------------------------------------------------
