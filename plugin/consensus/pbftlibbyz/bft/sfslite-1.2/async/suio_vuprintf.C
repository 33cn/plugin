/* $Id: suio_vuprintf.C 3146 2007-12-20 17:44:02Z max $ */

/*-
 * Copyright (c) 1990 The Regents of the University of California.
 * All rights reserved.
 *
 * This code is derived from software contributed to Berkeley by
 * Chris Torek.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions
 * are met:
 * 1. Redistributions of source code must retain the above copyright
 *    notice, this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright
 *    notice, this list of conditions and the following disclaimer in the
 *    documentation and/or other materials provided with the distribution.
 * 3. All advertising materials mentioning features or use of this software
 *    must display the following acknowledgement:
 *	This product includes software developed by the University of
 *	California, Berkeley and its contributors.
 * 4. Neither the name of the University nor the names of its contributors
 *    may be used to endorse or promote products derived from this software
 *    without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE REGENTS AND CONTRIBUTORS ``AS IS'' AND
 * ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED.  IN NO EVENT SHALL THE REGENTS OR CONTRIBUTORS BE LIABLE
 * FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
 * DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS
 * OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
 * HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
 * LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY
 * OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF
 * SUCH DAMAGE.
 */

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

#undef FLOATING_POINT

#include "suio++.h"

#ifdef FLOATING_POINT
#include <locale.h>
#include <math.h>

/* 11-bit exponent (VAX G floating point) is 308 decimal digits */
#define MAXEXP          308
/* 128 bit fraction takes up 39 decimal digits; max reasonable precision */
#define MAXFRACT        39

#define	BUF		(MAXEXP+MAXFRACT+1)	/* + decimal point */
#define	DEFPREC		6

static char *cvt (double, int, int, char *, int *, int, int *);
static int exponent (char *, int, int);

#else /* ! FLOATING_POINT */

#define	BUF		40

#endif /* ! FLOATING_POINT */


/*
 * Macros for converting digits to letters and vice versa
 */
#define	to_digit(c)	((c) - '0')
#define is_digit(c)	((unsigned)to_digit(c) <= 9)
#define	to_char(n)	((n) + '0')

/*
 * Flags used during conversion.
 */
#define	ALT		0x001	/* alternate form */
#define	HEXPREFIX	0x002	/* add 0x or 0X prefix */
#define	LADJUST		0x004	/* left adjustment */
#define	LONGDBL		0x008	/* long double; unimplemented */
#define	LONGINT		0x010	/* long integer */
#define	QUADINT		0x020	/* quad integer */
#define	SHORTINT	0x040	/* short integer */
#define	ZEROPAD		0x080	/* zero (as opposed to blank) pad */
#define FPT		0x100	/* Floating point number */
#define NOCOPY          0x200   /* String contents doesn't have to be copied */
#define SIZET           0x400   /* Variable of type size_t */

static size_t
my_strnlen (const char *s, size_t len)
{
  const char *cp;
  size_t r;

  for (cp = s, r = 0; r < len && *cp; cp++, r++) ;
  return r;
}

void
#ifndef DSPRINTF_DEBUG
suio_vuprintf (struct suio *uio, const char *_fmt, va_list ap)
#else /* DSPRINTF_DEBUG */
__suio_vuprintf (const char *line, struct suio *uio,
		 const char *_fmt, va_list ap)
