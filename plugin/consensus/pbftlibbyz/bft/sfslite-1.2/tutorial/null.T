#include "tame.h"
#include "async.h"

tamed static void foo() {}

int
main (int argc, char *argv[])
{
  foo ();
}
