#ifndef _Time_h
#define _Time_h 1

/*
 * Definitions of various types.
 */
#include <limits.h>
#include <sys/socket.h>
#include <sys/param.h>
#include <netinet/in.h> 
#include <sys/time.h>
#include <unistd.h>

#ifdef USE_GETTIMEOFDAY

typedef struct timeval Time;
static inline Time currentTime() {
  Time t;
  int ret = gettimeofday(&t, 0);
  th_assert(ret == 0, "gettimeofday failed");
  return t;
}

static inline Time zeroTime() {  
  Time t;
  t.tv_sec = 0;
  t.tv_usec = 0
  return t; 
}

static inline long long diffTime(Time t1, Time t2) {
  // t1-t2 in microseconds.
  return (((unsigned long long)(t1.tv_sec-t2.tv_sec)) << 20) + (t1.tv_usec-t2.tv_usec);
}

static inline bool lessThanTime(Time t1, Time t2) {
  return t1.tv_sec < t2.tv_sec ||  
    (t1.tv_sec == t2.tv_sec &&  t1.tv_usec < t2.tv_usec);
}
#else

typedef long long Time;

#include "Cycle_counter.h"

extern long long clock_mhz;
// Clock frequency in MHz

extern void init_clock_mhz();
// Effects: Initialize "clock_mhz".

static inline Time currentTime() { return rdtsc(); }

static inline Time zeroTime() { return 0; }

static inline long long diffTime(Time t1, Time t2) {
  return (t1-t2)/clock_mhz;
}

static inline bool lessThanTime(Time t1, Time t2) {
  return t1 < t2;

}

#endif

#endif // _Time_h 
