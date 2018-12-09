// -*-c++-*-
/* $Id: test_axprt.C 2 2003-09-24 14:35:33Z max $ */

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
#include "axprt_crypt.h"

struct xprtest {
  enum { pktsize = 0x10000 };
  enum { npkt = 128 };

  cbv cb;
  str testname;
  int nopen;

  struct state {
    int sctr;
    int rctr;
    ptr<axprt> x;
    arc4 snd;
    arc4 rcv;

    state () : sctr (0), rctr (0) {}

    static u_int32_t getint (arc4 &as) {
      return (as.getbyte () << 24 | as.getbyte () << 16
	      | as.getbyte () << 8 | as.getbyte ());
    }

    void input (xprtest *xt, const char *pkt, ssize_t len, const sockaddr *) {
      // warn ("input %p: %d\n", this, len);
      if (len <= 0)
	panic << xt->testname << ": receive error\n";
      if (rctr > xprtest::npkt)
	panic << xt->testname << ": too many packets received\n";

      if ((size_t) len != (getint (rcv) % xprtest::pktsize & ~3))
	panic << xt->testname << ": received packet of incorrect size\n";

      u_char *goodpkt = New u_char[len];
      for (int i = 0; i < len; i++)
	goodpkt[i] = rcv.getbyte ();
      if (memcmp (pkt, goodpkt, len))
	  panic << xt->testname << ": bad byte in packet\n";
      delete[] goodpkt;

      if (++rctr == xprtest::npkt && !--xt->nopen) {
	cbv cb = xt->cb;
	delete xt;
	(*cb) ();
      }
    }

    void output () {
      for (int pn = 0; pn < 3 && sctr < xprtest::npkt; pn++, sctr++) {
	char pkt[pktsize];
	int len = getint (snd) % xprtest::pktsize & ~3;
	for (int j = 0; j < len; j++)
	  pkt[j] = snd.getbyte ();
	// warn ("output %p: %d\n", this, len);
	x->send (pkt, len, NULL);
      }
      if (sctr < xprtest::npkt)
	x->setwcb (wrap (this, &state::output));
    }
  };
  state s[2];

  xprtest (str name, ptr<axprt> a, ptr<axprt> b, cbv cb)
    : cb (cb), testname (name), nopen (2) {
    s[0].x = a;
    s[0].snd.setkey ("AS/BR", 5);
    s[0].rcv.setkey ("BS/AR", 5);

    s[1].x = b;
    s[1].snd.setkey ("BS/AR", 5);
    s[1].rcv.setkey ("AS/BR", 5);

    s[0].x->setrcb (wrap (&s[0], &state::input, this));
    s[0].output ();
    s[1].x->setrcb (wrap (&s[1], &state::input, this));
    s[1].output ();
  }
  ~xprtest () {
    s[0].x->setrcb (NULL);
    s[1].x->setrcb (NULL);
  }
};

struct bigtest {
  enum { npkt = 10 };

  str name;
  ref<axprt> snd;
  ref<axprt> rcv;
  const ssize_t size;
  u_char *const msg;
  int count;
  cbv cb;

  void input (const char *pkt, ssize_t len, const sockaddr *) {
    if (len <= 0)
      panic << name << ": receive error\n";
    if (len != size)
      panic << name << ": bad packet size\n";
    if (memcmp (pkt, msg, size))
      panic << name << ": bad packet contents\n";
    if (!--count) {
      cbv c = cb;
      delete this;
      (*c) ();
    }
  }

  bigtest (str name, ref<axprt> snd, ref<axprt> rcv, size_t size, cbv cb)
    : name (name), snd (snd), rcv (rcv), size (size),
      msg (New u_char[size]), count (npkt), cb (cb) {
    arc4 gen;
    gen.setkey ("bigmsgkey", 9);
    for (u_char *p = msg; p < msg + size; p++)
      *p = gen.getbyte ();
    rcv->setrcb (wrap (this, &bigtest::input));
    for (int i = 0; i < npkt; i++)
      snd->send (msg, size, NULL);
  }
  ~bigtest () { rcv->setrcb (NULL); delete[] msg; }
};

ptr<axprt_stream> sta, stb;
ptr<axprt_crypt> cra, crb;

static inline const u_char *
s2ucp (const str &s)
{
  return reinterpret_cast<const u_char *> (s.cstr ());
}

static void dobig (bool last);

static void
docrypt ()
{
  str kab = "key from a to b";
  str kba = "key from b to a";

  cra->encrypt (s2ucp (kab), kab.len (), s2ucp (kba), kba.len ());
  crb->encrypt (s2ucp (kba), kba.len (), s2ucp (kab), kab.len ());

  vNew xprtest ("axprt_crypt (encrypted)", cra, crb, wrap (dobig, true));
}

static void
dobig (bool last)
{
  if (last)
    vNew bigtest ("axprt_crypt (encrypted, big messages)",
		  cra, crb, axprt_stream::defps, wrap (exit, 0));
  else
    vNew bigtest ("axprt_crypt (unencrypted, big messages)",
		  cra, crb, axprt_stream::defps, wrap (docrypt));
}

static void
startcrypt ()
{
  sta = stb = NULL;
  vNew xprtest ("axprt_crypt (unencrypted)", cra, crb, wrap (dobig, false));
}

void
startstream ()
{
  vNew xprtest ("axprt_stream", sta, stb, wrap (startcrypt));
}

int
main (int argc, char **argv)
{
  setprogname (argv[0]);

  int fds[2];
  if (socketpair (AF_UNIX, SOCK_STREAM, 0, fds) < 0)
    fatal ("socketpair: %m\n");
  sta = axprt_crypt::alloc (fds[0]);
  stb = axprt_crypt::alloc (fds[1]);

  if (socketpair (AF_UNIX, SOCK_STREAM, 0, fds) < 0)
    fatal ("socketpair: %m\n");
  cra = axprt_crypt::alloc (fds[0]);
  crb = axprt_crypt::alloc (fds[1]);

  // startstream ();
  docrypt ();

  amain ();
}
