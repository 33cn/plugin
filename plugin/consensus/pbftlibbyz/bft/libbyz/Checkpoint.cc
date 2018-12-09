#include "th_assert.h"
#include "Message_tags.h"
#include "Checkpoint.h"
#include "Replica.h"
#include "Principal.h"
 
Checkpoint::Checkpoint(Seqno s, Digest &d, bool stable) : 
#ifndef USE_PKEY
  Message(Checkpoint_tag, sizeof(Checkpoint_rep) + node->auth_size()) {
#else
  Message(Checkpoint_tag, sizeof(Checkpoint_rep) + node->sig_size()) {
#endif
    rep().extra = (stable) ? 1 : 0;
    rep().seqno = s;
    rep().digest = d;
    rep().id = node->id();
    rep().padding = 0;

#ifndef USE_PKEY
    node->gen_auth_out(contents(), sizeof(Checkpoint_rep));
#else
    node->gen_signature(contents(), sizeof(Checkpoint_rep), 
		      contents()+sizeof(Checkpoint_rep));
#endif
}

void Checkpoint::re_authenticate(Principal *p, bool stable) {
#ifndef USE_PKEY
  if (stable) rep().extra = 1;
  node->gen_auth_out(contents(), sizeof(Checkpoint_rep));
#else
  if (rep().extra != 1 && stable) {
    rep().extra = 1;
    node->gen_signature(contents(), sizeof(Checkpoint_rep), 
			contents()+sizeof(Checkpoint_rep));
  }
#endif
}

bool Checkpoint::verify() {
  // Checkpoints must be sent by replicas.
  if (!node->is_replica(id())) return false;
  
  // Check signature size.
#ifndef USE_PKEY
  if (size()-(int)sizeof(Checkpoint_rep) < node->auth_size(id())) 
    return false;

  return node->verify_auth_in(id(), contents(), sizeof(Checkpoint_rep));
#else
  if (size()-(int)sizeof(Checkpoint_rep) < node->sig_size(id())) 
    return false;

  return node->i_to_p(id())->verify_signature(contents(), sizeof(Checkpoint_rep),
					      contents()+sizeof(Checkpoint_rep));
#endif
}

bool Checkpoint::convert(Message *m1, Checkpoint  *&m2) {
  if (!m1->has_tag(Checkpoint_tag, sizeof(Checkpoint_rep)))
    return false;
  m1->trim();
  m2 = (Checkpoint*)m1;
  return true;
}
