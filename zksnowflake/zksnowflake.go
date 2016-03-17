package zksnowflake

import (
	"github.com/samuel/go-zookeeper/zk"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
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

const (
	LOCK_PATH = "/_snowflake_/lock/"
	NODE_PATH = "/_snowflake_/nodes/"
)

var (
	_zk_conn    *zk.Conn
	_sfc_map    map[string]*SnowFlakeCloud
	global_lock sync.RWMutex
)

type int64arr []int64

func (a int64arr) Len() int           { return len(a) }
func (a int64arr) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a int64arr) Less(i, j int) bool { return a[i] < a[j] }

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

func Dialog() {
	t := make_timestamp()
	sid := make_snowflake(t, 1, 2)
	log.Println(melt_snowflakeid(sid))
	log.Println("max id", snowflake_maxid())
	log.Println("max process id", max_process_id)
	log.Println("algorithm expire at", snowflake_expire_date())
	log.Println("snowflake id:", sid)

}

type SnowFlake struct {
	sequence   int64
	last_ts    int64
	last_seq   int64
	process_id int64
	lock       sync.RWMutex
}

func (sf *SnowFlake) getSequnce() int64 {
	sf.sequence += 1
	sf.sequence %= max_sequence_id
	return sf.sequence
}

func (sf *SnowFlake) gen() int64 {
	sf.lock.Lock()
	defer sf.lock.Unlock()
	for {
		ts := make_timestamp()
		seq := sf.getSequnce()
		if ts == sf.last_ts && seq <= sf.last_seq {
			time.Sleep(1 * time.Millisecond)
		} else {
			sid := make_snowflake(ts, sf.process_id, seq)
			sf.last_ts = ts
			sf.last_seq = seq
			return sid
		}
	}
}

type SnowFlakeCloud struct {
	zk_namespace string //show be unique for each id field
	lock_path    string
	node_path    string
	zk_lock      *zk.Lock
	sf           *SnowFlake
}

func (sfc *SnowFlakeCloud) ensure_paths(paths string) {
	_path := ""
	for _, p := range strings.Split(paths[1:], "/") {
		_path += "/" + p
		_, err := _zk_conn.Create(_path, []byte{1}, 0, zk.WorldACL(zk.PermAll))
		if err == nil || err == zk.ErrNodeExists {

		} else {
			log.Panicln(err)
		}
	}
}

func (sfc *SnowFlakeCloud) Init() {
	sfc.ensure_paths(sfc.lock_path)
	sfc.ensure_paths(sfc.node_path)
	sfc.zk_lock = zk.NewLock(_zk_conn, sfc.lock_path, zk.WorldACL(zk.PermAll))
}

func (sfc *SnowFlakeCloud) register() {
	err := sfc.zk_lock.Lock()
	if err != nil {
		log.Panicln(err)
	}
	nodes, _, err := _zk_conn.Children(sfc.node_path)
	if err != nil {
		log.Panicln(err)
	}
	var pid int64
	pid = 1
	if len(nodes) > 0 {
		node_ids := make(int64arr, 0, len(nodes))
		for _, s := range nodes {
			i, _ := strconv.ParseInt(s, 10, 64)
			node_ids = append(node_ids, i)
		}
		sort.Sort(node_ids)
		log.Println(node_ids)
		if node_ids[len(node_ids)-1] < max_process_id-1 {
			pid = node_ids[len(node_ids)-1] + 1
		} else {
			log.Println(node_ids[len(node_ids)-1])
			for i := int64(0); i < node_ids[len(node_ids)-1]+1; i++ {
				log.Println(i)
				if i+1 != node_ids[i] {
					pid = i + 1
					break
				} else if i == int64(len(node_ids))-1 {
					pid = i + 2
					break
				} else if node_ids[i] != node_ids[i+1]-1 {
					pid = node_ids[i] + 1
					break
				}
			}
		}
	}
	log.Println("zksnowflake process_id", pid)
	_, err = _zk_conn.Create(sfc.node_path+"/"+strconv.FormatInt(pid, 10),
		[]byte{1}, zk.FlagEphemeral,
		zk.WorldACL(zk.PermAll),
	)
	if err != nil {
		log.Panicln(err)
	}
	sfc.sf = &SnowFlake{process_id: pid}
	err = sfc.zk_lock.Unlock()
	if err != nil {
		log.Panicln(err)
	}
}

func (sfc *SnowFlakeCloud) Gen() int64 {
	return sfc.sf.gen()
}

func Setup(zk_config interface{}) {
	global_lock.Lock()
	defer global_lock.Unlock()
	if _zk_conn != nil {
		log.Println("zookeeper already setup", _zk_conn)
		return
	}
	//Though can be imported many times, but only execute once
	if __zk_config, ok := zk_config.(string); ok {
		log.Println("set zookeeper hosts:", __zk_config)
		z, _, err := zk.Connect(strings.Split(__zk_config, ","), time.Second) //*10)
		if err != nil {
			log.Panicln("zookeeper fail", err)
		} else {
			log.Println("zookeeper connected")
		}
		_zk_conn = z
	} else if __zk_conn, ok := zk_config.(*zk.Conn); ok {
		log.Println("set zookeeper:", __zk_conn)
		_zk_conn = __zk_conn
	}

}

func GetGenerator(namespace string) *SnowFlakeCloud {
	global_lock.Lock()
	defer global_lock.Unlock()
	if sfc, exist := _sfc_map[namespace]; exist { //if in the map then return
		return sfc
	} else { //else create a new one
		sfc = &SnowFlakeCloud{
			zk_namespace: namespace,
			node_path:    NODE_PATH + namespace,
			lock_path:    LOCK_PATH + namespace,
		}
		sfc.Init()
		sfc.register()
		_sfc_map[namespace] = sfc
		return sfc
	}
}

func init() {
	_sfc_map = make(map[string]*SnowFlakeCloud)
}
