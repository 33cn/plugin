// -*-c++-*-
/* $Id: axprt_crypt.C 2531 2007-02-11 14:40:18Z max $ */

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

/*
 * This is an encrypting stream transport.  Encryption starts when the
 * axprt_crypt::encrypt method is called with two encryption keys.
 * One key is used for encrypting sent data, the other for decrypting
 * received data.  Because security relies on the arc4 stream cipher,
 * THE SEND AND RECEIVE KEYS MUST BE UNIQUE AND UNRELATED for the
 * transport to offer any security.  The encryption and MAC can be
 * trivially broken if a key is ever reused.
 *
 * Each key is used to initialize an arc4 stream cipher.  To encrypt a
 * packet of length n (where n is divisible by 4), the transport
 * performs the following calculations:
 * 
 * Let A[0] be the next byte of the arc4 stream, A[1] following, etc.
 * Let M[0] .. M[n - 1] be the message to encrypt
 * Let R[0] .. R[n + 19] be the encryption result transmitted on the wire
 * Let SHA-1/16 designate the first 16 bytes of a SHA-1 hash
 * 
 * u_char mackey1[16] := A[0] .. A[15]
 * u_char mackey2[16] := A[16] .. A[31]
 * 
 * P[0] .. P[3] := htonl (0x80000000|n)
 * P[4] .. P[3+n] := M[0] .. M[n - 1]
 * P[4+n] .. P[4+n+15] := SHA-1/16 (mackey1, P[0] .. p[3+n], mackey2)
 * 
 * for i := 0; i <= n + 19; i++
 *   R[i] := P[i] ^ A[32 + i]
 *
 * In other words, the the first 32 bytes of arc4 data get used to
 * compute a 16-byte MAC on the message length and contents.  Then
 * then entire packet including length, message contents, and MAC are
 * encrypted by XORing them with subsequent bytes from the arc4
 * stream.
 */

#include "axprt_crypt.h"
#include "sha1.h"
#include "serial.h"

ptr<axprt_stream> axprt_crypt_alloc_fn (size_t ps, int fd);
const axprtalloc_fn axprt_crypt_alloc
  = gwrap (axprt_crypt_alloc_fn, int (axprt_stream::defps));

ptr<axprt_stream>
axprt_crypt_alloc_fn (size_t ps, int fd)
{
  return axprt_crypt::alloc (fd, ps);
}

axprt_crypt::~axprt_crypt ()
{
  ctx_send.reset ();
  ctx_recv.reset ();
  lenpad = 0;
}

void
axprt_crypt::recvbreak ()
{
  fail ();
}

bool
axprt_crypt::getpkt (char **cpp, char *eom)
{
  if (!cryptrecv)
    return axprt_stream::getpkt (cpp, eom);

  if (!macset) {
    for (size_t i = 0; i < sizeof (mackey1); i++)
      mackey1[i] = ctx_recv.getbyte ();
    for (size_t i = 0; i < sizeof (mackey2); i++)
      mackey2[i] = ctx_recv.getbyte ();
    lenpad = (ctx_recv.getbyte () << 24 | ctx_recv.getbyte () << 16
	      | ctx_recv.getbyte () << 8 | ctx_recv.getbyte ());
    macset = true;
  }

  char *cp = *cpp;
  if (!cb || eom - cp < 4)
    return false;

  const u_char *ucp = reinterpret_cast <u_char *> (cp);
  int32_t len = getint (ucp) ^ lenpad;
  u_int32_t rawlen = htonl (len);
  cp += 4;

  if (!len) {
    *cpp = cp;
    recvbreak ();
    return true;
  }
  if (!checklen (&len))
    return false;

  char *pktlim = cp + len + macsize;
  if (pktlim > eom)
    return false;

  macset = false;
  for (char *p = cp; p < pktlim; p++)
    *p ^= ctx_recv.getbyte ();

  sha1ctx sc;
  sc.update (mackey1, sizeof (mackey1));
  sc.update (&rawlen, 4);
  sc.update (cp, len);
  sc.update (mackey2, sizeof (mackey2));

  u_char mac[sha1::hashsize];
  sc.final (mac);
  if (memcmp (mac, cp + len, macsize)) {
    warn ("axprt_crypt::getpkt: MAC failure\n");
    fail ();
    return false;
  }

  *cpp = pktlim;
  (*cb) (cp, len, NULL);
  return true;
}

void
axprt_crypt::sendv (const iovec *iov, int cnt, const sockaddr *)
{
  if (writefd < 0)
    panic ("axprt_stream::sendv: called after an EOF\n");

  if (!cryptsend) {
    axprt_stream::sendv (iov, cnt, NULL);
    return;
  }

  bool blocked = out->resid ();

  u_int32_t len = iovsize (iov, cnt);
  if (len > pktsize) {
    warn ("axprt_stream::sendv: packet too large\n");
    fail ();
    return;
  }

  u_char mk1[sizeof (mackey1)];
  u_char mk2[sizeof (mackey2)];
  for (size_t i = 0; i < sizeof (mk1); i++)
    mk1[i] = ctx_send.getbyte ();
  for (size_t i = 0; i < sizeof (mk2); i++)
    mk2[i] = ctx_send.getbyte ();

  sha1ctx sc;
  sc.update (mk1, sizeof (mackey1));

  u_char *msgbuf
    = reinterpret_cast<u_char *> (out->getspace (len + macsize + 4));
  u_char *cp = msgbuf;

  putint (cp, 0x80000000 | len);
  cp += 4;

  for (const iovec *lastiov = iov + cnt; iov < lastiov; iov++) {
    const char *p = static_cast<char *> (iov->iov_base);
    memcpy (cp, p, iov->iov_len);
    cp += iov->iov_len;
  }

  cp = msgbuf;
  sc.update (cp, len + 4);
  for (u_int32_t i = 0; i < len + 4; i++)
    *cp++ ^= ctx_send.getbyte ();

  sc.update (mk2, sizeof (mackey2));
  u_char mac[sha1::hashsize];
  sc.final (mac);
  for (int i = 0; i < macsize; i++)
    *cp++ = mac[i] ^ ctx_send.getbyte ();

  assert (msgbuf + len + macsize + 4 == cp);

  out->print (msgbuf, cp - msgbuf);
  raw_bytes_sent += cp - msgbuf;

  if (!blocked)
    output ();

#if 0
  void (axprt_crypt::*op) () = &axprt_crypt::output;
  fdcb (fd, selwrite, wrap (this, op));
  wcbset = true;
#endif
}

void
axprt_crypt::encrypt (const void *sendkey, size_t sendkeylen,
		      const void *recvkey, size_t recvkeylen)
{
  if (xhip && xhip->svcnum ()) {
    warn ("axprt_crypt::encrypt called while serving RPCs\n");
    fail ();
    return;
  }
  ctx_send.setkey (sendkey, sendkeylen);
  ctx_recv.setkey (recvkey, recvkeylen);
  cryptsend = cryptrecv = true;
}

ref<axprt_crypt>
axprt_crypt::alloc (int f, size_t ps)
{
  return New refcounted<axprt_crypt> (f, ps);
}
