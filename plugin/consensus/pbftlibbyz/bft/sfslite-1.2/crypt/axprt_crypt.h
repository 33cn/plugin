// -*-c++-*-
/* $Id: axprt_crypt.h 1117 2005-11-01 16:20:39Z max $ */

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

#ifndef _AXPRT_CRYPT_H_
#define _AXPRT_CRYPT_H_ 1

#include "arpc.h"
#include "arc4.h"

class axprt_crypt : public axprt_stream {
  enum { macsize = 16 };

  arc4 ctx_send;
  arc4 ctx_recv;

  bool cryptsend;
  bool cryptrecv;
  bool macset;

  u_char mackey1[16];
  u_char mackey2[16];
  u_int32_t lenpad;

protected:
  axprt_crypt (int f, size_t ps)
    : axprt_stream (f, ps, ps + macsize + 4),
      cryptsend (false), cryptrecv (false), macset (false)
    {}
  virtual ~axprt_crypt ();
  virtual bool getpkt (char **, char *);
  virtual void recvbreak ();

public:
  virtual void sendv (const iovec *, int, const sockaddr * = NULL);
  void encrypt (const void *sendkey, size_t sendkeylen,
		const void *recvkey, size_t recvkeylen);
  void encrypt (const str &sendkey, const str &recvkey)
    { encrypt (sendkey, sendkey.len (), recvkey, recvkey.len ()); }

  static ref<axprt_crypt> alloc (int, size_t = axprt_stream::defps);
};

extern const axprtalloc_fn axprt_crypt_alloc;

#endif /* !_AXPRT_CRYPT_H_ */
