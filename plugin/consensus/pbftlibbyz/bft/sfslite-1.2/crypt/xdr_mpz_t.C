/* $Id: xdr_mpz_t.C 3758 2008-11-13 00:36:00Z max $ */

#include "arpc.h"
#include "bigint.h"

/*
 * An external data representation format for MP_INTs.
 *
 * This is basically the same as opaque XDR data:  Size followed by
 * data followed by padding.  However, the encoded size is always
 * chosen to be a multiple of 4, so so the padding is never used.
 *
 * The external representation has the following three parts:
 *
 *     size: 4-byte big-endian word indicating the number of bytes
 *           used to encode the number.  The ENCODE operation rounds
 *           this up to the next multiple of 4.
 *
 *   number: a size byte, 2's complement represantation of the MP_INT
 *           (The most significant bit is the sign bit.  The number
 *           -1 is represented by size bytes of value 0xff - where
 *           size would most likely be 1.)
 *
 *  padding: (4 - size) % 4 bytes of padding to make the whole thing
 *           a multiple of four bytes.  (Since size is 0 bytes, the
 *           encoding will always have a zero-length padding.)
 */
bool_t
xdr_mpz_t (register XDR *xdrs, MP_INT *z)
{
  u_int32_t size;
  char *cp;

  switch (xdrs->x_op) {
  case XDR_ENCODE:
    size = (mpz_rawsize (z) + 3) & ~3;
    if (!xdr_putint (xdrs, size))
      return FALSE;
    if (!(cp = (char *) XDR_INLINE (xdrs, size)))
      return FALSE;
    mpz_get_raw (cp, size, z);
    break;

  case XDR_DECODE:
    if (!z->_mp_d)
      mpz_init (z);
    if (!xdr_getint (xdrs, size) || (int32_t) size < 0)
      return FALSE;
    if (!(cp = (char *) XDR_INLINE (xdrs, (size + 3) & ~3)))
      return FALSE;
    mpz_set_raw (z, cp, size);
    break;

  case XDR_FREE:
    if (z->_mp_d)
      mpz_clear (z);
    z->_mp_d = NULL;
  }

  return TRUE;
}
