
// -*-c++-*-
/* $Id: async.h 3492 2008-08-05 21:38:00Z max $ */

#include "async.h"
#include "tame.h"
#include "parseopt.h"
#include "arpc.h"
#include "rtftp_prot.h"
#include "sha1.h"
#include "rxx.h"
#include "rtftp.h"

//-----------------------------------------------------------------------

enum { RC_OK = 0, RC_ERPC = -2, RC_ERR = -1, RC_EXISTS = 1 };

#define WINDOWSZ 30

class cli_t {
public:
  cli_t (const str &d, const str &f, const str &h, int p, size_t cs, size_t ws)
    : _verbose (false), _dir (d), _file (f), _host (h), _port (p),
      _chunk_sz (cs), _window_sz (ws) {}

  virtual ~cli_t () {}
  void set_verbose (bool b) { _verbose = b; }
  void run (CLOSURE);
  virtual bool init () { return true; }

protected:
  virtual void perform (evi_t ev) = 0;
  void connect (evb_t ev, CLOSURE);

  bool _verbose;
  str _dir;
  str _file;
  str _host;
  int _port;

  ptr<axprt_stream> _x;
  ptr<aclnt> _cli;
  size_t _chunk_sz;

public:
  size_t _window_sz;
};

//-----------------------------------------------------------------------

class put_cli_t : public cli_t {
public:
  put_cli_t (const str &d, const str &f, const str &h, int p, 
	     size_t cs, size_t ws) 
    : cli_t (d, f, h, p, cs, ws) {}
  void perform (evi_t ev) { perform_T (ev); }
private:
  void perform_T (evi_t ev, CLOSURE);
};

//-----------------------------------------------------------------------

class put2_cli_t : public cli_t {
public:
  put2_cli_t (const str &d, const str &f, const str &h, int p, 
	      size_t cs, size_t ws)
    : cli_t (d, f, h, p, cs, ws) , _sz (0) {}
  void perform (evi_t ev) { perform_T (ev); }
private:
  void perform_T (evi_t ev, CLOSURE);
  void register_file (evi_t ev, CLOSURE);
  void transfer_file (evi_t ev, CLOSURE);
  void put_footer (size_t fsz, u_int32_t nchnk, evi_t ev, CLOSURE);

  int _fd;
  rtftp_xfer_id_t _id;
  sha1ctx _ctx;
  size_t _sz;
};

//-----------------------------------------------------------------------

class get_cli_t : public cli_t {
public:
  get_cli_t (const str &d, const str &f, const str &h, int p, 
	     size_t cs, size_t ws)
    : cli_t (d, f, h, p, cs, ws) {}
  void perform (evi_t ev) { perform_T (ev); }
  bool init ();
private:
  int write_file (const str &s);
  void perform_T (evi_t ev, CLOSURE);
};

//-----------------------------------------------------------------------

class get2_cli_t : public cli_t {
public:
  get2_cli_t (const str &d, const str &f, const str &h, int p, 
	      size_t cs, size_t ws)
    : cli_t (d, f, h, p, cs, ws) , _fd (-1), _sz (0) {}
  void perform (evi_t ev) { perform_T (ev); }
  bool init ();
private:
  void perform_T (evi_t ev, CLOSURE);
  int open_file (int mode);
  void transfer_file (evi_t ev, CLOSURE);
  void request_file (evi_t ev, CLOSURE);
  bool check_file ();
  void send_eof_to_srv (evv_t ev, CLOSURE);
  
  int _fd;
  rtftp_xfer_id_t _id;
  sha1ctx _ctx;
  size_t _sz;
  char _hsh[RTFTP_HASHSZ];
  size_t _recv_bytes;
};

//-----------------------------------------------------------------------

static void
usage ()
{
  warnx << "usage: " << progname << " [-v] [-2] [-c<chnksz>] [-w<winsz>] "
	<< "<put|get> <dir> <file> <host>\n";
}

