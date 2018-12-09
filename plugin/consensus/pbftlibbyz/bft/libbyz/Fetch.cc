#include "th_assert.h"
#include "Message_tags.h"
#include "Partition.h"
#include "Fetch.h"
#include "Node.h"
#include "Replica.h"
#include "Principal.h"
#include "State_defs.h"

Fetch::Fetch(Request_id rid, Seqno lu, int level, int index,
#ifndef NO_STATE_TRANSLATION
	     int chunkn,
#endif
	     Seqno rc, int repid) :
  Message(Fetch_tag, sizeof(Fetch_rep) + node->auth_size()) {
  rep().rid = rid;
  rep().lu = lu;
  rep().level = level;
  rep().index = index;
  rep().rc = rc;
  rep().repid = repid;
  rep().id = node->id();
#ifndef NO_STATE_TRANSLATION
  rep().chunk_no = chunkn;
  rep().padding = 0;
#endif
  node->gen_auth_in(contents(), sizeof(Fetch_rep));
}

void Fetch::re_authenticate(Principal *p) {
  node->gen_auth_in(contents(), sizeof(Fetch_rep));
}

bool Fetch::verify() {
  if (!node->is_replica(id())) 
    return false;
  
  if (level() < 0 || level() >= PLevels)
    return false;
  
  if (index() < 0 || index() >=  PLevelSize[level()])
    return false;
  
  if (checkpoint() == -1 && replier() != -1)
    return false; 

#ifndef NO_STATE_TRANSLATION
  if (chunk_number() < 0)
    return false;
#endif

  // Check signature size.
  if (size()-(int)sizeof(Fetch_rep) < node->auth_size(id())) 
    return false;

  return node->verify_auth_out(id(), contents(), sizeof(Fetch_rep));
}


bool Fetch::convert(Message *m1, Fetch  *&m2) {
  if (!m1->has_tag(Fetch_tag, sizeof(Fetch_rep)))
    return false;

  m2 = (Fetch*)m1;
  m2->trim();
  return true;
}
 


