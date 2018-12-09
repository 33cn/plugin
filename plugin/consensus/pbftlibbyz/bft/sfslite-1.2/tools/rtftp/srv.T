
// -*-c++-*-
/* $Id: async.h 3492 2008-08-05 21:38:00Z max $ */

#include "async.h"
#include "tame.h"
#include "parseopt.h"
#include "arpc.h"
#include "rtftp_prot.h"
#include "crypt.h"
#include "rxx.h"
#include "rtftp.h"
#include <grp.h>

//-----------------------------------------------------------------------

class srv_t {
public:
  srv_t () 
    : _verbose (false), _debug (false), _port (RTFTP_TCP_PORT), _sfd (-1),
      _chroot (false) {}
  int config (int argc, char *argv[]);
  void newcli ();
  bool init ();
  void run ();
private:
  void shutdown (int sig);
  bool init_chroot ();
  bool init_port ();
  bool init_signals ();
  bool init_daemon ();
  bool init_perms ();

  bool _verbose, _debug;
  str _dir;
  int _port;
  int _sfd;

  bool _chroot;
  str _user, _group;
  int _uid, _gid;
};

//-----------------------------------------------------------------------

class file_t {
public:
  typedef enum { FILE_NONE = 0, FILE_PUT = 1, FILE_GET = 2 } mode_t ;

  file_t (const str &n, mode_t m) 
    : _file (n), 
      _fd (-1), 
      _sz (0), 
      _id (0),
      _eof (false), 
      _offset (0), 
      _header_offset (0),
      _mode (m), 
      _put_status (RTFTP_INCOMPLETE) {}

  ~file_t () { if (_fd >= 0) close (_fd); }
  bool open (int m);
  void set_mode (mode_t m) { _mode = m; }
  void set_put_status (rtftp_status_t st);
  void set_id (rtftp_xfer_id_t i) { _id = i; }
  void read_chunk (const rtftp_chunkid_t &chnk, rtftp_get2_res_t *res);
  rtftp_status_t write_chunk (const rtftp_chunk_t &dat);
  bool write_header_placeholder ();
  bool write_header (const rtftp_header_t &h);
  rtftp_status_t put_footer (const rtftp_footer_t &footer);
  bool eof () const { return _eof; }
  bool read_header ();
  size_t expected_size () const { return _expected_sz; }
  void set_xfer_header (rtftp_get2_res_t *res);
  void clean (bool b);
private:
  bool my_write (const void *v, size_t len);
  const str _file;
  int _fd;
  size_t _sz;
  rtftp_xfer_id_t _id;
  sha1ctx _ctx;
  bool _eof;
  off_t _offset;
  off_t _header_offset;

  size_t _expected_sz;
  char _expected_hash[RTFTP_HASHSZ];
  mode_t _mode;
  rtftp_status_t _put_status;
};

//-----------------------------------------------------------------------

class cli_t {
public:
  cli_t (int f, const char *a, bool v)
    : _fd (f), 
      _x (axprt_stream::alloc (f)),
      _srv (asrv::alloc (_x, rtftp_program_1, wrap (this, &cli_t::dispatch))),
      _addr (a),
      _verbose (v),
      _id (0) {}
  ~cli_t () { clean_files (); }
  void dispatch (svccb *sbp);
private:
  rtftp_status_t do_put (const rtftp_file_t *arg);
  void do_get (const rtftp_id_t &arg, rtftp_get_res_t *res);
  void do_get2 (const rtftp_get2_arg_t &arg, rtftp_get2_res_t *res);
  void do_put2 (const rtftp_put2_arg_t &arg, rtftp_put2_res_t *res);

  void clean_files ();

  int _fd;
  ptr<axprt> _x;
  ptr<asrv> _srv;
  const str _addr;
  bool _verbose;
  qhash<rtftp_xfer_id_t, ptr<file_t> > _tab;
  rtftp_xfer_id_t _id;
};

//-----------------------------------------------------------------------

