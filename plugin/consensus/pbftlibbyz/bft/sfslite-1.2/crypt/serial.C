/* $Id: serial.C 1117 2005-11-01 16:20:39Z max $ */

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

#include "crypt.h"
#include "blowfish.h"
#include "password.h"
#include "rxx.h"
#include "srp.h"

// XXX - need explicit instantiation for KCC
template const strbuf &strbuf_cat (const strbuf &,
				   const strbufcatobj<bigint, int> &);

const rxx srp_import_format ("^N=(0x[0-9a-f]+),g=(0x[0-9a-f]+)$");

bool
import_srp_params (str raw, bigint *Np, bigint *gp)
{
  if (!raw)
    return false;

  rxx r (srp_import_format);
  if (!r.search (raw))
    return false;

  *Np = r[1];
  *gp = r[2];
  return true;
}

str
export_srp_params (const bigint &N, const bigint &g)
{
  return strbuf ("N=0x") << N.getstr (16) << ",g=0x" << g.getstr (16);
}
