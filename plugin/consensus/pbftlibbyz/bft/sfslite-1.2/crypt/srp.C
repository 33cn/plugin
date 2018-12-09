/* $Id: srp.C 1117 2005-11-01 16:20:39Z max $ */

/*
 *
 * Copyright (C) 1999 David Mazieres (dm@uun.org)
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


#include "crypt.h"
#include "prime.h"
#include "password.h"
#include "rxx.h"
#include "crypt_prot.h"
#include "srp.h"

bigint srp_base::k1 (1);
bigint srp_base::k3 (3);
u_int srp_base::minprimsize;
srp_base::paramcache srp_base::cache[srp_base::cachesize];
int srp_base::lastpos;

bool
srp_base::setS (const bigint &SS)
{
  S = SS;

  sha1ctx sc;
  if (!datasink_catxdr (sc, sessid)
      || !datasink_catxdr (sc, N)
      || !datasink_catxdr (sc, g)
      || !datasink_catxdr<str> (sc, user)
      || !datasink_catxdr (sc, salt)
      || !datasink_catxdr (sc, A)
      || !datasink_catxdr (sc, B)
      || !datasink_catxdr (sc, S, true))
    return false;
  sc.final (M.base ());

  sc.reset ();
  if (!datasink_catxdr (sc, sessid)
      || !datasink_catxdr (sc, A)
      || !datasink_catxdr (sc, M)
      || !datasink_catxdr (sc, S, true))
    return false;
  sc.final (H.base ());

  return true;
}

bool
srp_base::checkparam (const bigint &N, const bigint &g, u_int iter)
{
  bigint N1 (N - 1);
  if (N.nbits () < minprimsize || g != g % N || g == N1
      || powm (g, N >> 1, N) != N1)
    return false;

  for (int i = 0; i < cachesize; i++)
    if (cache[i].N == N && cache[i].iter >= iter && !!N) {
      lastpos = i;
      return true;
    }

  if (!srpprime_test (N, iter))
    return false;

  lastpos = (lastpos + 1) % cachesize;
  cache[lastpos].N = N;
  cache[lastpos].iter = iter;

  return true;
}

bool
srp_base::seedparam (const bigint &N, const bigint &g, u_int iter)
{
  if (!checkparam (N, g, iter))
    return false;
  cache[lastpos].iter = 0x10000;
  return true;
}

void
srp_base::genparam (size_t nbits, bigint *Np, bigint *gp)
{
  *Np = random_srpprime (nbits);

  /* XXX - written in C-like syntax to work around bugs in gcc-2.95.2 */
  mpz_t q, t;
  mpz_init (q);
  mpz_init (t);

  /* XXX - why not just mpz_tdiv_q_2exp (q, Np, 1)? */
  mpz_sub_ui (q, Np, 1);
  mpz_tdiv_q_2exp (q, q, 1);

  for (mpz_set_ui (gp, 2);; mpz_add_ui (gp, gp, 1)) {
    mpz_powm (t, gp, q, Np);
    if (mpz_cmp_ui (t, 1)) {
      mpz_clear (q);
      mpz_clear (t);
      return;
    }
  }
}

srpres
srp_client::init (srpmsg *msgout, const srp_hash &sid,
                  str uu, str pp, int version)
{
  k = (version < 6) ? &k1 : &k3;  // the former is for SRP-3, the latter SRP-6
  user = uu;
  pwd = pp;
  host = NULL;
  host_ok = false;
  sessid = sid;
  msgout->setsize (0);
  phase = 1;
  return SRP_NEXT;
}

srpres
srp_client::phase1a (srpmsg *msgout, const srpmsg *msgin)
{
  srp_msg1 m;
  if (!bytes2xdr (m, *msgin))
    return SRP_FAIL;

  if (m.N != N || m.g != g) {
    setparam (m.N, m.g);
    if (!checkparam ())
      return SRP_FAIL;
  }

  salt = m.salt;
  if (!pw_dearmorsalt (&cost, NULL, &host, salt))
    return SRP_FAIL;

  if (!pwd) {
    phase = 0x1b;
    return SRP_SETPWD;
  }
  else
    return phase1b (msgout, msgin);
}

srpres
srp_client::phase1b (srpmsg *msgout, const srpmsg *msgin)
{
  x = pw_getint (pwd, salt, N.nbits () - 1, &eksb);
  pwd = NULL;

  a = random_zn (N);
  A = powm (g, a, N);
  if (!xdr2bytes (*msgout, A))
    return SRP_FAIL;

  phase = 3;
  return SRP_NEXT;
}

