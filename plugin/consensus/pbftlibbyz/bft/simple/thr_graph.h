#ifndef _thr_graph_h
#define _thr_graph_h 1

static const int thr_up = 1;
static const int thr_start = 2;
static const int thr_done = 3;
static const int thr_end = 4;

struct thr_command {
  int tag;
  int cid;
  int op;
  int num_iter;
  int read_only;
  float elapsed;
};

#endif //_thr_graph_h
