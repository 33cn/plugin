#include <stdio.h>
#include <sys/time.h>

main() {
  struct timeval tval;
  gettimeofday(&tval, 0);
  long long usecs = ((long long)(tval.tv_sec))*1000000+tval.tv_usec;
  printf("Time = %qd usecs\n", usecs);
}
