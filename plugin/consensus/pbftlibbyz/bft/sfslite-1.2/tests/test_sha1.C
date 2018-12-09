/* $Id: test_sha1.C 3256 2008-05-14 04:02:23Z max $ */

#define USE_PCTR 0

#include "sha1.h"
#include "bench.h"
#include "err.h"

/* SHA1 test vectors, from Handbook of Applied Cryptography Menezes,
   Oorschot and Vanstone */

#define NTEST 6U
struct tv_t {
  const char *in;	        	/* input */
  u_int8_t res[sha1ctx::hashsize]; /* expected result */
} tv[NTEST] = {
  {
    "", {
      0xda, 0x39, 0xa3, 0xee, 0x5e,
	0x6b, 0x4b, 0x0d, 0x32, 0x55,
	0xbf, 0xef, 0x95, 0x60, 0x18,
	0x90, 0xaf, 0xd8, 0x07, 0x09
    }
  },
  {
    "a", {
      0x86, 0xf7, 0xe4, 0x37, 0xfa,
	0xa5, 0xa7, 0xfc, 0xe1, 0x5d,
	0x1d, 0xdc, 0xb9, 0xea, 0xea,
	0xea, 0x37, 0x76, 0x67, 0xb8
    }
  },
  {
    "abc", {
      0xA9, 0x99, 0x3e, 0x36, 0x47,
	0x06, 0x81, 0x6a, 0xba, 0x3e,
	0x25, 0x71, 0x78, 0x50, 0xc2,
	0x6c, 0x9c, 0xd0, 0xd8, 0x9d
    }
  },
  {
    "abcdbcdecdefdefgefghfghighijhijkijkljklmklmnlmnomnopnopq",
    {
      0x84, 0x98, 0x3E, 0x44,
	0x1C, 0x3B, 0xD2, 0x6E,
	0xBA, 0xAE, 0x4A, 0xA1,
	0xF9, 0x51, 0x29, 0xE5,
	0xE5, 0x46, 0x70, 0xF1
    }
  },
  {
    "abcdefghijklmnopqrstuvwxyz", {
      0x32, 0xd1, 0x0c, 0x7b, 0x8c,
	0xf9, 0x65, 0x70, 0xca, 0x04,
	0xce, 0x37, 0xf2, 0xa1, 0x9d,
	0x84, 0x24, 0x0d, 0x3a, 0x89
    }
  },
  /* a million a's, 10**5 of this string */
  {
    "aaaaaaaaaa", {
      0x34, 0xAA, 0x97, 0x3C,
	0xD4, 0xC4, 0xDA, 0xA4,
	0xF6, 0x1E, 0xEB, 0x2B,
	0xDB, 0xAD, 0x27, 0x31,
	0x65, 0x34, 0x01, 0x6F
    }
  },
};

void 
printbs (u_char * bs, int len)
{
  u_char *end = bs + len;
  while (bs < end)
    printf ("0x%02x, ", *bs++);
  printf ("\n");
}
int 
main (int argc, char **argv)
{
  bool opt_v = false;
  sha1ctx c;
  u_int8_t h[sha1ctx::hashsize];
  char buf[100];
  u_int i, j;

  if (argc > 1 && !strcmp (argv[1], "-v"))
    opt_v = true;

  for (i = 0; i < NTEST - 1; i++) {
    c.reset ();
    strncpy (buf, tv[i].in, 100);
    c.update (buf, strlen (buf));
    c.final (h);
    if (strcmp (tv[i].in, buf)) {
      printf ("%s--is now--%s\n", tv[i].in, buf);
      abort ();
    }
    if (memcmp (h, tv[i].res, sha1ctx::hashsize)) {
      printf ("h(%s) = ", buf);
      printbs (h, sha1ctx::hashsize);
      abort ();
    }
  }
  strncpy (buf, tv[i].in, 100);
  c.reset ();
  for (j = 0; j < 100000; j++) {
    c.update (buf, strlen (buf));
  }
  c.final (h);
  if (strcmp (tv[i].in, buf)) {
    printf ("%s--is now--%s\n", tv[i].in, buf);
    abort ();
  }
  if (memcmp (h, tv[i].res, sha1ctx::hashsize)) {
    printf ("h(%s) = ", buf);
    printbs (h, sha1ctx::hashsize);
    abort ();
  }
  if (opt_v) {
    const u_int8_t hok[20] = {
      0x10, 0x9B, 0x42, 0x6B, 0x74, 0xC3, 0xDC, 0x1B, 0xD0, 0xE1,
      0x5D, 0x35, 0x24, 0xC5, 0xB8, 0x37, 0x55, 0x76, 0x47, 0xF2,
    };
    char bigbuf[500000];

    for (i = 0; i < sizeof (bigbuf); i++)
      bigbuf[i] = 'a';
    c.reset ();

    u_int64_t htime = get_time ();

    for (i = 0; i < 10000; i++)
      c.update (bigbuf, sizeof (bigbuf));
    c.final (h);

    htime = get_time () - htime;
    warn ("Hashed 5,000,000,000 bytes in %" U64F "u " TIME_LABEL "\n", htime);

    if (memcmp (h, hok, sizeof (hok))) {
      fprintf (stderr, "5 Gigabyte test failed.  5 Billon 'a's hashed to:\n ");
      for (i = 0; i < sizeof (h); i++)
	fprintf (stderr, "%02x", h[i]);
      fprintf (stderr, "\n");
      abort ();
    }
  }
  return 0;
}
