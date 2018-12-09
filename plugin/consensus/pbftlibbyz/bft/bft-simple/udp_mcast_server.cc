#include <stdio.h>
#include <sys/types.h>   /* include files for IP Sockets */
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>

#define TRUE 1
#define BUFSIZE   4096
#define TTL_VALUE 1
#define TEST_ADDR "227.4.1.9"
#define TEST_PORT 3456
#define LOOPMAX 10000
#define WAITTIME 20
int main(){
  struct sockaddr_in stLocal, stTo, stFrom;
  char achIn[BUFSIZE];
  char achOut[] = "1234567";
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
  stLocal.sin_port =     htons(TEST_PORT);
  bind(s, (struct sockaddr*) &stLocal, sizeof(stLocal));
  
  /* join the multicast group. */
  stMreq.imr_multiaddr.s_addr = inet_addr(TEST_ADDR);
  stMreq.imr_interface.s_addr = INADDR_ANY;
  i = setsockopt(s, 
	     IPPROTO_IP, 
	     IP_ADD_MEMBERSHIP, 
	     (char *)&stMreq, 
	     sizeof(stMreq));
  if (i < 0) {
    perror("unable to join group");
    exit(1);
  }

  /* set TTL to traverse up to multiple routers 
  iTmp = TTL_VALUE;
  i = setsockopt(s, 
	     IPPROTO_IP, 
	     IP_MULTICAST_TTL, 
	     (char *)&iTmp, 
	     sizeof(iTmp)); 

  if (i < 0) {
    perror("unable to change TTL value");
    exit(1);
  } */

  /* disable loopback */
  loop = 0;
  i = setsockopt(s, 
	     IPPROTO_IP, 
	     IP_MULTICAST_LOOP, 
	     &loop, 
	     sizeof(loop)); 
  
  if (i < 0) {
    perror("unable to disable loopback");
    exit(1);
  }


  /* assign our destination address */
  stTo.sin_family =      AF_INET;
  stTo.sin_addr.s_addr = inet_addr(TEST_ADDR);
  stTo.sin_port =        htons(TEST_PORT);
  
  while (1) {
    socklen_t addr_size = sizeof(struct sockaddr_in);
    int error;
    int waitime = 0;
    int count;    
    
    error = recvfrom(s, achIn, BUFSIZE, 0,
		     (struct sockaddr*)&stFrom, &addr_size);
    if (error < 0) 
      perror("recvfrom() failed\n");

    error = sendto(s, achOut, sizeof(achOut), 
		   0,(struct sockaddr*)&stFrom, addr_size);
    if (error < 0) {
      perror("sendto() failed\n");
    } 
  }
} /* end main() */  


