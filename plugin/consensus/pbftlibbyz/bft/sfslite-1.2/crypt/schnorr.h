// -*-c++-*-
// The above line tells emacs that this header likes C++

#ifndef _SCHNORR_H_
#define _SCHNORR_H_

#include <assert.h>

#include "crypt.h"
#include "bigint.h"
#include "sha1.h"


class ephem_key_pair {
protected:
  bigint k, r;

public:
  const bigint &private_half() const { return k; }
  const bigint &public_half()  const { return r; }

  ephem_key_pair (const bigint &kk, const bigint &rr)
    : k (kk), r (rr) { }
};

class schnorr_pub {
protected:
  const bigint p;
  const bigint q;
  const bigint g;
  const bigint y;

protected:
  bool is_group_elem (const bigint &elem) const
  { return powm (elem, q, p) == 1; }

  void random_group_log (bigint *log) const
  { assert (log != NULL); *log = random_bigint (q.nbits () - 1); }

  void elem_from_log (bigint *elem, const bigint &log) const
  { assert (elem != NULL); *elem = powm (g, log, p); }

  void bind_r_to_m (bigint *e, const str &m, const bigint &r) const;

  bool check_signature (const bigint &r, const bigint &s,
			const bigint &e, const bigint &y_v) const {
    bigint gs (powm (g, s, p)), 
           ye (powm (y_v, e, p));
    bigint should_be_gs (r * ye);

    should_be_gs %= p;

    return gs == should_be_gs;
  }

public:
  schnorr_pub (const bigint &pp, const bigint &qq,
	       const bigint &gg, const bigint &yy)
    : p (pp), q (qq), g (gg), y (yy) {}
  virtual ~schnorr_pub () {} 

  const bigint &modulus    () const { return p; }
  const bigint &order      () const { return q; }
  const bigint &generator  () const { return g; }
  const bigint &public_key () const { return y; }

  virtual const bigint private_share () const { return 0; }

  ref<schnorr_pub> clone_schnorr_pub () const 
  { return New refcounted<schnorr_pub> (p, q, g, y); }

  bool verify (const str &msg, const bigint &r, const bigint &s) const {
    bigint e;    

    if  (is_group_elem (r) && (s > 0) && (s < q)) {
      bind_r_to_m (&e, msg, r);
      return check_signature (r, s, e, y);
    }
    else {
      return false;
    }
  }

  const ref<ephem_key_pair> make_ephem_key_pair () const {
    bigint log, elem;

    random_group_log (&log);
    elem_from_log (&elem, log);

    return New refcounted<ephem_key_pair> (log, elem);
  }
};

class schnorr_clnt_priv : public schnorr_pub {
public:
  const bigint x_clnt;     	  /* first of two additive shares of
				     the Discrete Log of the y field */

public:
  schnorr_clnt_priv (const bigint &pp, const bigint &qq, const bigint &gg,
		     const bigint &yy, const bigint &x1)
    : schnorr_pub (pp, qq, gg, yy), x_clnt (x1) { }

  const bigint private_share () const { return x_clnt; }

  bool complete_signature (bigint *r, bigint *s, const str &msg,
			   const bigint &r_clnt, const bigint &k_clnt, 
			   const bigint &r_srv, const bigint &s_srv);
  ptr<schnorr_clnt_priv> update (bigint *delta) const;
};

class schnorr_srv_priv : public schnorr_pub {
public:
  const bigint x_srv;     	  /* second of two additive shares of
				     the Discrete Log of the y field */

public:
  schnorr_srv_priv (const bigint &pp, const bigint &qq, const bigint &gg,
		    const bigint &yy, const bigint &x2)
    : schnorr_pub (pp, qq, gg, yy), x_srv (x2) {}

  const bigint private_share () const { return x_srv; }

  bool endorse_signature (bigint *r_srv, bigint *s_srv,
				    const str &msg, const bigint &r_clnt);
  ptr<schnorr_srv_priv> update (const bigint &delta) const;

};

class schnorr_priv : public schnorr_pub {
public:
  schnorr_priv (const bigint &pp, const bigint &qq, const bigint &gg,
		const bigint &yy, const bigint &xx) 
    : schnorr_pub (pp, qq, gg, yy), x (xx), ekp (make_ephem_key_pair ()) {}

  bool sign (bigint *r, bigint *s, const str &msg);
  const bigint private_share () const { return x; }
private:
  const bigint x;
  void make_ekp ();
  ptr<ephem_key_pair> ekp;
};

/* 
 *  This algorithm is based on the Standard published in FIPS PUB 186-2.
 *  First it generates the group parameters p, q and g such that p is an
 *  n-bit prime, q is a 160-bit prime dividing p-1, and g is a random
 *  element of Z_p^* of order q.
 *  Then it chooses two random elements x_c, and x_s \in Z_q, and
 *  computes the two schnorr shares.
 */
struct schnorr_gen {
  static ptr<schnorr_gen> rgen (u_int pbits, u_int iter = 32);
  void gen (u_int iter);
  schnorr_gen (u_int p);
  ~schnorr_gen () { if (seed) delete [] seed; delete [] raw_p; }

  ptr<schnorr_clnt_priv> csk;
  ptr<schnorr_srv_priv> ssk;
  ptr<schnorr_priv> wsk;
  u_int64_t *seed;
  u_int seedsize;

private:
  bool gen_p (bigint *p, const bigint &q, u_int iter);
  void gen_q (bigint *q);
  void gen_g (bigint *g, const bigint &p, const bigint &q);

  char *raw_p;
  u_int raw_psize;
  u_int num_p_hashes;
  u_int pbits;
  u_int pbytes;
  u_int num_p_candidates;
};

#endif /* !_SCHNORR_H_ */
