#include "fips186.h"

static const u_int HASHSIZE = sha1::hashsize;

#define DIV_ROUNDUP(p,q) ((p) / (q) + ((p) % (q) == 0 ? 0 : 1))

fips186_gen::fips186_gen (u_int p, u_int iter) : seed (NULL), pbits (p)
{
  pbytes = p >> 3;
  num_p_hashes = DIV_ROUNDUP (pbytes, HASHSIZE);
  raw_psize = num_p_hashes * HASHSIZE;
  raw_p = New char[raw_psize];
  num_p_candidates = pbits << 2;  // shouldn't fail -- 4x expected # of trials
  
  seedsize = DIV_ROUNDUP (HASHSIZE, 8) + 1; // 8 bytes in a u_int64_t
  seed = New u_int64_t[seedsize];
  for (u_int i = 0; i < seedsize; i++)
    seed[i] = rnd.gethyper ();
}

void
fips186_gen::gen_q (bigint *q)
{
  bigint u1, u2;
  char digest[HASHSIZE];
  do {
    sha1_hash (digest, seed, seedsize << 3); // seedsize * 8
    mpz_set_rawmag_le (&u1, digest, HASHSIZE);
    seed[3]++;
    sha1_hash (digest, seed, seedsize << 3); // seedsize * 8
    mpz_set_rawmag_le (&u2, digest, HASHSIZE);
    mpz_xor (q, &u1, &u2);
    mpz_setbit (q, (HASHSIZE << 3) - 1);     // set high bit
    mpz_setbit (q, 0);                       // set low bit
  } while (!q->probab_prime (5));
}

bool
fips186_gen::gen_p (bigint *p, const bigint &q, u_int iter)
{
  bigint X, c;
  for (u_int i = 0; i < num_p_candidates; i++) {
    for (u_int off = 0; off < raw_psize; off += HASHSIZE) {
      seed[0]++;
      sha1_hash (raw_p + off, seed, seedsize << 3); // seedsize * 8
    }
    mpz_set_rawmag_le (&X, raw_p, pbytes);
    mpz_setbit (&X, pbits - 1);
    c = X;
    c = mod (c, q * 2);
    *p = (X - c + 1);

    if (p->probab_prime (iter))
      return true;
  }
  return false;
}

void
fips186_gen::gen_g (bigint *g, const bigint &p, const bigint &q)
{
  bigint e = (p - 1) / q;
  bigint h;
  bigint p_3 = p - 3;
 
  do h = random_zn (p_3);
  while ((*g = powm (++h, e, p)) == 1);
}


