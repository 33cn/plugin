/* $Id: test_itree.C 2 2003-09-24 14:35:33Z max $ */

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

/* Note:  This test sucks.  Someone should write a real one. */

#define DEBUG_REDBLACK 1

#include "async.h"
#include "itree.h"

struct fubar {
  int key;
  int val;
  struct itree_entry<fubar> link;
};

itree<int, fubar, &fubar::key, &fubar::link> tt;

void
rbtest (void)
{
  const int iter = 25;
  const int mi = 17;
  struct fubar *fbp, *nfbp;
  struct fubar *fba[iter];
  int i;

  fba[0] = 0;
  for (i = 1; i < iter; i++) {
    fbp = New fubar;
    fbp->key = i;
    fbp->val = 0;
    tt.insert (fbp);
    fba[i] = fbp;
  }

  for (i = iter - 1; i > 0; i -= 3)
    tt.remove (fba[i]);

  for (i = 1; i < 10; i++) {
    fbp = New fubar;
    fbp->key = mi;
    fbp->val = i;
    tt.insert (fbp);
  }

  for (fbp = tt.first (); fbp; fbp = nfbp) {
    nfbp = tt.next (fbp);
    tt.remove (fbp);
  }

  for (int i = 0; i < iter; i++) {
    fbp = New fubar;
    fbp->key = 9999;
    fbp->val = i;
    tt.insert (fbp);
  }
  i = 0;
  for (fbp = tt[9999]; fbp; fbp = tt.next (fbp))
    i++;
  assert (i == iter);
}

int
main (int argc, char **argv)
{
  rbtest ();
  return 0;
}
