#include <stdio.h>
#include <string.h>
#include <signal.h>
#include <stdlib.h>
#include <sys/param.h>
#include <unistd.h>
#include <iostream>

#include "th_assert.h"
#include "libbyz.h"
#include "Statistics.h"
#include "Timer.h"

#include "simple.h"

using std::cerr;

static int exec_count = 0;
static int max_exec_count = 10000; // how many ops to run tests for
static Timer t;

static void dump_profile(int sig) {
 profil(0,0,0,0);

 stats.print_stats();
 
 exit(0);
}

// Service specific functions.
int exec_command(Byz_req *inb, Byz_rep *outb, Byz_buffer *non_det, int client, bool ro) {

  if(exec_count++ == max_exec_count) {
    cerr << "starting timing at " << max_exec_count << " ops\n";
    t.start();
  } else if(exec_count == 2*max_exec_count) {
    cerr << "stopping execution at " << 2*max_exec_count << " ops\n";
    t.stop();
	cerr << "Throughput: " << max_exec_count/t.elapsed() << "\n";
    dump_profile(0);
  }

#ifdef RETS_GRAPH
  int size = *((int*)(inb->contents));
  bzero(outb->contents, size);
  outb->size = size;    
  return 0;
#else
  // A simple service.
  if (inb->contents[0] == 1) {
    th_assert(inb->size == 8, "Invalid request");
    bzero(outb->contents, Simple_size);
    outb->size = Simple_size;    
    return 0;
  }
  
  th_assert((inb->contents[0] == 2 && inb->size == Simple_size) ||
	    (inb->contents[0] == 0 && inb->size == 8), "Invalid request");
  *((long long*)(outb->contents)) = 0;
  outb->size = 8;
  return 0;
#endif
}

int main(int argc, char **argv) {
  // Process command line options.
  char config[PATH_MAX];
  char config_priv[PATH_MAX];
  config[0] = config_priv[0] = 0;

  int opt;
  while ((opt = getopt(argc, argv, "c:p:")) != EOF) {
    switch (opt) {
    case 'c':
      strncpy(config, optarg, PATH_MAX);
      break;
    
    case 'p':
      strncpy(config_priv, optarg, PATH_MAX);
      break;

    default:
      fprintf(stderr, "%s -c config_file -p config_priv_file", argv[0]);
      exit(-1);
    }
  }
 
  if (config[0] == 0) {
    // Try to open default file
    strcpy(config, "./config");
  }

  if (config_priv[0] == 0) {
    // Try to open default file
    char hname[MAXHOSTNAMELEN];
    gethostname(hname, MAXHOSTNAMELEN);
    sprintf(config_priv, "config_private/%s", hname);
  }

  // signal handler to dump profile information.
  struct sigaction act;
  act.sa_handler = dump_profile;
  sigemptyset (&act.sa_mask);
  act.sa_flags = 0;
  sigaction (SIGINT, &act, NULL);
  sigaction (SIGTERM, &act, NULL);

  int mem_size = 205*8192;
  char *mem = (char*)valloc(mem_size);
  bzero(mem, mem_size);

  Byz_init_replica(config, config_priv, mem, mem_size, exec_command, 0, 0);

  stats.zero_stats();

  // Loop executing requests.
  Byz_replica_run();
}


  
