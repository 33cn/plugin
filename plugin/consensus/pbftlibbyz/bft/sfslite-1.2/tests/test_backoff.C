/* $Id: test_backoff.C 2 2003-09-24 14:35:33Z max $ */

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

#include "async.h"
#include "backoff.h"

struct tmqtst {
  int n;
  tmoq_entry<tmqtst> tlink;
  tmqtst () : n (0) {}
  void xmit (int) { n++; alarm (0); sleep (3); alarm (8); }
  void timeout () { assert (n == 3); exit (0); }
};

static tmoq<tmqtst, &tmqtst::tlink, 1, 3> q;

void
stuck ()
{
  panic ("stuck\n");
}

int
main (int argc, char **argv)
{
  sigcb (SIGALRM, wrap (stuck));
  alarm (8);
  q.start (New tmqtst);
  amain ();
}
