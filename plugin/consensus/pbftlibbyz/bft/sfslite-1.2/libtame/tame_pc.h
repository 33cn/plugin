

// -*-c++-*-
/* $Id: tame.h 2077 2006-07-07 18:24:23Z max $ */

#ifndef _LIBTAME_PC_H_
#define _LIBTAME_PC_H_

#include "async.h"
#include "list.h"
#include "tame.h"

//
// producer/consumer with 1 object
//

namespace tame {

  template<class V>
  class pc1_t {
  public:
    pc1_t (const V &v) : _set (false), _done (false), _val (v) {}

    void set (const V &v) 
    {
      assert (!_set);
      assert (!_done);

      _val = v;
      if (_cb) {
	_done = true;
	typename callback<void, bool, V>::ref c (_cb);
	_cb = NULL;
	c->trigger (true, _val);
      } else {
	_set = true;
      }
    }

    void get (typename callback<void, bool, V>::ref c)
    {
      bool ret (false);
      if (!_done) {
	if (_set) {
	  assert (!_cb);
	  _done = true;
	  ret = true;
	} else if (!_cb) {
	  _cb = c;
	  return;
	}
      }
      c->trigger (ret, _val);
    }

    bool has (V *v)
    {
      if (_set && !_done) {
	assert (!_cb);
	*v = _val;
	return true;
      }
      return false;
    }
   
  private:
    bool _set;
    bool _done;
    V _val;
    typename callback<void, bool, V>::ptr _cb;
  };  


  template<class V>
  class pc_t {
  private:
    typedef typename event<V>::ref _ev_t;
  public:
    pc_t (size_t mx = 0) : _maxsz (mx) {}

    bool produce (const V &v)
    {
      if (_maxsz > 0 && _q.size () >= _maxsz)
	return false;

      _q.push_back (v);
      if (_waiters.size ()) {
	_ev_t ev (_waiters.pop_front ());
	V v (_q.pop_front ());
	ev->trigger (v);
      }
      return true;
    }

    void consume (_ev_t e)
    {
      if (_q.size ()) {
	V v (_q.pop_front ());
	e->trigger (v);
      } else {
	_waiters.push_back (e);
      }
    }

  private:
    vec<V> _q;
    vec<_ev_t> _waiters;
    size_t _maxsz;
  };

};


#endif /* _LIBTAME_PC_H_ */

