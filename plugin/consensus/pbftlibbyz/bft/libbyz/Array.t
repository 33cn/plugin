#include "Array.h"
	
template <class T>								     
Array<T>::Array(int predict) {					    
    th_assert(predict >= 0, "negative count supplied to array operation");
    alloc_ = 0;								  
    size_ = 0;								  
    store_ = 0;								  
    enlarge_to(predict);						  
    size_ = 0;								  
}									  

template <class T>								
Array<T>::~Array() {						    
    if (alloc_ > 0) delete [] store_;					
}

template <class T>
Array<T>::Array(T const* src, int s) {
    th_assert(s >= 0, "negative count supplied to array operation");
    alloc_ = 0;								
    size_  = 0;								
    store_ = 0;								
    enlarge_to(s);							
    for (int i = 0; i < s; i++)						
	store_[i] = src[i];						
}									

template <class T>							
Array<T>::Array(Array const& d) {				    
    alloc_ = 0;							
    size_  = 0;							
    store_ = 0;							
    enlarge_to(d.size_);					
    for (int i = 0; i < size_; i++)				
	store_[i] = d.store_[i];				
}								

template <class T>		
Array<T>::Array(T element, int num) {	
    th_assert(num >= 0, "negative count supplied to array operation");	      
    alloc_ = 0;								      
    size_ = 0;								      
    store_ = 0;								      
    enlarge_to(num);							      
    for (int i = 0; i < num; i++)					      
	store_[i] = element;						      
}									      

template <class T>							
Array<T>& Array<T>::operator=(Array const& d) {			      
    size_ = 0;								      
    enlarge_to(d.size_);						      
    for (int i = 0; i < size_; i++)					      
	store_[i] = d.store_[i];					      
    return (*this);							      
}									      

									      
template <class T>
void Array<T>::append(T element, int n) {			      
    th_assert(n >= 0, "negative count supplied to array operation");	      
    int oldsize = size_;	
    enlarge_to(size_ + n);		
    for (int i = 0; i < n; i++)		
	store_[i + oldsize] = element;	
}					
		
template <class T>			
void Array<T>::concat(Array const& d) {	
    int oldsize = size_;		
    enlarge_to(size_ + d.size_);	
    for (int i = 0; i < d.size_; i++)	
	store_[i+oldsize] = d.store_[i];
}					
			
template <class T>		
void Array<T>::concat(T const* src, int s) {
    th_assert(s >= 0, "negative count supplied to array operation");
    int oldsize = size_;						
    enlarge_to(s + size_);						
    for (int i = 0; i < s; i++)						
	store_[i+oldsize] = src[i];					
}									
			
template <class T>					
void Array<T>::predict(int new_alloc) {				
    if (new_alloc > alloc_)					
        enlarge_allocation_to(new_alloc);			
}								
 			
template <class T>					
void Array<T>::enlarge_to(int newsize) {				
    if (newsize > alloc_)					
	enlarge_allocation_to(newsize);				
   size_ = newsize;						
}								
			
template <class T>					
void Array<T>::enlarge_allocation_to(int newsize) {		
    int newalloc = alloc_ * 2;					
    if (newsize > newalloc) newalloc = newsize;			
								
    T* oldstore = store_;					
    store_ = new T[newalloc];					
								
    for (int i = 0; i < size_; i++)				
	store_[i] = oldstore[i];				
								
    if (alloc_ > 0) delete [] oldstore;			
    alloc_ = newalloc;					
}							
			
template <class T>				
void Array<T>::reclaim() {					
    if (alloc_ > size_) {				
	/* Some free entries that can be reclaimed */	
	if (size_ > 0) {				
	    /* Array not empty - create new store */	
	    T* newstore = new T[size_];			
	    for (int i = 0; i < size_; i++)		
		newstore[i] = store_[i];		
	    delete [] store_;				
	    alloc_ = size_;				
	    store_ = newstore;				
	}						
	else {						
	    /* Array empty - delete old store */	
	    if (alloc_ > 0) {				
		delete [] store_;			
		alloc_ = 0;				
	    }						
	}						
    }							
}	

