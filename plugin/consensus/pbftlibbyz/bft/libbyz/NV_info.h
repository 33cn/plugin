#ifndef _NV_info_h
#define _NV_info_h 1

#include <sys/time.h>
#include "Bitmap.h"
#include "parameters.h"
#include "types.h"
#include "Time.h"
#include "Digest.h"
#include "Pre_prepare_info.h"
#include "Array.h"

class View_info;
class View_change;
class New_view;
class Pre_prepare;
class View_change_ack;
class Status;
class Pre_prepare;
class Prepare;

class NV_info {
  //
  // Holds information concerning a specific new-view message.
  //
public:
  NV_info();
  // Effects: Creates an empty object.
    
  ~NV_info();
  // Effects: Discards any new-view, view-change and view-change ack
  // messages stored in this.
    
  void clear();
  // Effects: Makes this empty -- deletes all contained messages
  // and sets view() == 0.

  View_change *mark_stale(int id);
  // Effects: If "is_complete()", does nothing. Otherwise, deletes
  // all messages contained in this and sets view() == 0. Except
  // that it does not delete and it returns any view-change message
  // from replica "id".
  
  New_view *new_view(Time **t=0);
  // Effects: If there is a new-view message stored in "this"
  // returns it. Otherwise, returns 0.

  View_change *view_change(int id);
  // Effects: If there is any view-change message from replica "id"
  // stored in this returns it. Otherwise, returns 0.

  View view() const;
  // Effects: Returns the view number of the messages currently
  // stored on this or 0 if this is empty (i.e., if it does not
  // contain any new-view message.)

  bool complete() const;
  // Effects: Returns true iff this contains all the necessary
  // information to move to the new-view.

  void make_complete(View v);
  // Effects: Mark this as complete in view "v".

  bool add(New_view* nv, View_info *parent);
  // Requires: "nv.verify() || node->id() == node->primary(nv->view())"
  // Effects: If "nv->view() <= view()", it does not modify this and
  // deletes "nv".  Otherwise, it adds "nv" to this and if "view()
  // != 0", it deletes any new-view and view-change messages stored
  // in this. Returns true iff it adds "nv" to this.
  
  bool add(View_change* vc, bool verified);
  // Requires:  "vc->view() == view()" and "verified == vc->verify()"
  // Effects: If "vc" is one of the messages referenced in the
  // new-view message contained in this and is valid , add "vc" to
  // this and return true. Otherwise, do nothing and return false.

  bool add(View_change_ack* vca);
  // Requires: "vca->view() == view() && vca.verify()"
  // Effects: If there is no view-change corresponding to "vca"
  // in this or referenced by a new-view in this, returns false.
  // Otherwise, it inserts "vca" in this (if its digest matches the
  // view-change's) or deletes vca (otherwise) and returns true.

  Pre_prepare* fetch_request(Seqno n, Digest &d);
  // Requires: "complete" and "n" is the sequence number of a request
  // in new-view message.
  // Effects: Sets "d" to the digest of the request with sequence
  // number "n". If enough information to make a pre-prepare is available, it 
  // returns an appropriate pre-prepare. Otherwise, returns zero. The
  // caller is responsible for deallocating any returned pre-prepare.

  void add_missing(Pre_prepare* pp);
  // Effects: Checks if "pp" is a pre-prepare that is needed to
  // complete a view-change. If it is stores "pp", otherwise deletes
  // "pp".

  void add_missing(Digest& rd, Seqno n, int i);
  // Effects: Records that the big request with digest "rd" that is
  // referenced by a pre-prepare with sequence number "n" as the i-th
  // big request is cached.

  void add_missing(Prepare* p);
  // Effects: Checks if "p" is a prepare that is needed to
  // complete a view-change. If it is stores "pp", otherwise deletes "pp".

  void set_received_vcs(Status *m);
  // Effects: Mutates "m" to record which view change messages were
  // accepted in the current view or are not needed to complete new-view.

