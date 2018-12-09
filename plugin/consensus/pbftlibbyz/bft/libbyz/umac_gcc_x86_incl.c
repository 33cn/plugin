/* ----------------------------------------------------------------------
 * Rules for writing an architecture specific include file
 *
 * - For any "class" of functions written here (eg. ARCH_ROTL), all
 *   functions in the class must be written here.
 * - For each "class" written, define the class macro as 1
 *   (eg. #define ARCH_ROTL 1).
 * - This file is included because we are using extensions to ANSI C,
 *   but you must distinguish between "intrinsic" and "intrinsic+asm"
 *   extensions. This is easily done by writing this file in three
 *   sections: (1) intrinsic only functions, (2) intrinsic functions for
 *   which there exists an assembly equivalent and (3) assembly functions.
 *   If we are to do "intrinsic" extensions, then (1) and (2) should be
 *   compiled, otherwise if we are "intrinsic+asm", then (1) and (3).
 *   The assumption is that for speed, C < C+intrinsics < C+assembly.
 * ---------------------------------------------------------------------- */


/* ---------------------------------------------------------------------- */
/* ---------------------------------------------------------------------- */
/* ---------------------------------------------------------------------- */
/* First define routines which are only written using compiler intrinsics */
/* ---------------------------------------------------------------------- */
/* ---------------------------------------------------------------------- */
/* ---------------------------------------------------------------------- */

              /* GCC has no intrinsics */

/* ----------------------------------------------------------------------
 * ----------------------------------------------------------------------
 * ----------------------------------------------------------------------
 * Second define routines which are written using compiler intrinsics but
 * which have assembly equivalents in the third section.
 * ----------------------------------------------------------------------
 * ----------------------------------------------------------------------
 * ---------------------------------------------------------------------- */
#if ( ! USE_C_AND_ASSEMBLY)  /* Intrinsics only allowed */
/* ---------------------------------------------------------------------- */

              /* GCC has no intrinsics */


/* ----------------------------------------------------------------------
 * ----------------------------------------------------------------------
 * ----------------------------------------------------------------------
 * Third define routines which are written using inline assembly.
 * ----------------------------------------------------------------------
 * ----------------------------------------------------------------------
 * ---------------------------------------------------------------------- */
#else /* (USE_C_AND_ASSEMBLY) */
/* ---------------------------------------------------------------------- */


/* ---------------------------------------------------------------------- */
#define ARCH_ENDIAN_LS  1
/* ---------------------------------------------------------------------- */
#if (ARCH_ENDIAN_LS)

static UINT32 LOAD_UINT32_REVERSED(void *ptr)
{
    UINT32 temp;
    asm volatile("bswap %0" : "=r" (temp) : "0" (*(UINT32 *)ptr)); 
    return temp;
}

static void STORE_UINT32_REVERSED(void *ptr, UINT32 x)
{
    asm volatile("bswap %0" : "=r" (*(UINT32 *)ptr) : "0" (x)); 
}


 UINT16 LOAD_UINT16_REVERSED(void *ptr)
{
    UINT16 temp;
    asm volatile(
        "movw (%1), %0\n\t"
        "rolw $8, %0\n\t"
    : "=r" (temp) : "r" (ptr)  ); 
    return temp;
}

 void STORE_UINT16_REVERSED(void *ptr, UINT16 x)
{
    asm volatile(
        "rolw $8, %0\n\t"
        "movw %0, (%1)\n\t"
    :  "+r"(x) : "r" (ptr) ); 
}

#endif

/* ---------------------------------------------------------------------- */
#define ARCH_RC6  1
/* ---------------------------------------------------------------------- */
#if (ARCH_RC6)

#define RC6_BLOCK(a,b,c,d,n) \
    "leal 1(%%"#b",%%"#b"),%%eax\n\t" \
    "imul %%"#b",%%eax\n\t" \
    "roll $5,%%eax\n\t" \
    "leal 1(%%"#d",%%"#d"),%%ecx\n\t" \
    "imul %%"#d",%%ecx\n\t" \
    "roll $5,%%ecx\n\t" \
    "xorl %%eax,%%"#a"\n\t" \
    "roll %%cl,%%"#a"\n\t" \
    "addl "#n"(%%esi),%%"#a"\n\t" \
    "xorl %%ecx,%%"#c"\n\t" \
    "movl %%eax,%%ecx\n\t" \
    "roll %%cl,%%"#c"\n\t" \
    "addl "#n"+4(%%esi),%%"#c"\n\t"

