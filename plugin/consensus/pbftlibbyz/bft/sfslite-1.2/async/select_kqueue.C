
#include "sfs_select.h"
#include <time.h>
#include "litetime.h"
#include "async.h"

#ifdef HAVE_KQUEUE

namespace sfs_core {

  //-----------------------------------------------------------------------
  
  static void
  kq_warn (const str &s, const struct kevent &kev, const kqueue_fd_t *fd)
  {
    strbuf b;
    b << s << ": "
	 << "fd=" << kev.ident << "; "
	 << "filter=" << kev.filter << "; "
	 << "flags=" << kev.flags << "; "
	 << "data=" << kev.data;

    if (fd && fd->file ()) {
      b << "; file=" << fd->file () << ":" << fd->line ();
    }
    b << "\n";
    str tmp = b;
    fprintf (stderr, tmp.cstr ());
  }
   
  //-----------------------------------------------------------------------
  
  kqueue_selector_t::kqueue_selector_t (selector_t *old)
    : selector_t (old)
  {
    if ((_kq = kqueue ()) < 0)
      panic ("kqueue: %m\n");
  }

  //-----------------------------------------------------------------------
  
  kqueue_selector_t::~kqueue_selector_t () {}

  //-----------------------------------------------------------------------

  kqueue_fd_t::kqueue_fd_t ()
    : _flips (0), _on (false), _file (NULL), _line (-1) {}

  //-----------------------------------------------------------------------

  void
  kqueue_fd_t::clear ()
  {
    _flips = 0;
  }

  //-----------------------------------------------------------------------

  bool
  kqueue_fd_t::toggle (bool on, const char *file, int line)
  {
    bool rc = (_flips == 0);
    _flips ++;
    _on = on;
    if (_on) {
      _file = file;
      _line = line;
    }
    return rc;
  }

  //-----------------------------------------------------------------------

  void
  kqueue_fd_set_t::toggle (bool on, int fd, selop op, const char *file,
			   int line)
  {
    if (int (_fds[op].size ()) <= fd) {
      _fds[op].setsize (fd + 1);
    }
    if (_fds[op][fd].toggle (on, file, line)) {
      _active.push_back (kqueue_fd_id_t (fd, op));
    }
  }

  //-----------------------------------------------------------------------

  void 
  kqueue_fd_set_t::export_to_kernel (vec<struct kevent> *out)
  {
    out->setsize (0);
    for (size_t i = 0; i < _active.size (); i++) {
      const kqueue_fd_id_t &id = _active[i];
      size_t fd_i = id.fd ();
      assert (_fds[id._op].size () > fd_i);
      kqueue_fd_t &fd = _fds[id._op][fd_i];
      if (fd.any_flips ()) {
	struct kevent &kev = out->push_back ();
	short filter = (id._op == selread) ? EVFILT_READ : EVFILT_WRITE;
	u_short flags = fd.on () ? EV_ADD : EV_DELETE;
	memset (&kev, 0, sizeof (kev));
	EV_SET (&kev, fd_i, filter, flags, 0, 0, 0);
	fd.set_removal_bit ();
      }
      fd.clear ();
    }
    _active.setsize (0);
  }

  //-----------------------------------------------------------------------
  
  void
  kqueue_selector_t::_fdcb (int fd, selop op, cbv::ptr cb, const char *f, int l)
  {
    assert (fd >= 0);
    assert (fd < maxfd);

    _fdcbs[op][fd] = cb;
    _set.toggle (cb, fd, op, f, l);
  }

  //-----------------------------------------------------------------------
  
  static void
  val2spec (const struct timeval *in, struct timespec *out)
  {
    out->tv_sec = in->tv_sec;
    out->tv_nsec = 1000 * in->tv_usec;
  }

  //-----------------------------------------------------------------------
  
  bool
  kqueue_fd_id_t::convert (const struct kevent &kev)
  {
    bool ret = true;
    _fd = kev.ident;
    switch (kev.filter) {
    case EVFILT_READ:
      _op = int (selread);
      break;
    case EVFILT_WRITE:
      _op = int (selwrite);
      break;
    default:
      ret = false;
      break;
    }
    return ret;
  }

  //-----------------------------------------------------------------------

  const kqueue_fd_t *
  kqueue_fd_set_t::lookup (const struct kevent &kev) const
  {
    const kqueue_fd_t *ret = NULL;
    kqueue_fd_id_t id;
    if (id.convert (kev))
      ret = lookup (id);
    return ret;
  }

  //-----------------------------------------------------------------------

  const kqueue_fd_t *
  kqueue_fd_set_t::lookup (const kqueue_fd_id_t &id) const
  {
    const kqueue_fd_t *ret = NULL;
    size_t fd_i = id.fd ();
    if (fd_i < _fds[id._op].size ()) {
      ret = &_fds[id._op][fd_i];
    }
    return ret;
  }

  //-----------------------------------------------------------------------
  
  void
  kqueue_selector_t::fdcb_check (struct timeval *selwait)
  {
    struct timespec  ts;
    val2spec (selwait, &ts);
    _set.export_to_kernel (&_kq_changes);

    size_t outsz = max<size_t> (_kq_changes.size (), MIN_CHANGE_Q_SIZE);
    _kq_events_out.setsize (outsz);

    int rc = kevent (_kq, 
		     _kq_changes.base (), _kq_changes.size (), 
		     _kq_events_out.base (), outsz,
		     &ts);
    if (rc < 0) {
      if (errno == EINTR) { 
	fprintf (stderr, "kqueue resumable error (%d)\n", errno);
      } else {
	panic ("kqueue failure %m (%d)\n", errno);
      }
    } else {
      assert (rc <= int (outsz));
    }

    sfs_set_global_timestamp ();
    sigcb_check ();

    for (int i = 0; i < rc; i++) {
      const struct kevent &kev = _kq_events_out[i];
      kqueue_fd_id_t id;
      if (id.convert (kev)) {
	const kqueue_fd_t *fd = _set.lookup (id);

	if (kev.flags & EV_ERROR) {
	  if (!fd || !fd->removal ())
	    kq_warn ("kqueue kernel error", kev, fd);
	} else {
	  cbv::ptr cb = _fdcbs[id._op][id._fd];
	  if (cb) {
	    sfs_leave_sel_loop ();
	    (*cb) ();
	  }
	}
      } else {
	kq_warn ("kqueue unexpected event", kev, NULL);
      }
    }
  }
  //-----------------------------------------------------------------------

};

#endif /* HAVE_KQUEUE */
