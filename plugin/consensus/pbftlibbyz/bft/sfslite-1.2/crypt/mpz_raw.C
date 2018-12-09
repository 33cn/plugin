/* $Id: mpz_raw.C 1117 2005-11-01 16:20:39Z max $ */

/*
 *
 * Copyright (C) 1998 David Mazieres (dm@uun.org)
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

#include "sysconf.h"
#include "bigint.h"

#define mpz_get_rawmag mpz_get_rawmag_be
#define mpz_set_rawmag mpz_set_rawmag_be

#undef min
#define min(a, b) (((a) < (b)) ? (a) : (b))

static inline void
assert_limb_size ()
{
  switch (0) case 0: case GMP_LIMB_SIZE == sizeof (mp_limb_t):;
#if GMP_LIMB_SIZE != 2 && GMP_LIMB_SIZE != 4 && GMP_LIMB_SIZE != 8
# error Cannot handle size of GMP limbs
#endif /* GMP_LIMB_SIZE not 2, 4 or 8 */
}

#define COPYLIMB_BYTE(dst, src, SW, n) \
  ((char *) (dst))[n] = ((char *) (src))[SW (n)]

#if GMP_LIMB_SIZE == 2
# define COPYLIMB(dst, src, SW)		\
do {					\
  COPYLIMB_BYTE (dst, src, SW, 0);	\
  COPYLIMB_BYTE (dst, src, SW, 1);	\
} while (0)
#elif GMP_LIMB_SIZE == 4
# define COPYLIMB(dst, src, SW)		\
do {					\
  COPYLIMB_BYTE (dst, src, SW, 0);	\
  COPYLIMB_BYTE (dst, src, SW, 1);	\
  COPYLIMB_BYTE (dst, src, SW, 2);	\
  COPYLIMB_BYTE (dst, src, SW, 3);	\
} while (0)
#elif GMP_LIMB_SIZE == 8
# define COPYLIMB(dst, src, SW)		\
do {					\
  COPYLIMB_BYTE (dst, src, SW, 0);	\
  COPYLIMB_BYTE (dst, src, SW, 1);	\
  COPYLIMB_BYTE (dst, src, SW, 2);	\
  COPYLIMB_BYTE (dst, src, SW, 3);	\
  COPYLIMB_BYTE (dst, src, SW, 4);	\
  COPYLIMB_BYTE (dst, src, SW, 5);	\
  COPYLIMB_BYTE (dst, src, SW, 6);	\
  COPYLIMB_BYTE (dst, src, SW, 7);	\
} while (0)
#else /* GMP_LIMB_SIZE != 2, 4 or 8 */
# error Cannot handle size of GMP limbs
#endif /* GMP_LIMB_SIZE != 2, 4 or 8 */

#define LSWAP(n) ((n & -GMP_LIMB_SIZE) + GMP_LIMB_SIZE-1 - n % GMP_LIMB_SIZE)

#ifdef WORDS_BIGENDIAN
# define LE_POS(n) LSWAP(n)
# define BE_POS(n) n
#else /* !WORDS_BIGENDIAN */
# define LE_POS(n) n
# define BE_POS(n) LSWAP(n)
#endif /* !WORDS_BIGENDIAN */

size_t
mpz_rawsize (const MP_INT *mp)
{
  size_t nbits = mpz_sizeinbase2 (mp);
  if (nbits)
    return (nbits>>3) + 1;	/* Not (nbits+7)/8, because we need sign bit */
  else
    return 0;
}

void
mpz_get_rawmag_le (char *buf, size_t size, const MP_INT *mp)
{
  char *bp = buf;
  const mp_limb_t *sp = mp->_mp_d;
  const mp_limb_t *ep = sp + min (size / GMP_LIMB_SIZE,
				  (size_t) ABS (mp->_mp_size));

  while (sp < ep) {
    COPYLIMB (bp, sp, LE_POS);
    bp += GMP_LIMB_SIZE;
    sp++;
  }
  size_t n = size - (bp - buf);
  if (n < GMP_LIMB_SIZE && sp < mp->_mp_d + ABS (mp->_mp_size)) {
    mp_limb_t v = *sp;
    for (char *e = bp + n; bp < e; v >>= 8)
      *bp++ = v;
  }
  else
    bzero (bp, n);
}

