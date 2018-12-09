
#include "sp_gc.h"

namespace sp {
namespace gc {

  //=======================================================================

  static mgr_t<> *_g_mgr;

  mgr_t<> *
  meta_mgr_t<>::get ()
  {
    if (!_g_mgr) {
      _g_mgr = New std_mgr_t<> (std_cfg_t ());
    }
    return _g_mgr;
  }

  void meta_mgr_t<>::set (mgr_t<> *m) { _g_mgr = m; }

  //=======================================================================

#ifdef CGC_DEBUG
  bool debug_mem = true;
#else /* CGC_DEBUG */
  bool debug_mem;
#endif /* CGC_DEBUG */

  int debug_warnings;

  void mark_deallocated (void *p, size_t sz)
  {
    if (debug_mem) {
      if (debug_warnings)
	warn ("mark deallocated: %p to %p\n", p, (char *)p + sz);
      memset (p, 0xdf, sz);
    }
  }

  void mark_unitialized (void *p, size_t sz)
  {
    if (debug_mem)
      memset (p, 0xca, sz);
  }

  //=======================================================================
  
  size_t smallobj_sizer_t::_sizes[] =  { 4, 8, 12, 16,
					 24, 32, 40, 48, 56, 64,
					 80, 96, 112, 128, 
					 160, 192, 224, 256,
					 320, 384, 448, 512,
					 640, 768, 896, 1024 };
  
  //-----------------------------------------------------------------------
  
  smallobj_sizer_t::smallobj_sizer_t ()
    : _n_sizes (sizeof (_sizes) / sizeof (_sizes[0])) {}
  
  //-----------------------------------------------------------------------
  
  size_t
  smallobj_sizer_t::ind2size (int sz) const
  {
    if (sz < 0) return 0;
    assert (sz < int (_n_sizes));
    return _sizes[sz];
  }
  
  //-----------------------------------------------------------------------

  size_t
  smallobj_sizer_t::find (size_t sz, int *ip) const
  {
    // Binary search the sizes vector (above)
    int lim = _n_sizes;

    int l, m, h;
    h = lim - 1;
    l = 0;
    while ( l <= h ) {
      m = (l + h)/2;
      if (_sizes[m] > sz) { h = m - 1; }
      else if (_sizes[m] < sz) { l = m + 1; }
      else { l = m; break; }
    }

    if (l < lim && _sizes[l] < sz) l++;

    size_t ret = 0;
    if (l < lim) ret = _sizes[l];
    else         l = -1;

    if (ip)
      *ip = l;

    return ret;
  }

  //=======================================================================
  
  static size_t pagesz;

  size_t get_pagesz ()
  {
    if (!pagesz) {
      pagesz = sysconf (_SC_PAGE_SIZE);
    }
    return pagesz;
  }
  
  //=======================================================================
  
  void *
  cgc_mmap (size_t sz)
  {
    void *v = mmap (NULL, sz, PROT_READ | PROT_WRITE, 
		    MAP_PRIVATE | MAP_ANON , -1, 0);

    mark_unitialized (v, sz);

    if (!v) 
      panic ("mmap failed: %m\n");
    return v;
  }

  //=======================================================================

  size_t
  align (size_t in, size_t a)
  {
    a--;
    return (in + a) & ~a;
  }

  //-----------------------------------------------------------------------

  size_t
  boa_obj_align (size_t sz)
  {
    return align (sz, sizeof (void *));
  }

  //=======================================================================

};
};