//-----------------------------------------------------------------------

static bool
parse_host (const str &s, str *hp, int *pp) 
{
  static rxx x ("([^:]+)(:(\\d+))?");
  bool ret = x.match (s);
  if (ret) {
    *hp = x[1];
    if (x[2]) {
      bool rc = convertint (x[3], pp);
      assert (rc);
    }
  }
  return ret;
}

//-----------------------------------------------------------------------

static int
config (cli_t **clip, int argc, char *argv[])
{
  int rc = 1;
  bool verbose = false;
  int port = RTFTP_TCP_PORT;
  str host;
  str file;
  str dir;
  cli_t *cli = NULL;
  int ch;
  int version = 2;
  size_t cs = CHUNKSZ;
  size_t ws = WINDOWSZ;
  
  while ((ch = getopt (argc, argv, "vh12c:w:")) != -1) {
    switch (ch) {
    case '1':
      version = 1;
      break;
    case '2':
      version = 2;
      break;
    case 'v':
      verbose = true;
      break;
    case 'c':
      if (!convertint (optarg, &cs)) {
	warn << "cannot convert to int: " << optarg << "\n";
	usage ();
      }
      break;
    case 'w':
      if (!convertint (optarg, &ws)) {
	warn << "cannot convert to int: " << optarg << "\n";
	usage ();
      }
      break;
    case 'h':
      usage ();
      rc = 0;
      break;
    default:
      usage ();
      rc = -1;
    }
  }

  argc -= optind;
  argv += optind;

  if (rc > 0) {
    bool err = false;
    if (argc != 4) {
      err = true;
    } else if (!parse_host (argv[3], &host, &port)) {
      warn << "bad hostname/port: " << argv[3];
      err = true;
    } else {
      dir = argv[1];
      file = argv[2];
      if (strcmp (argv[0], "put") == 0) {
	if (version == 1) {
	  cli = New put_cli_t (dir, file, host, port, cs, ws);
	} else {
	  cli = New put2_cli_t (dir, file, host, port, cs, ws);
	}
      } else if (strcmp (argv[0], "get") == 0) {
	if (version == 1) {
	  cli = New get_cli_t (dir, file, host, port, cs, ws);
	} else {
	  cli = New get2_cli_t (dir, file, host, port, cs, ws);
	}
      } else {
	warn << "need either put/get operation mode\n";
	err = true;
      }
    }
   
    if (err) {
      usage ();
      rc = -1;
    } else {
      cli->set_verbose (verbose);
    }
  }
  *clip = cli;
  return rc;
}

//-----------------------------------------------------------------------

tamed void
cli_t::connect (evb_t ev)
{
  tvars {
    int fd;
    bool ret (true);
  }
  twait { tcpconnect (_host, _port, mkevent (fd)); }
  if (fd < 0) {
    warn << "connection to " << _host << ":" << _port << " failed\n";
    ret = false;
  } else {
    _x = axprt_stream::alloc (fd);
    _cli = aclnt::alloc (_x, rtftp_program_1);
  }
  ev->trigger (ret);
}

//-----------------------------------------------------------------------

tamed void
put2_cli_t::register_file (evi_t ev)
{
  tvars {
    rtftp_put2_arg_t arg (RTFTP_BEGIN);
    rtftp_put2_res_t res;
    clnt_stat err;
    int rc;
  }

  *arg.name = _file;
  twait { RPC::rtftp_program_1::rtftp_put2 (_cli, arg, &res, mkevent (err)); }
  if (err) { 
    warn << "RPC error: " << err << "\n";
    rc = RC_ERPC;
  } else if (res.status == RTFTP_EEXISTS) {
    warn << "File exists: " << _file << "\n";
    rc = RC_EXISTS;
  } else if (res.status != RTFTP_BEGIN) {
    warn << "Put failure on file " << _file << ": " << int (res.status) << "\n";
    rc = RC_ERR;
  } else {
    _id = *res.xfer_id;
    rc = RC_OK;
  }
  ev->trigger (rc);
}

