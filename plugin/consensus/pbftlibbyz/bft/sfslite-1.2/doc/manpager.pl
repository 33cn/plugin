#!/usr/local/bin/perl
#
# manpager.pl
#
#     A script used to process sfs.texi and to output a series of manpages
#     formatted for *roff.  Takes as arguments the input file (in all
#     seen cases, sfs.texi) and an output directory.  Will dump all
#     of the manpages in the output directory.
#
#     For the sfs.texi file to be successfully parsed by this program,
#     added structure is needed in the form of texi comments.  For
#     instance, here is a typical man page input:
#
#		@c @mp
#		@c @mp @command{foo}{foo processor}{1}
#		@c @mp @synopsis
#		@c @mpp foo [-i] @var{bar} 
#		@c @mp @end synopsis
#
#		@c @mp @description
#		Foo is a program used for the processing of foos,
#		the validation of the processing of foos, and the
#		generation of auto-validators thereof.
#		@c @mp @end description
#
#		@c @mp @options
#		@table @option
#		@item -I process, validate and auto-generate foo-validation
#		routines
#		@end @table
#		@c @mp @end @options
#		@c @mp @end @command
#		@c @mp
#
#     In this example, we've taken a .texi files, and added commands for
#     manpager.pl in the form of texi commands -- all lines of the form
#     "@c @mp ..." or "@c @mpp"
#
#     This program works by convering texi files into POD files, and then
#     by running pod2man translation.  Note that some functionality of 
#     the full *roff language is lost, notably, formatting in indented
#     environments. However, we've done our best to work around those 
#     problems.
#
#     Note that this program needs certain environment variables
#     to be given; for sfs.texi, VERSION will be needed.
#
#
#-----------------------------------------------------------------------
# $Id: manpager.pl 2 2003-09-24 14:35:33Z max $
#-----------------------------------------------------------------------
#

use strict;
use vars qw [ $pod2man_mod $pod2man_bin ];
if ($ENV{'POD_MAN'} and $ENV{'POD_MAN'} eq "yes") {
    require Pod::Man;
    $pod2man_mod = 1;
} elsif ($ENV{'POD2MAN'}) {
    $pod2man_bin = $ENV{'POD2MAN'};
} else {
    die "Could not find Pod::Man or pod2man\n";
}

use IO::File;
use Data::Dumper;
use vars qw [ $infile %TEXI_FORMATS %STYPES %TEXI_CHARS @TEXI_IGNORES_ARR
              %TEXI_IGNORES @STYPES_ARR @NATIVE_ENVS $NES @SEEALSO 
	      %TEXI_OPEN_ENV @TEXI_IGNORES_WL_ARR %TEXI_IGNORES_WL %GLOBALS
	      @GLOBAL_VARS $GVSWITCH @MANPAGE_SECTIONS 
	      $subsec_g %OUTPUT_ENV %SEEALSO ];


%TEXI_FORMATS = qw (  command    B
		      file       F
		      option     B
		      var        I
		      samp       B 
		      env        B  
		      uref       I
		      ref        I
		      emph	 I
		      strong     B
		      code       B
		      dfn	 B
		      );
$TEXI_FORMATS{xref} = "See I";

%TEXI_CHARS = qw (    dots       ...
		      dslash     /
		      );

%TEXI_OPEN_ENV = qw ( item       item 
		      itemx      item );

@TEXI_IGNORES_ARR = qw ( anchor table noindent pxref );
%TEXI_IGNORES = map { $_ => 1 } @TEXI_IGNORES_ARR;

# Ignore Whole-Line (WL)
@TEXI_IGNORES_WL_ARR = qw ( cindex );
%TEXI_IGNORES_WL = map { $_ => 1 } @TEXI_IGNORES_WL_ARR;

# Section Types
@STYPES_ARR = qw ( options synopsis command ignore description example
		   examples caveats
		   indent table conffile display itemize bugs files );
%STYPES = map { $_ => 1 } @STYPES_ARR;

@NATIVE_ENVS = qw ( example table display itemize ignore );
$NES = join("|", @NATIVE_ENVS);  # native env switch

@GLOBAL_VARS = qw ( VERSION SYSCONFDIR PKGDATADIR );
$GVSWITCH = join("|", @GLOBAL_VARS);

@MANPAGE_SECTIONS = qw ( synopsis description options files environment
			 diagostics examples caveats history see_also
			 bugs copying authors);

		      
		      

