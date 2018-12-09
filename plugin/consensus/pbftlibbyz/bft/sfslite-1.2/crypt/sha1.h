// -*-c++-*-
/* $Id: sha1.h 1117 2005-11-01 16:20:39Z max $ */

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


#ifndef _SHA1_H_
#define _SHA1_H_

#include "crypthash.h"

class sha1 : public mdblock {
public:
  enum { hashsize = 20 };
  enum { hashwords = hashsize/4 };

  void finish () { mdblock::finish_be (); }

  static void newstate (u_int32_t state[hashwords]);
  static void transform (u_int32_t[hashwords], const u_char[blocksize]);
  static void state2bytes (void *, const u_int32_t[hashwords]);
};

class sha1ctx : public sha1 {
protected:
  u_int32_t state[hashwords];

  void consume (const u_char *p) { transform (state, p); }
public:
  sha1ctx () { newstate (state); }
  void reset () { count = 0; newstate (state); }
  void final (void *digest) {
    finish ();
    state2bytes (digest, state);
    bzero (state, sizeof (state));
  }
};

/* The following undefined function is to catch stupid errors */
extern void sha1_hash (u_char *digest, const iovec *, size_t);

inline void
sha1_hash (void *digest, const void *buf, size_t len)
{
  sha1ctx sc;
  sc.update (buf, len);
  sc.final (digest);
}

inline void
sha1_hashv (void *digest, const iovec *iov, u_int cnt)
{
  sha1ctx sc;
  sc.updatev (iov, cnt);
  sc.final (digest);
}

#ifdef _ARPC_XDRMISC_H_
template<class T> bool
sha1_hashxdr (void *digest, const T &t, bool scrub = false)
{
  xdrsuio x (XDR_ENCODE, scrub);
  XDR *xp = &x;
  if (!rpc_traverse (xp, const_cast<T &> (t)))
    return false;
  sha1_hashv (digest, x.iov (), x.iovcnt ());
  return true;
}
#endif /* _ARPC_XDRMISC_H_ */

class sha1hmac : public sha1ctx {
  u_int32_t istate[hashwords];
  u_int32_t ostate[hashwords];
public:
  sha1hmac () {}		// Warning, no sanity check, must call setkey
  sha1hmac (const void *k, size_t klen) { setkey (k, klen); }
  void setkey (const void *, size_t);
  // void setkey2 (const void *, size_t, const void *, size_t);
  void reset () { count = blocksize; memcpy (state, istate, sizeof (state)); }
  void final (void *digest);
};

inline void
sha1_hmac (void *out, const void *key, size_t keylen,
	   const void *msg, size_t msglen)
{
  sha1hmac hc (key, keylen);
  hc.update (msg, msglen);
  hc.final (out);
}

#ifdef _ARPC_XDRMISC_H_
template<class T> bool
sha1_hmacxdr (void *digest, const void *k1, size_t k1l,
	      const T &t, bool scrub = false)
{
  xdrsuio x (XDR_ENCODE, scrub);
  XDR *xp = &x;
  if (!rpc_traverse (xp, const_cast<T &> (t)))
    return false;

  sha1hmac hc;
  hc.setkey (k1, k1l);
  hc.updatev (x.iov (), x.iovcnt ());
  hc.final (digest);

  if (scrub)
    hc.setkey (NULL, 0);

  return true;
}

template<class T> bool
sha1_hmacxdr_2 (void *digest, const void *k1, size_t k1l,
		const void *k2, size_t k2l,
		const T &t, bool scrub = false)
{
  xdrsuio x (XDR_ENCODE, scrub);
  XDR *xp = &x;
  if (!rpc_traverse (xp, const_cast<T &> (t)))
    return false;

  u_char *kbuf = static_cast<u_char *> (xmalloc (k1l + k2l));
  memcpy (kbuf, k1, k1l);
  memcpy (kbuf + k1l, k2, k2l);
  sha1hmac hc;
  hc.setkey (kbuf, k1l + k2l);
  bzero (kbuf, k1l + k2l);
  xfree (kbuf);

  hc.updatev (x.iov (), x.iovcnt ());
  hc.final (digest);

  if (scrub)
    hc.setkey (NULL, 0);

  return true;
}
#endif /* _ARPC_XDRMISC_H_ */

class sha1oracle : public sha1 {
  const size_t hashused;
  const size_t nctx;
  u_int32_t (*state)[hashwords];
  bool firstblock;

  void consume (const u_char *p);
public:
  const u_int64_t idx;
  const size_t resultsize;

  sha1oracle (size_t nbytes, u_int64_t idx = 0, size_t hashused = hashsize);
  ~sha1oracle ();
  void reset ();
  void final (void *);
};

#endif /* !_SHA1_H_ */