//-----------------------------------------------------------------------

bool
get2_cli_t::init ()
{
  bool rc;
  if (chdir (_dir) != 0) {
    warn ("cannot chdir to directory %s: %m\n", _dir.cstr ());
    rc = false;
  } else {
    rc = true;
  }
  return rc;
}
//-----------------------------------------------------------------------

bool
get_cli_t::init ()
{
  bool rc;
  if (chdir (_dir) != 0) {
    warn ("cannot chdir to directory %s: %m\n", _dir.cstr ());
    rc = false;
  } else {
    rc = true;
  }
  return rc;
}

//-----------------------------------------------------------------------

int 
get2_cli_t::open_file (int mode)
{
  return ::open_file (_file.cstr (), mode);
}

//-----------------------------------------------------------------------

tamed void
get2_cli_t::request_file (evi_t ev)
{
  tvars {
    rtftp_get2_arg_t arg (RTFTP_BEGIN);
    rtftp_get2_res_t res;
    clnt_stat err;
    int rc (0);
  }
  *arg.name = _file;
  twait { RPC::rtftp_program_1::rtftp_get2 (_cli, arg, &res, mkevent (err)); }
  if (err) {
    warn << "RPC error in file request for " << _file << ": " << err << "\n";
    rc = RC_ERPC;
  } else if (res.status == RTFTP_NOENT) {
    rc = RC_EXISTS;
  } else if (res.status != RTFTP_BEGIN) {
    warn << "Server error in GET for file " << _file << ": " 
	 << int (res.status) << "\n";
    rc = RC_ERR;
  } else {
    _id = res.header->xfer_id;
    _sz = res.header->size;
    memcpy (_hsh, res.header->hash.base (), RTFTP_HASHSZ);
  }
  ev->trigger (rc);
}

//-----------------------------------------------------------------------

tamed void
get2_cli_t::send_eof_to_srv (evv_t ev)
{
  tvars {
    rtftp_get2_arg_t arg (RTFTP_EOF);
    rtftp_get2_res_t res;
    clnt_stat s;
  }

  *arg.id = _id;

  twait { RPC::rtftp_program_1::rtftp_get2 (_cli, arg, &res, mkevent (s)); }
  if (s) {
    warn << "RPC failure in sending EOF to server: " << s << "\n";
  } else if (res.status != RTFTP_OK) {
    warn ("Server failed to close file %s: %d\n", 
	  _file.cstr (), int (res.status));
  }
  ev->trigger ();
}

//-----------------------------------------------------------------------

