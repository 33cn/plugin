
#include "ihash.h"
#include "qhash.h"
#include "async.h"

struct foo_t {

  foo_t (u_int k, u_int xx, const str &yy, bool zz)
    : key (k), x (xx), y (yy), z (zz) {}

  // can be an int or anything hashable
  u_int key;  

  // can now but arbitrary data in this struct
  u_int x;
  str y;
  bool z;

  // if you want to put it in an ihash, you need to have this
  // type of element
  ihash_entry<foo_t> link;
};


/* 
 * ignore these functions; they are just for output
 */
#define PRINT_BOOL(v) \
warn << "variable " << #v << " has the value: "  << (v ? "T" : "F") << "\n";
#define PRINT_INTP(v)                 \
warn ("int *" #v " = %p\n", v);       \
if (v) warn ("*" #v " = %d\n", *v);


int
main (int argc, char *argv[])
{
  // bhash is "binary hash;" it's helpful for building sets and
  // then testing membership;
  bhash<str> bh;
  bh.insert ("fish");
  bh.insert ("likes to");
  bh.insert ("complain about");
  bh.insert ("sfs. alot");

  bool b1 = bh["fish"];  // b1 is true
  bool b2 = bh["max"];   // b2 is false

  PRINT_BOOL(b1);
  PRINT_BOOL(b2);

  // qhash is short for "quick hash"
  // that is, no a priori struct is needed; qhash rolls one on
  // the fly to associate the given keys with the given values
  qhash<str,int> qh;
  qh.insert ("washington", 1776);
  qh.insert ("mckinley", 1900);
  qh.insert ("lincoln", 1860);
  qh.insert ("the jesus (tm)", 0);
  
  // to get values out of qhash, you need to traffic in pointers
  // to values that you expect out.
  int *wash = qh["washington"];
  int *jesus = qh["the jesus (tm)"];
  int *bush = qh["bush"];

  // "wash" will point to an integer whose value is 1776
  PRINT_INTP (wash);   

  // "jesus" will point to an integer whose value is 0
  PRINT_INTP (jesus);

  // "bush" was not found in the table; it will be zero-valued pointer
  PRINT_INTP (bush);

  // ihash is short for "intrusive hash"
  // if two values are hashed to the same bucket, the table needs a way
  // to store multiple <key,value> pairs under the same bucket.  
  // the obvious way to do this is with a linked-list.  by putting
  // an "ihash_entry" field in a structure, you're declaring that it
  // can be linked-listified, for insertion into such a hash table.
  // an example is foo_t, declared above;

  // first argument: the type of the key being used
  // second argument: the type of struct being stored
  // thrid argument: the position of the key in the struct
  // fourth argument: the posititon of the hash link in the struct;
  ihash<u_int, foo_t, &foo_t::key, &foo_t::link> ih;

  foo_t *f1 = New foo_t (10, 100, "square of 10", true);
  foo_t *f2 = New foo_t (314, 15927, "pieces of pi", false);
 
  // note, inserting into an ihash takes only 1 argument, since
  // the key and the values are stored in the struct, and the
  // templated class definition tells the compiler where to find them
  ih.insert (f1);
  ih.insert (f2);

  foo_t *nada = ih[40];
  // will output 0x0
  warn ("output=%p\n", nada);      

  foo_t *found_one = ih[10];
  // should output "square of 10"
  warn ("found_one->y = %s\n", found_one->y.cstr ()) ;

}