static void
usage ()
{
  warnx << "usage: " << progname 
	<< " [-r] [-u<usr>] [-g<grp>] [-dv] [-p<prt>] [-l<log>] <dir>\n"
	<< "\n"
	<< "Options:\n"
	<< "   -d       debug mode (no backgrounding)\n"
	<< "   -g<grp>  run-as group\n"
	<< "   -l<log>  log priority (e.g. daemon.notice)\n"
	<< "   -p<prt>  port to bind to\n"
	<< "   -r       chrooted operation\n"
	<< "   -u<usr>  run-as user\n"
	<< "   -v       verbose mode (display hits/misses)\n"
	<< "\n"
	<< "Arguments:\n"
	<< "   <dir>    the cache directory\n"
    ;
}

//-----------------------------------------------------------------------

rtftp_status_t
cli_t::do_put (const rtftp_file_t *arg)
{
  rtftp_status_t st;
  if (!check_file (*arg)) {
    st = RTFTP_CORRUPT;
  } else {
    str s = xdr2str (*arg);
    st = (write_file (arg->name, s) == 0) ? RTFTP_OK :  RTFTP_EFS;
  }
  if (_verbose) {
    warn << "PUT: " << arg->name << " -> ";
    rpc_print (warnx, st, 0, NULL, NULL);
    warnx << "\n";
  }
  return st;
}

//-----------------------------------------------------------------------

bool
file_t::open (int m)
{
  return ((_fd = open_file (_file, m)) >= 0);
}

//-----------------------------------------------------------------------

void
file_t::set_put_status (rtftp_status_t s)
{
  assert (_mode == FILE_PUT);
  _put_status = s;
}

//-----------------------------------------------------------------------

void
file_t::clean (bool v)
{
  if (_mode == FILE_PUT && _put_status != RTFTP_OK) {
    int rc = unlink (_file.cstr ());
    if (rc != 0) {
      warn ("Cannot unlink file %s: %m\n", _file.cstr ());
    }
    if (v) {
      warn << "PUT2: " << _file << " -> ";
      rpc_print (warnx, _put_status, 0, NULL, NULL);
      warnx << " (transmission failed)\n";
    }
  }
}

//-----------------------------------------------------------------------

void
cli_t::clean_files ()
{
  qhash_const_iterator_t<rtftp_xfer_id_t, ptr<file_t> > it (_tab);
  ptr<file_t> f;
  while (it.next (&f)) {
    f->clean (_verbose);
  }
}

//-----------------------------------------------------------------------

rtftp_status_t
file_t::write_chunk (const rtftp_chunk_t &arg)
{
  rtftp_status_t st;
  if (arg.id.offset != _sz) {
    st = RTFTP_OUT_OF_SEQ;
  } else {
    _ctx.update (arg.data.base (), arg.data.size ());
    if (!my_write (arg.data.base (), arg.data.size ())) {
      st = RTFTP_EFS;
    } else {
      st = RTFTP_OK;
      _sz += arg.data.size ();
    }
  }
  return st;
}

//-----------------------------------------------------------------------

bool
file_t::write_header_placeholder ()
{
  rtftp_header_t h;
  h.name = _file;
  h.magic = MAGIC;
  h.size = 0;
  memset (h.hash.base (), 0, RTFTP_HASHSZ);
  return write_header (h);
}


//-----------------------------------------------------------------------

bool
file_t::write_header (const rtftp_header_t &h)
{
  str s = xdr2str (h);
  size_t sz = s.len ();

  return 
    my_write ((const void *)&sz, sizeof (sz)) &&
    my_write (s.cstr (), s.len ());
}

//-----------------------------------------------------------------------

bool
file_t::my_write (const void *v, size_t len)
{
  ssize_t rc = write (_fd, v, len);
  bool ok = false;
  if (rc < 0) {
    warn ("write failed for file %s: %m\n", _file.cstr ());
  } else if (rc != ssize_t (len)) {
    warn ("short write for file %s\n", _file.cstr ());
  } else {
    ok = true;
  }
  return ok;
}

