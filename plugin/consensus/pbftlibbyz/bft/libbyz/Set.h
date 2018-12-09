#ifndef _Set_h
#define _Set_h 1

#include "types.h"

template <class T> class Set {
  //
  // Set of T*. Type T must have an "int id()" method which is used 
  // as the set's key. Furthermore, the  returned identifiers i must be
  // within the range 0 <= i < sz (where sz is the maximum number of 
  // elements in the set determined at creation time).
  //
public:
  Set(int sz);
  // Effects: Creates a set that can hold up to "sz" elements.

  ~Set();
  // Effects: Delete set and all associated storage.

  bool store(T *e);
  // Effects: Adds "e" to the set (if it has a valid id not in the
  // set) and returns true. Otherwise, returns false and does nothing.

  T* fetch(int id);
  // Effects: returns a pointer to the element in the set with 
  // identifier "id" or 0 if there is no such element.

  T* remove(int id);
  // Effects: Removes the element with identifier "id" from the set 
  // and returns a pointer to it.

  void clear();
  // Effects: Removes all elements from the set and deletes them.

  int size() const;
  // Effects: Returns the number of elements in the set.


private:
  int max_size;
  T **elems;
  int cur_size;
};

template <class T>
inline int Set<T>::size() const { return cur_size; }

#endif // _Set_h
