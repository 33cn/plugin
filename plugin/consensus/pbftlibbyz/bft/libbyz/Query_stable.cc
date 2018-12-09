#include "th_assert.h"
#include "Message_tags.h"
#include "Query_stable.h"
#include "Replica.h"
#include "Principal.h"
 
Query_stable::Query_stable() : 
  Message(Query_stable_tag, sizeof(Query_stable_rep) + node->auth_size()) {
  rep().id = node->id();
  rep().nonce = random_int();
  node->gen_auth_out(contents(), sizeof(Query_stable_rep));
}

void Query_stable::re_authenticate(Principal *p) {
  node->gen_auth_out(contents(), sizeof(Query_stable_rep));
}

bool Query_stable::verify() {
  // Query_stables must be sent by replicas.
  if (!node->is_replica(id())) return false;
  
  // Check signature size.
  if (size()-(int)sizeof(Query_stable_rep) < node->auth_size(id())) 
    return false;

  return node->verify_auth_in(id(), contents(), sizeof(Query_stable_rep));
}

bool Query_stable::convert(Message *m1, Query_stable  *&m2) {
  if (!m1->has_tag(Query_stable_tag, sizeof(Query_stable_rep)))
    return false;
  m1->trim();
  m2 = (Query_stable*)m1;
  return true;
}
