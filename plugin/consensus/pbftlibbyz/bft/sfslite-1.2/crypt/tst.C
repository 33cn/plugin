#if 0

#define USE_PCTR 0

#include "arpc.h"

int
main ()
{
  {
    xdrsuio x (XDR_ENCODE, true);
  }
  return 0;
}

#endif

#define USE_PCTR 0
#include "crypt.h"
#include "bench.h"

int
main (int argc, char **argv)
{
  random_update ();

#define HMAC(k, m)						\
do {								\
  u_char digest[sha1::hashsize];                                \
  sha1_hmac (digest, k, sizeof (k) - 1, m, sizeof (m) - 1);	\
  warn << "k = " << k << "\nm = " << m << "\n"			\
       << hexdump (digest, sizeof (digest)) << "\n";		\
} while (0)

#define HMAC2(k, k2, m)						\
do {								\
  u_char digest[sha1::hashsize];                                \
  sha1_hmac_2 (digest, k, sizeof (k) - 1, k2, sizeof (k2) - 1,	\
	       m, sizeof (m) - 1);				\
  warn << "k = " << k << "\nm = " << m << "\n"			\
       << hexdump (digest, sizeof (digest)) << "\n";		\
} while (0)

#if 0
  HMAC ("Jefe", "what do ya want for nothing?");
  HMAC ("\014\014\014\014\014\014\014\014\014\014\014\014\014\014\014\014\014\014\014\014", "Test With Truncation");
  //HMAC2 ("Je", "fe", "what do ya want for nothing?");
#endif

  bigint p ("c81698301db5fdba3c5fecfdd97ca952c1f0df3500740a567ecdb561555c8a34d0affcc99ae7a38b42d144373ae2f68b48064373b5baef7d25782fd07dc4b35f", 16);
  bigint q ("d32d977062a62dccfc4a37a21b03fca098973b72860002a3c05084060fbaa81b5c0fc636902a2959fb5ffd3d8a4969fbe9e15037c35477c9789da0b74ef32e3f", 16);
  bigint n ("a50e41c593b3b866bc4c72d0476611baab9bd54a22c62e11f536f87861ce592e7a101aea8652d3b949e66271b4497f91a861404eb5f3cba23f22b9b46fadda6cd327e3773eb23795e73ee06c16e5df18cf12e812fcd1bdbf3a4d7cca4fecd95fcbf248ac0534a3ebc67ebb06f9ca77d3ce1a5c4920da6d211b5f242e80d03661", 16);

  rsa_pub rsapub (n);
  str m ("a random string");
  bigint c = rsapub.encrypt (m);

  rsa_priv rsapriv (p, q);
  m = rsapriv.decrypt (c, m.len ());
  warn << "m " << m << "\n";

  rsa_priv x (rsa_keygen (1024));
  bigint pt (random_bigint (1019));
  bigint ct, pt2;
    
  BENCH (100000, ct = x.encrypt (pt));
  BENCH (1000, pt = x.decrypt (ct));

#if 0
  warn << pt.getstr (10) << "\n";
  ct = x.encrypt (pt);
  warn << ct.getstr (10) << "\n";;
  pt2 = x.decrypt (ct);
  warn << pt2.getstr (10) << "\n";
#endif

  rabin_priv xx (rabin_keygen (1280, 2));
  str pt3 ("plaintext message");

  BENCH (100000, ct = xx.encrypt (pt3));
  BENCH (1000, pt3 = xx.decrypt (ct, sizeof (pt3)));

#if 0
  BENCH (100, ct = x.sign (pt3));
  BENCH (1000, x.verify (pt3, ct));
  BENCH (1000, ct = x.encrypt (pt3));
#endif

  return 0;
}
