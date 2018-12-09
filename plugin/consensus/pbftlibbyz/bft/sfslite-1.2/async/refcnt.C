/* $Id: refcnt.C 3769 2008-11-13 20:21:34Z max $ */

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

#include "refcnt.h"
#include "amisc.h"

/*
 * XXX - workaround for egregious egcs bug.
 *
 * On OpenBSD, egcs does not support inline functions with static
 * variables properly.  For example, what we want is:
 *    class __globaldestruction_t {
 *      bool &started () { static bool val; return val; }
 *    public:
 *      ~__globaldestruction_t () { started () = true; }
 *      operator bool () { return started () }
 *    };
 * But that doesn't work, so we need a whole file just for one bit!
 */

bool __globaldestruction_t::started;

/*
 * XXX - for lack of a better place to put this:
 */

#include "callback.h"


static void
ignore_void ()
{
}

static void
ignore_int (int)
{
}

callback<void>::ref cbv_null (gwrap (ignore_void));
callback<void, int>::ref cbi_null (gwrap (ignore_int));

#include "err.h"
#include <typeinfo>

void
refcnt_warn (const char *op, const std::type_info &type, void *addr, int cnt)
{
  char buf[1024];
  sprintf (buf, "%.128s%s%.64s: %.512s (%p) -> %d\n",
	   progname ? progname.cstr () : "", progname ? ": " : "",
	   op, type.name (), addr, cnt);
  assert (memchr (buf, 0, sizeof (buf)));
  v_write (errfd, buf, strlen (buf));
}