sub read_infile {
    my ($fh) = @_;
    my $state = 0;
    my $lineno = 0;
    my $lines = [];
    while (<$fh>) {
	$lineno++;
	# Begin reading on "@c @mp"
	if ($state == 0 && /^\@c\s+\@mp\s*$/) {
	    $state = 2;
	# end reading on "@c @mp @end"
	} elsif ($state == 1 && /^\@c\s+\@mp\s+\@end\s*$/) {
	    $state = 0;
	}
	if ($state == 1) {
	    my $pushit = 0;


	    # Only Take Commented lines that are commands
	    if ( /^\@(c|comment)\s+/ ) {
		if ( $_ =~ s/^\@c\s+\@mpp// or $_ =~ s/^\@c\s+\@mp/\@mp/  ) {
		    $pushit = 1;
		}
	    } else {
		$pushit = 1;
	    }

	    push @$lines, [ $_, $lineno ]  if $pushit;
	}
	$state = 1 if $state == 2 ;
    }
    return $lines;
}

sub parse_clump {
    my ($lines) = @_;
    while (parse_globals($lines)) {}
    my $sec = parse_section($lines);
    return $sec if $sec;
    return parse_text($lines);
}

sub parse_globals {
    my ($lines) = @_;
    return 0 unless $lines and $lines->[0] and $lines->[0]->[0];
    my $line = $lines->[0]->[0];
    if ($line =~ /^\@mp\s+global\s+($GVSWITCH)\s+(\S*)/ ) {
	$GLOBALS{$1} = $2;
	shift @$lines;
	return 1;
    }
    return 0;
}

sub parse_text {
    my ($lines) = @_;
    my $olines;
    my $line;
    while ($line = shift @$lines) {
    	# stop reading text on beginning or end of @mp environment
	if ($line->[0] =~ /^\@mp\s+\@/ or
	    $line->[0] =~ /^\@(end\s+)?($NES)/ ) {
	    unshift @$lines, $line;
	    last;
	}
	push @$olines, $line;
    }
    return { "type" => "TEXT" ,
             "data" => $olines };
}

sub parse {
    my ($lines) = @_;
    my $out = [];
    while ($lines and $#$lines >= 0) {
	my $r = parse_clump($lines);
	last unless $r;
	push @$out, $r;
    }
    return $out;
}

sub parse_section {
    my ($lines) = @_;
    my $pair = shift @$lines;
    return undef unless $pair and $pair->[0];
    my $line = $pair->[0];
    my $start = $pair->[1];
    my $stype;

    my $native_env = 0;
    if ($line =~ /^\@mp\s+\@\s/ ) {
	die "$infile:$start: Illegal empty command tag found\n";
    } elsif ($line =~ s/^\@mp\s+\@(\w+)// ) {
	$stype = $1;
    } elsif ($line =~ s/^\@($NES)\b// ) {
	$stype = $1;
	$native_env = 1;
    } else {
	unshift @$lines, $pair;
	return undef;
    }

    my $lineno = $pair->[1];

    if ($stype eq "end") {
	die "$infile:$lineno: Unbalanced \@end tag found\n";
    }

    unless ($STYPES{$stype}) {
	unshift @$lines, $pair;
	die "$infile:$lineno: Unrecognized section type: $stype\n";
    }

    $line =~ s/^\s*\{// ;
    $line =~ s/\}\s*$//;
    my @args = split /\}\s*\{/, $line;

    my @clumps = ();

    while (1) {

	my $clump = parse_clump($lines);

	my $line = shift @$lines;
	if (!$line or !$line->[0]) {
		die qq#$infile:EOF: Hit EOF before end of <$stype> #,
	            qq# environment started on line $start#;
	}
	push @clumps, $clump;
	my $prefix = $native_env ? "" : '@mp\s+';
	if ($line->[0] =~ /^$prefix\@end\s+$stype/ ) {
	    last;
	}
	unshift @$lines, $line;
    }

    my $section = { "type" => $stype,
                    "args" => [ @args ], 
		    "data" => [ @clumps ] , 
		    "lineno" => $lineno  };
    return $section;
}

sub output {
    my ($doc, $outdir) = @_;
    my $manpages = 0;
    foreach my $manpage (@$doc) {
	if (ref($manpage) and ref($manpage) eq "HASH" and 
	    $manpage->{type} eq "command") {
	    output_command($manpage, $outdir);
	    $manpages ++;
	}
    }
}

