// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package raft

import (
	"io/ioutil"
	"net/http"
	"strconv"

	"fmt"

	"github.com/coreos/etcd/raft/raftpb"
)

// Handler for a http based httpRaftAPI backed by raft
type httpRaftAPI struct {
	confChangeC chan<- raftpb.ConfChange
}

func (h *httpRaftAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := r.RequestURI
	switch {
	case r.Method == "POST":
		url, err := ioutil.ReadAll(r.Body)
		if err != nil {
			rlog.Error(fmt.Sprintf("Failed to convert ID for conf change (%v)", err.Error()))
			http.Error(w, "Failed on POST", http.StatusBadRequest)
			return
		}

		nodeID, err := strconv.ParseUint(key[1:], 0, 64)
		if err != nil {
			rlog.Error(fmt.Sprintf("Failed to convert ID for conf change (%v)", err.Error()))
			http.Error(w, "Failed on POST", http.StatusBadRequest)
			return
		}

		cc := raftpb.ConfChange{
			Type:    raftpb.ConfChangeAddNode,
			NodeID:  nodeID,
			Context: url,
		}
		h.confChangeC <- cc
		// As above, optimistic that raft will apply the conf change
		w.WriteHeader(http.StatusCreated)
	case r.Method == "DELETE":
		nodeID, err := strconv.ParseUint(key[1:], 0, 64)
		if err != nil {
			rlog.Error(fmt.Sprintf("Failed to convert ID for conf change (%v)", err.Error()))
			http.Error(w, "Failed on DELETE", http.StatusBadRequest)
			return
		}
		cc := raftpb.ConfChange{
			Type:   raftpb.ConfChangeRemoveNode,
			NodeID: nodeID,
		}
		h.confChangeC <- cc
		// As above, optimistic that raft will apply the conf change
		w.WriteHeader(http.StatusAccepted)
	default:
		w.Header().Add("Allow", "POST")
		w.Header().Add("Allow", "DELETE")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func serveHTTPRaftAPI(port int, confChangeC chan<- raftpb.ConfChange, errorC <-chan error) {
	srv := http.Server{
		Addr: "localhost:" + strconv.Itoa(port),
		Handler: &httpRaftAPI{
			confChangeC: confChangeC,
		},
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			rlog.Error(fmt.Sprintf("ListenAndServe have a err: (%v)", err.Error()))
		}
	}()

	// exit when raft goes down
	if err, ok := <-errorC; ok {
		rlog.Error(fmt.Sprintf("the errorC chan receive a err (%v)\n", err.Error()))
	}
}
