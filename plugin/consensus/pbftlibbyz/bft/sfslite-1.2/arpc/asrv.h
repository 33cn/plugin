// -*-c++-*-
/* $Id: asrv.h 3714 2008-10-14 15:00:36Z max $ */

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

class asrv;

struct progvers {
  const u_int32_t prog;
  const u_int32_t vers;
  progvers (u_int32_t p, u_int32_t v) : prog (p), vers (v) {}
  operator hash_t() const { return prog | hash_rotate (vers, 20); }
  bool operator== (const progvers &a) const
    { return prog == a.prog && vers == a.vers; }
};

class svccb {
  friend class asrv;
  friend class asrv_replay;
  friend class asrv_unreliable;
  friend class asrv_resumable;
  friend class auto_ptr<svccb>;

  rpc_msg msg;			// RPC message
  void *arg;			// Argument to call
  mutable authunix_parms *aup;

  ptr<asrv> srv;
  sockaddr *addr;		// Address to reply to
  size_t addrlen;		// Length of address (for comparing)

  void *resdat;			// Unmarshaled result (if convenient)

  void *res;			// fields for replay cache
  size_t reslen;

  u_int64_t offset;             // Byte offset in underlying transport

  timespec ts_start;            // keep track of when it started

  svccb (const svccb &);	// No copying
  const svccb &operator= (const svccb &);

  void init (asrv *, const sockaddr *);

protected:
  svccb ();
  virtual ~svccb ();

public:
  tailq_entry<svccb> qlink;
  ihash_entry<svccb> hlink;

  u_int hash_value () const;
  bool operator== (const svccb &a) const;
  operator hash_t () const { return hash_value (); }

  u_int32_t xid () const { return msg.rm_xid; }
  u_int32_t prog () const { return msg.rm_call.cb_prog; }
  u_int32_t vers () const { return msg.rm_call.cb_vers; }
  u_int32_t proc () const { return msg.rm_call.cb_proc; }

  const ptr<asrv> &getsrv () const { return srv; }

  void *getvoidarg () { return arg; }
  const void *getvoidarg () const { return arg; }
  template<class T> T *getarg () { return static_cast<T *> (arg); }
  template<class T> const T *getarg () const { return static_cast<T *> (arg); }

  /* These return a properly initialized structure of the return type
   * of the RPC call.  The structure is automatically deleted when
   * with svccb, which may just happen to be handy. */
  void *getvoidres ();
  template<class T> T *getres () { return static_cast<T *> (getvoidres ()); }

  const opaque_auth *getcred () const { return &msg.rm_call.cb_cred; }
  const opaque_auth *getverf () const { return &msg.rm_call.cb_verf; }
  const authunix_parms *getaup () const;
  u_int32_t getaui () const;
  const sockaddr *getsa () const { return addr; }
  bool fromresvport () const;

  void reply (const void *, sfs::xdrproc_t = NULL, bool nocache = false);
  template<class T> void replyref (const T &res, bool nocache = false)
    { reply (&res, NULL, nocache); }
  void replyref (const int &res, bool nocache = false)
    { u_int32_t val = res; reply (&val, NULL, nocache); }

  void reject (auth_stat);
  void reject (accept_stat);
  void ignore ();
};


class asrv : public virtual refcount {
  friend class svccb;
protected:
  typedef callback<void, svccb *>::ref asrv_cb;

  const rpc_program *const rpcprog;
  const rpcgen_table *const tbl;
  const u_int32_t nproc;

private:
  asrv_cb::ptr cb;

  cbv::ptr recv_hook;

  static void seteof (ref<xhinfo>, const sockaddr *, bool force = false);

protected:
  ptr<xhinfo> xi;

