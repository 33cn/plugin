/* -----------------------------------------------------------------------
 * 
 * umac.c -- C Implementation UMAC Message Authentication
 *
 * Version 0.01 of draft-krovetz-umac-00.txt -- 2000 August
 *
 * For a full description of UMAC message authentication see the UMAC
 * world-wide-web page at http://www.cs.ucdavis.edu/~rogaway/umac
 * Please report bugs and suggestions to the UMAC webpage.
 *
 * Copyright (c) 1999-2000 Ted Krovetz (tdk@acm.org)
 *
 * This code is made available on the Internet by Phillip Rogaway      
 * (rogaway@cs.ucdavis.edu).  Permission to download this code to 
 * any site outside the United States is granted only if you agree 
 * to use this code solely for its intended purpose of generating 
 * a message authentication code.              
 *                                                                 
 * Permission to use, copy, modify, and distribute this software and  
 * its documentation for any purpose and without fee, is hereby granted,
 * provided that the above copyright notice appears in all copies and  
 * that both that copyright notice and this permission notice appear   
 * in supporting documentation, and that the names of the University of
 * California and Ted Krovetz not be used in advertising or publicity  
 * pertaining to distribution of the software without specific,        
 * written prior permission.                                          
 *                                                                   
 * The Regents of the University of California and Ted Krovetz disclaim 
 * all warranties with regard to this software, including all implied
 * warranties of merchantability and fitness.  In no event shall the  
 * University of California or Ted Krovetz be liable for any special,  
 * indirect or consequential damages or any damages whatsoever resulting
 * from loss of use, data or profits, whether in an action of contract,
 * negligence or other tortious action, arising out of or in connection
 * with the use or performance of this software.
 * 
 * ---------------------------------------------------------------------- */
 
/* ---------------------------------------------------------------------- */
/* -- Global Includes --------------------------------------------------- */
/* ---------------------------------------------------------------------- */

#include "umac.h"
#include <string.h>
#include <stdlib.h>
#include "bits.h"

/* ---------------------------------------------------------------------- */
/* --- User Switches ---------------------------------------------------- */
/* ---------------------------------------------------------------------- */

/* Following is the list of UMAC parameters supported by this code.       
 * The following parameters are fixed in this implementation.             
 *                                                                        
 *      ENDIAN_FAVORITE_LITTLE  = 1                                       
 *      L1-OPERATIONS-SIGN      = SIGNED   (when WORD_LEN == 2)           
 *      L1-OPERATIONS-SIGN      = UNSIGNED (when WORD_LEN == 4)           
 */
#define WORD_LEN                4   /* 2  | 4                             */
#define UMAC_OUTPUT_LEN         8   /* 4  | 8  | 12  | 16                 */
#define L1_KEY_LEN           1024   /* 32 | 64 | 128 | ... | 2^28         */
#define UMAC_KEY_LEN           16   /* 16 | 32                            */

/* To produce a prefix of a tag rather than the entire tag defined
 * by the above parameters, set the following constant to a number
 * less than UMAC_OUTPUT_LEN.
 */
#define UMAC_PREFIX_LEN  UMAC_OUTPUT_LEN

/* This file implements UMAC in ANSI C if the compiler supports 64-bit
 * integers. To accellerate the execution of the code, architecture-
 * specific replacements have been supplied for some compiler/instruction-
 * set combinations. To enable the features of these replacements, the
 * following compiler directives must be set appropriately. Some compilers
 * include "intrinsic" support of basic operations like register rotation,
 * byte reversal, or vector SIMD manipulation. To enable these intrinsics
 * set USE_C_AND_INTRINSICS to 1. Most compilers also allow for inline
 * assembly in the C code. To allow intrinsics and/or assembly routines
 * (whichever is faster) set only USE_C_AND_ASSEMBLY to 1.
 */
#define USE_C_ONLY            0  /* ANSI C and 64-bit integers req'd */
#define USE_C_AND_INTRINSICS  0  /* Intrinsics for rotation, MMX, etc.    */
#define USE_C_AND_ASSEMBLY    1  /* Intrinsics and assembly */

#if (USE_C_ONLY + USE_C_AND_INTRINSICS + USE_C_AND_ASSEMBLY != 1)
#error -- Only one setting may be nonzero
#endif

#define RUN_TESTS             0  /* Run basic correctness/ speed tests    */
#define HASH_ONLY             0  /* Only universal hash data, don't MAC   */

/* ---------------------------------------------------------------------- */
/* --- Primitive Data Types ---                                           */
/* ---------------------------------------------------------------------- */

#ifdef _MSC_VER
typedef char               INT8;   /* 1 byte   */
typedef __int16            INT16;  /* 2 byte   */
typedef unsigned __int16   UINT16; /* 2 byte   */
typedef __int32            INT32;  /* 4 byte   */
typedef unsigned __int32   UINT32; /* 4 byte   */
typedef __int64            INT64;  /* 8 bytes  */
typedef unsigned __int64   UINT64; /* 8 bytes  */
#else
typedef char               INT8;   /* 1 byte   */
typedef short              INT16;  /* 2 byte   */
typedef unsigned short     UINT16; /* 2 byte   */
typedef int                INT32;  /* 4 byte   */
typedef unsigned int       UINT32; /* 4 byte   */
typedef long long          INT64;  /* 8 bytes  */
typedef unsigned long long UINT64; /* 8 bytes  */
#endif
typedef long               WORD;   /* Register */
typedef unsigned long      UWORD;  /* Register */

/* ---------------------------------------------------------------------- */
/* --- Derived Constants ------------------------------------------------ */
/* ---------------------------------------------------------------------- */

#if (WORD_LEN == 4)

typedef INT32   SMALL_WORD;  
typedef UINT32  SMALL_UWORD;  
typedef INT64   LARGE_WORD;
typedef UINT64  LARGE_UWORD;

#elif (WORD_LEN == 2)
 
typedef INT16   SMALL_WORD;  
typedef UINT16  SMALL_UWORD;  
typedef INT32   LARGE_WORD;
typedef UINT32  LARGE_UWORD;

#endif

/* How many iterations, or streams, are needed to produce UMAC_OUTPUT_LEN
 * and UMAC_PREFIX_LEN bytes of output
 */
#define PREFIX_STREAMS    (UMAC_PREFIX_LEN / WORD_LEN)
#define OUTPUT_STREAMS    (UMAC_OUTPUT_LEN / WORD_LEN)

/* Three compiler environments are supported for accellerated
 * implementations: GNU gcc and Microsoft Visual C++ (and copycats) on x86,
 * and Metrowerks on PowerPC.
 */
#define GCC_X86         (__GNUC__ && __i386__)      /* GCC on IA-32       */
#define MSC_X86         (_MSC_VER && _M_IX86)       /* Microsoft on IA-32 */
#define MW_PPC          ((__MWERKS__ || __MRC__) && __POWERPC__)
                                                    /* Metrowerks on PPC  */
/* ---------------------------------------------------------------------- */
/* --- Host Computer Endian Definition ---------------------------------- */
/* ---------------------------------------------------------------------- */

/* Message "words" are read from memory in an endian-specific manner.     */
/* For this implementation to behave correctly, __LITTLE_ENDIAN__ must    */
/* be set true if the host computer is little-endian.                     */

#if __i386__ || __alpha__ || _M_IX86 || __LITTLE_ENDIAN
#define __LITTLE_ENDIAN__ 1
#else
#define __LITTLE_ENDIAN__ 0
#endif

/* ---------------------------------------------------------------------- */
/* ----- RC6 Function Family Constants ---------------------------------- */
/* ---------------------------------------------------------------------- */

#define RC6_KEY_BYTES    UMAC_KEY_LEN
#define RC6_ROUNDS       20       
#define RC6_KEY_WORDS    (UMAC_KEY_LEN/4)      
#define RC6_TABLE_WORDS  (2*RC6_ROUNDS+4)  
#define RC6_P            0xb7e15163u
#define RC6_Q            0x9e3779b9u

#define PRF_BLOCK_LEN   16                /* Constants for KDF and PDF */
#define PRF_KEY_LEN     (RC6_TABLE_WORDS*4)

/* ---------------------------------------------------------------------- */
/* ----- Poly hash and Inner-Product hash Constants --------------------- */
/* ---------------------------------------------------------------------- */

/* Primes */
const UINT32 p19 = (((UINT32)1 << 19) - (UINT32)1); /* 2^19 -  1 */
const UINT32 p32 = ((UINT32)0 - (UINT32)5);         /* 2^32 -  5 */

const UINT64 p36 = (((UINT64)1 << 36) - (UINT64)5); /* 2^36 -  5 */
const UINT64 p64 = ((UINT64)0 - (UINT64)59);        /* 2^64 - 59 */

/* Masks */
const UINT32 m19 = ((1ul << 19) - 1);        /* The low 19 of 32 bits */
const UINT64 m36 = (((UINT64)1 << 36) - 1);  /* The low 36 of 64 bits */

/* ---------------------------------------------------------------------- */
/* ---------------------------------------------------------------------- */
/* ----- Architecture Specific Routines --------------------------------- */
/* ---------------------------------------------------------------------- */
/* ---------------------------------------------------------------------- */

#if (GCC_X86 && ! USE_C_ONLY)
#include "umac_gcc_x86_incl.c"
#elif (MSC_X86 && ! USE_C_ONLY)
#include "umac_msc_x86_incl.c"
#elif (MW_PPC && ! USE_C_ONLY)
#include "umac_mw_ppc_incl.c"
#endif

/* ---------------------------------------------------------------------- */
/* ---------------------------------------------------------------------- */
/* ----- Primitive Routines --------------------------------------------- */
/* ---------------------------------------------------------------------- */
/* ---------------------------------------------------------------------- */

/* ---------------------------------------------------------------------- */
/* --- 32-Bit Rotation operators                                          */
/* ---------------------------------------------------------------------- */

/* Good compilers can detect when a rotate
 * is being constructed from bitshifting and bitwise OR and output the
 * assembly rotates. Other compilers require assembly or C intrinsics.
 */

#if (USE_C_ONLY || ! ARCH_ROTL)

#define ROTL32_VAR(r,n)  (((r) << ((n) & 31)) | \
                         ((UINT32)(r) >> (32 - ((n) & 31))))
#define ROTL32_CONST(r,n) (((r) <<  (n))       | \
                          ((UINT32)(r) >> (32 -  (n))))

#endif

/* ---------------------------------------------------------------------- */
/* --- 32-bit by 32-bit to 64-bit Multiplication ------------------------ */
/* ---------------------------------------------------------------------- */

#if (USE_C_ONLY || ! ARCH_MUL64)