#endif /* DSPRINTF_DEBUG */
{
  char *fmt = (char *) _fmt;
  register int ch;		/* character from fmt */
  register int n, m;		/* handy integers (short term usage) */
  register char *cp;		/* handy char pointer (short term usage) */
  register int flags;		/* flags as above */
  int width;			/* width from format (%8d), or 0 */
  int prec;			/* precision from format (%.3d), or -1 */
  char sign;			/* sign prefix (' ', '+', '-', or \0) */
#ifdef FLOATING_POINT
  char *decimal_point = ".";
  char softsign;		/* temporary negative sign for floats */
  double _double = 0.0;		/* double precision arguments %[eEfgG] */
  int expt;			/* integer value of exponent */
  int expsize = 0;			/* character count for expstr */
  int ndig;			/* actual number of digits returned by cvt */
  char expstr[7];		/* buffer for exponent string */
#endif

#define	quad_t	  int64_t
#define	u_quad_t  u_int64_t
  u_quad_t _uquad;		/* integer arguments %[diouxX] */
  enum {
    OCT, DEC, HEX
  } base;			/* base for [diouxX] conversion */
  int dprec;			/* a copy of prec if [diouxX], 0 otherwise */
  int realsz;			/* field size expanded by dprec */
  int size;			/* size of converted field or string */
  const char *xdigs = "";		/* digits for [xX] conversion */

  char buf[BUF];                /* space for %c, %[diouxX], %[eEfgG] */
  char ox[2];                   /* space for 0x hex-prefix */

  /*
   * To extend shorts properly, we need both signed and unsigned
   * argument extraction methods.
   */
#define	SARG()						\
  (flags&QUADINT ? va_arg(ap, quad_t) :			\
   flags&LONGINT ? va_arg(ap, long) :			\
   flags&SIZET ? va_arg(ap, ssize_t) :                  \
   flags&SHORTINT ? (long)(short)va_arg(ap, int) :	\
   (long)va_arg(ap, int))
#define	UARG()						\
  (flags&QUADINT ? va_arg(ap, u_quad_t) :		\
   flags&LONGINT ? va_arg(ap, u_long) :			\
   flags&SIZET ? va_arg(ap, size_t) :                   \
   flags&SHORTINT ? (u_long)(u_short)va_arg(ap, int) :	\
   (u_long)va_arg(ap, u_int))

  /*
   * Scan the format for conversions (`%' character).
   */
  for (;;) {
    cp = fmt;
    while (*fmt > 0) {
      if (*fmt == '%')
	break;
      fmt++;
    }
    if ((m = fmt - cp) != 0)
      suio_print (uio, cp, m);
    if (!*fmt)
      goto done;
    fmt++;			/* skip over '%' */

    flags = 0;
    dprec = 0;
    width = 0;
    prec = -1;
    sign = '\0';

  rflag:
    ch = *fmt++;
  reswitch:
    switch (ch) {
    case ' ':
      /*
       * ``If the space and + flags both appear, the space
       * flag will be ignored.''
       *      -- ANSI X3J11
       */
      if (!sign)
	sign = ' ';
      goto rflag;
    case '#':
      flags |= ALT;
      goto rflag;
    case '*':
      /*
       * ``A negative field width argument is taken as a
       * - flag followed by a positive field width.''
       *      -- ANSI X3J11
       * They don't exclude field widths read from args.
       */
      if ((width = va_arg (ap, int)) >= 0)
	  goto rflag;
      width = -width;
      /* FALLTHROUGH */
    case '-':
      flags |= LADJUST;
      goto rflag;
    case '+':
      sign = '+';
      goto rflag;
    case '.':
      if ((ch = *fmt++) == '*') {
	n = va_arg (ap, int);
	prec = n < 0 ? -1 : n;
	goto rflag;
      }
      n = 0;
      while (is_digit (ch)) {
	n = 10 * n + to_digit (ch);
	ch = *fmt++;
      }
      prec = n < 0 ? -1 : n;
      goto reswitch;
    case '0':
      /*
       * ``Note that 0 is taken as a flag, not as the
       * beginning of a field width.''
       *      -- ANSI X3J11
       */
      flags |= ZEROPAD;
      goto rflag;
    case '1':
    case '2':
    case '3':
    case '4':
    case '5':
    case '6':
    case '7':
    case '8':
    case '9':
      n = 0;
      do {
	n = 10 * n + to_digit (ch);
	ch = *fmt++;
      } while (is_digit (ch));
      width = n;
      goto reswitch;
#ifdef FLOATING_POINT
    case 'L':
      flags |= LONGDBL;
      goto rflag;
#endif
    case 'h':
      flags |= SHORTINT;
      goto rflag;
    case 'l':
      if (*fmt == 'l') {
	fmt++;
	flags |= QUADINT;
      }
      else {
	flags |= LONGINT;
      }
      goto rflag;
    case 'z':
      flags |= SIZET;
      goto rflag;
    case 'q':
      flags |= QUADINT;
      goto rflag;
    case 'c':
      *(cp = buf) = va_arg (ap, int);
      size = 1;
      sign = '\0';
      break;
    case 'D':
      flags |= LONGINT;
      /*FALLTHROUGH */
    case 'd':
    case 'i':
      _uquad = SARG ();
      if ((quad_t) _uquad < 0) {
	_uquad = -_uquad;
	sign = '-';
      }
      base = DEC;
      goto number;
#ifdef FLOATING_POINT
    case 'e':
    case 'E':
    case 'f':
    case 'g':
    case 'G':
      if (prec == -1) {
	prec = DEFPREC;
      }
      else if ((ch == 'g' || ch == 'G') && prec == 0) {
	prec = 1;
      }

      if (flags & LONGDBL) {
	_double = (double) va_arg (ap, long double);
      }
      else {
	_double = va_arg (ap, double);
      }

      /* do this before tricky precision changes */
      if (isinf (_double)) {
	if (_double < 0)
	  sign = '-';
	cp = "Inf";
	size = 3;
	break;
      }
      if (isnan (_double)) {
	cp = "NaN";
	size = 3;
	break;
      }

      flags |= FPT;
      cp = cvt (_double, prec, flags, &softsign,
		&expt, ch, &ndig);
      if (ch == 'g' || ch == 'G') {
	if (expt <= -4 || expt > prec)
	  ch = (ch == 'g') ? 'e' : 'E';
	else
	  ch = 'g';
      }
      if (ch <= 'e') {		/* 'e' or 'E' fmt */
	--expt;
	expsize = exponent (expstr, expt, ch);
	size = expsize + ndig;
	if (ndig > 1 || flags & ALT)
	  ++size;
      }
      else if (ch == 'f') {	/* f fmt */
	if (expt > 0) {
	  size = expt;
	  if (prec || flags & ALT)
	    size += prec + 1;
	}
	else			/* "0.X" */
	  size = prec + 2;
      }
      else if (expt >= ndig) {	/* fixed g fmt */
	size = expt;
	if (flags & ALT)
	  ++size;
      }
      else
	size = ndig + (expt > 0 ?
		       1 : 2 - expt);

      if (softsign)
	sign = '-';
      break;
#endif /* FLOATING_POINT */
#if 0
      /* Disable "%n", as we don't need it--you can always print to a
       * suio and use suio::resid () to get the buffer size. */
    case 'n':
      if (flags & QUADINT)
	*va_arg (ap, quad_t *) = uio->resid ();
      else if (flags & LONGINT)
	*va_arg (ap, long *) = uio->resid ();
      else if (flags & SHORTINT)
	*va_arg (ap, short *) = uio->resid ();
      else
	*va_arg (ap, int *) = uio->resid ();
      continue;			/* no output */
#endif
    case 'O':
      flags |= LONGINT;
      /*FALLTHROUGH */
    case 'o':
      _uquad = UARG ();
      base = OCT;
      goto nosign;
    case 'p':
      /*
       * ``The argument shall be a pointer to void.  The
       * value of the pointer is converted to a sequence
       * of printable characters, in an implementation-
       * defined manner.''
       *      -- ANSI X3J11
       */
      /* NOSTRICT */
      _uquad = (u_long) va_arg (ap, void *);
      base = HEX;
      xdigs = "0123456789abcdef";
      flags |= HEXPREFIX;
      ch = 'x';
      goto nosign;
    case 'm':
      cp = strerror (errno);
      goto gotcp;
    case 's':
      cp = va_arg (ap, char *);
    gotcp:
      if (cp == NULL)
	panic ("suio_vuprintf:  NULL pointer\n");
	//cp = "(null)";
      if (prec >= 0) {
	/*
	 * can't use strlen; can only look for the
	 * NUL in the first `prec' characters, and
	 * strlen() will go further.
	 *
	 * MK 8/16/07: Unfortunately, can't use memchr either,
	 * since dmalloc complains that memchr would scan off the
	 * end of the allocated string.
	 */
	size = my_strnlen (cp, prec);

	/*
	 * MK 8/16/07: used to be:
	 *
	  char *p = (char *) memchr (cp, 0, prec);
	  if (p != NULL) {
	  size = p - cp;
	  if (size > prec)
	  size = prec;
	  }
	  else
	  size = prec;
	*/
      }
      else
	size = strlen (cp);
      sign = '\0';
      if (!(flags & ALT))
	flags |= NOCOPY;
      break;
    case 'U':
      flags |= LONGINT;
      /*FALLTHROUGH */
    case 'u':
      _uquad = UARG ();
      base = DEC;
      goto nosign;
    case 'X':
      xdigs = "0123456789ABCDEF";
      goto hex;
    case 'x':
      xdigs = "0123456789abcdef";
    hex:
      _uquad = UARG ();
      base = HEX;
      /* leading 0x/X only if non-zero */
      if (flags & ALT && _uquad != 0)
	flags |= HEXPREFIX;

      /* unsigned conversions */
    nosign:
      sign = '\0';
      /*
       * ``... diouXx conversions ... if a precision is
       * specified, the 0 flag will be ignored.''
       *      -- ANSI X3J11
       */
    number:
      if ((dprec = prec) >= 0)
	flags &= ~ZEROPAD;

      /*
       * ``The result of converting a zero value with an
       * explicit precision of zero is no characters.''
       *      -- ANSI X3J11
       */
      cp = buf + BUF;
      if (_uquad != 0 || prec != 0) {
	/*
	 * Unsigned mod is hard, and unsigned mod
	 * by a constant is easier than that by
	 * a variable; hence this switch.
	 */
	switch (base) {
	case OCT:
	  do {
	    *--cp = to_char (_uquad & 7);
	    _uquad >>= 3;
	  } while (_uquad);
	  /* handle octal leading 0 */
	  if (flags & ALT && *cp != '0')
	    *--cp = '0';
	  break;

	case DEC:
	  /* many numbers are 1 digit */
	  while (_uquad >= 10) {
	    *--cp = to_char (_uquad % 10);
	    _uquad /= 10;
	  }
	  *--cp = to_char (_uquad);
	  break;

	case HEX:
	  do {
	    *--cp = xdigs[_uquad & 15];
	    _uquad >>= 4;
	  } while (_uquad);
	  break;

	default:
	  // XXX leak memory to satisfy compiler.  but it's in an error case
	  cp = strdup ("bug in vfprintf: bad base");
	  size = strlen (cp);
	  goto skipsize;
	}
      }
      size = buf + BUF - cp;
    skipsize:
      break;
    default:			/* "%?" prints ?, unless ? is NUL */
      if (ch == '\0')
	goto done;
      /* pretend it was %c with argument ch */
      cp = buf;
      *cp = ch;
      size = 1;
      sign = '\0';
      break;
    }

    /*
     * All reasonable formats wind up here.  At this point, `cp'
     * points to a string which (if not flags&LADJUST) should be
     * padded out to `width' places.  If flags&ZEROPAD, it should
     * first be prefixed by any sign or other prefix; otherwise,
     * it should be blank padded before the prefix is emitted.
     * After any left-hand padding and prefixing, emit zeroes
     * required by a decimal [diouxX] precision, then print the
     * string proper, then emit zeroes required by any leftover
     * floating precision; finally, if LADJUST, pad with blanks.
     *
     * Compute actual size, so we know how much to pad.
     * size excludes decimal prec; realsz includes it.
     */
    realsz = dprec > size ? dprec : size;
    if (sign)
      realsz++;
    else if (flags & HEXPREFIX)
      realsz += 2;

    /* right-adjusting blank padding */
    if ((flags & (LADJUST | ZEROPAD)) == 0)
      suio_fill (uio, ' ', width - realsz);

    /* prefix */
    if (sign) {
      suio_copy (uio, &sign, 1);
    }
    else if (flags & HEXPREFIX) {
      ox[0] = '0';
      ox[1] = ch;
      suio_copy (uio, ox, 2);
    }

    /* right-adjusting zero padding */
    if ((flags & (LADJUST | ZEROPAD)) == ZEROPAD)
      suio_fill (uio, '0', width - realsz);

    /* leading zeroes from decimal precision */
    suio_fill (uio, '0', dprec - size);

    /* the string or number proper */
#ifdef FLOATING_POINT
    if ((flags & FPT) == 0) {
      if (flags & NOCOPY)
	__suio_printcheck (line, uio, cp, size);
      else
	suio_copy (uio, cp, size);
    }
    else {			/* glue together f_p fragments */
      if (ch >= 'f') {		/* 'f' or 'g' */
	if (_double == 0) {
	  /* kludge for __dtoa irregularity */
	  suio_print (uio, "0", 1);
	  if (expt < ndig || (flags & ALT) != 0) {
	    suio_print (uio, decimal_point, 1);
	    suio_fill (uio, '0', ndig - 1);
	  }
	}
	else if (expt <= 0) {
	  suio_print (uio, "0", 1);
	  suio_print (uio, decimal_point, 1);
	  suio_fill (uio, '0', -expt);
	  suio_copy (uio, cp, ndig);
	}
	else if (expt >= ndig) {
	  suio_copy (uio, cp, ndig);
	  suio_fill (uio, '0', expt - ndig);
	  if (flags & ALT)
	    suio_print (uio, ".", 1);
	}
	else {
	  suio_copy (uio, cp, expt);
	  cp += expt;
	  suio_print (uio, ".", 1);
	  suio_copy (uio, cp, ndig - expt);
	}
      }
      else {			/* 'e' or 'E' */
	if (ndig > 1 || flags & ALT) {
	  ox[0] = *cp++;
	  ox[1] = '.';
	  suio_copy (uio, ox, 2);
	  if (_double || (flags & ALT) == 0) {
	    suio_copy (uio, cp, ndig - 1);
	  }
	  else			/* 0.[0..] */
	    /* __dtoa irregularity */
	    suio_fill (uio, '0', ndig - 1);
	}
	else			/* XeYYY */
	  suio_copy (uio, cp, 1);
	suio_copy (uio, expstr, expsize);
      }
    }
#else
    if (flags & NOCOPY)
#ifndef DSPRINTF_DEBUG
      suio_print (uio, cp, size);
#else /* DSPRINTF_DEBUG */
      __suio_printcheck (line, uio, cp, size);
#endif /* DSPRINTF_DEBUG */
    else
      suio_copy (uio, cp, size);
#endif
    /* left-adjusting padding (always blank) */
    if (flags & LADJUST)
      suio_fill (uio, ' ', width - realsz);
  }
done:
  return;
}

