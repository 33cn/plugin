#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <signal.h>
#include <sys/param.h>
#include <unistd.h>
#include <arpa/inet.h>

#include "th_assert.h"
#include "Timer.h"
#include "libbyz.h"
#include "Client.h"

#include "Statistics.h"

#include "thr_graph.h"
#include "simple.h"

int main(int argc, char **argv) {
  char config[PATH_MAX];
  char config_priv[PATH_MAX];
  config[0] = config_priv[0] = 0;
  short port=0; 

  // Process command line options.
  int opt;
  while ((opt = getopt(argc, argv, "c:p:m:")) != EOF) {
    switch (opt) {
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

  // Create socket to communicate with manager
  int manager;
  if ((manager = socket(AF_INET, SOCK_DGRAM, 0)) == -1) {
    th_fail("Could not create socket");
  }

  Address a;
  bzero((char *)&a, sizeof(a));
  a.sin_addr.s_addr = INADDR_ANY;
  a.sin_family = AF_INET;
  a.sin_port = htons(port+500);
  if (bind(manager, (struct sockaddr *) &a, sizeof(a)) == -1) {
    th_fail("Could not bind name to socket");
  }

  // Fill-in manager address.
  Address desta;
  bzero((char *)&desta, sizeof(desta));
  desta.sin_addr.s_addr = inet_addr("18.26.1.35"); //cello
  desta.sin_family = AF_INET;
  desta.sin_port = htons(3456);

  thr_command out, in;

  // Tell manager we are up
  out.tag = thr_up;
  out.cid = node->id();
  if (sendto(manager, &out, sizeof(out),
	     0, (struct sockaddr*)&desta, sizeof(desta)) <= 0){
    exit(-1);
  }

  int num_iter;
  int option;
  bool read_only;

  // Allocate request
  Byz_req req;
  Byz_alloc_request(&req, Simple_size);
  th_assert(Simple_size <= req.size, "Request too big");
    
  while (1) {
    // Wait for a command from the manager
    int ret = recvfrom(manager, &in, sizeof(in), 0, 0, 0);    
    if (ret != sizeof(in)) {
      exit(-1);
    }

    if (in.tag == thr_end) {
      exit(0);
    } else {
      num_iter = in.num_iter;
      option = in.op;
      read_only = (in.read_only != 0);
    }

    // printf("%d starting num_iter = %d op = %d read_only = %d\n",
    //   node->id(), num_iter, option, (int)read_only);
    //
    // Loop invoking requests:
    //
    
    // Store data into request
    for (int i=0; i < Simple_size; i++) {
      req.contents[i] = option;
    }
    
    if (option != 2) {
      req.size = 8;
    } else {
      req.size = Simple_size;
    }

    Byz_reset_client();

    Timer t;
    for (int k=0; k < 3; k++) {
      int niter;
  
      if (k == 1) {
	niter = num_iter;
	t.reset();
	t.start();
      } else {
	niter = num_iter;
	//	niter = num_iter/2;
      }
	
      Byz_rep rep;
      for (int i=0; i < niter; i++) {
	// Invoke request
	Byz_invoke(&req, &rep, read_only);
	
	// Check reply
	th_assert(((option == 2 || option == 0) && rep.size == 8) ||
		  (option == 1 && rep.size == Simple_size), "Invalid reply");
      
	// Free reply
	Byz_free_reply(&rep);
      }

      if (k == 1)      
	t.stop();
    }

    out.tag = thr_done;
    out.cid = node->id();
    out.elapsed = t.elapsed();
    if (sendto(manager, &out, sizeof(out),
	     0, (struct sockaddr*)&desta, sizeof(desta)) <= 0){
      exit(-1);
    }
  }
    
  Byz_free_request(&req);
}
  
