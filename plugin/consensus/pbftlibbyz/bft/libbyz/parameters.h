#ifndef _parameters_h
#define _parameters_h 1

const int Max_num_replicas = 32;

// Interval in sequence space between "checkpoint" states, i.e.,
// states that are checkpointed and for which Checkpoint messages are
// sent.
const int checkpoint_interval = 128;

// Maximum number of messages for which protocol can be
// simultaneously in progress, i.e., messages with sequence number
// higher than last_stable+max_out are ignored. It is required that
// max_out > checkpoint_interval. Otherwise, the algorithm will be
// unable to make progress.
const int max_out = 256;

#endif // _parameters_h
