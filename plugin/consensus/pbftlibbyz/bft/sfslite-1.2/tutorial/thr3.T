// -*-c++-*-
/* $Id: ex1.T 2236 2006-09-29 00:00:22Z max $ */

#include "async.h"
#include "tame.h"
#include <list.h>

struct req_t {
  req_t (int p, evi_t::ptr d) : _param (p), _done (d) {}
  int _param;
  evi_t::ptr _done;
};

struct worker_t {
  worker_t () : _shutdown (false), _req (NULL) {}
  void serve ();
  int blocking_action (int i);
  int _index;
  list_entry<worker_t> _rlnk;
  evv_t::ptr _ev;
  bool _shutdown;
  req_t *_req;
};

struct server_t {
public:
  server_t (size_t n) : _n_threads (n), _rv (__FILE__, __LINE__) { init (); }
  void run (evv_t cb, CLOSURE);
  void shutdown (evv_t cb, CLOSURE);
  void serve (int i, evi_t cb, CLOSURE);
private:
  void init ();
  void ready_thread (worker_t *w);
  void get_thread (event<worker_t *>::ref, CLOSURE);
  size_t _n_threads;
  vec<worker_t> _workers;
  list<worker_t, &worker_t::_rlnk> _ready;
  size_t _n_alive;
  vec<evv_t> _waiters;
  rendezvous_t<size_t> _rv;
  evv_t::ptr _shutdown_cb;
  bool _shutdown;
};

void
server_t::init ()
{
  _workers.setsize (_n_threads);

  for (size_t i = 0; i < _workers.size (); i++) {
    worker_t *w = &_workers[i];
    w->_index = i;
    tfork (_rv, i, wrap (w, &worker_t::serve));
    ready_thread (w);
  }
}

void
server_t::ready_thread (worker_t *w)
{
  _ready.insert_head (w);
  if (_waiters.size ()) {
    evv_t cb = _waiters.pop_front ();
    cb->trigger ();
  }
}

tamed void
server_t::run (evv_t cb)
{
  twait { _shutdown_cb = mkevent (); }
  cb->trigger ();
}

tamed void
server_t::shutdown (evv_t cb)
{
  tvars {
    size_t i;
  }
  warn << "s> shutdown started...\n";
  _shutdown = true;
  for (i = 0; i < _workers.size (); i++) {
    worker_t *w = &_workers[i];
    w->_shutdown = true;
    if (w->_ev) w->_ev->trigger ();
  }
  while (_rv.n_triggers_left ()) {
    twait (_rv, i);
    warn << "s> thread " << i << " exitted\n";
  }
  warn << "s> shutdown finished...\n";
  cb->trigger ();
}

void
worker_t::serve ()
{
  warn << "w> " << _index << " starting up...\n";
  while (!_shutdown) {
    if (!_req) {
      twait { _ev = mkevent (); }
      _ev = NULL;
    }
    if (_req) {

      int i = blocking_action (_req->_param);

      // 'dtrigger' calls the given callback, with the given parameter,
      // but only after this thread yields.  Calling req->_done directly
      // here makes for a nasty race condition between the server and
      // the worker thread.
      dtrigger (_req->_done, i);

      _req = NULL;
    }
  }
  warn << "w> " << _index << " shutting down...\n";
}

int
worker_t::blocking_action (int i)
{
  int s = rand () % 10 + 1;
  warn << "a> " << _index << ": server req=" << i <<  "; "
       << "sleep=" << s << "\n";
  sleep (s);
  warn << "a> " << _index << ": wake up\n";
  return s;
}

tamed void
server_t::get_thread (event<worker_t *>::ref cb)
{
  tvars {
    worker_t *w (NULL);
  }
  while (!_shutdown && !(w = _ready.first)) 
    twait { _waiters.push_back (mkevent ()); }
  if (w) _ready.remove (w);
  cb->trigger (w);
}

tamed void
server_t::serve (int i, evi_t cb)
{
  tvars {
    worker_t *w;
    int res (-1);
    req_t *r;
  }
  if (i == 0) {
    if (_shutdown_cb) {
      warn << "s> invoking shutdown_cb ...\n";
      evv_t c = _shutdown_cb;
      _shutdown_cb = NULL;
      c->trigger ();
    }
  } else {
    twait { get_thread (mkevent (w)); }
    if (w && !w->_shutdown) {
      twait { 
	r = New req_t (i, mkevent (res));
	w->_req = r;
	if (w->_ev) {
	  w->_ev->trigger ();
	}
      }
      delete r;
      ready_thread (w);
    }
  }
  cb->trigger (res);
}

tamed static void
run_server (server_t **sp, evv_t cb) 
{
  tvars {
    server_t *s (New server_t (6));
  }
  *sp = s;

  twait { s->run (mkevent ()); }
  warn << "m> server done running...\n";
  twait { s->shutdown (mkevent ()); }
  warn << "m> server done shutting down ...\n";
  delete s;
  cb->trigger ();
}

#define N_CALLS 30

tamed static void
run_client (server_t *s, evv_t cb)
{
  tvars {
    int res[N_CALLS];
    int i;
    rendezvous_t<int> rv (__FILE__, __LINE__);
    int sdres;
  }

  for (i = 1; i < N_CALLS; i++) {
    warn << "c> issue request " << i << "\n";
    s->serve (i, mkevent (rv, i, res[i]));
    twait { delaycb (1, 0, mkevent ()); }
  }
  while (rv.n_triggers_left ()) {
    twait (rv, i);
    warn << "c> request " << i << " returned: " << res[i] << "\n";
  }
  warn << "c> shutdown trigger\n";
  twait { s->serve (0, mkevent (sdres)); }
  cb->trigger ();
}

#undef N_CALLS
  
tamed static 
void main2 ()
{
  tvars {
    server_t *s;
  }
  twait {
    run_server (&s, mkevent ());
    run_client (s, mkevent ());
  }

  exit (0);
}


int main (int argc, char *argv[])
{
  main2 ();
  amain ();
}

