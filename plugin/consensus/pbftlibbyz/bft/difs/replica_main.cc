#include <stdio.h>
#include <string.h>
#include <sys/time.h>
#include <rpc/rpc.h>
#include <signal.h>
#include <unistd.h>

#include "th_assert.h"
#include "libbyz.h"
#include "Statistics.h"

// thresholds to accept timestamps from primary (in seconds)
#define LOWER_THRESHOLD 10
#define UPPER_THRESHOLD 10


// Service specific functions.

extern "C" int nfsd_dispatch(Byz_req *inb, Byz_rep *outb, Byz_buffer *non_det, 
			     int client, int ro);

extern "C" int get_obj(int n, char **obj);
extern "C" void put_objs(int num_objs, int *sizes, int *indices, char **objs);

extern "C" void shutdown_state(FILE *o);
extern "C" void restart_state(FILE *i);


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

bool check_non_det(Byz_buffer *non_det) {
  return TRUE;
  th_assert(non_det->size >= (int)sizeof(struct timeval), "Non-det buffer is too small");
  struct timeval t;
  gettimeofday(&t, 0);


  if ( (((struct timeval *)(non_det->contents))->tv_sec > 
	t.tv_sec + (-LOWER_THRESHOLD)) &&
       (((struct timeval *)(non_det->contents))->tv_sec 
	< t.tv_sec + UPPER_THRESHOLD) )
    return TRUE;

  fprintf(stderr, "\t\t> > > > Bad non-deterministic choices < < < <\n");
  return FALSE;
}

extern "C" int FS_init(char *, char *);

#ifdef STATS
extern int num_get_file;
extern int num_get_dir;
extern int num_get_free;
extern int num_put;
extern int num_put_file;
extern int num_put_dir;
extern int num_put_free;
extern int numcalls[];
#endif

static void dump_profile(int sig) {

#ifdef STATS
  show_stats(); // temporario - retirarXXX RR-TODO!!
  printf("# gets for: files = %d dirs = %d free = %d\n", num_get_file, num_get_dir, num_get_free);
  printf("# puts = %d.  For: files = %d dirs = %d free = %d\n", num_put, num_put_file, num_put_dir, num_put_free);
  for (int i=0; i<NUM_CALLS; i++)
    printf("#of %d calls %5d, ",i, numcalls[i]);
  fprintf(stderr, "\nLookups cached %d. Reads cached %d.\n", lookup_cached, read_cached);
#endif

 profil(0,0,0,0);

 stats.print_stats();

 exit(0);
}

char hostname[MAXHOSTNAMELEN+1];

int main(int argc, char **argv) {
  // Process command line options.
  char config[PATH_MAX];
  char config_priv[PATH_MAX];
  char dirname[PATH_MAX];
  int port = 0;
  hostname[0] = dirname[0] = config[0] = config_priv[0] = 0;


  // Process command line options.
  int opt;
  while ((opt = getopt(argc, argv, "c:p:h:d:r:")) != EOF) {
    switch (opt) {
    case 'h':
      strncpy(hostname, optarg, MAXHOSTNAMELEN+1);
      break;

    case 'd':
      strncpy(dirname, optarg, PATH_MAX);
      break;

    case 'p':
      strncpy(config_priv, optarg, PATH_MAX);
      break;

    case 'c':
      strncpy(config, optarg, PATH_MAX);
      break;

    case 'r':
      port = atoi(optarg);
      break;
      
      
    default:
      fprintf(stderr, "%s -c config_file -p config_priv_file -d dir -h hoste", argv[0]);
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
    if (gethostname(hname, MAXHOSTNAMELEN) < 0)
      sysfail("hafs replica: could not find local host name");
    sprintf(config_priv, "config_private/%s", hname);
  }

  if (dirname[0] == 0) {
    // Try to open default dir
    strcpy(dirname, "/tmp/hafs");
  }

  if (hostname[0] == 0) {
    // default hostname
    // Get my host name
    if (gethostname(hostname, MAXHOSTNAMELEN+1) < 0)
      sysfail("hafs: could not find local host name");
  }

  // signal handler to dump profile information and stats.
  struct sigaction act;
  act.sa_handler = dump_profile;
  sigemptyset (&act.sa_mask);
  act.sa_flags = 0;
  sigaction (SIGINT, &act, NULL);
  sigaction (SIGTERM, &act, NULL);


  // Initialize file system
  int npages = FS_init(hostname, dirname);

  if (npages < 0) {
    fprintf(stderr, "Error in FS_init. Check if I've been hit by a cosmic ray\n");
    exit(-1);
  }
  fprintf(stderr, "Init on port %d\n", port);
  // Initialize replica
  if (Byz_init_replica(config, config_priv, npages,
		       nfsd_dispatch1, non_det_choices, sizeof(struct timeval),
		       check_non_det, get_obj, put_objs,
		       shutdown_state, restart_state, port) < 0) {
    fprintf(stderr, "Error in Byz_init. Check if I've been hit by a cosmic ray\n");
    exit(-1);
  }

  stats.zero_stats();

  // Loop executing requests.
  Byz_replica_run();
}


  
