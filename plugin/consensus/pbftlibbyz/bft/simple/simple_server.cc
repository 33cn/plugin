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
#include "Timer.h"
#ifdef AUTHENTICATION
#include "simple_auth.h"
#else
#define SIMPLE_MAC_SIZE 0
#endif //AUTHENTICATION

#define SIMPLE_PORT 3460

int main(int argc, char **argv) {
  char in[8192+SIMPLE_MAC_SIZE];
  char out[8192+SIMPLE_MAC_SIZE];

  // Create sockets and name them.
  int source, dest;
  if ((source = socket(AF_INET, SOCK_DGRAM, 0)) == -1) {
    th_fail("Could not create socket");
  }

  struct sockaddr_in a;
  bzero((char *)&a, sizeof(a));
  a.sin_addr.s_addr = INADDR_ANY;
  a.sin_family = AF_INET;
  a.sin_port = htons(SIMPLE_PORT);
  if (bind(source, (struct sockaddr *) &a, sizeof(a)) == -1) {
    th_fail("Could not bind name to socket");
  }

  // Loop handling messages.
  int addr_len;
  while (1) {
    addr_len = sizeof(a);

    int ret = recvfrom(source, in, 8192+SIMPLE_MAC_SIZE, 0, 
		       (struct sockaddr*)&a, &addr_len);    
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

    if (in[0] == 1) {
      th_assert(ret == 8+SIMPLE_MAC_SIZE, "Invalid request");
      bzero(out, 4096);
      ret = 4096;
    } else {
      th_assert((in[0] == 2 && ret == 4096+SIMPLE_MAC_SIZE) ||
		(in[0] == 0 && ret == 8+SIMPLE_MAC_SIZE), "Invalid request");
      out[0] = 0;
      ret = 8;
    }

    // Send reply
#ifdef AUTHENTICATION
    gen_mac(out, ret, out+ret); 
#endif //AUTHENTICATION

    ret = sendto(source, out, ret+SIMPLE_MAC_SIZE, 0, 
		 (struct sockaddr*)&a, addr_len);
    if (ret <= 0) {
      // Error in sendto
      perror("sendto() failed\n");
      continue;
    }
  }
}