srpres
srp_client::phase3 (srpmsg *msgout, const srpmsg *msgin)
{
  srp_msg3 m;
  if (!bytes2xdr (m, *msgin) || !m.B || !m.u)
    return SRP_FAIL;

  B = m.B;
  if (!setS (powm (B - *k * powm (g, x, N), a + m.u * x, N)))
    return SRP_FAIL;

  if (!xdr2bytes (*msgout, M))
    return SRP_FAIL;

  phase = 5;
  return SRP_NEXT;
}

srpres
srp_client::phase5 (srpmsg *msgout, const srpmsg *msgin)
{
  srp_hash m;
  if (!bytes2xdr (m, *msgin) || m != H)
    return SRP_FAIL;
  host_ok = true;
  return SRP_DONE;
}

srpres
srp_client::next (srpmsg *msgout, const srpmsg *msgin)
{
  int ophase = phase;
  phase = -1;
  switch (ophase) {
  case 1:
    return phase1a (msgout, msgin);
  case 0x1b:
    return phase1b (msgout, msgin);
  case 3:
    return phase3 (msgout, msgin);
  case 5:
    return phase5 (msgout, msgin);
  default:
    return SRP_FAIL;
  }
}

static rxx hostrx ("^[\\w\\.\\-]*$");

str
srp_client::create (const bigint &NN, const bigint &gg,
		    str pp, str hh, u_int cost, u_int iter)
{
  phase = -1;
  setparam (NN, gg);
  if (!checkparam (iter) || !hostrx.search (hh))
    return NULL;
  pwd = NULL;
  host = hh;
  salt = pw_gensalt (cost, host);
  bigint x (pw_getint (pp, salt, N.nbits () - 1, &eksb));
  if (!x)
    return NULL;
  bigint v (powm (g, x, N));

  return strbuf () << "SRP,N=0x" << N.getstr (16)
		   << ",g=0x" << g.getstr (16)
		   << ",s=" << salt
		   << ",v=0x" << v.getstr (16);
}

const rxx srpinforx ("^SRP,N=(0x[\\da-f]+),g=(0x[\\da-f]+),"
		     "s=(\\d+\\$[A-Za-z0-9+/]+={0,2}\\$[\\w\\.\\-]*),"
		     "v=(0x[\\da-f]+)$");

bool
srp_server::sane (str info)
{
  rxx r (srpinforx);
  if (!info || !r.search (info))
    return false;
  bigint N (r[1]);
  bigint g (r[2]);
  if (!checkparam (N, g, 0))
    return false;
  return true;
}

srpres
srp_server::init (srpmsg *msgout, const srpmsg *msgin,
		  const srp_hash &sid, str uu, str info, int version)
{
  k = (version < 6) ? &k1 : &k3;  // the former is for SRP-3, the latter SRP-6
  if (msgin->size () || !info || !info.len ())
    return SRP_FAIL;
  rxx r (srpinforx);
  if (!r.search (info) || !(N = r[1], g = r[2], checkparam (1))) {
    /* Unnecessary sanity check */
    warn << "Bad SRP parameters for user " << uu << "\n";
    return SRP_FAIL;
  }
  user = uu;
  salt = r[3];
  sessid = sid;
  v = r[4];

  srp_msg1 m;
  m.salt = salt;
  m.N = N;
  m.g = g;
  if (!xdr2bytes (*msgout, m))
    return SRP_FAIL;

  phase = 2;
  return SRP_NEXT;
}

srpres
srp_server::phase2 (srpmsg *msgout, const srpmsg *msgin)
{
  if (!bytes2xdr (A, *msgin) || !A)
    return SRP_FAIL;

  b = random_zn (N);
  B = *k * v;           // XXX: want single expression; bigint.h bug?
  B += powm (g, b, N);
  B %= N;
  u = random_zn (N);

  srp_msg3 m;
  m.B = B;
  m.u = u;
  if (!xdr2bytes (*msgout, m))
    return SRP_FAIL;

  phase = 4;
  return SRP_NEXT;
}

srpres
srp_server::phase4 (srpmsg *msgout, const srpmsg *msgin)
{
  srp_hash m;
  if (!bytes2xdr (m, *msgin)
      || !setS (powm (A * powm (v, u, N), b, N))
      || m != M
      || !xdr2bytes (*msgout, H))
    return SRP_FAIL;
  return SRP_LAST;
}

srpres
srp_server::next (srpmsg *msgout, const srpmsg *msgin)
{
  int ophase = phase;
  phase = -1;
  switch (ophase) {
  case 2:
    return phase2 (msgout, msgin);
  case 4:
    return phase4 (msgout, msgin);
  default:
    return SRP_FAIL;
  }
}