static void RC6(UINT32 S[], void *pt, void *ct)
{ 
    UINT32 *s = S;
    UINT32 A = ((UINT32 *)pt)[0];
    UINT32 B = ((UINT32 *)pt)[1] + *s++;
    UINT32 C = ((UINT32 *)pt)[2];
    UINT32 D = ((UINT32 *)pt)[3] + *s++;


    asm volatile (
        "pushl %%ebp\n\t"
        "movl %%ecx,%%ebp\n\t"
      
        RC6_BLOCK(edi,ebx,ebp,edx,0)
        RC6_BLOCK(ebx,ebp,edx,edi,8)
        RC6_BLOCK(ebp,edx,edi,ebx,16)
        RC6_BLOCK(edx,edi,ebx,ebp,24)
      
        RC6_BLOCK(edi,ebx,ebp,edx,32)
        RC6_BLOCK(ebx,ebp,edx,edi,40)
        RC6_BLOCK(ebp,edx,edi,ebx,48)
        RC6_BLOCK(edx,edi,ebx,ebp,56)
      
        RC6_BLOCK(edi,ebx,ebp,edx,64)
        RC6_BLOCK(ebx,ebp,edx,edi,72)
        RC6_BLOCK(ebp,edx,edi,ebx,80)
        RC6_BLOCK(edx,edi,ebx,ebp,88)
      
        RC6_BLOCK(edi,ebx,ebp,edx,96)
        RC6_BLOCK(ebx,ebp,edx,edi,104)
        RC6_BLOCK(ebp,edx,edi,ebx,112)
        RC6_BLOCK(edx,edi,ebx,ebp,120)
      
        RC6_BLOCK(edi,ebx,ebp,edx,128)
        RC6_BLOCK(ebx,ebp,edx,edi,136)
        RC6_BLOCK(ebp,edx,edi,ebx,144)
        RC6_BLOCK(edx,edi,ebx,ebp,152)
      
        "movl %%ebp,%%ecx\n\t"
        "popl %%ebp"
        : "+D" (A), "+b" (B), "+c" (C), "+d" (D)
        : "S" (s)
        : "eax");
    

    A += *(s+40);
    C += *(s+41);
    ((UINT32 *)ct)[0] = A; 
    ((UINT32 *)ct)[1] = B;  
    ((UINT32 *)ct)[2] = C; 
    ((UINT32 *)ct)[3] = D;  
} 
#endif


#if (WORD_LEN == 4)
/* ---------------------------------------------------------------------- */
#define ARCH_NH32  1
/* ---------------------------------------------------------------------- */
#if (ARCH_NH32)

#define NH_BLOCK(n) \
    "movl "#n"(%%ebx),%%eax\n\t" \
    "movl "#n"+16(%%ebx),%%edx\n\t" \
    "addl "#n"(%%ecx),%%eax\n\t" \
    "addl "#n"+16(%%ecx),%%edx\n\t" \
    "mull %%edx\n\t" \
    "addl %%eax,%%esi\n\t" \
    "adcl %%edx,%%edi\n\t"
    
