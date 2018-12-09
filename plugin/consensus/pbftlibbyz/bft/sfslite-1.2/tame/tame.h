// -*-c++-*-
/* $Id: tame.h 3131 2007-11-17 05:00:05Z max $ */

/*
 *
 * Copyright (C) 2005 Max Krohn (email: my last name AT mit DOT edu)
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU General Public License as
 * published by the Free Software Foundation; either version 2, or (at
 * your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 59 Temple Place, Suite 330, Boston, MA 02111-1307
 * USA
 *
 */

#ifndef _TAME_TAME_H
#define _TAME_TAME_H

/*
 * netdb.h hack
 *
 * There is a conflict betwen flex version 2.5.4 and /usr/include/netdb.h
 * on Linux. flex generates code that #define's __unused to be empty, but
 * in struct gaicb within netdb.h, there is a member field of a struct
 * called __unused, which gets blanked out, causing a compile error.
 * (Note that netdb.h is included from sysconf.h).  Use this hack to
 * not include netdb.h for now...
 */
#ifndef _NETDB_H
# define _SKIP_NETDB_H
# define _NETDB_H
#endif

#include "amisc.h"

#ifdef _SKIP_NETDB_H
# undef _NETDB_H
# undef _SKIP_NETDB_H
#endif
/*
 * end netdb.h hack
 */

#include "vec.h"
#include "union.h"
#include "qhash.h"
#include "list.h"
#include "ihash.h"
#include <string.h>

extern int yylex ();
extern int yyparse ();
#undef yyerror
extern int yyerror (str);

extern int yyparse ();
extern int yydebug;
extern FILE *yyin;
extern int get_yy_lineno ();
extern str get_yy_loc ();

typedef enum { NONE = 0, ARG = 1, STACK = 2, CLASS = 3, EXPR = 4 } vartyp_t ;

str ws_strip (str s);
str template_args (str s);

class my_strbuf_t : public strbuf {
public:
  strbuf & mycat (const str &s) { cat (s.cstr (), true); return (*this); }
  void clear () { tosuio ()->clear (); }
};


class lstr : public str {
public:
  lstr () : str (), _lineno (0) {}
  lstr (const char *c) : str (c), _lineno (0) {}
  lstr (u_int ln, const str &s) : str (s), _lineno (ln) {}
  lstr (u_int ln) : str (""), _lineno (ln) {}
  void set_lineno (u_int l) { _lineno = l; }
  u_int lineno () const { return _lineno; }
private:
  u_int _lineno;
};


#define HOLDVAR_FLAG     (1 << 0)
#define CONST_FLAG       (2 << 0)
/*
 * object for holding type modifiers like "unsigned", "const",
 * "typename", etc, but also for holding some flags that should
 * never be output, that are internal to tame. Examples include
 * 'holdvar'.
 */
struct type_qualifier_t {
  type_qualifier_t () : _flags (0) {}
  type_qualifier_t (const type_qualifier_t &t) 
    : _lineno (t._lineno), _v (t._v), _flags (t._flags) {}
  type_qualifier_t (const lstr &s, u_int f = 0) 
    : _flags (f) { if (s) add_lstr (s); }

  void add_str (const str &s) { _v.push_back (s); }
  void add_lstr (const lstr &s) { add_str (s); _lineno = s.lineno (); }
  void add_flag (u_int f) { _flags = _flags | HOLDVAR_FLAG; }

  type_qualifier_t &concat (const type_qualifier_t &m);

  str to_str () const;
  u_int flags () const { return _flags; }
  u_int lineno () const { return _lineno; }
  
private:
  u_int _lineno;
  vec<str> _v;
  u_int _flags;
};


typedef enum { OUTPUT_NONE = 0,
	       OUTPUT_PASSTHROUGH = 1,
	       OUTPUT_TREADMILL = 2,
	       OUTPUT_BIG_NEW_CHUNK = 3 } output_mode_t;

