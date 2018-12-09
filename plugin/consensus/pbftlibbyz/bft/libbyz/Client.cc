#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/types.h> 
#include <unistd.h>
#include <sys/time.h>

#include "th_assert.h"
#include "Client.h"
#include "ITimer.h"
#include "Message.h"
#include "Reply.h"
#include "Request.h"

//#define ADJUST_RTIMEOUT 1

Client::Client(FILE *config_file, FILE *config_priv, short port) : 
  Node(config_file, config_priv, port), t_reps(2*f()+1), c_reps(f()+1) {
  // Fail if node is is a replica.
  if (is_replica(id())) th_fail("Node is a replica");

  rtimeout = 150; // Initial timeout value
  rtimer = new ITimer(rtimeout, rtimer_handler);

  out_rid = new_rid();
  out_req = 0;     

  // Multicast new key to all replicas.
  send_new_key();
  atimer->start();
}

Client::~Client() {
  delete rtimer;
}

void Client::reset() {
  rtimeout = 150;
}

bool Client::send_request(Request *req) {
  bool ro = req->is_read_only();
  if (out_req == 0) { 
    // Send request to service
    if (ro || req->size() > Request::big_req_thresh) {
      // read-only requests and big requests are multicast to all replicas.
      send(req, All_replicas);
    } else {
      // read-write requests are sent to the primary only.
      send(req, primary());
    }
    out_req = req;
    need_auth = false;
    n_retrans = 0;
    
#ifdef ADJUST_RTIMEOUT
    // Adjust timeout to reflect average latency
    rtimer->adjust(rtimeout);

    // Start timer to measure request latency
    latency.reset();
    latency.start();
#endif

    rtimer->start();
    return true;
  } else {
    // Another request is being processed.
    return false;
  }
}

Reply *Client::recv_reply() {
  if (out_req == 0)
    // Nothing to wait for.
    return 0;

  //
  // Wait for reply
  // 
  while (1) {
    Message* m = recv();
    
    Reply* rep;
    if (!Reply::convert(m, rep) || rep->request_id() != out_rid) {
      delete m;
      continue;
    }
    
    Certificate<Reply> &reps = (rep->is_tentative()) ? t_reps : c_reps;
    if (reps.is_complete()) {
    	// We have a complete certificate without a full reply.
    	if (!rep->full() || !rep->verify() || !rep->match(reps.cvalue())) {
    		delete rep;
    		continue;
    	}
    } else {
      reps.add(rep);
      rep = (reps.is_complete() && reps.cvalue()->full()) ? 
                                             reps.cvalue_clear() : 0;
    }
      
    if (rep) {
      // printf("request %d has committed\n", (int)rep->request_id());
    	
      out_rid = new_rid();
      rtimer->stop();
      out_req = 0;
      t_reps.clear();
      c_reps.clear();

      // Choose view in returned rep. TODO: could make performance
      // more robust to attacks by picking the median view in the
      // certificate.
      v = rep->view();
      cur_primary = v % num_replicas;

#ifdef ADJUST_RTIMEOUT
      latency.stop();
      rtimeout = (3*rtimeout+
		  latency.elapsed()*Rtimeout_mult/(clock_mhz*1000))/4+1;
#endif

      return rep;
    }
  }
}

void rtimer_handler() {
  th_assert(node, "Client is not initialized");
  ((Client*)node)->retransmit();
}


void Client::retransmit() {
  // Retransmit any outstanding request.
  static const int thresh = 1;
  static const int nk_thresh = 4;
  static const int nk_thresh_1 = 100;

  if (out_req != 0) {
    INCR_OP(req_retrans);

    //    fprintf(stderr, ".");
    n_retrans++;
    if (n_retrans == nk_thresh || n_retrans % nk_thresh_1 == 0) {
      send_new_key();
    }

    bool ro = out_req->is_read_only();
    bool change = (ro || out_req->replier() >= 0) && n_retrans > thresh;
    //    printf("%d %d %d %d\n", id(), n_retrans, ro, out_req->replier());

    if (need_auth || change) {
      // Compute new authenticator for request
      out_req->re_authenticate(change);
      need_auth = false;
      if (ro && change) t_reps.clear();
    }

    if (out_req->is_read_only() || n_retrans > thresh 
	|| out_req->size() > Request::big_req_thresh) {
      // read-only requests, requests retransmitted more than
      // mcast_threshold times, and big requests are multicast to all
      // replicas.
      send(out_req, All_replicas);
    } else {
      // read-write requests are sent to the primary only.
      send(out_req, primary());
    }
  }

#ifdef ADJUST_RTIMEOUT
  // exponential back off
  if (rtimeout < Min_rtimeout) rtimeout = 100;
  rtimeout = rtimeout+lrand48()%rtimeout;
  if (rtimeout > Max_rtimeout) rtimeout = Max_rtimeout;
  rtimer->adjust(rtimeout);
#endif

  rtimer->restart();
}

void Client::send_new_key() {
  Node::send_new_key();
  need_auth = true;

  // Cleanup reply messages authenticated with old keys.
  t_reps.clear();
  c_reps.clear();
}
