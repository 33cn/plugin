/* -*-fundamental-*- */
/* $Id: parse.yy 2272 2006-10-27 02:25:59Z max $ */

/*
 *
 * Copyright (C) 1998 David Mazieres (dm@uun.org)
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

%{
#include "tame.h"
#include "parseopt.h"
#define YYSTYPE YYSTYPE
int vars_lineno;
%}

%token <str> T_ID
%token <str> T_NUM
%token <str> T_PASSTHROUGH

/* Tokens for C++ Variable Modifiers */
%token T_CONST
%token T_STRUCT
%token T_TYPENAME
%token T_VOID
%token T_CHAR
%token T_SHORT
%token T_INT
%token T_LONG
%token T_LONG_LONG
%token T_FLOAT
%token T_DOUBLE
%token T_SIGNED
%token T_UNSIGNED
%token T_STATIC
%token T_TEMPLATE
%token T_HOLDVAR

%token T_2COLON
%token T_RETURN

/* Keywords for our new filter */
%token T_TAME
%token T_TAMED
%token T_VARS
%token T_JOIN
%token T_TWAIT
%token T_DEFAULT_RETURN

%token T_2DOLLAR

%type <str> pointer pointer_opt template_instantiation_arg pointer_or_ref
%type <str> template_instantiation_list template_instantiation
%type <str> template_instantiation_opt typedef_name_single
%type <str> template_instantiation_list_opt identifier
%type <str> typedef_name
%type <str> type_specifier 
%type <str> passthrough 
%type <typ_mod> type_modifier type_modifier_list declaration_specifiers 
%type <typ_mod> type_qualifier_list_opt type_qualifier_list type_qualifier 
%type <initializer> cpp_initializer_opt

%type <decl> init_declarator declarator direct_declarator 
%type <decl> declarator_cpp direct_declarator_cpp


%type <vars> parameter_type_list_opt parameter_type_list parameter_list
%type <exprs> join_list id_list_opt id_list 

%type <opt>  const_opt
%type <fn>   fn_declaration tame_decl

%type <var>  parameter_declaration

%type <el>   fn_tame vars return_statement twait 
%type <el>   block_body twait_body wait_body
%type <el>   default_return

%type <opts> static_opt
%type <fn_spc> fn_specifiers template_decl

%%


file:  passthrough 			{ state->passthrough ($1); }
	| file fn_or_twait passthrough 	{ state->passthrough ($3); }
	;

fn_or_twait: fn	
	| twait				{ state->new_el ($1); }
	;

passthrough: /* empty */	    { $$ = lstr (get_yy_lineno ()); }
	| passthrough T_PASSTHROUGH 
	{
	   strbuf b ($1);
	   b << $2;
	   $$ = lstr ($1.lineno (), b);
	}
	;

tame_decl: T_TAME '(' fn_declaration ')'   { $$ = $3; }
	| T_TAMED fn_declaration 	   { $$ = $2; }
	;

fn:	tame_decl '{'
	{
	  state->new_fn ($1);
	  state->push_list ($1);
	  $1->set_lbrace_lineno (get_yy_lineno ());
	}
	fn_statements '}'
	{
	  if (!state->function ()->check_return_type ()) {
	    yyerror ("Function has non-void return type but no "
	    	     "DEFAULT_RETURN specified");
 	  }
	  state->push (New tame_fn_return_t (get_yy_lineno (), 
				            state->function ()));
	  state->passthrough (lstr (get_yy_lineno (), "}"));
	  state->pop_list ();
	  state->clear_fn ();
	}
	;

fn_specifiers:	template_decl { $$ = $1; }
	| T_STATIC	      { $$ = fn_specifier_t (STATIC_DECL); }
	| /* empty */	      { $$ = fn_specifier_t (0); }
	;

template_decl: T_TEMPLATE '<' passthrough '>' static_opt
	{
	   $$ = fn_specifier_t ($5, $3);
	}
	;

static_opt: T_STATIC 	{ $$ = STATIC_DECL; }
	| /* empty */	{ $$ = 0; }
	;

/* declaration_specifiers is no longer optional ?! */
fn_declaration: fn_specifiers declaration_specifiers declarator const_opt
	{
	   $$ = New tame_fn_t ($1, $2.to_str (), $3, $4, get_yy_lineno (), 
	                       get_yy_loc ());
	}
	;

const_opt: /* empty */		{ $$ = false; }
	| T_CONST		{ $$ = true; }
	;

fn_statements: passthrough			
	{
	  state->passthrough ($1);
	}
	| fn_statements fn_tame passthrough
	{
   	  if ($2) state->push ($2);
	  state->passthrough ($3);
	}
	;

fn_tame: vars
	| twait
	| return_statement
	| default_return
	;

default_return: T_DEFAULT_RETURN '{' passthrough '}'
	{
	  // this thing will not be output anywhere near where
	  // it's being input, so don't associate it in the 
	  // element list as usual.
	  if (!state->function ()->set_default_return ($3)) {
	    yyerror ("DEFAULT_RETURN specified more than once");
	  }
	  $$ = NULL;
	}
	;

