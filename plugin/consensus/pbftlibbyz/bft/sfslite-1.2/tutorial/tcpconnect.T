// -*-c++-*-
/* $Id: tcpconnect.T 2693 2007-04-08 19:47:33Z max $ */

/*
 *
 * Copyright (C) 2003 David Mazieres (dm@uun.org)
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU General Public License as
 * published by the Free Software Foundation; either version 2, or (at
 * your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 59 Temple Place, Suite 330, Boston, MA 02111-1307
 * USA
 *
 */

#include "async.h"
#include "dns.h"
#include "parseopt.h"
#include "tame.h"
#include "tame_io.h"
#include "tame_connectors.h"

/**
 * Given a UNIX in_addr structure and a port, make a new TCP connect
 * and test that it succeeded by selecting for write on the new socket.
 */
tamed static void
connect_to_in_addr (in_addr a, int port, evi_t ev)
{
  tvars {
    sockaddr_in sin;
    socklen_t sn;
    int err;
    int fd (-1);
    outcome_t outc;
  }

  bzero (&sin, sizeof (sin));
  sin.sin_family = AF_INET;
  sin.sin_port = htons (port);
  sin.sin_addr = a;

  fd = inetsocket (SOCK_STREAM);
  if (fd >= 0) {
    make_async (fd);
    close_on_exec (fd);
    if (connect (fd, (sockaddr *) &sin, sizeof (sin)) >= 0 || 
	errno == EINPROGRESS) {

      twait { tame::waitwrite (fd, connector::cnc (mkevent (), ev, &outc)); }
      warn << "return from function\n";

      if (outc == OUTCOME_SUCC) {
	sn = sizeof (sin);
	if (getpeername (fd, (sockaddr *) &sin, &sn)) {
	  err = 0;
	  sn = sizeof (err);
	  getsockopt (fd, SOL_SOCKET, SO_ERROR, (char *) &err, &sn);
	  if (err)
	    fd = -ECONNREFUSED;
	}
      } else {
	tame::clearwrite (fd);
      }
    } else {
      fd = -errno;
    }
  }

  // calls (*callercv)(fd); also enforces that this
  // callback is called exactly once.
  ev->trigger (fd);
}

/**
 * Given a hostname and port, make a TCP connection to the remote host,
 * returning the result as an opened file descriptor, or negative for
 * error.
 *
 * @param hostname
 * @param port
 * @param cv tha caller's CV
 * @param dnssearch whether to search DNS domains
 * @param namep the name of this host as gotten from DNS
 */
tamed static void
my_tcpconnect (str hostname, u_int16_t port, evi_t ev,
	       bool dnssearch, str *namep)
{
  tvars {
    dnsreq_t *dnsp (NULL);
    ptr<hostent> h;
    int err;
    int ret (-1);
    outcome_t outc;
  }

  // launch 2 parallel calls that will race: the DNS requester
  // and the canceller.
  twait { 
    dnsp = dns_hostbyname (hostname, 
			   connector::cnc (mkevent (h, err), ev, &outc), 
			   dnssearch); 
  }

  if (outc != OUTCOME_SUCC) {
    warn << "DNS lookup cancelled!\n";
    dnsreq_cancel (dnsp);
  } else {
    // if we keep going, note that the canceller is still outstanding,
    // and might still cancel us.
    dnsp = NULL;
    if (!h) {
      ret = dns_tmperr (err) ? -EAGAIN : -ENOENT;
    } else {
      if (namep)
	*namep = h->h_name;

      twait { 
	connect_to_in_addr (*(in_addr *) h->h_addr, 
			    port, 
			    connector::cnc (mkevent (ret), ev));
      }
    }
  }
  ev->trigger (ret);
}

static void usage ()
{
  fatal << "usage: " << progname << " <hostname> <port>\n";
}

tamed static void 
run (str s, int p, evv_t ev)
{
  tvars {
    rendezvous_t<bool> rv (__FILE__, __LINE__);
    int fd;
    bool succeeded;
  }
  warn << "connect on " << s << ":" << p << "\n";
  my_tcpconnect (s, p, mkevent (rv, true, fd), false, NULL);
  delaycb (4, 0, mkevent (rv, false));
  twait (rv, succeeded);
  if (succeeded) {
    warn << "Succeeded; fd=" << fd << "\n";
  } else {
    warn << "Had to cancel!\n";
  }
  rv.cancel ();
  ev->trigger ();
}

tamed static void
main2 (int argc, char **argv)
{
  tvars {
    int port;
  }
  if (argc != 3 || !convertint (argv[2], &port))
    usage ();
  twait { run (argv[1], port, mkevent ()); }
  twait { delaycb (10, 0, mkevent()); }
  exit (0);
}

int
main (int argc, char *argv[])
{
  setprogname (argv[0]);
  main2 (argc, argv);
  amain ();
}
