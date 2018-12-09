#include "crypt.h"
#include "bench.h"
#include "hashcash.h"

char payment[sha1::blocksize];
char inithash[sha1::hashsize];
char target[sha1::hashsize];

int
main (int argc, char **argv)
{
  u_long j;
  bool opt_verbose = false;

  if (argc > 1 && !strcmp (argv[1], "-v"))
    opt_verbose = true;

  for (unsigned int i = 0; i < 20; i++) {
    if (opt_verbose) {
      TIME(j = hashcash_pay (payment, inithash, target, i););
      warnx << "bitcost " << i << " " << j << " iterations\n";
    }
    else
      j = hashcash_pay (payment, inithash, target, i);
    if (!hashcash_check (payment, inithash, target, i))
      panic ("payment doesn't match target\n");
  }
}
