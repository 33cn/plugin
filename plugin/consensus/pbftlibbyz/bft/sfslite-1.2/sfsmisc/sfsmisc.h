// -*-c++-*-
/* $Id: sfsmisc.h 2742 2007-04-16 21:03:33Z max $ */

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

#ifndef _SFSMISC_H_
#define _SFSMISC_H_ 1

#include "amisc.h"

#ifdef HAVE_SFSMISC
#include "nfs3_prot.h"
#include "sfs_prot.h"
#endif /* HAVE_SFSMISC */

struct svccb;
struct aclnt;
struct asrv;
struct rabin_priv;
struct axprt_crypt;
struct sfspriv;

#ifdef HAVE_SFSMISC
/* nfs3_err.C */
extern void nfs3_err (svccb *sbp, nfsstat3 status);
extern void nfs3exp_err (svccb *sbp, nfsstat3 status);
const strbuf &strbuf_cat (const strbuf &sb, nfsstat3 err);

/* sfspath.C */
enum { ascii_hostid_len = (sizeof (sfs_hash) * 8 + 4) / 5 };
bool sfs_ascii2hostid (sfs_hash *hid, const char *p);
bool sfsgethost_label (const char *&p);
bool sfsgethost_dotlabel (const char *&p);
str sfsgethost (const char *&p, bool qualified = true);
bool sfsgetlocation (const char *&pp, str *hostp = NULL,
		     u_int16_t *portp = NULL, bool qualified = true);
bool sfsgetatlocation (const char *&pp, str *hostp = NULL,
		       u_int16_t *portp = NULL, bool qualified = true);
bool sfs_parsepath (str path, str *host = NULL, sfs_hash *hostid = NULL, 
		    u_int16_t *portp = NULL, int *vers = NULL);
bool sfs_parsepath_v1 (str path, str *host, sfs_hash *hostid,
		       u_int16_t *portp);
bool sfs_parsepath_v2 (str path, str *host, sfs_hash *hostid, 
		       u_int16_t *portp);



class sfs_servinfo_w {
public:
  static ptr<sfs_servinfo_w> alloc (const sfs_servinfo &s);
  static ref<sfs_servinfo_w> alloc (const sfs_hostinfo2 &h);
  sfs_servinfo_w (const sfs_servinfo &s) : si (s) {}
  sfs_servinfo get_xdr () const { return si; }
  virtual ~sfs_servinfo_w () {}

  bool mkhostid_client (sfs_hash *h) const { return mkhostid (h, get_vers ());}
  bool mkhostid (sfs_hash *h, int vers = 2) const;
  bool ckpath (const str &path) const;
  str mkpath (int vers = 2, int port = -1) const;
  str mkpath_client () const { return mkpath (get_vers ()); }
  bool operator== (const sfs_servinfo_w &s) const;
  bool ckci (const sfs_connectinfo &ci) const;
  bool ckhostid (const sfs_hash *id, int vers = 2) const;
  bool ckhostid_client (const sfs_hash *id) const 
  { return ckhostid (id, get_vers ()); }

  virtual bigint get_rabin_pubkey () const = 0;
  virtual int get_vers () const = 0;
  virtual str get_hostname () const = 0;
  virtual int get_port () const = 0;
  virtual sfs_pubkey2 get_pubkey () const = 0;
  virtual int get_relno () const = 0;
  virtual int get_progno () const = 0;
  virtual int get_versno () const = 0;
protected:
  const sfs_servinfo si;
private:
  bool mkhostid_v1 (sfs_hash *id) const;
  bool mkhostid_v2 (sfs_hash *id) const;
};

class sfs_servinfo_w_v2 : public sfs_servinfo_w {
public:
  sfs_servinfo_w_v2 (const sfs_servinfo &s) : sfs_servinfo_w (s) {}
  int get_vers () const { return 2; }
  str get_hostname () const { return si.cr7->host.hostname; }
  int get_port () const { return si.cr7->host.port; }
  sfs_pubkey2 get_pubkey () const { return si.cr7->host.pubkey; }
  bigint get_rabin_pubkey () const ;
  int get_relno () const { return si.cr7->release ; }
  int get_progno () const { return si.cr7->prog; }
  int get_versno () const { return si.cr7->vers; }
};

