
#include "sfs_select.h"
#include "litetime.h"
#include "async.h"

#ifdef HAVE_EPOLL

namespace sfs_core {

  //-----------------------------------------------------------------------

  epoll_selector_t::epoll_selector_t (selector_t *old)
    : selector_t (old),
      _maxevents (maxfd *2)
  {
    if ((_epfd = epoll_create (maxfd)) < 0) {
      panic ("epoll_create(%d): %m\n", maxfd);
    }
    _ret_events = (epoll_event*)xmalloc(sizeof(struct epoll_event)*_maxevents);
    bzero(_ret_events, sizeof(struct epoll_event)*_maxevents);
    _epoll_states = (epoll_state*)xmalloc(sizeof(struct epoll_state)*maxfd);
    bzero(_epoll_states, sizeof(struct epoll_state)*maxfd);

  }
 
  //-----------------------------------------------------------------------

  epoll_selector_t::~epoll_selector_t ()
  {
    xfree (_ret_events);
    xfree (_epoll_states);
    close (_epfd);
  }

  //-----------------------------------------------------------------------

#define EV_READ_BIT   1
#define EV_WRITE_BIT  2
#define EV_READ_EVENTS  (EPOLLIN  | EPOLLHUP | EPOLLERR | EPOLLPRI)
#define EV_WRITE_EVENTS (EPOLLOUT | EPOLLHUP | EPOLLERR)
  
  // maps our change-in-event-state to instructions to epoll
  int
  epoll_selector_t::update_epoll_state(epoll_state* es)
  {
    int epoll_op = (es->user_events 
		    ? (es->in_epoll 
		       ? EPOLL_CTL_MOD 
		       : EPOLL_CTL_ADD)
		    : EPOLL_CTL_DEL);
    
    es->in_epoll = (es->user_events != 0);
    return epoll_op;
  }

  //-----------------------------------------------------------------------

  int
  epoll_selector_t::user_events_to_epoll_events(epoll_state* es)
  {
    int ret = 0;
    
    /* 
     * we're registering for hang up events and then
     * passing them onto our callers when our callers have
     * registered for read. that's because libasync doesn't have a
     * fdcb(fd, selerrror, cb) interface. In other words, people
     * using libasync expect only read and write events.
     */
    if (es->user_events & EV_READ_BIT)
      ret |= EV_READ_EVENTS;
    
    if (es->user_events & EV_WRITE_BIT)
      ret |= EV_WRITE_EVENTS;
    
    return ret;
  }
  
  //-----------------------------------------------------------------------

  void
  epoll_selector_t::_fdcb (int fd, selop op, cbv::ptr cb, const char *file,
			   int line)
  {
    assert(fd >= 0);
    assert(fd < maxfd);
    
    epoll_event ev;
    int epoll_op;
    epoll_state* es = &_epoll_states[fd];
    
    _fdcbs[op][fd] = cb;

    // keep gcc4 happy
    int op_as_int = static_cast<int>(op);

    if (cb) {
	/* analog of FD_SET */
	es->user_events |= (1 << op_as_int);
    } else {
	/* analog of FD_CLR */
	es->user_events &= ~(1 << op_as_int);
    }

    epoll_op   = update_epoll_state(es);
    ev.events  = user_events_to_epoll_events(es);
    ev.data.fd = fd;

    epoll_ctl(_epfd, epoll_op, fd, &ev);
  }
  
  //-----------------------------------------------------------------------

  // version of the "select loop" that uses epoll_wait instead.
  void
  epoll_selector_t::fdcb_check (struct timeval *selwait)
  {
    int timeout_ms = selwait->tv_usec / 1000 + selwait->tv_sec * 1000;
    int n = epoll_wait(_epfd, _ret_events, _maxevents, timeout_ms);
    
    if (n < 0 && errno != EINTR)
      panic ("epoll_wait: %m\n");
    
    sfs_set_global_timestamp ();

    sigcb_check();
    
    if (n < 0) return;
    
    for (int i = 0; i < n; i++) {
      
      epoll_event* eventp = &_ret_events[i];
      int fd = eventp->data.fd;
      int* interest = &_epoll_states[fd].user_events;
      
      /* analogous to calling FD_ISSET on the returned fd_set. second
       * condition is analogous to calling FD_ISSET on the 'master
       * copy'; we need to make sure that the event is still set (an
       * earlier event handler could have unreg'ed the handler for the
       * current socket fd). */
      if ( (eventp->events & EV_READ_EVENTS) && (*interest & EV_READ_BIT)) {
	sfs_leave_sel_loop ();
	(*_fdcbs[selread][fd]) ();
      }
      
      if ( (eventp->events & EV_WRITE_EVENTS) && (*interest & EV_WRITE_BIT)) {
	sfs_leave_sel_loop ();
	(*_fdcbs[selwrite][fd]) ();
      }
    }
  }

  //-----------------------------------------------------------------------

};


#endif /* HAVE_EPOLL */
