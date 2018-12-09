#include "tame.h"
#include "rxx.h"

u_int count_newlines (str s)
{
  const char *end = s.cstr () + s.len ();
  u_int c = 0;
  for (const char *cp = s.cstr (); cp < end; cp++) {
    if (*cp == '\n')
      c++;
  }
  return c;
}

bool
outputter_t::init ()
{
  if (_outfn && _outfn != "-") {
    if ((_fd = open (_outfn.cstr (), O_CREAT|O_WRONLY|O_TRUNC, 0644)) < 0) {
      warn << "cannot open file for writing: " << _outfn << "\n";
      return false;
    }
  } else {
    _fd = 1;
  }
  return true;
}

void
outputter_t::start_output ()
{
  _lineno = 1;

  // switching from NONE to PASSTHROUGH
  switch_to_mode (OUTPUT_PASSTHROUGH);
}

void
outputter_t::output_line_number ()
{
  if (_output_xlate &&
      (_did_output || _last_lineno != _lineno)) {
    strbuf b;
    if (!_last_char_was_nl)
      b << "\n";
    b << "# " << _lineno << " \"" << _infn << "\"\n";
    _output_str (b, false);
    _last_lineno = _lineno;
    _did_output = false;
  }
}

void
outputter_H_t::output_str (str s)
{
  if (_mode == OUTPUT_TREADMILL) {
    static rxx x ("\n");
    vec<str> v;
    output_line_number ();
    split (&v, x, s);
    for (u_int i = 0; i < v.size (); i++) {
      // only output a newline on the last line
      _output_str (v[i], (i == v.size () - 1 ? "\n" : " "));
    }
  } else {
    outputter_t::output_str (s);
  }
}

void
outputter_t::output_str (str s)
{
  if (_mode == OUTPUT_TREADMILL) {
    static rxx x ("\n");
    vec<str> v;
    split (&v, x, s);
    for (u_int i = 0; i < v.size (); i++) {
      output_line_number ();
      _output_str (v[i], "\n");
    }
  } else {

    // we might have set up a defered output_line_number from
    // within switch_to_mode; now is the time to do it.
    if (s.len () && _do_output_line_number) {
      output_line_number ();
      _do_output_line_number = false;
    }

    _output_str (s, false);
    if (_mode == OUTPUT_PASSTHROUGH)
      _lineno += count_newlines (s);
  }
}

void
outputter_t::_output_str (str s, str sep_str)
{
  if (!s.len ())
    return;

  _last_output_in_mode = _mode;
  _did_output = true;

  _strs.push_back (s);
  _buf << s;
  if (sep_str) {
    _buf << sep_str;
    _last_char_was_nl = (sep_str[sep_str.len () - 1] == '\n');
  } else {
    if (s && s.len () && s[s.len () - 1] == '\n') {
      _last_char_was_nl = true;
    } else {
      _last_char_was_nl = false;
    }
  }
}

void
outputter_t::flush ()
{
  _buf.tosuio ()->output (_fd);
}

outputter_t::~outputter_t ()
{
  if (_fd >= 0) {
    flush ();
    close (_fd); 
  }
}

output_mode_t 
outputter_t::switch_to_mode (output_mode_t m, int nln)
{
  output_mode_t old = _mode;
  int oln = _lineno;

  if (nln >= 0)
    _lineno = nln;
  else
    nln = oln;

  if (m == OUTPUT_PASSTHROUGH &&
      (oln != nln ||
       (old != OUTPUT_NONE && 
	_last_output_in_mode != OUTPUT_PASSTHROUGH))) {

    // don't call output_line_number() directly, since
    // maybe this will be an empty environment
    _do_output_line_number = true;
  }
  _mode = m;
  return old;
}
