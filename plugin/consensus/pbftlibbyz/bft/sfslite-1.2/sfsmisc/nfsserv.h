// -*-c++-*-
/* $Id: nfsserv.h 2679 2007-04-04 20:53:20Z max $ */

/*
 *
 * Copyright (C) 2000-2002 David Mazieres (dm@uun.org)
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

#ifndef _SFSMISC_NFSSERV_H_
#define _SFSMISC_NFSSERV_H_ 1

/*
 *  This file adds a layer of indirection between the RPC layer and
 *  code implementing an NFS3 server.  Thus, one can reuse the server
 *  code with non-NFS3 clients by adding a translation layer (e.g.,
 *  from UVFS to NFS3, or NFS2 to NFS3).  The class adds one fake NFS
 *  call, CLOSE, which in NFS must be simulated using timing
 *  heuristics, but can be implemented reliably under UVFS.
 */

#include "arpc.h"
#include "nfs3_prot.h"
#include "nfs3close_prot.h"
#include "blowfish.h"
#include "sfsmisc.h"
#include "qhash.h"
#include "nfstrans.h"

#if 0
enum { lastnfscall = NFSPROC3_COMMIT };
enum { NFSPROC_CLOSE = lastnfscall+1 };
#else
enum { NFSPROC_CLOSE = cl_NFSPROC3_CLOSE };
#endif

template<int N> struct nfs3proc {};
#define defineproc(proc, arg, res)		\
template<> struct nfs3proc<proc> {		\
  typedef arg arg_type;				\
  typedef res res_type;				\
};
NFS_PROGRAM_3_APPLY (defineproc)
defineproc (NFSPROC_CLOSE, nfs_fh3, nfsstat3)
#undef defineproc

struct nfsserv;

struct nfscall {
  static const rpcgen_table closert;

  const authunix_parms *const aup;
  const u_int32_t procno;
  void *const argp;
  void *resp;
  sfs::xdrproc_t xdr_res;
  accept_stat acstat;
  auth_stat austat;
  const time_t rqtime;
  bool nocache;
  bool nofree;

  nfsserv *stopserv;
  nfsserv *curserv;

  nfscall (const authunix_parms *aup, u_int32_t p, void *a);
  virtual ~nfscall () { clearres (); }
  void sendreply ();
  void setreply (void *, sfs::xdrproc_t = NULL, bool nc = false);
  void reply (void *, sfs::xdrproc_t = NULL, bool nc = false);
  void seterr (nfsstat3 err);
  void error (nfsstat3 err) { seterr (err); sendreply (); }
  void reject (accept_stat s) { acstat = s; error (NFS3ERR_IO); }
  void reject (auth_stat s) { austat = s; error (NFS3ERR_PERM); }
  const rpcgen_table &getrpct () const;
  void pinres ();

  sfs_aid getaid () const { return aup2aid (aup); }
  const authunix_parms *getaup () const { return aup; }
  u_int32_t proc () const { return procno; }
  void *getvoidarg () { return argp; }
  nfs_fh3 *getfh3arg () { return static_cast<nfs_fh3 *> (argp); }
  template<class T> T *getarg () {
#ifdef CHECK_BOUNDS
    assert ((typeid (T) == typeid (nfs_fh3) && proc () != NFSPROC3_NULL)
	    || typeid (T) == *getrpct ().type_arg);
#endif /* !CHECK_BOUNDS */
    return static_cast<T *> (getvoidarg ());
  }
  template<class T> const T *getarg () const {
    return const_cast<nfscall *> (this)->template getarg<T> ();
  }
  void clearres ();
  void *getvoidres ();
  template<class T> T *getres () { return static_cast<T *> (getvoidres ()); }
};

template<int N> class nfscall_cb : public nfscall {
  typedef typename nfs3proc<N>::arg_type *arg_type;
  typedef typename nfs3proc<N>::res_type *res_type;
  typedef ref<callback<void, res_type> > cb_t;
  cb_t cb;
public:
  nfscall_cb (const authunix_parms *au, arg_type a, cb_t c, nfsserv *srv);
  ~nfscall_cb () {
    /* Note, if xdr_res is not the default, we could always marshall
     * and unmarshall the result to get it in the right type.  That
     * would be slow, however, and if it were actually happening, it
     * would probably indicate something wrong with the code.  Thus,
     * we force an assertion failure in this case. */
    assert (!xdr_res);
    (*cb) (static_cast<res_type> (resp));
  }
};

struct nfscall_rpc : nfscall {
  svccb *sbp;
  nfscall_rpc (svccb *sbp);
  ~nfscall_rpc ();
};

struct nfsserv : public virtual refcount {
  typedef callback<void, nfscall *>::ref cb_t;
  static const cb_t stalecb;
  cb_t cb;
  const ptr<nfsserv> nextserv;
  explicit nfsserv (ptr<nfsserv> n = NULL);
  void setcb (const cb_t &c) { cb = c; }
  void mkcb (nfscall *nc) { nc->curserv = this; (*cb) (nc); }
  virtual void getcall (nfscall *nc) { mkcb (nc); }
  virtual void getreply (nfscall *nc) { nc->sendreply (); }
  virtual bool encodefh (nfs_fh3 &fh);
};

