#ifndef _Log_h
#define _Log_h 1

#include "types.h"
#include "parameters.h"

template <class T> class Log {
  //
  // Log of T ordered by sequence number. 
  //
  // Requires that "t" has a method:
  // void clear();

public:
  Log(int sz=max_out, Seqno h=1);
  // Requires: "sz" is a power of 2 (allows for more efficient implementation).
  // Effects: Creates a log that holds "sz" elements and has
  // head equal to "h". The log only maintains elements with sequence
  // number higher than "head" and lower than "tail" = "head"+"max_size"-1 

  ~Log();
  // Effects: Delete log and all associated storage.

  void clear(Seqno h);
  // Effects: Calls "clear" for all elements in log and sets head to "h"

  T &fetch(Seqno seqno);
  // Requires: "within_range(seqno)"
  // Effects: Returns the entry corresponding to "seqno".

  void truncate(Seqno new_head);
  // Effects: Truncates the log clearing all elements with sequence
  // number lower than new_head.

  bool within_range(Seqno seqno) const;
  // Effects: Returns true iff "seqno" is within range.
  
  Seqno head_seqno() const;
  // Effects: Returns the sequence number for the head of the log.

  Seqno max_seqno() const;
  // Effects: Returns the maximum sequence number that can be
  // stored in the log.

protected:
  unsigned mod(Seqno s) const;
  // Effects: Computes "s" modulo the size of the log.

  Seqno head;
  int max_size;
  T *elems;
  Seqno mask;
};

template <class T>
inline unsigned Log<T>::mod(Seqno s) const { return s & mask; }

template <class T>
inline bool Log<T>::within_range(Seqno seqno) const { 
  return seqno >= head && seqno < head + max_size; 
}
  
template <class T>
inline Seqno Log<T>::head_seqno() const {
  return head; 
}

template <class T>
inline Seqno Log<T>::max_seqno() const {
  return head+max_size-1;
}


#endif // _Log_h
