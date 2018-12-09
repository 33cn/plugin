/* $Id: xdrmisc.C 1693 2006-04-28 23:17:35Z max $ */

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

const stompcast_t _stompcast = stompcast_t ();
struct rpc_clear_t _rpcclear;
struct rpc_wipe_t _rpcwipe;
const str rpc_emptystr ("");
const char __xdr_zero_bytes[4] = { 0, 0, 0, 0 };

BOOL
xdr_void (XDR *xdrs, void *)
{
  return true;
}
void *
void_alloc ()
{
  return NULL;
}

BOOL
xdr_false (XDR *xdrs, void *)
{
  return false;
}
void *
false_alloc ()
{
  return NULL;
}

BOOL
xdr_string (XDR *xdrs, void *objp)
{
  return rpc_traverse (xdrs, *static_cast<rpc_str<RPC_INFINITY> *> (objp));
}
void *
string_alloc ()
{
  return New rpc_str<RPC_INFINITY>;
}

BOOL
xdr_int (XDR *xdrs, void *objp)
{
  u_int32_t val;
  switch (xdrs->x_op) {
  case XDR_ENCODE:
    val = *static_cast<int *> (objp);
    return rpc_traverse (xdrs, val);
  case XDR_DECODE:
    val = 0; // silence buggy warning message in gcc 4.1
    if (!rpc_traverse (xdrs, val))
      return false;
    *static_cast<int *> (objp) = val;
  default:
    return true;
  }
}
void *
int_alloc ()
{
  return New int;
}

#define DEFXDR(type)						\
BOOL								\
xdr_##type (XDR *xdrs, void *objp)				\
{								\
  return rpc_traverse (xdrs, *static_cast<type *> (objp));	\
}								\
void *								\
type##_alloc ()							\
{								\
  return New type;						\
}

DEFXDR(bool)
DEFXDR(int32_t)
DEFXDR(u_int32_t)
DEFXDR(int64_t)
DEFXDR(u_int64_t)

RPC_PRINT_GEN (char, sb.fmt ("%02x", (int) obj & 0xff))
RPC_PRINT_GEN (int32_t, sb << obj)
RPC_PRINT_GEN (u_int32_t, sb.fmt ("0x%x", obj))
RPC_PRINT_GEN (int64_t, sb << obj)
RPC_PRINT_GEN (u_int64_t, sb.fmt ("0x%" U64F "x", obj))
RPC_PRINT_GEN (bool, sb << (obj ? "true" : "false"))

RPC_PRINT_DEFINE(bool)
RPC_PRINT_DEFINE(int32_t)
RPC_PRINT_DEFINE(u_int32_t)
RPC_PRINT_DEFINE(int64_t)
RPC_PRINT_DEFINE(u_int64_t)
