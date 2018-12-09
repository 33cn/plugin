#include <stdlib.h>
#include <string.h>

#include "Rep_info.h"
#include "Replica.h"
#include "Reply.h"
#include "Req_queue.h"

#include "Statistics.h"
#include "State_defs.h"

#include "Array.t"

#ifndef NO_STATE_TRANSLATION

Rep_info::Rep_info(int n) {
  th_assert(n != 0, "Invalid argument");

  nps = n;
  mem = (char *)valloc(size());

#else

Rep_info::Rep_info(char *m, int sz, int n) {
  th_assert(n != 0, "Invalid argument");

  nps = n;
  mem = m;

  if (sz < (nps+1)*Max_rep_size)
    th_fail("Memory is too small to hold replies for all principals");
 
#endif
  
  int old_nps = *((Long*)mem);
  if (old_nps != 0) {
    // Memory has already been initialized.
    if (nps != old_nps)
      th_fail("Changing number of principals. Not implemented yet");
  } else {
    // Initialize memory.
    bzero(mem, (nps+1)*Max_rep_size);
    for (int i=0; i < nps; i++) {
      // Wasting first page just to store the number of principals.
      Reply_rep* rr = (Reply_rep*)(mem+(i+1)*Max_rep_size);
      rr->tag = Reply_tag;
      rr->reply_size = -1;
      rr->rid = 0;
    }
    *((Long*)mem) = nps;
  }

  struct Rinfo ri;
  ri.tentative = true;
  ri.lsent = zeroTime();

  for (int i=0; i < nps; i++) {
    Reply_rep *rr = (Reply_rep*)(mem+(i+1)*Max_rep_size);
    th_assert(rr->tag == Reply_tag, "Corrupt memory");
    reps.append(new Reply(rr));
    ireps.append(ri);
  }
}


Rep_info::~Rep_info() {
  for (int i=0; i < nps; i++) 
    delete reps[i];
}


char* Rep_info::new_reply(int pid, int &sz) {
  Reply* r = reps[pid];

#ifndef NO_STATE_TRANSLATION
  for(int i=(r->contents()-mem)/Block_size;
      i<=(r->contents()+Max_rep_size-1-mem)/Block_size;i++) {
    //    fprintf(stderr,"modifying reply: mem %d begin %d end %d \t page %d size %d\n",mem, r->contents(), r->contents()+Max_rep_size-1, i,size());
    replica->modify_index_replies(i);
  }
#else
  replica->modify(r->contents(), Max_rep_size);
#endif
  ireps[pid].tentative = true;
  ireps[pid].lsent = zeroTime();
  r->rep().reply_size = -1;
  sz = Max_rep_size-sizeof(Reply_rep)-MAC_size;
  return r->contents()+sizeof(Reply_rep);
}


void Rep_info::end_reply(int pid, Request_id rid, int sz) {
  Reply* r = reps[pid];
  th_assert(r->rep().reply_size == -1, "Invalid state");

  Reply_rep& rr = r->rep();
  rr.rid = rid;
  rr.reply_size = sz;
  rr.digest = Digest(r->contents()+sizeof(Reply_rep), sz);

  int old_size = sizeof(Reply_rep)+rr.reply_size;
  r->set_size(old_size+MAC_size);
  bzero(r->contents()+old_size, MAC_size);
}

void Rep_info::send_reply(int pid, View v, int id, bool tentative) {
  Reply *r = reps[pid];
  Reply_rep& rr = r->rep();
  int old_size = sizeof(Reply_rep)+rr.reply_size;

  th_assert(rr.reply_size != -1, "Invalid state");
  th_assert(rr.extra == 0 && rr.v == 0 && rr.replica == 0, "Invalid state");

  if (!tentative && ireps[pid].tentative) {
    ireps[pid].tentative = false;
    ireps[pid].lsent = zeroTime();
  }

  Time cur;
  Time& lsent = ireps[pid].lsent;
  if (lsent != 0) {
    cur = currentTime();
    if (diffTime(cur, lsent) <= 10000) 
      return;

    lsent = cur;
  }
  
  if (ireps[pid].tentative) rr.extra = 1;
  rr.v = v;
  rr.replica = id;
  Principal *p = node->i_to_p(pid);

  INCR_OP(reply_auth);
  START_CC(reply_auth_cycles);
  p->gen_mac_out(r->contents(), sizeof(Reply_rep), r->contents()+old_size);
  STOP_CC(reply_auth_cycles);

  node->send(r, pid);

  // Undo changes. To ensure state matches across all replicas.
  rr.extra = 0;
  rr.v = 0;
  rr.replica = 0;
  bzero(r->contents()+old_size, MAC_size);
}


bool Rep_info::new_state(Req_queue *rset) {
  bool first=false;
  for (int i=0; i < nps; i++) {
    commit_reply(i);

    // Remove requests from rset with stale timestamps.
    if (rset->remove(i, req_id(i)))
      first = true;
  }
  return first;
}