vars:	T_VARS 
	{
	  vars_lineno = get_yy_lineno ();
	} 
	'{' declaration_list_opt '}'
	{
	  tame_vars_t *v = New tame_vars_t (state->function (), vars_lineno);
	  if (state->function ()->get_vars ()) {
	    strbuf b;
	    b << "VARS{} section specified twice (before on line " 
	      << state->function ()->get_vars ()->lineno () << ")\n";
	    yyerror (b);
	  }
	  if (!state->function ()->set_vars (v)) {
	    yyerror ("The tvars{} section must come before any twait "
	             " statement or environment");
	  }
	  $$ = v;
	}
	;

return_statement: T_RETURN passthrough ';'
	{
	   tame_ret_t *r = New tame_ret_t (get_yy_lineno (), 
			  	    state->function ());	
	   if ($2)
	     r->add_params ($2);
 	   r->passthrough (lstr (get_yy_lineno (), ";"));
	   $$ = r;
	}
	;

block_body: '{' 
	{
	  tame_fn_t *fn = state->function ();
	  tame_block_t *bb;
	  if (fn) {
 	    bb = New tame_block_ev_t (fn, get_yy_lineno ());
	    fn->add_env (bb);
	    fn->hit_tame_block ();
	  } else {
	    bb = New tame_block_thr_t (get_yy_lineno ());
	  }
	  state->new_block (bb);
	  state->push_list (bb);
	}
	fn_statements '}'
	{
 	  state->pop_list ();
	  $$ = state->block ();
	  state->clear_block ();
	}
	;

id_list_opt: /* empty */  { $$ = NULL; }
	| id_list
	;

id_list:  ',' identifier
	{
	  $$ = New refcounted<expr_list_t> ();
	  $$->push_back (var_t ()); // reserve 1 empty spot!
	  $$->push_back (var_t ($2, STACK));
	}
	| id_list ',' identifier
	{
	  $1->push_back (var_t ($3, STACK));
	  $$ = $1;
	}
	;

join_list: passthrough id_list_opt
	{
	  if ($2) {
	    (*$2)[0] = var_t ($1, EXPR);
	    $$ = $2;
	  } else {
	    $$ = New refcounted<expr_list_t> ();
	    $$->push_back (var_t ($1, EXPR));
	  }
	}
	;

wait_body: '(' join_list ')' ';'
	{
	  tame_fn_t *fn = state->function ();
	  tame_wait_t *w = New tame_wait_t (fn, $2, get_yy_lineno ());
	  if (fn) fn->add_env (w);
	  else {
	    yyerror ("Cannot have a twait() statement outside of a "
	 	     "tamed function body.");
	  }
	  $$ = w;
	}
	;

twait: T_TWAIT twait_body { $$ = $2; }
	;

twait_body: wait_body
	| block_body
	;

identifier: T_ID
	;

/*
 * No longer need casts..
 *
cast: '(' declaration_specifiers pointer_opt ')'
	{
	  $$ = type_t (ws_strip ($2), ws_strip ($3));
	}
	;
 */

declaration_list_opt: /* empty */
	| declaration_list
	;

declaration_list: declaration
	| declaration_list declaration
	;

parameter_type_list_opt: /* empty */ { $$ = NULL; }
	| parameter_type_list
	;

/* missing: '...'
 */
parameter_type_list: parameter_list
	;

parameter_list: parameter_declaration 
	{
	  $$ = New refcounted<vartab_t> ($1);
	}
	| parameter_list ',' parameter_declaration
	{
	  if (! ($1)->add ($3) ) {
	    strbuf b;
	    b << "duplicated parameter in param list: " << $3.name ();
	    yyerror (b);
          } else {
 	    $$ = $1;
          }
	}
	;

/* missing: abstract declarators
 */
parameter_declaration: declaration_specifiers declarator
	{
	  if ($2->params ()) {
	    warn << "parameters found when not expected\n";
	  }
	  $$ = var_t ($1, $2, ARG);
	}
	;

declaration: declaration_specifiers 
	{
	  state->set_decl_specifier ($1);
	}
	init_declarator_list_opt ';'
	;

init_declarator_list_opt: /* empty */
	| init_declarator_list
	;

init_declarator_list:  init_declarator			{}
	| init_declarator_list ',' init_declarator	{}
	;

/* missing: C++-style initialization, C-style initiatlization
 */
init_declarator: declarator_cpp cpp_initializer_opt
	{
	  assert ($2);
	  $1->set_initializer ($2);

	  vartab_t *t = state->stack_vars ();

	  var_t v (state->decl_specifier (), $1, STACK);
	  if (state->args () &&
              state->args ()->exists (v.name ())) {
	    strbuf b;
	    b << "stack variable '" << v.name () << "' shadows a parameter";
	    yyerror (b);
	  }
	  if (!t->add (v)) {
	    strbuf b;
	    b << "redefinition of stack variable: " << v.name () ;
 	    yyerror (b);
          }
	}
	;

