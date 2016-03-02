package main

import (
	"fmt"
	"time"
)

const (
	twepoch          = 1441622034706
	timestamp_bits   = 42
	process_id_bits  = 9
	sequence_id_bits = 12
	max_timestamp    = 1 << timestamp_bits
	max_process_id   = 1 << process_id_bits
	max_sequence_id  = 1 << sequence_id_bits
)

func make_timestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func make_snowflake(timestamp_ms int64, process_id int64, sequence_id int64) int64 {
	snowflakeid := ((timestamp_ms - twepoch) % max_timestamp) << process_id_bits << sequence_id_bits
	snowflakeid += (process_id % max_process_id) << sequence_id_bits
	snowflakeid += sequence_id % max_sequence_id
	return snowflakeid
}

func melt_snowflakeid(snowflakeid int64) (int64, int64, int64) {
	sequence_id := snowflakeid & (max_sequence_id - 1)
	process_id := (snowflakeid >> sequence_id_bits) & (max_process_id - 1)
	timestamp_ms := snowflakeid >> sequence_id_bits >> process_id_bits
	timestamp_ms += twepoch
	return timestamp_ms, process_id, sequence_id
}

func snowflake_maxid() int64 {
	var _max int64
	_max = (max_timestamp - 1) << process_id_bits << sequence_id_bits
	_max += (max_process_id - 1) << sequence_id_bits
	_max += max_sequence_id - 1
	return _max
}

func snowflake_expire_date() time.Time {
	return time.Unix((max_timestamp-1+twepoch)/1000, 0)
}

func dialog() {
	t := make_timestamp()
	sid := make_snowflake(t, 1, 2)
	fmt.Println(melt_snowflakeid(sid))
	fmt.Println("max id", snowflake_maxid())
	fmt.Println("algorithm expire at", snowflake_expire_date())
	fmt.Println("snowflake id:", sid)

}
func main() {
	dialog()
}