static UINT64 MUL64(UINT32 a, UINT32 b)
{
    return (UINT64)a * (UINT64)b;
}
               
#endif

/* ---------------------------------------------------------------------- */
/* --- Endian Conversion --- Forcing assembly on some platforms           */
/* ---------------------------------------------------------------------- */

/* Lots of endian reversals happen in UMAC. PowerPC and Intel Architechture
 * both support efficient endian conversion, but compilers seem unable to
 * automatically utilize the efficient assembly opcodes. The architechture-
 * specific versions utilize them.
 */

#if (USE_C_ONLY || ! ARCH_ENDIAN_LS)

static UINT32 LOAD_UINT32_REVERSED(void *ptr)
{
    UINT32 temp = *(UINT32 *)ptr;
    temp = (temp >> 24) | ((temp & 0x00FF0000) >> 8 )
         | ((temp & 0x0000FF00) << 8 ) | (temp << 24);
    return (UINT32)temp;
}
               
static void STORE_UINT32_REVERSED(void *ptr, UINT32 x)
{
    UINT32 i = (UINT32)x;
    *(UINT32 *)ptr = (i >> 24) | ((i & 0x00FF0000) >> 8 )
                   | ((i & 0x0000FF00) << 8 ) | (i << 24);
}

static UINT16 LOAD_UINT16_REVERSED(void *ptr)
{
    UINT16 temp = *(UINT16 *)ptr;
    temp = (temp >> 8) | (temp << 8);
    return (UINT16)temp;
}
               
static void STORE_UINT16_REVERSED(void *ptr, UINT16 x)
{
    UINT16 temp = (UINT16)x;
    *(UINT16 *)ptr = (temp >> 8) | (temp << 8);
}
               
#endif

/* The following routines use the above reversal-primitives to do the right
 * thing on endian specific load and stores.
 */

static UINT16 LOAD_UINT16_LITTLE(void *ptr)
{
    #if ( ! __LITTLE_ENDIAN__)
    return LOAD_UINT16_REVERSED(ptr);
    #else
    return *(UINT16 *)ptr;
    #endif
}

static void STORE_UINT16_BIG(void *ptr, UINT16 x)
{
    #if (__LITTLE_ENDIAN__)
    STORE_UINT16_REVERSED(ptr,x);
    #else
    *(UINT16 *)ptr = x;
    #endif
}

static UINT32 LOAD_UINT32_LITTLE(void *ptr)
{
    #if ( ! __LITTLE_ENDIAN__)
    return LOAD_UINT32_REVERSED(ptr);
    #else
    return *(UINT32 *)ptr;
    #endif
}

static UINT32 LOAD_UINT32_BIG(void *ptr)
{
    #if (__LITTLE_ENDIAN__)
    return LOAD_UINT32_REVERSED(ptr);
    #else
    return *(UINT32 *)ptr;
    #endif
}

static void STORE_UINT32_BIG(void *ptr, UINT32 x)
{
    #if (__LITTLE_ENDIAN__)
    STORE_UINT32_REVERSED(ptr,x);
    #else
    *(UINT32 *)ptr = x;
    #endif
}

static void STORE_UINT32_LITTLE(void *ptr, UINT32 x)
{
    #if ( ! __LITTLE_ENDIAN__)
    STORE_UINT32_REVERSED(ptr,x);
    #else
    *(UINT32 *)ptr = x;
    #endif
}

/* ---------------------------------------------------------------------- */
/* ---------------------------------------------------------------------- */
/* ----- Begin Cryptographic Primitive Section -------------------------- */
/* ---------------------------------------------------------------------- */
/* ---------------------------------------------------------------------- */

/* The UMAC specification requires the use of AES for it's cryptographic
 * component. Until AES is finalized, we use RC6, and AES finalist, in its
 * place.
 */

/* ---------------------------------------------------------------------- */
#if (USE_C_ONLY || ! ARCH_RC6)
/* ---------------------------------------------------------------------- */

#define RC6_BLOCK(a,b,c,d,n)    \
        t = b*(2*b+1);          \
        t = ROTL32_CONST(t,5);    \
        u = d*(2*d+1);          \
        u = ROTL32_CONST(u,5);    \
        a ^= t;                 \
        a = ROTL32_VAR(a,u);  \
        a += s[n];              \
        c ^= u;                 \
        c = ROTL32_VAR(c,t);  \
        c += s[n+1];

static void RC6(UINT32 S[], void *pt, void *ct)
{ 
    const UINT32 *s = (UINT32 *)S;
    UINT32 A = LOAD_UINT32_LITTLE((UINT32 *)pt  );
    UINT32 B = LOAD_UINT32_LITTLE((UINT32 *)pt+1) + s[0];
    UINT32 C = LOAD_UINT32_LITTLE((UINT32 *)pt+2);
    UINT32 D = LOAD_UINT32_LITTLE((UINT32 *)pt+3) + s[1];
    UINT32 t,u;
    
    RC6_BLOCK(A,B,C,D, 2)
    RC6_BLOCK(B,C,D,A, 4)
    RC6_BLOCK(C,D,A,B, 6)
    RC6_BLOCK(D,A,B,C, 8)

    RC6_BLOCK(A,B,C,D,10)
    RC6_BLOCK(B,C,D,A,12)
    RC6_BLOCK(C,D,A,B,14)
    RC6_BLOCK(D,A,B,C,16)

    RC6_BLOCK(A,B,C,D,18)
    RC6_BLOCK(B,C,D,A,20)
    RC6_BLOCK(C,D,A,B,22)
    RC6_BLOCK(D,A,B,C,24)

    RC6_BLOCK(A,B,C,D,26)
    RC6_BLOCK(B,C,D,A,28)
    RC6_BLOCK(C,D,A,B,30)
    RC6_BLOCK(D,A,B,C,32)

    RC6_BLOCK(A,B,C,D,34)
    RC6_BLOCK(B,C,D,A,36)
    RC6_BLOCK(C,D,A,B,38)
    RC6_BLOCK(D,A,B,C,40)

    A += s[42];
    C += s[43];
    STORE_UINT32_LITTLE((UINT32 *)ct  , A); 
    STORE_UINT32_LITTLE((UINT32 *)ct+1, B); 
    STORE_UINT32_LITTLE((UINT32 *)ct+2, C); 
    STORE_UINT32_LITTLE((UINT32 *)ct+3, D); 
} 

/* ---------------------------------------------------------------------- */
#endif
/* ---------------------------------------------------------------------- */

static void RC6_SETUP(INT8 *K, UINT32 S[])
{
    UWORD i, j, k;
    UINT32 A, B, L[RC6_KEY_WORDS]; 

    /* Load little endian key into L */
    #if (__LITTLE_ENDIAN__)
    memcpy(L, K, RC6_KEY_BYTES);
    #else
    for (i = 0; i < RC6_KEY_WORDS; i++)
        STORE_UINT32_LITTLE(L+i, LOAD_UINT32_BIG((UINT32 *)K+i));
    #endif
    
    /* Preload S with P and Q derived constants */
    S[0]=RC6_P;
    for (i = 1; i < RC6_TABLE_WORDS; i++)
        S[i] = S[i-1] + RC6_Q;
    
    /* Mix L into S */
    A=B=i=j=k=0;
    for (k = 0; k < 3*RC6_TABLE_WORDS; k++) {
        A = S[i] = ROTL32_CONST(S[i]+(A+B),3);  
        B = L[j] = ROTL32_VAR(L[j]+(A+B),(A+B));
        i=(i+1)%RC6_TABLE_WORDS;
        j=(j+1)%RC6_KEY_WORDS;
    } 
} 

/* ---------------------------------------------------------------------- */
/* ---------------------------------------------------------------------- */
/* ----- Begin KDF & PDF Section ---------------------------------------- */
/* ---------------------------------------------------------------------- */
/* ---------------------------------------------------------------------- */

/* The user-supplied UMAC key is stretched using AES in an output feedback
 * mode to supply all random bits needed by UMAC. The kdf function takes
 * and AES's internal key representation 'key' and writes a stream of
 * 'nbytes' bytes to the memory pointed at by 'buffer_ptr'. Each distinct
 * 'index' causes a distinct byte stream.
 */
static void kdf(void *buffer_ptr, INT8 *key, INT8 index, int nbytes)
{
    INT8 chain[PRF_BLOCK_LEN] = {0};
    INT8 *dst_buf = (INT8 *)buffer_ptr;
    
    chain[PRF_BLOCK_LEN-1] = index;
    
    while (nbytes >= PRF_BLOCK_LEN) {
        RC6((UINT32 *)key,chain,chain);
        memcpy(dst_buf,chain,PRF_BLOCK_LEN);
        nbytes -= PRF_BLOCK_LEN;
        dst_buf += PRF_BLOCK_LEN;
    }
    if (nbytes) {
        RC6((UINT32 *)key,chain,chain);
        memcpy(dst_buf,chain,nbytes);
    }
}

/* The final UHASH result is XOR'd with the output of a pseudorandom
 * function. Here, we use AES to generate random output and 
 * xor the appropriate bytes depending on the last bits of nonce.
 * This code is optimized for sequential, increasing big-endian nonces.
 */

typedef struct {
    INT8 cache[PRF_BLOCK_LEN];
    INT8 nonce[PRF_BLOCK_LEN];
    INT8 prf_key[PRF_KEY_LEN]; /* Expanded AES Key                       */
} pdf_ctx;

static void pdf_init(pdf_ctx *pc, INT8 *prf_key)
{
    INT8 buf[UMAC_KEY_LEN];
    
    kdf(buf, prf_key, (INT8)128, UMAC_KEY_LEN);
    RC6_SETUP(buf, (UINT32 *)pc->prf_key);
    
    /* Initialize pdf and cache */
    memset(pc->nonce, 0, sizeof(pc->nonce));
    RC6((UINT32 *)pc->prf_key, pc->nonce, pc->cache);
}

