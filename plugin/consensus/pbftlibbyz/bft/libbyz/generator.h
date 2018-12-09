#ifndef _GENERATOR_H
#define _GENERATOR_H

#pragma interface

#include "basic.h"

template <class T>
class Generator {
  public:
	Generator() {}
    virtual bool get(T&) = 0;
    virtual ~Generator() {};
};

#endif /* _GENERATOR_H */
