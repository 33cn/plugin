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

#define SIMPLE_PORT 3460

int main(int argc, char **argv) {
  short port=0; 

  // Process command line options.
  int opt;
  while ((opt = getopt(argc, argv, "p:")) != EOF) {
    switch (opt) {
    case 'p':
      port = atoi(optarg);
      break;
    }
  }

  char server_addr[] = "18.26.1.241"; // sarod

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
  out.cid = 0;
  if (sendto(manager, &out, sizeof(out),
	     0, (struct sockaddr*)&desta, sizeof(desta)) <= 0){
    exit(-1);
  }

  char inb[4096];
  char outb[4096];

  // Create socket and name it.
  int dest;
  if ((dest = socket(AF_INET, SOCK_DGRAM, 0)) == -1) {
    th_fail("Could not create socket");
  }

  bzero((char *)&a, sizeof(a));
  a.sin_addr.s_addr = INADDR_ANY;
  a.sin_family = AF_INET;
  a.sin_port = htons(port);
  if (bind(dest, (struct sockaddr *) &a, sizeof(a)) == -1) {
    th_fail("Could not bind name to socket");
  }

  // Fill-in server address.
  Address destb;
  bzero((char *)&destb, sizeof(destb));
  destb.sin_addr.s_addr = inet_addr(server_addr);
  destb.sin_family = AF_INET;
  destb.sin_port = htons(SIMPLE_PORT);
  
 
  int num_iter;
  int option;
  bool read_only;
  int cid;

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
      cid = in.cid;
    }

    //    printf("%d starting num_iter = %d op = %d read_only = %d\n",
    // cid, num_iter, option, (int)read_only);

    //
    // Loop invoking requests:
    //

    int len;    
    if (option != 2) {
      len = 8;
    } else {
      len = Simple_size;
    }

    // Fill out buffer with option.
    for (int i=0; i < 4096; i++) {
      outb[i] = option;
    }


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
	
      for (int i=0; i < num_iter; i++) {
	int ret = sendto(dest, outb, len, 
			 0, (struct sockaddr*)&destb, sizeof(destb));
	if (ret <= 0) {
	  // Error in sendto
	  perror("sendto() failed\n");
	  continue;
	}


	// Wait for reply
	ret = recvfrom(dest, inb, 4096, 0, 0, 0);    
	if (ret <= 0) {
	  // Error while receiving.
	  perror("recvfrom() failed\n");
	  continue;
	}
      }

      if (k == 1)      
	t.stop();
    }

    out.tag = thr_done;
    out.cid = cid;
    out.elapsed = t.elapsed();
    if (sendto(manager, &out, sizeof(out),
	     0, (struct sockaddr*)&desta, sizeof(desta)) <= 0){
      exit(-1);
    }
  }
}
  