class outputter_t {
public:
  outputter_t (const str &in, const str &out, bool ox) 
    : _mode (OUTPUT_NONE), _infn (in), _outfn (out), _fd (-1), 
      _lineno (1), _output_xlate (ox), _need_xlate (false), 
      _last_char_was_nl (false), _last_output_in_mode (OUTPUT_NONE),
      _last_lineno (-1), _did_output (false), 
      _do_output_line_number (false) {}
  virtual ~outputter_t ();
  bool init ();

  void start_output ();
  void flush ();

  output_mode_t switch_to_mode (output_mode_t m, int ln = -1);
  output_mode_t mode () const { return _mode; }
  virtual void output_str (str s);
protected:
  void output_line_number ();
  void _output_str (str s, str sep_str = NULL);
  output_mode_t _mode;
private:
  const str _infn, _outfn;
  int _fd;
  int _lineno;
  bool _output_xlate;

  strbuf _buf;
  vec<str> _strs;
  bool _need_xlate;
  bool _last_char_was_nl;
  output_mode_t _last_output_in_mode;
  int _last_lineno;
  bool _did_output;
  bool _do_output_line_number;
};

/**
 * Horizontal outputter -- doesn't output newlines when in treadmill mode
 */
class outputter_H_t : public outputter_t {
public:
  outputter_H_t (const str &in, const str &out, bool ox) 
    : outputter_t (in, out, ox) {}
protected:
  void output_str (str s);
};

class tame_el_t {
public:
  tame_el_t () {}
  virtual ~tame_el_t () {}
  virtual bool append (const str &s) { return false; }
  virtual void output (outputter_t *o) = 0;
  virtual bool goes_after_vars () const { return true; }
  tailq_entry<tame_el_t> _lnk;
};

class element_list_t : public tame_el_t {
public:
  virtual void output (outputter_t *o);
  void passthrough (const lstr &l) ;
  void push (tame_el_t *e) { push_hook (e); _lst.insert_tail (e); }
  virtual void push_hook (tame_el_t *e) {}
protected:
  tailq<tame_el_t, &tame_el_t::_lnk> _lst;
};

class tame_env_t : public element_list_t {
public:
  virtual void output (outputter_t *o) { element_list_t::output (o); }
  virtual bool is_jumpto () const { return false; }
  virtual void set_id (int id) {}
  virtual int id () const { return 0; }
  virtual bool needs_counter () const { return false; }
};

class tame_fn_t; 

class tame_vars_t : public tame_el_t {
public:
  tame_vars_t (tame_fn_t *fn, int lineno) : _fn (fn), _lineno (lineno) {}
  void output (outputter_t *o) ;
  bool goes_after_vars () const { return false; }
  int lineno () const { return _lineno; }
private:
  tame_fn_t *_fn;
  int _lineno;
};

class tame_passthrough_t : public tame_el_t {
public:
  tame_passthrough_t (const lstr &s) { append (s); }
  bool append (const lstr &s) { _strs.push_back (s); _buf << s; return true; }
  void output (outputter_t *o);
private:
  strbuf _buf;
  vec<lstr> _strs;
};

class type_t {
public:
  type_t () {}
  type_t (const str &t, const str &p)
    : _base_type (t), _pointer (p) {}
  type_t (const str &t, const str &p, const str &ta)
    : _base_type (t), _pointer (p), _template_args (ta) {}
  str base_type () const { return _base_type; }
  str pointer () const { return _pointer; }
  str to_str () const;
  str to_str_w_template_args (bool p = true) const;
  str mk_ptr () const;
  str alloc_ptr (const str &nm, const str &args) const;
  str type_without_pointer () const;
  void set_base_type (const str &t) { _base_type = t; }
  void set_pointer (const str &p) { _pointer = p; }
  bool is_complete () const { return _base_type; }
  bool is_void () const 
  { return (_base_type == "void" && (!_pointer || _pointer.len () == 0)); } 
  bool is_ref () const { return _pointer && strchr (_pointer.cstr (), '&'); }
private:
  str _base_type, _pointer, _template_args;
};

class initializer_t {
public:
  initializer_t () : _value (0, NULL) {}
  initializer_t (const lstr &v) : _value (v) {}
  virtual ~initializer_t () {}
  virtual str output_in_constructor () const { return ""; }
  virtual str output_in_declaration () const { return ""; }
  virtual bool do_constructor_output () const { return false; }
  virtual str ref_prefix () const;
protected:
  lstr _value;
};

