// -*-c++-*-
/* $Id: dnsimpl.h 1117 2005-11-01 16:20:39Z max $ */

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

#ifndef _ASYNC_DNSIMPL_H_
#define _ASYNC_DNSIMPL_H_ 1

#include "dnsparse.h"
#include "ihash.h"
#include "backoff.h"

class resolver;
class dnsreq {
  int ntries;
  int srchno;

public:
  resolver *const resp;
  bool usetcp;
private:
  bool constructed;
  bool intable;
protected:
  void remove ();
public:
  int error;
  u_int16_t id;			// DNS query ID
  str basename;			// Name for which to search
  str name;			// Name for which to query
  u_int16_t type;		// Type of query (T_A, T_PTR, etc.)
  ihash_entry<dnsreq> hlink;	// Per-id hash table link
  tmoq_entry<dnsreq> tlink;	// Retransmit queue link

  dnsreq (resolver *, str, u_int16_t, bool search = false);
  virtual ~dnsreq ();
  void start (bool);
  void xmit (int = 0);
  virtual void readreply (dnsparse *) = 0;
  void timeout ();
  void fail (int);
};

class dnsreq_a : public dnsreq {
  bool checkaddr;		// Non-zero when arr_addr must be checked
  in_addr addr;			// Adress of inverse queries (for checking)
  cbhent cb;			// Callback for hostbyname/addr
  dnsreq_a ();
public:
  dnsreq_a (resolver *rp, str n, cbhent c, bool s = false)
    : dnsreq (rp, n, T_A, s), checkaddr (false), cb (c) {}
  dnsreq_a (resolver *rp, str n, cbhent c, const in_addr &a)
    : dnsreq (rp, n, T_A), checkaddr (true), addr (a), cb (c) {}
  void readreply (dnsparse *);
};

class dnsreq_mx : public dnsreq {
  cbmxlist cb;
  dnsreq_mx ();
public:
  dnsreq_mx (resolver *rp, str n, cbmxlist c, bool s)
    : dnsreq (rp, n, T_MX, s), cb (c) {}
  void readreply (dnsparse *);
};

class dnsreq_srv : public dnsreq {
  cbsrvlist cb;
  dnsreq_srv ();
public:
  dnsreq_srv (resolver *rp, str n, cbsrvlist c, bool s)
    : dnsreq (rp, n, T_SRV, s), cb (c) {}
  void readreply (dnsparse *);
};

class dnsreq_ptr : public dnsreq {
  in_addr addr;
  cbhent cb;			// Callback for hostbyname/addr

  int napending;
  vec<str, 2> vnames;
  vec<dnsreq_a *> vrfyv;

  static void maybe_push (vec<str, 2> *sv, const char *s);
public:
  static str inaddr_arpa (in_addr);
  dnsreq_ptr (resolver *rp, in_addr a, cbhent c)
    : dnsreq (rp, inaddr_arpa (a), T_PTR), addr (a), cb (c) {}
  ~dnsreq_ptr ();
  void readreply (dnsparse *);
  void readvrfy (int i, ptr<hostent> h, int err);
};

class dnsreq_txt : public dnsreq {
  cbtxtlist cb;
public:
  dnsreq_txt (resolver *rp, str n, cbtxtlist c, bool s = false)
    : dnsreq (rp, n, T_TXT, s), cb (c) {}
  void readreply (dnsparse *);
};

class dnssock {
public:
  typedef callback<void, u_char *, ssize_t>::ref cb_t;
protected:
  cb_t cb;
  ref<bool> destroyed;
public:
  const bool reliable;
  dnssock (bool r, cb_t c)
    : cb (c), destroyed (New refcounted<bool> (false)),
      reliable (r) {}
  virtual ~dnssock () { *destroyed = false; }
  virtual void sendpkt (const u_char *pkt, size_t size) = 0;
};

class dnssock_udp : public dnssock {
  int fd;
  void rcb ();
public:
  dnssock_udp (int f, cb_t cb);
  ~dnssock_udp ();
  void sendpkt (const u_char *pkt, size_t size);
};

class dnssock_tcp : public dnssock {
  int fd;
  bool write_ok;
  dnstcppkt tcpstate;
  void rcb ();
  void wcb (bool selected = false);
public:
  dnssock_tcp (int f, cb_t cb);
  ~dnssock_tcp ();
  void sendpkt (const u_char *pkt, size_t size);
};

class resolver {
protected:
  ptr<dnssock> udpsock;
  ptr<dnssock> tcpsock;
  int nbump;			// # of bumpsocks since last good reply
  sockaddr *addr;
  socklen_t addrlen;
  dnsreq *udpcheck_req;

  virtual bool bumpsock (bool failure) = 0;
  bool udpinit ();
  bool tcpinit ();
  void cantsend ();
  bool resend (bool udp, bool tcp);
  static void failreq (int err, dnsreq *r) { r->fail (err); }
  void pktready (bool tcp, u_char *qb, ssize_t size);
  void udpcheck_cb (ptr<hostent> h, int err);
public:
  time_t last_resp;		// Last time of valid reply from this server
  time_t last_bump;
  ref<bool> destroyed;
  ihash<u_int16_t, dnsreq, &dnsreq::id, &dnsreq::hlink> reqtab;
  tmoq<dnsreq, &dnsreq::tlink, 1, 5> reqtoq;

  resolver ();
  virtual ~resolver ();
  bool setsock (bool failure);
  void sendreq (dnsreq *r);
  virtual const char *srchlist (int n) { return n <= 0 ? "" : NULL; }
  u_int16_t genid ();
  void udpcheck ();
};

class resolv_conf : public resolver {
protected:
  int ns_idx;
  sockaddr_in srvaddr;
  time_t last_reload;
  bool reload_lock;
  ifchgcb_t *ifc;
  ref<bool> destroyed;

  void reload (bool failure);
  static void reload_dumpres (int fd);
  void reload_cb (ref<bool> d, bool ifchange, str newres);

protected:
  bool bumpsock (bool failure);

public:
  resolv_conf ();
  ~resolv_conf ();
  const char *srchlist (int n);
};

#endif /* _ASYNC_DNSIMPL_H_ */
