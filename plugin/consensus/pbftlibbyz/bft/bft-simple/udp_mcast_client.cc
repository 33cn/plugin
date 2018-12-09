#include <stdio.h>
#include <sys/types.h>   /* include files for IP Sockets */
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>

#include "Timer.h"
#include "Cycle_counter.h"

#define TRUE 1
#define BUFSIZE 4096
#define TTL_VALUE 0
#define TEST_ADDR "227.4.1.9"
#define TEST_PORT 3456
#define LOOPMAX 1000
#define WAITTIME 20
int main(){
  struct sockaddr_in stLocal, stTo, stFrom;
  char achIn[BUFSIZE];
  char achOut[BUFSIZE];

  int s;
  struct ip_mreq stMreq;
  int iTmp, i;
  u_char loop;

  /* get a datagram socket */
  s = socket(AF_INET, SOCK_DGRAM, 0);
  
  /* avoid EADDRINUSE error on bind() */ 
  iTmp = TRUE;
  setsockopt(s, SOL_SOCKET, SO_REUSEADDR, 
	     (char *)&iTmp, sizeof(iTmp));
  
  /* name the socket */
  stLocal.sin_family =   AF_INET;
  stLocal.sin_addr.s_addr = htonl(INADDR_ANY);
  stLocal.sin_port =   0;
  bind(s, (struct sockaddr*) &stLocal, sizeof(stLocal));

  /* set TTL to traverse up to multiple routers */
  iTmp = TTL_VALUE;
  i = setsockopt(s, 
		 IPPROTO_IP, 
		 IP_MULTICAST_TTL, 
		 (char *)&iTmp, 
		 sizeof(iTmp)); 
  
  if (i < 0) {
    perror("unable to change TTL value");
    exit(1);
  }

#if 0 
  iTmp = TRUE;
  i = setsockopt(s, SOL_SOCKET, SO_BROADCAST, 
	     (char *)&iTmp, sizeof(iTmp));
   if (i < 0) {
    perror("unable to set bcast");
    exit(1);
  }
#endif 
  /* assign our destination address */
  stTo.sin_family =      AF_INET;
  stTo.sin_addr.s_addr = inet_addr(TEST_ADDR);
  stTo.sin_port =        htons(TEST_PORT);
  
  Cycle_counter c;
  Timer t;
  t.reset();
  t.start();
  for (i=0;i<LOOPMAX;i++) {
    //c.reset();
    //c.start();
    socklen_t addr_size = sizeof(struct sockaddr_in);
    int error;
    int waitime = 0;
    int count;
    int j;

    
#if 0
    error = sendto(s, achOut, BUFSIZE, 
		   0,(struct sockaddr*)&stTo, addr_size);
#else
    error = sendto(s, achOut, 4096, 
		   0,(struct sockaddr*)&stTo, addr_size);
#endif
    if (error < 0) {
      perror("sendto() failed\n");
    } 
    
    for (j=0; j < 1; j++) {
      error = recvfrom(s, achIn, BUFSIZE, 0,
		       (struct sockaddr*)&stFrom, &addr_size);
      if (error < 0) {
	perror("recvfrom() failed\n");
      } 
    }
    //c.stop();
    //printf("%qd\n", c.elapsed());
  }
  t.stop();
  printf("Elapsed time = %f\n", t.elapsed());
} /* end main() */  


