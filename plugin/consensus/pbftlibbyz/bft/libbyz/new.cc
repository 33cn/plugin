#include <new.h>

void *operator new(size_t sz) {
  return malloc(sz);
}

void operator delete(void *p) {
  return free(p);
}
