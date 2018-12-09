//
// NFSD proxy on client machine.
//

#define _SOCKADDR_LEN
#include <stdio.h>
#include <stdlib.h>
#include <signal.h>
#include <resolv.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <rpc/rpc.h>
#include <rpc/pmap_clnt.h>
#include <sys/time.h>
#include <unistd.h>

#include "nfs.h"

#define READ_ONLY_OPT 1

#include "libbyz.h"
#include "th_assert.h"

static int relay(int s);
// Effects: Loop forever relaying messages received in socket "s" to
// the replicas associated with client "c".

int main (int argc, char *argv[]) {
  // Process command line options.
  bool multi = false;
  bool mmulti = false;
  int opt;
  while ((opt = getopt(argc, argv, "mM")) != EOF) {
    switch (opt) {
    case 'm':
      multi = true;
      break;

    case 'M':
      multi = true;
      mmulti = true;
      break;

    default:
      fprintf(stderr, "%s -m ", argv[0]);
      exit(-1);
    }
  }

  //
  // Initialize nfsd proxy
  //
  
  // Create socket and name it.
  int nfsd_socket;
  if ((nfsd_socket = socket(AF_INET, SOCK_DGRAM, 0)) == -1) {
    th_fail("Could not create socket");
  }
  
  struct sockaddr_in a;
  bzero((char *)&a, sizeof(a));
  a.sin_addr.s_addr = inet_addr("127.0.0.1");
  a.sin_family = AF_INET;
  a.sin_port = htons(NFSD_PROXY_PORT);
  if (bind(nfsd_socket, (struct sockaddr *) &a, sizeof(a)) == -1) {
    th_fail("Could not bind name to socket");
  }
  
  pid_t pid1 = 0;
  pid_t pid2 = 0;
  if (multi) {
    // Fork child.
    pid1 = fork();
    if (pid1 == -1) {
      th_fail("Fork failed");
    }

    if (mmulti) {
      pid2 = fork();
      if (pid2 == -1) {
	th_fail("Fork failed");
      }
    }
  }

  // Get name of private config file
  char hname[MAXHOSTNAMELEN];
  gethostname(hname, MAXHOSTNAMELEN);
  char config_priv[PATH_MAX];
  sprintf(config_priv, "config_private/%s", hname);

  short port_offset = ((pid1 != 0) << 1) |  (pid2 != 0); 
  short port = 3458 + port_offset;
  
  // Initialize client
  Byz_init_client("./config", config_priv, port);

  // Loop relaying messages.
  relay(nfsd_socket);
}

static int relay(int s) {  
  fprintf(stderr, "Client pid=%d is ready\n", getpid());

#ifdef LOGOPS
  struct timeval tval;
  char log_file_name[PATH_MAX];
  sprintf(log_file_name, "client_log%d", getpid());
  FILE *log_file = fopen(log_file_name, "w"); 
#endif //LOGOPS

  XDR in;
  Byz_req req;
  Byz_alloc_request(&req, 0);
  int max_len = req.size;
  Byz_rep rep;

  static unsigned current_xid = (unsigned)-1;
  struct sockaddr_in  a;
  unsigned addr_len = sizeof(a);
  while (1) {
    int ret = recvfrom(s, req.contents, max_len, 0, 
		       (struct sockaddr*)&a, &addr_len);    
    if (ret <= 0) {
      perror("recvfrom() failed or request too big\n");
      continue;
    }

    struct rpc_msg *m = (struct rpc_msg *)req.contents;
    if (m->rm_xid == current_xid) {
      // NFS client will trigger retransmissions. Thus, need to check
      // transaction identifier to ignore retransmissions.
      continue;
    }

    req.size = ret;

    // Check if request is read-only (i.e. getattr).
    xdrmem_create(&in, req.contents, ret, XDR_DECODE);
    struct rpc_msg rm; 
    char cred_area[MAX_AUTH_BYTES];
    rm.rm_call.cb_cred.oa_base = cred_area;
    rm.rm_call.cb_verf.oa_base = cred_area;
    xdr_callmsg(&in, &rm);
#ifdef READ_ONLY_OPT
    bool read_only = (rm.rm_call.cb_proc == 1) || (rm.rm_call.cb_proc == 4) 
      || (rm.rm_call.cb_proc == 6) ;
#else
    bool read_only = (rm.rm_call.cb_proc == 1);
#endif

#ifdef LOGOPS
    // Log operation start:
    gettimeofday(&tval, 0);
    long usecs = (long)(tval.tv_sec)*1000000+tval.tv_usec;
    fprintf(log_file, "time_in=%ld proc=%d arg=%d ", usecs, rm.rm_call.cb_proc, ret);
#endif // LOGOPS

    Byz_invoke(&req, &rep, read_only);

#ifdef LOGOPS
    // Log operation end:
    gettimeofday(&tval, 0);
    usecs = (long)(tval.tv_sec)*1000000+tval.tv_usec;
    fprintf(log_file, "rep=%d time_out=%ld\n", rep_len, usecs);
#endif // LOGOPS

    ret = sendto(s, rep.contents, rep.size, 0, (struct sockaddr*)&a, addr_len);
    if (ret < 0) {
      perror("sendto() failed\n");
    }

    Byz_free_reply(&rep);

    current_xid = m->rm_xid;
  }
}
