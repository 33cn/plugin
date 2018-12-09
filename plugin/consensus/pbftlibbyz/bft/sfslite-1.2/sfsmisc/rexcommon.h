// -*-c++-*-
/* $Id: rexcommon.h 1754 2006-05-19 20:59:19Z max $ */

/*
 *
 * Copyright (C) 2003 David Mazieres (dm@uun.org)
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

#ifndef _SFSMISC_REXCOMMON_H_
#define _SFSMISC_REXCOMMON_H_ 1

#include "sfsmisc.h"
#include "rex_prot.h"
#include "sfsagent.h"

struct sfsagent_rex_res_w : public sfsagent_rex_res {
  sfsagent_rex_res_w () : sfsagent_rex_res () {}
  sfsagent_rex_res_w (bool s) : sfsagent_rex_res (s) {}
  ~sfsagent_rex_res_w ()
    { rpc_wipe (implicit_cast<sfsagent_rex_res &> (*this)); }
};

void
rex_mkkeys (rpc_bytes<> *ksc, rpc_bytes<> *kcs, sfs_hash *sessid,
	    sfs_seqno seqno, const sfs_kmsg &sdat, const sfs_kmsg &cdat);

void
rex_mksecretid (vec<char> &secretid, rpc_bytes<> &ksc, rpc_bytes<> &kcs);

#endif /* _SFSMISC_REXCOMMON_H_ */
