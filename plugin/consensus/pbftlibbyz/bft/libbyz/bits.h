/*
\section{Bit-Fields of various Sizes}

This file contains definitions of bitfields of various sizes.  The
definitions vary by compiler and architectures.  The unsigned types
defined here are "Ubits8", "Ubits16", "Ubits32", and "Ubits64".  
Note that unsigned arithmetic can lead to unexpected results.

*/

#ifndef _BITS_H
#define _BITS_H

#include "assert.h"

#define byte_bits 8
typedef unsigned int  Uint;

typedef unsigned char  Ubits8;
typedef unsigned short Ubits16;
typedef unsigned int   Ubits32;

typedef char  Bits8;
typedef short Bits16;
typedef int   Bits32;

#ifdef __alpha

#define LONG_SIGN_BIT_MASK 0x8000000000000000UL
#define INT_SIGN_BIT_MASK 0x80000000UL

#define INT_BITS 32
#define LONG_BITS 64

#define POINTER_INT long

#define _BITS_H_OK

typedef long int Bits64;
/* Lformat is the format string for printing purposes */
#endif


#ifdef __x86_64__

#define LONG_SIGN_BIT_MASK 0x8000000000000000UL
#define INT_SIGN_BIT_MASK 0x80000000UL

#define INT_BITS 32
#define LONG_BITS 64

#define POINTER_INT long

#define _BITS_H_OK

typedef long Bits64;

#endif // __X86_64__


#ifdef __i386__ 

#define LONG_SIGN_BIT_MASK 0x80000000UL
#define INT_SIGN_BIT_MASK 0x80000000UL

#define INT_BITS 32
#define LONG_BITS 32

#define POINTER_INT int

#define _BITS_H_OK

typedef long long Bits64;


#endif

#undef _BITS_H_OK


#endif