//-----------------------------------------------------------------------

void
file_t::read_chunk (const rtftp_chunkid_t &chnk, rtftp_get2_res_t *res)
{
  res->set_status (RTFTP_OK);
  res->chunk->data.setsize (chnk.size);
  off_t o = chnk.offset + _header_offset;
  int rc = 0;
  if (o != _offset) {
    rc = lseek (_fd, o, SEEK_SET);
    if (rc < 0) {
      warn ("seek failed on file %s: %m\n", _file.cstr ());
      res->set_status (RTFTP_EFS);
    }
  }
  if (rc >= 0) {
    int rc = read (_fd, res->chunk->data.base (), chnk.size);
    if (rc == 0) {
      res->set_status (RTFTP_EOF);
    } else if (rc < 0) {
      warn ("read error on file %s: %m\n", _file.cstr ());
      res->set_status (RTFTP_EFS);
    } else {
      _offset += rc;
      res->chunk->data.setsize (rc);
      res->chunk->id = chnk;
    }
  }
}

//-----------------------------------------------------------------------

void
cli_t::do_put2 (const rtftp_put2_arg_t &arg, rtftp_put2_res_t *res)
{
  switch (arg.status) {
  case RTFTP_BEGIN:
    {
      ptr<file_t> f = New refcounted<file_t> (*arg.name, file_t::FILE_PUT);
      if (!f->open (O_WRONLY|O_CREAT) || !f->write_header_placeholder ()) {
	str s = *arg.name;
	warn ("failed to write file %s: %m\n", s.cstr ());
	res->set_status (RTFTP_EFS);
      } else {
	rtftp_xfer_id_t id = _id ++;
	f->set_id (id);
	_tab.insert (id, f);
	res->set_status (RTFTP_BEGIN);
	*res->xfer_id = id;
      }
    }
    if (_verbose) {
      warn << "PUT2: " << *arg.name << " -> ";
      rpc_print (warnx, res->status, 0, NULL, NULL);
      warnx << "\n";
    }
    break;

  case RTFTP_OK:
    {
      rtftp_xfer_id_t id = arg.data->id.xfer_id;
      ptr<file_t> *f = _tab[id];
      if (f) {
	rtftp_status_t s = (*f)->write_chunk (*arg.data);
	res->set_status (s);
      } else {
	res->set_status (RTFTP_NOENT);
      }
    }
    break;

  case RTFTP_EOF:
    {
      rtftp_xfer_id_t id = arg.footer->xfer_id;
      ptr<file_t> *f = _tab[id];
      if (f) {
	rtftp_status_t s = (*f)->put_footer (*arg.footer);
	res->set_status (s);
	(*f)->set_put_status (s);
      } else {
	res->set_status (RTFTP_NOENT);
      }
    }
    break;

  default:
    res->set_status (RTFTP_ERR);
    break;
  }
}

//-----------------------------------------------------------------------

rtftp_status_t
file_t::put_footer (const rtftp_footer_t &footer)
{
  rtftp_status_t st;
  char buf[RTFTP_HASHSZ];
  _ctx.final (buf);
  if (_sz != footer.size) {
    warn << "wrong number of bytes in file " << _file << "; "
	 << "expected " << _sz << " but got " << footer.size << "\n";
    st = RTFTP_CORRUPT;
  } else if (memcmp (footer.hash.base (), buf, RTFTP_HASHSZ) != 0) {
    warn << "hash mismatch for file " << _file << "\n";
    st = RTFTP_CORRUPT;
  } else {
    st = RTFTP_OK;
    off_t rc = lseek (_fd, 0, SEEK_SET);
    if (rc < 0) {
      warn ("cannot seek through file %s: %m\n", _file.cstr ());
      st = RTFTP_EFS;
    } else {
      rtftp_header_t h;
      h.name = _file;
      h.magic = MAGIC;
      h.size = _sz;
      memcpy (h.hash.base (), buf, RTFTP_HASHSZ);
      st = write_header (h) ? RTFTP_OK : RTFTP_EFS;
    }
  }
  return st;
}

