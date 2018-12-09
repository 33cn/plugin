#include "Time.h"

long long clock_mhz = 0;

void init_clock_mhz() {
#ifndef USE_GETTIMEOFDAY
  struct timeval t0,t1;
  
  long long c0 = rdtsc();
  gettimeofday(&t0, 0);
  sleep(1);
  long long c1 = rdtsc();
  gettimeofday(&t1, 0);
    
  clock_mhz = (c1-c0)/((t1.tv_sec-t0.tv_sec)*1000000+t1.tv_usec-t0.tv_usec);
#endif
}

