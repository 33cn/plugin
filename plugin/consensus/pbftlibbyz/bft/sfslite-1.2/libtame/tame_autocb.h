
// -*-c++-*-
/* $Id: tame.h 2077 2006-07-07 18:24:23Z max $ */

#ifndef _LIBTAME_AUTOCB_H_
#define _LIBTAME_AUTOCB_H_

#include "async.h"
#include "callback.h"

template<class B1=void, class B2=void, class B3=void, class B4=void> 
class autocb_t;

template<>
class autocb_t<> {
public:
  autocb_t (callback<void>::ref cb) : _cb (cb), _cancelled (false) {}

  ~autocb_t () { if (!_cancelled) TRIGGER (_cb); }
  void cancel () { _cancelled = true; }
private:
  callback<void>::ref _cb;
  bool _cancelled;
};

template<class B1>
class autocb_t<B1> {
public:
  autocb_t (ref<callback<void, B1> > cb, B1 &b1) 
    : _cb (cb), _b1 (b1), _cancelled (false) {}

  ~autocb_t () { if (!_cancelled) TRIGGER (_cb, _b1); }
  void cancel () { _cancelled = true; }
private:
  ref<callback<void, B1> > _cb;
  B1 &_b1;
  bool _cancelled;
};

template<class B1, class B2>
class autocb_t<B1,B2> {
public:
  autocb_t (ref<callback<void, B1, B2> > cb, B1 &b1, B2 &b2) 
    : _cb (cb), _b1 (b1), _b2 (b2), _cancelled (false) {}

  ~autocb_t () { if (!_cancelled) TRIGGER (_cb, _b1, _b2); }
  void cancel () { _cancelled = true; }
private:
  ref<callback<void, B1, B2> > _cb;
  B1 &_b1;
  B2 &_b2;
  bool _cancelled;
};

template<class B1, class B2, class B3>
class autocb_t<B1,B2,B3> {
public:
  autocb_t (ref<callback<void, B1, B2, B3> > cb, B1 &b1, B2 &b2, B3 &b3) 
    : _cb (cb), _b1 (b1), _b2 (b2), _b3 (b3), _cancelled (false) {}

  ~autocb_t () { if (!_cancelled) TRIGGER (_cb, _b1, _b2, _b3); }
  void cancel () { _cancelled = true; }
private:
  ref<callback<void, B1, B2, B3> > _cb;
  B1 &_b1;
  B2 &_b2;
  B3 &_b3;
  bool _cancelled;
};

#endif /* _LIBAME_AUTOCB_H_ */