//-----------------------------------------------------------------------

bool
file_t::read_header ()
{
  size_t sz = 0;
  size_t szsz = sizeof (sz);
  bool ok = false;
  
  ssize_t rc = read (_fd, (void *)&sz, szsz);
  if (rc < 0) {
    warn ("read error on file %s: %m\n", _file.cstr ());
  } else if (rc != ssize_t (szsz)) {
    warn ("cannot read file size on file %s\n", _file.cstr ());
  } else if (sz > 0x10000) {
    warn ("size is way too big");
  } else {
    _offset += szsz;
    char *buf = New char[sz];
    rc = read (_fd, buf, sz);
    if (rc != ssize_t (sz)) {
      warn ("cannot read header on file %s\n", _file.cstr ());
    } else {
      rtftp_header_t h;
      if (!buf2xdr (h, buf, sz)) {
	warn ("cannot read header on file %s\n", _file.cstr ());
      } else if (h.magic != MAGIC) {
	warn ("corrupted magic on file %s\n", _file.cstr ());
      } else {
	_offset += sz;
	memcpy (_expected_hash, h.hash.base (), RTFTP_HASHSZ);
	_expected_sz = h.size;
	ok = true;
	_header_offset = _offset;
      }
    }
    delete [] buf;
  }
  return ok;
}


//-----------------------------------------------------------------------

void
file_t::set_xfer_header (rtftp_get2_res_t *res)
{
  res->set_status (RTFTP_BEGIN);
  res->header->xfer_id = _id;
  res->header->size = _expected_sz;
  memcpy (res->header->hash.base (), _expected_hash, RTFTP_HASHSZ);
}

//-----------------------------------------------------------------------

void
cli_t::do_get2 (const rtftp_get2_arg_t &arg, rtftp_get2_res_t *res)
{
  switch (arg.status) {

  case RTFTP_BEGIN:
    {
      ptr<file_t> f = New refcounted<file_t> (*arg.name, file_t::FILE_GET);
      if (!f->open (O_RDONLY)) {
	res->set_status (RTFTP_NOENT);
      } else if (!f->read_header ()) {
	res->set_status (RTFTP_CORRUPT);
      } else {
	rtftp_xfer_id_t id = _id ++;
	f->set_id (id);
	_tab.insert (id, f);
	f->set_xfer_header (res);
      }
      if (_verbose) {
	warn << "GET2: " << *arg.name << " -> ";
	rpc_print (warnx, res->status, 0, NULL, NULL);
	warnx << "\n";
      }
    }
    break;

  case RTFTP_OK:
    {
      rtftp_xfer_id_t id = arg.chunk->xfer_id;
      ptr<file_t> *f = _tab[id];
      if (f) {
	(*f)->read_chunk (*arg.chunk, res);
      } else {
	res->set_status (RTFTP_NOENT);
      }
    }
    break;

  case RTFTP_EOF:
    {
      rtftp_xfer_id_t id = *arg.id;
      if (_tab[id]) {
	_tab.remove (id);
	res->set_status (RTFTP_OK);
      } else {
	res->set_status (RTFTP_NOENT);
      }
    }
    break;

  default:
    res->set_status (RTFTP_ERR);
    break;
  }
}


//-----------------------------------------------------------------------

void
cli_t::do_get (const rtftp_id_t &arg, rtftp_get_res_t *res)
{
  str d = file2str (arg);
  res->set_status (RTFTP_OK);
  if (!d) {
    res->set_status (RTFTP_NOENT);
  } else if (!str2xdr (*res->file, d)) {
    warn << "cannot un-XDR file: " << arg << "\n";
    res->set_status (RTFTP_CORRUPT);
  } else if (!check_file (*res->file)) {
    warn << "checksum failed on read file: " << arg << "\n";
    res->set_status (RTFTP_CORRUPT);
  }
  if (_verbose) {
    warn << "GET: " << arg << " -> ";
    rpc_print (warnx, res->status, 0, NULL, NULL);
    warnx << "\n";
  }
}

