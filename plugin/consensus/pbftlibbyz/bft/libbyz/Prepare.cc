#include "th_assert.h"
#include "Message_tags.h"
#include "Prepare.h"
#include "Node.h"
#include "Replica.h"
#include "Principal.h"

Prepare::Prepare(View v, Seqno s, Digest &d, Principal* dst) : 
  Message(Prepare_tag, sizeof(Prepare_rep) 
#ifndef USE_PKEY
	  + ((dst) ? MAC_size : node->auth_size())) {
#else
          + ((dst) ? MAC_size : node->sig_size())) {
#endif
    rep().extra = (dst) ? 1 : 0;
    rep().view = v;
    rep().seqno = s;
    rep().digest = d;
    rep().id = node->id();
    rep().padding = 0;
    if (!dst) {
#ifndef USE_PKEY
      node->gen_auth_out(contents(), sizeof(Prepare_rep));
#else
      node->gen_signature(contents(), sizeof(Prepare_rep), 
			  contents()+sizeof(Prepare_rep));
#endif
    } else {
      dst->gen_mac_out(contents(), sizeof(Prepare_rep), 
		       contents()+sizeof(Prepare_rep));
    } 
}


void Prepare::re_authenticate(Principal *p) {
  if (rep().extra == 0) {
#ifndef USE_PKEY
    node->gen_auth_out(contents(), sizeof(Prepare_rep));
#endif
  } else
    p->gen_mac_out(contents(), sizeof(Prepare_rep), 
		   contents()+sizeof(Prepare_rep));
}


bool Prepare::verify() {
  // This type of message should only be sent by a replica other than me
  // and different from the primary
  if (!node->is_replica(id()) || id() == node->id()) 
    return false;

  if (rep().extra == 0) {
    // Check signature size.
#ifndef USE_PKEY
    if (replica->primary(view()) == id() || 
	size()-(int)sizeof(Prepare_rep) < node->auth_size(id())) 
      return false;

    return node->verify_auth_in(id(), contents(), sizeof(Prepare_rep));
#else
    if (replica->primary(view()) == id() || 
	size()-(int)sizeof(Prepare_rep) < node->sig_size(id())) 
      return false;

    return node->i_to_p(id())->verify_signature(contents(), sizeof(Prepare_rep), 
						contents()+sizeof(Prepare_rep));
#endif


  } else {
    if (size()-(int)sizeof(Prepare_rep) < MAC_size)
      return false;

    return node->i_to_p(id())->verify_mac_in(contents(), sizeof(Prepare_rep), 
				      contents()+sizeof(Prepare_rep));
  }

  return false;
}


bool Prepare::convert(Message *m1, Prepare  *&m2) {
  if (!m1->has_tag(Prepare_tag, sizeof(Prepare_rep)))
    return false;

  m2 = (Prepare*)m1;
  m2->trim();
  return true;
}
 