static void nh_aux_8(void *kp, void *dp, void *hp, UINT32 dlen)
/* NH hashing primitive. Previous (partial) hash result is loaded and     */
/* then stored via hp pointer. The length of the data pointed at by dp is */
/* guaranteed to be divisible by HASH_BUF_BYTES (64), which means we can   */
/* optimize by unrolling the loop. 64 bits are written at hp.             */
{
  UINT32 *p = (UINT32 *)hp;
  
  asm volatile (
    "\n\t"
    "pushl %%ebp\n\t"
    "movl %%eax,%%ebp\n\t"
    "shrl $6,%%ebp\n\t"
    "pushl %%edx\n\t"
    "pushl %%eax\n\t"
    "testl %%ebp,%%ebp\n\t"
    "je 2f\n\t"
    ".align 4,0x90\n"
    "1:\n\t"
    
    NH_BLOCK(0)
    NH_BLOCK(4)
    NH_BLOCK(8)
    NH_BLOCK(12)
    NH_BLOCK(32)
    NH_BLOCK(36)
    NH_BLOCK(40)
    NH_BLOCK(44)

    "addl $64,%%ecx\n\t"
    "addl $64,%%ebx\n\t"
    "decl %%ebp\n\t"
    "jne 1b\n\t"
    ".align 4\n"
    "2:\n\t"
    "movl (%%esp),%%ebp\n\t"
    "andl $63,%%ebp\n\t"
    "shrl $5,%%ebp\n\t"
    "testl %%ebp,%%ebp\n\t"
    "je 4f\n\t"
    ".align 4,0x90\n"
    "3:\n\t"
    
    NH_BLOCK(0)
    NH_BLOCK(4)
    NH_BLOCK(8)
    NH_BLOCK(12)

    "addl $32,%%ecx\n\t"
    "addl $32,%%ebx\n\t"
    "decl %%ebp\n\t"
    "jne 3b\n\t"
    ".align 4\n"
    "4:\n\t"
    "popl %%eax\n\t"
    "popl %%edx\n\t"
    "popl %%ebp"
    : "+S" (p[0]), "+D" (p[1]), "+c" (kp), "+b" (dp)
    : "a" (dlen)
    );
}

/* ---------------------------------------------------------------------- */

static void nh_aux_16(void *kp, void *dp, void *hp, UINT32 dlen)
/* NH hashing primitive. 128 bits are written at hp by performing two     */
/* passes over the data with the second key being the toeplitz shift of   */
/* the first.                                                             */
{
    nh_aux_8(kp,dp,hp,dlen);
    nh_aux_8((INT8 *)kp+16,dp,(INT8 *)hp+8,dlen);
}

/* ---------------------------------------------------------------------- */

#endif
#else

/* ---------------------------------------------------------------------- */
#define ARCH_NH16  1
/* ---------------------------------------------------------------------- */
#if (ARCH_NH16)
#define MMX_BLOCK_128(n) \
    "movq "#n"+0(%%eax),%%mm0\n\t" \
    "movq "#n"+16(%%eax),%%mm1\n\t" \
    "movq "#n"+0(%%ebx),%%mm2\n\t" \
    "movq "#n"+16(%%ebx),%%mm3\n\t" \
    "paddw %%mm0,%%mm2\n\t" \
    "paddw %%mm1,%%mm3\n\t" \
    "pmaddwd %%mm3,%%mm2\n\t" \
    "paddd %%mm2,%%mm4\n\t" \
    "psubw %%mm1,%%mm3\n\t"         \
    "movq "#n"+32(%%ebx),%%mm2\n\t" \
    "paddw %%mm0,%%mm3\n\t" \
    "paddw %%mm1,%%mm2\n\t" \
    "pmaddwd %%mm2,%%mm3\n\t" \
    "paddd %%mm3,%%mm5\n\t" \
    "psubw %%mm1,%%mm2\n\t"         \
    "movq "#n"+48(%%ebx),%%mm3\n\t" \
    "paddw %%mm0,%%mm2\n\t" \
    "paddw %%mm1,%%mm3\n\t" \
    "pmaddwd %%mm3,%%mm2\n\t" \
    "paddd %%mm2,%%mm6\n\t" \
    "psubw %%mm1,%%mm3\n\t"         \
    "movq "#n"+64(%%ebx),%%mm2\n\t" \
    "paddw %%mm0,%%mm3\n\t" \
    "paddw %%mm1,%%mm2\n\t" \
    "pmaddwd %%mm2,%%mm3\n\t" \
    "paddd %%mm3,%%mm7\n\t" \
    "movq "#n"+8(%%eax),%%mm0\n\t" \
    "movq "#n"+24(%%eax),%%mm1\n\t" \
    "movq "#n"+8(%%ebx),%%mm2\n\t" \
    "movq "#n"+24(%%ebx),%%mm3\n\t" \
    "paddw %%mm0,%%mm2\n\t" \
    "paddw %%mm1,%%mm3\n\t" \
    "pmaddwd %%mm3,%%mm2\n\t" \
    "paddd %%mm2,%%mm4\n\t" \
    "psubw %%mm1,%%mm3\n\t"         \
    "movq "#n"+40(%%ebx),%%mm2\n\t" \
    "paddw %%mm0,%%mm3\n\t" \
    "paddw %%mm1,%%mm2\n\t" \
    "pmaddwd %%mm2,%%mm3\n\t" \
    "paddd %%mm3,%%mm5\n\t" \
    "psubw %%mm1,%%mm2\n\t"         \
    "movq "#n"+56(%%ebx),%%mm3\n\t" \
    "paddw %%mm0,%%mm2\n\t" \
    "paddw %%mm1,%%mm3\n\t" \
    "pmaddwd %%mm3,%%mm2\n\t" \
    "paddd %%mm2,%%mm6\n\t" \
    "psubw %%mm1,%%mm3\n\t"         \
    "movq "#n"+72(%%ebx),%%mm2\n\t" \
    "paddw %%mm0,%%mm3\n\t" \
    "paddw %%mm1,%%mm2\n\t" \
    "pmaddwd %%mm2,%%mm3\n\t" \
    "paddd %%mm3,%%mm7\n\t"

