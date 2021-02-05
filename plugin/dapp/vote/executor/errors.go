package executor

import "errors"

var (
	errEmptyName            = errors.New("ErrEmptyName")
	errInvalidMemberWeights = errors.New("errInvalidMemberWeights")
	errNilMember            = errors.New("errNilMember")
	errDuplicateMember      = errors.New("errDuplicateMember")
	errDuplicateGroup       = errors.New("errDuplicateGroup")
	errDuplicateAdmin       = errors.New("errDuplicateAdmin")
	errInvalidVoteTime      = errors.New("errInvalidVoteTime")
	errInvalidVoteOption    = errors.New("errInvalidVoteOption")
	errEmptyVoteGroup       = errors.New("errEmptyVoteGroup")
	errVoteNotExist         = errors.New("errVoteNotExist")
	errGroupNotExist        = errors.New("errGroupNotExist")
	errStateDBGet           = errors.New("errStateDBGet")
	errInvalidVoteID        = errors.New("errInvalidVoteID")
	errInvalidGroupID       = errors.New("errInvalidGroupID")
	errInvalidOptionIndex   = errors.New("errInvalidOptionIndex")
	errAddrAlreadyVoted     = errors.New("errAddrAlreadyVoted")
	errInvalidGroupMember   = errors.New("errInvalidGroupMember")
	errVoteAlreadyFinished  = errors.New("errVoteAlreadyFinished")
	errVoteAlreadyClosed    = errors.New("errVoteAlreadyClosed")
	errAddrPermissionDenied = errors.New("errAddrPermissionDenied")
)
