
// -*-c++-*-
#include "tame.h"
#include "tame_io.h"
#include "tame_connectors.h"
#include "tame_nlock.h"

struct selop_which_t {
  selop_which_t (int f, selop w) : _fd (f), _which (w) {}
  int _fd;
  selop _which;
};

template<> struct equals<selop_which_t> {
  equals () {}
  bool operator() (const selop_which_t &a, const selop_which_t &b) const
  { return (a._fd == b._fd && a._which == b._which); }
};

template<> struct hashfn<selop_which_t> {
  hashfn () {}
  bool operator() (const selop_which_t &x) const
  { return hash_t ((x._fd << 3) | int (x._which)); }
};

namespace tame {

void
clearread (int fd)
{
  fdcb (fd, selread, NULL);
}

void
clearwrite (int fd)
{
  fdcb (fd, selwrite, NULL);
}

tamed void
sigcb1 (int sig, evv_t cb)
{
  twait { sigcb (sig, connector::cnc (mkevent (), cb)); }
  sigcb (sig, NULL);
  cb->trigger ();
}


void
waitwrite (int fd, evv_t cb)
{
  fdcb1(fd, selwrite, cb);
}

void
waitread (int fd, evv_t cb)
{
  fdcb1(fd, selread, cb);
}


tamed void
iofd_t::on (evv_t cb)
{
  tvars {
    outcome_t oc (OUTCOME_SUCC);
  }
  _on = true;
  twait { fdcb (_fd, _op, connector::cnc (mkevent (), cb, &oc)); }
  _on = false;
  if (oc != OUTCOME_CANCELLED) {
    cb->trigger ();
  }
  if (!_on) {
    off (false);
  }
}

void
iofd_t::off (bool check)
{
  if (_on || !check) {
    fdcb (_fd, _op, NULL);
    _on = false;
  }
}

typedef enum { READ, WRITE, ENDPROG, POKE } ev_t;

void
iofd_sticky_t::on ()
{
  if (!_on && _ev)
    fdcb (_fd, _op, _ev);

  if (_ev) 
    _on = true;
}

void
iofd_sticky_t::off ()
{
  if (_on)
    fdcb (_fd, _op, NULL);

  _on = false;
}

void
iofd_sticky_t::finish ()
{
  off ();
  _ev = NULL;
}

tamed void
proxy (int infd, int outfd, evv_t ev)
{
  tvars {
    ref<std_proxy_t> px (New refcounted<std_proxy_t> ());
  }
  twait { px->go (infd, outfd, mkevent ()); }
  ev->trigger ();
}

bool
std_proxy_t::is_readable () const
{
  return (_sz > _buf.resid ());
}

bool
std_proxy_t::is_writable () const 
{
  return (_buf.resid () > 0);
}

int
std_proxy_t::v_read (int fd)
{
  return _buf.input (fd, _sz - _buf.resid ());
}

int 
std_proxy_t::v_write (int fd)
{
  return _buf.output (fd);
}

std_proxy_t::std_proxy_t (const str &d, ssize_t s) : 
  proxy_t (d), 
  _sz (s > 0 ? s : 0x4000) {}

std_proxy_t::~std_proxy_t () {}

void
proxy_t::do_debug (const str &msg) const
{
  if (msg && _debug_level > 0) {
    str s = _debug_name;
    if (!s) 
      s = "<anonymous>";
    warn << s << ": " << msg;
    if (msg[msg.len () - 1] != '\n')
      warnx << "\n";
  }
}

tamed void
proxy_t::go (int infd, int outfd, evv_t ev)
{
  tvars {
    rendezvous_t<ev_t> rv (__FILE__, __LINE__);
    ev_t which;
    int rc;
    bool err (false);
    iofd_sticky_t read (infd, selread), write (outfd, selwrite);
    bool cancelled (false);
    str s;
    holdvar ptr<proxy_t> hold (mkref (_self));
  }

  // If the parent process gives up on us, then we'll get a callback here
  ev->set_cancel_notifier (mkevent (rv, ENDPROG));

  read.setev (mkevent (rv, READ));
  write.setev (mkevent (rv, WRITE));

  _poke_ev = mkevent (rv, POKE);
  _poke_ev->set_reuse (true);

  while (!err && (!(is_eof() || cancelled) || is_writable ())) {
    
    if (is_readable () && !is_eof ()) read.on ();
    else read.off ();
    
    if (is_writable ()) write.on ();
    else write.off ();

    twait (rv, which);

    switch (which) {
    case POKE:
      break;
    case READ: 
      if ((rc = v_read (read.fd())) == 0) {
	do_debug ("EOF on socket");
	set_eof ();
      } else if (rc < 0) {
	if (read_error (&s)) {
	  warn ("read error: %s\n", s.cstr ());
	  err = true;
	} else if (errno != EAGAIN) {
	  warn ("read error: %m\n");
	  err = true;
	}
      }
      do_debug (strbuf ("read() => (%d,%d)", rc, errno));
      break;
    case WRITE:
      if ((rc = v_write (write.fd())) < 0) {
	if (write_error (&s)) {
	  warn ("write error: %s\n", s.cstr ());
	  err = true;
	} else if (errno != EAGAIN) {
	  warn ("write error: %m\n");
	  err = true;
	}
      }
      do_debug (strbuf ("write() => (%d,%d)", rc, errno));
      break;
    case ENDPROG:
      do_debug ("end program");
      cancelled = true;
      break;
    default:
      panic ("unexpected case.\n");
    }
  }

  do_debug ("leaving proxy loop");

  read.finish ();
  write.finish ();
  _poke_ev = NULL;

  ev->trigger ();
  rv.cancel (); // for the timer (if necessary )
}

bool
proxy_t::poke ()
{
  if (_poke_ev) {
    _poke_ev->trigger ();
    return true;
  } else {
    return false;
  }
}

tamed void
read (int fd, char *buf, size_t sz, evi_t ev)
{
  tvars { int rc; }

  twait { fdcb (fd, selread, mkevent ()); }
  fdcb (fd, selread, NULL);
  rc = ::read (fd, buf, sz);
  ev->trigger (rc);
}

tamed void
write (int fd, const char *buf, size_t sz, evi_t ev)
{
  tvars { int rc; }

  twait { fdcb (fd, selwrite, mkevent ()); }
  fdcb (fd, selwrite, NULL);
  rc = ::write (fd, buf, sz);
  ev->trigger (rc);
}

tamed void 
accept (int sockfd, struct sockaddr *addr, socklen_t *addrlen, evi_t ev)
{
  tvars { int rc; }
  
  twait { fdcb (sockfd, selread, mkevent ()); }
  fdcb (sockfd, selread, NULL);
  rc = ::accept (sockfd, addr, addrlen);
  ev->trigger (rc);
}
 
  //-----------------------------------------------------------------------

  lock_table_t<selop_which_t> locktab;
  
  tamed void 
  fdcb1 (int fd, selop which, evv_t cb)
  {
    tvars {
      ptr<lock_handle_t<selop_which_t> > lh;
      selop_which_t sw (fd, which);
    }

    twait { locktab.acquire (&lh, sw, lock_t::EXCLUSIVE, mkevent ()); }
    twait { fdcb (fd, which, mkevent ()); }
    fdcb (fd, which, NULL);
    lh->release ();
    cb->trigger ();
  }

  //-----------------------------------------------------------------------

};
