/* $Id: itree.C 1117 2005-11-01 16:20:39Z max $ */

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


#include "itree.h"

#define oc __opaquecontainer_pointer
#define itree_entry __itree_entry_private

/* Define some gross macros to avoid a lot of typing.  These asssume a
 * local variable 'os' contains the offset of the itree_entry
 * structure in the opaque container structure. */
static inline struct itree_entry *oc2rb (oc, const int)
     __attribute__ ((const));
static inline struct itree_entry *
oc2rb (oc o, const int os)
{
  return (struct itree_entry *) ((char *) o + os);
}
#define up(n) oc2rb (n, os)->rbe_up
#define left(n) oc2rb (n, os)->rbe_left
#define right(n) oc2rb (n, os)->rbe_right
#define color(n) ((n) ? oc2rb (n, os)->rbe_color : BLACK)
#define lcolor(n) oc2rb (n, os)->rbe_color

static void
itree_left_rotate (oc *r, oc x, const int os)
{
  oc y = right (x);
  if ((right (x) = left (y)))
    up (left (y)) = x;
  up (y) = up (x);
  if (!up (x))
    *r = y;
  else if (x == left (up (x)))
    left (up (x)) = y;
  else
    right (up (x)) = y;
  left (y) = x;
  up (x) = y;
}

static void
itree_right_rotate (oc *r, oc y, const int os)
{
  oc x = left (y);
  if ((left (y) = right (x)))
    up (right (x)) = y;
  up (x) = up (y);
  if (!up (y))
    *r = x;
  else if (y == right (up (y)))
    right (up (y)) = x;
  else
    left (up (y)) = x;
  right (x) = y;
  up (y) = x;
}

/* Turn any newly inserted node red.  Then fix things up if the new
 * node's parent is also red.  Note that this routine assumes the root
 * of the tree is black. */
static inline void
itree_insert_fix (oc *r, oc x, const int os)
{
  oc y;

  lcolor (x) = RED;
  while (color (up (x)) == RED) {
    oc gp = up (up (x));
    if (up (x) == left (gp)) {
      y = right (gp);
      if (color (y) == RED) {
	lcolor (up (x)) = BLACK;
	lcolor (y) = BLACK;
	lcolor (gp) = RED;
	x = gp;
      }
      else {
	if (x == right (up (x))) {
	  x = up (x);
	  itree_left_rotate (r, x, os);
	}
	lcolor (up (x)) = BLACK;
	lcolor (up (up (x))) = RED;
	itree_right_rotate (r, up (up (x)), os);
      }
    }
    else {
      y = left (gp);
      if (color (y) == RED) {
	lcolor (up (x)) = BLACK;
	lcolor (y) = BLACK;
	lcolor (gp) = RED;
	x = gp;
      }
      else {
	if (x == left (up (x))) {
	  x = up (x);
	  itree_right_rotate (r, x, os);
	}
	lcolor (up (x)) = BLACK;
	lcolor (up (up (x))) = RED;
	itree_left_rotate (r, up (up (x)), os);
      }
    }
  }
  lcolor (*r) = BLACK;
}

void
itree_insert (oc *r, oc z, const int os,
	      int (*cmpfn) (void *, oc, oc), void *cmparg)
{
  oc y, x;
  int cmpres = 0;

  left (z) = right (z) = y = NULL;
  for (x = *r; x; x = cmpres < 0 ? left (x) : right (x)) {
    y = x;
    cmpres = cmpfn (cmparg, z, x);
  }
  up (z) = y;
  if (!y)
    *r = z;
  else if (cmpres < 0)
    left (y) = z;
  else
    right (y) = z;
  itree_insert_fix (r, z, os);
}

static inline oc
itree_minimum (oc x, const int os)
{
  oc y;
  while ((y = left (x)))
    x = y;
  return x;
}

oc
itree_successor (oc x, const int os)
{
  oc y;
  if ((y = right (x)))
    return itree_minimum (y, os);
  for (y = up (x); y && x == right (y); y = up (y))
    x = y;
  return y;
}

static inline oc
itree_maximum (oc x, const int os)
{
  oc y;
  while ((y = right (x)))
    x = y;
  return x;
}

oc
itree_predecessor (oc x, const int os)
{
  oc y;
  if ((y = left (x)))
    return itree_maximum (y, os);
  for (y = up (x); y && x == left (y); y = up (y))
    x = y;
  return y;
}

