%define shortversion 0.8
Summary: SFSlite -- SFS development libraries 
Name: sfslite
Version: 0.8.11
Release: 1
Copyright: GPL 
Group: Applications/File
Source: http://dist.okws.org/dist/sfslite-%{version}.tar.gz 
URL: http://www.okws.org/doku.php?id=okws:sfslite
Packager: OKWS Developers <http://www.okws.org/>
BuildRoot: %{_tmppath}/%{name}-%{version}-buildroot
Requires: kernel >= 2.2.14, gmp >= 2.0
BuildRequires: gcc, gcc-c++
BuildConflicts: gcc = 2.96-85, gcc-c++ = 2.96-85, gcc = 2.96-98, gcc-c++ = 2.96-98,gcc = 2.96-108, gcc-c++ = 2.96-108, gcc = 2.96-110, gcc-c++ = 2.96-110

%description
This is a port of the SFS-Lite development libraries.  The SFS toolkit
was developed to support the SFS distributed file system (see
http://www.fs.net).  But because others use the toolkit for other
reasons, we're making SFS's libraries available as a separate,
lightweight package.  sfslite compiles much faster and can be
installed as different non-conflicting build modes (such as
sfslite-dbg or sfslite-noopt) so might be better for some applications
that need the SFS libraries but not SFS.

Maintained as port of the OKWS distribution by Maxwell Krohn.

%prep
%setup -q

%build
%configure --enable-all
make 

%install
rm -rf $RPM_BUILD_ROOT
make install-strip DESTDIR=$RPM_BUILD_ROOT

%clean
rm -rf $RPM_BUILD_ROOT

%files
%defattr(-,root,root)
%dir %{_libdir}/sfslite-%{shortversion}
%doc AUTHORS COPYING ChangeLog NEWS README STANDARDS TODO
%{_bindir}/*
%{_includedir}/sfs.h
%{_includedir}/sfslite-%{shortversion}
%{_libdir}/sfslite-%{shortversion}/*

%changelog
* Sun Dec 03 2006 Emil Sit <sit@mit.edu>
- Develop sfslite package from sfs-0.8pre.spec

* Mon Jan 03 2005 Emil Sit <sit@mit.edu>
- Build under FC2.

* Tue Dec 03 2002 Kevin Fu <fubob@mit.edu>
- Merge the sfs-devel package into the core sfs package.  Move some 
  misplaced files between sfs and sfs-servers packages.  Clean up the
  upgrade process.  Make /sfs a %%ghost directory for easier removal.

* Mon Dec 02 2002 Kevin Fu <fubob@mit.edu>
- Release of SFS 0.7

* Sun Dec 01 2002 Kevin Fu <fubob@mit.edu>
- Add conditional magic to avoid compiling with gcc 2.96

* Sat Nov 30 2002 Kevin Fu <fubob@mit.edu>
- Created new spec file based on sfs.spec and 0.7pre12 build
