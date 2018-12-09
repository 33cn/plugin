// -*-c++-*-
/* $Id: dns.h 1117 2005-11-01 16:20:39Z max $ */

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


#ifndef _DNS_H_
#define _DNS_H_ 1

#include "async.h"

extern "C" {
#define class rr_class
#if HAVE_ARPA_NAMESER_COMPAT_H
#include <arpa/nameser_compat.h>
#include <arpa/nameser.h>
#else /* !HAVE_ARPA_NAMESER_COMPAT_H */
#include <arpa/nameser.h>
#endif /* !HAVE_ARPA_NAMESER_COMPAT_H */
#undef class
#include <resolv.h>

/* Declarations missing on some OS's */
#ifdef NEED_RES_INIT_DECL
void res_init ();
#endif /* NEED_RES_INIT_DECL */
#ifdef NEED_RES_MKQUERY_DECL
int res_mkquery (int, const char *, int, int,
		 const u_char *, int, const u_char *,
		 u_char *, int);
#endif /* NEED_RES_MKQUERY_DECL */
}

struct addrhint {
  char *h_name;
  int h_addrtype;
  int h_length;
  char h_address[16];
};

struct hostent;

struct mxrec {
  u_int16_t pref;
  char *name;
};
struct mxlist {
  char *m_name;		    /* Name of host for which MX list taken */
  struct addrhint **m_hints;
  u_short m_nmx;		/* Number of mx records */
  struct mxrec m_mxes[1];	/* MX records */
};

struct srvrec {
  u_int16_t prio;
  u_int16_t weight;
  u_int16_t port;
  char *name;
};
struct srvlist {
  char *s_name;
  struct addrhint **s_hints;
  u_short s_nsrv;
  struct srvrec s_srvs[1];
};

struct txtlist {
  char *t_name;
  u_short t_ntxt;
  char *t_txts[1];
};

/* Extender error types for ar_errno */
#define ARERR_NXREC 0x101	/* No records of appropriate type */
#define ARERR_TIMEOUT 0x102	/* Query timed out */
#define ARERR_PTRSPOOF 0x103	/* PTR response was a lie! */
#define ARERR_BADRESP 0x104	/* Nameserver replied with malformed packet */
#define ARERR_CANTSEND 0x105	/* Can't send to name server */
#define ARERR_REQINVAL 0x106	/* Request was for malformed domain name */
#define ARERR_CNAMELOOP 0x107   /* CNAME records form loop */

typedef struct dnsreq dnsreq_t;
void dnsreq_cancel (dnsreq_t *rqp);

typedef callback<void, ptr<hostent>, int>::ref cbhent;
dnsreq_t *dns_hostbyname (str, cbhent,
			  bool search = false, bool addrok = true);
dnsreq_t *dns_hostbyaddr (const in_addr, cbhent);

typedef callback<void, ptr<mxlist>, int>::ref cbmxlist;
dnsreq_t *dns_mxbyname (str, cbmxlist, bool search = false);

typedef callback<void, ptr<srvlist>, int>::ref cbsrvlist;
dnsreq_t *dns_srvbyname (str name, cbsrvlist, bool search = false);

typedef callback<void, ptr<txtlist>, int>::ref cbtxtlist;
dnsreq_t *dns_txtbyname (str name, cbtxtlist cb, bool search = false);

inline dnsreq_t *
dns_srvbyname (const char *name, const char *proto, const char *srv,
	       cbsrvlist cb, bool search = false)
{
  return dns_srvbyname (strbuf ("_%s._%s.%s", srv, proto, name), cb, search);
}

const char *dns_strerror (int);
int dns_tmperr (int);

void printaddrs (const char *, ptr<hostent>, int = 0);
void printmxlist (const char *, ptr<mxlist>, int = 0);
void printsrvlist (const char *msg, ptr<srvlist> s, int = 0);
void printtxtlist (const char *msg, ptr<txtlist> s, int = 0);

void dns_reload ();

#endif /* !_DNS_H_ */