  asrv (ref<xhinfo>, const rpc_program &, asrv_cb::ptr);
  virtual ~asrv ();
  virtual bool isreplay (svccb *) { return false; }
  virtual void sendreply (svccb *, xdrsuio *, bool nocache);
  virtual void inc_svccb_count () {}
  virtual void dec_svccb_count () {}

public:
  const progvers pv;
  ihash_entry<asrv> xhlink;
  const ref<axprt> &xprt () const;

  void start ();
  void stop ();

  virtual void setcb (asrv_cb::ptr c);
  bool hascb () { return cb; }

  void set_recv_hook (cbv::ptr cb) { recv_hook = cb; }

  static void dispatch (ref<xhinfo>, const char *, ssize_t, const sockaddr *);
  static ptr<asrv> alloc (ref<axprt>, const rpc_program &,
			  asrv_cb::ptr = NULL);
};

class asrv_replay : public asrv {
protected:
  u_int rsize;
  tailq<svccb, &svccb::qlink> rq;
  shash<svccb, &svccb::hlink> rtab;

  void delsbp (svccb *);

  svccb *lookup (svccb *);
  void sendreply (svccb *sbp, xdrsuio *, bool nocache);

  asrv_replay (ref<xhinfo> x, const rpc_program &rp, asrv_cb::ptr cb)
    : asrv (x, rp, cb), rsize (0) {}
  ~asrv_replay ();
};

//
// An asrv class that masks an EOF until all outstanding svccb's
// have been acted on. Semantics are:
//   - On an EOF, if no svccb's outstanding, then callback the asrv_cb
//     with NULL.
//   - On an EOF, if svccb's are outstanding, then callback the eofcb.
//      - Once all svccb's have been replied to (and replies are swallowed
//        after EOFs), then call the asrv_cb with NULL.
//
// Thus, the callback possibilities are either an eofcb, and later a 
// asrv_cb (NULL), or just an asrv_cb (NULL).
//  
// 
class asrv_delayed_eof : public asrv {
private:
  int _count;
  bool _eof;
  asrv_cb::ptr _asrv_cb;
  cbv::ptr _eofcb;

protected:
  asrv_delayed_eof (ref<xhinfo>, const rpc_program &, asrv_cb::ptr, 
		    cbv::ptr eofcb);
  void dispatch (svccb *sbp);
  void inc_svccb_count () { _count ++; }
  void dec_svccb_count ();
  void sendreply (svccb *, xdrsuio *, bool nocache) ;
  
public:
  static ptr<asrv_delayed_eof> 
  alloc (ref<axprt>, const rpc_program &, 
	 asrv_cb::ptr cb = NULL,
	 cbv::ptr eofcb = NULL);
  bool is_eof () const { return _eof; }
  void setcb (asrv_cb::ptr cb);
};

class asrv_unreliable : public asrv_replay {
  const u_int maxrsize;

  bool isreplay (svccb *sbp);
  void sendreply (svccb *sbp, xdrsuio *, bool nocache);

protected:
  asrv_unreliable (ref<xhinfo> x, const rpc_program &rp,
		   asrv_cb::ptr cb, u_int rs = 16)
    : asrv_replay (x, rp, cb), maxrsize (rs) {}
};

class asrv_resumable : public asrv_replay {
  bool isreplay (svccb *sbp);
  void sendreply (svccb *sbp, xdrsuio *, bool nocache);

protected:
  asrv_resumable (ref<xhinfo> x, const rpc_program &rp, asrv_cb::ptr cb)
    : asrv_replay (x, rp, cb) {}

public:
  bool resume (ref<axprt>);

  static ptr<asrv_resumable> alloc (ref<axprt>, const rpc_program &,
                                    asrv_cb::ptr);
};

ptr<asrv> asrv_alloc (ref<axprt>, const rpc_program &,
    		      callback<void, svccb *>::ptr, bool);

str sock2str (const struct sockaddr *sp);
#ifdef MAINTAINER
void set_asrvtrace (int l);
int get_asrvtrace (void);
void set_asrvtime (bool b);
bool get_asrvtime (void);
#endif /* MAINTAINER */