static void pdf_gen_xor(pdf_ctx *pc, INT8 nonce[8], INT8 buf[8])
{
    /* This implementation requires UMAC_OUTPUT_LEN to divide PRF_BLOCK_LEN
     * or be at least 1/2 its length. 'index' indicates that we'll be using
     * the index-th UMAC_OUTPUT_LEN-length element of the AES output. If
     * last time around we returned the index-1 element, then we may have
     * the result in the cache already.
     */
    INT8 tmp_nonce_lo[4];
    int index = nonce[7] % (PRF_BLOCK_LEN / UMAC_OUTPUT_LEN);
    
    *(UINT32 *)tmp_nonce_lo = ((UINT32 *)nonce)[1];
    tmp_nonce_lo[3] ^= index; /* zero some bits */
    
    if ( (((UINT32 *)tmp_nonce_lo)[0] != ((UINT32 *)pc->nonce)[1]) ||
         (((UINT32 *)nonce)[0] != ((UINT32 *)pc->nonce)[0]) )
    {
        ((UINT32 *)pc->nonce)[0] = ((UINT32 *)nonce)[0];
        ((UINT32 *)pc->nonce)[1] = ((UINT32 *)tmp_nonce_lo)[0];
        RC6((UINT32 *)pc->prf_key, pc->nonce, pc->cache);
    }
    
    #if (UMAC_OUTPUT_LEN == 2)
        *((UINT16 *)buf) ^= ((UINT16 *)pc->cache)[index];
    #elif (UMAC_OUTPUT_LEN == 4)
        *((UINT32 *)buf) ^= ((UINT32 *)pc->cache)[index];
    #elif (UMAC_OUTPUT_LEN == 8)
        *((UINT64 *)buf) ^= ((UINT64 *)pc->cache)[index];
    #elif (UMAC_OUTPUT_LEN == 12) 
        ((UINT64 *)buf)[0] ^= ((UINT64 *)pc->cache)[0];
        ((UINT32 *)buf)[2] ^= ((UINT32 *)pc->cache)[2];
    #elif (UMAC_OUTPUT_LEN == 16) 
        ((UINT64 *)buf)[0] ^= ((UINT64 *)pc->cache)[0];
        ((UINT64 *)buf)[1] ^= ((UINT64 *)pc->cache)[1];
    #else
        #error only 2,4,8,12,16 byte output supported.
    #endif
}

/* ---------------------------------------------------------------------- */
/* ---------------------------------------------------------------------- */
/* ----- Begin NH Hash Section ------------------------------------------ */
/* ---------------------------------------------------------------------- */
/* ---------------------------------------------------------------------- */

/* The NH-based hash functions used in UMAC are described in the UMAC paper
 * and specification, both of which can be found at the UMAC website.     
 * The interface to this implementation has two         
 * versions, one expects the entire message being hashed to be passed
 * in a single buffer and returns the hash result immediately. The second
 * allows the message to be passed in a sequence of buffers. In the          
 * muliple-buffer interface, the client calls the routine nh_update() as     
 * many times as necessary. When there is no more data to be fed to the   
 * hash, the client calls nh_final() which calculates the hash output.    
 * Before beginning another hash calculation the nh_reset() routine       
 * must be called. The single-buffer routine, nh(), is equivalent to  
 * the sequence of calls nh_update() and nh_final(); however it is        
 * optimized and should be prefered whenever the multiple-buffer interface
 * is not necessary. When using either interface, it is the client's         
 * responsability to pass no more than L1_KEY_LEN bytes per hash result.            
 *                                                                        
 * The routine nh_init() initializes the nh_ctx data structure and        
 * must be called once, before any other PDF routine.                     
 */
 
 /* The "nh_aux_*" routines do the actual NH hashing work. The versions
  * prefixed with "nh_aux_hb" expect the buffers passed to them to be a
  * multiple of HASH_BUF_BYTES, allowing increased optimization. The versions
  * prefixed with "nh_aux_pb" expect buffers to be multiples of
  * L1_PAD_BOUNDARY. These routines produce output for all PREFIX_STREAMS
  * NH iterations in one call, allowing the parallel implementation of the
  * streams.
  */
#if   (UMAC_PREFIX_LEN == 2)
#define nh_aux   nh_aux_4
#elif (UMAC_PREFIX_LEN == 4)
#define nh_aux   nh_aux_8
#elif (UMAC_PREFIX_LEN == 8)
#define nh_aux   nh_aux_16
#elif (UMAC_PREFIX_LEN == 12)
#define nh_aux   nh_aux_24
#elif (UMAC_PREFIX_LEN == 16)
#define nh_aux   nh_aux_32
#endif

#define L1_KEY_SHIFT         16     /* Toeplitz key shift between streams */
#define L1_PAD_BOUNDARY      32     /* pad message to boundary multiple   */
#define ALLOC_BOUNDARY       16     /* Keep buffers aligned to this       */
#define HASH_BUF_BYTES      128     /* nh_aux_hb buffer multiple          */

/* How many extra bytes are needed for Toeplitz shift? */
#define TOEPLITZ_EXTRA       ((PREFIX_STREAMS - 1) * L1_KEY_SHIFT)

typedef struct {
    INT8  nh_key [L1_KEY_LEN + TOEPLITZ_EXTRA]; /* NH Key */
    INT8  data   [HASH_BUF_BYTES];    /* Incomming data buffer            */
    int next_data_empty;    /* Bookeeping variable for data buffer.       */
    int bytes_hashed;        /* Bytes (out of L1_KEY_LEN) incorperated.   */
    LARGE_UWORD state[PREFIX_STREAMS];               /* on-line state     */
} nh_ctx;


/* ---------------------------------------------------------------------- */
#if (WORD_LEN == 4)
/* ---------------------------------------------------------------------- */

/* ---------------------------------------------------------------------- */
#if (USE_C_ONLY || ! ARCH_NH32)
/* ---------------------------------------------------------------------- */

static void nh_aux_8(void *kp, void *dp, void *hp, UINT32 dlen)
/* NH hashing primitive. Previous (partial) hash result is loaded and     
 * then stored via hp pointer. The length of the data pointed at by "dp",
 * "dlen", is guaranteed to be divisible by L1_PAD_BOUNDARY (32).  Key
 * is expected to be endian compensated in memory at key setup.    
 */
{
  UINT64 h;
  UWORD c = dlen / 32;
  UINT32 *k = (UINT32 *)kp;
  UINT32 *d = (UINT32 *)dp;
  UINT32 d0,d1,d2,d3,d4,d5,d6,d7;
  UINT32 k0,k1,k2,k3,k4,k5,k6,k7;

  h = *((UINT64 *)hp);
  do {
    d0 = LOAD_UINT32_LITTLE(d+0); d1 = LOAD_UINT32_LITTLE(d+1);
    d2 = LOAD_UINT32_LITTLE(d+2); d3 = LOAD_UINT32_LITTLE(d+3);
    d4 = LOAD_UINT32_LITTLE(d+4); d5 = LOAD_UINT32_LITTLE(d+5);
    d6 = LOAD_UINT32_LITTLE(d+6); d7 = LOAD_UINT32_LITTLE(d+7);
    k0 = *(k+0); k1 = *(k+1); k2 = *(k+2); k3 = *(k+3);
    k4 = *(k+4); k5 = *(k+5); k6 = *(k+6); k7 = *(k+7);
    h += MUL64((k0 + d0), (k4 + d4));
    h += MUL64((k1 + d1), (k5 + d5));
    h += MUL64((k2 + d2), (k6 + d6));
    h += MUL64((k3 + d3), (k7 + d7));

    d += 8;
    k += 8;
  } while (--c);
  *((UINT64 *)hp) = h;
}

/* ---------------------------------------------------------------------- */

static void nh_aux_16(void *kp, void *dp, void *hp, UINT32 dlen)
/* Same as nh_aux_8, but two streams are handled in one pass,
 * reading and writing 16 bytes of hash-state per call.
 */
{
  UINT64 h1,h2;
  UWORD c = dlen / 32;
  UINT32 *k = (UINT32 *)kp;
  UINT32 *d = (UINT32 *)dp;
  UINT32 d0,d1,d2,d3,d4,d5,d6,d7;
  UINT32 k0,k1,k2,k3,k4,k5,k6,k7,
        k8,k9,k10,k11;

  h1 = *((UINT64 *)hp);
  h2 = *((UINT64 *)hp + 1);
  k0 = *(k+0); k1 = *(k+1); k2 = *(k+2); k3 = *(k+3);
  do {
    d0 = LOAD_UINT32_LITTLE(d+0); d1 = LOAD_UINT32_LITTLE(d+1);
    d2 = LOAD_UINT32_LITTLE(d+2); d3 = LOAD_UINT32_LITTLE(d+3);
    d4 = LOAD_UINT32_LITTLE(d+4); d5 = LOAD_UINT32_LITTLE(d+5);
    d6 = LOAD_UINT32_LITTLE(d+6); d7 = LOAD_UINT32_LITTLE(d+7);
    k4 = *(k+4); k5 = *(k+5); k6 = *(k+6); k7 = *(k+7);
    k8 = *(k+8); k9 = *(k+9); k10 = *(k+10); k11 = *(k+11);

    h1 += MUL64((k0 + d0), (k4 + d4));
    h2 += MUL64((k4 + d0), (k8 + d4));

    h1 += MUL64((k1 + d1), (k5 + d5));
    h2 += MUL64((k5 + d1), (k9 + d5));

    h1 += MUL64((k2 + d2), (k6 + d6));
    h2 += MUL64((k6 + d2), (k10 + d6));

    h1 += MUL64((k3 + d3), (k7 + d7));
    h2 += MUL64((k7 + d3), (k11 + d7));

    k0 = k8; k1 = k9; k2 = k10; k3 = k11;

    d += 8;
    k += 8;
  } while (--c);
  ((UINT64 *)hp)[0] = h1;
  ((UINT64 *)hp)[1] = h2;
}

/* ---------------------------------------------------------------------- */


/* ---------------------------------------------------------------------- */



/* ---------------------------------------------------------------------- */
#endif
/* ---------------------------------------------------------------------- */

/* ---------------------------------------------------------------------- */
/* ----- NH16 Universal Hash -------------------------------------------- */
/* ---------------------------------------------------------------------- */

#else


/* ---------------------------------------------------------------------- */
#if (USE_C_ONLY || ! ARCH_NH16)
/* ---------------------------------------------------------------------- */

