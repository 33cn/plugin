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

int main(int argc, char **argv) {
  char config[PATH_MAX];
  char config_priv[PATH_MAX];
  config[0] = config_priv[0] = 0;
  int num_iter = 1000;
  int option = 0; // null command
  short port=0; 

  // Process command line options.
  int opt;
  while ((opt = getopt(argc, argv, "c:p:i:m:")) != EOF) {
    switch (opt) {
    case 'i':
      num_iter = atoi(optarg);
      break;

    case 'm':
      port = atoi(optarg);
      break;
    
    case 'c':
      strncpy(config, optarg, PATH_MAX);
      config[PATH_MAX] = 0;
      break;
    
    case 'p':
      strncpy(config_priv, optarg, PATH_MAX);
      config[PATH_MAX] = 0;
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


  // Initialize client
  Byz_init_client(config, config_priv, port);


  //
  // Loop invoking requests and measuring latency
  //
  int sizes[] = {8, 64, 128, 256, 512, 1024, 
		 2048, 3072, 4096, 5120, 6144, 7168, 8192};

  // variances
  float vars[sizeof(sizes)/sizeof(int)];

  static const int num_points = 3;

  //stats.zero_stats();
  // Allocate request
  Byz_req req;
  Byz_alloc_request(&req, 8192);

  bool read_only = false;  
  for (int l=0; l < 2; l++) {
    // Get two graphs one with read-only optimization and one without

    printf("Latency (usecs) graph for different arg sizes (bytes)\n");
    printf("Read-only = %s iterations = %d\n", 
	   ((read_only) ? "true" : "false"), num_iter);
    printf("Averages: \n");

    for (int j=0; j < sizeof(sizes)/sizeof(int); j++) {
      int rsize = sizes[j];
    
      // Store data into request
      *((int*)(req.contents)) = rsize;
      for (int i=4; i < rsize; i++) {
	req.contents[i] = option;
      }

#ifdef RETS_GRAPH
      req.size = 8;
#else
      req.size = rsize;
#endif

      float results[num_points];
      for (int k = 0; k < num_points; k++) {
	Timer t;
	t.reset();
	t.start();
	Byz_rep rep;
	for (int i=0; i < num_iter; i++) {
	  // Invoke request
	  Byz_invoke(&req, &rep, read_only);
	
	  // Free reply
	  Byz_free_reply(&rep);
	}
	t.stop();
	results[k] = t.elapsed();
	sleep(2);
      }
	
      // Compute average
      double sum = 0;
      for (int k=0; k < num_points; k++) {
	sum += results[k];
      }
      float avg = sum/num_points * (1000000.0/num_iter);
      printf("%d %f\n", rsize, avg);

      // Compute std
      sum = 0;
      for (int k=0; k < num_points; k++) {
	double val = results[k]*(1000000.0/num_iter) - avg;
	sum += val*val;
      }
      vars[j] =  sum/(float)(num_points-1);
    }
    
    printf("Standard deviations:\n");
    for (int j=0; j < sizeof(sizes)/sizeof(int); j++) {
      printf("%d std = %f std mean = %f\n", 
	     sizes[j],sqrt(vars[j]), sqrt(vars[j])/sqrt(num_points));
    }

    read_only = true;
  }

  Byz_free_request(&req);

  // stats.print_stats();
}
  
