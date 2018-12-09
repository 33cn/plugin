#include <stdio.h>
#include <string.h>
#include <sys/time.h>
#include <rpc/rpc.h>
#include <signal.h>
#include <unistd.h>

#include "th_assert.h"
#include "libbyz.h"
#include "Statistics.h"


// Service specific functions.

extern "C" int nfsd_dispatch(Byz_req *inb, Byz_rep *outb, Byz_buffer *non_det, 
			     int client, int ro);

int nfsd_dispatch1(Byz_req *inb, Byz_rep *outb, Byz_buffer *non_det, 
		   int client, bool ro) {
  // This wrapper solves a problem that happens on linux with optimization.
  // nfsd_dispatch1 has C++ linkage as the replication library expects.
  return nfsd_dispatch(inb, outb, non_det, client, (ro) ? 1 : 0);
}


void non_det_choices(Seqno s, Byz_buffer *ndet) {
  th_assert(ndet->size >= (int)sizeof(struct timeval), "Non-det buffer is too small");
  ndet->size = sizeof(struct timeval);
  gettimeofday((struct timeval *)ndet->contents, 0);
}

extern "C" void FS_init(int sz, int usz);
extern "C" void FS_map(char *file, char **mem, int *sz);

extern struct rpc_msg *msg_buf;

static void dump_profile(int sig) {
 profil(0,0,0,0);

 stats.print_stats();

 exit(0);
}


int main(int argc, char **argv) {
  // Process command line options.
  char config[PATH_MAX];
  char config_priv[PATH_MAX];
  char file_system[PATH_MAX];
  file_system[0] = config[0] = config_priv[0] = 0;


  // Process command line options.
  int opt;
  while ((opt = getopt(argc, argv, "c:p:f:")) != EOF) {
    switch (opt) {
    case 'f':
      strncpy(file_system, optarg, PATH_MAX);
      file_system[PATH_MAX-1] = 0;
      break;

    case 'p':
      strncpy(config_priv, optarg, PATH_MAX);
      config_priv[PATH_MAX-1] = 0;
      break;

    case 'c':
      strncpy(config, optarg, PATH_MAX);
      config[PATH_MAX-1] = 0;
      break;
      
    default:
      fprintf(stderr, "%s -c config_file -p config_priv_file -f file_system", argv[0]);
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

  if (file_system[0] == 0) {
    // Try to open default file system
    strcpy(file_system, "./fs");
  }

  // signal handler to dump profile information and stats.
  struct sigaction act;
  act.sa_handler = dump_profile;
  sigemptyset (&act.sa_mask);
  act.sa_flags = 0;
  sigaction (SIGINT, &act, NULL);
  sigaction (SIGTERM, &act, NULL);


  // Initialize file system
  char *mem;
  int mem_size;
  FS_map(file_system, &mem, &mem_size);

  // Initialize replica
  int used_bytes = Byz_init_replica(config, config_priv, mem, mem_size, 
		   nfsd_dispatch1, non_det_choices, sizeof(struct timeval));

  if (used_bytes < 0)
    exit(-1);

  // Can only be called here because it requires replica to be initialized.
  FS_init(mem_size, used_bytes);

  stats.zero_stats();

  // Loop executing requests.
  Byz_replica_run();
}


  