template<int N> inline
nfscall_cb<N>::nfscall_cb (const authunix_parms *au, arg_type a, cb_t c,
			   nfsserv *srv)
  : nfscall (au, N, a), cb (c)
{
  if ((stopserv = srv))
    srv->mkcb (this);
}

class nfsserv_udp : public nfsserv {
  int fd;
  ptr<axprt> x;
  ptr<asrv> s;
  void getsbp (svccb *sbp);
public:
  nfsserv_udp ();
  nfsserv_udp (const rpc_program&);
  // void getreply (nfscall *nc);
  int getfd () { return fd; }
};

/* Work around some bugs in NFS servers */
class nfsserv_fixup : public nfsserv {
  void getattr (nfscall *nc, nfs_fh3 *, getattr3res *res);
public:
  explicit nfsserv_fixup (ref<nfsserv> s) : nfsserv (s) {}
  void getreply (nfscall *nc);
};


struct nfsserv_link : public nfsserv {
  ref<nfsserv> next;
  nfsserv_link (const ref<nfsserv> &n) : next (n) {}
  virtual void getreply (nfscall *);
};

class nfsdemux : public virtual refcount {
public:
  struct nfsserv_cryptfh : public nfsserv {
    const ref<nfsdemux> d;
    const u_int32_t srvno;
    ihash_entry <nfsserv_cryptfh> hlink;
    nfsserv_cryptfh (const ref<nfsdemux> &dd, u_int32_t s);
    ~nfsserv_cryptfh ();
    virtual void getcall (nfscall *nc) { mkcb (nc); }
    virtual void getreply (nfscall *nc);
    bool encodefh (nfs_fh3 &fh);
  };
  friend class nfsserv_cryptfh;

private:
  const ref<nfsserv> ns;
  u_int32_t srvnoctr;
  ihash<const u_int32_t, nfsserv_cryptfh,
    &nfsserv_cryptfh::srvno, &nfsserv_cryptfh::hlink> srvnotab;
  void getcall (nfscall *nc);
public:
  blowfish fhkey;
  nfsdemux (const ref<nfsserv> &n);
  ref<nfsserv_cryptfh> servalloc ();
};

/* Attribute caching manipulator.
 *
 * This server expects
 *
 * Warning, this server must be pushed after (i.e., later in the
 * stream than) any servers that manipulate (e.g., encrypt/decrypt)
 * file handles.
 */
class attr_cache {
  struct access_dat {
    u_int32_t mask;
    u_int32_t perm;

    access_dat (u_int32_t m, u_int32_t p) : mask (m), perm (p) {}
  };

public:
  struct attr_dat {
    attr_cache *const cache;
    const nfs_fh3 fh;
    fattr3exp attr;
    qhash<sfs_aid, access_dat> access;

    ihash_entry<attr_dat> fhlink;
    tailq_entry<attr_dat> lrulink;

    attr_dat (attr_cache *c, const nfs_fh3 &f, const fattr3exp *a);
    ~attr_dat ();
    void touch ();
    void set (const fattr3exp *a, const wcc_attr *w);
    bool valid () 
    { return sfs_get_timenow() < implicit_cast<time_t> (attr.expire); }
  };

private:
  friend class attr_dat;
  ihash<const nfs_fh3, attr_dat, &attr_dat::fh, &attr_dat::fhlink> attrs;

  static void remove_aid (sfs_aid aid, attr_dat *ad)
    { ad->access.remove (aid); }

public:
  ~attr_cache () { attrs.deleteall (); }
  void flush_attr () { attrs.deleteall (); }
  void flush_access (sfs_aid aid) { attrs.traverse (wrap (remove_aid, aid)); }
  void flush_access (const nfs_fh3 &fh, sfs_aid);

  void attr_enter (const nfs_fh3 &, const fattr3exp *, const wcc_attr *);
  void attr_enter (const nfs_fh3 &fh, const ex_fattr3 *a, const wcc_attr *w)
    { attr_enter (fh, reinterpret_cast<const fattr3exp *> (a), w); }
  const fattr3exp *attr_lookup (const nfs_fh3 &);

  void access_enter (const nfs_fh3 &, sfs_aid aid,
		     u_int32_t mask, u_int32_t perm);
  int32_t access_lookup (const nfs_fh3 &, sfs_aid, u_int32_t mask);
};

class nfsserv_ac : public nfsserv {
public:
  attr_cache ac;
  explicit nfsserv_ac (ref<nfsserv> s) : nfsserv (s) {}
  void getcall (nfscall *nc);
  void getreply (nfscall *nc);
};

ref<nfsserv> close_simulate (ref<nfsserv> ns);

#endif /* _SFSMISC_NFSSERV_H_ */
