package executor

import "errors"

var (
	errEmptyName            = errors.New("errEmptyName")
	errDuplicateMember      = errors.New("errDuplicateMember")
	errDuplicateAdmin       = errors.New("errDuplicateAdmin")
	errInvalidVoteTime      = errors.New("errInvalidVoteTime")
	errInvalidVoteOption    = errors.New("errInvalidVoteOption")
	errVoteNotExist         = errors.New("errVoteNotExist")
	errGroupNotExist        = errors.New("errGroupNotExist")
	errStateDBGet           = errors.New("errStateDBGet")
	errInvalidVoteID        = errors.New("errInvalidVoteID")
	errInvalidGroupID       = errors.New("errInvalidGroupID")
	errInvalidOptionIndex   = errors.New("errInvalidOptionIndex")
	errAddrAlreadyVoted     = errors.New("errAddrAlreadyVoted")
	errVoteAlreadyFinished  = errors.New("errVoteAlreadyFinished")
	errVoteNotStarted       = errors.New("errVoteNotStarted")
	errVoteAlreadyClosed    = errors.New("errVoteAlreadyClosed")
	errAddrPermissionDenied = errors.New("errAddrPermissionDenied")
)
