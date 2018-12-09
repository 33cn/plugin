// -*-c++-*-

#include "tame_aio.h"

namespace tame {
  
  tamed void 
  aiofh_t::open (const str &fn, int flg, int mode, evi_t ev)
  {
    tvars {
      int rc;
    }
    _fn = fn;
    twait { _aiod->open (fn, flg, mode, mkevent (_fh, rc)); }
    ev->trigger (rc);
  }

  tamed void 
  aiofh_t::read (size_t sz, aio_read_ev_t ev)
  {
    tvars {
      ptr<aiobuf> rbuf;
      ssize_t rsz;
      int rc;
    }
    assert (_fh);
    assert (_aiod);

    if (!_buf || sz > _bufsz) {
      _buf = _aiod->bufalloc (sz);
      _bufsz = sz;
    }
    if (!_buf) {
      rsz = -1;
    } else {
      twait { _fh->read (_off, _buf, mkevent (rbuf, rsz, rc)); }
      if (rc != 0) {
	warn ("Read error on file (%s): %m\n", _fn.cstr ());
	rsz = -1;
      } else if (rsz < 0) {
	warn ("Read error (short read) on file %s\n", _fn.cstr ());
      } else {
	_off += rsz;
      }
    }
    ev->trigger (rbuf, rsz);
  }

  tamed void
  aiofh_t::close (evi_t::ptr ev)
  {
    tvars {
      int ret;
    }
    assert (_fh);
    twait { 
      _fh->close (mkevent (ret)); 
      _fh = NULL;
    }
    if (ev)
      ev->trigger (ret);
  }

  aiofh_t::~aiofh_t ()
  {
    if (_fh) {
      close ();
    }
  }

};
