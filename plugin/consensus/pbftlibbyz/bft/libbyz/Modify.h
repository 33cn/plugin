#ifndef _Modify_h
#define _Modify_h 1
#include "State_defs.h"

/* Needs to be kept in sync with Bitmap.h */

#ifdef NO_STATE_TRANSLATION

extern unsigned long *_Byz_cow_bits;
extern char *_Byz_mem;
extern void _Byz_modify_index(int bindex);
static const int _LONGBITS = sizeof(long)*8;

#define _Byz_cow_bit(bindex) (_Byz_cow_bits[bindex/_LONGBITS] & (1UL << (bindex%_LONGBITS)))

#define _Byz_modify1(mem)  do {                                    \
  int bindex;                                                      \
  bindex = ((char*)(mem)-_Byz_mem)/Block_size;                     \
  if (_Byz_cow_bit(bindex) == 0)                                   \
    _Byz_modify_index(bindex);                                     \
} while(0)

#define _Byz_modify2(mem,size) do {                                \
  int bindex1;                                                     \
  int bindex2;                                                     \
  char *_mem;                                                      \
  int _size;                                                       \
  _mem = (char*)(mem);                                             \
  _size = size;                                                    \
  bindex1 = (_mem-_Byz_mem)/Block_size;                            \
  bindex2 = (_mem+_size-1-_Byz_mem)/Block_size;                    \
  if ((_Byz_cow_bit(bindex1) == 0) | (_Byz_cow_bit(bindex2) == 0)) \
    Byz_modify(_mem,_size);                                        \
} while(0)

#endif

#endif /* _Modify_h */
