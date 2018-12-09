#include "Node.h"
#include "Prepared_cert.h"

#include "Certificate.t"
template class Certificate<Prepare>;

Prepared_cert::Prepared_cert() : pc(node->f()*2), primary(false) {}


Prepared_cert::~Prepared_cert() { pi.clear(); }


bool Prepared_cert::is_pp_correct() {
  if (pi.pre_prepare()) {
    Certificate<Prepare>::Val_iter viter(&pc);
    int vc;
    Prepare* val;
    while (viter.get(val, vc)) {
      if (vc >= node->f() && pi.pre_prepare()->match(val)) {
	return true;
      } 
    }
  }
  return false;
}
  

bool Prepared_cert::add(Pre_prepare *m) {
  if (pi.pre_prepare() == 0) {
    Prepare* p = pc.mine();
    
    if (p == 0) {
      if (m->verify()) {
	pi.add(m);
	return true;
      }

      if (m->verify(Pre_prepare::NRC)) { 
	// Check if there is some value that matches pp and has f
	// senders.
	Certificate<Prepare>::Val_iter viter(&pc);
	int vc;
	Prepare* val;
	while (viter.get(val, vc)) {
	  if (vc >= node->f() && m->match(val)) {
	    pi.add(m);
	    return true;
	  } 
	}
      }
    } else {
      // If we sent a prepare, we only accept a matching pre-prepare.
      if (m->match(p) && m->verify(Pre_prepare::NRC)) {
	pi.add(m);
	return true;
      }
    }
  }
  delete m;
  return false;
}

  
bool Prepared_cert::encode(FILE* o) {
  bool ret = pc.encode(o);
  ret &= pi.encode(o);
  int sz = fwrite(&primary, sizeof(bool), 1, o);
  return ret & (sz == 1);
}
  
  
bool Prepared_cert::decode(FILE* i) {
  th_assert(pi.pre_prepare() == 0, "Invalid state");
  
  bool ret = pc.decode(i);
  ret &= pi.decode(i);
  int sz = fread(&primary, sizeof(bool), 1, i);
  t_sent = zeroTime();
  
  return ret & (sz == 1);
}


bool Prepared_cert::is_empty() const {
  return pi.pre_prepare() == 0 && pc.is_empty();
}