static void nh_aux_4(void *kp, void *dp, void *hp, UINT32 dlen)
/* NH hashing primitive. Previous (partial) hash result is loaded and     
 * then stored via hp pointer. The length of the data pointed at by "dp",
 * "dlen", is guaranteed to be divisible by L1_PAD_BOUNDARY (32).  Key
 * is expected to be endian compensated in memory at key setup.    
 */
{
    INT32 h;
    UINT32 c = dlen / 32;
    INT16 *k = (INT16 *)kp;
    INT16 *d = (INT16 *)dp;

    h = *(INT32 *)hp;
    do {
        h += (INT32)(INT16)(*(k+0)  + LOAD_UINT16_LITTLE(d+0)) * 
             (INT32)(INT16)(*(k+8)  + LOAD_UINT16_LITTLE(d+8));
        h += (INT32)(INT16)(*(k+1)  + LOAD_UINT16_LITTLE(d+1)) * 
             (INT32)(INT16)(*(k+9)  + LOAD_UINT16_LITTLE(d+9));
        h += (INT32)(INT16)(*(k+2)  + LOAD_UINT16_LITTLE(d+2)) * 
             (INT32)(INT16)(*(k+10) + LOAD_UINT16_LITTLE(d+10));
        h += (INT32)(INT16)(*(k+3)  + LOAD_UINT16_LITTLE(d+3)) * 
             (INT32)(INT16)(*(k+11) + LOAD_UINT16_LITTLE(d+11));
        h += (INT32)(INT16)(*(k+4)  + LOAD_UINT16_LITTLE(d+4)) * 
             (INT32)(INT16)(*(k+12) + LOAD_UINT16_LITTLE(d+12));
        h += (INT32)(INT16)(*(k+5)  + LOAD_UINT16_LITTLE(d+5)) * 
             (INT32)(INT16)(*(k+13) + LOAD_UINT16_LITTLE(d+13));
        h += (INT32)(INT16)(*(k+6)  + LOAD_UINT16_LITTLE(d+6)) * 
             (INT32)(INT16)(*(k+14) + LOAD_UINT16_LITTLE(d+14));
        h += (INT32)(INT16)(*(k+7)  + LOAD_UINT16_LITTLE(d+7)) * 
             (INT32)(INT16)(*(k+15) + LOAD_UINT16_LITTLE(d+15));
        d += 16;
        k += 16;
    } while (--c);
    *(INT32 *)hp = h;
}

/* ---------------------------------------------------------------------- */

static void nh_aux_8(void *kp, void *dp, void *hp, UINT32 dlen)
/* Same as nh_aux_4, but two streams are handled in one pass,
 * reading and writing 8 bytes of hash-state per call.
 */
{
    nh_aux_4(kp,dp,hp,dlen);
    nh_aux_4((INT8 *)kp+16,dp,(INT8 *)hp+4,dlen);
}

/* ---------------------------------------------------------------------- */

static void nh_aux_16(void *kp, void *dp, void *hp, UINT32 dlen)
/* Same as nh_aux_8, but four streams are handled in one pass,
 * reading and writing 16 bytes of hash-state per call.
 */
{
    nh_aux_4(kp,dp,hp,dlen);
    nh_aux_4((INT8 *)kp+16,dp,(INT8 *)hp+4,dlen);
    nh_aux_4((INT8 *)kp+32,dp,(INT8 *)hp+8,dlen);
    nh_aux_4((INT8 *)kp+48,dp,(INT8 *)hp+12,dlen);
}

/* ---------------------------------------------------------------------- */



/* ---------------------------------------------------------------------- */
#endif  /* UMAC-MMX hashes              */
/* ---------------------------------------------------------------------- */
#endif  /* UMAC-MMX hashes              */
/* ---------------------------------------------------------------------- */

/* The following two routines use previously defined ones to build up longer
 * outputs of 24 or 32 bytes.
 */
 

/* ---------------------------------------------------------------------- */

static void nh_aux_24(void *kp, void *dp, void *hp, UINT32 dlen)
{
    nh_aux_16(kp,dp,hp,dlen);
    nh_aux_8((INT8 *)kp+((8/WORD_LEN)*L1_KEY_SHIFT),dp,
                                           (INT8 *)hp+16,dlen);
}

/* ---------------------------------------------------------------------- */

/* ---------------------------------------------------------------------- */

static void nh_aux_32(void *kp, void *dp, void *hp, UINT32 dlen)
{
    nh_aux_16(kp,dp,hp,dlen);
    nh_aux_16((INT8 *)kp+((8/WORD_LEN)*L1_KEY_SHIFT),
                                      dp,(INT8 *)hp+16,dlen);
}

/* ---------------------------------------------------------------------- */


/* ---------------------------------------------------------------------- */

static void nh_transform(nh_ctx *hc, INT8 *buf, UINT32 nbytes)
/* This function is a wrapper for the primitive NH hash functions. It takes
 * as argument "hc" the current hash context and a buffer which must be a
 * multiple of L1_PAD_BOUNDARY. The key passed to nh_aux is offset
 * appropriately according to how much message has been hashed already.
 */
{
    INT8 *key;
  
    if (nbytes) {
        key = hc->nh_key + hc->bytes_hashed;
        nh_aux(key, buf, hc->state, nbytes);
    }
}

/* ---------------------------------------------------------------------- */

static void endian_convert(void *buf, UWORD bpw, UINT32 num_bytes)
/* We endian convert the keys on little-endian computers to               */
/* compensate for the lack of big-endian memory reads during hashing.     */
{
    UWORD iters = num_bytes / bpw;
    if (bpw == 2) {
        UINT16 *p = (UINT16 *)buf;
        do {
            *p = ((UINT16)*p >> 8) | (*p << 8);
            p++;
        } while (--iters);
    } else if (bpw == 4) {
        UINT32 *p = (UINT32 *)buf;
        do {
            *p = LOAD_UINT32_REVERSED(p);
            p++;
        } while (--iters);
    } else if (bpw == 8) {
        UINT32 *p = (UINT32 *)buf;
        UINT32 t;
        do {
            t = LOAD_UINT32_REVERSED(p+1);
            p[1] = LOAD_UINT32_REVERSED(p);
            p[0] = t;
            p += 2;
        } while (--iters);
    }
}
#if (__LITTLE_ENDIAN__)
#define endian_convert_if_le(x,y,z) endian_convert((x),(y),(z))
#else
#define endian_convert_if_le(x,y,z) do{}while(0)  /* Do nothing */
#endif

/* ---------------------------------------------------------------------- */

static void nh_reset(nh_ctx *hc)
/* Reset nh_ctx to ready for hashing of new data */
{
    hc->bytes_hashed = 0;
    hc->next_data_empty = 0;
    hc->state[0] = 0;
    #if (PREFIX_STREAMS > 1)
    hc->state[1] = 0;
    #if (PREFIX_STREAMS > 2)
    hc->state[2] = 0;
    hc->state[3] = 0;
    #if (PREFIX_STREAMS > 4)
    hc->state[4] = 0;
    hc->state[5] = 0;
    #if (PREFIX_STREAMS > 6)
    hc->state[6] = 0;
    hc->state[7] = 0;
    #endif
    #endif
    #endif
    #endif
}

/* ---------------------------------------------------------------------- */

static void nh_init(nh_ctx *hc, INT8 *prf_key)
/* Generate nh_key, endian convert and reset to be ready for hashing.   */
{
    
    kdf(hc->nh_key, prf_key, 0, sizeof(hc->nh_key));
    endian_convert_if_le(hc->nh_key, WORD_LEN, sizeof(hc->nh_key));
    #if (ARCH_KEY_MODIFICATION)
    /* Some specialized code may need the key in an altered format.
     * They will define ARCH_KEY_MODIFICATION == 1 and provide a
     * arch_key_modification(INT8 *, int) function
     */
    arch_key_modification(hc->nh_key, sizeof(hc->nh_key));
    #endif
    nh_reset(hc);
}

/* ---------------------------------------------------------------------- */

static void nh_update(nh_ctx *hc, INT8 *buf, UINT32 nbytes)
/* Incorporate nbytes of data into a nh_ctx, buffer whatever is not an    */
/* even multiple of HASH_BUF_BYTES.                                       */
{
    UINT32 i,j;
    
    j = hc->next_data_empty;
    if ((j + nbytes) >= HASH_BUF_BYTES) {
        if (j) {
            i = HASH_BUF_BYTES - j;
            memcpy(hc->data+j, buf, i);
            nh_transform(hc,hc->data,HASH_BUF_BYTES);
            nbytes -= i;
            buf += i;
            hc->bytes_hashed += HASH_BUF_BYTES;
        }
        if (nbytes >= HASH_BUF_BYTES) {
            /* i = nbytes - (nbytes % HASH_BUF_BYTES); */
            i = nbytes & ~(HASH_BUF_BYTES - 1);
            nh_transform(hc, buf, i);
            nbytes -= i;
            buf += i;
            hc->bytes_hashed += i;
        }
        j = 0;
    }
    memcpy(hc->data + j, buf, nbytes);
    hc->next_data_empty = j + nbytes;
}

/* ---------------------------------------------------------------------- */

static void zero_pad(INT8 *p, int nbytes)
{
/* Write "nbytes" of zeroes, beginning at "p" */
    if (nbytes >= (int)sizeof(UWORD)) {
        while ((POINTER_INT)p % sizeof(UWORD)) {
            *p = 0;
            nbytes--;
            p++;
        }
        while (nbytes >= (int)sizeof(UWORD)) {
            *(UWORD *)p = 0;
            nbytes -= sizeof(UWORD);
            p += sizeof(UWORD);
        }
    }
    while (nbytes) {
        *p = 0;
        nbytes--;
        p++;
    }
}

/* ---------------------------------------------------------------------- */

static void nh_final(nh_ctx *hc, INT8 *result)
/* After passing some number of data buffers to nh_update() for integration
 * into an NH context, nh_final is called to produce a hash result. If any
 * bytes are in the buffer hc->data, incorporate them into the
 * NH context. Finally, add into the NH accumulation "state" the total number
 * of bits hashed. The resulting numbers are written to the buffer "result".
 */
{
    int nh_len, nbits;

    if (hc->next_data_empty) {
        nh_len = ((hc->next_data_empty + 
                  (L1_PAD_BOUNDARY - 1)) & ~(L1_PAD_BOUNDARY - 1));
        zero_pad(hc->data + hc->next_data_empty, 
                 nh_len - hc->next_data_empty);
        nh_transform(hc, hc->data, nh_len);
        hc->bytes_hashed += hc->next_data_empty;
    }
    nbits = (hc->bytes_hashed << 3);
    ((LARGE_UWORD *)result)[0] = ((LARGE_UWORD *)hc->state)[0] + nbits;
    #if (PREFIX_STREAMS > 1)
    ((LARGE_UWORD *)result)[1] = ((LARGE_UWORD *)hc->state)[1] + nbits;
    #if (PREFIX_STREAMS > 2)
    ((LARGE_UWORD *)result)[2] = ((LARGE_UWORD *)hc->state)[2] + nbits;
    ((LARGE_UWORD *)result)[3] = ((LARGE_UWORD *)hc->state)[3] + nbits;
    #if (PREFIX_STREAMS > 4)
    ((LARGE_UWORD *)result)[4] = ((LARGE_UWORD *)hc->state)[4] + nbits;
    ((LARGE_UWORD *)result)[5] = ((LARGE_UWORD *)hc->state)[5] + nbits;
    #if (PREFIX_STREAMS > 6)
    ((LARGE_UWORD *)result)[6] = ((LARGE_UWORD *)hc->state)[6] + nbits;
    ((LARGE_UWORD *)result)[7] = ((LARGE_UWORD *)hc->state)[7] + nbits;
    #endif
    #endif
    #endif
    #endif
    nh_reset(hc);
}