void
mpz_get_rawmag_be (char *buf, size_t size, const MP_INT *mp)
{
  char *bp = buf + size;
  const mp_limb_t *sp = mp->_mp_d;
  const mp_limb_t *ep = sp + min (size / GMP_LIMB_SIZE,
				  (size_t) ABS (mp->_mp_size));

  while (sp < ep) {
    bp -= GMP_LIMB_SIZE;
    COPYLIMB (bp, sp, BE_POS);
    sp++;
  }
  size_t n = bp - buf;
  if (n < GMP_LIMB_SIZE && sp < mp->_mp_d + ABS (mp->_mp_size)) {
    mp_limb_t v = *sp;
    for (; bp > buf; v >>= 8)
      *--bp = v;
  }
  else
    bzero (buf, n);
}

void
mpz_get_raw (char *buf, size_t size, const MP_INT *mp)
{
  if (mp->_mp_size < 0) {
    mpz_t neg;
    mpz_init (neg);
    mpz_umod_2exp (neg, mp, size * 8);
    mpz_get_rawmag (buf, size, neg);
    mpz_clear (neg);
  }
  else
    mpz_get_rawmag (buf, size, mp);
}

void
mpz_set_rawmag_le (MP_INT *mp, const char *buf, size_t size)
{
  const char *bp = buf;
  size_t nlimbs = (size + sizeof (mp_limb_t)) / sizeof (mp_limb_t);
  mp_limb_t *sp;
  mp_limb_t *ep;

  mp->_mp_size = nlimbs;
  if (nlimbs > (u_long) mp->_mp_alloc)
    _mpz_realloc (mp, nlimbs);
  sp = mp->_mp_d;
  ep = sp + size / sizeof (mp_limb_t);

  while (sp < ep) {
    COPYLIMB (sp, bp, LE_POS);
    bp += GMP_LIMB_SIZE;
    sp++;
  }

  const char *ebp = buf + size;
  if (bp < ebp) {
    mp_limb_t v = (u_char) *--ebp;
    while (bp < ebp)
      v = v << 8 | (u_char) *--ebp;
    *sp++ = v;
  }

  while (sp > mp->_mp_d && !sp[-1])
    sp--;
  mp->_mp_size = sp - mp->_mp_d;
}

void
mpz_set_rawmag_be (MP_INT *mp, const char *buf, size_t size)
{
  const char *bp = buf + size;
  size_t nlimbs = (size + sizeof (mp_limb_t)) / sizeof (mp_limb_t);
  mp_limb_t *sp;
  mp_limb_t *ep;

  mp->_mp_size = nlimbs;
  if (nlimbs > (u_long) mp->_mp_alloc)
    _mpz_realloc (mp, nlimbs);
  sp = mp->_mp_d;
  ep = sp + size / sizeof (mp_limb_t);

  while (sp < ep) {
    bp -= GMP_LIMB_SIZE;
    COPYLIMB (sp, bp, BE_POS);
    sp++;
  }

  if (bp > buf) {
    mp_limb_t v = (u_char) *buf++;
    while (bp > buf) {
      v <<= 8;
      v |= (u_char) *buf++;
    }
    *sp++ = v;
  }

  while (sp > mp->_mp_d && !sp[-1])
    sp--;
  mp->_mp_size = sp - mp->_mp_d;
}

void
mpz_set_raw (MP_INT *mp, const char *buf, size_t size)
{
  mpz_set_rawmag (mp, buf, size);
  if (*buf & 0x80) {
    mp->_mp_size = - mp->_mp_size;
    mpz_umod_2exp (mp, mp, 8 * size);
    mp->_mp_size = - mp->_mp_size;
  }
}