class cpp_initializer_t : public initializer_t {
public:
  cpp_initializer_t (const lstr &v) : initializer_t (v) {}
  str output_in_constructor () const;
  bool do_constructor_output () const { return true; }
};

class array_initializer_t : public initializer_t {
public:
  array_initializer_t (const lstr &v) : initializer_t (v) {}
  str output_in_declaration () const;
  str ref_prefix () const;
};

class declarator_t;
class var_t {
public:
  var_t () {}
  var_t (const str &n, vartyp_t a = NONE) : _name (n), _asc (a), _flags (0) {}
  var_t (const type_qualifier_t &m, ptr<declarator_t> d, vartyp_t a = NONE);
  var_t (const str &t, const str &p, const str &n, vartyp_t a = NONE) : 
    _type (t, p), _name (n), _asc (a), _flags (0) {}
  var_t (const type_t &t, const str &n, vartyp_t a = NONE)
    : _type (t), _name (n), _asc (a), _flags (0) {}
  var_t (const str &t, const str &p, const str &n, vartyp_t a, const str &ta)
    : _type (t, p, ta), _name (n), _asc (a), _flags (0) {}
protected:
  type_t _type;

public:
  const type_t &type () const { return _type; }
  const str &name () const { return _name; }
  type_t *get_type () { return &_type; }
  const type_t * get_type_const () const { return &_type; }
  bool is_complete () const { return _type.is_complete (); }

  // ASC = Args, Stack or Class
  void set_asc (vartyp_t a) { _asc = a; }
  vartyp_t get_asc () const { return _asc; }
  void reset_name (const str &s) { _name = s; }

  void set_type (const type_t &t) { _type = t; }
  ptr<initializer_t> initializer () const { return _initializer; }
  bool do_output () const;

  str decl () const;
  str decl (const str &prfx, int n) const;
  str decl (const str &prfx) const;
  str ref_decl () const;
  str _name;

protected:
  vartyp_t _asc;
  ptr<initializer_t> _initializer;
  u_int _flags;
};

class expr_list_t : public vec<var_t>
{
public:
  bool output_vars (strbuf &b, bool first = false, const str &prfx = NULL,
		    const str &sffx = NULL);
};

typedef enum { DECLARATIONS, NAMES, TYPES } list_mode_t;

class vartab_t {
public:
  ~vartab_t () {}
  vartab_t () {}
  vartab_t (var_t v) { add (v); }
  size_t size () const { return _vars.size (); }
  bool add (var_t v) ;
  void declarations (strbuf &b, const str &padding) const;
  void paramlist (strbuf &b, list_mode_t m, str prfx = NULL) const;
  void initialize (strbuf &b, bool self) const;
  bool exists (const str &n) const { return _tab[n]; }
  const var_t *lookup (const str &n) const;
  void copy_in (const vartab_t &vt);

  vec<var_t> _vars;
  qhash<str, u_int> _tab;
};


class tame_block_t;
class tame_nonblock_t;
class tame_join_t;
class tame_fork_t;

/*
 * corresponds to the yacc rule for parsing C declarators -- either for
 * variables or for function definitions
 */
class declarator_t {
public:
  declarator_t (const str &n, ptr<vartab_t> v)
    : _name (n), _params (v) {}
  declarator_t (const str &n) : _name (n) {}
  void set_pointer (const str &s) { _pointer = s; }
  str pointer () const { return _pointer; }
  str name () const { return _name; }
  ptr<vartab_t> params () { return _params; }
  void set_params (ptr<vartab_t> v) { _params = v; }
  void set_initializer (ptr<initializer_t> i) { _initializer = i; }
  void dump () const ;
  ptr<initializer_t> initializer () const { return _initializer; }
private:
  const str _name;
  str _pointer;
  ptr<vartab_t> _params;
  ptr<initializer_t> _initializer;
};

// Convert:
//
//   foo_t::max<int,int> => foo_t__max_int_int_
//
str mangle (const str &in);

