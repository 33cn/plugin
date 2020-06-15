package events

// -----------------------------------------------------
// 	Events: Events maintains a mapping of events to an array
//		of claims made by validators.
// -----------------------------------------------------

// EventRecords : map of transaction hashes to LockEvent structs
var EventRecords = make(map[string]LockEvent)

// NewEventWrite : add a validator's address to the official claims list
func NewEventWrite(txHash string, event LockEvent) {
	EventRecords[txHash] = event
}