//-----------------------------------------------------------------------

void
cli_t::dispatch (svccb *sbp)
{
  if (!sbp) {
    _srv = NULL;
    _x = NULL;
    delete this;
    return;
  }

  switch (sbp->proc ()) {
  case RTFTP_NULL: 
    {
      RPC::rtftp_program_1::rtftp_null_srv_t<svccb> srv (sbp);
      srv.reply ();
      break;
    }

  case RTFTP_PUT:
    {
      RPC::rtftp_program_1::rtftp_put_srv_t<svccb> srv (sbp);
      rtftp_status_t s = do_put (srv.getarg ());
      srv.reply (s);
      break;
    }

  case RTFTP_GET:
    {
      rtftp_get_res_t res;
      RPC::rtftp_program_1::rtftp_get_srv_t<svccb> srv (sbp);
      do_get (*srv.getarg (), &res);
      srv.reply (res);
      break;
    }

  case RTFTP_GET2:
    {
      rtftp_get2_res_t res;
      RPC::rtftp_program_1::rtftp_get2_srv_t<svccb> srv (sbp);
      do_get2 (*srv.getarg (), &res);
      srv.reply (res);
      break;
    }

  case RTFTP_PUT2:
    {
      rtftp_put2_res_t res;
      RPC::rtftp_program_1::rtftp_put2_srv_t<svccb> srv (sbp);
      do_put2 (*srv.getarg (), &res);
      srv.reply (res);
      break;
    }

  default: 
    {
      sbp->reject (PROC_UNAVAIL);
      break;
    }
  }
}

//-----------------------------------------------------------------------

void
srv_t::newcli ()
{
  sockaddr_in sin;
  socklen_t sinlen = sizeof (sockaddr_in);
  bzero (&sin, sinlen);
  
  int fd = accept (_sfd, reinterpret_cast<sockaddr *> (&sin), &sinlen);
  if (fd < 0) {
    warn ("accept error: %m\n");
  } else {
    vNew cli_t (fd, inet_ntoa (sin.sin_addr), _verbose);
  }
}

//-----------------------------------------------------------------------

void
srv_t::shutdown (int sig)
{
  warn << "shutting down with signal=" << sig << "\n";
  exit (0);
}


//-----------------------------------------------------------------------

bool
srv_t::init_signals ()
{
  sigcb (SIGTERM, wrap (this, &srv_t::shutdown, SIGTERM));
  sigcb (SIGQUIT, wrap (this, &srv_t::shutdown, SIGQUIT));
  sigcb (SIGINT , wrap (this, &srv_t::shutdown, SIGINT));
  return true;
}

//-----------------------------------------------------------------------

bool
srv_t::init ()
{
  return 
    init_signals () && 
    init_port () && 
    init_chroot () &&
    init_daemon () &&
    init_perms ();
}

//-----------------------------------------------------------------------

static int
gname2gid (const str &g)
{
  struct group *gr;
  int ret = -1;
  if ((gr = getgrnam (g))) {
    ret = gr->gr_gid;
  }
  endgrent ();
  return ret;
}

//-----------------------------------------------------------------------

static int
uname2uid (const str &n)
{
  struct passwd *pw;
  int ret = -1;
  if ((pw = getpwnam (n))) {
    ret = pw->pw_uid;
  }
  endpwent ();
  return ret;
}

//-----------------------------------------------------------------------

bool
srv_t::init_port ()
{
  bool rc = true;
  u_int32_t addr = INADDR_ANY;
  u_int16_t port = _port;
  int type = SOCK_STREAM;
  _sfd = inetsocket (type, port, addr);
  if (_sfd < 0) {
    warn ("cannot bind to port %d: %m\n", _port);
    rc = false;
  }
  return rc;
}

