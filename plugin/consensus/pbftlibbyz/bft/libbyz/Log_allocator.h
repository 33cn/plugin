#ifndef _Log_allocator_h
#define _Log_allocator_h 1

#include <stdio.h>
#include <string.h>
#include "types.h"
#include "th_assert.h"


// Since messages may contain other messages in the payload. It is
// important to ensure proper alignment to allow access to the fields
// of embedded messages. The following macros are used to check and
// enforce alignment requirements. All message pointers and message
// sizes must satisfy ALIGNED.

// Minimum required alignment for correctly accessing message fields.
// Must be a power of 2.
#define ALIGNMENT 8

// bool ALIGNED(void *ptr) or bool ALIGNED(long sz)
// Effects: Returns true iff the argument is aligned to ALIGNMENT
#define ALIGNED(ptr) (((long)(ptr))%ALIGNMENT == 0)

// int ALIGNED_SIZE(int sz)
// Effects: Increases sz to the least multiple of ALIGNMENT greater 
// than size.
#define ALIGNED_SIZE(sz) ((ALIGNED(sz)) ? (sz) : (sz)-(sz)%ALIGNMENT+ALIGNMENT)

#ifndef NDEBUG
#define DEBUG_ALLOC 1
#endif

class Log_allocator {
  // Overview: A fast and space efficient memory allocator. It assumes
  // objects that are allocated close together in time are freed close
  // together in time (otherwise it may waste a lot of memory). For
  // example, this assumption holds if the heap objects are allocated
  // as part of a sequential log and are deallocated when the log is
  // truncated. 

public:
  Log_allocator(int csz=65536, int nc=8);
  // Requires: "csz" is a multiple of the operating system vm page size.
  // Effects: Creates an allocator object with chunks of size "csz" and
  // an area for allocating chunks that can hold up to "nc" chunks.

  char *malloc(int sz);
  // Requires: size > 0 and sz < this.chunk_size
  // Effects: Allocates a heap block with "sz" bytes. The user of the
  // abstraction is responsible for keeping track of the size of the
  // returned block. 

  void free(char *p, int sz);
  // Requires: "p" was allocated by this allocator and has size "sz"
  // Effects: Frees "p". 

  bool realloc(char *p, int osz, int nsz);
  // Requires: "p" was allocated by this allocator and has size "osz".
  // Effects: Returns true, if it suceeds in converting "p" into a block
  // of size "nsz" (allocating more space or freeing it as necessary).
  // Otherwise, returns false and does nothing.

  void debug_print();
  // Effects: Prints debug information

private:
  struct Chunk {
    char *next;   // Pointer to beginning of free area
    char *max;    // Pointer to end of free area
    Long nb; // Reference count (number of blocks allocated in this) 
                  // plus one when this is the current block.
    char data[1];
    // Followed by extra data

    void debug_print() {
      printf("Chunk %p: next=%p max=%p nb=%d \n", this, next, max, (int)nb); 
    }
  };

  Chunk *alloc_chunk();
  // Effects: Allocates a new (current) chunk and initializes it

  void free_chunk(Chunk *p);
  // Effects: Frees the chunk pointed to by "p"
  

  Chunk *cur;         // current chunk
  int chunk_size;     // size of chunk

  char *chunks;       // array of "chunk_size" chunks
  int max_num_chunks; // maximum number of chunks in "chunks"
  int num_chunks;     // number of chunks already allocated in "chunks".

  Chunk *free_chunks; // list of free chunks
};

inline char *Log_allocator::malloc(int sz) {
  th_assert(sz > 0 && sz < chunk_size, "Invalid argument");
  th_assert(ALIGNED_SIZE(sz), "Invalid argument");
  register char *next;

  while (1) {
    next = cur->next;
    if (next+sz < cur->max) {
      // There is space in the current chunk
      cur->next = next+sz;
      cur->nb++;
#ifdef DEBUG_ALLOC
      bzero(next, sz);
#endif
      return next;
    }

    // Current chunk is full. Allocate a new one.
    if (cur->nb == 1) {
      // Can reuse current block.
      cur->next = cur->data;
    } else {
      // Allocate a new chunk
      cur->nb--; // To allow old chunk to be deallocated
      cur = alloc_chunk();
    }
  }
}

inline void Log_allocator::free_chunk(Chunk *p) {
  p->next = (char*)free_chunks;
  free_chunks = p;
}

#ifdef DEBUG_ALLOC
const int Log_allocator_magic = 0x386592a7;
#endif

inline void Log_allocator::free(char *p, int sz) {
  th_assert(ALIGNED_SIZE(sz), "Invalid argument");
  th_assert(ALIGNED(p), "Invalid argument");
  Chunk *pc = (Chunk*)((long)p & ~((long)chunk_size-1));

#ifdef DEBUG_ALLOC
  int *pi = (int *)p;
  for(int i=0; i < sz/4; i++) {
    if (*(pi+i) == Log_allocator_magic)
      fprintf(stderr, "WARNING: Storage possibly freed twice\n");
    *(pi+i) = Log_allocator_magic;
  }
#endif

  if (pc == cur && p+sz == cur->next) {
    // Adjust pointer to reuse space in current chunk
    cur->next -= sz;
  }
  
  pc->nb--;
  if (pc->nb == 0) {
    // The chunk can be freed
    th_assert(pc != cur, "Invalid state");
    free_chunk(pc);
  }
}

inline bool Log_allocator::realloc(char *p, int osz, int nsz) {
  Chunk *pc = (Chunk*)((long)p & ~((long)chunk_size-1));
  if (pc == cur && p+osz == cur->next) {
    int diff = nsz-osz;
    if (cur->next + diff < cur->max) {
      cur->next += diff;
      return true;
    } 
  }
  return false;
}

#endif // _Log_allocator_h
