#include <stdio.h>
#include <string.h>
#include <sys/time.h>
#include <resolv.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <rpc/rpc.h>
#include <sys/time.h>
#include <arpa/inet.h>

#include "th_assert.h"
#include "types.h"
#include "Timer.h"
#ifdef AUTHENTICATION
#include "simple_auth.h"
#else
#define SIMPLE_MAC_SIZE 0
#endif //AUTHENTICATION


#define SIMPLE_PORT 3460

int main(int argc, char **argv) {
  // Process command line options.
  char *server_addr = 0;
  int num_iter = 1000;
  int port = SIMPLE_PORT;
  int option = 0; // null command

  int opt;
  while ((opt = getopt(argc, argv, "s:i:nrwp:")) != EOF) {
    switch (opt) {
    case 's':
      server_addr = optarg;
      break;

    case 'i':
      num_iter = atoi(optarg);
      break;

    case 'p':
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

    default:
      fprintf(stderr, "%s -f file_system -s server_addr", argv[0]);
      exit(-1);
    }
  }

  if (server_addr == 0) {
    server_addr = "18.26.1.241"; // sarod
  }

  char in[4096+SIMPLE_MAC_SIZE];
  char out[4096+SIMPLE_MAC_SIZE];

  // Create socket and name them.
  int dest;
  if ((dest = socket(AF_INET, SOCK_DGRAM, 0)) == -1) {
    th_fail("Could not create socket");
  }

  Address a;
  bzero((char *)&a, sizeof(a));
  a.sin_addr.s_addr = INADDR_ANY;
  a.sin_family = AF_INET;
  a.sin_port = htons(port);
  if (bind(dest, (struct sockaddr *) &a, sizeof(a)) == -1) {
    th_fail("Could not bind name to socket");
  }

  // Fill-in server address.
  Address desta;
  bzero((char *)&desta, sizeof(desta));
  desta.sin_addr.s_addr = inet_addr(server_addr);
  desta.sin_family = AF_INET;
  desta.sin_port = htons(SIMPLE_PORT);

  // Fill out buffer with option.
  for (int i=0; i < 4096; i++) {
    out[i] = option;
  }
  
  Timer t;
  t.start();

  for (int i=0; i < num_iter; i++) {
    int len = (option == 2) ? 4096 : 8;

#ifdef AUTHENTICATION
    gen_mac(out, len, out+len); 
#endif //AUTHENTICATION

    int ret = sendto(dest, out, len+SIMPLE_MAC_SIZE, 
		 0, (struct sockaddr*)&desta, sizeof(desta));
    if (ret <= 0) {
      // Error in sendto
      perror("sendto() failed\n");
      continue;
    }


    // Wait for reply
    ret = recvfrom(dest, in, 4096+SIMPLE_MAC_SIZE, 0, 0, 0);    
    if (ret <= 0) {
      // Error while receiving.
      perror("recvfrom() failed\n");
      continue;
    }

#ifdef AUTHENTICATION
    if (ret < SIMPLE_MAC_SIZE || 
	!verify_mac(in, ret-SIMPLE_MAC_SIZE, in+ret-SIMPLE_MAC_SIZE)) {
      th_fail("Could not authenticate");
    }
#endif //AUTHENTICATION
    /*
    if (i%1000 == 0) {
      printf("%d operations complete", i);
    }
    */
    th_assert(((option == 2 || option == 0) && ret == 8+SIMPLE_MAC_SIZE) ||
	    (option == 1 && ret == 4096+SIMPLE_MAC_SIZE), "Invalid reply");
  }
  t.stop();
  printf("Elapsed time %f for %d iterations of operation %d.\n", t.elapsed(), 
	 num_iter, option);

}


