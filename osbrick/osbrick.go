package osbrick

import "sync"

// command binary
var (
	BinaryIscsiadm  = "iscsiadm"
	BinaryMultipath = "multipath"
	BinaryBlockdev  = "blockdev"
	BinaryTee       = "tee"
)

// command mutex
var (
	commandMu sync.Mutex
)
