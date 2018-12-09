/* $Id: myname.C 3758 2008-11-13 00:36:00Z max $ */

/*
 *
 * Copyright (C) 1998 David Mazieres (dm@uun.org)
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

#include "dns.h"

str
mydomain ()
{
  if (!(_res.options & RES_INIT))
    res_init ();
  return _res.dnsrch[0];
}


#ifndef MAXHOSTNAMELEN
#define MAXHOSTNAMELEN 256
#endif /* !MAXHOSTNAMELEN */

str
myname ()
{
  char namebuf[MAXHOSTNAMELEN+1];
  namebuf[MAXHOSTNAMELEN] = '\0';
  if (gethostname (namebuf, MAXHOSTNAMELEN) < 0)
    panic ("gethostname: %m\n");

  if (strchr (namebuf, '.'))
    return namebuf;

  if (!(_res.options & RES_INIT))
    res_init ();
  if (_res.dnsrch[0] && _res.dnsrch[0][0])
    return strbuf ("%s.%s", namebuf, _res.dnsrch[0]);

  if (hostent *hp = gethostbyname (namebuf)) {
    if (strchr (hp->h_name, '.')) {
      return hp->h_name;
    } else {
      for (char **np = hp->h_aliases; *np; np++) {
	if (strchr (*np, '.')) {
	  return *np;
	}
      }
    }
  }

  vec<in_addr> av;
  if (myipaddrs (&av)) {
    for (in_addr *ap = av.base (); ap < av.lim (); ap++) {
      if (ap->s_addr != htonl (INADDR_LOOPBACK)) {
	if (hostent *hp = gethostbyaddr ((char *) ap, sizeof (*ap), AF_INET)) {
	  if (strchr (hp->h_name, '.')) {
	    return hp->h_name;
	  } else {
	    for (char **np = hp->h_aliases; *np; np++) {
	      if (strchr (*np, '.')) {
		return *np;
	      }
	    }
	  }
	}
      }
    }
  }

  warn ("cannot find fully qualified domain name of this host\n");
  warn ("set system name to fully-qualified domain name "
	"or fix /etc/resolv.conf\n");
  return NULL;
}
