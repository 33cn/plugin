/* Mounts a user-level NFS daemon */

#define _SOCKADDR_LEN
#include "nfsconf.h"
#include <netdb.h>
#include <sys/mount.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <rpc/rpc.h>
#include <arpa/inet.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <errno.h>
#include <ctype.h>
#include <unistd.h>
#include "nfs.h"

#include "th_assert.h"

static void init_args(struct nfs_args*, struct sockaddr_in*, char*, int cache);

int main(int argc, char **argv) {
  char path[PATH_MAX];
  char hostname[PATH_MAX];
  int to_cache;
  struct hostent *h;
  struct nfs_args args;
  struct sockaddr_in client;

  // defaults:
  to_cache = 1;                     // cache by default
  strcpy(hostname, "127.0.0.1");    // mount in loopback interface
  strcpy(path, "./mnt");        
    
  int opt;
  while ((opt = getopt(argc, argv, "d:h:c:")) != EOF) {
    switch (opt) {
    case 'd':
      // directory to mount
      strncpy(path, optarg, PATH_MAX);
      path[PATH_MAX-1] = 0;
      break;
     
    case 'h':
      strncpy(hostname, optarg, PATH_MAX);
      hostname[PATH_MAX-1] = 0;
      break;
     
    case 'c':
      to_cache = atoi(optarg);
      break;
     
    default:
      fprintf(stderr, "%s [-d mnt_path] [-h hostname] [-c 0|1]", argv[0]);
      exit(-1);
    }
  }

  fprintf(stderr, "Mounting port %d on %s.\n", NFSD_PROXY_PORT, path);

  bzero((char *) &client, sizeof(client));
  client.sin_family = AF_INET;
  client.sin_port = htons(NFSD_PROXY_PORT);
  if (!isdigit(hostname[0])) {
    h = gethostbyname(hostname);
    if (!h) {
      fprintf(stderr, "Unknown host: %s\n", hostname);
      exit(-1);
    } else {
      memcpy(&client.sin_addr, h->h_addr, h->h_length);
    }
  } else {
    client.sin_addr.s_addr = inet_addr(hostname);
  }
  
  init_args(&args, &client, hostname, to_cache);
  
  if (SYS_MOUNT(hostname, MOUNT_NFS, path, 0, (caddr_t) &args) < 0) {
    perror("mount");
    return 1;
  }
  
#ifndef __linux__
  free(args.fh);
#endif

  return 0;
}

static void init_args(struct nfs_args* args, struct sockaddr_in* addr,
		      char *hostname, int to_cache) {

  //
  //  The root directory is  (0, 0).
  //
  caddr_t fh = (caddr_t)malloc(NFS_FHSIZE);
  bzero((char *) fh, NFS_FHSIZE);
  unsigned int *datap = (unsigned int *) fh;
  datap[0] = 0;
  datap[1] = 0;

  bzero((char *) args, sizeof(*args));
#ifndef __linux__
  args->addr          = addr;
  args->fh            = fh;
  args->flags         = NFSMNT_SOFT | NFSMNT_HOSTNAME;
  args->hostname      = hostname;
  args->netname       = "lcs.mit.edu";
  args->pathconf      = 0;
#else
  args->addr		= *addr;
  memcpy (&args->root, fh, NFS_FHSIZE);
  args->flags		= NFSMNT_SOFT;
  args->fd = socket (AF_INET, SOCK_DGRAM, 0);
  if (args->fd < 0)
     th_fail("could not create socket");
 
  sockaddr_in tmp;
  tmp.sin_family =   AF_INET;
  tmp.sin_addr.s_addr = INADDR_ANY;
  tmp.sin_port = htons(NFSD_PROXY_PORT+5);
  int error = bind(args->fd, (struct sockaddr*)&tmp, sizeof(sockaddr_in));
  if (error < 0) 
    th_fail("could not bind socket");
  
  error = connect(args->fd, (sockaddr *) addr, sizeof (sockaddr_in));
  if (error < 0) 
    th_fail("could not connect socket");
 
  args->version = 2;
  memcpy(args->hostname, hostname, strlen(hostname)+1);
  args->namlen = strlen(hostname);
#endif

  if (to_cache == 0) {
    // Attribute caching is disabled unless explicitly requested
    args->flags |= NFSMNT_NOAC;
  }

  args->wsize		= 4096;
  args->rsize		= 4096;
  args->timeo		= 11;
  args->retrans	        = 4;  
  
  if (to_cache == 0) {
    args->acregmin = 0;
    args->acregmax = 0;
    args->acdirmin = 0;
    args->acdirmax = 0;
  } else {
    args->acdirmin = 30;
    args->acdirmax = 60;
    args->acregmin = 3;
    args->acregmax = 60;
  }  
}