#ifdef FLOATING_POINT

extern char *__dtoa (double, int, int, int *, int *, char **);

static char *
cvt (value, ndigits, flags, sign, decpt, ch, length)
     double value;
     int ndigits, flags, *decpt, ch, *length;
     char *sign;
{
  int mode, dsgn;
  char *digits, *bp, *rve;

  if (ch == 'f') {
    mode = 3;			/* ndigits after the decimal point */
  }
  else {
    /* To obtain ndigits after the decimal point for the 'e'
     * and 'E' formats, round to ndigits + 1 significant
     * figures.
     */
    if (ch == 'e' || ch == 'E') {
      ndigits++;
    }
    mode = 2;			/* ndigits significant digits */
  }

  if (value < 0) {
    value = -value;
    *sign = '-';
  }
  else
    *sign = '\000';
  digits = __dtoa (value, mode, ndigits, decpt, &dsgn, &rve);
  if ((ch != 'g' && ch != 'G') || flags & ALT) {	/* Print trailing zeros */
    bp = digits + ndigits;
    if (ch == 'f') {
      if (*digits == '0' && value)
	*decpt = -ndigits + 1;
      bp += *decpt;
    }
    if (value == 0)		/* kludge for __dtoa irregularity */
      rve = bp;
    while (rve < bp)
      *rve++ = '0';
  }
  *length = rve - digits;
  return (digits);
}

