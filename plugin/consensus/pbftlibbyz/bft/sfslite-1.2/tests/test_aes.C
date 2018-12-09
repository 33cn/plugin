/* $Id: test_aes.C 2 2003-09-24 14:35:33Z max $ */

/*
 *
 * Copyright (C) 2000 David Mazieres (dm@uun.org)
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

#define USE_PCTR 0

#include "crypt.h"
#include "aes.h"
#include "bench.h"

struct testvec {
  int klen;
  u_char key[32];
  u_char ptext[16];
  u_char ctext[16];
};

const testvec vectors[] = {
  { 16,
    { 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
      0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f },
    { 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77,
      0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff },
    { 0x69, 0xc4, 0xe0, 0xd8, 0x6a, 0x7b, 0x04, 0x30,
      0xd8, 0xcd, 0xb7, 0x80, 0x70, 0xb4, 0xc5, 0x5a } },

  { 16,
    { 0x00, 0x01, 0x02, 0x03, 0x05, 0x06, 0x07, 0x08,
      0x0A, 0x0B, 0x0C, 0x0D, 0x0F, 0x10, 0x11, 0x12 },
    { 0x50, 0x68, 0x12, 0xA4, 0x5F, 0x08, 0xC8, 0x89,
      0xB9, 0x7F, 0x59, 0x80, 0x03, 0x8B, 0x83, 0x59 },
    { 0xD8, 0xF5, 0x32, 0x53, 0x82, 0x89, 0xEF, 0x7D,
      0x06, 0xB5, 0x06, 0xA4, 0xFD, 0x5B, 0xE9, 0xC9 } },

  { 16,
    { 0xDC, 0xDD, 0xDE, 0xDF, 0xE1, 0xE2, 0xE3, 0xE4,
      0xE6, 0xE7, 0xE8, 0xE9, 0xEB, 0xEC, 0xED, 0xEE },
    { 0x23, 0x60, 0x5A, 0x82, 0x43, 0xD0, 0x77, 0x64,
      0x54, 0x1B, 0xC5, 0xAD, 0x35, 0x5B, 0x31, 0x29 },
    { 0x6D, 0x96, 0xFE, 0xF7, 0xD6, 0x65, 0x90, 0xA7,
      0x7A, 0x77, 0xBB, 0x20, 0x56, 0x66, 0x7F, 0x7F } },

  { 24,
    { 0x78, 0x79, 0x7A, 0x7B, 0x7D, 0x7E, 0x7F, 0x80,
      0x82, 0x83, 0x84, 0x85, 0x87, 0x88, 0x89, 0x8A,
      0x8C, 0x8D, 0x8E, 0x8F, 0x91, 0x92, 0x93, 0x94 },
    { 0x02, 0xAE, 0xA8, 0x6E, 0x57, 0x2E, 0xEA, 0xB6,
      0x6B, 0x2C, 0x3A, 0xF5, 0xE9, 0xA4, 0x6F, 0xD6 },
    { 0x8F, 0x67, 0x86, 0xBD, 0x00, 0x75, 0x28, 0xBA,
      0x26, 0x60, 0x3C, 0x16, 0x01, 0xCD, 0xD0, 0xD8 } },

  { 32,
    { 0x60, 0x61, 0x62, 0x63, 0x65, 0x66, 0x67, 0x68,
      0x6A, 0x6B, 0x6C, 0x6D, 0x6F, 0x70, 0x71, 0x72,
      0x74, 0x75, 0x76, 0x77, 0x79, 0x7A, 0x7B, 0x7C,
      0x7E, 0x7F, 0x80, 0x81, 0x83, 0x84, 0x85, 0x86 },
    { 0xD2, 0x1E, 0x43, 0x9F, 0xF7, 0x49, 0xAC, 0x8F,
      0x18, 0xD6, 0xD4, 0xB1, 0x05, 0xE0, 0x38, 0x95 },
    { 0x9A, 0x6D, 0xB0, 0xC0, 0x86, 0x2E, 0x50, 0x6A,
      0x9E, 0x39, 0x72, 0x25, 0x88, 0x40, 0x41, 0xD7 } },
};
const int ntestvec = sizeof (vectors) / sizeof (vectors[0]);

struct b16 {
  enum { nc = 16, nl = nc / sizeof (long) };
  union {
    char c[nc];
    long l[nl];
  };
};
inline void
b16xor (b16 *d, const b16 &s)
{
  for (int i = 0; i < b16::nl; i++)
    d->l[i] ^= s.l[i];
}
inline void
b16xor (b16 *d, const b16 &s1, const b16 &s2)
{
  for (int i = 0; i < b16::nl; i++)
    d->l[i] = s1.l[i] ^ s2.l[i];
}


static void
cbcencrypt (aes *cp, void *_d, const void *_s, int len)
{
  assert (!(len & 15));
  len >>= 4;
  const b16 *s = static_cast<const b16 *> (_s);
  b16 *d = static_cast<b16 *> (_d);

  if (len-- > 0) {
    cp->encipher_bytes (d->c, (s++)->c);
    while (len-- > 0) {
      b16 tmp;
      b16xor (&tmp, *d++, *s++);
      cp->encipher_bytes (d->c, tmp.c);
    }
  }
}

static void
cbcdecrypt (aes *cp, void *_d, const void *_s, int len)
{
  assert (!(len & 15));
  len >>= 4;
  const b16 *s = static_cast<const b16 *> (_s) + len;
  b16 *d = static_cast<b16 *> (_d) + len;

  if (len-- > 0) {
    --s;
    while (len-- > 0) {
      cp->decipher_bytes ((--d)->c, s->c);
      b16xor (d, *--s);
    }
    cp->decipher_bytes ((--d)->c, s->c);
  }
}

int
main (int argc, char **argv)
{
  aes ctx;
  u_char buf[16];
  bool opt_verbose = false;

  if (argc > 1 && !strcmp (argv[1], "-v"))
    opt_verbose = true;

  for (int i = 0; i < ntestvec; i++) {
    ctx.setkey (vectors[i].key, vectors[i].klen);
    memcpy (buf, vectors[i].ptext, 16);
    ctx.encipher_bytes (buf);
    if (memcmp (buf, vectors[i].ctext, sizeof (buf)))
      panic ("test %d encipher failed\n", i);
    ctx.decipher_bytes (buf);
    if (memcmp (buf, vectors[i].ptext, sizeof (buf)))
      panic ("test %d decipher failed\n", i);
  }

  char key[] = "This is a test key of 32 bytes.";
  char pbuf[0x100000];
  char cbuf[sizeof (pbuf)];
  char tbuf[sizeof (pbuf)];
  random_update ();
  rnd.getbytes (pbuf, sizeof (pbuf));
  ctx.setkey (key, sizeof (key));

  if (opt_verbose) {
    BENCH (655360, ctx.encipher_bytes (cbuf, pbuf));
    BENCH (655360, ctx.decipher_bytes (tbuf, cbuf));
  }

  if (opt_verbose) {
    BENCH (100, cbcencrypt (&ctx, cbuf, pbuf, sizeof (pbuf)));
    BENCH (100, cbcdecrypt (&ctx, tbuf, cbuf, sizeof (pbuf)));
  }
  else {
    cbcencrypt (&ctx, cbuf, pbuf, sizeof (pbuf));
    cbcdecrypt (&ctx, tbuf, cbuf, sizeof (pbuf));
  }
  if (memcmp (pbuf, tbuf, sizeof (pbuf)))
    panic ("cbc encryption/decryption failed\n");

  return 0;
}
