// -*-c++-*-
/* $Id: sfsclient.h 2515 2007-01-25 05:58:24Z max $ */

/*
 *
 * Copyright (C) 2000 David Mazieres (dm@uun.org)
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

#ifndef _SFSMISC_SFSCLIENT_H_
#define _SFSMISC_SFSCLIENT_H_ 1

#include "nfsserv.h"
#include "sfsagent.h"
#include "vec.h"
#include "qhash.h"
#include "axprt_crypt.h"
#include "sfscrypt.h"
#include "sfscd_prot.h"

struct sfscd_mountarg;
class rabin_priv;
class sfsprog;

template<> struct qhash_lookup_return<auto_auth> {
  typedef AUTH *type;
  typedef AUTH *const_type;
  static type ret (auto_auth *a) { if (a) return *a; return NULL; }
};

struct sfsserverargs {
  typedef callback<void, const nfs_fh3 *>::ref fhcb;

  ref<nfsserv> ns;
  int fd;
  sfsprog *p;
  sfscd_mountarg *ma;
  fhcb cb;

  sfsserverargs (ref<nfsserv> ns, int fd, sfsprog *p,
		 sfscd_mountarg *ma, fhcb cb)
    : ns (ns), fd (fd), p (p), ma (ma), cb (cb) {}
};

class sfsserver : public virtual refcount {
public:
  typedef callback<void, const nfs_fh3 *>::ref fhcb;
private:
  bool lock_flag;
  bool condemn_flag;
  bool destroyed;
  int recon_backoff;
  timecb_t *recon_tmo;
  ptr<srvlist> srvl;

  void lock () { assert (!lock_flag); lock_flag = true; }
  void unlock ();

  void reconnect ();
  void reconnect_0 ();
  void reconnect_1 (int fd);
  void reconnect_2 (ref<sfs_connectres> res, clnt_stat err);
  void getsessid (const sfs_hash *);
  void getfsinfo (sfs_fsinfo *fsi, clnt_stat err);
  void getfsinfo2 (sfs_fsinfo *fsi, bool err);
  void connection_failure (bool permanent = false);
  void retrootfh (fhcb cb);

protected:
  ref<nfsserv> ns;
  vec<cbv> waitq;

  sfsserver (const sfsserverargs &a);
  virtual ~sfsserver ();

  void sfsdispatch (svccb *sbp);
  virtual void setfd (int fd);
  virtual sfs_fsinfo *fsinfo_alloc () { return New sfs_fsinfo; }
  virtual void fsinfo_free (sfs_fsinfo *fsi) { delete fsi; }
  virtual sfs::xdrproc_t fsinfo_marshall () { return xdr_sfs_fsinfo; }
  typedef callback<void, const sfs_hash *>::ref crypt_cb;
  virtual void crypt (sfs_connectok cres, ref<const sfs_servinfo_w> si, 
		      crypt_cb cb);
  virtual bool authok (nfscall *) { return true; }
  virtual void initstate () {}
  virtual void flushstate () {}
  virtual void setx (int fd) { x = axprt_stream::alloc (fd); }

  virtual void setrootfh (const sfs_fsinfo *fsi,
			  callback<void, bool>::ref err_cb) = 0;
  virtual void dispatch (nfscall *nc) = 0;

public:
  sfsprog &prog;
  str path;
  str dnsname;
  int portno;
  const sfs_connectinfo carg;
  ptr<const sfs_servinfo_w> si;
  sfs_authinfo authinfo;
  sfs_fsinfo *fsinfo;
  nfs_fh3 rootfh;
  time_t lastuse;

  ptr<axprt_stream> x;
  ptr<aclnt> sfsc;
  ptr<asrv> sfscbs;

  ihash_entry<sfsserver> pathlink;
  tailq_entry<sfsserver> idlelink;

  bool condemned () { return condemn_flag; }
  bool locked () { return lock_flag; }
  void touch ();
  void condemn ();
  void destroy ();
  void getnfscall (nfscall *nc);
  virtual AUTH *authof (sfs_aid) { return NULL; }
  virtual void authclear (sfs_aid) {}
};

class sfsserver_auth : public sfsserver {
protected:
  struct userauth : public virtual refcount {
    const sfs_aid aid;
  protected:
    const ref<sfsserver_auth> sp;
    vec<nfscall *> ncvec;
    timecb_t *tmo;
    callbase *cbase;
    bool aborted;
    sfsagent_auth_res ares;
    sfs_loginres_old sres;
    sfs_seqno seqno;
    int ntries;
    u_int32_t authno;

    userauth (sfs_aid aid, const ref<sfsserver_auth> &s);
    ~userauth ();
    virtual void finish ();
    void aresult (clnt_stat);
    void sresult (clnt_stat);
    void timeout ();

  public:
    void sendreq ();
    void pushreq (nfscall *nc);
    void abort ();
  };
  friend class userauth;

  sfs_seqno seqno;
  qhash<sfs_aid, auto_auth> authnos;
  qhash<sfs_aid, ref<userauth> > authpending;

  static ptr<sfspriv> privkey;
  static timecb_t *keytmo;
  static void keyexpire ();

  sfsserver_auth (const sfsserverargs &a)
    : sfsserver (a), seqno (0)  {}
  void setx (int fd) { x = xc = axprt_crypt::alloc (fd); }
  virtual void crypt (sfs_connectok cres, ref<const sfs_servinfo_w> si, 
		      crypt_cb cb);

  virtual ref<userauth> userauth_alloc (sfs_aid);
  virtual bool authok (nfscall *);
  virtual AUTH *authof (sfs_aid aid) { return authnos[aid]; }
  virtual void flushstate ();

public:
  ptr<axprt_crypt> xc;
  static void keygen ();
  virtual void authclear (sfs_aid);
};

class sfsserver_credmap : public sfsserver_auth {
protected:
  struct userauth_credmap : public userauth {
    ref<sfsauth_cred> cred;

    userauth_credmap (sfs_aid aid, const ref<sfsserver_credmap> &s);
    ~userauth_credmap () {}
    virtual void finish ();
    void cresult (clnt_stat);
  };
  friend class userauth_credmap;

  sfsserver_credmap (const sfsserverargs &a) : sfsserver_auth (a) {}
  ref<userauth> userauth_alloc (sfs_aid);
  void flushstate ();
public:
  qhash<sfs_aid, ref<sfsauth_cred> > credmap;
  qhash<u_int32_t, u_int32_t> uidmap;

  void authclear (sfs_aid);
  bool nomap (const authunix_parms *aup);
  void mapcred (const authunix_parms *aup, ex_fattr3 *fp,
		u_int32_t unknown_uid, u_int32_t unknown_gid);
};

template<class T> void
sfsserver_alloc (sfsprog *prog, ref<nfsserv> ns, int tcpfd,
		 sfscd_mountarg *ma, sfsserver::fhcb cb)
{
  if (!ma->cres || (ma->carg.civers == 5
		    && !sfs_parsepath (ma->carg.ci5->sname))) {
    (*cb) (NULL);		// Named protocols intercept here
    return;
  }
  vNew refcounted<T> (sfsserverargs (ns, tcpfd, prog, ma, cb));
}

template<class T> void
sfsserver_cache_alloc (sfsprog *prog, ref<nfsserv> ns, int tcpfd,
		 sfscd_mountarg *ma, sfsserver::fhcb cb)
{
  if (!ma->cres || (ma->carg.civers == 5
		    && !sfs_parsepath (ma->carg.ci5->sname))) {
    (*cb) (NULL);		// Named protocols intercept here
    return;
  }
  ref<nfsserv_ac> as = New refcounted<nfsserv_ac> (ns);
  vNew refcounted<T> (sfsserverargs (as, tcpfd, prog, ma, cb), &as->ac);
}

class sfsprog {
  struct sfsctl {
    struct fileinfo {
      const str fspath;
      const nfs_fh3 fh;
      fileinfo (sfsserver *s, const nfs_fh3 &fh) : fspath (s->path), fh (fh) {}
    };

    sfsprog *const prog;
    ptr<asrv> s;
    const sfs_aid aid;
    int32_t pid;
    struct authunix_parms aup;
    ihash_entry<sfsctl> hlink;
    ptr<fileinfo> fip;

    sfsctl (ref<axprt_stream> x, const authunix_parms *aup, sfsprog *p);
    ~sfsctl ();
    void setpid (int32_t npid);
    void dispatch (svccb *sbp);
  };
  friend class sfsctl;
  ihash2<const sfs_aid, int32_t, sfsctl, &sfsprog::sfsctl::aid,
	 &sfsprog::sfsctl::pid, &sfsprog::sfsctl::hlink> ctltab;

  virtual ~sfsprog () { panic ("sfsprog deleted\n"); }
  virtual void mountcb (svccb *sbp, const nfs_fh3 *fhp);
  void cddispatch (svccb *sbp);
  void ctlaccept (ptr<axprt_unix> x, const authunix_parms *aup);
  void linkdispatch (nfscall *nc);
  void tmosched (bool expired = false);
  void sockcheck ();

public:
  typedef void (*allocfn_t) (sfsprog *, ref<nfsserv>, int,
			     sfscd_mountarg *, sfsserver::fhcb);

  const allocfn_t newserver;
  const bool needclose;
  const bool mntwithclose;
  ref<axprt_unix> x;
  ref<aclnt> cdc;
  ref<asrv> cds;
  ref<nfsserv_udp> ns;
  ref<nfsdemux> nd;
  ref<nfsserv> linkserv;
  timecb_t *idletmo;
  ihash<const str, sfsserver,
    &sfsserver::path, &sfsserver::pathlink> pathtab;
  tailq<sfsserver, &sfsserver::idlelink> idleq;

  sfsprog (ref<axprt_unix> cdx, allocfn_t f, bool needclose = false,
           bool mntwithclose = false);
  virtual bool intercept (sfsserver *s, nfscall *nc);
};

#endif /* _SFSMISC_SFSCLIENT_H_ */
