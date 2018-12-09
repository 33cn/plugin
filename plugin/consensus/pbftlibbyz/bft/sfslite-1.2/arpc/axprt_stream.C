/* $Id: axprt_stream.C 2603 2007-03-23 16:08:56Z max $ */

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

#include "arpc.h"


axprt_stream::axprt_stream (int f, size_t ps, size_t bs)
  : axprt_pipe(f, f, ps, bs)
{
}

int
axprt_stream::reclaim ()
{
  int r,w;
  axprt_pipe::reclaim (&r, &w);
  assert (r == w);
  return r;
}
