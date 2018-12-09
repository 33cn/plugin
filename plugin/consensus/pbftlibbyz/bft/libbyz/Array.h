/*
\section{Generic Resizable Array}

"Array.h" provides a generic resizable array. These arrays can be
stack allocated as long as you are careful that they do not get copied
excessively.  (One reason for stack allocation of arrays is so that
you can use the nice [] syntactic sugar.)

Arrays grow and shrink at the high end.  The low end always has index 0.

// \subsection{Constructor Specifications}


    Array()
	effects  - Creates empty array

    Array(int predict)
	requires - predict >= 0
	effects  - Creates empty array.  Extra storage is preallocated
		   under the assumption that the array will be grown
		   to contain predict elements.  This constructor form
		   helps avoids unnecessary copying.

    Array(T const* x, int count)
	requires - count >= 0 and x is a pointer to at least count elements.
	effects  - creates an array with a copy of the count elements pointed
		   to by x.  This constructor is useful for initializing an
		   array from a builtin array.

    Array(Array const& x)
	effects  - creates an array that is a copy of x.

    Array(T x, int count)
	requires - count >= 0
	effects  - creates an array with count copies of x.
    */

// \subsection{Destructor Specifications}

    /*
    ~Array()
	effects  - releases all storage for the array.
    */

// \subsection{Operation Specifications}

    /*
    Array& operator=(Array const& x)
	effects  - copies the contents of x into *this.

    T& operator[](int index) const
	requires - index >= 0 and index < size().
	effects  - returns a reference to the indexth element in *this.
		   Therefore, you can say things like --
			Array a;
			...
			a[i] = a[i+1];

    T& slot(int index) const
	requires - index >= 0 and index < size().
	effects  - returns a reference to the indexth element in *this.
		   This operation is identical to operator [].  It is
		   just more convenient to use with an Array*.

			Array* a;
			...
			a->slot(i) = a->slot(i+1);

    int  size() const
	effects	 - returns the number of elements in *this.

    T& high() const
	effects  - returns a reference to the last element in *this.

    void append(T v)
	modifies - *this
	effects  - grows array by one by adding v to the high end.

    void append(T v, int n)
	requires - n >= 0
	modifies - *this
	effects  - grows array by n by adding n copies of v to the high end.

    void concat(T const* x, int n)
	requires - n >= 0, x points to at least n Ts.
	modifies - *this
	effects	 - grows array by n by adding the n elements pointed to by x.

    void concat(Array const& x)
	modifies - *this
	effects  - append the contents of x to *this.

    T remove()
	requires - array is not empty
	modifies - *this
	effects  - removes last element and return a copy of it.

    void remove(int num)
	requires - num >= 0 and array has at least num elements
	modifies - *this
	effects  - removes the last num elements.

    void clear()
	modifies - *this
	effects  - removes all elements from *this.

    void reclaim()
	effects  - reclaim unused storage.

    T* as_pointer() const
	requires - returned value is not used across changes in array size
		   or calls to reclaim.
	effects  - returns a pointer to the first element in the array.
		   The returned pointer is useful for interacting with
		   code that manipulates builtin arrays of T.

    void predict(int new_alloc);
         effects - Does not change the abstract state. If the allocated 
                   storage is smaller than new_alloc elements, enlarges the
                   storage to new_alloc elements.
 
    void _enlarge_by(int n);
	requires - n >= 0
	effects  - appends "n" UNITIALIZATED entries to the array.
		   This is an unsafe operation that is mostly useful
		   when reading the contents of an array over the net.
		   Use it carefully.

    */

#ifndef _Array_h
#define _Array_h 1

#include "th_assert.h"

template <class T> class Array {
  public:
								
    /* Constructors */						
    Array();			/* Empty array */	
    Array(int predict);		/* Empty array with size predict */
    Array(T const*, int);	/* Initialized with C array */	  
    Array(Array const&);	/* Initialized with another Array */
    Array(T, int);		/* Fill with n copies of T */ 
									    
    /* Destructor */							    
    ~Array();							    
									    
    /* Assignment operator */						    
    Array& operator=(Array const&);				    
									    
    /* Array indexing */						    
    T& operator[](int index) const;				   
    T& slot(int index) const;					   
									   
    /* Other Array operations */				
    int  size() const;			/* Return size; */	
    T& high() const;		/* Return last T */	
    T* as_pointer() const;	/* Return as pointer to base */	
    void append(T v);		/* append an T */		
    void append(T, int n);	/* Append n copies of T */ 
    void concat(T const*, int);	/* Concatenate C array */	 
    void concat(Array const&);	/* Concatenate another Array */	     
    T remove();			/* Remove and return last T */
    void remove(int num);		/* Remove last num Ts */	    
    void clear();			/* Remove all Ts */	    
    void predict(int new_alloc);        /* Increase allocation */	    
									    
									    
    /* Storage stuff */							    
    void reclaim();			/* Reclaim all unused space */	    
    void _enlarge_by(int n);		/* Enlarge array by n */	    
  private:								    
    T*	store_;			/* Actual storage */		    
    int		alloc_;			/* Size of allocated storage */	    
    int		size_;			/* Size of used storage */	    
									    
    /* Storage enlargers */						    
    void enlarge_allocation_to(int s);	/* Enlarge to s */		    
    void enlarge_to(int s);		/* Enlarge to s if necessary */	    
};									    

template <class T>
inline Array<T>::Array() {						      
    alloc_ = 0;								      
    size_  = 0;								      
    store_ = 0;								      
}
									      
template <class T>								      
inline int Array<T>::size() const {					      
    return size_;							      
}			
						      
template <class T>								      
inline T& Array<T>::operator[](int index) const {		      
    th_assert((index >= 0) && (index < size_), "array index out of bounds");  
    return store_[index];						      
}									      

template <class T>								      
inline T& Array<T>::slot(int index) const {			      
    th_assert((index >= 0) && (index < size_), "array index out of bounds");  
    return store_[index];						      
}									      
			
template <class T>					      
inline T& Array<T>::high() const {				      
    th_assert(size_ > 0, "array index out of bounds");			      
    return store_[size_-1];						      
}									      
			
template <class T>				      
inline T* Array<T>::as_pointer() const {				      
    return store_;							      
}									      
			
template <class T>				      
inline void Array<T>::append(T v) {				      
    if (size_ >= alloc_)						      
	enlarge_allocation_to(size_+1);					      
    store_[size_++] = v;						      
}									      
			
template <class T>				      
inline T Array<T>::remove() {					      
    if (size_ > 0) size_--;						      
    return store_[size_];						      
}									      
			
template <class T>				      
inline void Array<T>::remove(int num) {				      
    th_assert((num >= 0) && (num <= size_), "invalid array remove count");    
    size_ -= num;							      
}									      
			
template <class T>				      
inline void Array<T>::clear() {					      
    size_ = 0;								      
}									      
			
template <class T>				      
inline void Array<T>::_enlarge_by(int n) {				      
    th_assert(n >= 0, "negative count supplied to array operation");	      
    int newsize = size_ + n;                                                  
    if (newsize > alloc_)						      
	enlarge_allocation_to(newsize);					      
    size_ = newsize;                                                          
}									      
							  
#endif // _Array_h

