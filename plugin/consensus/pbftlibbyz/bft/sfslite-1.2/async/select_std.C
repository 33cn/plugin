
#include "sfs_select.h"
#include "async.h"
#include "litetime.h"
#include "corebench.h"

namespace sfs_core {

  //-----------------------------------------------------------------------

  std_selector_t::std_selector_t ()
    : selector_t (),
      _compact_interval (0),
      _n_fdcb_iter (0),
      _nselfd (0),
      _busywait (false),
      _last_fd (-1),
      _last_i (-1),
      _n_repeats (0)
  {
    init_fdsets ();
    for (size_t i = 0; i < fdsn; i++) {
      _src_locs[i] = New src_loc_t[maxfd];
    }
  }

  //-----------------------------------------------------------------------

  std_selector_t::std_selector_t (selector_t *old)
    : selector_t (old),
      _compact_interval (0),
      _n_fdcb_iter (0),
      _nselfd (0),
      _busywait (false)
  {
    init_fdsets ();
  }

  //-----------------------------------------------------------------------

  std_selector_t::~std_selector_t ()
  {
    for (int i = 0; i < fdsn; i++) {
      xfree (_fdsp[i]);
      xfree (_fdspt[i]);
      delete [] _src_locs[i];
    }
  }

  //-----------------------------------------------------------------------

  void
  std_selector_t::init_fdsets ()
  {
    for (int i = 0; i < fdsn; i++) {
      _fdsp[i] = (fd_set *) xmalloc (fd_set_bytes);
      bzero (_fdsp[i], fd_set_bytes);
      _fdspt[i] = (fd_set *) xmalloc (fd_set_bytes);
      bzero (_fdspt[i], fd_set_bytes);
    }
  }

  //-----------------------------------------------------------------------

  void
  std_selector_t::_fdcb (int fd, selop op, cbv::ptr cb, const char *f, int l)
  {
    assert (fd >= 0);
    assert (fd < maxfd);
    _fdcbs[op][fd] = cb;
    if (cb) {
      _src_locs[op][fd].set (f, l);
      sfs_add_new_cb ();
      if (fd >= _nselfd)
	_nselfd = fd + 1;
      FD_SET (fd, _fdsp[op]);
    } else {
      _src_locs[op][fd].clear ();
      FD_CLR (fd, _fdsp[op]);
    }
  }

  //-----------------------------------------------------------------------

  void
  std_selector_t::compact_nselfd ()
  {
    int max_tmp = 0;
    for (int i = 0; i < _nselfd; i++) {
      for (int j = 0; j < fdsn; j++) {
	if (FD_ISSET(i, _fdsp[j]))
	  max_tmp = i;
      }
    }
    _nselfd = max_tmp + 1;
  }

  //-----------------------------------------------------------------------

  void
  std_selector_t::select_failure ()
  {
    warn ("select: %m\n");
    const char *typ[] = { "reading" , "writing" };
    for (int k = 0; k < 2; k++) {
      warnx << "Select Set Dump: " << typ[k] << " { " ;
      for (int j = 0; j < maxfd; j++) {
	if (FD_ISSET (j, _fdspt[k])) {
	  warnx << j << " ";
	}
      }
      warnx << " }\n";
    }
    panic ("Aborting due to select() failure\n");
  }

  //-----------------------------------------------------------------------
  
  void
  std_selector_t::fdcb_check (struct timeval *selwait)
  {

    //
    // If there was a request to compact nselfd every compact_interval,
    // then examine the fd sets and make the adjustment.
    //
    if (_compact_interval && (++_n_fdcb_iter % _compact_interval) == 0) 
      compact_nselfd ();
  
    for (int i = 0; i < fdsn; i++)
      memcpy (_fdspt[i], _fdsp[i], fd_set_bytes);

    if (_busywait) {
      memset (selwait, 0, sizeof (*selwait));
    }
    
    int n = SFS_SELECT (_nselfd, _fdspt[0], _fdspt[1], NULL, selwait);

    // warn << "select exit rc=" << n << "\n";
    if (n < 0 && errno != EINTR) {
      select_failure ();
    }

    sfs_set_global_timestamp ();
    sigcb_check ();

    for (int fd = 0; fd < maxfd && n > 0; fd++)
      for (int i = 0; i < fdsn; i++)
	if (FD_ISSET (fd, _fdspt[i])) {
	  n--;
	  if (FD_ISSET (fd, _fdsp[i])) {

	    // Comment out this check for now; since the bug has
	    // been fixed, I don't think we need it any more.
	    if (0 && _last_fd == fd && _last_i == i) {
	      _n_repeats ++;
	      if (_n_repeats > 0 && _n_repeats % 1000 == 0) {
		strbuf b;
		str l = _src_locs[i][fd].to_str ();
		b << "XXX repeatedly (" << _n_repeats 
		  << ") ready FD indidicates possible bug: "
		  << "fd=" << fd << "; op=" << i << "; "
		  << "cb=" << l << "\n";
		str s = b;
		fprintf (stderr, "%s", s.cstr ());
	      }
	    } else {
	      _n_repeats = 0;
	      _last_fd = fd;
	      _last_i = i;
	    }


#ifdef WRAP_DEBUG
	    callback_trace_fdcb (i, fd, _fdcbs[i][fd]);
#endif /* WRAP_DEBUG */
	    STOP_ACHECK_TIMER ();
	    sfs_leave_sel_loop ();
	    (*_fdcbs[i][fd]) ();
	    START_ACHECK_TIMER ();
	  }
	}
  }
  
  //-----------------------------------------------------------------------

  void
  src_loc_t::set (const char *f, int l) 
  {
    _file = f;
    _line = l;
  }

  //-----------------------------------------------------------------------

  void
  src_loc_t::clear () 
  {
    _file = NULL;
    _line = 0;
  }

  //-----------------------------------------------------------------------

  str
  src_loc_t::to_str () const
  {
    str ret;
    if (!_line) {
      ret = "<N/A>";
    } else {
      strbuf b;
      b << _file << ":" << _line;
      ret = b;
    }
    return ret;
  }

  //-----------------------------------------------------------------------
  
};