#define MMX_BLOCK_64(n)                \
    "movq "#n"+0(%%eax),%%mm0\n\t"  \
    "movq "#n"+16(%%eax),%%mm1\n\t" \
    "movq "#n"+0(%%ebx),%%mm2\n\t"  \
    "movq "#n"+16(%%ebx),%%mm3\n\t" \
    "paddw %%mm0,%%mm2\n\t"         \
    "paddw %%mm1,%%mm3\n\t"         \
    "pmaddwd %%mm3,%%mm2\n\t"       \
    "paddd %%mm2,%%mm4\n\t"         \
    "psubw %%mm1,%%mm3\n\t"         \
    "movq "#n"+32(%%ebx),%%mm2\n\t" \
    "paddw %%mm0,%%mm3\n\t"         \
    "paddw %%mm1,%%mm2\n\t"         \
    "pmaddwd %%mm3,%%mm2\n\t"       \
    "paddd %%mm2,%%mm5\n\t"         \
    "movq "#n"+8(%%eax),%%mm0\n\t"  \
    "movq "#n"+24(%%eax),%%mm1\n\t" \
    "movq "#n"+8(%%ebx),%%mm2\n\t"  \
    "movq "#n"+24(%%ebx),%%mm3\n\t" \
    "paddw %%mm0,%%mm2\n\t"         \
    "paddw %%mm1,%%mm3\n\t"         \
    "pmaddwd %%mm3,%%mm2\n\t"       \
    "paddd %%mm2,%%mm4\n\t"         \
    "psubw %%mm1,%%mm3\n\t"         \
    "movq "#n"+40(%%ebx),%%mm2\n\t" \
    "paddw %%mm0,%%mm3\n\t"         \
    "paddw %%mm1,%%mm2\n\t"         \
    "pmaddwd %%mm3,%%mm2\n\t"       \
    "paddd %%mm2,%%mm5\n\t"

#define MMX_BLOCK_32(n)                \
    "movq "#n"+0(%%eax),%%mm0\n\t"  \
    "movq "#n"+16(%%eax),%%mm1\n\t" \
    "movq "#n"+0(%%ebx),%%mm2\n\t"  \
    "movq "#n"+16(%%ebx),%%mm3\n\t" \
    "paddw %%mm0,%%mm2\n\t"         \
    "paddw %%mm1,%%mm3\n\t"         \
    "movq "#n"+8(%%eax),%%mm4\n\t"  \
    "movq "#n"+24(%%eax),%%mm6\n\t" \
    "pmaddwd %%mm3,%%mm2\n\t"       \
    "movq "#n"+8(%%ebx),%%mm5\n\t"  \
    "movq "#n"+24(%%ebx),%%mm3\n\t" \
    "paddd %%mm2,%%mm7\n\t"         \
    "paddw %%mm4,%%mm5\n\t"         \
    "paddw %%mm6,%%mm3\n\t"         \
    "pmaddwd %%mm3,%%mm5\n\t"       \
    "paddd %%mm5,%%mm7\n\t"         

