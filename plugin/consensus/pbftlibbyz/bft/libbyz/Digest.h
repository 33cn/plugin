#ifndef _Digest_h
#define _Digest_h 1

class Digest { 
public:
  inline Digest() { for(int i=0; i < 4; i++) d[i] = 0; }
  Digest(char *s, unsigned n);
  // Effects: Creates a digest for string "s" with length "n"

  inline Digest(Digest const &x) { 
    d[0] = x.d[0]; d[1] = x.d[1]; d[2] = x.d[2];  d[3] = x.d[3];
  }

  inline ~Digest() {}
  // Effects: Deallocates all storage associated with digest.

  inline void zero() {
    for(int i=0; i < 4; i++) d[i] = 0;
  }

  inline bool is_zero() const {
    return d[0] == 0;
  }

  inline bool operator==(Digest const &x) const { 
    return (d[0] == x.d[0]) & (d[1] == x.d[1]) &
      (d[2] == x.d[2]) & (d[3] == x.d[3]);
  }

  inline bool operator==(unsigned int *e) const { 
    return (d[0] == e[0]) & (d[1] == e[1]) &
      (d[2] == e[2]) & (d[3] == e[3]);
  }

  inline bool operator!=(Digest const &x) const { 
    return !(*this == x);
  }

  inline Digest& operator=(Digest const &x) { 
    d[0] = x.d[0]; d[1] = x.d[1]; d[2] = x.d[2];  d[3] = x.d[3];
    return *this;
  }

  inline int hash() const {
    return d[0];
  }

  char* digest() { return (char *)d; }
  unsigned* udigest() { return d; } 

  void print();
  // Effects: Prints digest in stdout.

private:
  unsigned int d[4]; 
};

#endif // _Digest_h
