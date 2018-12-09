#include <stdlib.h>
#include "th_assert.h"
#include "types.h"
#include "Log.h"

template <class T> 
Log<T>::Log(int sz, Seqno h) : head(h), max_size(sz) {
  elems = new T[sz];
  mask = max_size-1;
}

template <class T> 
Log<T>::~Log() {
  delete [] elems;
}


template <class T> 
void Log<T>::clear(Seqno h) {
  for (int i=0; i < max_size; i++) 
    elems[i].clear();  

  head = h;    
}


template <class T>  
T &Log<T>::fetch(Seqno seqno) {
  th_assert(within_range(seqno), "Invalid argument\n");
  return elems[mod(seqno)];
}


template <class T>
void Log<T>::truncate(Seqno new_head) {
  if (new_head <= head) return;
  
  int i = head;
  int max = new_head;
  if (new_head - head >= max_size) {
    i = 0;
    max = max_size;
  }
    
  for (; i < max; i++) {
    elems[mod(i)].clear();
  }

  head = new_head;
}
