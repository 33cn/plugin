#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/param.h>
#include <unistd.h>

#include "th_assert.h"
#include "Timer.h"
#include "libbyz.h"

#include "Statistics.h"

#include "thr_graph.h"
#include "simple.h"

static const int num_replicas = 4;
static const int num_clients = 200;
static const int num_client_nodes = 5;
static const int max_clients_node = num_clients/num_client_nodes;

Address client_addrs[num_clients];
static int clients;
void start_experiment(thr_command& out, int nc) {
  if (nc <= 0 || nc > num_clients || nc % num_client_nodes != 0)
    th_fail("Invalid arguments");

  int count = nc/num_client_nodes;
  for (int i=0; i < count; i++) {
    for (int j=0; j < num_client_nodes; j++) {
      int index = i+j*max_clients_node;
      if (sendto(clients, &out, sizeof(out),0, 
		 (struct sockaddr*)&(client_addrs[index]), sizeof(Address))<=0){
	th_fail("Sendto");
      }
    }
  }
}

int main(int argc, char **argv) {
  // Create socket to communicate with clients
  if ((clients = socket(AF_INET, SOCK_DGRAM, 0)) == -1) {
    th_fail("Could not create socket");
  }

  Address a;
  bzero((char *)&a, sizeof(a));
  a.sin_addr.s_addr = INADDR_ANY;
  a.sin_family = AF_INET;
  a.sin_port = htons(3456);
  if (bind(clients, (struct sockaddr *) &a, sizeof(a)) == -1) {
    th_fail("Could not bind name to socket");
  }

  thr_command out, in;

  // Wait for clients to tell us they are up
  for (int i=0; i < num_clients; i++) {
    size_t addr_len = sizeof(a);
    int ret = recvfrom(clients, &in, sizeof(in), 0, 
		   (struct sockaddr*)&a, &addr_len);    
    if (ret != sizeof(in) || in.tag != thr_up) {
      th_fail("Invalid message");
    }

    client_addrs[in.cid - num_replicas] = a;
    // Assuming client processes are evenly divided by the client
    // machines in blocks of contiguous cids.
  }
  printf("All ready\n");

  //
  // Run experiments
  //

  
  //int client_count[] = {5, 10, 15, 20, 30, 40, 50, 60, 80, 100, 120, 140};
  int client_count[] = {15, 40, 60, 80, 100, 120, 140};
  int niters[2][3] = {{600000, 330000, 60000},{780000, 420000, 60000}};
//  int niters[2][3] = {{6000, 3300, 6000},{7800, 420000, 60000}};
  static const int num_points = 3;
  double vars[sizeof(client_count)/sizeof(int)];

  for (int read_only = 1; read_only < 2; read_only++) {
    out.read_only = read_only;

    printf("Experiments with read only = %d\n", read_only);
    for (int opt = 1; opt < 2; opt++) {
      printf("Experiments for opt = %d\n", opt);
      out.tag = thr_start;
      out.op = opt;

      int max = sizeof(client_count)/sizeof(int);
      for (int i=0; i < max; i++) {
	int nc = client_count[i];
	
	int tot_iter = niters[read_only][opt]/5;
	out.num_iter = tot_iter/nc;
	tot_iter = out.num_iter*nc;

	double maxs[num_points];
	for (int j=0; j < num_points; j++) {
	  start_experiment(out, nc);
	  
	  // Wait for results
	  float max=0, min=3600;
	  for (int k=0; k < nc; k++) {
	    int ret = recvfrom(clients, &in, sizeof(in), 0, 0, 0);    
	    if (ret != sizeof(in) || in.tag != thr_done) {
	      th_fail("Invalid message");
	    }

	    if (in.elapsed > max)
	      max = in.elapsed;

	    if (in.elapsed < min)
	      min = in.elapsed;
	  }
	  if ((max-min)/max > .5)
	    printf("Large variance\n");

	  maxs[j] = max;

	  // Separate out experiments to improve independence
	  sleep(5); 
	}
	
	// Compute averages
	double sum = 0;
	for (int j=0; j < num_points; j++) {
	  sum += maxs[j];
	}
	float avg = tot_iter*num_points/sum;
	printf("%d %f\n", nc, avg);

	// Compute std
	sum = 0;
	for (int j=0; j < num_points; j++) {
	  double val = tot_iter/maxs[j] - avg;
	  sum += val*val;
	}
	vars[i] =  sum/(float)(num_points-1);
      }
    
      printf("Standard deviations:\n");
      for (int i=0; i < max; i++) {
	printf("%d std = %f std mean = %f\n", 
	       client_count[i],sqrt(vars[i]), sqrt(vars[i])/sqrt(num_points));
      }
    }
  }
  
  // Kill clients
  out.tag = thr_end;
  start_experiment(out, num_clients);
}
  