static void nh_aux_4(void *kp, void *dp, void *hp, UINT32 dlen)
/* NH hashing primitive. Previous (partial) hash result is loaded and     */
/* then stored via hp pointer. The length of the data pointed at by dp is */
/* guaranteed to be divisible by HASH_BUF_BYTES (64), which means we can   */
/* optimize by unrolling the loop. 64 bits are written at hp by           */
/* performing two passes over the data with the second key being the      */
/* toeplitz shift of the first.                                           */
{
  UINT32 t[2];
  UINT32 *p = (UINT32 *)hp;
  
  asm volatile (
    "\n\t"
    "pushl %%edx\n\t"
    "shrl $7,%%edx\n\t"
    "pxor %%mm7,%%mm7\n\t"
    "testl %%edx,%%edx\n\t"
    "je 2f\n\t"
    ".align 4,0x90\n"
    "1:\t"

    MMX_BLOCK_32(0)
    MMX_BLOCK_32(32)
    MMX_BLOCK_32(64)
    MMX_BLOCK_32(96)
    
    "addl $128,%%eax\n\t"
    "addl $128,%%ebx\n\t"
    "decl %%edx\n\t"
    "jne 1b\n"
    ".align 4\n"
    "2:\n\t"
    "movl (%%esp),%%edx\n\t"
    "andl $127,%%edx\n\t"
    "shrl $5,%%edx\n\t"
    "testl %%edx,%%edx\n\t"
    "je 4f\n\t"
    ".align 4,0x90\n"
    "3:\t"

    MMX_BLOCK_32(0)
    
    "addl $32,%%eax\n\t"
    "addl $32,%%ebx\n\t"
    "decl %%edx\n\t"
    "jne 3b\n"
    ".align 4\n"
    "4:\n\t"
    "movq %%mm7,(%%ecx)\n\t"
    "popl %%edx\n\t"
    "emms"
    : "+a" (dp), "+b" (kp), "+d" (dlen)
    : "c" (t+0)
    : "memory");
    p[0] = p[0] + t[0] + t[1];
}

/* ---------------------------------------------------------------------- */

static void nh_aux_8(void *kp, void *dp, void *hp, UINT32 dlen)
/* NH hashing primitive. Previous (partial) hash result is loaded and     */
/* then stored via hp pointer. The length of the data pointed at by dp is */
/* guaranteed to be divisible by HASH_BUF_BYTES (64), which means we can   */
/* optimize by unrolling the loop. 64 bits are written at hp by           */
/* performing two passes over the data with the second key being the      */
/* toeplitz shift of the first.                                           */
{
  UINT32 t[4];
  UINT32 *p = (UINT32 *)hp;
  
  asm volatile (
    "\n\t"
    "pushl %%edx\n\t"
    "shrl $7,%%edx\n\t"
    "pxor %%mm4,%%mm4\n\t"
    "pxor %%mm5,%%mm5\n\t"
    "testl %%edx,%%edx\n\t"
    "je 2f\n\t"
    ".align 4,0x90\n"
    "1:\t"

    MMX_BLOCK_64(0)
    MMX_BLOCK_64(32)
    MMX_BLOCK_64(64)
    MMX_BLOCK_64(96)
    
    "addl $128,%%eax\n\t"
    "addl $128,%%ebx\n\t"
    "decl %%edx\n\t"
    "jne 1b\n"
    ".align 4\n"
    "2:\n\t"
    "movl (%%esp),%%edx\n\t"
    "andl $127,%%edx\n\t"
    "shrl $5,%%edx\n\t"
    "testl %%edx,%%edx\n\t"
    "je 4f\n\t"
    ".align 4,0x90\n"
    "3:\t"

    MMX_BLOCK_64(0)
    
    "addl $32,%%eax\n\t"
    "addl $32,%%ebx\n\t"
    "decl %%edx\n\t"
    "jne 3b\n"
    ".align 4\n"
    "4:\n\t"
    "movq %%mm4,(%%ecx)\n\t"
    "movq %%mm5,8(%%ecx)\n\t"
    "popl %%edx\n\t"
    "emms"
    : "+a" (dp), "+b" (kp), "+d" (dlen)
    : "c" (t+0)
    : "memory");
    p[0] = p[0] + t[0] + t[1];
    p[1] = p[1] + t[2] + t[3];
}

