// -*-c++-*-
/* $Id: tame_core.h 2654 2007-03-31 05:42:21Z max $ */

/*
 *
 * Copyright (C) 2005 Max Krohn (max@okws.org)
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

#include "tame_run.h"
#include "tame_event.h"
#include "tame_closure.h"
#include "qhash.h"

void report_leaks (event_cancel_list_t *lst)
{

  qhash<str, int> tab;
  vec<str> v;
  _event_cancel_base *p;
  for (p = lst->first ; p ; p = lst->next (p)) {
    strbuf b;
    str t = p->loc ();
    b << t << ": event object leaked";
    str s = b;
    int *n = tab[s];
    if (n) { (*n)++; }
    else { 
      tab.insert (s, 1); 
      v.push_back (s);
    }
  }

  for (size_t i = 0; i < v.size (); i++) {
    if (!(tame_options & TAME_ERROR_SILENT)) {
      str s = v[i];
      warn << s;
      if (*tab[s] > 1) 
	warnx << " (" << *tab[s] << " times)";
      warnx << "\n";
    }
  }

  if (v.size () > 0 && (tame_options & TAME_ERROR_FATAL))
    panic ("abort on TAME failure\n");
}

vec<weakref<rendezvous_base_t> > tame_collect_rv_vec;
bool tame_collect_rv_flag;

void
start_rendezvous_collection ()
{
  tame_collect_rv_flag = true;
  tame_collect_rv_vec.clear ();
}

void
collect_rendezvous (weakref<rendezvous_base_t> r)
{
  if (tame_collect_rv_flag) 
    tame_collect_rv_vec.push_back (r);
}

void
closure_t::collect_rendezvous ()
{
  for (u_int i = 0; i < tame_collect_rv_vec.size (); i++) {
    const weakref<rendezvous_base_t> &rv = tame_collect_rv_vec[i];
    if (is_onstack (rv.pointer ()))
      _rvs.push_back (rv);
  }
  tame_collect_rv_flag = false;
  tame_collect_rv_vec.clear ();
}
