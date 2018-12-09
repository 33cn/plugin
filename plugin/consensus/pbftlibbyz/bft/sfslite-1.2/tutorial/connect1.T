// -*-c++-*-
/* $Id: test.T 2468 2007-01-02 22:07:27Z max $ */

#include "async.h"
#include "tame.h"
#include "qhash.h"
#include "tame_connectors.h"
#include "parseopt.h"

int count = 0;

tamed static void
__toconn (event<int>::ptr *out, int n, event<int>::ref in)
{
  tvars {
    rendezvous_t<outcome_t> rv (__FILE__, __LINE__);
    timecb_t *tcb (NULL);
    int serial;
    outcome_t outcome;
  }
  serial = count++;

  warn << "connector(" << serial << "): timeout in " << n << " seconds..\n";

  tcb = delaycb (n, 0, mkevent (rv, OUTCOME_TIMEDOUT));
  in->set_cancel_notifier (mkevent (rv, OUTCOME_DISCARDED));
  *out = mkevent_rs (in->slot_set (), rv, OUTCOME_SUCC);
  
  twait (rv, outcome);

  warn << "connector(" << serial << "): notifed " << int (outcome) << "\n";

  if (outcome != OUTCOME_TIMEDOUT) {
    timecb_remove (tcb);
  }
  
  rv.cancel ();
  in->trigger_no_assign ();
}

static event<int>::ptr
toconn (int n, event<int>::ref e)
{
  event<int>::ptr out;
  __toconn (&out, n, e);
  return out;
}

static void usage ()
{
  warnx << "usage: " << progname << 
    " [-t<time>] [-l<lo>] [-h<hi>] [-R] [-n <nconn>]\n";
  exit (-1);
}

int myrand (int lo, int hi)
{
  return ((rand() % (hi -lo)) + lo);
}


event<int>::ref
mk_chain (event<int>::ref e, int n, int lo, int hi)
{
  int i;

  for (i = 0; i < n; i++) {
    e = toconn (myrand (lo, hi), e);
  }
  return e;
}

tamed static void
fn (int t, evi_t ev)
{
  twait { delaycb (t, 0, mkevent ()); }
  ev->trigger (10);
}

tamed static void
run_experiment (int t, int n, int lo, int hi, evv_t e)
{
  tvars {
    int res (-1);
  }
  warn << "Running expriment; operation completes in " << t << " secs\n";
  warn (  "   - with %d connectors in the range %d-%d\n", n, lo, hi);

  twait { fn (t, mk_chain (mkevent (res), n, lo, hi)); }
  warn << "Experiment finished: res=" << res << "\n";
  e->trigger ();
}


tamed static void 
main2 (int argc, char **argv)
{
  tvars {
    int t (5), n (4), lo (4), hi (10);
    bool r (false);
    int ch;
  }

  while ((ch = getopt (argc, argv, "t:n:l:h:r")) != -1) {
    switch (ch) {
    case 'r':
      r = true;
      break;
    case 'n':
      if (!convertint (optarg, &n))
	usage ();
      break;
    case 'l':
      if (!convertint (optarg, &lo))
	usage ();
      break;
    case 'h':
      if (!convertint (optarg, &hi))
	usage ();
      break;
    case 't':
      if (!convertint (optarg, &t))
	usage ();
      break;
    default:
      usage ();
      break;
    }
  }
  if (r) 
    srand (time (NULL) ^ getpid ());

  twait { run_experiment (t, n, lo, hi, mkevent ()); }
  warn << "Delaying 5 seconds...\n";
  twait { delaycb (5, 0, mkevent ()); }
  warn << "Exiting....\n";
  exit (0);
}

int main (int argc, char *argv[])
{
  setprogname (argv[0]);
  main2 (argc, argv);
  amain ();
}