sub output_command {
    my ($mp, $outdir) = @_;
    die "$infile:$mp->{lineno}: 3 Arguments expected to \@mp \@command"
	unless $#{$mp->{args}} == 2;
    my $name = $mp->{args}->[0];
    my $sdesc = $mp->{args}->[1];
    my $msec = $mp->{args}->[2];
    my $outfnp = "$outdir/$name.$msec.pod";
    my $outfn = "$outdir/$name.$msec";
    my ($dsec,$min,$hour,$mday,$mon,$year,$wday,$yday,$isdst) =
              localtime(time);
    my @months = qw [ Jan Feb Mar Apr May Jun Jul Aug Sep Oct Nov Dec ];
    my $date = $months[$mon] . " " . ($year + 1900);

    foreach ($outfnp, $outfn) { unlink $_ if -f $_ and !(-w $_); }

    my $outfh = new IO::File(">$outfnp");
    die "Cannot open $outfn for writing" unless $outfh;
    print $outfh "=head1 NAME\n\n$name - $sdesc\n\n";

    my %dispatch = ( "see_also"   => \&output_see_also,
		     "authors"    => \&output_authors,
		     "files"      => \&output_files );

    foreach my $ss (@MANPAGE_SECTIONS) {
	my $subsec = $subsec_g = $mp->{subsec}->{$ss};
	my $func = $dispatch{$ss};
	if ($func and !$subsec) {
	    &$func($outfh, $name, $msec); 
	} else {
	    output_subsection($subsec, $outfh) if $subsec;
	}
    }
    $outfh->close();

    # if we have another global variable this should become a function;
    # the next thing to think about is scoping.
    my $v = get_value("VERSION", 0, "0.0");


    if ($pod2man_mod) {
	my $parser = Pod::Man->new( release => "SFS $v", section => $msec,
                                center => "SFS $v", name => uc ($name));
	$parser->parse_from_file($outfnp, $outfn);
    } elsif ($pod2man_bin) {
	system (qq!$pod2man_bin --release="SFS $v" --section="$msec" !.
		qq!--center="SFS $v" $outfnp > $outfn!);
    } else {
	die "Cannot find a valid pod2man\n";
    }
    unlink($outfnp);


}

sub output_subsection {
    my ($ssec_arr, $outfh) = @_;
    return unless $ssec_arr and $#$ssec_arr >= 0;
    print $outfh "=head1 ", uc($ssec_arr->[0]->{type}), "\n\n";
    foreach my $ssec (@$ssec_arr) {
	output_normal($ssec, $outfh);
        # print a newline ?
    }
    print $outfh "\n";
}

sub output_normal {
    my ($sec, $outfh) = @_;
    my $over = 0;
    my %oenv_copy;
    if ($sec->{type} eq "TEXT") {
	output_text($sec->{data}, $outfh);
    } elsif ($sec->{type} eq "ignore") {
	return;
    } elsif ($sec->{data}) {
	%oenv_copy = %OUTPUT_ENV;
	if ($sec->{type} eq "indent" or $sec->{type} eq "table" or 
	    $sec->{type} eq "itemize" ) {
	    print $outfh "\n\n=over 4\n\n";
	    $over = 1;
	}
	if ($sec->{type} eq "itemize" ) {
	    $OUTPUT_ENV{itemize} = 1;
	}
	    
	# need to do special handling of example.  also @display mode?
	if ($sec->{type} eq "example" and !($subsec_g eq "synopsis")) {
	    $OUTPUT_ENV{example} = 1;
	}
	if ($sec->{type} eq "display" ) {
	    $OUTPUT_ENV{display} = 1;
	}
	foreach my $ssec (@{$sec->{data}}) {
	    output_normal($ssec, $outfh);
        }
	if ($over) {
	    print $outfh "\n\n=back\n\n";
	}
	%OUTPUT_ENV = %oenv_copy;
    }
}

#
# order: from systerm environment variables
#        then, from global variables defined in file
#        then, default value
#
sub get_value {
    my ($var, $line, $def) = @_;
    my $val = $ENV{$var};
    $val = $GLOBALS{$var} unless defined($val);
    $val = $def if (defined($def) and !defined($val));
    warn "$infile:$line: Variable $var undefined\n" unless $val;
    $val = "" unless defined($val);
    return $val;
}

