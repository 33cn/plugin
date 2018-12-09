// -*-c++-*-
/* $Id: dnsparse.h 1117 2005-11-01 16:20:39Z max $ */

/*
 *
 * Copyright (C) 1998-2003 David Mazieres (dm@uun.org)
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


#ifndef _DNSPARSE_H_
#define _DNSPARSE_H_ 1

#include "dns.h"
#include "qhash.h"

#define QBSIZE 512

#ifndef T_DNAME
# define T_DNAME 39
#endif /* !T_DNAME */
#ifndef T_SRV
# define T_SRV 33
#endif /* !T_SRV */

#ifndef MAXDNAME
# define MAXDNAME 1025
#endif /* !MAXDNAME */

class dnstcppkt {
  u_int inbufsize;
  u_int inbufpos;
  u_int inbufused;
  u_char *inbuf;
  suio outbuf;
  void compact ();
  u_int pktsize ()
    { return inbufpos < inbufused + 2 ? 2 : getshort (inbuf + inbufused) + 2; }
public:
  dnstcppkt ();
  ~dnstcppkt ();
  void reset ();
  int input (int fd);		// Returns size of packet, 0 incomplete, -1 err
  int output (int fd);		// 1 drained, 0 need again, -1 err
  bool getpkt (u_char **pp, size_t *sp);
  void putpkt (const u_char *p, size_t s);
};

class nameset {
  typedef qhash<str, u_int> map_t;
  u_int pos;
  map_t name2pos;
public:
  nameset () : pos (0) {}
  char *store (str s);
  char *lookup (str s) const;
  bool present (str s) const { return name2pos[s]; }
  static char *xlat (char *base, char *p) { return base + (p - (char *) 0); }
  u_int size () const { return pos; }
  void put (char *dst) const;
};

struct resrec {
  struct rd_mx {
    u_int16_t mx_pref;
    char mx_exch[MAXDNAME];
  };
  struct rd_soa {
    char soa_mname[MAXDNAME];
    char soa_rname[MAXDNAME];
    u_int32_t soa_serial;
    u_int32_t soa_refresh;
    u_int32_t soa_retry;
    u_int32_t soa_expire;
    u_int32_t soa_minimum;
  };
  struct rd_srv {
    u_int16_t srv_prio;
    u_int16_t srv_weight;
    u_int16_t srv_port;
    char srv_target[MAXDNAME];
  };

  char rr_name[MAXDNAME];
  u_int16_t rr_class;
  u_int16_t rr_type;
  u_int32_t rr_ttl;
  u_int16_t rr_rdlen;
  union {
    char rr_ns[MAXDNAME];
    in_addr rr_a;
    char rr_cname[MAXDNAME];
    rd_soa rr_soa;
    char rr_ptr[MAXDNAME];
    rd_mx rr_mx;
    char rr_txt[sizeof (rd_soa)];
    rd_srv rr_srv;
  };
};

struct question {
  char q_name[MAXDNAME];
  u_int16_t q_class;
  u_int16_t q_type;
};

class dnsparse {
  dnsparse ();

  const u_char *const buf;
  const u_char *const eom;
  const u_char *anp;

  static int mxrec_cmp (const void *, const void *);
  static int srvrec_cmp (const void *, const void *);
  static void srvrec_randomize (srvrec *base, srvrec *last);
  static size_t hintsize (u_int nhints)
    { return nhints * sizeof (addrhint) + (nhints+1) * sizeof (addrhint *); }
  static addrhint **puthints (char *dst, const vec<addrhint> &hv,
			     char *namebase);
  bool gethints (vec<addrhint> *hv, const nameset &nset);

public:
  int error;
  const HEADER *const hdr;
  const u_int ancount;
  const u_int nscount;
  const u_int arcount;

  dnsparse (const u_char *buf, size_t len, bool answer = true);

  const u_char *getqp () { return hdr ? buf + sizeof (HEADER) : NULL; }
  const u_char *getanp () { return anp; }

  bool qparse (question *);
  bool qparse (const u_char **, question *);
  bool rrparse (const u_char **, resrec *);

  bool skipnrecs (const u_char **, u_int);

  ptr<hostent> tohostent ();
  ptr<mxlist> tomxlist ();
  ptr<srvlist> tosrvlist ();
  ptr<txtlist> totxtlist ();
};

#endif /* !_DNSPARSE_H_ */