/* ---------------------------------------------------------------------- */

static void nh(nh_ctx *hc, INT8 *buf, UINT32 padded_len,
               UINT32 unpadded_len, INT8 *result)
/* All-in-one nh_update() and nh_final() equivalent.
 * Assumes that padded_len is divisible by L1_PAD_BOUNDARY and result is
 * well aligned
 */
{
    UINT32 nbits;
    
    /* Initialize the hash state */
    nbits = (unpadded_len << 3);
    
	((LARGE_UWORD *)result)[0] = nbits;
    #if (PREFIX_STREAMS > 1)
    ((LARGE_UWORD *)result)[1] = nbits;
    #if (PREFIX_STREAMS > 2)
    ((LARGE_UWORD *)result)[2] = nbits;
    ((LARGE_UWORD *)result)[3] = nbits;
    #if (PREFIX_STREAMS > 4)
    ((LARGE_UWORD *)result)[4] = nbits;
    ((LARGE_UWORD *)result)[5] = nbits;
    #if (PREFIX_STREAMS > 6)
    ((LARGE_UWORD *)result)[6] = nbits;
    ((LARGE_UWORD *)result)[7] = nbits;
    #endif
    #endif
    #endif
    #endif
    
	nh_aux((INT8 *)hc->nh_key, (INT8 *)buf, result, padded_len);
}

/* ---------------------------------------------------------------------- */
/* ---------------------------------------------------------------------- */
/* ----- Begin UHASH Section -------------------------------------------- */
/* ---------------------------------------------------------------------- */
/* ---------------------------------------------------------------------- */

/* UHASH is a multi-layered algorithm. Data presented to UHASH is first
 * hashed by NH. The NH output is then hashed by a polynomial-hash layer
 * unless the initial data to be hashed is short. After the polynomial-
 * layer, an inner-product hash is used to produce the final UHASH output.
 *
 * UHASH provides two interfaces, one all-at-once and another where data
 * buffers are presented sequentially. In the sequential interface, the
 * UHASH client calls the routine uhash_update() as many times as necessary.
 * When there is no more data to be fed to UHASH, the client calls
 * uhash_final() which          
 * calculates the UHASH output. Before beginning another UHASH calculation    
 * the uhash_reset() routine must be called. The all-at-once UHASH routine,   
 * uhash(), is equivalent to the sequence of calls uhash_update() and         
 * uhash_final(); however it is optimized and should be                     
 * used whenever the sequential interface is not necessary.              
 *                                                                        
 * The routine uhash_init() initializes the uhash_ctx data structure and    
 * must be called once, before any other UHASH routine.
 */                                                        

/* ---------------------------------------------------------------------- */
/* ----- PDF Constants and uhash_ctx ------------------------------------ */
/* ---------------------------------------------------------------------- */



/* ---------------------------------------------------------------------- */

typedef struct uhash_ctx {
    nh_ctx hash;                        /* Hash context for L1 NH hash    */
    /* Extra stuff for the WORD_LEN == 2 case, where a polyhash tansition
     * may occur between p32 and p64
     */
    #if (WORD_LEN == 2)
    UINT32 poly_key_4[PREFIX_STREAMS]; /* p32 Poly keys                   */
    UINT64 poly_store[PREFIX_STREAMS]; /* To buffer NH-16 output for p64  */
    int poly_store_full;               /* Flag for poly_store             */
    UINT32 poly_invocations;           /* Number of p32 words hashed      */
    #endif
    UINT64 poly_key_8[PREFIX_STREAMS];    /* p64 poly keys                */
    UINT64 poly_accum[PREFIX_STREAMS];    /* poly hash result             */
    LARGE_UWORD ip_keys[PREFIX_STREAMS*4];/* Inner-product keys           */
    SMALL_UWORD ip_trans[PREFIX_STREAMS]; /* Inner-product translation    */
    UINT32 msg_len;               /* Total length of data passed to uhash */
} uhash_ctx;

/* ---------------------------------------------------------------------- */


/* ---------------------------------------------------------------------- */
#if (USE_C_ONLY || ! ARCH_POLY)
/* ---------------------------------------------------------------------- */

/* The polynomial hashes use Horner's rule to evaluate a polynomial one
 * word at a time. As described in the specification, poly32 and poly64
 * require keys from special domains. The following impelementations exploit
 * the special domains to avoid overflow. The results are not guaranteed to
 * be within Z_p32 and Z_p64, but the Inner-Product hash implementation
 * patches any errant values.
 */
static UINT32 poly32(UINT32 cur, UINT32 key, UINT32 data)
/* requires 29 bit keys */
{
    UINT64 t;
    UINT32 hi, lo;
    
    t = cur * (UINT64)key;
    hi = (UINT32)(t >> 32);
    lo = (UINT32)t;
    hi *= 5;
    lo += hi;
    if (lo < hi)
        lo += 5;
    lo += data;
    if (lo < data)
        lo += 5;
    return lo;
}

static UINT64 poly64(UINT64 cur, UINT64 key, UINT64 data)
{
    UINT32 key_hi = (UINT32)(key >> 32),
           key_lo = (UINT32)key,
           cur_hi = (UINT32)(cur >> 32),
           cur_lo = (UINT32)cur,
           x_lo,
           x_hi;
    UINT64 X,T,res;
    
    X =  MUL64(key_hi, cur_lo) + MUL64(cur_hi, key_lo);
    x_lo = (UINT32)X;
    x_hi = (UINT32)(X >> 32);
    
    res = (MUL64(key_hi, cur_hi) + x_hi) * 59 + MUL64(key_lo, cur_lo);
     
    T = ((UINT64)x_lo << 32);
    res += T;
    if (res < T)
        res += 59;

    res += data;
    if (res < data)
        res += 59;

    return res;
}

#endif



#if (WORD_LEN == 2)

/* Although UMAC is specified to use a ramped polynomial hash scheme, this
 * impelemtation does not handle all ramp levels. When WORD_LEN is 2, we only
 * handle the p32 and p64 modulus polynomial calculations. Because we don't
 * handle the ramp up to p128 modulus in this implementation, we are limited
 * to 2^31 poly_hash() invocations per stream (for a total capacity of 2^41
 * bytes per tag input to UMAC).
 */
const UINT32 poly_crossover = (1ul << 9);

static void poly_hash(uhash_ctx_t hc, UINT32 data[])
{
    int i;

    if (hc->poly_invocations < poly_crossover) { /* Use poly32 */
        for (i = 0; i < PREFIX_STREAMS; i++) {
            /* If the data passed in is out of range, we hash a marker
             * and then hash the data offset to be in range.
             */
            if (data[i] >= (p32-1)) {
                hc->poly_accum[i] = poly32((UINT32)hc->poly_accum[i],
                                           hc->poly_key_4[i], p32-1);
                hc->poly_accum[i] = poly32((UINT32)hc->poly_accum[i],
                                           hc->poly_key_4[i], (data[i] - 5));
            } else
                hc->poly_accum[i] = poly32((UINT32)hc->poly_accum[i], 
                                           hc->poly_key_4[i], data[i]);
        }
    } else if (hc->poly_invocations > poly_crossover) {      /* Use poly64 */
        /* We must buffer every other 32-bit word to build up a 64-bit one */
        if ( ! hc->poly_store_full) {
            for (i = 0; i < PREFIX_STREAMS; i++) {
                hc->poly_store[i] = ((UINT64)data[i]) << 32;
            }
            hc->poly_store_full = 1;
        } else {
            for (i = 0; i < PREFIX_STREAMS; i++) {
                /* If the data passed in is out of range, we hash a marker
                 * and then hash the data offset to be in range.
                 */
                if ((UINT32)(hc->poly_store[i] >> 32) == 0xfffffffful) {
                    hc->poly_accum[i] = poly64(hc->poly_accum[i], 
                                               hc->poly_key_8[i], p64 - 1);
                    hc->poly_accum[i] = poly64(hc->poly_accum[i], 
                                        hc->poly_key_8[i],
                                        (hc->poly_store[i] + data[i] - 59));
                } else {
                    hc->poly_accum[i] = poly64(hc->poly_accum[i],
                                        hc->poly_key_8[i],
                                        hc->poly_store[i] + data[i]);
                }
                hc->poly_store_full = 0;
            }
        }
    } else { /* (hc->poly_invocations == poly_crossover) */
        /* Implement the ramp from p32 to p64 hashing    */
        for (i = 0; i < PREFIX_STREAMS; i++) {
            hc->poly_accum[i] = poly64(1, hc->poly_key_8[i],
                                       hc->poly_accum[i]);
            hc->poly_store[i] = ((UINT64)data[i]) << 32;
        }
        hc->poly_store_full = 1;
    }
    hc->poly_invocations += 1;
}

#else /* WORD_LEN == 4 */

/* Although UMAC is specified to use a ramped polynomial hash scheme, this
 * impelemtation does not handle all ramp levels. When WORD_LEN is 4, we only
 * handle the p64 modulus polynomial calculations. Because we don't handle
 * the ramp up to p128 modulus in this implementation, we are limited to
 * 2^14 poly_hash() invocations per stream (for a total capacity of 2^24
 * bytes per tag input to UMAC).
 */
static void poly_hash(uhash_ctx_t hc, UINT32 data_in[])
{
/* This routine is simpler than that above because there is no ramping. */ 
    int i;
    UINT64 *data=(UINT64*)data_in;
    
    for (i = 0; i < PREFIX_STREAMS; i++) {
        if ((UINT32)(data[i] >> 32) == 0xfffffffful) {
            hc->poly_accum[i] = poly64(hc->poly_accum[i], 
                                       hc->poly_key_8[i], p64 - 1);
            hc->poly_accum[i] = poly64(hc->poly_accum[i],
                                       hc->poly_key_8[i], (data[i] - 59));
        } else {
            hc->poly_accum[i] = poly64(hc->poly_accum[i],
                                       hc->poly_key_8[i], data[i]);
        }
    }
}

#endif

/* ---------------------------------------------------------------------- */

#if (WORD_LEN == 4)

/* ---------------------------------------------------------------------- */
#if (USE_C_ONLY || ! ARCH_IP)
/* ---------------------------------------------------------------------- */

/* The final step in UHASH is an inner-product hash. The poly hash
 * produces a result not neccesarily WORD_LEN bytes long. The inner-
 * product hash breaks the polyhash output into 16-bit chunks and
 * multiplies each with a 36 bit key.
 */
