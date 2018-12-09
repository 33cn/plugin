// -*-c++-*-
/* $Id: ex7.T 2693 2007-04-08 19:47:33Z max $ */

#include "tame.h"
#include "parseopt.h"
#include "tame_io.h"

typedef event<int,int>::ref cbii ;

/**
 * Given a vector of N hosts, connect to all of them on the given port.
 * When the first connection is established, return controle via 'done',
 * and close the remaining stragglers.
 *
 * @param hosts the hosts to try
 * @param port the port to try on
 * @param done the callback to call when the first has returned.
 */
tamed static void 
connect (vec<str> hosts, int port, cbii done)
{
  tvars {
    u_int i;
    rendezvous_t<u_int, ptr<int> > rv;
    bool got_one (false);
    ptr<int> fd;
  }

  for (i = 0; i < hosts.size (); i++) {
    fd = New refcounted<int> (-1);
    tcpconnect (hosts[i], port, mkevent (rv,i,fd,*fd));
  }
  
  while (rv.need_wait ()) {
    twait (rv, i, fd);
    warn << hosts[i]  << ":" << port << ": ";
    if (*fd >= 0) {
      warnx << "connection succeeded";
      if (!got_one) {
	done->trigger(*fd, i);
	got_one = true;
      } else {
	warnx << "... but too late!";
	close (*fd);
      }
      warnx << "\n";
    } else {
      warnx << "connection failed\n";
    }
  }
  if (!got_one) {
    done->trigger (-1, -1);
  }
}

/**
 * Get fastest Web page, where 'fast' is defined by how fast the DNS
 * lookup and TCP session establishment are.  Once connected, request
 * for "/" and dump the response to standard output.
 *
 * @param hosts the hosts to try
 * @param port the port to try on
 * @param done the callback to call when done
 */
tamed static void
get_fastest_web_page (vec<str> hosts, int port, cbb cb)
{
  tvars {
    int fd (-1), rc;
    strbuf req, resp;
    bool ret (true);
    int host;
  }

  //
  // get the fastest connection, and dump the result into 'fd'
  //
  twait { connect (hosts, port, mkevent(fd, host)); }
  if (fd < 0) {
    ret = false;
    goto done;
  }

  //
  // A dirt simple HTTP 1.0 request
  //
  req << "GET / HTTP/1.0\r\n"
         "Host: " << hosts[host] << "\r\n"
	 "\r\n";

  // suio::resid() returns the # of bytes left to write.
  while (req.tosuio ()->resid ()) {

    // Wait on fd to become writable, and get called back when so.
    // Unlike libasync's fdcb, waitwrite (and waitread) are not
    // "sticky" --- they're one-shots.
    twait { tame::waitwrite (fd, mkevent ()); }

    //
    // Use this syntax to output the results of the string buffer
    // 'req' to the socket 'fd'. At this point, fdcb has returned,
    // signalling that the socket 'fd' is writable.  If for some
    // reason we were lied to, write() will return <0 inside of
    // suio::output() below, but with errno set to EAGAIN; then 
    // suio::output() will return 0, and we'll try the write again the 
    // next time through the loop.  A return from suio::output() that
    // is negative signals a non-retryable error, and we'll bail out.
    //
    if (req.tosuio ()->output (fd) < 0) {
      warn << "write failed...\n";
      ret = false;
      goto done;
    }
  }

  //
  // The details of the reading loop are almost identical to the
  // writing loop above.
  //
  while (true) {
    twait { tame::waitread (fd, mkevent ()); }
    if ((rc = resp.tosuio ()->input (fd)) < 0 && errno != EAGAIN) {
      ret = false;
      goto done;
    }
    if (rc == 0) {
      break;
    }
  }

  // 
  // dump the response to standard output
  //
  resp.tosuio ()->output (1);

  // 
  // success!
  //
 done:
  if (fd >= 0)
    close (fd);

  (*cb) (ret);
}

static void finish (bool rc)
{
  delaycb (3, 0, wrap (exit, rc ? 0 : -1));
}

int
main (int argc, char *argv[])
{
  vec<str> hosts;
  int port;
  if (argc < 3 || !convertint (argv[1], &port))
    fatal << "usage: ex2 <port> <host1> <host2> ...\n";

  for (int i = 2; i < argc; i++) 
    hosts.push_back (argv[i]);

  get_fastest_web_page (hosts, port, wrap (finish));

  amain ();
}
