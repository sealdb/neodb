/*
 * NeoDB
 *
 * Copyright 2018 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package syncer

import (
	"os"
	"testing"

	"github.com/sealdb/mysqlstack/xlog"
	"github.com/stretchr/testify/assert"
)

const (
	testPeerdir = "_test_syncer_peer"
)

func testRemovePeerdir() {
	os.RemoveAll(testPeerdir)
}

func TestPeer(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))

	// Create test dir "_test_syncer_peer"
	if err := os.MkdirAll(testPeerdir, os.ModePerm); err != nil {
		log.Error("syncer.peer.test.mkdir.error:%+v", err)
	}
	defer testRemovePeerdir()

	peer := NewPeer(log, testPeerdir, "192.168.0.1:8080")
	assert.NotNil(t, peer)

	// peer.json not exist
	err := peer.LoadConfig()
	assert.Nil(t, err)

	// Add peers.
	{
		peer.Add("192.168.0.2:8080")
		peer.Add("192.168.0.3:8080")
		peer.Add("192.168.0.4:8080")

		want := []string{
			"192.168.0.1:8080",
			"192.168.0.2:8080",
			"192.168.0.3:8080",
			"192.168.0.4:8080",
		}
		got := peer.peers
		assert.Equal(t, want, got)
	}

	// Remove peers.
	{
		peer.Remove("192.168.0.3:8080")
		peer.Remove("192.168.0.4:8080")

		want := []string{
			"192.168.0.1:8080",
			"192.168.0.2:8080",
		}
		got := peer.peers
		assert.Equal(t, want, got)
	}

	// Load.
	{
		err := peer.LoadConfig()
		assert.Nil(t, err)
		want := []string{
			"192.168.0.1:8080",
			"192.168.0.2:8080",
		}
		got := peer.peers
		assert.Equal(t, want, got)
	}
}

func TestPeerError(t *testing.T) {
	defer testRemovePeerdir()

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	peer := NewPeer(log, testPeerdir, "192.168.0.1:8080")

	// Dir not exist, add fail.
	{
		err := peer.Add("192.168.0.2:8080")
		assert.NotNil(t, err)
	}

	// Create test dir
	if err := os.MkdirAll(testPeerdir, os.ModePerm); err != nil {
		log.Error("syncer.peer.test.Mkdir.error:%+v", err)
	}

	// Add empty peer.
	{
		err := peer.Add("")
		assert.NotNil(t, err)
	}

}
