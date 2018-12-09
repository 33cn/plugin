#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/param.h>
#include <unistd.h>
#include <sys/time.h>
#include <resolv.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <rpc/rpc.h>
#include <sys/time.h>

#include "th_assert.h"
#include "libbyz.h"
#include "nfs.h"

extern "C" void FS_init(int sz, int usz);
extern "C" void FS_map(char *file, char **mem, int *sz);

void service_reqs();

int main(int argc, char **argv) {
  // Process command line options.
  char file_system[PATH_MAX];
  int opt;
  while ((opt = getopt(argc, argv, "f:")) != EOF) {
    switch (opt) {
    case 'f':
      strcpy(file_system, optarg);
      break;

    default:
      fprintf(stderr, "%s -f file_system", argv[0]);
      exit(-1);
    }
  }
 
  // Initialize file system
  char *mem;
  int mem_size;
  FS_map(file_system, &mem, &mem_size);
  FS_init(mem_size, 0);

  // Service requests
  service_reqs();
}


extern "C" int nfsd_dispatch(Byz_req *inb, Byz_rep *outb, Byz_buffer *non_det, 
			     int client, int ro);

#define MAX_MSG_SIZE 8192*2

void service_reqs() {
  struct timeval tval;
  Byz_buffer non_det;
  
  non_det.contents = (char*)&tval;
  non_det.size = sizeof(tval);

  Byz_req inb; 
  Byz_rep outb;

  inb.contents = (char*)malloc(MAX_MSG_SIZE);
  inb.size = MAX_MSG_SIZE;

  outb.contents = (char*)malloc(MAX_MSG_SIZE);
  outb.size = MAX_MSG_SIZE;
  
  
  // Create socket and name it.
  int s;
  if ((s = socket(AF_INET, SOCK_DGRAM, 0)) == -1) {
    th_fail("Could not create socket");
  }

  struct sockaddr_in a;
  bzero((char *)&a, sizeof(a));
  a.sin_addr.s_addr = INADDR_ANY;
  a.sin_family = AF_INET;
  a.sin_port = htons(NFSD_PROXY_PORT);
  if (bind(s, (struct sockaddr *) &a, sizeof(a)) == -1) {
    th_fail("Could not bind name to socket");
  }

  // Loop handling messages.
  unsigned addr_len;
  while (1) {
    addr_len = sizeof(a);

    int ret = recvfrom(s, inb.contents, MAX_MSG_SIZE, 0, 
		       (struct sockaddr*)&a, &addr_len);    
    if (ret <= 0) {
      // Error while receiving.
      perror("recvfrom() failed\n");
      continue;
    }

    inb.size = ret;
    outb.size = MAX_MSG_SIZE;

    // Get current time
    gettimeofday(&tval, 0);

    // Handle request
    nfsd_dispatch(&inb, &outb, &non_det, 0, 0);

    // Send reply
    ret = sendto(s, outb.contents, outb.size, 0, (struct sockaddr*)&a, addr_len);
    if (ret <= 0) {
      // Error in sendto
      perror("sendto() failed\n");
      continue;
    }
  }
}