//
// convert 
//
//   foo_t::max<int,int>::bar  => bar
//
str strip_to_method (const str &in);
str strip_off_method (const str &in);

class tame_block_t;

#define STATIC_DECL           (1 << 0)
#define CONST_DECL            (2 << 0)

// function specifier embodies static and template keywords and options
// at present, and perhaps more in the future.
struct fn_specifier_t {
  fn_specifier_t (u_int o = 0, str t = NULL)
    : _opts (o), _template (t) {}
  u_int _opts;
  str _template;
};

//
// Unwrap Function Type
//
class tame_fn_t : public element_list_t {
public:
  tame_fn_t (const fn_specifier_t &fn, const str &r, ptr<declarator_t> d, 
	     bool c, u_int l, str loc)
    : _ret_type (ws_strip (r), 
		 d->pointer () ? ws_strip (d->pointer ()) : NULL), 
      _name (d->name ()),
      _name_mangled (mangle (_name)), 
      _method_name (strip_to_method (_name)),
      _class (strip_off_method (_name)), 
      _self (c ? str (strbuf ("const ") << _class) : _class, "*", "_self"),
      _isconst (c),
      _template (fn._template),
      _template_args (_class ? template_args (_class) : NULL),
      _closure (mk_closure ()), 
      _args (d->params ()), 
      _opts (fn._opts),
      _lineno (l),
      _n_labels (0),
      _n_blocks (0),
      _loc (loc),
      _lbrace_lineno (0),
      _vars (NULL),
      _after_vars_el_encountered (false)
  { }

  vartab_t *stack_vars () { return &_stack_vars; }
  vartab_t *args () { return _args; }
  vartab_t *class_vars_tmp () { return &_class_vars_tmp; }

  void push_hook (tame_el_t *el)
  {
    if (el->goes_after_vars ())
      _after_vars_el_encountered = true;
  }

  // called from tame_vars_t class
  void output_vars (outputter_t *o, int ln);

  // default return statement is "return;"; can be overidden,
  // but only once.
  bool set_default_return (str s) 
  { 
    bool ret = _default_return ? false : true;
    _default_return = s; 
    return ret;
  }

  // if non-void return, then there must be a default return
  bool check_return_type () const 
  { return (_ret_type.is_void () || _default_return); }

  str classname () const { return _class; }
  str name () const { return _name; }
  str signature (bool decl, str prfx = NULL, bool sttc = false) const;

  void set_opts (int i) { _opts = i; }
  int opts () const { return _opts; }

  bool need_self () const { return (_class && !(_opts & STATIC_DECL)); }

  void jump_out (strbuf &b, int i);

  void output (outputter_t *o);

  void add_env (tame_env_t *g) ;

  str fn_prefix () const { return _name_mangled; }

  static var_t closure_generic () ;
  str decl_casted_closure (bool do_lhs) const;
  var_t closure () const { return _closure; }
  static var_t trig () ;

  void hit_tame_block () { _n_blocks++; }

  str closure_nm () const { return _closure.name (); }
  str reenter_fn  () const ;
  str frozen_arg (const str &i) const ;

  str label (str s) const;
  str label (u_int id) const ;
  str loc () const { return _loc; }

  str return_expr () const;

  str template_str () const
  { return (_template ? str (strbuf ("template< " ) << _template << " >") 
	    : NULL); }

  void set_lbrace_lineno (u_int i) { _lbrace_lineno = i ; }

  bool set_vars (tame_vars_t *v) 
  { _vars = v; return (!_after_vars_el_encountered); }
  const tame_vars_t *get_vars () const { return _vars; }

private:
  const type_t _ret_type;
  const str _name;
  const str _name_mangled;
  const str _method_name;
  const str _class;
  const var_t _self;

  const bool _isconst;
  str _template;
  str _template_args;
  const var_t _closure;

  var_t mk_closure () const ;

  ptr<vartab_t> _args;
  vartab_t _stack_vars;
  vartab_t _class_vars_tmp;
  vec<tame_env_t *> _envs;