sub output_text {
    my ($data, $outfh) = @_;

    my $fswitch = join("|", keys %TEXI_FORMATS);
    my $cswitch = join ("|", keys %TEXI_CHARS);
    my $iswitch = join ("|", keys %TEXI_IGNORES);
    my $iwlswitch = join("|", keys %TEXI_IGNORES_WL);
    my $oeswitch = join ("|", keys %TEXI_OPEN_ENV);

    # first do commands that need to keep track of newlines
    my $txt = "";
    my $startpar = -1;
    my $notflushlines = 0;
    my $example = $OUTPUT_ENV{example};
    my $itemize = $OUTPUT_ENV{itemize};
    my $display = $OUTPUT_ENV{display};
    my $fmt = $display ? 0 : 1;
    foreach my $pair (@$data) {
	my $line = $pair->[0];
	my $lineno = $pair->[1];
	$startpar = $lineno if $startpar < 0;
	$line =~ s/\@($iwlswitch)\s+.*$//g;
	$line =~ s/\@value\{(\w*)\}/get_value($1,$lineno, "")/eg;
	$line =~ s{\@($oeswitch)\s+(.*)}{\n\n=$TEXI_OPEN_ENV{$1} $2\n\n}g 
	    unless $itemize;
	$line =~ s/^\@end.*//g;

	# example environments that need indentation cannot have
	# formatting.  this is a limitation of POD
	if ($example) { $notflushlines++ if $line =~ /^\s/ ; }
	else { $line =~ s/^[ \t]+//; }
	if ($display) { $txt .= "  "; }
	$txt .= $line;
    }
    $fmt = 0 if $notflushlines > 0;

    #
    # itemized lists should be bulleted
    if ($itemize) { $txt =~ s/\@item/\n\n=item *\n\n/gs; }

    #
    # get rid of constructs like " SFS, @ref{SFS}."
    #
    $txt =~ s/,\s+\@ref{\w*}\././sgo;

    # now treat all of the text as one line
    $txt =~ s/(\S)\-{2}/$1-/gso;
    $txt =~ s/(\S)\-{3}/$1--/gso;
    $txt =~ s/\@($iswitch)(\s)/$2/gso;
    $txt =~ s/\@($iswitch)\{(([^\}@]|@.)*)\}//gso;
    $txt =~ s/\@($cswitch)(\{\})?/$TEXI_CHARS{$1}/gso;
    $txt = unnest($txt);
    die "$infile:$startpar: Unbalanced nested expression in paragraph" 
	unless defined($txt);
    my $c = 0;
    $txt =~ s/\@uref\{([^,\}]+),([^\}]+)\}/"$2 (I<$1>)"/egso;
    $txt =~ s/\@($fswitch)\{(([^\}@]|@.)*)\}/tfmt($1,$2,$fmt,\$c)/egso;
    if ($c == 0 and $example) {
	$txt =~ s/^(.)/  $1/ ;
	$txt =~ s/\n/\n  /sg;
    }

    # strip away all white space if possible
    unless ($example) {
	$txt =~ s/^[ \t]*// ;
	$txt =~ s/\n[ \t]*/\n/sg;
    }

    # unescape "@@" sequences
    $txt =~ s/@([@\{\}])/$1/gso;
    print $outfh $txt if $txt;

}

sub tfmt {
    my ($cmd, $arg, $fmt, $c) = @_;
    $$c++;
    return $arg unless $fmt;
    return $TEXI_FORMATS{$cmd} . "<$arg>";
}

sub unnest_parse {
    my ($line, $cmd, $err) = @_;

    my $ret;
    $ret->{cmd} = $cmd if defined($cmd);

    my $cflag = 1;
    while ($$line and length($$line) >= 0 and $cflag ) {
	$cflag = 0;
	if ($$line =~ s/^(([^}@]|@.)*?)@(\w+)\{//s ) {
	    push @{$ret->{data}}, $1 if $1 and length($1) > 0;
	    my $sret = unnest_parse($line, $3, $err);
	    push @{$ret->{data}}, $sret if $sret;
	    $$err = 1 unless $$line =~ s/^\}// ;
	    $cflag = 1;
	# Must fix this!  - MK 06/17/02 - Gobble all non-} or @} combinations
	} elsif ($$line =~ s/^([^@}]|(@.))+@?//s ) {
	    push @{$ret->{data}}, $&;
	}
    }
    return ($#{$ret->{data}} >= 0 ? $ret : undef );
}

