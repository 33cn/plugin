
#include "py_rpctypes.h"


#ifdef PYDEBUG

qhash<str,int> g_new_cnt;
qhash<str,int> g_del_cnt;

static void
memreport_trav (strbuf b, const str &k, int *np)
{
  int d = g_del_cnt[k] ? *(g_del_cnt[k]) : 0;
  int n = np ? *np : 0;
  b << k << "\t" << "new:" << n << "\t" << "del:" << d;
  if (n != d) {
    b << "\t**MEMORY LEAK=" << n - d ;
  }
  b << "\n";
}

void 
pydebug_memreport (const strbuf &b)
{
  b << "=============== New/Delete Memory Statistics ===================\n";
  g_new_cnt.traverse (wrap (memreport_trav, b));
  b << "============= End New/Delete Memory Statistics =================\n";
}


void 
pydebug_inc (const str &k, qhash<str,int> *hsh)
{

  int *i;
  if ((i = (*hsh)[k]))  (*i) ++ ;
  else hsh->insert (k, 1);
}

void 
pydebug_dec (const str &k, qhash<str,int> *hsh)
{
  int *i;
  if (!(i = (*hsh)[k])) {
    warn << "dec without inc: " << k << "\n";
  } else {
    (*i) -- ;
  }
}


#endif /* PYDEBUG */