static int
exponent (p0, exp, fmtch)
     char *p0;
     int exp, fmtch;
{
  register char *p, *t;
  char expbuf[MAXEXP];

  p = p0;
  *p++ = fmtch;
  if (exp < 0) {
    exp = -exp;
    *p++ = '-';
  }
  else
    *p++ = '+';
  t = expbuf + MAXEXP;
  if (exp > 9) {
    do {
      *--t = to_char (exp % 10);
    } while ((exp /= 10) > 9);
    *--t = to_char (exp);
    for (; t < expbuf + MAXEXP; *p++ = *t++);
  }
  else {
    *p++ = '0';
    *p++ = to_char (exp);
  }
  return (p - p0);
}
#endif /* FLOATING_POINT */

void
#ifndef DSPRINTF_DEBUG
suio_uprintf (struct suio *uio, const char *fmt, ...)
#else /* DSPRINTF_DEBUG */
__suio_uprintf (const char *line, struct suio *uio, const char *fmt, ...)
#endif /* DSPRINTF_DEBUG */
{
  va_list ap;
  va_start (ap, fmt);
#ifndef DSPRINTF_DEBUG
  suio_vuprintf (uio, fmt, ap);
#else /* DSPRINTF_DEBUG */
  __suio_vuprintf (line, uio, fmt, ap);
#endif /* DSPRINTF_DEBUG */
  va_end (ap);
}
