/* $Id: hashtab.h 435 2004-06-02 15:46:36Z max $ */

/*
 *
 * Copyright (C) 1998, 1999 David Mazieres (dm@uun.org)
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


#ifndef _HASHTAB_H_
#define _HASHTAB_H_ 1

#include <stddef.h>

struct hashtab_entry {
  void *hte_next;
  void **hte_prev;
  u_int hte_hval;
};

struct _hashtab {
  u_int ht_buckets;
  u_int ht_entries;
  void **ht_tab;
};

#define HASHSEED 5381
static inline u_int
hash (u_int seed, const u_char *key, int len)
{
  const u_char *end;

  for (end = key + len; key < end; key++)
    seed = ((seed << 5) + seed) ^ *key;
  return seed;
}

extern void _hashtab_grow (struct _hashtab *, u_int);
extern const u_int exp2primes[];

#define hashtab_decl(name, type, field)					\
struct name {								\
  struct _hashtab ht;							\
};									\
static inline void							\
name ## _init (struct name *_htp)					\
{									\
  struct _hashtab *htp = &_htp->ht;					\
  htp->ht_entries = 0;							\
  htp->ht_buckets = 0;							\
  htp->ht_tab = NULL;							\
}									\
static inline void							\
name ## _clear (struct name *_htp)					\
{									\
  struct _hashtab *htp = &_htp->ht;					\
  xfree (htp->ht_tab);							\
}									\
static inline void							\
name ## _insert (struct name *_htp, struct type *elm, u_int hval)	\
{									\
  struct _hashtab *htp = &_htp->ht;					\
  u_int bn;								\
  struct type *np;							\
									\
  if (++htp->ht_entries >= htp->ht_buckets)				\
    _hashtab_grow (htp, offsetof (struct type, field));			\
  elm->field.hte_hval = hval;						\
									\
  bn = hval % htp->ht_buckets;						\
  if ((np = htp->ht_tab[bn]))						\
    np->field.hte_prev = &elm->field.hte_next;				\
  elm->field.hte_next = np;						\
  elm->field.hte_prev = &htp->ht_tab[bn];				\
  htp->ht_tab[bn] = elm;						\
}									\
static inline void							\
name ## _delete (struct name *_htp, struct type *elm)			\
{									\
  struct _hashtab *htp = &_htp->ht;					\
  htp->ht_entries--;							\
  if (elm->field.hte_next)						\
    ((struct type *) elm->field.hte_next)->field.hte_prev		\
      = elm->field.hte_prev;						\
  *elm->field.hte_prev = elm->field.hte_next;				\
}									\
static inline struct type *						\
name ## _chain (struct name *_htp, u_int hval)				\
{									\
  struct _hashtab *htp = &_htp->ht;					\
  return htp->ht_buckets ? htp->ht_tab[hval % htp->ht_buckets] : 0;	\
}									\
static inline struct type *						\
name ## _next (struct type *elm)					\
{									\
  return elm->field.hte_next;						\
}									\
static inline void							\
name ## _traverse (struct name *_htp,					\
		   void (*fn) (void *, struct type *),			\
		   void *arg)						\
{									\
  struct _hashtab *htp = &_htp->ht;					\
  struct type *e, *ne;							\
  u_int i;								\
  for (i = 0; i < htp->ht_buckets; i++)					\
    for (e = htp->ht_tab[i]; e; e = ne) {				\
      ne = e->field.hte_next;						\
      fn (arg, e);							\
    }									\
}									\
static inline u_int							\
name ## _size (struct name *_htp)					\
{									\
  struct _hashtab *htp = &_htp->ht;					\
  return htp->ht_entries;						\
}									\
struct name

static inline u_int
hash_string (const char *_key)
{
  const u_char *key = (const u_char *) _key;
  u_int seed = 5381;
  while (*key)
    seed = ((seed << 5) + seed) ^ *key++;
  return seed;
}


#endif /* _HASHTAB_H_ */
