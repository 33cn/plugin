/* $Id: sfshostalias.C 1754 2006-05-19 20:59:19Z max $ */

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

#include "sfsconnect.h"
#include "parseopt.h"

#define SFSHOSTALIAS_VERBOSE 0

sfs_hosttab_t sfs_hosttab;

sfs_host::sfs_host (const str &h, const sfs_sa &s, socklen_t sl)
  : host (h), sa (s), sa_len (sl)
{
#if SFSHOSTALIAS_VERBOSE
  warn ("initialized host %s => address %s%%%d\n",
	host.cstr (), inet_ntoa (sa.sa_in.sin_addr),
	ntohs (sa.sa_in.sin_port));
#endif /* SFSHOSTALIAS_VERBOSE */
}

void
sfs_hosttab_t::errmsg (int line, str msg)
{
  warn << filename << ":" << line << ": " << msg << "\n";
}

void
sfs_hosttab_t::loadfd (int fd)
{
  suio uio;
  int lineno = 0;
  sfs_sa sa;
  bzero (&sa, sizeof (sa));
  sa.sa_in.sin_family = AF_INET;

  for (;;) {
    lineno++;
    size_t n;
    while (!(n = uio.linelen ()))
      if (uio.input (fd) <= 0) {
	if (uio.resid ())
	  errmsg (lineno, "incomplete last line");
	return;
      }
    mstr m (n);
    uio.copyout (m, n - 1);
    m[n-1] = '\0';
    uio.rembytes (n);
    if (m[0] == '#' || m[0] == '\n')
      continue;

    char *p = m;
    char *host = strnnsep (&p, " \t\r\n");
    char *port = host;
    xstrsep (&port, "%");

    if (inet_aton (host, &sa.sa_in.sin_addr) <= 0) {
      errmsg (lineno, "invalid IP address");
      continue;
    }
    if (!port || !*port)
      sa.sa_in.sin_port = htons (SFS_PORT);
    else if (convertint (port, &sa.sa_in.sin_port))
      sa.sa_in.sin_port = htons (sa.sa_in.sin_port);
    else {
      errmsg (lineno, "invalid port number");
      continue;
    }

    while ((host = strnnsep (&p, " \t\r\n"))) {
      str h (mytolower (host));
      if (sfs_host *hp = tab[h])
	delete hp;
      tab.insert (New sfs_host (h, sa, sizeof (sa.sa_in)));
    }
  }
}

bool
sfs_hosttab_t::load (const char *path)
{
  loaded = true;

  int fd = open (path, O_RDONLY);
  if (fd < 0) {
    if (errno != ENOENT)
      warn ("%s: %m\n", path);
    return false;
  }

  filename = path;
  loadfd (fd);
  close (fd);
  return true;
}

const sfs_host *
sfs_hosttab_t::lookup (str name) const
{
  const sfs_host *hp = tab[mytolower (name)];
#if SFSHOSTALIAS_VERBOSE
  warn ("lookup %s => %s\n", name.cstr (),
	hp ? inet_ntoa (hp->sa.sa_in.sin_addr) : "NONE");
#endif /* SFSHOSTALIAS_VERBOSE */
  return hp;
}

inline void
sfs_hosttab_init_dir (const char *s)
{
  if (s) {
    str path (strbuf ("%s/sfs_hosts", s));
    sfs_hosttab.load (path);
  }
}
void
sfs_hosttab_init ()
{
  if (!sfs_hosttab.loaded) {
    sfs_hosttab_init_dir (etc3dir);
    sfs_hosttab_init_dir (etc2dir);
    sfs_hosttab_init_dir (etc1dir);
  }
}
