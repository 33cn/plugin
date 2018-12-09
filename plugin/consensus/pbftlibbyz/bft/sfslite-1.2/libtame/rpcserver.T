
// -*-c++-*-

#include "tame_rpcserver.h"
#include "parseopt.h"

namespace tame {

  server_t::server_t (int fd, int v) : _verbosity (v)
  {
    tcp_nodelay (fd);
    _x = axprt_stream::alloc (fd);
  }

  tamed void
  server_t::runloop ()
  {
    tvars {
      rendezvous_t<> rv (__FILE__, __LINE__);
      event<svccb *>::ptr ev;
      svccb *sbp;
      ptr<asrv> s;
    }
    
    ev = mkevent (rv, sbp);
    ev->set_reuse (true);
    
    s = asrv::alloc (_x, get_prog (), ev);
    
    do {
      twait (rv);
      if (sbp) {
	dispatch (sbp);
      }
    } while (sbp);
    
    if (_verbosity >= VERB_MED)
      warn << "EOF on socket received; deleting server...\n";

    ev->finish ();
    
    delete this;
  }

  void
  server_factory_t::new_connection (int lfd)
  {
    sockaddr_in sin;
    socklen_t sinlen = sizeof (sin);
    bzero (&sin, sinlen);
    int newfd = accept (lfd, reinterpret_cast<sockaddr *> (&sin), &sinlen);
    if (newfd >= 0) {
      if (_verbosity >= VERB_MED)
	warn ("accepting connection from %s\n", inet_ntoa (sin.sin_addr));
      server_t *srv = alloc_server (newfd, _verbosity);
      srv->runloop ();
    } else if (errno != EAGAIN) {
      if (_verbosity >= VERB_LOW)
	warn ("accept failure: %m\n");
    }
  }

  void
  server_factory_t::run (const str &s, evb_t done)
  {
    int port;
    if (!convertint (s, &port)) {
      warn << "Could not convert string to port: " << s << "\n";
      done->trigger (false);
    } else if (port <= 0 || port > int (USHRT_MAX)) {
      warn << "Invalid port specified: " << port << "\n";
      done->trigger (false);
    } else {
      run (u_int (port), done);
    }
  }

  tamed void
  server_factory_t::run_T (u_int port, evb_t done)
  {
    tvars {
      int fd;
      event<>::ptr ev;
      bool ret (false);
      bool go (true);
      rendezvous_t<bool> rv (__FILE__, __LINE__);
    }
    
    fd = inetsocket (SOCK_STREAM, port);
    if (fd <= 0) {
      warn << "cannot allocate TCP port: " << port << "\n";
    } else {
      ret = true;
      
      sigcb (SIGINT, mkevent (rv, false));
      sigcb (SIGTERM, mkevent (rv, false));
      
      close_on_exec (fd);
      listen (fd, 200);
      
      ev = mkevent (rv, true);
      ev->set_reuse (true);
      
      fdcb (fd, selread, ev);
      
      while (go) {
	twait (rv, go);
	if (go) 
	  new_connection (fd);
      }
      ev->finish ();
    }
    done->trigger (ret);
  }
  
};
