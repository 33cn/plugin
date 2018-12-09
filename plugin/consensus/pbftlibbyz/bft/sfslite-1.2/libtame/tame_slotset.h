// -*-c++-*-
/* $Id: tame_valueset.h 2673 2007-04-03 04:59:39Z max $ */

#ifndef _LIBTAME_TAME_SLOTSET_H_
#define _LIBTAME_TAME_SLOTSET_H_

template <typename T1=void, typename T2=void, typename T3=void, 
	  typename T4=void>
struct _tame_slot_set {
  _tame_slot_set(T1 *p1_, T2 *p2_, T3 *p3_, T4 *p4_) 
    : p1(p1_), p2(p2_), p3(p3_), p4(p4_) { }
  void assign(const T1 &v1, const T2 &v2, const T3 &v3, const T4 &v4) {
    *p1 = v1; *p2 = v2; *p3 = v3; *p4 = v4;
  }
  T1 *p1;
  T2 *p2;
  T3 *p3;
  T4 *p4;
};

template <typename T1, typename T2, typename T3>
struct _tame_slot_set<T1, T2, T3, void> {
    _tame_slot_set(T1 *p1_, T2 *p2_, T3 *p3_) : p1(p1_), p2(p2_), p3(p3_) { }
    void assign(const T1 &v1, const T2 &v2, const T3 &v3) {
	*p1 = v1; *p2 = v2; *p3 = v3;
    }
    T1 *p1;
    T2 *p2;
    T3 *p3;
};

template <typename T1, typename T2>
struct _tame_slot_set<T1, T2, void, void> {
    _tame_slot_set(T1 *p1_, T2 *p2_) : p1(p1_), p2(p2_) { }
    void assign(const T1 &v1, const T2 &v2) {
	*p1 = v1; *p2 = v2;
    }
    T1 *p1;
    T2 *p2;
};

template <typename T1>
struct _tame_slot_set<T1, void, void, void> {
    _tame_slot_set(T1 *p1_) : p1(p1_) { }
    void assign(const T1 &v1) {
	*p1 = v1;
    }
    T1 *p1;
};

template <>
struct _tame_slot_set<void, void, void, void> {
    _tame_slot_set() { }
    void assign() { }
};


struct nil_t {};

#endif /* _LIBTAME_TAME_SLOTSET_H_ */
