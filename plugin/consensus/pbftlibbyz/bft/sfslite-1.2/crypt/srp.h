// -*-c++-*-
/* $Id: srp.h 1117 2005-11-01 16:20:39Z max $ */

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

#ifndef _SRP_H_
#define _SRP_H_ 1

#include "bigint.h"
#include "sha1.h"
#include "blowfish.h"

typedef rpc_opaque<20> srp_hash;
typedef rpc_bytes<RPC_INFINITY> srpmsg;

enum srpres { SRP_FAIL, SRP_NEXT, SRP_SETPWD, SRP_LAST, SRP_DONE };

class srp_base {
protected:
  str salt;
  bigint A;
  bigint B;
  srp_hash M;
  srp_hash H;
  static bigint k1;
  static bigint k3;
  bigint *k;

  void setparam (const bigint &NN, const bigint &gg) { N = NN; g = gg; }
  bool checkparam (u_int iter = 32)
    { if (checkparam (N, g, iter)) return true; N = g = 0; return false; }
  bool setS (const bigint &S);

  struct paramcache {
    bigint N;
    u_int iter;
    paramcache () : iter (0) {}
  };
  enum { cachesize = 2 };
  static paramcache cache[cachesize];
  static int lastpos;

public:
  srp_hash sessid;
  str user;
  static u_int minprimsize;
  bigint S;
  bigint N;
  bigint g;

  static bool checkparam (const bigint &N, const bigint &g, u_int iter = 32);
  static bool seedparam (const bigint &N, const bigint &g, u_int iter = 1);
  static void genparam (size_t nbits, bigint *Np, bigint *gp);
};

class srp_client : public srp_base {
  int phase;
  bigint x;
  bigint a;

  srpres phase1a (srpmsg *msgout, const srpmsg *in);
  srpres phase1b (srpmsg *msgout, const srpmsg *in);
  srpres phase3 (srpmsg *msgout, const srpmsg *in);
  srpres phase5 (srpmsg *msgout, const srpmsg *in);

public:
  str pwd;
  str host;
  bool host_ok;
  u_int cost;
  eksblowfish eksb;

  srp_client () : phase (-1), host_ok (false) {}
  void setpwd (const str &p) { pwd = p; }
  str getname () const
    { return strbuf () << user << "@" << host << "/" << N.nbits (); }
  srpres init (srpmsg *msgout, const srp_hash &sessid,
	       str user, str pwd = NULL, int version = 6);
  srpres next (srpmsg *msgout, const srpmsg *in);

  str create (const bigint &N, const bigint &g,
	      str pwd, str host, u_int cost, u_int iter = 32);
};

class srp_server : public srp_base {
  int phase;
  bigint v;
  bigint b;
  bigint u;

  srpres phase2 (srpmsg *msgout, const srpmsg *msgin);
  srpres phase4 (srpmsg *msgout, const srpmsg *msgin);

public:
  srp_server () : phase (-1) {}
  srpres init (srpmsg *msgout, const srpmsg *msgin,
	       const srp_hash &sessid, str user, str info, int version = 6);
  srpres next (srpmsg *msgout, const srpmsg *msgin);

  static bool sane (str info);
};

bool import_srp_params (str raw, bigint *Np, bigint *gp);
str export_srp_params (const bigint &N, const bigint &g);

#endif /* !_SRP_H_  */
