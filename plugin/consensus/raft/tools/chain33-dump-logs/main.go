// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"strconv"

	ttypes "github.com/33cn/chain33/types"
	"github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/coreos/etcd/pkg/pbutil"
	"github.com/coreos/etcd/pkg/types"
	"github.com/coreos/etcd/raft/raftpb"
	raftsnap "github.com/coreos/etcd/snap"
	"github.com/coreos/etcd/wal"
	"github.com/coreos/etcd/wal/walpb"
	"github.com/golang/protobuf/proto"
)

func main() {
	snapfile := flag.String("start-snap", "", "The base name of snapshot file to start dumping")
	index := flag.String("start-index", "", "The index to start dumping")
	flag.Parse()

	if len(flag.Args()) != 1 {
		log.Fatalf("Must provide data-dir argument (got %+v)", flag.Args())
	}
	dataDir := flag.Args()[0]

	if *snapfile != "" && *index != "" {
		log.Fatal("start-snap and start-index flags cannot be used together.")
	}

	var (
		walsnap  walpb.Snapshot
		snapshot *raftpb.Snapshot
		err      error
	)

	isIndex := *index != ""

	if isIndex {
		fmt.Printf("Start dumping log entries from index %s.\n", *index)
		walsnap.Index, _ = strconv.ParseUint(*index, 10, 64)
	} else {
		if *snapfile == "" {
			ss := raftsnap.New(snapDir(dataDir))
			snapshot, err = ss.Load()
		} else {
			snapshot, err = raftsnap.Read(filepath.Join(snapDir(dataDir), *snapfile))
		}

		switch err {
		case nil:
			walsnap.Index, walsnap.Term = snapshot.Metadata.Index, snapshot.Metadata.Term
			nodes := genIDSlice(snapshot.Metadata.ConfState.Nodes)
			fmt.Printf("Snapshot:\nterm=%d index=%d nodes=%s\n",
				walsnap.Term, walsnap.Index, nodes)
		case raftsnap.ErrNoSnapshot:
			fmt.Printf("Snapshot:\nempty\n")
		default:
			log.Fatalf("Failed loading snapshot: %v", err)
		}
		fmt.Println("Start dupmping log entries from snapshot.")
	}

	w, err := wal.OpenForRead(walDir(dataDir), walsnap)
	if err != nil {
		log.Fatalf("Failed opening WAL: %v", err)
	}
	wmetadata, state, ents, err := w.ReadAll()
	w.Close()
	if err != nil && (!isIndex || err != wal.ErrSnapshotNotFound) {
		log.Fatalf("Failed reading WAL: %v", err)
	}
	id, cid := parseWALMetadata(wmetadata)
	vid := types.ID(state.Vote)
	fmt.Printf("WAL metadata:\nnodeID=%s clusterID=%s term=%d commitIndex=%d vote=%s\n",
		id, cid, state.Term, state.Commit, vid)

	fmt.Printf("WAL entries:\n")
	fmt.Printf("lastIndex=%d\n", ents[len(ents)-1].Index)
	fmt.Printf("%4s\t%10s\ttype\tdata\n", "term", "index")
	for _, e := range ents {
		msg := fmt.Sprintf("%4d\t%10d", e.Term, e.Index)
		switch e.Type {
		case raftpb.EntryNormal:
			msg = fmt.Sprintf("%s\tnorm", msg)
			if len(e.Data) == 0 {
				break
			}
			// 解码
			block := &ttypes.Block{}
			if err := proto.Unmarshal(e.Data, block); err != nil {
				log.Printf("failed to unmarshal: %v", err)
				break
			}
			msg = fmt.Sprintf("%s\tHeight=%d\tTxHash=%X", msg, block.Height, block.TxHash)
		case raftpb.EntryConfChange:
			msg = fmt.Sprintf("%s\tconf", msg)
			var r raftpb.ConfChange
			if err := r.Unmarshal(e.Data); err != nil {
				msg = fmt.Sprintf("%s\t???", msg)
			} else {
				msg = fmt.Sprintf("%s\tmethod=%s id=%s", msg, r.Type, types.ID(r.NodeID))
			}
		}
		fmt.Println(msg)
	}
}

func walDir(dataDir string) string { return filepath.Join(dataDir, "wal") }

func snapDir(dataDir string) string { return filepath.Join(dataDir, "snap") }

func parseWALMetadata(b []byte) (id, cid types.ID) {
	var metadata etcdserverpb.Metadata
	pbutil.MustUnmarshal(&metadata, b)
	id = types.ID(metadata.NodeID)
	cid = types.ID(metadata.ClusterID)
	return id, cid
}

func genIDSlice(a []uint64) []types.ID {
	ids := make([]types.ID, len(a))
	for i, id := range a {
		ids[i] = types.ID(id)
	}
	return ids
}
