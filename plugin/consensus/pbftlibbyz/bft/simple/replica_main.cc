#include <stdio.h>
#include <string.h>
#include <signal.h>
#include <stdlib.h>
#include <sys/param.h>
#include <unistd.h>

#include "th_assert.h"
#include "libbyz.h"
#include "Statistics.h"

#include "simple.h"


int mem_size;
char *mem;
int used_mem;
int first_read=1;
int debug=0;

int get_segment(int n, char **page)
{
  int size;

#ifndef VAR_SIZE
  size = Block_size;
#else
  size = (n%2 ? SIZE2 : SIZE1);
#endif

  if (n<0 || n>NPAGES-1)
    th_fail("requested invalid page number\n");
  *page = new char[size];
#ifndef VAR_SIZE
  memcpy(*page, mem+n*size, size);
#else
  memcpy(*page, mem+(n/2)*(SIZE2+SIZE1)+(n%2 ? SIZE1 : 0), size);
#endif
  return size;
}
  
void put_segments(int count, int *sizes, int *indices, char **pages)
{
  for (int i=0; i<count; i++) { 
#ifndef VAR_SIZE
    if (indices[i]<0 || indices[i]>NPAGES-1 || sizes[i]!= Block_size)
      th_fail("requested invalid page number\n");
    //    fprintf(stderr, "Put for page %d size %d\n", indices[i], sizes[i]);
    memcpy(mem+indices[i]*Block_size, pages[i], Block_size);
#else
    th_assert(sizes[i] == (indices[i]%2 ? SIZE2 : SIZE1), "Wrong size in put");
    //    fprintf(stderr, "Put for page %d size %d\n", indices[i], sizes[i]);
    memcpy(mem+(indices[i]/2)*(SIZE2+SIZE1)+(indices[i]%2 ? SIZE1 : 0), pages[i],
	   sizes[i]);
#endif
  }
}

bool check_nond(Byz_buffer *b)
{
  return true;
}



// Service specific functions.
int exec_command(Byz_req *inb, Byz_rep *outb, Byz_buffer *non_det, int client, bool ro) {


  int n=0, page;
  char newcont;


  // A simple service.
  if (inb->contents[0] == 1) {
    th_assert(inb->size == 8, "Invalid request");
    bzero(outb->contents, 4096);
    outb->size = 4096;    
    return 0;
  }


  if (inb->contents[0] == 4) {

    n=*((int *)&(inb->contents[4]));
#ifndef VAR_SIZE
    page=n/Block_size;
#else
    page = (n / (SIZE2 + SIZE1)) * 2 + ( (n%(SIZE2 + SIZE1) < SIZE1) ? 0 : 1);
#endif

    newcont=inb->contents[8];

    if (debug)
      fprintf(stderr,">modifying ... pos %d page %d New contents = %d\n", n, page, newcont);
    Byz_modify(1,&page);

    mem[n]=newcont;

    bzero(outb->contents, 4096);
    outb->size = 4096;    
    return 0;
  }


  if (inb->contents[0] == 3) {

    n=*((int *)&(inb->contents[4]));

    newcont=(char)rand();

    if (debug)
      fprintf(stderr,"reading ... pos %d  contents %d\n", n, mem[n]);

    bzero(outb->contents, 8);
    outb->contents[0] = mem[n];
    outb->size = 8;    
    return 0;
  }

  
  th_assert((inb->contents[0] == 2 && inb->size == Simple_size) ||
	    (inb->contents[0] == 0 && inb->size == 8), "Invalid request");
  *((long long*)(outb->contents)) = 0;
  outb->size = 8;
  return 0;
}

static void dump_profile(int sig) {
 profil(0,0,0,0);

 stats.print_stats();
 
 exit(0);
}


int main(int argc, char **argv) {
  // Process command line options.
  char config[PATH_MAX];
  char config_priv[PATH_MAX];
  config[0] = config_priv[0] = 0;
  short port = 0;

  srand(0);

  int opt;
  while ((opt = getopt(argc, argv, "c:p:r:d")) != EOF) {
    switch (opt) {
    case 'c':
      strncpy(config, optarg, PATH_MAX);
      break;
    
    case 'p':
      strncpy(config_priv, optarg, PATH_MAX);
      break;

    case 'r':
      port = (short)atoi(optarg);
      break;

    case 'd':
      debug=1;
      break;      

    default:
      fprintf(stderr, "%s [-c config_file] [-p config_priv_file] [-r port] [-d]", argv[0]);
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


#ifndef VAR_SIZE
  mem_size = NPAGES*Block_size;
#else
  assert(NPAGES%2==0);
  mem_size = (NPAGES/2)*(SIZE2+SIZE1);
#endif
  mem = (char*)valloc(mem_size);
  bzero(mem, mem_size);

  //  printf("port %d\n",port);

  used_mem=Byz_init_replica(config, config_priv, NPAGES, exec_command, 0, 0,
		            check_nond, get_segment, put_segments, 0, 0, port);

  stats.zero_stats();

  // Loop executing requests.
  Byz_replica_run();
}