tamed void
get2_cli_t::transfer_file (evi_t ev)
{
  tvars {
    vec<int> ids;
    vec<rtftp_get2_arg_t> args;
    vec<rtftp_get2_res_t> ress;
    size_t wsz (_self->_window_sz);
    u_int32_t cr (0), cs (0); // chunk receive, chunk sent
    rendezvous_t<int> rv (__FILE__, __LINE__);
    vec<clnt_stat> stats;
    int rc (0);
    int id;
    size_t nbrq (0), nbrc (0); // n bytes requsted and received
  }

  ids.setsize (wsz);
  args.setsize (wsz);
  stats.setsize (wsz);
  ress.setsize (wsz);
  for (size_t i = 0; i < wsz; i++) { ids[i] = wsz - (i + 1); }

  while ((nbrq < _sz || ids.size () < wsz) && rc >= 0) {
    assert (cr <= cs);
    if (ids.size () && (nbrq < _sz)) {
      id = ids.pop_back ();

      clnt_stat &stat = stats[id];
      rtftp_get2_arg_t &arg = args[id];
      rtftp_get2_res_t &res = ress[id];

      arg.set_status (RTFTP_OK);
      arg.chunk->xfer_id = _id;
      arg.chunk->offset = nbrq;
      arg.chunk->size = _chunk_sz;

      nbrq += arg.chunk->size;
      cs++;

      RPC::rtftp_program_1::rtftp_get2 (_cli, arg, &res,
					mkevent (rv, id, stat));
    } else {
      twait (rv, id);

      const clnt_stat &stat = stats[id];
      const rtftp_get2_res_t &res = ress[id];
      
      if (stat) {
	warn << "RPC failure in get2 for file " << _file 
	     << ": " << stat << "\n";
	rc = RC_ERPC;
      } else if (res.status != RTFTP_OK) {
	warn << "Server failure for file " << _file << ": " 
	     << int (res.status) << "\n";
	rc = RC_ERR;
      } else if (res.chunk->id.xfer_id != _id) {
	warn << "Got data for wrong file; expected " 
	     << _id << " but got " << res.chunk->id.xfer_id << "\n";
	rc = RC_ERR;
      } else if (res.chunk->id.offset != nbrc) {
	warn << "Chunks out of sequence; expected offset " 
	     << nbrc << " but got " << res.chunk->id.offset << "\n";
	rc = RC_ERR;
      } else {
	size_t sz = res.chunk->data.size ();
	ssize_t wrc = write (_fd, res.chunk->data.base (), sz);
	if (wrc < 0) {
	  warn ("failed write on file %s: %m\n", _file.cstr ());
	  rc = RC_ERR;
	} else if (wrc != ssize_t (sz)) {
	  warn ("short write on file %s\n", _file.cstr ());
	  rc = RC_ERR;
	} else {
	  _ctx.update (res.chunk->data.base (), sz);
	  nbrc += sz;
	}
      }
      cr ++;
      ids.push_back (id);
    }
  }
  close (_fd);
  _fd = -1;

  _recv_bytes = nbrc;

  if (rc == 0) {
    assert (cr == cs);
    assert (nbrq >= _sz);
    assert (ids.size () == wsz);
  } else {
    rv.cancel ();
  }

  twait { send_eof_to_srv (mkevent ()); }

  ev->trigger (rc);

}

//-----------------------------------------------------------------------

bool
get2_cli_t::check_file ()
{
  char hsh[RTFTP_HASHSZ];
  _ctx.final (hsh);

  bool ok (false);

  if (_recv_bytes != _sz) {
    warn << "File size mismatch for file " << _file << "; "
	 << "expected " << _sz << ", but got " << _recv_bytes << "\n";
  } else if (memcmp (hsh, _hsh, RTFTP_HASHSZ) != 0) {
    warn ("Hash mismatch on file %s\n", _file.cstr ());
  } else {
    ok = true;
  }
  return ok;
}

//-----------------------------------------------------------------------

tamed void
get2_cli_t::perform_T (evi_t ev)
{
  tvars {
    int rc;
  }
  twait { request_file (mkevent (rc)); }
  if (rc == RC_EXISTS) {
    /* file does not exist; noop! */
  } else if (rc < 0) {
    warn ("failed to start transfer for file %s\n", _file.cstr ());
  } else {
    if ((_fd = open_file (O_WRONLY|O_CREAT|O_TRUNC)) < 0) {
      warn ("cannot create/open file %s: %m\n", _file.cstr ());
    } else {
      twait { transfer_file (mkevent (rc)); }
      if (rc < 0) {
	warn ("failed to transfer file %s\n", _file.cstr ());
      } else if (!check_file ()) {
	rc = RC_ERR;
      }
    }
  }
  if (_fd) {
    close (_fd);
  }
  ev->trigger (rc);
}

//-----------------------------------------------------------------------

