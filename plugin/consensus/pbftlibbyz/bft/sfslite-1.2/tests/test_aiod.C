/* $Id: test_aiod.C 2 2003-09-24 14:35:33Z max $ */

/*
 *
 * Copyright (C) 1999 David Mazieres (dm@uun.org)
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

#include "aiod.h"

int x = 32;
int fctr;

struct aiotst : public virtual refcount {
  aiod *a;
  str path;
  ptr<aiobuf> buf;

  aiotst (aiod *a) : a (a) { start (); }
  void start () {
    char tfile[80];
    sprintf (tfile, "aio.%02d~", fctr++);
    path = tfile;
    str msg ("This is " << path << ".\n");
    buf = a->bufalloc (msg.len () + 1);
    if (!buf) {
      a->bufwait (wrap (this, &aiotst::start));
      return;
    }
    strcpy (buf->base (), msg);
    a->open (path, O_CREAT|O_RDWR|O_TRUNC, 0666,
	     wrap (mkref (this), &aiotst::opencb));
    a->finalize ();
  }
  void opencb (ptr<aiofh> fh, int err) {
    if (!fh)
      panic ("open: %s\n", strerror (err));
    fh->write (0, buf, wrap (mkref (this), &aiotst::writecb));
  }
  void writecb (ptr<aiobuf>, ssize_t, int err) {
    if (err)
      panic ("write %s: %s\n", path.cstr (), strerror (err));
    a->unlink (path, wrap (mkref (this), &aiotst::unlinkcb));
  }
  void unlinkcb (int err) {
    if (err)
      panic ("unlink %s: %s\n", path.cstr (), strerror (err));
  }

  ~aiotst () {
    if (!--x)
      exit (0);
  }
};

int
main (int argc, char **argv)
{
  setprogname (argv[0]);
  char *dir = getcwd (NULL, PATH_MAX);
  str aiodpath (strbuf ("%s/../async/aiod", dir));
  free (dir);

  for (int i = x; i-- > 0;) {
    aiod *a = New aiod (1, 0x10000, 0x10000, false, aiodpath);
    New refcounted<aiotst> (a);
  }
  amain ();
}
