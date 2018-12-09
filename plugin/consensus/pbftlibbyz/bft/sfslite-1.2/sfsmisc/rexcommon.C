/* $Id: rexcommon.C 1754 2006-05-19 20:59:19Z max $ */

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

#include "rexcommon.h"
#include "sha1.h"

void
rex_mkkeys (rpc_bytes<> *ksc, rpc_bytes<> *kcs, sfs_hash *sessid,
	    sfs_seqno seqno, const sfs_kmsg &sdat, const sfs_kmsg &cdat)
{
  sfs_sessinfo sess;
  sess.type = SFS_SESSINFO;

  rex_sesskeydat skd;
  skd.seqno = seqno;

  skd.type = SFS_KCS;
  sess.kcs.setsize (sha1::hashsize);
  sha1_hmacxdr_2 (sess.kcs.base (),
		  sdat.kcs_share.base (), sdat.kcs_share.size (),
		  cdat.kcs_share.base (), cdat.kcs_share.size (),
		  skd, true);

  skd.type = SFS_KSC;
  sess.ksc.setsize (sha1::hashsize);
  sha1_hmacxdr_2 (sess.ksc.base (),
		  sdat.ksc_share.base (), sdat.ksc_share.size (),
		  cdat.ksc_share.base (), cdat.ksc_share.size (),
		  skd, true);

  if (sessid)
    sha1_hashxdr (sessid->base (), sess, true);
  if (kcs)
    swap (*kcs, sess.kcs);
  if (ksc)
    swap (*ksc, sess.ksc);

  bzero (sess.kcs.base (), sess.kcs.size ());
  bzero (sess.ksc.base (), sess.ksc.size ());
}

void
rex_mksecretid (vec<char> &secretid, rpc_bytes<> &ksc, rpc_bytes<> &kcs)
{
    sfs_sessinfo si;
    si.type = SFS_SESSINFO_SECRETID;
    si.ksc = ksc;
    si.kcs = kcs;

    sfs_hash dummy_hash;
    secretid.setsize (dummy_hash.size ());

    sha1_hashxdr (secretid.base (), si, true);
    bzero (si.kcs.base (), si.kcs.size ());
    bzero (si.ksc.base (), si.ksc.size ());
}

