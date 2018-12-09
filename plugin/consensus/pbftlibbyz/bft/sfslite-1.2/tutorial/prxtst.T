
// -*-c++-*-
/* $Id: tcpconnect.T 2272 2006-10-27 02:25:59Z max $ */

#include "async.h"
#include "tame.h"
#include "tame_io.h"
#include "rxx.h"
#include "parseopt.h"

static void
usage ()
{
  warnx << "usage: progname <localport> <hostname>:<port>\n";
  exit (-1);
}

static int
listen_on (int port)
{
  int fd = inetsocket (SOCK_STREAM, port);
  if (fd >= 0) {
    close_on_exec (fd);
    listen (fd, 20);
  } else {
    warn ("listen failed on port %d: %m\n", port);
  }
  return fd;
}

typedef enum { PROXY_DONE, PROXY_END, READ, WRITE } ev_t;

tamed static void
do_proxy (int infd, str hn, int port, cbv cb)
{
  tvars {
    int outfd;
  }
  warn << "TCPconnect to " << hn << ":" << port << "\n";
  twait { tcpconnect (hn, port, mkevent (outfd)); }
  if (outfd >= 0) {
    twait {
      tame::proxy (infd, outfd, mkevent ());
      tame::proxy (outfd, infd, mkevent ());
    }
  } else {
    warn ("tcpconnect failed: %m\n");
  }
  (*cb) ();
}

tamed static void
listen_loop (int lport, str hn, int port, evv_t cb)
{
  tvars {
    int lfd;
    bool go (true);
    rendezvous_t<ev_t> G (__FILE__, __LINE__);
    int nfd;
    ev_t which;
    int proxies_out (0);
    sockaddr_in sin;
    socklen_t sinlen (sizeof (sin));
  }

  if ((lfd = listen_on (lport)) >= 0) {

    cb->set_cancel_notifier (mkevent (G, PROXY_END));

    while (go || proxies_out > 0 ) {

      if (go)
	tame::waitread (lfd, mkevent (G, READ));

      twait (G, which);

      switch (which) {
      case READ:
	{
	  bzero (&sin, sinlen);
	  if ((nfd = accept (lfd, reinterpret_cast<sockaddr *> (&sin), 
			     &sinlen)) >= 0) {
	    warn ("accepting connection from %s\n", inet_ntoa (sin.sin_addr));
	    proxies_out ++;
	    do_proxy (nfd, hn, port, mkevent (G, PROXY_DONE));
	  } else {
	    warn ("accept failed: %m\n");
	  }
	  break;
	}
      case PROXY_DONE:
	proxies_out --;
	break;
      case PROXY_END:
	warn << "Caught shutdown signal...\n";
	go = false;
	break;
      default:
	panic ("weird\n");
      }
    }

    warn << "Exiting listen loop....\n";
    G.cancel ();
  }
  (*cb) ();
}

tamed static void
main2 (int lport, str hn, int port)
{
  tvars {
    rendezvous_t<bool> rv (__FILE__, __LINE__);
    bool ok;
  }

  listen_loop (lport, hn, port, mkevent (rv, true));
  tame::sigcb1 (SIGINT, mkevent (rv, false));
  tame::sigcb1 (SIGTERM, mkevent (rv, false));

  twait (rv, ok);

  rv.cancel ();

  warn << "Waiting 2 second before exit...\n";
  twait { delaycb (2, 0, mkevent ()); }

  exit (0);
}


int 
main (int argc, char *argv[])
{
  static rxx x ("([^:]+):(\\d+)");
  int lport = 0;
  int port = 0;
  str hn;
  setprogname (argv[0]);

  if (argc == 3 && convertint (argv[1], &lport) && 
      x.match (argv[2]) && convertint (x[2], &port) &&
      (hn = x[1])) {
    main2 (lport, hn, port);
  } else {
    usage ();
  }
  amain ();
}

