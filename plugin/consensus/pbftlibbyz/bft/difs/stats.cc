#include <stdio.h>
#include "stats.h"
#include "../libbyz/Cycle_counter.h"

Cycle_counter cc[NUM_COUNTERS];

void init_stats()
{
  for(int i=0; i<NUM_COUNTERS; i++)
    cc[i].reset();
}
    
 

void start_counter(int num)
{
  cc[num].start();
}

void stop_counter(int num)
{
  cc[num].stop();
}

void show_stats()
{
  for(int i=0; i<NUM_COUNTERS; i++)
    fprintf(stderr, "Counter %d had %lld cycles\n", i, cc[i].elapsed());
}