  void output_reenter (strbuf &b);
  void output_closure (outputter_t *o);
  void output_fn (outputter_t *o);
  void output_static_decl (outputter_t *o);
  void output_stack_vars (strbuf &b);
  void output_arg_references (strbuf &b);
  void output_jump_tab (strbuf &b);
  void output_set_method_pointer (my_strbuf_t &b);
  void output_block_cb_switch (strbuf &b);
  
  int _opts;
  u_int _lineno;
  u_int _n_labels;
  u_int _n_blocks;
  str _default_return;
  str _loc; // filename:linenumber where this function was declared
  u_int _lbrace_lineno;  // void foo () { ... where the '{' was
  tame_vars_t *_vars;
  bool _after_vars_el_encountered;
};


class tame_ret_t : public tame_env_t 
{
public:
  tame_ret_t (u_int l, tame_fn_t *f) : _line_number (l), _fn (f) {}
  void add_params (const lstr &l) { _params = l; }
  virtual void output (outputter_t *o);
protected:
  u_int _line_number;
  tame_fn_t *_fn;
  lstr _params;
};

class tame_fn_return_t : public tame_ret_t {
public:
  tame_fn_return_t (u_int l, tame_fn_t *f) : tame_ret_t (l, f) {}
  void output (outputter_t *o);
};

class tame_unblock_t : public tame_ret_t {
public:
  tame_unblock_t (u_int l, tame_fn_t *f) : tame_ret_t (l, f) {}
  virtual ~tame_unblock_t () {}
  void output (outputter_t *o);
  virtual str macro_name () const { return "SIGNAL"; }
  virtual void do_return_statement (my_strbuf_t &b) const {}
};

class tame_resume_t : public tame_unblock_t {
public:
  tame_resume_t (u_int l, tame_fn_t *f) : tame_unblock_t (l, f) {}
  ~tame_resume_t () {}
  str macro_name () const { return "RESUME"; }
  void do_return_statement (my_strbuf_t &b) const;
};

class parse_state_t : public element_list_t {
public:
  parse_state_t () : _xlate_line_numbers (false),
		     _need_line_xlate (true) 
  {
    _lists.push_back (this);
  }

  void new_fn (tame_fn_t *f) { new_el (f); _fn = f; }
  void new_el (tame_el_t *e) { _fn = NULL; push (e); }
  void set_fn (tame_fn_t *f) { _fn = f; }
  void clear_fn () { set_fn (NULL); }
  tame_fn_t *fn () { return _fn; }

  void passthrough (const lstr &l) { top_list ()->passthrough (l); }
  void push (tame_el_t *e) { top_list ()->push (e); }

  element_list_t *top_list () { return _lists.back (); }
  void push_list (element_list_t *l) { _lists.push_back (l); }
  void pop_list () { _lists.pop_back (); }

  // access variable tables in the currently active function
  vartab_t *stack_vars () { return _fn ? _fn->stack_vars () : NULL; }
  vartab_t *class_vars_tmp () { return _fn ? _fn->class_vars_tmp () : NULL ; }
  vartab_t *args () { return _fn ? _fn->args () : NULL; }

  void set_decl_specifier (const type_qualifier_t &m) { _decl_specifier = m; }
  const type_qualifier_t &decl_specifier () const { return _decl_specifier; }
  tame_fn_t *function () { return _fn; }
  const tame_fn_t &function_const () const { return *_fn; }

  void new_block (tame_block_t *g);
  tame_block_t *block () { return _block; }
  void clear_block () { _block = NULL; }

  void new_nonblock (tame_nonblock_t *s);
  tame_nonblock_t *nonblock () { return _nonblock; }

  void new_fork (tame_fork_t *f);
  tame_fork_t *fork () { return _fork; }

  void output (outputter_t *o);

  void clear_sym_bit () { _sym_bit = false; }
  void set_sym_bit () { _sym_bit = true; }
  bool get_sym_bit () const { return _sym_bit; }

  void set_infile_name (const str &i) { _infile_name = i; }
  str infile_name () const { return _infile_name; }
  str loc (u_int l) const ;

protected:
  type_qualifier_t _decl_specifier;
  tame_fn_t *_fn;
  tame_block_t *_block;
  tame_nonblock_t *_nonblock;
  tame_fork_t *_fork;
  bool _sym_bit;

