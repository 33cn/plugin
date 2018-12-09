#include <sys/mman.h>
#include "Log_allocator.h"

#ifndef MAP_VARIABLE
#define MAP_VARIABLE 0x00
#endif

Log_allocator::Log_allocator(int csz, int nc) {
  chunk_size = csz;
  max_num_chunks = nc;
  num_chunks = nc;
  free_chunks = 0;
  chunks = 0;
  chunks += chunk_size*max_num_chunks;
  cur = alloc_chunk();
}

Log_allocator::Chunk *Log_allocator::alloc_chunk() {
  Chunk *ret;
  if (free_chunks != 0) {
    // First try to allocate from free list
    ret = free_chunks;
    free_chunks = (Chunk*)(free_chunks->next);
  } else if (num_chunks < max_num_chunks) {
    // Try to allocate from the current chunks array.
    ret = (Chunk*)(chunks+chunk_size*num_chunks);
    num_chunks++;
  } else {
    // Allocate a new chunks array. The array must be chunk_size-aligned.
    void *addr = (void*)-1;
    for (int i=1; i <= 1000; i++) {
      addr = (void*)(chunks+(chunk_size*max_num_chunks)*i);
      addr = mmap(addr, chunk_size*max_num_chunks,
                 PROT_READ | PROT_WRITE, // R&W access
                 MAP_ANONYMOUS |         // Not from a file (zeroed pages)
                 MAP_VARIABLE |          
                 MAP_PRIVATE,            // Changes are private
                 -1,                     // No file
                 0);                     // No offset in file
      if (addr != (void*)-1 && ((long)addr)%chunk_size == 0) {
	break;
      } else { 
	addr = (void*)-1;
      }
    }

    if (addr == (void*)-1) {
      th_fail("Unable to allocate memory");
    }
    chunks = (char*)addr;
    ret = (Chunk*)chunks;
    num_chunks = 1;
  }

  ret->next = ret->data;
  ret->max = ret->next+(chunk_size-sizeof(Chunk));
  ret->nb = 1; // this is the current chunk
  return ret;
}


void Log_allocator::debug_print() {
  printf("Free space: current chunk\n");
  if (cur) {
    cur->debug_print();
  } else {
    printf("(null)\n");
  }

  printf("Free chunks:\n");
  for (Chunk *p=free_chunks; p != 0; p = (Chunk*)(p->next)) {
    p->debug_print();
  }

  printf("All chunks:");
  for (int i=0; i < max_num_chunks; i++) {
    Chunk *p = (Chunk *) (chunks+chunk_size*i);
    p->debug_print();
  }
}
