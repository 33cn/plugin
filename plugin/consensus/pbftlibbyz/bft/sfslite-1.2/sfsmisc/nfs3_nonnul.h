/* -*-c++-*- */
/* $Id: nfs3_nonnul.h 1754 2006-05-19 20:59:19Z max $ */

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

#ifndef _SFSMISC_NFS3_NONNUL_H_
#define _SFSMISC_NFS3_NONNUL_H_ 1

struct nfs3void {
  nfs3void () {}
  void set_status (nfsstat3) {}
  template<class T> nfs3void (T) {}
};

template<class T> inline bool
rpc_traverse (T &, nfs3void &)
{
  return true;
}

#define NFS_PROGRAM_3_APPLY_NONULL(macro)	\
  NFS_PROGRAM_3_APPLY_NOVOID (macro, nfs3void)

#define ex_NFS_PROGRAM_3_APPLY_NONULL(macro)	\
  ex_NFS_PROGRAM_3_APPLY_NOVOID (macro, nfs3void)

template<class T>
inline bool nfs3_traverse_arg (T &t, u_int32_t nfs_proc, void *argp)
{
#define transarg(proc, arg_t, res_t)				\
 case proc:							\
   return rpc_traverse (t, *static_cast<arg_t *> (argp));

  switch (nfs_proc) {
    NFS_PROGRAM_3_APPLY_NONULL (transarg);
  default:
    panic ("bad NFS proc %d\n", nfs_proc);
  }

#undef transarg
}

template<class T>
inline bool nfs3_traverse_res (T &t, u_int32_t nfs_proc, void *resp)
{
#define transres(proc, arg_t, res_t)				\
 case proc:							\
   return rpc_traverse (t, *static_cast<res_t *> (resp));

  switch (nfs_proc) {
    NFS_PROGRAM_3_APPLY_NONULL (transres);
  default:
    panic ("bad NFS proc %d\n", nfs_proc);
  }

#undef transres
}

#endif /* _SFSMISC_NFS3_NONNUL_H_ */
