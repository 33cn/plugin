#include "fileid_hash.h"
#include <rpc/types.h>
#include <rpc/auth.h>
#include <stdio.h>
#include "nfsd.h"
#include "fs.h"
#include "map.h"
#include "valuekey.h"
#include "bhash.t"
#include "buckets.t"
#include "th_assert.h"

Map<UIntKey,int> fileid2inum(NUM_LOGICAL_INODES);

void init_map_fileid() {}

void add_fileid(unsigned int key, int value)
{
  fileid2inum.add(key, value);
}


int find_fileid(unsigned int key)
{
  int retval;
  if (fileid2inum.find(key, retval))
    return retval;
  else
    return -1;
}

int remove_fileid(unsigned int key)
{
  int retval;
  if (fileid2inum.remove(key, retval))
    return retval;
  else
    return -1;
}

void fileid_map_clear()
{
  fileid2inum.clear();
}

void save_fileid_map(FILE *o)
{
  size_t wb = 0;
  size_t ab = 0;

  int sz = fileid2inum.size();

  wb += fwrite(&sz, sizeof(int), 1, o);
  ab++;

  MapGenerator<UIntKey, int> g(fileid2inum);
  UIntKey fileidkey;
  int v;
  while (g.get(fileidkey, v)) {
    unsigned int fileid = (unsigned int) fileidkey.val;
    wb += fwrite(&fileid, sizeof(unsigned int), 1, o);
    ab++;
    wb += fwrite(&v, sizeof(int), 1, o);
    ab++;
  }
  th_assert(ab == 2*sz + 1 && wb == ab, "Write fileid map failed");
}

void read_fileid_map(FILE *in)
{
  size_t rb = 0;
  size_t ab = 0;
  int sz, i, inum;
  unsigned int fileid;

  rb += fread(&sz, sizeof(int), 1, in);
  ab++;

  for (i=0; i<sz; i++) {
    rb += fread(&fileid, sizeof(unsigned int), 1, in);
    ab++;
    rb += fread(&inum, sizeof(int), 1, in);
    ab++;
    fileid2inum.add(fileid, inum);
  }
  th_assert(ab == 2*sz + 1 && rb == ab, "Read fileid map failed");
}