  void set_missing_pps(Status *m);
  // Effects: Mutates "m" to record which pre-prepares are missing

  void mark_stable(Seqno ls);
  // Effects: Informs "this" that checkpoint sequence number "ls" is
  // stable.

private:
  View v;
  New_view *nv;
  int vc_target; // Number of view-change messages in nv
  int vc_cur;    // Number of view-change messages already matched with target.
  
  // Buffer for view changes associated with "nv" and their acks.
  struct VC_info {
    View_change* vc;
    int ack_count;
    Bitmap ack_reps;
    bool req_sum;
    
    VC_info();
    ~VC_info();
    void clear();
  }; 
  Array<VC_info> vcs;

  //
  // Data structures to check the correctness and completness of
  // new-view messages.
  //
  struct Ckpt_sum {
    Seqno n;
    Digest d;
    int n_proofs;
    int n_le;
    Seqno max_seqno;
    int id; // identifier of first replica proposing this.
  };

  Array<Ckpt_sum> ckpts; // Array of candidate checkpoints
  int chosen_ckpt;       // Index of chosen checkpoint (-1 if no checkpoint chosen.)

  Seqno min;             // Sequence number of chosen checkpoint
  Seqno max;             // All requests that will propagate to the next view have 
                         // sequence number less than max.
  Seqno base;            // reqs and comp_reqs are indexed by sequence number minus base.

  struct Req_sum {
    View v;
    Digest d;
    int n_proofs;
    int n_pproofs;       // Number of positive proofs.
    Bitmap r_pproofs;    // Replicas that sent positive proofs.
    Pre_prepare_info pi;
    int n_le; 
    int id; // identifier of first replica proposing this.
    Req_sum(); 
    ~Req_sum();
  };
  Array<Array<Req_sum> > reqs; // reqs contains an array for each sequence number 
                               // above min an less than or equal to max.
  Array<int> comp_reqs; // For each row index in reqs, contains either -1 
                        // if no request with complete information for that 
                        // index or the column index of the complete Req_sum.  
  int n_complete;       // Number of complete entries with seqnos between min and max.

  
  View_info *vi; // Pointer to parent view-info.

  bool is_complete; // True iff this contains all the necessary information for a new-view

  Time nv_sent; // time at which my new-view was last sent

  //
  // Auxiliary methods:

  void summarize(View_change *vc);
  // Effects: Summarizes the information in "vc" and combine it with
  // information from other view-change messages.
  
  void summarize_reqs(View_change *vc);
  // Effects: Summarize the request information in "vc" and combine it
  // with information from other view-change messages. 
  
  void choose_ckpt(int index);
  // Requires: "c = sum->ckpts[index]" has n_proofs and n_le greater
  // than or equal to "node->n_f()" 
  // Effects: If "c" has a higher max_stable, than  sum->ckpts[chosen]
  // make it the new chosen checkpoint. Otherwise, do nothing.
  
  void check_comp(Req_sum& cur, Seqno i, int j);
  // Requires: !complete()
  // Effects: Checks if the request corresponding to cur has enough information to 
  // be propagated to the next view. If it does and complete() becomes true, calls
  // make_new_view.

  void make_new_view();
  // Requires: complete() 
  // Effects: Completes the construction of the new-view message.

  bool check_new_view();
  // Effects: Checks if this contains a correct new-view and if so
  // returns true.

  void get_proofs(Req_sum& cur, View_change *vc, Seqno i);
  // Effects: Update proofs for cur with invormation in "vc" for
  // sequence number "i".

  Seqno known_stable();
  // Effects: Computes the maximum sequence number that is known to be
  // stable.

};


inline View NV_info::view() const {
  return v;
}

inline bool NV_info::complete() const {
  return is_complete;
}

inline New_view *NV_info::new_view(Time **t) {
  if (t && nv)
    *t = &nv_sent;
  return nv;
}

inline  View_change *NV_info::view_change(int id) {
  return vcs[id].vc;
}



#endif // _NV_info_h
