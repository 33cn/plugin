#include "th_assert.h"
#include "Message_tags.h"
#include "Reply_stable.h"
#include "Replica.h"
#include "Principal.h"
 
Reply_stable::Reply_stable(Seqno lc, Seqno lp, int n, Principal* p) :
  Message(Reply_stable_tag, sizeof(Reply_stable_rep) + MAC_size) {
  rep().lc = lc;
  rep().lp = lp;
  rep().id = node->id();
  rep().nonce = n;

  p->gen_mac_out(contents(), sizeof(Reply_stable_rep), 
		 contents()+sizeof(Reply_stable_rep));
}


void Reply_stable::re_authenticate(Principal *p) {
  p->gen_mac_out(contents(), sizeof(Reply_stable_rep), 
		 contents()+sizeof(Reply_stable_rep));
}


bool Reply_stable::verify() {
  // Reply_stables must be sent by replicas.
  if (!node->is_replica(id())) return false;
  
  if (rep().lc%checkpoint_interval != 0 || rep().lc > rep().lp) return false;

  // Check size.
  if (size()-(int)sizeof(Reply_stable_rep) < MAC_size) 
    return false;

  // Check MAC
  Principal *p = node->i_to_p(id());
  if (p) {
    return p->verify_mac_in(contents(), sizeof(Reply_stable_rep), 
			    contents()+sizeof(Reply_stable_rep));
  }

  return false;
}


bool Reply_stable::convert(Message *m1, Reply_stable  *&m2) {
  if (!m1->has_tag(Reply_stable_tag, sizeof(Reply_stable_rep)))
    return false;
  m1->trim();
  m2 = (Reply_stable*)m1;
  return true;
}