static UINT64 ip_aux(UINT64 t, UINT64 *ipkp, UINT64 data)
{
    t = t + ipkp[0] * (UINT64)(UINT16)(data >> 48);
    t = t + ipkp[1] * (UINT64)(UINT16)(data >> 32);
    t = t + ipkp[2] * (UINT64)(UINT16)(data >> 16);
    t = t + ipkp[3] * (UINT64)(UINT16)(data);
    
    return t;
}

static SMALL_UWORD ip_reduce_p36(LARGE_UWORD t)
{
/* Divisionless modular reduction */
    UINT64 ret;
    
    ret = (t & m36) + 5 * (t >> 36);
    if (ret >= p36)
        ret -= p36;

    /* return least significant 32 bits */
    return (SMALL_UWORD)(ret);
}

#endif

/* If the data being hashed by UHASH is no longer than L1_KEY_LEN, then
 * the polyhash stage is skipped and ip_short is applied directly to the
 * NH output.
 */
static void ip_short(uhash_ctx_t ahc, INT8 *nh_res, char *res)
{
    UINT64 t;
    UINT64 *nhp = (UINT64 *)nh_res;
    
    t  = ip_aux(0,ahc->ip_keys, nhp[0]);
    STORE_UINT32_BIG((UINT32 *)res+0, ip_reduce_p36(t) ^ ahc->ip_trans[0]);
    #if (PREFIX_STREAMS > 1)
    t  = ip_aux(0,ahc->ip_keys+1, nhp[1]);
    STORE_UINT32_BIG((UINT32 *)res+1, ip_reduce_p36(t) ^ ahc->ip_trans[1]);
    #if (PREFIX_STREAMS > 2)
    t  = ip_aux(0,ahc->ip_keys+2, nhp[2]);
    STORE_UINT32_BIG((UINT32 *)res+2, ip_reduce_p36(t) ^ ahc->ip_trans[2]);
    t  = ip_aux(0,ahc->ip_keys+3, nhp[3]);
    STORE_UINT32_BIG((UINT32 *)res+3, ip_reduce_p36(t) ^ ahc->ip_trans[3]);
    #if (PREFIX_STREAMS > 4)
    t  = ip_aux(0,ahc->ip_keys+4, nhp[4]);
    STORE_UINT32_BIG((UINT32 *)res+4, ip_reduce_p36(t) ^ ahc->ip_trans[4]);
    t  = ip_aux(0,ahc->ip_keys+5, nhp[5]);
    STORE_UINT32_BIG((UINT32 *)res+5, ip_reduce_p36(t) ^ ahc->ip_trans[5]);
    #if (PREFIX_STREAMS > 6)
    t  = ip_aux(0,ahc->ip_keys+6, nhp[6]);
    STORE_UINT32_BIG((UINT32 *)res+6, ip_reduce_p36(t) ^ ahc->ip_trans[6]);
    t  = ip_aux(0,ahc->ip_keys+7, nhp[7]);
    STORE_UINT32_BIG((UINT32 *)res+7, ip_reduce_p36(t) ^ ahc->ip_trans[7]);
    #endif
    #endif
    #endif
    #endif
}

/* If the data being hashed by UHASH is longer than L1_KEY_LEN, then
 * the polyhash stage is not skipped and ip_long is applied to the
 * polyhash output.
 */
static void ip_long(uhash_ctx_t ahc, char *res)
{
    int i;
    UINT64 t;

    for (i = 0; i < PREFIX_STREAMS; i++) {
        /* fix polyhash output not in Z_p64 */
        if (ahc->poly_accum[i] >= p64)
            ahc->poly_accum[i] -= p64;
        t  = ip_aux(0,ahc->ip_keys+i, ahc->poly_accum[i]);
        STORE_UINT32_BIG((UINT32 *)res+i, 
                         ip_reduce_p36(t) ^ ahc->ip_trans[i]);
    }
}


/* ---------------------------------------------------------------------- */
#elif (WORD_LEN == 2)
/* ---------------------------------------------------------------------- */

/* ---------------------------------------------------------------------- */
#if (USE_C_ONLY || ! ARCH_IP)
/* ---------------------------------------------------------------------- */

/* The final step in UHASH is an inner-product hash. The poly hash
 * produces a result not neccesarily WORD_LEN bytes long. The inner-
 * product hash breaks the polyhash output into 16-bit chunks and
 * multiplies each with a 19 bit key.
 */
static UINT64 ip_aux(UINT64 t, UINT32 *ipkp, UINT32 data)
{
    t = t + MUL64(ipkp[0], (data >> 16));
    t = t + MUL64(ipkp[1], (UINT16)(data));
    
    return t;
}

static UINT16 ip_reduce_p19(UINT64 t)
{
/* Divisionless modular reduction */
    UINT32 ret;
    
    ret = ((UINT32)t & m19) + (UINT32)(t >> 19);
    if (ret >= p19)
        ret -= p19;

    /* return least significant 16 bits */
    return (UINT16)(ret);
}


#endif


/* If the data being hashed by UHASH is no longer than L1_KEY_LEN, then
 * the polyhash stage is skipped and ip_short is applied directly to the
 * NH output.
 */
static void ip_short(uhash_ctx_t ahc, INT8 *nh_res, char *res)
{
    UINT64 t;
    UINT32 *nhp = (UINT32 *)nh_res;
    
    t  = ip_aux(0,ahc->ip_keys+2, nhp[0]);
    STORE_UINT16_BIG((UINT16 *)res+0, (UINT16)(ip_reduce_p19(t) ^ ahc->ip_trans[0]));
    #if (PREFIX_STREAMS > 1)
    t  = ip_aux(0,ahc->ip_keys+6, nhp[1]);
    STORE_UINT16_BIG((UINT16 *)res+1, (UINT16)(ip_reduce_p19(t) ^ ahc->ip_trans[1]));
    #if (PREFIX_STREAMS > 2)
    t  = ip_aux(0,ahc->ip_keys+10, nhp[2]);
    STORE_UINT16_BIG((UINT16 *)res+2, (UINT16)(ip_reduce_p19(t) ^ ahc->ip_trans[2]));
    t  = ip_aux(0,ahc->ip_keys+14, nhp[3]);
    STORE_UINT16_BIG((UINT16 *)res+3, (UINT16)(ip_reduce_p19(t) ^ ahc->ip_trans[3]));
    #if (PREFIX_STREAMS > 4)
    t  = ip_aux(0,ahc->ip_keys+18, nhp[4]);
    STORE_UINT16_BIG((UINT16 *)res+4, (UINT16)(ip_reduce_p19(t) ^ ahc->ip_trans[4]));
    t  = ip_aux(0,ahc->ip_keys+22, nhp[5]);
    STORE_UINT16_BIG((UINT16 *)res+5, (UINT16)(ip_reduce_p19(t) ^ ahc->ip_trans[5]));
    #if (PREFIX_STREAMS > 6)
    t  = ip_aux(0,ahc->ip_keys+26, nhp[6]);
    STORE_UINT16_BIG((UINT16 *)res+6, (UINT16)(ip_reduce_p19(t) ^ ahc->ip_trans[6]));
    t  = ip_aux(0,ahc->ip_keys+30, nhp[7]);
    STORE_UINT16_BIG((UINT16 *)res+7, (UINT16)(ip_reduce_p19(t) ^ ahc->ip_trans[7]));
    #endif
    #endif
    #endif
    #endif
}

/* If the data being hashed by UHASH is longer than L1_KEY_LEN, then
 * the polyhash stage is not skipped and ip_long is applied to the
 * polyhash output.
 */
static void ip_long(uhash_ctx_t ahc, char *res)
{
    int i;
    UINT64 t;

    if (ahc->poly_invocations > poly_crossover) { /* hash 64 bits */
        for (i = 0; i < PREFIX_STREAMS; i++) {
            if (ahc->poly_accum[i] >= p64)
                ahc->poly_accum[i] -= p64;
            t = ip_aux(0,ahc->ip_keys+i*4,(UINT32)(ahc->poly_accum[i] >> 32));
            t = ip_aux(t,ahc->ip_keys+i*4+2,(UINT32)ahc->poly_accum[i]);
            /* Store result big endian for sonsistancy across architectures */
            STORE_UINT16_BIG((UINT16 *)res+i, 
                             (UINT16)(ip_reduce_p19(t) ^ ahc->ip_trans[i]));
        }
    } else {                                      /* hash 32 bits */
        for (i = 0; i < PREFIX_STREAMS; i++) {
            if (ahc->poly_accum[i] >= p32)
                ahc->poly_accum[i] -= p32;
            t  = ip_aux(0,ahc->ip_keys+i*4+2,(UINT32)ahc->poly_accum[i]);
            STORE_UINT16_BIG((UINT16 *)res+i, 
                             (UINT16)(ip_reduce_p19(t) ^ ahc->ip_trans[i]));
        }
    }
}

#endif

/* ---------------------------------------------------------------------- */

/* Reset uhash context for next hash session */
int uhash_reset(uhash_ctx_t pc)
{
    nh_reset(&pc->hash);
    pc->msg_len = 0;
    #if (WORD_LEN == 2)
    pc->poly_invocations = 0;
    pc->poly_store_full = 0;
    #endif
    pc->poly_accum[0] = 1;
    #if (PREFIX_STREAMS > 1)
    pc->poly_accum[1] = 1;
    #endif
    #if (PREFIX_STREAMS > 2)
    pc->poly_accum[2] = 1;
    pc->poly_accum[3] = 1;
    #endif
    #if (PREFIX_STREAMS > 4)
    pc->poly_accum[4] = 1;
    pc->poly_accum[5] = 1;
    #endif
    #if (PREFIX_STREAMS > 6)
    pc->poly_accum[6] = 1;
    pc->poly_accum[7] = 1;
    #endif
    return 1;
}

/* ---------------------------------------------------------------------- */

/* Given a pointer to the internal key needed by kdf() and a uhash context,
 * initialize the NH context and generate keys needed for poly and inner-
 * product hashing. All keys are endian adjusted in memory so that native
 * loads cause correct keys to be in registers during calculation.
 */
