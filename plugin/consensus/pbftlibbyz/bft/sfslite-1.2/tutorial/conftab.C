
#include "parseopt.h"

void
okd_t::got_pubd_unix (vec<str> s, str loc, bool *errp)
{
  str name;
  if (s.size () != 2 || access (s[1], R_OK) != 0) {
    warn << loc << ": usage: PubdUnix <socketpath>\n";
    *errp = true;
  } else if (!is_safe (s[1])) {
    warn << loc << ": Pubd socket path (" << s[1]
	 << ") contains unsafe substrings\n";
    *errp = true;
  } else {
    pubd = New helper_unix_t (pub_program_1, s[1]);
  }
}

void
okd_t::got_pubd_exec (vec<str> s, str loc, bool *errp)
{
  if (s.size () <= 1) {
    warn << loc << ": usage: PubdExecPath <path-to-pubd>\n";
    *errp = true;
    return;
  } else if (!is_safe (s[1])) {
    warn << loc << ": pubd exec path (" << s[1] 
	 << ") contains unsafe substrings\n";
    *errp = true;
    return;
  }
  str prog = okws_exec (s[1]);
  str err = can_exec (prog);
  if (err) {
    warn << loc << ": cannot open pubd: " << err << "\n";
    *errp = true;
  } else {
    s.pop_front ();
    s[0] = prog;
    pubd = New helper_exec_t (pub_program_1, s);
  }
}

void
okd_t::parseconfig ()
{
  const str &cf = configfile;
  warn << "using config file: " << cf << "\n";
  parseargs pa (cf);
  bool errors = false;

  int line;
  vec<str> av;

  str un, gn;
  conftab ct;

  ct.add ("TopDir", &topdir) // set a string
    .add ("SyscallStatDumpInterval", &ok_ssdi, 0, 1000) // set an int
    .add ("PubdUnix", wrap (this, &okd_t::got_pubd_unix)) // trigger a func
    .add ("PubdInet", wrap (this, &okd_t::got_pubd_inet))
    .ignore ("MmapClockDaemon")  // ignore these commands
    .ignore ("Service");

  while (pa.getline (&av, &line)) {
    if (!ct.match (av, cf, line, &errors)) {
      warn << cf << ":" << line << ": unknown config parameter\n";
      errors = true;
    }
  }

  // will be set if any of the config lines failed
  if (errors)
    exit (1);
}