class sfs_servinfo_w_v1 : public sfs_servinfo_w {
public:
  sfs_servinfo_w_v1 (const sfs_servinfo &s) : sfs_servinfo_w (s) {}
  int get_vers () const { return 1; }
  str get_hostname () const { return si.cr5->host.hostname; }
  int get_port () const { return 0; }
  sfs_pubkey2 get_pubkey () const ;
  bigint get_rabin_pubkey () const { return si.cr5->host.pubkey; }
  int get_relno () const { return si.sivers ; }
  int get_progno () const { return si.cr5->prog; }
  int get_versno () const { return si.cr5->vers; }
};

class sfs_pathrevoke_w {
public:
  sfs_pathrevoke_w (const sfs_pathrevoke &r);
  ~sfs_pathrevoke_w ();
  bool check (sfs_hash *p = NULL);

  const sfs_pathrevoke rev;
  ptr<sfs_servinfo_w> si;
  ptr<sfs_servinfo_w> rsi;
};

/* sfs_err.C */
const strbuf &strbuf_cat (const strbuf &sb, sfsstat err);
const strbuf &strbuf_cat (const strbuf &sb, sfsauth_stat status);

#endif /* HAVE_SFSMISC */

/* sfsconst.C */
extern u_int32_t sfs_release;
extern u_int16_t sfs_defport;
extern uid_t sfs_uid;
extern gid_t sfs_gid;
extern uid_t nobody_uid;
extern gid_t nobody_gid;
extern u_int32_t sfs_resvgid_start;
extern u_int sfs_resvgid_count;
#ifdef MAINTAINER
extern bool runinplace;
#else /* !MAINTAINER */
enum { runinplace = false };
#endif /* !MAINTAINER */
extern const char *sfsroot;
extern str sfsdir;
extern str sfssockdir;
extern str sfsdevdb;
extern const char *etc1dir;
extern const char *etc2dir;
extern const char *etc3dir;
extern const char *sfs_authd_syslog_priority;
extern u_int sfs_rsasize;
extern u_int sfs_dlogsize;
extern u_int sfs_mindlogsize;
extern u_int sfs_maxdlogsize;
extern u_int sfs_minrsasize;
extern u_int sfs_maxrsasize;
extern u_int sfs_pwdcost;
const u_int sfs_maxpwdcost = 32;
extern u_int sfs_hashcost;
extern u_int sfs_maxhashcost;

void sfsconst_init (bool lite_mode = false);
str sfsconst_etcfile (const char *name);
str sfsconst_etcfile (const char *name, const char *const *path);
str sfsconst_etcfile_required (const char *name);
str sfsconst_etcfile_required (const char *name, const char *const *path,
			       bool ftl = true);
void mksfsdir (str path, mode_t mode,
	       struct stat *sbp = NULL, uid_t uid = sfs_uid);
str sfshostname ();

#ifdef HAVE_SFSMISC

/* sfsaid.C */
extern const bool sfsaid_shift;
typedef u_int64_t sfs_aid;
extern const sfs_aid sfsaid_sfs;
extern const sfs_aid sfsaid_nobody;
bool sfs_specaid (sfs_aid);
sfs_aid sfs_mkaid (u_int32_t uid, u_int32_t gid);
sfs_aid aup2aid (const authunix_parms *aup);
sfs_aid myaid ();

/* validshell.C */
bool validshell (const char *shell);

/* suidgetfd.C */
int suidgetfd (str prog);
int suidgetfd_required (str prog);

/* unixserv.C */
struct axprt_unix;
typedef callback<void, ptr<axprt_unix>,
                 const authunix_parms *>::ref suidservcb;
void sfs_unixserv (str sock, cbi cb, mode_t = 0600);
void sfs_suidserv (str prog, suidservcb cb);

#endif /* HAVE_SFSMISC */

#include "keyfunc.h"

template<> struct hashfn<vec<str> > {
  hashfn () {}
  hash_t operator() (const vec<str> &v) const {
    u_int val = HASHSEED;
    for (const str *sp = v.base (); sp < v.lim (); sp++)
      val = hash_bytes (sp->cstr (), sp->len (), val);
    return val;
  }
};
template<> struct equals<vec<str> > {
  equals () {}
  bool operator() (const vec<str> &a, const vec<str> &b) const {
    size_t n = a.size ();
    if (n != b.size ())
      return false;
    while (n-- > 0)
      if (a[n] != b[n])
	return false;
    return true;
  }
};

/* pathexpand.C */
int path2sch (str path, str *sch);

#ifdef HAVE_SFSMISC
/* here to avoid circular dependecies; needed by sfscrypt.h */
typedef callback<void, str, ptr<sfs_sig2> >::ref cbsign;
#else
void rndbkd (const str &msg = NULL);
#endif /*  HAVE_SFSMISC */


#endif /* _SFSMISC_H_ */