/* ---------------------------------------------------------------------- */

static void nh_aux_16(void *kp, void *dp, void *hp, UINT32 dlen)
/* NH hashing primitive. Previous (partial) hash result is loaded and     */
/* then stored via hp pointer. The length of the data pointed at by dp is */
/* guaranteed to be divisible by HASH_BUF_BYTES (64), which means we can   */
/* optimize by unrolling the loop. 128 bits are written at hp by          */
/* performing four passes over the data with the later keys being the     */
/* toeplitz shift of the first.                                           */
{
  UINT32 t[8];
  UINT32 *p = (UINT32 *)hp;
  
  asm volatile (
    "\n\t"
    "pushl %%edx\n\t"
    "shrl $7,%%edx\n\t"
    "pxor %%mm4,%%mm4\n\t"
    "pxor %%mm5,%%mm5\n\t"
    "pxor %%mm6,%%mm6\n\t"
    "pxor %%mm7,%%mm7\n\t"
    "testl %%edx,%%edx\n\t"
    "je 2f\n\t"
    ".align 4,0x90\n"
    "1:\t"

    MMX_BLOCK_128(0)
    MMX_BLOCK_128(32)
    MMX_BLOCK_128(64)
    MMX_BLOCK_128(96)
    
    "addl $128,%%eax\n\t"
    "addl $128,%%ebx\n\t"
    "decl %%edx\n\t"
    "jne 1b\n"
    ".align 4\n"
    "2:\n\t"
    "movl (%%esp),%%edx\n\t"
    "andl $127,%%edx\n\t"
    "shrl $5,%%edx\n\t"
    "testl %%edx,%%edx\n\t"
    "je 4f\n\t"
    ".align 4,0x90\n"
    "3:\t"

    MMX_BLOCK_128(0)
    
    "addl $32,%%eax\n\t"
    "addl $32,%%ebx\n\t"
    "decl %%edx\n\t"
    "jne 3b\n"
    ".align 4\n"
    "4:\n\t"
    "movq %%mm4,(%%ecx)\n\t"
    "movq %%mm5,8(%%ecx)\n\t"
    "movq %%mm6,16(%%ecx)\n\t"
    "movq %%mm7,24(%%ecx)\n\t"
    "popl %%edx\n\t"
    "emms"
    : "+a" (dp), "+b" (kp), "+d" (dlen)
    : "c" (t+0)
    : "memory");
    p[0] = p[0] + t[0] + t[1];
    p[1] = p[1] + t[2] + t[3];
    p[2] = p[2] + t[4] + t[5];
    p[3] = p[3] + t[6] + t[7];
}

/* ---------------------------------------------------------------------- */

#endif
#endif

/* ---------------------------------------------------------------------- */
#define ARCH_POLY  1
/* ---------------------------------------------------------------------- */
#if (ARCH_POLY)

static UINT32 poly32(UINT32 cur, UINT32 key, UINT32 data)
{
    asm volatile (
        "mull %1\n\t"
        "leal (%%edx, %%edx, 4), %%edx\n\t"
        "addl %%edx,%%eax\n\t"
        "leal 5(%%eax),%%edx\n\t"
        "cmovb %%edx,%%eax\n\t"
        "addl %2,%%eax\n\t"
        "leal 5(%%eax),%%edx\n\t"
        "cmovb %%edx,%%eax"
    : "+a"(cur)
    : "g"(key), "g"(data)
    : "edx");
    return cur;
}

