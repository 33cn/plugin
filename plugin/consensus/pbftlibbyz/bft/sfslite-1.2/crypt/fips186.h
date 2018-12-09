
#ifndef _FIPS186_H_
#define _FIPS186_H_

#include <assert.h>

#include "crypt.h"
#include "bigint.h"
#include "sha1.h"

/* 
 *  This algorithm is based on the Standard published in FIPS PUB 186-2.
 *  First it generates the group parameters p, q and g such that p is an
 *  n-bit prime, q is a 160-bit prime dividing p-1, and g is a random
 *  element of Z_p^* of order q.
 */
struct fips186_gen {
  virtual void gen (u_int iter)
  { fatal ("Don't instantiate the fips186 class directly\n"); }
  fips186_gen (u_int pbits, u_int iter);
  virtual ~fips186_gen () { if (seed) delete [] seed; delete [] raw_p; }

  u_int64_t *seed;
  u_int seedsize;

protected:
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

#endif /* !_FIPS186_H_ */
