#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/param.h>
#include <unistd.h>

#include "th_assert.h"
#include "Timer.h"
#include "libbyz.h"

#include "Statistics.h"

#include "simple.h"

const int num_iter = 1000;

int main(int argc, char **argv) {
  char config[PATH_MAX];
  char config_priv[PATH_MAX];
  config[0] = config_priv[0] = 0;
  int option = 0; // null command
  bool read_only = false;
  short port=0; 
  char requests[num_iter];
  char replies[num_iter];
  int pos[num_iter];
  int debug=0;
      

  // Process command line options.
  int opt;
  while ((opt = getopt(argc, argv, "c:p:i:nrxdwom:")) != EOF) {
    switch (opt) {
    case 'i':
      fprintf(stderr,"-i option no longer supported.\n");
      //      num_iter = atoi(optarg);
      break;

    case 'm':
      port = atoi(optarg);
      break;

    case 'n':
      option = 0;
      break;

    case 'r':
      option = 1;
      break;

    case 'w':
      option = 2;
      break;
    
    case 'o':
      read_only = true;
      break;

    case 'c':
      strncpy(config, optarg, PATH_MAX);
      config[PATH_MAX] = 0;
      break;
    
    case 'p':
      strncpy(config_priv, optarg, PATH_MAX);
      config[PATH_MAX] = 0;
      break;

    case 'x':
      option = 4;
      break;

    case 'd':
      debug = 1;
      break;

    
    default:
      fprintf(stderr, "%s -c config_file -p config_priv_file -m port", argv[0]);
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


  srand(0);

  // Initialize client
  Byz_init_client(config, config_priv, port);


  //
  // Loop invoking requests:
  //

  // Allocate request
  Byz_req req;
  Byz_alloc_request(&req, Simple_size);
  th_assert(Simple_size <= req.size, "Request too big");

  // Store data into request
  for (int i=0; i < Simple_size; i++) {
    req.contents[i] = option;
  }
  

  req.size = 8;
  if (option==4)
    req.size = 16;
  if (option==2)
    req.size = 4096;

  stats.zero_stats();

  Timer t;
  t.start();
  Byz_rep rep;
  for (int i=0; i < num_iter; i++) {
    // Invoke request

    if(option==4) {
      int *a=(int*)&(req.contents[4]);
      char *b=&(req.contents[8]);
#ifdef VAR_SIZE
      *a = rand()%(NPAGES/2*(SIZE2+SIZE1));
#else
      *a = rand()%(NPAGES*Block_size);
#endif
      *b = rand();
      requests[i]=*b;
      pos[i]=*a;
    }
      
    Byz_invoke(&req, &rep, read_only);

    if (debug)
      fprintf(stderr,"invoked req %d pos=%d   newcont=%d pedidio=%d\n", i, *((int*)&(req.contents[4])) ,req.contents[8], requests[i]);

    // Check reply
        th_assert(((option == 2 || option == 0) && rep.size == 8) ||
    	    (option == 1 && rep.size == Simple_size) ||
	 (option==4) , "Invalid reply");
    
    // Free reply
    Byz_free_reply(&rep);
  }
  if (option==4) {
    read_only=true;
    req.contents[0] = 3;
    srand(0);
    for (int i=0; i < num_iter; i++) {

      int *c=(int *)&(req.contents[4]);
#ifdef VAR_SIZE
      *c = rand()%(NPAGES/2*(SIZE2+SIZE1)); 
#else
      *c = rand()%(NPAGES*Block_size);
#endif
      rand();
      // Invoke request
      Byz_invoke(&req, &rep, read_only);
      
      
      // Check reply
      if (debug)
	fprintf(stderr,"Executed %d Replied %d\n", i,rep.contents[0]);
      
      replies[i]=rep.contents[0];

      // Free reply
      Byz_free_reply(&rep);
    }
  }
  t.stop();
  printf("Elapsed time %f for %d iterations of operation %d\n", t.elapsed(), 
	 num_iter, option);

  stats.print_stats();

  for (int i=0;i<num_iter;i++)
    if (requests[i]!=replies[i]) {
      option=1;
      for (int j=i+1; j<num_iter; j++)
	if (pos[i]==pos[j])
	  if(requests[i]!=requests[j])
	    option=0;
	  else
	    option=1;
      if (option)
	fprintf(stderr,"mismatch indice %d requests %d replies %d\n",i,requests[i],replies[i]);
    }
    else {
      option=0;
      for (int j=i+1; j<num_iter; j++)
	if (pos[i]==pos[j])
	  if (requests[i]!=requests[j])
	    option=j;
	  else
	    option=0;
      if (option)
	fprintf(stderr,"mismatch (SHOULD BE DIFF'T) indice %d igual a %d - requests %d replies %d\n",i,option,requests[i],replies[i]);
    }

  Byz_free_request(&req);
}
  

