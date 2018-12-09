#ifndef _Certificate_h
#define _Certificate_h 1

#include <sys/time.h>
#include "types.h"
#include "Time.h"
#include "parameters.h"
#include "Bitmap.h"

template <class T> class Certificate {
  //
  // A certificate is a set of "matching" messages from different
  // replicas. 
  //
  // T must have the following methods:
  // bool match(T*);
  // // Effects: Returns true iff the messages match
  //
  // int id();
  // // Effects: Returns the identifier of the principal that
  // // sent the message.
  //
  // bool verify();
  // // Effects: Returns true iff the message is properly authenticated
  // // and statically correct.
  //
  // bool full();
  // // Effects: Returns true iff the message is full
  //
  // bool encode(FILE* o);
  // bool decode(FILE* i);
  // Effects: Encodes and decodes object state from stream. Return
  // true if successful and false otherwise.
  
public:
  Certificate(int complete=0);
  // Requires: "complete" >= f+1 or 0
  // Effects: Creates an empty certificate. The certificate is
  // complete when it contains at least "complete" matching messages
  // from different replicas. If the complete argument is omitted (or
  // 0) it is taken to be 2f+1.

  ~Certificate();
  // Effects: Deletes certificate and all the messages it contains.
 
  bool add(T *m);
  // Effects: Adds "m" to the certificate and returns true provided
  // "m" satisfies:
  // 1. there is no message from "m.id()" in the certificate
  // 2. "m->verify() == true"
  // 3. if "cvalue() != 0", "cvalue()->match(m)";
  // otherwise, it has no effect on this and returns false.  This
  // becomes the owner of "m" (i.e., no other code should delete "m"
  // or retain pointers to "m".)

  bool add_mine(T *m);
  // Requires: The identifier of the calling principal is "m->id()"
  // and "mine()==0" and m is full.
  // Effects: If "cvalue() != 0" and "!cvalue()->match(m)", it has no
  // effect and returns false. Otherwise, adds "m" to the certificate
  // and returns. This becomes the owner of "m"

  T *mine(Time **t=0);
  // Effects: Returns caller's message in certificate or 0 if there is
  // no such message. If "t" is not null, sets it to point to the time
  // at which I last sent my message.

  T *cvalue() const;
  // Effects: Returns the correct message value for this certificate
  // or 0 if this value is not known. Note that the certificate
  // retains ownership over the returned value (e.g., if clear or
  // mark_stale are called the value may be deleted.)

  T *cvalue_clear();
  // Effects: Returns the correct message value for this certificate
  // or 0 if this value is not known. If it returns the correct value,
  // it removes the message from the certificate and clears the
  // certificate (that is the caller gains ownership over the returned
  // value.)

  int num_correct() const;
  // Effects: Returns the number of messages with the correct value
  // in this.

  bool is_complete() const;
  void make_complete();
  // Effects: If cvalue() is not null, makes the certificate
  // complete.
    
  void mark_stale();
  // Effects: Discards all messages in certificate except mine. 

  void clear(); 
  // Effects: Discards all messages in certificate

  bool is_empty() const;
  // Effects: Returns true iff the certificate is empty

  class Val_iter {
    // An iterator for yielding all the distinct values in a
    // certificate and the number of messages matching each value. The
    // certificate cannot be modified while it is being iterated on.
  public:
    Val_iter(Certificate<T>* c);
    // Effects: Return an iterator for the values in "c"
	
    bool get(T*& m, int& count);
    // Effects: Updates "m" to point to the next value in the
    // certificate and count to contain the number of messages
    // matching this value and returns true. If there are no more
    // values, it returns false.

  private:
    Certificate<T>* cert; 
    int next;
  };
  friend  class Val_iter;

  bool encode(FILE* o);
  bool decode(FILE* i);
  // Effects: Encodes and decodes object state from stream. Return
  // true if successful and false otherwise.

private:
  Bitmap bmap; // bitmap with replicas whose message is in this.

  class Message_val {
  public:
    T *m;
    int count; 
    
    inline Message_val() { m = 0; count = 0; }
    inline void clear() { 
      delete m; 
      m = 0;
      count = 0;
    }
    inline ~Message_val() { clear(); }
    
  };
  Message_val *vals;    // vector with all distinct message values in this
  int max_size;         // maximum number of elements in vals, f+1
  int cur_size;         // current number of elements in vals

  int correct;          // value is correct if it appears in at least "correct" messages
  Message_val *c;       // correct certificate value or 0 if unknown.

  int complete;         // certificate is complete if "num_correct() >= complete"

  T *mym; // my message in this or null if I have no message in this 
  Time t_sent; // time at which mym was last sent

  // The implementation assumes:
  // correct > 0 and complete > correct 
};

template <class T> 
inline T *Certificate<T>::mine(Time **t) { 
  if (t && mym) *t = &t_sent;
  return mym;
}

template <class T> 
inline T *Certificate<T>::cvalue() const { return (c) ? c->m : 0; }

template <class T> 
inline int Certificate<T>::num_correct() const {  return (c) ? c->count : 0; }

template <class T>
inline bool Certificate<T>::is_complete() const { return num_correct() >= complete; }

template <class T>
inline  void Certificate<T>::make_complete() {
  if (c) {
    c->count = complete;
  }
}


template <class T>
inline bool Certificate<T>::is_empty() const {
  return bmap.all_zero();
}


template <class T>
inline void Certificate<T>::clear() {
  for (int i=0; i < cur_size; i++) vals[i].clear(); 
  bmap.clear();
  c = 0;
  mym = 0;
  t_sent = 0;
  cur_size = 0;
}


template <class T>
inline Certificate<T>::Val_iter::Val_iter(Certificate<T>* c) {
  cert = c;
  next = 0;
}
	

template <class T>
inline bool Certificate<T>::Val_iter::get(T*& m, int& count) {
  if (next < cert->cur_size) {
    m = cert->vals[next].m;
    count = cert->vals[next].count;
    next++;
    return true;
  } 
  return false;
}

#endif // Certificate_h
  










