#include <strings.h>
#include "th_assert.h"
#include "Message_tags.h"
#include "Reply.h"
#include "Node.h"
#include "Principal.h"

#include "Statistics.h"

Reply::Reply(View view, Request_id req, int replica) : 
  Message(Reply_tag, Max_message_size) {
    rep().v = view; 
    rep().rid = req;
    rep().replica = replica;
    rep().reply_size = 0;
    set_size(sizeof(Reply_rep));
}


Reply::Reply(Reply_rep *r) : Message(r) {}


Reply::Reply(View view, Request_id req, int replica, Digest &d, 
	     Principal *p, bool tentative) :
  Message(Reply_tag, sizeof(Reply_rep)+MAC_size) {

	if (tentative) rep().extra = 1;
	else rep().extra = 0;
	
    rep().v = view; 
    rep().rid = req;
    rep().replica = replica;
    rep().reply_size = -1;
    rep().digest = d;

    INCR_OP(reply_auth);
    START_CC(reply_auth_cycles);
    p->gen_mac_out(contents(), sizeof(Reply_rep), contents()+sizeof(Reply_rep));
    STOP_CC(reply_auth_cycles);
}


Reply* Reply::copy(int id) const {
  Reply* ret = (Reply*)new Message(msg->size);
  memcpy(ret->msg, msg, msg->size);
  ret->rep().replica = id;
  return ret;
}


char *Reply::store_reply(int &max_len) {
  max_len = msize()-sizeof(Reply_rep)-MAC_size;
  return contents()+sizeof(Reply_rep);
}


void Reply::authenticate(Principal *p, int act_len, bool tentative) {
  th_assert((unsigned)act_len <= msize()-sizeof(Reply_rep)-MAC_size,
	    "Invalid reply size");

  if (tentative) rep().extra = 1;
  
  rep().reply_size = act_len;
  rep().digest = Digest(contents()+sizeof(Reply_rep), act_len);
  int old_size = sizeof(Reply_rep)+act_len;
  set_size(old_size+MAC_size);

  INCR_OP(reply_auth);
  START_CC(reply_auth_cycles);
  p->gen_mac_out(contents(), sizeof(Reply_rep), contents()+old_size);
  STOP_CC(reply_auth_cycles);

  trim();
}


void Reply::re_authenticate(Principal *p) {
  int old_size = sizeof(Reply_rep)+rep().reply_size;

  INCR_OP(reply_auth);
  START_CC(reply_auth_cycles);
  p->gen_mac_out(contents(), sizeof(Reply_rep), contents()+old_size);
  STOP_CC(reply_auth_cycles);
}


void Reply::commit(Principal *p) {
  if (rep().extra == 0) return; // Reply is already committed.

  rep().extra = 0;
  int old_size = sizeof(Reply_rep)+rep().reply_size;
  p->gen_mac_out(contents(), sizeof(Reply_rep), contents()+old_size);
}


bool Reply::verify() {
  // Replies must be sent by replicas.
  if (!node->is_replica(id())) 
    return false;

  // Check sizes
  int rep_size = (full()) ? rep().reply_size : 0;
  if (size()-(int)sizeof(Reply_rep)-rep_size < MAC_size) 
    return false;

  // Check reply
  if (full()) {
    Digest d(contents()+sizeof(Reply_rep), rep_size);
    if (d != rep().digest)
      return false;
  }

  // Check signature.
  Principal *replica = node->i_to_p(rep().replica);
  int size_wo_MAC = sizeof(Reply_rep)+rep_size;
  
  INCR_OP(reply_auth_ver);
  START_CC(reply_auth_ver_cycles);

  bool ret = replica->verify_mac_in(contents(), sizeof(Reply_rep), contents()+size_wo_MAC);

  STOP_CC(reply_auth_ver_cycles);

  return ret;
}


bool Reply::convert(Message *m1, Reply *&m2) {
  if (!m1->has_tag(Reply_tag, sizeof(Reply_rep)))
    return false;
  
  m1->trim();
  m2 = (Reply*)m1;
  return true;
}
