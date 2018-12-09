// -*-c++-*-
/* $Id: crypthash.h 1117 2005-11-01 16:20:39Z max $ */

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


#ifndef _MDBLOCK_H_
#define _MDBLOCK_H_ 1

#include "sysconf.h"

struct datasink {
  virtual void update (const void *, size_t) = 0;
  virtual ~datasink () {}
};

class mdblock : public datasink {
  /* The following undefined function is to catch stupid errors.  A
   * call to update with an iovec * is almost certainly a typo that
   * should have read updatev. */
  void update (const iovec *, size_t);

public:
  enum { blocksize = 64 };

protected:
  u_int64_t count;
  u_char buffer[blocksize];

  mdblock () : count (0) {}
  virtual ~mdblock () { count = 0; bzero (buffer, sizeof (buffer)); }

  void finish_le ();
  void finish_be ();
  virtual void consume (const u_char *) = 0;

public:
  void update (const void *data, size_t len);
  void updatev (const iovec *iov, u_int cnt);
};

#ifdef _ARPC_XDRMISC_H_
template<class T> bool
datasink_catxdr (datasink &dst, const T &t, bool scrub = false)
{
  xdrsuio x (XDR_ENCODE, scrub);
  XDR *xp = &x;
  if (!rpc_traverse (xp, const_cast<T &> (t)))
    return false;
  for (const iovec *iov = x.iov (), *end = iov + x.iovcnt (); iov < end; iov++)
    dst.update (iov->iov_base, iov->iov_len);
  return true;
}
#endif /* _ARPC_XDRMISC_H_ */

#endif /* !_MDBLOCK_H_ */