tamed void
put2_cli_t::transfer_file (evi_t ev)
{
  tvars {
    vec<int> ids;
    vec<rtftp_put2_arg_t> args;
    vec<rtftp_put2_res_t> ress;
    size_t wsz (_self->_window_sz);
    size_t fsz (0);
    u_int32_t cr (0), cs (0); // chunk receive, chunk sent
    rendezvous_t<int> rv (__FILE__, __LINE__);
    vec<clnt_stat> stats;
    bool eof (false);
    int rc (0);
    int id;
  }

  ids.setsize (wsz);
  args.setsize (wsz);
  stats.setsize (wsz);
  ress.setsize (wsz);
  for (size_t i = 0; i < wsz; i++) { ids[i] = wsz - (i + 1); }

  while ((!eof || ids.size () < wsz) && rc >= 0) {
    assert (cr <= cs);
    if (ids.size () && !eof) {
      id = ids.pop_back ();
      clnt_stat &stat = stats[id];
      rtftp_put2_arg_t &arg = args[id];
      rtftp_put2_res_t &res = ress[id];

      arg.set_status (RTFTP_OK);
      arg.data->data.setsize (_chunk_sz);
      rc = read (_fd, arg.data->data.base (), _chunk_sz);
      if (rc < 0) {
	warn ("read error on file %s: %m\n", _file.cstr ());
	ids.push_back (id);
      } else if (rc == 0) {
	eof = true;
	ids.push_back (id);
      } else {

	assert (size_t (rc) <= _chunk_sz);
	arg.data->id.xfer_id = _id;
	arg.data->id.offset = fsz;
	arg.data->id.size = rc;

	fsz += rc;
	cs++;

	arg.data->data.setsize (rc);
	_ctx.update (arg.data->data.base (), rc);
	RPC::rtftp_program_1::rtftp_put2 (_cli, arg, &res, 
					  mkevent (rv, id, stat));
      }
    } else {
      twait (rv, id);

      const clnt_stat &stat = stats[id];
      const rtftp_put2_res_t &res = ress[id];
      if (stat) {
	warn << "RPC error in putting file " << _file << ": " << stat << "\n";
	rc = RC_ERPC;
      } else if (res.status != RTFTP_OK) {
	warn ("Server rejected put on file %s: %d\n", _file.cstr (), 
	      int (res.status));
	rc = RC_ERR;
      }
      cr ++;
      ids.push_back (id);
    }
  }

  if (rc == 0) {
    assert (eof);
    assert (ids.size () == wsz);
    assert (cr == cs);
    twait { put_footer (fsz, cr, mkevent (rc)); }

  } else {
    rv.cancel ();
  }

  ev->trigger (rc);
}

//-----------------------------------------------------------------------

tamed void
put2_cli_t::put_footer (size_t fsz, u_int32_t nchnk, evi_t ev)
{
  tvars {
    rtftp_put2_arg_t arg;
    clnt_stat stat;
    rtftp_put2_res_t res;
    int rc (0);
  }
    
  arg.set_status (RTFTP_EOF);
  arg.footer->xfer_id = _id;
  arg.footer->size = fsz;
  arg.footer->n_chunks = nchnk;
  _ctx.final (arg.footer->hash.base ());
  twait { 
    RPC::rtftp_program_1::rtftp_put2 (_cli, arg, &res, mkevent (stat));
  }
  if (stat) {
    warn << "RPC error in putting footer for file " << _file << ": "
	 << int (stat) << "\n";
    rc = RC_ERPC;
  } else if (res.status != RTFTP_OK) {
    warn << "Server reported error when putting footer for file "
	 << _file << ": " << int (res.status) << "\n";
    rc = RC_ERR;
  }
  ev->trigger (rc);
}

//-----------------------------------------------------------------------

tamed void
put2_cli_t::perform_T (evi_t ev)
{
  tvars {
    str file_full;
    int rc;
  }

  {
    strbuf b;
    b << _dir << "/" << _file;
    file_full = b;
  }

  if ((_fd = open (file_full.cstr(), O_RDONLY)) < 0) {
    warn ("cannot open file %s: %m\n", file_full.cstr ());
    rc = -1;
  } else {
    twait { register_file (mkevent (rc)); }
    if (rc < 0) {
      warn ("failed to register file %s\n", _file.cstr ());
    } else {
      twait { transfer_file (mkevent (rc)); }
      if (rc < 0) {
	warn ("file transfer failed on file %s\n", _file.cstr ()); 
      }
    }
  }
  if (_fd >= 0) 
    close (_fd);

  ev->trigger (rc);
}