static void uhash_init(uhash_ctx_t ahc, INT8 *prf_key)
{
    int i;
    INT8 buf[256];
    
    /* Zero the entire uhash context */
    memset(ahc, 0, sizeof(uhash_ctx));

    /* Initialize the L1 hash */
    nh_init(&ahc->hash, prf_key);
    
    /* Setup L2 hash variables */
    kdf(buf, prf_key, 1, sizeof(buf));    /* Fill buffer with index 1 key */
    for (i = 0; i < PREFIX_STREAMS; i++) {
        /* Fill keys from the buffer, skipping bytes in the buffer not
         * used by this implementation. Endian reverse the keys if on a
         * little-endian computer.
         */
        #if (WORD_LEN == 2)
        memcpy(ahc->poly_key_4+i, buf+28*i, 4);
        memcpy(ahc->poly_key_8+i, buf+28*i+4, 8);
        endian_convert_if_le(ahc->poly_key_4+i, 4, 4);
        ahc->poly_key_4[i] &= 0x1fffffff;  /* Mask to special domain */
        #elif (WORD_LEN == 4)
        memcpy(ahc->poly_key_8+i, buf+24*i, 8);
        #endif
        endian_convert_if_le(ahc->poly_key_8+i, 8, 8);
        /* Mask the 64-bit keys to their special domain */
        ahc->poly_key_8[i] &= ((UINT64)0x01ffffffu << 32) + 0x01ffffffu;
        ahc->poly_accum[i] = 1;  /* Our polyhash prepends a non-zero word */
    }
    
    /* Setup L3-1 hash variables */
    kdf(buf, prf_key, 2, sizeof(buf)); /* Fill buffer with index 2 key */
    for (i = 0; i < PREFIX_STREAMS; i++)
          memcpy(ahc->ip_keys+4*i, buf+8*sizeof(LARGE_WORD)*i,
                 4*sizeof(LARGE_WORD));
    endian_convert_if_le(ahc->ip_keys, sizeof(LARGE_WORD), 
                         sizeof(ahc->ip_keys));
    for (i = 0; i < PREFIX_STREAMS*4; i++)
        #if (WORD_LEN == 2)
        ahc->ip_keys[i] %= p19;  /* Bring into Z_p19 */
        #elif (WORD_LEN == 4)
        ahc->ip_keys[i] %= p36;  /* Bring into Z_p36 */
        #endif
    
    /* Setup L3-2 hash variables    */
    /* Fill buffer with index 3 key */
    kdf(ahc->ip_trans, prf_key, 3, PREFIX_STREAMS * sizeof(SMALL_UWORD));
    endian_convert_if_le(ahc->ip_trans, sizeof(SMALL_UWORD),
                         PREFIX_STREAMS * sizeof(SMALL_UWORD));
}

/* ---------------------------------------------------------------------- */

uhash_ctx_t uhash_alloc(char key[])
{
/* Allocate memory and force to a 16-byte boundary. */
    uhash_ctx_t ctx;
    char bytes_to_add;
    UINT32 prf_key[RC6_TABLE_WORDS];
    
    ctx = (uhash_ctx_t)malloc(sizeof(uhash_ctx)+ALLOC_BOUNDARY);
    if (ctx) {
        if (ALLOC_BOUNDARY) {
            bytes_to_add = ALLOC_BOUNDARY - ((POINTER_INT)ctx & (ALLOC_BOUNDARY -1));
            ctx = (uhash_ctx_t)((char *)ctx + bytes_to_add);
            *((char *)ctx - 1) = bytes_to_add;
        }
        RC6_SETUP(key, prf_key);  /* Intitialize the block-cipher */
        uhash_init(ctx, (INT8 *)prf_key);
    }
    return (ctx);
}

/* ---------------------------------------------------------------------- */

int uhash_free(uhash_ctx_t ctx)
{
/* Free memory allocated by uhash_alloc */
    char bytes_to_sub;
    
    if (ctx) {
        if (ALLOC_BOUNDARY) {
            bytes_to_sub = *((char *)ctx - 1);
            ctx = (uhash_ctx_t)((char *)ctx - bytes_to_sub);
        }
        free(ctx);
    }
    return (1);
}

/* ---------------------------------------------------------------------- */

int uhash_update(uhash_ctx_t ctx, char *input, long len)
/* Given len bytes of data, we parse it into L1_KEY_LEN chunks and
 * hash each one with NH, calling the polyhash on each NH output.
 */
{
    UWORD bytes_hashed, bytes_remaining;
    INT8 nh_result[PREFIX_STREAMS*sizeof(LARGE_WORD)];
    
    if (ctx->msg_len + len <= L1_KEY_LEN) {
        nh_update(&ctx->hash, (INT8 *)input, len);
        ctx->msg_len += len;
    } else {
    
         bytes_hashed = ctx->msg_len % L1_KEY_LEN;
         if (ctx->msg_len == L1_KEY_LEN)
             bytes_hashed = L1_KEY_LEN;

         if (bytes_hashed + len >= L1_KEY_LEN) {

             /* If some bytes have been passed to the hash function      */
             /* then we want to pass at most (L1_KEY_LEN - bytes_hashed) */
             /* bytes to complete the current nh_block.                  */
             if (bytes_hashed) {
                 bytes_remaining = (L1_KEY_LEN - bytes_hashed);
                 nh_update(&ctx->hash, input, bytes_remaining);
                 nh_final(&ctx->hash, nh_result);
                 ctx->msg_len += bytes_remaining;
                 poly_hash(ctx,(UINT32 *)nh_result);
                 len -= bytes_remaining;
                 input += bytes_remaining;
             }

             /* Hash directly from input stream if enough bytes */
             while (len >= L1_KEY_LEN) {
                 nh(&ctx->hash, (INT8 *)input, L1_KEY_LEN,
                                   L1_KEY_LEN, nh_result);
                 ctx->msg_len += L1_KEY_LEN;
                 len -= L1_KEY_LEN;
                 input += L1_KEY_LEN;
                 poly_hash(ctx,(UINT32 *)nh_result);
             }
         }

         /* pass remaining < L1_KEY_LEN bytes of input data to NH */
         if (len) {
             nh_update(&ctx->hash, (INT8 *)input, len);
             ctx->msg_len += len;
         }
     }

    return (1);
}

/* ---------------------------------------------------------------------- */

int uhash_final(uhash_ctx_t ctx, char *res)
/* Incorporate any pending data, pad, and generate tag */
{
    INT8 nh_result[PREFIX_STREAMS*sizeof(LARGE_WORD)];

    if (ctx->msg_len > L1_KEY_LEN) {
        if (ctx->msg_len % L1_KEY_LEN) {
            nh_final(&ctx->hash, nh_result);
            poly_hash(ctx,(UINT32 *)nh_result);
        }
        /* If WORD_LEN == 2 and we have ramped-up to p64 in the polyhash,
         * then we must pad the data passed to poly64 with a 1 bit and then
         * zero bits up to the next multiple of 64 bits.
         */
        #if (WORD_LEN == 2)
        if (ctx->poly_invocations > poly_crossover) {
            UINT32 tmp[PREFIX_STREAMS];
            int i;
            for (i = 0; i < PREFIX_STREAMS; i++)
                tmp[i] = 0x80000000u;
            poly_hash(ctx,tmp);
            if (ctx->poly_store_full) {
                for (i = 0; i < PREFIX_STREAMS; i++)
                    tmp[i] = 0;
                poly_hash(ctx,tmp);
            }
        }
        #endif
        ip_long(ctx, res);
    } else {
        nh_final(&ctx->hash, nh_result);
        ip_short(ctx,nh_result, res);
    }
    uhash_reset(ctx);
    return (1);
}

/* ---------------------------------------------------------------------- */

int uhash(uhash_ctx_t ahc, char *msg, long len, char *res)
/* assumes that msg is in a writable buffer of length divisible by */
/* L1_PAD_BOUNDARY. Bytes beyond msg[len] may be zeroed.           */
/* Does not handle zero length message                             */
{
    INT8 nh_result[PREFIX_STREAMS*sizeof(LARGE_WORD)];
    UINT32 nh_len;
    int extra_zeroes_needed;
    INT8 padded_msg[L1_KEY_LEN/sizeof(INT8)];
        
    /* If the message to be hashed is no longer than L1_HASH_LEN, we skip
     * the polyhash.
     */
    if (len <= L1_KEY_LEN) {
        nh_len = ((len + (L1_PAD_BOUNDARY - 1)) & ~(L1_PAD_BOUNDARY - 1));
        extra_zeroes_needed = nh_len - len;
	if (extra_zeroes_needed) {  
	  memcpy(padded_msg, msg, len);
	  msg = padded_msg;
	  zero_pad((INT8 *)msg + len, extra_zeroes_needed);
	}
        nh(&ahc->hash, msg, nh_len, len, nh_result);
        ip_short(ahc,nh_result, res);
    } else {
        /* Otherwise, we hash each L1_KEY_LEN chunk with NH, passing the NH
         * output to poly_hash().
         */
        do {
            nh(&ahc->hash, msg, L1_KEY_LEN, L1_KEY_LEN, nh_result);
            poly_hash(ahc,(UINT32 *)nh_result);
            len -= L1_KEY_LEN;
            msg += L1_KEY_LEN;
        } while (len >= L1_KEY_LEN);
        if (len) {
            nh_len = ((len + (L1_PAD_BOUNDARY - 1)) & ~(L1_PAD_BOUNDARY - 1));
            extra_zeroes_needed = nh_len - len;
	    if (extra_zeroes_needed) {  
	      memcpy(padded_msg, msg, len);
	      msg = padded_msg;
	      zero_pad((INT8 *)msg + len, extra_zeroes_needed);
	    }
            nh(&ahc->hash, msg, nh_len, len, nh_result);
            poly_hash(ahc,(UINT32 *)nh_result);
        }
        /* If WORD_LEN == 2 and we have ramped-up to p64 in the polyhash,
         * then we must pad the data passed to poly64 with a 1 bit and then
         * zero bits up to the next multiple of 64 bits.
         */
        #if (WORD_LEN == 2)
        if (ahc->poly_invocations > poly_crossover) {
            UINT32 tmp[PREFIX_STREAMS];
            int i;
            for (i = 0; i < PREFIX_STREAMS; i++)
                tmp[i] = 0x80000000u;
            poly_hash(ahc,tmp);
            if (ahc->poly_store_full) {
                for (i = 0; i < PREFIX_STREAMS; i++)
                    tmp[i] = 0;
                poly_hash(ahc,tmp);
            }
        }
        #endif
        ip_long(ahc, res);
    }
    
    uhash_reset(ahc);
    return 1;
}

/* ---------------------------------------------------------------------- */
/* ---------------------------------------------------------------------- */
/* ----- Begin UMAC Section --------------------------------------------- */
/* ---------------------------------------------------------------------- */
/* ---------------------------------------------------------------------- */

/* The UMAC interface has two interfaces, an all-at-once interface where
 * the entire message to be authenticated is passed to UMAC in one buffer,
 * and a sequential interface where the message is presented a bit at a time.   
 * The all-at-once is more optimaized than the sequential version and should
 * be preferred when the sequential interface is not required. 
 */
typedef struct umac_ctx {
    uhash_ctx hash;          /* Hash function for message compression    */
    pdf_ctx pdf;             /* PDF for hashed output                    */
} umac_ctx;

