// -*-c++-*-
/* $Id: uvfstrans.h 1754 2006-05-19 20:59:19Z max $ */

/*
 *
 * Copyright (C) 1999 Michael Kaminsky (kaminsky@lcs.mit.edu)
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
#include "ihash.h"
#include "uvfs_prot.h"

struct uvfstrans;

struct uvfsstate {
  typedef rpc_bytes<NFS3_FHSIZE> fh_t;

  struct fhpair {
    u_int32_t srvno;
    fh_t ufh;
    fh_t nfh;
    ihash_entry<fhpair> ulink;
    ihash_entry<fhpair> nlink;
  };

  u_int32_t fhcount;

  ihash<fh_t, fhpair, &uvfsstate::fhpair::ufh,
    &uvfsstate::fhpair::ulink> uton_tab;
  ihash2<fh_t, u_int32_t, fhpair, &uvfsstate::fhpair::nfh,
    &uvfsstate::fhpair::srvno, &uvfsstate::fhpair::nlink> ntou_tab;

  fh_t nextfh () {
    fh_t obj;
    obj.setsize (4);
    do {
      putint (obj.base(), fhcount++);
    } while (uton_tab[obj]);
    return obj;
  }

  bool translate (uvfstrans &, fh_t &);
  void reclaim (fh_t &);
};

struct uvfstrans {
  enum op { TOUVFS, FROMUVFS } dir;
  bool srvno_valid;
  u_int32_t srvno;
  uvfsstate *state;
  int error;

  uvfstrans (op d, uvfsstate *s)
    : dir (d), srvno_valid (false), state (s), error (0)
    { assert (dir == FROMUVFS); }
  uvfstrans (op d, uvfsstate *s, u_int32_t sn)
    : dir (d), srvno_valid (true), srvno (sn), state (s), error (0)
    { assert (dir == TOUVFS); }
};

DUMBTRAVERSE (uvfstrans)

inline bool
rpc_traverse (uvfstrans &fht, uvfs_fh &fh)
{
  return fht.state->translate (fht, fh.data);
}
bool uvfs_transarg (uvfstrans &fht, void *objp, u_int32_t proc);
bool uvfs_transres (uvfstrans &fht, void *objp, u_int32_t proc);
int uvfs_open ();
void uvfs_err (svccb *sbp, uvfsstat s);
