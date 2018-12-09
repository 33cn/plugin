#ifndef _VALUEKEY_H
#define _VALUEKEY_H

#define DECLARE_value_key(CLASS,T,HASH)					      \
/*									      \
									      \
    DECLARE_value_key(name CLASS, type T, int HASH(T));			      \
									      \
    Declare a class that can act as a key for the "map" class (see	      \
    "map.h".) The class created wraps an existing type and adds the	      \
    "hash" method required of a map key. It is particularly convenient	      \
    for turning primitive types into keys.				      \
									      \
    Note that HASH may be the name of a macro.				      \
									      \
    type T where T {							      \
	    T();							      \
	    T(T const &);						      \
	    void operator=(T const &);					      \
	    bool operator==(T const &);					      \
	}								      \
									      \
    int HASH(T);							      \
 */									      \
class CLASS {								      \
public:									      \
    CLASS() {} /* need this for use with generators, annoyingly	*/	      \
    CLASS(T a) : val(a) {}						      \
    void operator=(CLASS const &x) { val = x.val; }			      \
    int hash() const { return HASH(val); }				      \
    bool operator==(CLASS const &x)					      \
	{ return (x.val == val) ? true : false; }			      \
    T val;								      \
};									      \


#define DECLARE_value_elem(CLASS,T,HASH)				      \
/*									      \
									      \
    DECLARE_value_elem(name CLASS, type T, int HASH(T))			      \
									      \
    Declare a class that can act as an element for the set class ("set.h")    \
    "map.h".) The class created wraps an existing type and adds the	      \
    "similar" and "hash" methods required of a set element. It is useful      \
    for turning primitive types into keys.				      \
									      \
    Note that HASH may be the name of a macro.				      \
									      \
    type T where T {							      \
	    T();							      \
	    T(T const &);						      \
	    void operator=(T const &);					      \
	    bool operator==(T const &);					      \
	}								      \
									      \
    int HASH(T);							      \
 */									      \
class CLASS {								      \
public:									      \
    CLASS() {} /* need this for use with generators, annoyingly */	      \
    CLASS(T a) : val(a) {}						      \
    void operator=(CLASS const &x) { val = x.val; }			      \
    int hash() const { return HASH(val); }				      \
    bool operator==(CLASS const &x)					      \
	{ return (x.val == val) ? true : false; }			      \
    bool similar(CLASS const &x)					      \
	{ return (x.val == val) ? true : false; }			      \
    T val;								      \
};									      \


/*
    The commonly used keys "IntKey", "UIntKey", and "LongKey" are
    provided here.
*/
#define ident_hash(x) ((int)(x));

DECLARE_value_key(IntKey, int, ident_hash);
DECLARE_value_key(LongKey, int, ident_hash);
DECLARE_value_key(UIntKey, unsigned int, ident_hash);

DECLARE_value_elem(IntElem, int, ident_hash);
DECLARE_value_elem(LongElem, int, ident_hash);
DECLARE_value_elem(UIntElem, unsigned int, ident_hash);

#endif /* _VALUEKEY_H */
