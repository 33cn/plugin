#include "Replica.h"
#include "Reply_stable.h"
#include "Stable_estimator.h"
#include "K_max.h"

Stable_estimator::Stable_estimator() {
  nv = node->n();
  vals = new Val[nv];
  est = -1;
}


Stable_estimator::~Stable_estimator() {
  delete [] vals;
}


bool Stable_estimator::add(Reply_stable* m, bool mine) {
  if (mine || m->verify()) {
    const int id = m->id();
    const int lc = m->last_checkpoint();
    const int lp = m->last_prepared();

    Val& val = vals[id];
    Seqno oc = Seqno_max;
    Seqno op = -1;

    if (lc < val.lc) {
      oc = val.lc;
      val.lc = lc;
      val.lec = 1;
      val.gep = 1;
    }

    if (val.lp < lp) {
      op = val.lp;
      val.lp = lp;
    }

    const int nge = node->f()+1;
    const int nle = 2*node->f()+1;
    for (int i=0; i < nv; i++) {
      if (i == id) continue;

      Val &v = vals[i];

      if ((oc > v.lc) && (lc <= v.lc))
	v.lec++;
	
      if (lc >= v.lc)
	val.lec++;

      if ((op < v.lc) && (lp >= v.lc))
	v.gep++;
      
      if (lc <= v.lp)
	v.gep++;
    
      if ((v.lec >= nle) && (v.gep >= nge)) {
	est = v.lc;
	break;
      }
    }

    if ((est < 0) && (val.lec >= nle) && (val.gep >= nge)) 
      est = val.lc;      
  }
  delete m;
  return est >= 0;
}


void Stable_estimator::mark_stale() {
  for (int i=0; i < node->n(); i++) {
    vals[i].clear();
  }
}

  
void Stable_estimator::clear() {
  mark_stale();
  est = -1;
}

Seqno Stable_estimator::low_estimate() {
  Seqno lps[Max_num_replicas];

  for (int i=0; i < node->n(); i++) {
    lps[i] = vals[i].lp;
  }

  Seqno mlp = K_max(node->f()+1, lps, node->n(), Seqno_max);
  return (mlp-max_out+checkpoint_interval-2)/checkpoint_interval * checkpoint_interval;
}