  // lists of elements (to reflect nested structure)
  vec<element_list_t *> _lists;

  str _infile_name;
  bool _xlate_line_numbers;
  bool _need_line_xlate;
};

class tame_block_t : public tame_env_t {
public:
  tame_block_t (int l) : _lineno (l) {}
  virtual ~tame_block_t () {}
protected:
  int _lineno;
};

class tame_block_thr_t : public tame_block_t {
public:
  tame_block_thr_t (int l) : tame_block_t (l) {}
  void output (outputter_t *o);
};

class tame_block_ev_t : public tame_block_t {
public:
  tame_block_ev_t (tame_fn_t *f, int l) 
    : tame_block_t (l), _fn (f), _id (0) {}
  ~tame_block_ev_t () {}
  
  void output (outputter_t *o);
  bool is_jumpto () const { return true; }
  void set_id (int i) { _id = i; }
  int id () const { return _id; }
  void add_class_var (const var_t &v) { _class_vars.add (v); }
  bool needs_counter () const { return true; }
  
protected:
  tame_fn_t *_fn;
  int _id;
  vartab_t _class_vars;
};


  
class tame_nonblock_t : public tame_env_t {
public:
  tame_nonblock_t (ptr<expr_list_t> l) : _args (l) {}
  ~tame_nonblock_t () {}
  void output (outputter_t *o) { element_list_t::output (o); }
  ptr<expr_list_t> args () const { return _args; }
private:
  ptr<expr_list_t> _args;
};

class tame_join_t : public tame_env_t {
public:
  tame_join_t (tame_fn_t *f, ptr<expr_list_t> l) : _fn (f), _args (l) {}
  bool is_jumpto () const { return true; }
  void set_id (int i) { _id = i; }
  int id () const { return _id; }
  virtual void output (outputter_t *o) = 0;
  var_t join_group () const { return (*_args)[0]; }
  var_t arg (u_int i) const { return (*_args)[i+1]; }
  size_t n_args () const 
  { assert (_args->size () > 0); return _args->size () - 1; }
protected:
  void output_blocked (my_strbuf_t &b, const str &jgn);
  tame_fn_t *_fn;
  ptr<expr_list_t> _args;
  int _id;
};

class tame_fork_t : public tame_env_t {
public:
  tame_fork_t (tame_fn_t *f, ptr<expr_list_t> l) : _fn (f), _args (l) {}
  bool is_jumpto () const { return false; }
  void set_id (int i) { _id = i; }
  int id () const { return _id; }
  var_t join_group () const { return (*_args)[0]; }
  var_t arg (u_int i) const { return (*_args)[i+1]; }
  size_t n_args () const 
  { assert (_args->size () > 0); return _args->size () - 1; }
protected:
  tame_fn_t *_fn;
  ptr<expr_list_t> _args;
  int _id;
};

class tame_wait_t : public tame_join_t {
public:
  tame_wait_t (tame_fn_t *f, ptr<expr_list_t> l, int ln) 
    : tame_join_t (f, l), _lineno (ln) {}
  void output (outputter_t *o);
private:
  int _lineno;
};

extern parse_state_t *state;
extern str infile_name;

struct YYSTYPE {
  ::lstr            str;
  ptr<declarator_t> decl;
  ptr<vartab_t>     vars;
  var_t             var;
  bool              opt;
  char              ch;
  tame_fn_t *       fn;
  tame_el_t *       el;
  ptr<expr_list_t>  exprs;
  type_t            typ;
  vec<ptr<declarator_t> > decls;
  u_int             opts;
  tame_ret_t *       ret;
  fn_specifier_t     fn_spc;
  type_qualifier_t    typ_mod;
  ptr<initializer_t> initializer;
};
extern YYSTYPE yylval;
extern str filename;

#define CONCAT(ln,in,out)                                 \
do {                                                      \
      strbuf b;                                           \
      b << in;                                            \
      out = lstr (ln, b);                                 \
} while (0)


#endif /* _TAME_TAME_H */
