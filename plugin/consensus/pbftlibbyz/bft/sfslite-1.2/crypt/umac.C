/* $Id: umac.C 3758 2008-11-13 00:36:00Z max $ */

/*
 *
 * Copyright (C) 2003 David Mazieres (dm@uun.org)
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

#include "umac.h"
#include "serial.h"

const bigint umac_prime<128>::prime ("0xffffffffffffffffffffffffffffff61");
const bigint umac_prime<128>::marker (umac_prime<128>::prime - 1);
const bigint umac_prime<128>::maxword ("0xffffffff000000000000000000000000");

const bigint umac_poly<64>::prime ((u_int64_t) umac_prime<64>::prime);
const bigint umac_poly<64>::marker ((u_int64_t) umac_prime<64>::marker);
const bigint umac_poly<64>::maxword (umac_prime<64>::maxword);

const bigint umac::mask128 ("0x01ffffff01ffffff01ffffff01ffffff");

umac_u32_le::word_t
umac_u32_le::l3hash (int polyno, u_int64_t val)
{
  u_int16_t m[8] = { 0, 0, 0, 0 };
  for (int i = 0; i < 4; i++)
    m[7-i] = val >> (16 * i);
  return ((m[0]*k31[polyno][0] + m[1]*k31[polyno][1]
	   + m[2]*k31[polyno][2] + m[3]*k31[polyno][3]
	   + m[4]*k31[polyno][4] + m[5]*k31[polyno][5]
	   + m[6]*k31[polyno][6] + m[7]*k31[polyno][7])
	  % umac_prime<36>::prime) ^ k32[polyno];
}

umac_u32_le::word_t
umac_u32_le::l3hash (int polyno, bigint val)
{
  u_int16_t m[8];
  for (int i = 0; i < 8; i++) {
    m[7-i] = val.getui ();
    val >>= 16;
  }
  return ((m[0]*k31[polyno][0] + m[1]*k31[polyno][1]
	   + m[2]*k31[polyno][2] + m[3]*k31[polyno][3]
	   + m[4]*k31[polyno][4] + m[5]*k31[polyno][5]
	   + m[6]*k31[polyno][6] + m[7]*k31[polyno][7])
	  % umac_prime<36>::prime) ^ k32[polyno];
}

void
umac_u32_le::setkey2 (const aes_e &ek)
{
  {
    char buf[24 * output_words];
    umac::kdf (buf, sizeof (buf), ek, 1);
    const char *cp = buf;
    for (int i = 0; i < output_words; i++) {
      k21[i] = umac_word<8>::getwordbe (cp) & umac::mask64;
      cp += 8;
      k22[i] = umac_word<16>::getwordbe (cp) & umac::mask128;
      cp += 16;
    }
  }

  {
    char buf[8 * sizeof (dword_t) * output_words];
    umac::kdf (buf, sizeof (buf), ek, 2);
    const char *cp = buf;
    for (int i = 0; i < output_words; i++)
      for (int j = 0; j < 8; j++) {
	k31[i][j] = (umac_word<sizeof (dword_t)>::getwordbe (cp)
		     % umac_prime<36>::prime);
	cp += sizeof (dword_t);
      }
  }

  {
    char buf[sizeof (word_t) * output_words];
    umac::kdf (buf, sizeof (buf), ek, 3);
    const char *cp = buf;
    for (int i = 0; i < output_words; i++) {
      k32[i] = umac_word<sizeof (word_t)>::getwordbe (cp);
      cp += sizeof (word_t);
    }
  }
}

void
umac_u32_le::poly_reset ()
{
  for (int i = 0; i < output_words; i++) {
    y1[i].poly_reset ();
    y2[i].poly_reset ();
  }
}

void
umac_u32_le::poly_update (int polyno, umac_u32_le::dword_t val)
{
  // XXX - need to switch to k22 for big messages
  y1[polyno].poly_inner (k21[polyno], val);
}

void
umac_u32_le::poly_final (void *_dp)
{
  // XXX - need to switch to y2 for big messages
  char *dp = static_cast<char *> (_dp);
  for (int i = 0; i < output_words; i++) {
    umac_word<sizeof (word_t)>::putwordbe (dp, l3hash (i, y1[i].yp));
    dp += sizeof (word_t);
  }
}

void
umac::size_sanity ()
{
  switch (0) {
  case 0:
  case (output_bits == output_words * word_size * 8):;
  }
}

void
umac::kdf (void *out, u_int len, const aes_e &ek, u_char index)
{
  u_char buf[16] = { 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, index };
  u_char *s = buf;
  u_char *d = static_cast<u_char *> (out);

  while (len >= 16) {
    ek.encipher_bytes (d, s);
    s = d;
    d += 16;
    len -= 16;
  }
  if (len > 0) {
    ek.encipher_bytes (buf, s);
    memcpy (d, buf, len);
  }
}

void
umac::kdfw (umac::word_t *out, u_int len, const aes_e &ek, u_char index)
{
  assert (!(len & word_size_mask));

  enum { words_per_block = 16 / word_size };

  word_t buf[words_per_block];
  bzero (buf, sizeof (buf) - word_size);
  buf[words_per_block-1] = htonw (index);

  while (len >= 16) {
    ek.encipher_bytes (buf);
    for (int i = 0; i < words_per_block; i++)
      out[i] = ntohw (buf[i]);
    len -= 16;
    out += words_per_block;
  }
  if (len > 0) {
    ek.encipher_bytes (buf);
    len /= word_size;
    for (u_int i = 0; i < len; i++)
      out[i] = ntohw (buf[i]);
  }
}

umac::dword_t
umac::nh (const umac::word_t *k, const umac::word_t *m)
{
  dword_t y = l1_key_len * 8;
  const word_t *ek = k + (l1_key_len / word_size);
  while (k < ek) {
    y += nh_inner (k, m);
    k += l1_block_size / word_size;
    m += l1_block_size / word_size;
  }
  return y;
}

umac::dword_t
umac::nh (const umac::word_t *k, const umac::word_t *m, u_int len)
{
  dword_t y = len * 8;
  u_int extra = len & (l1_block_size - 1);
  const word_t *ek = k + (len - extra) / word_size;
  while (k < ek) {
    y += nh_inner (k, m);
    k += l1_block_size / word_size;
    m += l1_block_size / word_size;
  }
  if (extra) {
    word_t buf[l1_block_size/word_size];
    bzero (buf, sizeof (buf));
    // XXX - assumes copied into oversized/aligned buffer
    memcpy (buf, m, (extra + word_size_mask) & ~word_size_mask);
    y += nh_inner (k, buf);
  }
  return y;
}

void
umac::consume ()
{
  totlen += l1_key_len;
  for (int i = 0; i < output_words; i++) {
    dword_t yt = nh (k1 + i * l1_key_shift, wbuf);
    poly_update (i, yt);
  }
}

void
umac::final (void *_mac)
{
  char *mac = static_cast<char *> (_mac);
  if (!totlen) {
    for (int i = 0; i < output_words; i++) {
      dword_t yt = nh (k1 + i * l1_key_shift, wbuf, l1len);
      poly_set (i, yt);
    }
  }
  else if (l1len) {
    for (int i = 0; i < output_words; i++) {
      dword_t yt = nh (k1 + i * l1_key_shift, wbuf, l1len);
      poly_update (i, yt);
    }
  }
  poly_final (mac);
  // XXX - XOR in PAD
}

void
umac::setkey (const void *key, u_int keylen)
{
  aes_e ek;
  ek.setkey (key, keylen);

  char kp[16];
  kdf (kp, sizeof (kp), ek, 128);
  kpad.setkey (kp, sizeof (kp));

  kdfw (k1, sizeof (k1), ek, 0);
  setkey2 (ek);

  reset ();
}

void
umac::reset ()
{
  poly_reset ();
  l1len = 0;
  totlen = 0;
}

void
umac::update (const void *_dp, size_t len)
{
  const u_int8_t *dp = static_cast <const u_int8_t *> (_dp);

  if (!len)
    return;

  if (l1len & word_size_mask) {
    u_int8_t c[word_size];

    u_int i = 0;
    while (i < (l1len & word_size_mask))
      c[i++] = 0;
    while (len > 0 && i < word_size) {
      c[i++] = *dp++;
      l1len++;
      len--;
    }
    while (i < word_size)
      c[i++] = 0;
    wbuf[l1len++ / word_size] |= getword (c);
  }

  u_int l1pos = l1len / word_size;

  while (len > l1_key_len - l1pos * word_size) {
    len -= l1_key_len - l1pos * word_size;
    while (l1pos < l1_key_words) {
      wbuf[l1pos++] = getword (dp);
      dp += word_size;
    }
    consume ();
    l1pos = 0;
  }

  while (len >= word_size) {
    wbuf[l1pos++] = getword (dp);
    dp += word_size;
    len -= word_size;
  }

  l1len = l1pos * word_size;

  if (len > 0) {
    u_int8_t c[word_size];
    for (u_int i = 0; i < word_size; i++)
      c[i] = i < len ? dp[i] : 0;
    wbuf[l1len / word_size] = getword (c);
    l1len += len;
  }
}
