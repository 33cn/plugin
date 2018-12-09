#include <rpc/rpc.h>
#include "nfsd.h"

bool_t
xdr_nfsstat(XDR *xdrs, nfsstat *objp)
{
	if (!xdr_enum(xdrs, (enum_t *)objp)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_ftype(XDR *xdrs, ftype *objp)
{
	if (!xdr_enum(xdrs, (enum_t *)objp)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_fhandle(XDR *xdrs, fhandle *objp)
{
	if (!xdr_opaque(xdrs, objp->data, FHSIZE)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_nfstimeval(XDR *xdrs, nfstimeval *objp)
{
	if (!xdr_u_int(xdrs, &objp->seconds)) {
		return (FALSE);
	}
	if (!xdr_u_int(xdrs, &objp->useconds)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_fattr(XDR *xdrs, fattr *objp)
{
	if (!xdr_ftype(xdrs, &objp->type)) {
		return (FALSE);
	}
	if (!xdr_u_int(xdrs, &objp->mode)) {
		return (FALSE);
	}
	if (!xdr_u_int(xdrs, &objp->nlink)) {
		return (FALSE);
	}
	if (!xdr_u_int(xdrs, &objp->uid)) {
		return (FALSE);
	}
	if (!xdr_u_int(xdrs, &objp->gid)) {
		return (FALSE);
	}
	if (!xdr_u_int(xdrs, &objp->size)) {
		return (FALSE);
	}
	if (!xdr_u_int(xdrs, &objp->blocksize)) {
		return (FALSE);
	}
	if (!xdr_u_int(xdrs, &objp->rdev)) {
		return (FALSE);
	}
	if (!xdr_u_int(xdrs, &objp->blocks)) {
		return (FALSE);
	}
	if (!xdr_u_int(xdrs, &objp->fsid)) {
		return (FALSE);
	}
	if (!xdr_u_int(xdrs, &objp->fileid)) {
		return (FALSE);
	}
	if (!xdr_nfstimeval(xdrs, &objp->atime)) {
		return (FALSE);
	}
	if (!xdr_nfstimeval(xdrs, &objp->mtime)) {
		return (FALSE);
	}
	if (!xdr_nfstimeval(xdrs, &objp->ctime)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_sattr(XDR *xdrs, sattr *objp)
{
	if (!xdr_u_int(xdrs, &objp->mode)) {
		return (FALSE);
	}
	if (!xdr_u_int(xdrs, &objp->uid)) {
		return (FALSE);
	}
	if (!xdr_u_int(xdrs, &objp->gid)) {
		return (FALSE);
	}
	if (!xdr_u_int(xdrs, &objp->size)) {
		return (FALSE);
	}
	if (!xdr_nfstimeval(xdrs, &objp->atime)) {
		return (FALSE);
	}
	if (!xdr_nfstimeval(xdrs, &objp->mtime)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_filename(XDR *xdrs, filename *objp)
{
	if (!xdr_string(xdrs, objp, NFS_MAXNAMLEN)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_path(XDR *xdrs, path *objp)
{
	if (!xdr_string(xdrs, objp, NFS_MAXPATHLEN)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_attrstat(XDR *xdrs, attrstat *objp)
{
	if (!xdr_nfsstat(xdrs, &objp->status)) {
		return (FALSE);
	}
	switch (objp->status) {
	case NFS_OK:
		if (!xdr_fattr(xdrs, &objp->attributes)) {
			return (FALSE);
		}
		break;
	default:
		break;
	}
	return (TRUE);
}




bool_t
xdr_diropargs(XDR *xdrs, diropargs *objp)
{
	if (!xdr_fhandle(xdrs, &objp->dir)) {
		return (FALSE);
	}
	if (!xdr_filename(xdrs, &objp->name)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_diropokres(XDR *xdrs, diropokres *objp)
{
	if (!xdr_fhandle(xdrs, &objp->file)) {
		return (FALSE);
	}
	if (!xdr_fattr(xdrs, &objp->attributes)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_diropres(XDR *xdrs, diropres *objp)
{
	if (!xdr_nfsstat(xdrs, &objp->status)) {
		return (FALSE);
	}
	switch (objp->status) {
	case NFS_OK:
		if (!xdr_diropokres(xdrs, &objp->diropok)) {
			return (FALSE);
		}
		break;
	default:
		break;
	}
	return (TRUE);
}




bool_t
xdr_sattrargs(XDR *xdrs, sattrargs *objp)
{
	if (!xdr_fhandle(xdrs, &objp->file)) {
		return (FALSE);
	}
	if (!xdr_sattr(xdrs, &objp->attributes)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_readlinkres(XDR *xdrs, readlinkres *objp)
{
	if (!xdr_nfsstat(xdrs, &objp->status)) {
		return (FALSE);
	}
	switch (objp->status) {
	case NFS_OK:
		if (!xdr_path(xdrs, &objp->data)) {
			return (FALSE);
		}
		break;
	default:
		break;
	}
	return (TRUE);
}




bool_t
xdr_readargs(XDR *xdrs, readargs *objp)
{
	if (!xdr_fhandle(xdrs, &objp->file)) {
		return (FALSE);
	}
	if (!xdr_u_int(xdrs, &objp->offset)) {
		return (FALSE);
	}
	if (!xdr_u_int(xdrs, &objp->count)) {
		return (FALSE);
	}
	if (!xdr_u_int(xdrs, &objp->totalcount)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_readokres(XDR *xdrs, readokres *objp)
{
	if (!xdr_fattr(xdrs, &objp->attributes)) {
		return (FALSE);
	}
	if (!xdr_bytes(xdrs, (char **)&objp->data.data_val,
		       (u_int *)&objp->data.data_len, NFS_MAXDATA)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_readres(XDR *xdrs, readres *objp)
{
	if (!xdr_nfsstat(xdrs, &objp->status)) {
		return (FALSE);
	}
	switch (objp->status) {
	case NFS_OK:
		if (!xdr_readokres(xdrs, &objp->reply)) {
			return (FALSE);
		}
		break;
	default:
		break;
	}
	return (TRUE);
}




bool_t
xdr_writeargs(XDR *xdrs, writeargs *objp)
{
	if (!xdr_fhandle(xdrs, &objp->file)) {
		return (FALSE);
	}
	if (!xdr_u_int(xdrs, &objp->beginoffset)) {
		return (FALSE);
	}
	if (!xdr_u_int(xdrs, &objp->offset)) {
		return (FALSE);
	}
	if (!xdr_u_int(xdrs, &objp->totalcount)) {
		return (FALSE);
	}
	if (!xdr_bytes(xdrs, (char **)&objp->data.data_val,
		       (u_int *)&objp->data.data_len, NFS_MAXDATA)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_createargs(XDR *xdrs, createargs *objp)
{
	if (!xdr_diropargs(xdrs, &objp->where)) {
		return (FALSE);
	}
	if (!xdr_sattr(xdrs, &objp->attributes)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_renameargs(XDR *xdrs, renameargs *objp)
{
	if (!xdr_diropargs(xdrs, &objp->from)) {
		return (FALSE);
	}
	if (!xdr_diropargs(xdrs, &objp->to)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_linkargs(XDR *xdrs, linkargs *objp)
{
	if (!xdr_fhandle(xdrs, &objp->from)) {
		return (FALSE);
	}
	if (!xdr_diropargs(xdrs, &objp->to)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_symlinkargs(XDR *xdrs, symlinkargs *objp)
{
	if (!xdr_diropargs(xdrs, &objp->from)) {
		return (FALSE);
	}
	if (!xdr_path(xdrs, &objp->to)) {
		return (FALSE);
	}
	if (!xdr_sattr(xdrs, &objp->attributes)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_nfscookie(XDR *xdrs, nfscookie *objp)
{
	if (!xdr_opaque(xdrs, objp->data, COOKIESIZE)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_readdirargs(XDR *xdrs, readdirargs *objp)
{
	if (!xdr_fhandle(xdrs, &objp->dir)) {
		return (FALSE);
	}
	if (!xdr_nfscookie(xdrs, &objp->cookie)) {
		return (FALSE);
	}
	if (!xdr_u_int(xdrs, &objp->count)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_entry(XDR *xdrs, entry *objp)
{
	if (!xdr_u_int(xdrs, &objp->fileid)) {
		return (FALSE);
	}
	if (!xdr_filename(xdrs, &objp->name)) {
		return (FALSE);
	}
	if (!xdr_nfscookie(xdrs, &objp->cookie)) {
		return (FALSE);
	}
	if (!xdr_pointer(xdrs, (char **)&objp->nextentry, sizeof(entry),
			 (xdrproc_t) xdr_entry)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_dirlist(XDR *xdrs, dirlist *objp)
{
	if (!xdr_pointer(xdrs, (char **)&objp->entries, sizeof(entry),
			 (xdrproc_t) xdr_entry)) {
		return (FALSE);
	}
	if (!xdr_bool(xdrs, &objp->eof)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_readdirres(XDR *xdrs, readdirres *objp)
{
	if (!xdr_nfsstat(xdrs, &objp->status)) {
		return (FALSE);
	}
	switch (objp->status) {
	case NFS_OK:
		if (!xdr_dirlist(xdrs, &objp->readdirok)) {
			return (FALSE);
		}
		break;
	default:
		break;
	}
	return (TRUE);
}




bool_t
xdr_statfsokres(XDR *xdrs, statfsokres *objp)
{
	if (!xdr_u_int(xdrs, &objp->tsize)) {
		return (FALSE);
	}
	if (!xdr_u_int(xdrs, &objp->bsize)) {
		return (FALSE);
	}
	if (!xdr_u_int(xdrs, &objp->blocks)) {
		return (FALSE);
	}
	if (!xdr_u_int(xdrs, &objp->bfree)) {
		return (FALSE);
	}
	if (!xdr_u_int(xdrs, &objp->bavail)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_statfsres(XDR *xdrs, statfsres *objp)
{
	if (!xdr_nfsstat(xdrs, &objp->status)) {
		return (FALSE);
	}
	switch (objp->status) {
	case NFS_OK:
		if (!xdr_statfsokres(xdrs, &objp->info)) {
			return (FALSE);
		}
		break;
	default:
		break;
	}
	return (TRUE);
}




bool_t
xdr_fhstatus(XDR *xdrs, fhstatus *objp)
{
	if (!xdr_u_int(xdrs, &objp->status)) {
		return (FALSE);
	}
	switch (objp->status) {
	case 0:
		if (!xdr_fhandle(xdrs, &objp->directory)) {
			return (FALSE);
		}
		break;
	default:
		break;
	}
	return (TRUE);
}




bool_t
xdr_dirpath(XDR *xdrs, dirpath *objp)
{
	if (!xdr_string(xdrs, objp, MNTPATHLEN)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_name(XDR *xdrs, name *objp)
{
	if (!xdr_string(xdrs, objp, MNTNAMLEN)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_mountlist(XDR *xdrs, mountlist *objp)
{
	if (!xdr_pointer(xdrs, (char **)objp, sizeof(struct mountnode),
			 (xdrproc_t) xdr_mountnode)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_mountnode(XDR *xdrs, mountnode *objp)
{
	if (!xdr_name(xdrs, &objp->hostname)) {
		return (FALSE);
	}
	if (!xdr_dirpath(xdrs, &objp->directory)) {
		return (FALSE);
	}
	if (!xdr_mountlist(xdrs, &objp->nextentry)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_groups(XDR *xdrs, groups *objp)
{
	if (!xdr_pointer(xdrs, (char **)objp, sizeof(struct groupnode),
			 (xdrproc_t) xdr_groupnode)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_groupnode(XDR *xdrs, groupnode *objp)
{
	if (!xdr_name(xdrs, &objp->grname)) {
		return (FALSE);
	}
	if (!xdr_groups(xdrs, &objp->grnext)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_exportlist(XDR *xdrs, exportlist *objp)
{
	if (!xdr_pointer(xdrs, (char **)objp, sizeof(struct exportnode),
			 (xdrproc_t) xdr_exportnode)) {
		return (FALSE);
	}
	return (TRUE);
}




bool_t
xdr_exportnode(XDR *xdrs, exportnode *objp)
{
	if (!xdr_dirpath(xdrs, &objp->filesys)) {
		return (FALSE);
	}
	if (!xdr_groups(xdrs, &objp->groups)) {
		return (FALSE);
	}
	if (!xdr_exportlist(xdrs, &objp->next)) {
		return (FALSE);
	}
	return (TRUE);
}


#ifdef KERBEROS

bool_t
xdr_krbtkt(XDR *xdrs, krbtkt *objp)
{
	if (!xdr_opaque(xdrs, objp->data, KRBTKTSIZE)) {
		return (FALSE);
	}
	return (TRUE);
}

#endif /* KERBEROS */