static UINT64 poly64(UINT64 cur, UINT64 key, UINT64 data)
{
    UINT32 data_hi = (UINT32)(data >> 32),
           data_lo = (UINT32)(data);
    
    /* 8(esp) = scratch
       4(esp) = cur
       0(esp) = cur */
    asm volatile(
        "subl $12,%%esp\n\t"
        "movl %%esp,%%ecx\n\t"              /* (ecx) = cur */
        "movl %%edx, 4(%%ecx)\n\t"
        "movl %%eax, 0(%%ecx)\n\t"
        "movl 0(%%ebx),%%eax\n\t"        /* eax = keylo                   */
        "mull 4(%%ecx)\n\t"              /* edx:eax = eax * curhi         */
        "movl %%eax, %%esi\n\t"          /* edi:esi = edx:eax */
        "movl %%edx, %%edi\n\t"
        "movl 4(%%ebx),%%eax\n\t"        /* eax = keyhi                   */
        "mull 0(%%ecx)\n\t"              /* edx:eax  = eax * curlo        */
        "addl %%eax, %%esi\n\t"          /* edi:(esp) += edx:eax           */
        "adcl %%edx, %%edi\n\t"
        "movl %%esi,8(%%esp)\n\t"
        "movl 4(%%ebx),%%eax\n\t"        /* eax = keyhi                   */
        "mull 4(%%ecx)\n\t"              /* edx:eax  = eax * curhi        */
        "xorl %%esi,%%esi\n\t"           /* esi:edi  = (edi:(esp) >> 32)  */
        "addl %%eax,%%edi\n\t"           /* esi:edi += edx:eax            */
        "adcl %%edx,%%esi\n\t"   
        "movl %%esi,%%edx\n\t"           /* edx:eax  = esi:edi */
        "movl %%edi,%%eax\n\t"     
        "shldl $6,%%edi,%%esi\n\t"       /* esi:edi <<= 6                 */
        "shll $6,%%edi\n\t"
        "subl %%eax,%%edi\n\t"           /* esi:edi  -= edx:eax           */
        "sbbl %%edx,%%esi\n\t"     
        "shldl $2,%%eax,%%edx\n\t"       /* esi:edi <<= 2                 */
        "shll $2,%%eax\n\t"
        "subl %%eax,%%edi\n\t"           /* esi:edi  -= edx:eax */
        "sbbl %%edx,%%esi\n\t"
        "movl 0(%%ebx),%%eax\n\t"        /* eax = keylo                   */
        "mull 0(%%ecx)\n\t"              /* edx:eax = eax * curlo         */
        "addl %%edi,%%eax\n\t"           /* edx:eax += esi:edi */
        "adcl %%esi,%%edx\n\t"
        "addl 8(%%esp), %%edx\n\t"    /* edx:eax += ((esp) << 32) */
        "movl $59,%%esi\n\t"       /* esi = 59                      */
        "movl $0,%%edi\n\t"        /* edi = 0                       */
        "cmovb %%esi,%%edi\n\t"    /* if (carry) edi = 59 else edi = 0 */
        "addl $12,%%esp\n\t"
        "addl %3, %%eax\n\t"       /* edx:eax += data               */
        "adcl %4, %%edx\n\t"
        "movl $0,%%ecx\n\t"        /* edi = 0                       */
        "cmovb %%esi,%%ecx\n\t"    /* if (carry) edi = 59 else edi = 0 */
        "addl %%ecx, %%edi\n\t"    /* edx:eax += edi                */
        "addl %%edi, %%eax\n\t"    /* edx:eax += edi                */
        "adcl $0,%%edx\n\t"
    : "+A"(cur)
    : "m"(key),"b"(&key),"g"(data_lo),"g"(data_hi)
    : "ecx","edi","esi");
    
    return cur;
}

#endif

#if (WORD_LEN == 4)
/* ---------------------------------------------------------------------- */
#define ARCH_IP  1
/* ---------------------------------------------------------------------- */
#if (ARCH_IP)

