/* $Id: myaddrs.C 1117 2005-11-01 16:20:39Z max $ */

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


#include "amisc.h"
#include "qhash.h"

#include <net/if.h>
#ifdef HAVE_SYS_SOCKIO_H
#include <sys/sockio.h>
#endif /* HAVE_SYS_SOCKIO_H */

#define MAXCONFBUF 0x10000

bool
myipaddrs (vec<in_addr> *res)
{
  static u_int len = 512;
  struct ifconf ifc;
  char *p, *e;
  int s;

  if ((s = socket (AF_INET, SOCK_DGRAM, 0)) < 0) {
    warn ("socket: %m\n");
    return false;
  }

  errno = 0;
  ifc.ifc_len = MAXCONFBUF;
  ifc.ifc_buf = NULL;
  for (;;) {
    ifc.ifc_len = len;
    xfree (ifc.ifc_buf);
    ifc.ifc_buf = (char *) xmalloc (ifc.ifc_len);
    if (ioctl (s, SIOCGIFCONF, &ifc) < 0) {
      warn ("SIOCGIFCONF: %m\n");
      close (s);
      xfree (ifc.ifc_buf);
      return false;
    }
    /* The +64 is for large addresses (e.g., IPv6), in which sa_len
     * may be greater than the struct sockaddr inside ifreq. */
    if (ifc.ifc_len + sizeof (struct ifreq) + 64 < len)
      break;
    len *= 2;
    if (len >= MAXCONFBUF) {
      warn ("SIOCGIFCONF: buffer too large\n");
      close (s);
      xfree (ifc.ifc_buf);
      return false;
    }
  }

  res->clear ();
  bhash<in_addr> addrs;

  p = ifc.ifc_buf;
  e = p + ifc.ifc_len;
  while (p < e) {
    struct ifreq *ifrp = (struct ifreq *) p;
    struct ifreq ifr = *ifrp;
#ifndef HAVE_SA_LEN
    p += sizeof (ifr);
#else /* !HAVE_SA_LEN */
    p += sizeof (ifrp->ifr_name)
      + max (sizeof (ifrp->ifr_addr), (size_t) ifrp->ifr_addr.sa_len);
#endif /* !HAVE_SA_LEN */
    if (ifrp->ifr_addr.sa_family != AF_INET)
      continue;
    if (ioctl (s, SIOCGIFFLAGS, &ifr) < 0) {
      warn ("SIOCGIFFLAGS (%.*s): %m\n", (int) sizeof (ifr.ifr_name),
	    ifr.ifr_name);
      continue;
    }
    in_addr a = ((struct sockaddr_in *) &ifrp->ifr_addr)->sin_addr;
    if ((ifr.ifr_flags & IFF_UP) && !addrs[a]) {
      addrs.insert (a);
      res->push_back (a);
    }
  }

  xfree (ifc.ifc_buf);
  close (s);

  return true;
}
