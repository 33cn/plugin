#include "Set.h"

template <class T> 
Set<T>::Set(int sz) : max_size(sz), cur_size(0) {
  elems = (T**) malloc(sizeof(T*)*sz);  
  for (int  i=0; i < max_size; i++)
    elems[i] = 0;
}

template <class T> 
Set<T>::~Set() {
  for (int i=0; i < max_size; i++)
    if (elems[i]) delete elems[i];
  free(elems);
}

template <class T> 
bool Set<T>::store(T *e) {
  if (e->id() >= max_size || e->id() < 0 || elems[e->id()] != 0) 
    return false;
  elems[e->id()] = e;
  ++cur_size;
  return true;
}

template <class T>
T* Set<T>::fetch(int id) {
  if (id >= max_size || id < 0) 
    return 0;
  return elems[id];
}

template <class T>
T *Set<T>::remove(int id) {
  if (id >= max_size || id < 0 || elems[id] == 0)
    return 0;
  T *ret = elems[id];
  elems[id] = 0;
  cur_size--;
  return ret;
}


template <class T>
void Set<T>::clear() {
  for (int i=0; i < max_size; i++) {
    delete elems[i];
    elems[i] = 0;
  }
}
