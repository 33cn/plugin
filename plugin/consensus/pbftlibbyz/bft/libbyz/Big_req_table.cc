#include "Request.h"
#include "Pre_prepare.h"
#include "Big_req_table.h"
#include "Replica.h"

#include "Array.t"
#include "bhash.t"
#include "buckets.t"


struct Waiting_pp {
  Seqno n;
  int i;
};

class BR_entry {
public:
  inline BR_entry() : r(0), maxn(-1), maxv(-1) {}
  inline ~BR_entry() { delete r; }

  Digest rd;            // Request's digest
  Request *r;           // Request or 0 is request not received
  Array<Waiting_pp> waiting; // if r=0, Seqnos of pre-prepares waiting for request
  Seqno maxn;           // Maximum seqno of pre-prepare referencing request 
  View maxv;            // Maximum view in which this entry was marked useful
};


Big_req_table::Big_req_table() : breqs(max_out), unmatched((Request*)0, node->np()) {
  max_entries = max_out*Pre_prepare::big_req_max;
}


Big_req_table::~Big_req_table() {
  MapGenerator<Digest,BR_entry*> g(breqs);
  Digest d;
  BR_entry* bre;
  while (g.get(d, bre)) {
    delete bre;
  }
  breqs.clear();
} 

inline void Big_req_table::remove_unmatched(BR_entry* bre) {
  if (bre->maxn < 0) {
    th_assert(bre->r != 0, "Invalid state");
    unmatched[bre->r->client_id()] = 0;      
  }
}

bool Big_req_table::add_pre_prepare(Digest& rd, int i, Seqno n, View v) {
  BR_entry* bre;
  if (breqs.find(rd, bre)) {
    remove_unmatched(bre);

    if (n > bre->maxn)
      bre->maxn = n;

    if (v > bre->maxv)
      bre->maxv = v;

    if (bre->r) {
      return true;
    } else {
      Waiting_pp wp;
      wp.i = i;
      wp.n = n;
      bre->waiting.append(wp);
    }
  } else {
    // No entry in breqs for rd
    bre = new BR_entry;
    bre->rd = rd;
    Waiting_pp wp;
    wp.i = i;
    wp.n = n;
    bre->waiting.append(wp);
    bre->maxn = n;
    bre->maxv = v;
    breqs.add(rd, bre);
  }

  return false;
}


void Big_req_table::add_pre_prepare(Request* r, Seqno n, View v) {
  BR_entry* bre;
  Digest rd = r->digest();
  if (breqs.find(rd, bre)) {
    remove_unmatched(bre);

    if (n > bre->maxn)
      bre->maxn = n;

    if (v > bre->maxv)
      bre->maxv = v;

    if (bre->r == 0) {
      bre->r = r;
    } else {
      delete r;
    } 
  } else {
    // No entry in breqs for rd
    bre = new BR_entry;
    bre->rd = rd;
    bre->r = r;
    bre->maxn = n;
    bre->maxv = v;
    breqs.add(rd, bre);
  }
}


bool Big_req_table::check_pcerts(BR_entry* bre) {
  th_assert(replica->has_new_view(), "Invalid state");

  for (int i=0; i < bre->waiting.size(); i++) {
    Waiting_pp wp = bre->waiting[i];
    if (replica->plog.within_range(wp.n)) {
      Prepared_cert& pc = replica->plog.fetch(wp.n);
      if (pc.is_pp_correct())
	return true;
    }
  }
  return false;
}


bool Big_req_table::add_request(Request* r, bool verified) {
  th_assert(r->size() > Request::big_req_thresh && !r->is_read_only(),
	    "Invalid Argument");
  BR_entry* bre;
  Digest rd = r->digest();
  if (breqs.find(rd, bre)) {
    if (bre->r == 0 
	&& (verified || !replica->has_new_view() || check_pcerts(bre))) {
      bre->r = r;
      while (bre->waiting.size()) {
	Waiting_pp wp = bre->waiting.high();
	Seqno n = wp.n;
	int i = wp.i;

	if (replica->has_new_view()) {
	  // Missing pre-prepare is in replica's plog.
	  if (replica->plog.within_range(n)) {
	    Prepared_cert& pc = replica->plog.fetch(n);
	    pc.add(bre->rd, i);
	    replica->send_prepare(pc);
	    if (pc.is_complete())
	      replica->send_commit(n);
	  }
	} else {
	  // Missing pre-prepare is in replica's view-info
	  replica->vi.add_missing(bre->rd, n, i);
	}

	bre->waiting.remove();				       
      }
      return true;
    }
  } else if (verified) {
    // Buffer the request with largest timestamp from client.
    int cid = r->client_id();
    Request_id rid = r->request_id();
    Request* old_req = unmatched[cid];

    if (old_req == 0 || old_req->request_id() < rid) {
      bre = new BR_entry;
      bre->rd = rd;
      bre->r = r;
      breqs.add(rd, bre);

      if (old_req) {
	breqs.remove_fast(old_req->digest());
	delete old_req;
      }

      unmatched[cid] = r;
      return true;
    } 
  }
  return false;
}


Request* Big_req_table::lookup(Digest& rd) {
  BR_entry* bre;
  if (breqs.find(rd, bre)) {
    return bre->r;
  }
  return 0;
}


void Big_req_table::mark_stable(Seqno ls) {
  if (breqs.size() > 0) {
    MapGenerator<Digest,BR_entry*> g(breqs);
    Digest d;
    BR_entry* bre;
    while (g.get(d, bre)) {
      if (bre->maxn <= ls && bre->maxv >= 0) {
	remove_unmatched(bre);
	delete bre;
	g.remove();
      }
    }
  }
}


void Big_req_table::view_change(View v) {
  if (breqs.size() > 0) {
    MapGenerator<Digest,BR_entry*> g(breqs);
    Digest d;
    BR_entry* bre;
    while (g.get(d, bre)) {
      if (bre->maxv < v) {
	remove_unmatched(bre);
	delete bre;
	g.remove();
      }
    }
  }
}

