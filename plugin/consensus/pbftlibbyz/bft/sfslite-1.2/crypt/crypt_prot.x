/* $Id: crypt_prot.x 2330 2006-11-19 18:18:00Z max $ */

/*
 * This file was written by David Mazieres.  Its contents is
 * uncopyrighted and in the public domain.  Of course, standards of
 * academic honesty nonetheless prevent anyone in research from
 * falsely claiming credit for this work.
 */

%#include "bigint.h"


/*
 * These structures define the raw byte formats of messages exchanged
 * by the SRP protocol, as published in:
 *
 * T. Wu, The Secure Remote Password Protocol, in Proceedings of the
 * 1998 Internet Society Network and Distributed System Security
 * Symposium, San Diego, CA, Mar 1998, pp. 97-111.
 *
 *   sessid is a session identifier known by the user and server to be fresh
 *
 *   N is a prime number such that (N-1)/2 is also prime
 *   g is a generator of Z_N^*
 *
 *   x is a function of the user's password and salt
 *   v is g^x mod N
 *
 *   a is a random element of Z_N^* selected by the user (client)
 *   A = g^a mod N
 *
 *   b and u are random elements of Z_N^* picked by the server
 *   B = v + g^b mod N       (in version 3 of the protocol)
 *   B = 3v + g^b mod N      (in version 6 of the protocol)
 *
 *   S = g^{ab} * g^{xub}
 *   M = SHA-1 (sessid, N, g, user, salt, A, B, S)
 *   H = SHA-1 (sessid, A, M, S)
 *
 * The protocol proceeds as follows:
 *
 *   User -> Server:  username
 *   Server -> User:  salt, N, g
 *   User -> Server:  A
 *   Server -> User:  B, u
 *   User -> Server:  M
 *   Server -> User:  H
 *
 * After this, S can be used to generate secret session keys for use
 * between the user and server.
 */

/*
 * By default, the SRP code now uses an updated scheme designed to
 * prevent two-for-one password guessing by an active attacker
 * impersonating the server, as described in:
 *
 *   T. Wu, SRP-6: Improvements and Refinements to the Secure Remote
 *   Password Protocol, Submission to the IEEE P1363 Working Group,
 *   Oct 2002.
 *
 * The protocol is the same as the one above (called "SRP-3"), but a
 * constant k=3 is used to remove the symmetry in the calculation of B
 * by the server and of S by the client:
 *
 *   (server)   B = kv + g^b mod N
 *
 *   (client)   S = (B - kv)^(a + ux)
 *
 * The resulting value of S (as a function of all the other
 * parameters) is the same as before, and there is no change in the
 * sequence of messages exchanged.
 */

typedef opaque _srp_hash[20];

/* server to client */
struct srp_msg1 {
  string salt<>;
  bigint N;
  bigint g;
};

/* client to server */
struct srp_msg2 {
  bigint A;
};

/* server to client */
struct srp_msg3 {
  bigint B;
  bigint u;
};

/* hashed, then client to server */
struct srp_msg4_src {
  _srp_hash sessid;
  bigint N;
  bigint g;
  string user<>;
  string salt<>;
  bigint A;
  bigint B;
  bigint S;
};

/* hashed, then server to client */
struct srp_msg5_src {
  _srp_hash sessid;
  bigint A;
  _srp_hash M;
  bigint S;
};

#if 0
/* Info stored by server */
struct srp_info {
  bigint n;
  bigint g;
  string salt<>;
  bigint v;
};
#endif


enum crypt_keytype {
  CRYPT_NOKEY = 0,
  CRYPT_RABIN = 1,
  CRYPT_2SCHNORR = 2,          /* proactive 2-Schnorr -- private */
  CRYPT_SCHNORR = 3,           /* either *Schnorr -- public */
  CRYPT_1SCHNORR = 4,          /* standard 1-Schnorr -- private */
  CRYPT_ESIGN = 5,
  CRYPT_PAILLIER = 6,
  CRYPT_ELGAMAL = 7
};

struct elgamal_ctext {
  bigint r;
  bigint m;
};

union crypt_ctext switch (crypt_keytype type) {
 case CRYPT_RABIN:
   bigint rabin;
 case CRYPT_PAILLIER:
   bigint paillier;
 case CRYPT_ELGAMAL:
   elgamal_ctext elgamal;
 default:
   void;
};
