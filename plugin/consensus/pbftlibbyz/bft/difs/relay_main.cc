#include <stdio.h>
#include <string.h>
#include <sys/time.h>
#include <resolv.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <rpc/rpc.h>
#include <sys/time.h>
#include <arpa/inet.h>
#include <unistd.h>

#include "th_assert.h"
#include "nfs.h"

void service_reqs(char *saddr, bool multi);

int main(int argc, char **argv) {
  // Process command line options.
  char *server_addr = 0;
  bool multi = false;

  int opt;
  while ((opt = getopt(argc, argv, "s:m")) != EOF) {
    switch (opt) {
    case 's':
      server_addr = optarg;
      break;

    case 'm':
      multi = true;
      break;

    default:
      fprintf(stderr, "%s -s server_addr [-m] ", argv[0]);
      exit(-1);
    }
  }

  if (server_addr == 0) {
    server_addr = "18.26.1.241"; // sarod
  }
    
  // Service requests
  service_reqs(server_addr, multi);
}

#define MAX_MSG_SIZE 8192

void service_reqs(char *saddr, bool multi) {
  char *in = (char*)malloc(MAX_MSG_SIZE);
  char *out = (char*)malloc(MAX_MSG_SIZE);

  // Create sockets and name them.
  int source, dest;
  if ((source = socket(AF_INET, SOCK_DGRAM, 0)) == -1) {
    th_fail("Could not create socket");
  }

  struct sockaddr_in a;
  bzero((char *)&a, sizeof(a));
  a.sin_addr.s_addr = inet_addr("127.0.0.1");
  a.sin_family = AF_INET;
  a.sin_port = htons(NFSD_PROXY_PORT);
  if (bind(source, (struct sockaddr *) &a, sizeof(a)) == -1) {
    th_fail("Could not bind name to socket");
  }

  pid_t pid1 = 0;
  if (multi) {
    // Fork child.
    pid1 = fork();
    if (pid1 == -1) {
      th_fail("Fork failed");
    }
  }
  
  if ((dest = socket(AF_INET, SOCK_DGRAM, 0)) == -1) {
    th_fail("Could not create socket");
  }
  bzero((char *)&a, sizeof(a));
  a.sin_addr.s_addr = INADDR_ANY;
  a.sin_family = AF_INET;
  a.sin_port = htons(NFSD_PROXY_PORT+((pid1 == 0) ? 1 : 2));
  if (bind(dest, (struct sockaddr *) &a, sizeof(a)) == -1) {
    th_fail("Could not bind name to socket");
  }

  // Fill-in server address.
  struct sockaddr_in desta;
  bzero((char *)&desta, sizeof(desta));
  desta.sin_addr.s_addr = inet_addr(saddr);
  desta.sin_family = AF_INET;
  desta.sin_port = htons(NFSD_PROXY_PORT);


  printf("Ready %d\n", getpid());

  // Loop handling messages.
  unsigned addr_len;
  while (1) {
    addr_len = sizeof(a);
    int ret = recvfrom(source, in, MAX_MSG_SIZE, 0, 
		       (struct sockaddr*)&a, &addr_len);    
    if (ret <= 0) {
      // Error while receiving.
      continue;
    }

    unsigned sent_xid = ((struct rpc_msg *)in)->rm_xid;
    int req_size = ret;
    bool to_send = true;
    while (1) {
      if (to_send) {
	// Relay message.
	ret = sendto(dest, in, req_size, 0, 
		     (struct sockaddr*)&desta, sizeof(desta));
	if (ret <= 0) {
	  // Error in sendto
	  continue;
	}
      }
    
      // Wait for reply.
      struct timeval timeout;
      timeout.tv_sec = 0;
      timeout.tv_usec = 200000;
      fd_set fdset;
      FD_ZERO(&fdset);
      FD_SET(dest, &fdset);
      ret = select(dest+1, &fdset, 0, 0, &timeout); 
      if (ret <= 0 || !FD_ISSET(dest, &fdset)) {
	// Timeout before receiving reply
	to_send = true;
	continue;
      }

      ret = recvfrom(dest, out, MAX_MSG_SIZE, 0, 0, 0); 
      if (ret <= 0 || ((struct rpc_msg *)out)->rm_xid != sent_xid) {
	// Error while receiving.
	to_send = false;
	continue;
      } 
      
      break;
    }

    // Send reply
    ret = sendto(source, out, ret, 0, (struct sockaddr*)&a, addr_len);
    if (ret <= 0) {
      // Error in sendto
      continue;
    }
  }
}
