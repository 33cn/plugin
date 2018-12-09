#include <string.h>

#include "th_assert.h"
#include "Message_tags.h"
#include "New_view.h"
#include "Replica.h"
#include "Principal.h"
 

New_view::New_view(View v) : Message(New_view_tag, Max_message_size) {
  rep().v = v;
  rep().min = -1;
  rep().max = -1;

  // Initialize vc_info
  for (int i=0; i < node->n(); i++) {
    vc_info()[i].d.zero();
  }
}


void  New_view::add_view_change(int id, Digest &d) {
  th_assert(node->is_replica(id), "Not a replica");
  th_assert(vc_info()[id].d == Digest(), "Duplicate");

  VC_info& vci =  vc_info()[id];
  vci.d = d;
}


void New_view::set_min(Seqno min) {
  th_assert(rep().min == -1, "Invalid state");
  rep().min = min;
}

void New_view::set_max(Seqno max) {
  th_assert(min() >= 0, "Invalid state");
  rep().max = max;
  th_assert(max >= min() && max-min() <= max_out+1, "Invalid arguments");
}

void New_view::pick(int id, Seqno n) {
  th_assert(min() >= 0, "Invalid state");
  th_assert(node->is_replica(id), "Not a replica");
  th_assert(vc_info()[id].d != Digest(), "Invalid argument");
  th_assert(n >= min() && n <= min()+max_out, "Invalid argument");

  picked()[n-min()] = id;
}


void New_view::re_authenticate(Principal *p) {
  int old_size = sizeof(New_view_rep)+sizeof(VC_info)*node->n()+max()-min();

  // Compute authenticator and update size. 
  th_assert(Max_message_size-old_size >= node->auth_size(), "Message is too small");
  set_size(old_size+node->auth_size());
  node->gen_auth_out(contents(), old_size, contents()+old_size);
  trim();
}


bool New_view::view_change(int id, Digest& d) {
  if (id < 0  || id >= node->n())
    return false;

  VC_info& vci = vc_info()[id];
  if (vci.d.is_zero())
    return false;

  d = vci.d;

  return true;
}


bool New_view::verify() {
  if (view() <= 0 || min() < 0 || max() < 0 || max() < min() || max()-min() > max_out+1)
    return false;

  // Check that each entry in picked is set to the identifier of a replica
  // whose view-change digest is in this.
  for (Seqno i = min(); i < max(); i++) {
    int vci = picked()[i-min()];
    if (!node->is_replica(vci) || vc_info()[vci].d.is_zero())
      return false;
  }

  int old_size = sizeof(New_view_rep) + sizeof(VC_info)*node->n()+max()-min();
  if (Max_message_size-old_size < node->auth_size(id()))
    return false;

  // Check authenticator
  if (!node->verify_auth_in(id(), contents(), old_size,  contents()+old_size))
    return false;

  return true;
}


bool New_view::convert(Message *m1, New_view  *&m2) {
  if (!m1->has_tag(New_view_tag, sizeof(New_view_rep)))
    return false;

  m1->trim();
  m2 = (New_view*)m1;
  return true;
}