static inline void
itree_delete_fixup (oc *r, oc x, oc p, const int os)
{
  oc w;

  assert (!x || up (x) == p);
  while (x != *r && color (x) == BLACK) {
    if (x)
      p = up (x);
    if (x == left (p)) {
      w = right (p);
      if (color (w) == RED) {
	lcolor (w) = BLACK;
	lcolor (p) = RED;
	itree_left_rotate (r, p, os);
	w = right (p);
      }
      if (color (left (w)) == BLACK && color (right (w)) == BLACK) {
	lcolor (w) = RED;
	x = p;
      }
      else {
	if (color (right (w)) == BLACK) {
	  lcolor (left (w)) = BLACK;
	  lcolor (w) = RED;
	  itree_right_rotate (r, w, os);
	  w = right (p);
	}
	lcolor (w) = color (p);
	lcolor (p) = BLACK;
	lcolor (right (w)) = BLACK;
	itree_left_rotate (r, p, os);
	x = *r;
      }
    }
    else {
      w = left (p);
      if (color (w) == RED) {
	lcolor (w) = BLACK;
	lcolor (p) = RED;
	itree_right_rotate (r, p, os);
	w = left (p);
      }
      if (color (right (w)) == BLACK && color (left (w)) == BLACK) {
	lcolor (w) = RED;
	x = p;
      }
      else {
	if (color (left (w)) == BLACK) {
	  lcolor (right (w)) = BLACK;
	  lcolor (w) = RED;
	  itree_left_rotate (r, w, os);
	  w = left (p);
	}
	lcolor (w) = color (p);
	lcolor (p) = BLACK;
	lcolor (left (w)) = BLACK;
	itree_right_rotate (r, p, os);
	x = *r;
      }
    }
  }
  if (x)
    lcolor (x) = BLACK;
}

void
itree_delete (oc *r, oc z, const int os)
{
  oc x, y, p;
  enum itree_color c;

  y = left (z) && right (z) ? itree_successor (z, os) : z;
  p = up (y);
  if ((x = left (y)) || (x = right (y)))
    up (x) = p;
  if (!p)
    *r = x;
  else if (y == left (p))
    left (p) = x;
  else
    right (p) = x;
  c = color (y);

  if (y != z) {
    oc pz = up (z);
    if (pz) {
      if (z == left (pz))
	left (pz) = y;
      else
	right (pz) = y;
    }
    else
      *r = y;
    if ((pz = left (z)))
      up (pz) = y;
    if ((pz = right (z)))
      up (pz) = y;
    *oc2rb (y, os) = *oc2rb (z, os);
    if (p == z)
      p = y;
  }

  if (c == BLACK)
    itree_delete_fixup (r, x, p, os);
}

static void
itree_check_node (oc x, oc low, oc high, int bd, const int lbd,
		  const int os,
		  int (*cmpfn) (void *, oc, oc), void *cmparg)
{
  volatile oc l, r, p;
  volatile enum itree_color cx, cl, cr, cp;

  if (color (x) == BLACK)
    bd++;
  if (!x) {
    assert (bd == lbd);
    return;
  }

  cx = color (x);
  p = up (x);
  cp = color (p);
  l = left (x);
  cl = color (l);
  r = right (x);
  cr = color (r);

  assert (!l || up (l) == x);
  assert (!r || up (r) == x);
  assert (cx == BLACK || cx == RED);
  assert (cx == BLACK || (cl == BLACK && cr == BLACK));
  assert (!low || cmpfn (cmparg, low, x) <= 0);
  assert (!high || cmpfn (cmparg, x, high) <= 0);

  itree_check_node (l, low, x, bd, lbd, os, cmpfn, cmparg);
  itree_check_node (r, x, high, bd, lbd, os, cmpfn, cmparg);
}

void
__itree_check (oc *r, const int os,
	       int (*cmpfn) (void *, oc, oc), void *cmparg)
{
  oc x;
  int lbd = 0;

  assert (color (*r) == BLACK);
  for (x = *r; x;) {
    x = left (x);
    if (color (x) == BLACK)
      lbd++;
  }

  assert (!*r || !up (*r));
  itree_check_node (*r, NULL, NULL, -1, lbd, os, cmpfn, cmparg);
}
