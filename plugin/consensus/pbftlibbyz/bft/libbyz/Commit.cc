#include "th_assert.h"
#include "Message_tags.h"
#include "Commit.h"
#include "Node.h"
#include "Replica.h"
#include "Principal.h"

Commit::Commit(View v, Seqno s) : 
  Message(Commit_tag, sizeof(Commit_rep) + node->auth_size()) {
    rep().view = v;
    rep().seqno = s;
    rep().id = node->id(); 
    rep().padding = 0;
    node->gen_auth_out(contents(), sizeof(Commit_rep));
}


Commit::Commit(Commit_rep *contents) : Message(contents) {}

void Commit::re_authenticate(Principal *p) {
  node->gen_auth_out(contents(), sizeof(Commit_rep));
}

bool Commit::verify() {
  // Commits must be sent by replicas.
  if (!node->is_replica(id()) || id() == node->id()) return false;

  // Check signature size.
  if (size()-(int)sizeof(Commit_rep) < node->auth_size(id())) 
    return false;

  return node->verify_auth_in(id(), contents(), sizeof(Commit_rep));
}


bool Commit::convert(Message *m1, Commit  *&m2) {
  if (!m1->has_tag(Commit_tag, sizeof(Commit_rep)))
    return false;

  m2 = (Commit*)m1;
  m2->trim();
  return true;
}

bool Commit::convert(char *m1, unsigned max_len, Commit &m2) {
  // First check if we can use m1 to create a Commit.
  if (!Message::convert(m1, max_len, Commit_tag, sizeof(Commit_rep),m2)) 
    return false;
  return true;
}
 