declarator: pointer_opt direct_declarator
	{
	  if ($1.len () > 0) 
	    $2->set_pointer ($1);
  	  $$ = $2;
	}
	;

declarator_cpp: pointer_opt direct_declarator_cpp
	{
	  if ($1.len () > 0) 
	    $2->set_pointer ($1);
  	  $$ = $2;
	}
	;

cpp_initializer_opt: /* empty */ 
	{ 
	  $$ = New refcounted<initializer_t> (); 
	}
	| '(' passthrough ')'
	{
	  $$ = New refcounted<cpp_initializer_t> ($2); 
	}
	| '[' passthrough ']'
	{
	  $$ = New refcounted<array_initializer_t> ($2);
	}
	;

direct_declarator_cpp:	identifier 
	{
	  $$ = New refcounted<declarator_t> ($1);
	}
	;

/* 
 * use "typedef_name" instead of identifier for C++-style names
 *
 * simplified to not be recursive...
 */
direct_declarator: typedef_name
	{
	   $$ = New refcounted<declarator_t> ($1);
	}
	| typedef_name '(' parameter_type_list_opt ')'
	{
	   $$ = New refcounted<declarator_t> ($1, $3);
	}
	;


/* missing: first rule:
 *	storage_class_specifier declaration_specifiers_opt
 *
 * changed rule, to eliminate s/r conflicts
 *
 * Returns: <str> element, with the type of the variable (unparsed)
 */
declaration_specifiers: type_modifier_list type_specifier
	{
	   $1.add_lstr ($2);
	   $$ = $1;
	}
	;

/*
 * new rule to eliminate s/r conflicts
 */
type_modifier:  type_qualifier
	| T_SIGNED		{ $$ = type_qualifier_t ("signed"); }
	| T_UNSIGNED		{ $$ = type_qualifier_t ("unsigned"); }
	;

type_modifier_list: /* empty */ { $$ = type_qualifier_t (""); }
	| type_modifier_list type_modifier
	{
	  $$ = $1.concat ($2);
	}
	;
	

/* missing: struct and enum rules:
 *	| struct_or_union_specifier
 *	| enum_specifier
 */
type_specifier: T_VOID		{ $$ = "void" ; }
	| T_CHAR 		{ $$ = "char";  }
	| T_SHORT		{ $$ = "short"; }
	| T_INT			{ $$ = "int" ; }
	| T_LONG		{ $$ = "long" ; }
	| T_LONG_LONG		{ $$ = "long long"; }
	| T_FLOAT		{ $$ = "float"; }
	| T_DOUBLE		{ $$ = "double" ; }
	| typedef_name
	;

/*
 * hack for now -- not real C syntax
 */
type_qualifier:	T_CONST	{ $$ = type_qualifier_t ("const", CONST_FLAG); }
	| T_STRUCT	{ $$ = type_qualifier_t ("struct"); }
	| T_TYPENAME	{ $$ = type_qualifier_t ("typename"); }
	| T_HOLDVAR	{ $$ = type_qualifier_t (lstr (), HOLDVAR_FLAG); }
	;

type_qualifier_list: type_qualifier
	| type_qualifier_list type_qualifier
	{
	  $$ = $1.concat ($2);
	}
	;

type_qualifier_list_opt: /* empty */ { $$ = type_qualifier_t (""); }
	| type_qualifier_list        { $$ = $1; }
	;

/*
 * foo<int, char *>::bar_t::my_class<int> -> 
 *   foo<> bar_t my_class<>
 */
typedef_name:  typedef_name_single
	| typedef_name T_2COLON typedef_name_single
	{
	   CONCAT($1.lineno (), $1 << "::" << $3, $$);
	}
	;

typedef_name_single: identifier template_instantiation_opt
	{
          CONCAT($1.lineno (), $1 << $2, $$);
	}
	;

template_instantiation_opt: /* empty */ 	{ $$ = ""; }
	| template_instantiation	
	;

template_instantiation: '<' template_instantiation_list_opt '>'
	{
	  CONCAT($2.lineno (), "<" << $2 << ">", $$);
	}
	;

template_instantiation_list_opt: /* empty */   { $$ = "" ; }
	| template_instantiation_list
	;

template_instantiation_list: template_instantiation_arg
	| template_instantiation_list ',' template_instantiation_arg
	{
	  CONCAT($1.lineno (), $1 << " , " << $3, $$);
	}
	;

template_instantiation_arg: declaration_specifiers pointer_opt
	{
	  CONCAT($1.lineno (), $1.to_str () << " " << $2, $$);
	}
	;

pointer_opt: /* empty */	{ $$ = ""; }
	| pointer
	;

pointer_or_ref:	'*'		{ $$ = "*"; }
	| '&'			{ $$ = "&"; }
	;

pointer: pointer_or_ref		{ $$ = $1; }
	| pointer_or_ref type_qualifier_list_opt pointer
	{
	  CONCAT($2.lineno (), " " << $1 << " " << $2.to_str () << $3, $$);
	}
	;

%%

