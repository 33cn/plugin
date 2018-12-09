#include "fh_hash.h"
#include <rpc/types.h>
#include <rpc/auth.h>
#include "nfsd.h"
#include "fs.h"
#include "map.h"
#include "bhash.t"
#include "buckets.t"


class fhandleKey {
public:
  fhandleKey(fhandle const &k);
  ~fhandleKey() {};
  void operator=(fhandleKey const &k);
  bool operator==(fhandleKey const &key2) const;
  int hash() const;
  
  fhandle val;
};

inline fhandleKey::fhandleKey(fhandle const &k) {
  for (int i=0; i<FHSIZE; i++)
    val.data[i] = k.data[i];
}

inline void fhandleKey::operator=(fhandleKey const &k) {
  for (int i=0; i<FHSIZE; i++)
    val.data[i] = k.val.data[i];
}

inline bool fhandleKey::operator==(fhandleKey const &key2) const {
  for (int i=0; i<FHSIZE; i++)
    if (val.data[i]!=key2.val.data[i])
      return false;
  return true;
}

inline int fhandleKey::hash() const {
  return *((int *)&val.data[24]);
}


Map<fhandleKey,int> nfs2cli(NUM_LOGICAL_INODES);

void init_map() {}

void add(fhandle *key, int value)
{
  nfs2cli.add(*key, value);
}


int find(fhandle *key)
{
  int retval;
  if (nfs2cli.find(*key, retval))
    return retval;
  else
    return -1;
}

int remove_fh(fhandle *key)
{
  int retval;
  if (nfs2cli.remove(*key, retval))
    return retval;
  else
    return -1;
}

void fh_map_clear()
{
  nfs2cli.clear();
}