static UINT64 ip_aux(UINT64 t, UINT64 *ipkp, UINT64 data)
{
    UINT32 dummy1, dummy2;
    asm volatile(
        "pushl %%ebp\n\t"
        "movl %%eax,%%esi\n\t"
        "movl %%edx,%%ebp\n\t"
        "movl %%ebx,%%eax\n\t"
        "shrl $16,%%eax\n\t"
        "mull 0(%%edi)\n\t"
        "addl %%eax,%%esi\n\t"
        "adcl %%edx,%%ebp\n\t"
        "movl %%ebx,%%eax\n\t"
        "shrl $16,%%eax\n\t"
        "mull 4(%%edi)\n\t"
        "addl %%eax,%%ebp\n\t"

        "movzwl %%bx,%%eax\n\t"
        "mull 8(%%edi)\n\t"
        "addl %%eax,%%esi\n\t"
        "adcl %%edx,%%ebp\n\t"
        "movzwl %%bx,%%eax\n\t"
        "mull 12(%%edi)\n\t"
        "addl %%eax,%%ebp\n\t"

        "movl %%ecx,%%eax\n\t"
        "shrl $16,%%eax\n\t"
        "mull 16(%%edi)\n\t"
        "addl %%eax,%%esi\n\t"
        "adcl %%edx,%%ebp\n\t"
        "movl %%ecx,%%eax\n\t"
        "shrl $16,%%eax\n\t"
        "mull 20(%%edi)\n\t"
        "addl %%eax,%%ebp\n\t"

        "movzwl %%cx,%%eax\n\t"
        "mull 24(%%edi)\n\t"
        "addl %%eax,%%esi\n\t"
        "adcl %%edx,%%ebp\n\t"
        "movzwl %%cx,%%eax\n\t"
        "mull 28(%%edi)\n\t"
        "leal (%%eax,%%ebp),%%edx\n\t"
        "movl %%esi,%%eax\n\t"
        "popl %%ebp"
    : "+A"(t), "=b"(dummy1), "=c"(dummy2)
    : "D"(ipkp), "1"((UINT32)(data>>32)), "2"((UINT32)data)
    : "esi");
    
    return t;
}

static UINT32 ip_reduce_p36(UINT64 t)
{
    asm volatile(
        "movl %%edx,%%edi\n\t"
        "andl $15,%%edx\n\t"
        "shrl $4,%%edi\n\t"
        "leal (%%edi,%%edi,4),%%edi\n\t"
        "addl %%edi,%%eax\n\t"
        "adcl $0,%%edx\n\t"
    : "+A"(t)
    :
    : "edi");

    if (t >= p36)
        t -= p36;

    return (SMALL_UWORD)(t);
}
#endif
#else

/* ---------------------------------------------------------------------- */
#define ARCH_IP  1
/* ---------------------------------------------------------------------- */
#if (ARCH_IP)

static UINT64 ip_aux(UINT64 t, UINT32 *ipkp, UINT32 data)
{
    UINT32 dummy;
    asm volatile(
        "movl %%eax,%%esi\n\t"
        "movl %%edx,%%edi\n\t"
        "movl %%ecx,%%eax\n\t"
        "shrl $16,%%eax\n\t"
        "mull (%2)\n\t"
        "addl %%eax,%%esi\n\t"
        "adcl %%edx,%%edi\n\t"
        "movzwl %%cx,%%eax\n\t"
        "mull 4(%2)\n\t"
        "addl %%esi,%%eax\n\t"
        "adcl %%edi,%%edx"
    : "+A"(t), "=c"(dummy)
    : "r"(ipkp), "1"(data)
    : "esi", "edi");
    
    return t;
}

static SMALL_UWORD ip_reduce_p19(UINT64 t)
{
    UINT32 ret;
    
    asm volatile(
        "shldl $13,%%eax,%%edx\n\t"
        "andl $524287,%%eax\n\t"
        "addl %%edx,%%eax\n\t"
    : "+A"(t)
    :
    : "edi");

    ret = (UINT32)t;
    if (ret >= p19)
        ret -= p19;

    return (SMALL_UWORD)(t);
    
}
#endif
#endif
#endif