/* ---------------------------------------------------------------------- */

int umac_reset(umac_ctx_t ctx)
/* Reset the hash function to begin a new authentication.        */
{
    uhash_reset(&ctx->hash);
    return (1);
}

/* ---------------------------------------------------------------------- */

int umac_delete(umac_ctx_t ctx)
/* Deallocate the ctx structure */
{
    char bytes_to_sub;
    
    if (ctx) {
        if (ALLOC_BOUNDARY) {
            bytes_to_sub = *((char *)ctx - 1);
            ctx = (umac_ctx_t)((char *)ctx - bytes_to_sub);
        }
        free(ctx);
    }
    return (1);
}

/* ---------------------------------------------------------------------- */

umac_ctx_t umac_new(char key[])
/* Dynamically allocate a umac_ctx struct, initialize variables, 
 * generate subkeys from key. Align to 16-byte boundary.
 */
{
    umac_ctx_t ctx;
    char bytes_to_add;
    UINT32 prf_key[RC6_TABLE_WORDS];
    
    ctx = (umac_ctx_t)malloc(sizeof(umac_ctx)+ALLOC_BOUNDARY);
    if (ctx) {
        if (ALLOC_BOUNDARY) {
            bytes_to_add = ALLOC_BOUNDARY - ((POINTER_INT)ctx & (ALLOC_BOUNDARY - 1));
            ctx = (umac_ctx_t)((char *)ctx + bytes_to_add);
            *((char *)ctx - 1) = bytes_to_add;
        }
        RC6_SETUP(key, prf_key);
        pdf_init(&ctx->pdf, (INT8 *)prf_key);
        uhash_init(&ctx->hash, (INT8 *)prf_key);
    }
        
    return (ctx);
}

/* ---------------------------------------------------------------------- */

int umac_final(umac_ctx_t ctx, char tag[], char nonce[8])
/* Incorporate any pending data, pad, and generate tag */
{
    /* pdf_gen_xor writes OUTPUT_STREAMS * WORD_LEN bytes to its output
     * buffer, so if PREFIX_STREAMS == OUTPUT_STREAMS, we write directly
     * to the buffer supplied by the client. Otherwise we use a temporary
     * buffer.
     */
    #if ((PREFIX_STREAMS == OUTPUT_STREAMS) || HASH_ONLY)
    INT8 *uhash_result = (INT8 *)tag;
    #else
    INT8 uhash_result[UMAC_OUTPUT_LEN];
    #endif
    
    uhash_final(&ctx->hash, uhash_result);
    #if ( ! HASH_ONLY)
    pdf_gen_xor(&ctx->pdf, nonce, uhash_result);
    #endif
    
    #if ((PREFIX_STREAMS != OUTPUT_STREAMS) && ! HASH_ONLY)
    memcpy(tag,uhash_result,UMAC_PREFIX_LEN);
    #endif
    
    return (1);
}

/* ---------------------------------------------------------------------- */

int umac_update(umac_ctx_t ctx, char *input, long len)
/* Given len bytes of data, we parse it into L1_KEY_LEN chunks and   */
/* hash each one, calling the PDF on the hashed output whenever the hash- */
/* output buffer is full.                                                 */
{
    uhash_update(&ctx->hash, input, len);
    return (1);
}

/* ---------------------------------------------------------------------- */

int umac(umac_ctx_t ctx, char *input, 
         long len, char tag[],
         char nonce[8])
/* All-in-one version simply calls umac_update() and umac_final().        */
{
    #if ((PREFIX_STREAMS == OUTPUT_STREAMS) || HASH_ONLY)
    INT8 *uhash_result = (INT8 *)tag;
    #else
    INT8 uhash_result[UMAC_OUTPUT_LEN];
    #endif
    
    uhash(&ctx->hash, input, len, uhash_result);
    #if ( ! HASH_ONLY)
    pdf_gen_xor(&ctx->pdf, nonce, uhash_result);
    #endif
    
    #if ((PREFIX_STREAMS != OUTPUT_STREAMS) && ! HASH_ONLY)
    memcpy(tag,uhash_result,UMAC_PREFIX_LEN);
    #endif

    return (1);
}

/* ---------------------------------------------------------------------- */
/* ---------------------------------------------------------------------- */
/* ----- End UMAC Section ----------------------------------------------- */
/* ---------------------------------------------------------------------- */
/* ---------------------------------------------------------------------- */

/* If RUN_TESTS is defined non-zero, then we define a main() function and */
/* run some verification and speed tests.                                 */

#if RUN_TESTS

#include <stdio.h>
#include <time.h>

static void pbuf(void *buf, UWORD n, char *s)
{
    UWORD i;
    INT8 *cp = (INT8 *)buf;
    
    if (n <= 0 || n >= 30)
        n = 30;
    
    if (s)
        printf("%s: ", s);
        
    for (i = 0; i < n; i++)
        printf("%2X", (unsigned char)cp[i]);
    printf("\n");
}

static void primitive_verify(void)
{
    #if (UMAC_KEY_LEN == 16)
    char key[] = "\x01\x23\x45\x67\x89\xAB\xCD\xEF"
                 "\x01\x12\x23\x34\x45\x56\x67\x78";
    char res[] = "524E192F4715C6231F51F6367EA43F18";
    #elif (UMAC_KEY_LEN == 32)
    char key[] = "\x01\x23\x45\x67\x89\xAB\xCD\xEF"
                 "\x01\x12\x23\x34\x45\x56\x67\x78"
                 "\x89\x9A\xAB\xBC\xCD\xDE\xEF\xF0"
                 "\x10\x32\x54\x76\x98\xba\xdc\xfe";
    char res[] = "C8241816F0D7E48920AD16A1674E5D48";
    #endif
    char pt[]  = "\x02\x13\x24\x35\x46\x57\x68\x79"
                 "\x8A\x9B\xAC\xBD\xCE\xDF\xE0\xF1";
    UINT32 k1[RC6_TABLE_WORDS];
    
    RC6_SETUP(key, k1);
    RC6(k1, pt, pt);
    printf("\nRC6 Test\n");
    pbuf(pt, 16, "Digest is       ");
    printf("Digest should be: %s\n", res);
}

static void umac_verify(void)
{
    umac_ctx_t ctx;
    char *data_ptr;
    int data_len = 4 * 1024;
    char nonce[8] = {0};
    char tag[21] = {0};
    char tag2[21] = {0};
    int bytes_over_boundary, i, j;
    int inc[] = {1,99,512};
    
    /* Initialize Memory and UMAC */
    nonce[7] = 1;
    data_ptr = (char *)malloc(data_len + 16);
    bytes_over_boundary = (int)data_ptr & (16 - 1);
    if (bytes_over_boundary != 0)
        data_ptr += (16 - bytes_over_boundary);
    for (i = 0; i < data_len; i++)
        data_ptr[i] = (i%127) * (i%123) % 127;
    ctx = umac_new("abcdefghijklmnopqrstuvwxyz");
    
    umac(ctx, data_ptr, data_len, tag, nonce);
    umac_reset(ctx);

    #if ((WORD_LEN == 2) && (UMAC_OUTPUT_LEN == 8) && \
         (L1_KEY_LEN == 1024) && (UMAC_KEY_LEN == 16))
    printf("UMAC-2/8/1024/16/LITTLE/SIGNED Test\n");
    pbuf(tag, PREFIX_STREAMS*WORD_LEN, "Tag is                   ");
    printf("Tag should be a prefix of: %s\n", "F238FCB732528C51");
    #elif ((WORD_LEN == 4) && (UMAC_OUTPUT_LEN == 8) && \
         (L1_KEY_LEN == 1024) && (UMAC_KEY_LEN == 16))
    printf("UMAC-2/8/1024/16/LITTLE/SIGNED Test\n");
    pbuf(tag, PREFIX_STREAMS*WORD_LEN, "Tag is                   ");
    printf("Tag should be a prefix of: %s\n", "64BA28EFC26175AB");
    #endif



    printf("\nVerifying consistancy of single- and"
           " multiple-call interfaces.\n");
    for (i = 1; i < (int)(sizeof(inc)/sizeof(inc[0])); i++) {
            for (j = 0; j <= data_len-inc[i]; j+=inc[i])
                umac_update(ctx, data_ptr+j, inc[i]);
            umac_final(ctx, tag, nonce);
            umac_reset(ctx);

            umac(ctx, data_ptr, (data_len/inc[i])*inc[i], tag2, nonce);
            umac_reset(ctx);
            nonce[7]++;
            
            if (memcmp(tag,tag2,sizeof(tag)))
                printf("\ninc = %d data_len = %d failed!\n",
                       inc[i], data_len);
    }
    printf("Done.\n");
    umac_delete(ctx);
}

static void speed_test(void)
{
    clock_t ticks;
    double secs,gbits;
    umac_ctx_t uc;
    int bytes_over_boundary;
    char *data, *orig_data;
    char key[32] = "abcdefghijklmnopqrst";
    int data_len = 2048;
    char nonce[8] = {0};
    char tag[20];
    long iters_per_tag = 64;
    long tag_iters = (500 * 1024 * 1024) / (data_len * iters_per_tag),
         i, j;

    /* Allocate memory and align to boundary multiple */
    orig_data = data = (char *)malloc(data_len + ALLOC_BOUNDARY);
    bytes_over_boundary = (int)data & (ALLOC_BOUNDARY - 1);
    if ((bytes_over_boundary != 0) && (ALLOC_BOUNDARY))
        data += (ALLOC_BOUNDARY - bytes_over_boundary);
    for (i = 0; i < data_len; i++)
        data[i] = (i*i) % 127;
    uc = umac_new(key);
    
    printf("\nSpeed Test (%ld bytes/iter * %ld iters) ...\n",
            iters_per_tag * data_len, tag_iters);

    i = tag_iters;
    ticks = clock();
    do {
        j = iters_per_tag;
        do {
            umac_update(uc, data, data_len);
        } while (--j);
        umac_final(uc, tag, nonce);
        nonce[7] += 1;
    } while (--i);
    ticks = clock() - ticks;
    
    umac_delete(uc);
    free(orig_data);
    
    secs = (double)ticks / CLOCKS_PER_SEC;
    gbits = (8.0 * data_len * tag_iters * iters_per_tag) / (1.0e9);
    
    printf("MAC'd %G MB in %G seconds (%G Gbit/sec).\n",
           (1000.0/8.0) * gbits, secs, gbits/secs);
}

int main(void)
{
    umac_verify();
    primitive_verify();
    speed_test();
    /* printf("Push return to continue\n"); getchar(); */
    return (1);
}

#endif