sub unnest_output {
    my ($expr) = @_;
    my $ret = "";
    return $ret unless defined($expr);
    foreach my $d (@{$expr->{data}}) {
	if (ref($d)) { $ret .= unnest_output($d); }
	elsif (!defined($expr->{cmd})) { $ret .=  $d ; }
	else { $ret .= '@' . $expr->{cmd} . '{' . $d . '}' ; }
    }
    return $ret;
}

#
# unnests nested expressions
#
# input:  a b @c{@d{e}} f @h{ i j @k{l @m{n @o{p}} q} r}
# output: a b @d{e} f @h{ i j }@k{l }@m{n }@o{p}@k{ q}@h{ r}
#
# also watch out "@}" type escape routines:
# input:  a b @c{@}@@@d{e}}
# output: a b @c{@}@@}@d{e}
#
sub unnest {
    my ($line) = @_;
    my $err = 0;
    my $expr = unnest_parse(\$line, undef, \$err);
    return undef if $err;
    return unnest_output($expr);
}

sub output_authors {
    my ($outfh) = @_;
    print $outfh "\n\n=head1 AUTHOR\n\nsfsdev\@redlab.lcs.mit.edu\n\n",
}

sub sacmp {
    if ($a->[1] == $b->[1]) { return $a->[0] cmp $b->[0] ; }
    else { return $a->[1] <=> $b->[1] ; }
}

sub output_see_also {
    my ($outfh, $cmd, $sec) = @_;
    print $outfh "\n\n=head1 SEE ALSO\n\n";
    my $first = 1;
    foreach my $seealso (@SEEALSO) {
	next if $seealso->[0] eq $cmd and $seealso->[1] eq $sec;
	print $outfh ", " unless $first;
	$first = 0;
	print $outfh $seealso->[0], "(", $seealso->[1], ")";
    }
    print $outfh <<EOF;


The full documentation for B<SFS> is maintained as a Texinfo
manual.  If the B<info> and B<SFS> programs are properly installed
at your site, the command B<info SFS>
should give you access to the complete manual.

For updates, documentation, and software distribution, please
see the B<SFS> website at I<http://www.fs.net>.

EOF

}

sub output_files {
	my ($outfh, $cmd, $sec) = @_;
	my $conf = ($sec == 5) ? $cmd : $cmd . "_config";
	my $sdesc;
	return unless $sdesc = $SEEALSO{5}->{$conf} ;
	my $etcdir = get_value("ETCDIR", 0, "/etc/sfs");
	my $defdir = get_value("PKGDATADIR", 0, "/usr/local/share/sfs");
	print $outfh "\n\n=head1 FILES\n\n=over 4\n\n";
	foreach my $d ($etcdir, $defdir) {
	    print $outfh "=item F<$d/$conf>\n\n";
	}
	print $outfh "$sdesc\n\n=back\n\n",
	             "(Files in F<$etcdir> supersede default versions",
                     " in F<$defdir>.)\n\n";
}

#
# associates a man-page command with its constituent subsections
#
sub cleanup {
    my ($doc) = @_;
    my ($secs, $sstype);
    foreach my $sec (@$doc) {
	if ($sec->{type} eq "conffile") {
	    $sec->{type} = "command";
	    $sec->{args}->[2] = 5;
	}
	next unless $sec and ref($sec) eq "HASH" and 
	    ($sec->{type} eq "command" or $sec->{type} eq "conf" );
	foreach my $subsec (@{$sec->{data}}) {
	    unless ($subsec->{type} eq "TEXT") {
		$sstype = $subsec->{type};
	        push @{$sec->{subsec}->{$sstype}}, $subsec;
            }
	}
	my $args = $sec->{args};
	push @$secs, $sec;
	push @SEEALSO, [ $args->[0], $args->[2] ];
	$SEEALSO{$args->[2]}->{$args->[0]} = $args->[1];
    }
    @SEEALSO = sort sacmp @SEEALSO;
    return $secs;
}

#
# do config test first
#
if ($#ARGV == 0 and $ARGV[0] eq "-t") { exit(0); }
sub usage {
    die "usage: $0 <infile.texi> <out-directory>\n";
}

usage() unless $#ARGV == 1;
$infile = $ARGV[0];
my $infh = new IO::File("<$ARGV[0]");
my $outdir = $ARGV[1];
die "Cannot open infile: $infile" unless $infile;
die "Cannot find destination directory: $outdir" unless -d $outdir;

my $lines = read_infile($infh);
$infh->close();

my $doc = parse($lines);
cleanup($doc);
output($doc, $outdir, undef);

exit(0);

