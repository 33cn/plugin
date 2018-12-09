#include "Pre_prepare.h"
#include "Pre_prepare_info.h"
#include "Replica.h"

Pre_prepare_info::~Pre_prepare_info() {
  delete pp;
}


void Pre_prepare_info::add(Pre_prepare* p) {
  th_assert(pp == 0, "Invalid state");
  pp = p;
  mreqs = p->num_big_reqs();
  mrmap = 0;
  Big_req_table* brt = replica->big_reqs();
  
  for (int i=0; i < p->num_big_reqs(); i++) {
    if (brt->add_pre_prepare(p->big_req_digest(i), i, p->seqno(), p->view())) {
      mreqs--;
      Bits_set((char*)&mrmap, i);
    }
  }
}


void Pre_prepare_info::add(Digest& rd, int i) {
  if (pp && pp->big_req_digest(i) == rd) {
    mreqs--;  
    Bits_set((char*)&mrmap, i);
  }
}

 
bool Pre_prepare_info::encode(FILE* o) {
  bool hpp = pp != 0;
  size_t sz = fwrite(&hpp, sizeof(bool), 1, o);
  bool ret = true;
  if (hpp)
    ret = pp->encode(o);
  return ret & (sz == 1);
}
  

bool Pre_prepare_info::decode(FILE* i) {
  bool hpp;
  size_t sz = fread(&hpp, sizeof(bool), 1, i);
  bool ret = true;
  if (hpp) {
    pp = (Pre_prepare*) new Message;
    ret &= pp->decode(i);
  }
  return ret & (sz == 1);
}


Pre_prepare_info::BRS_iter::BRS_iter(Pre_prepare_info const *p, BR_map m) {
  ppi = p;
  mrmap = m;
  next = 0;
}
    
bool Pre_prepare_info::BRS_iter::get(Request*& r) {
  Pre_prepare* pp = ppi->pp;
  while (pp && next < pp->num_big_reqs()) {
    if (!Bits_test((char*)&mrmap, next) & Bits_test((char*)&(ppi->mrmap), next)) {
      r = replica->big_reqs()->lookup(pp->big_req_digest(next));
      th_assert(r != 0, "Invalid state");
      next++;
      return true;
    }
    next++;
  }
  return false;
}
