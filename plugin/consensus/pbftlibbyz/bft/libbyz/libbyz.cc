#include <stdio.h>
#include <stdlib.h>
#include <signal.h>
#include <sys/types.h>
#include <sys/wait.h>

#include "libbyz.h"
#include "Request.h"
#include "Reply.h"
#include "Client.h"
#include "Replica.h"
#include "Statistics.h"
#include "State_defs.h"

static void wait_chld(int sig) {
  // Get rid of zombies created by sfs code.
  while (waitpid(-1, 0, WNOHANG) > 0);
}


int Byz_init_client(char *conf, char *conf_priv, short port) {
  // signal handler to get rid of zombies
  struct sigaction act;
  act.sa_handler = wait_chld;
  sigemptyset (&act.sa_mask);
  act.sa_flags = 0;
  sigaction (SIGCHLD, &act, NULL);
  
  FILE *config_file = fopen(conf, "r");
  if (config_file == 0) {
    fprintf(stderr, "libbyz: Invalid configuration file %s \n", conf);
    return -1;
  }
    
  FILE *config_priv_file = fopen(conf_priv, "r");
  if (config_priv_file == 0) {
    fprintf(stderr, "libbyz: Invalid private configuration file %s \n", conf);
    return -1;
  }

  // Initialize random number generator
  srand48(getpid());

  Client* client = new Client(config_file, config_priv_file, port);
  node = client;
  return 0;
}

void Byz_reset_client() {
  ((Client*)node)->reset();
}

int Byz_alloc_request(Byz_req *req, int size) {
  Request* request = new Request((Request_id)0);
  if (request == 0) return -1;

  int len;
  req->contents = request->store_command(len);
  req->size = len;
  req->opaque = (void*)request;
  return 0;
}


int Byz_send_request(Byz_req *req, bool ro) {
  Request *request = (Request *) req->opaque;
  request->request_id() = ((Client*)node)->get_rid();
  request->authenticate(req->size, ro);
  
  bool retval = ((Client*)node)->send_request(request);
  return (retval) ? 0 : -1;
}


int Byz_recv_reply(Byz_rep *rep) {
  Reply *reply = ((Client*)node)->recv_reply();
  if (reply==NULL)
    return -1;
  rep->contents = reply->reply(rep->size);
  rep->opaque = reply;
  return 0;
}

int Byz_invoke(Byz_req *req, Byz_rep *rep, bool ro) {
  if (Byz_send_request(req, ro)==-1)
    return -1;
  return Byz_recv_reply(rep);
}



void Byz_free_request(Byz_req *req) {
  Request *request = (Request *) req->opaque;
  delete request;
}


void Byz_free_reply(Byz_rep *rep) {
  Reply *reply = (Reply *) rep->opaque;
  delete reply;
}


#ifndef NO_STATE_TRANSLATION

int Byz_init_replica(char *conf, char *conf_priv, unsigned int num_objs, 
		     int (*exec)(Byz_req *, Byz_rep *, Byz_buffer*, int, bool),
		     void (*comp_ndet)(Seqno, Byz_buffer *), int ndet_max_len,
		     bool (*check_ndet)(Byz_buffer *),
		     int (*get)(int, char **),
		     void (*put)(int, int *, int *, char **),
		     void (*shutdown_proc)(FILE *o),
		     void (*restart_proc)(FILE *i),
		     short port) {

  // signal handler to get rid of zombies
  struct sigaction act;
  act.sa_handler = wait_chld;
  sigemptyset (&act.sa_mask);
  act.sa_flags = 0;
  sigaction (SIGCHLD, &act, NULL);

  FILE *config_file = fopen(conf, "r");
  if (config_file == 0) {
    fprintf(stderr, "libbyz: Invalid configuration file %s \n", conf);
    return -1;
  }
  
  FILE *config_priv_file = fopen(conf_priv, "r");
  if (config_priv_file == 0) {
    fprintf(stderr, "libbyz: Invalid private configuration file %s \n", conf_priv);
    return -1;
  }

  // Initialize random number generator
  srand48(getpid());

  replica = new Replica(config_file, config_priv_file, num_objs,
			get, put, shutdown_proc, restart_proc, port);
  node = replica;

  // Register service-specific functions.
  replica->register_exec(exec);
  replica->register_nondet_choices(comp_ndet, ndet_max_len, check_ndet);
  return replica->used_state_pages();
}

void Byz_modify(int npages, int* pages) {

  for (int i=0; i<npages; i++)
    replica->modify_index(pages[i]);
}

#else

int Byz_init_replica(char *conf, char *conf_priv, char *mem, unsigned int size,
		     int (*exec)(Byz_req *, Byz_rep *, Byz_buffer *, int, bool),
		     void (*comp_ndet)(Seqno, Byz_buffer *), int ndet_max_len) {
  // signal handler to get rid of zombies
  struct sigaction act;
  act.sa_handler = wait_chld;
  sigemptyset (&act.sa_mask);
  act.sa_flags = 0;
  sigaction (SIGCHLD, &act, NULL);

  FILE *config_file = fopen(conf, "r");
  if (config_file == 0) {
    fprintf(stderr, "libbyz: Invalid configuration file %s \n", conf);
    return -1;
  }
  
  FILE *config_priv_file = fopen(conf_priv, "r");
  if (config_priv_file == 0) {
    fprintf(stderr, "libbyz: Invalid private configuration file %s \n", conf_priv);
    return -1;
  }

  // Initialize random number generator
  srand48(getpid());

  replica = new Replica(config_file, config_priv_file, mem, size);
  node = replica;

  // Register service-specific functions.
  replica->register_exec(exec);
  replica->register_nondet_choices(comp_ndet, ndet_max_len);
  return replica->used_state_bytes();
}

void Byz_modify(char *mem, int size) {
  replica->modify(mem, size);
}

#endif

void Byz_replica_run() {
  stats.zero_stats();
  replica->recv();
}

#ifdef NO_STATE_TRANSLATION
void _Byz_modify_index(int bindex) {
  replica->modify_index(bindex);
}
#endif

void Byz_reset_stats() {
  stats.zero_stats();
}

void Byz_print_stats() {
  stats.print_stats();
}

#ifndef NO_STATE_TRANSLATION
char* Byz_get_cached_object(int i) {
  return replica->get_cached_obj(i);
}
#endif
