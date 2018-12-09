#include "th_assert.h"
#include "Message_tags.h"
#include "View_change_ack.h"
#include "Node.h"
#include "Principal.h"

View_change_ack::View_change_ack(View v, int id, int vcid, Digest const &vcd) :
  Message(View_change_ack_tag, sizeof(View_change_ack_rep) + MAC_size) {
    rep().v = v;
    rep().id = node->id();
    rep().vcid = vcid;
    rep().vcd = vcd;
    
    int old_size = sizeof(View_change_ack_rep);
    set_size(old_size+MAC_size);
    Principal *p = node->i_to_p(node->primary(v));
    p->gen_mac_out(contents(), old_size, contents()+old_size);
}

void View_change_ack::re_authenticate(Principal *p) {
  p->gen_mac_out(contents(), sizeof(View_change_ack_rep), contents()+sizeof(View_change_ack_rep));
}

bool View_change_ack::verify() {
  // These messages must be sent by replicas other than me, the replica that sent the 
  // corresponding view-change, or the primary.
  if (!node->is_replica(id()) || id() == node->id() 
      || id() == vc_id() || node->primary(view()) == id())
    return false;

  if (view() <= 0 || !node->is_replica(vc_id()))
    return false;

  // Check sizes
  if (size()-(int)sizeof(View_change_ack) < MAC_size) 
    return false;

  // Check MAC.
  Principal *p = node->i_to_p(id());
  int old_size = sizeof(View_change_ack_rep);
  if (!p->verify_mac_in(contents(), old_size, contents()+old_size))
    return false;

  return true;
}


bool View_change_ack::convert(Message *m1, View_change_ack  *&m2) {
  if (!m1->has_tag(View_change_ack_tag, sizeof(View_change_ack_rep)))
    return false;

  m2 = (View_change_ack*)m1;
  m2->trim();
  return true;
}