//-----------------------------------------------------------------------

// trigger: -2 on RPC failure
//          -1 on internal failure
//           0 on success
//           1 if no put due to file already existing
tamed void
put_cli_t::perform_T (evi_t ev)
{
  tvars {
    str dat;
    rtftp_file_t file;
    rtftp_status_t status;
    clnt_stat err;
    int rc (0);
    str file_full;
  }
  
  {
    strbuf b;
    b << _dir << "/" << _file;
    file_full = b;
  }

  if (!(dat = file2str (file_full))) {
    warn ("cannot open file %s: %m\n", file_full.cstr ());
    rc = -1;
  } else {
    file.magic = MAGIC;
    file.name = _file;
    file.data = dat;
    sha1_hash (file.hash.base (), file.data.base (), file.data.size ());
    twait { 
      RPC::rtftp_program_1::rtftp_put (_cli, file, &status, mkevent (err)); 
    }
    if (err) {
      warn  << "RPC failure: "  << err << "\n";
      rc = -2;
    } else if (status == RTFTP_EEXISTS) {
      if (_verbose) {
	warn << "file already exists\n";
      }
      rc = 1;
    } else if (status != RTFTP_OK) {
      warn << "cannot do PUT transfer: " << int (status) << "\n";;
      rc = -1;
    } else {
      rc = 0;
    }
  }
  ev->trigger (rc);
}

//-----------------------------------------------------------------------

//
// write out the data 's' to the file given by _file.  Might need to make
// some parent directories if the filename contains slashes.
//
// return -1 if failure, and 0 if wrote the file out OK.
//
int
get_cli_t::write_file (const str &s)
{
  int rc = ::write_file (_file, s);
  return rc;
}
 
//-----------------------------------------------------------------------

// Trigger: -2 on RPC failure
//          -1 on internal failure
//           0 if file found and gotten
//           1 if file not found
tamed void
get_cli_t::perform_T (evi_t ev)
{
  tvars {
    rtftp_get_res_t res;
    rtftp_id_t arg;
    clnt_stat err;
    int rc (0);
  }

  arg = _file;
  twait { RPC::rtftp_program_1::rtftp_get (_cli, arg, &res, mkevent (err)); }
  if (err) {
    warn << "RPC failure: " << err << "\n";
    rc = -2;
  } else if (res.status == RTFTP_NOENT) {
    rc = 1;
  } else if (res.status != RTFTP_OK) {
    warn << "GET failed: " << int (res.status) << "\n";
    rc = -1;
  } else {
    if (!check_file (*res.file)) {
      warn << "hash mismatch on file!\n";
      rc = -1;
    } else {
      mstr dat (res.file->data.size ());
      memcpy (dat.cstr (), res.file->data.base (), res.file->data.size ());
      dat.setlen (res.file->data.size ());
      rc = write_file (dat);
    }
  }
  ev->trigger (rc);
}

//-----------------------------------------------------------------------

tamed void
cli_t::run ()
{
  tvars {
    bool ok;
    int rc (0);
  }

  twait { connect (mkevent (ok)); }
  if (ok) {
    twait { perform (mkevent (rc)); }
  } else {
    warn << "connection failed: " << _host << ":" << _port << "\n";
    rc = -1;
  }

  exit (rc);
}

//-----------------------------------------------------------------------

int 
main (int argc, char *argv[])
{
  cli_t *cli;
  int rc;
  setprogname (argv[0]);

  if ((rc = config (&cli, argc, argv)) <= 0) {
    return rc;
  }

  if (!cli->init ()) {
    return -1;
  }

  cli->run ();
  amain ();

  return 0; // never get here
}

//-----------------------------------------------------------------------