//-----------------------------------------------------------------------

bool
srv_t::init_perms ()
{
  bool rc = false;
  if (_chroot) {
    if (chroot (_dir.cstr ()) != 0) {
      warn ("Cannot chroot to directory %s: %m\n", _dir.cstr ());
    } else if (setgid (_gid) != 0) {
      warn ("Cannot set group id to %s/%d: %m\n", _group.cstr (), _gid);
    } else if (setuid (_uid) != 0) {
      warn ("Cannot set user id to %s/%d: %m\n", _user.cstr (), _uid);
    } else if (chdir ("/") != 0) {
      warn ("Cannot chdir to top directory after chroot: %m\n");
    } else {
      rc = true;
    }
  } else {
    rc = true;
  }
  return rc;
}

//-----------------------------------------------------------------------

bool
srv_t::init_chroot ()
{
  bool rc = false;
  if (access (_dir.cstr (), X_OK|R_OK) != 0) {
    warn ("Cannot access directory %s: %m\n", _dir.cstr ());
  } else if (getuid () == 0) {
    if (_chroot) {
      if ((_uid  = uname2uid (_user)) < 0) {
	warn ("Cannot find user %s\n", _user.cstr ());
      } else if ((_gid = gname2gid (_group)) < 0) {
	warn ("Cannot find group %s\n", _group.cstr ());
      } else {
	rc = true;
      }
    } else {
      warn ("Cannot run as root without chroot!\n");
    }
  } else if (chdir (_dir.cstr ()) != 0) {
    warn ("cannot change to data directory %s: %m\n", _dir.cstr ());
  } else {
    rc = true;
  }
  return rc;
}

//-----------------------------------------------------------------------

bool
srv_t::init_daemon ()
{
  if (!_debug) {
    daemonize ("rtftpd");
  }
  return true;
}

//-----------------------------------------------------------------------

void
srv_t::run ()
{
  warn << "starting up (pid=" << getpid () << ")\n";
  if (_verbose) {
    if (_user && _group) {
      warn ("running in dir=%s as user=%s and group=%s\n",
	    _dir.cstr (), _user.cstr (), _group.cstr ());
    } else {
      warn ("running unprivileged in dir=%s\n", _dir.cstr ()); 
    }
  }
  listen (_sfd, 200);
  fdcb (_sfd, selread, wrap (this, &srv_t::newcli));
}

//-----------------------------------------------------------------------

int
srv_t::config (int argc, char *argv[])
{
  int ch;
  int rc = 1;

  while ((ch = getopt (argc, argv, "vdru:g:l:")) != -1) {
    switch (ch) {
    case 'l':
      syslog_priority = optarg;
      break;
    case 'r':
      _chroot = true;
      break;
    case 'u':
      _user = optarg;
      break;
    case 'g':
      _group = optarg;
      break;
    case 'd':
      _debug = true;
      break;
    case 'v':
      _verbose = true;
      break;
    case 'p':
      if (!convertint (optarg, &_port)) {
	usage ();
	rc = -1;
	break;
      }
    case 'h':
      rc = 0;
      usage ();
      break;
    default:
      usage ();
      rc = -1;
      break;
    }
  }

  if (_chroot && (!_user || !_group)) {
    warn << "With -r, need to supply -u<user> and -g<group>\n";
    rc = -1;
  }

  if (rc > 0) {
    argc -= optind;
    argv += optind;
    if (argc != 1) {
      usage ();
      rc = -1;
    } else {
      _dir = argv[0];
    }
  }
  return rc;
}

//-----------------------------------------------------------------------

int
main (int argc, char *argv[])
{
  srv_t srv;
  int rc;
  setprogname (argv[0]);

  if ((rc = srv.config (argc, argv)) <= 0) {
    return rc;
  }

  if (!srv.init ()) {
    return -1;
  }

  srv.run ();

  amain ();
  return 0;
}

//-----------------------------------------------------------------------
