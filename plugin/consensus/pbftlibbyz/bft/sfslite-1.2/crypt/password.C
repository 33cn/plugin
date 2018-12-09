/* $Id: password.C 3758 2008-11-13 00:36:00Z max $ */

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
#include "password.h"
#include "rxx.h"
#include "parseopt.h"

inline void
hashptext (char *dst, size_t dstlen, const str &src)
{
  sha1oracle ora (dstlen);
  ora.update (src.cstr (), src.len ());
  ora.final (reinterpret_cast<u_char *> (dst));
}

str
pw_armorsalt (u_int cost, str bsalt, str ptext)
{
  return strbuf ("%d$", cost) << armor64 (bsalt) << "$" << ptext;
}

#define A64STR "[A-Za-z0-9+/]+={0,2}"
static rxx saltrx ("^(\\d+)\\$(" A64STR ")\\$(.*)$");

bool
pw_dearmorsalt (u_int *costp, str *bsaltp, str *ptextp, str armor)
{
  if (!(armor / saltrx))
    return false;
  
  str s = dearmor64 (saltrx[2]);
  if (!s)
    return false;
  if (bsaltp)
    *bsaltp = s;
  if (costp)
    *costp = strtoi64 (saltrx[1]);
  if (ptextp)
    *ptextp = saltrx[3];

  return true;
}

str
pw_dorawcrypt (str ptext, size_t outsize, eksblowfish *eksb)
{
  wmstr m ((outsize + 7) & ~7);
  hashptext (m, m.len (), ptext);

  cbc64iv iv (*eksb);
  for (int i = 0; i < 64; i++) {
    iv.setiv (0, 0);
    iv.encipher_bytes (m, m.len ());
  }

  return m;
}

str
pw_rawcrypt (u_int cost, str pwd, str bsalt,
	     str ptext, size_t outsize, eksblowfish *eksb)
{
  u_int maxlen = 56;
  if (!outsize)
    outsize = ptext.len ();
  eksblowfish *eksbdel = NULL;
  if (!eksb)
    eksbdel = eksb = New eksblowfish;
  if (pwd.len () > maxlen) {
    char hsh[2 * sha1::hashsize];
    sha1_hash (reinterpret_cast <u_char *> (hsh), pwd.cstr (), pwd.len ());
    sha1_hash (reinterpret_cast <u_char *> (hsh + sha1::hashsize), 
      str (hsh, sha1::hashsize), sha1::hashsize);
    pwd = str (hsh, (2*sha1::hashsize < maxlen) ? 2*sha1::hashsize : maxlen);
  }
  eksb->eksetkey (cost, pwd.cstr (), pwd.len (), bsalt.cstr (), bsalt.len ());

  str ret = pw_dorawcrypt (ptext, outsize, eksb);
  delete eksbdel;
  return ret;
}

str
pw_gensalt (u_int cost, str ptext)
{
  mstr m (16);
  rnd.getbytes (m, m.len ());
  str bsalt (m);
  return pw_armorsalt (cost, bsalt, ptext);
}

str
pw_getptext (str salt)
{
  if (salt / saltrx)
    return saltrx[3];
  return NULL;
}

str
pw_crypt (str pwd, str salt, size_t outsize, eksblowfish *eksb)
{
  u_int cost;
  str bsalt, ptext;
  if (!pw_dearmorsalt (&cost, &bsalt, &ptext, salt))
    return NULL;
  return pw_rawcrypt (cost, pwd, bsalt, ptext, outsize, eksb);
}

bigint
pw_getint (str pwd, str salt, size_t nbits, eksblowfish *eksb)
{
  str raw = pw_crypt (pwd, salt, (nbits + 7) >> 3, eksb);
  if (!raw)
    return 0;
  bigint res;
  mpz_set_rawmag_le (&res, raw.cstr (), raw.len ());
  res.trunc (nbits);
  return res;
}
