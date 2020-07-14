//: ----------------------------------------------------------------------------
//: Copyright (C) 2017 Verizon.  All Rights Reserved.
//: All Rights Reserved
//:
//: file:    memcache.go
//: details: handles template caching in memory with sharding feature
//: author:  Mehrdad Arshad Rad
//: date:    02/01/2017
//:
//: Licensed under the Apache License, Version 2.0 (the "License");
//: you may not use this file except in compliance with the License.
//: You may obtain a copy of the License at
//:
//:     http://www.apache.org/licenses/LICENSE-2.0
//:
//: Unless required by applicable law or agreed to in writing, software
//: distributed under the License is distributed on an "AS IS" BASIS,
//: WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//: See the License for the specific language governing permissions and
//: limitations under the License.
//: ----------------------------------------------------------------------------

package ipfix

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"hash/fnv"
	"io/ioutil"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/VerizonDigital/vflow/log"
)

var shardNo = 32

// MemCache represents templates shards
type MemCache []*TemplatesShard

// Data represents template records and
// updated timestamp
type Data struct {
	Template  TemplateRecord
	Timestamp int64
}

// TemplatesShard represents a shard
type TemplatesShard struct {
	Templates map[uint32]Data
	sync.RWMutex
}
type memCacheDisk struct {
	Cache   MemCache
	ShardNo int
}

// GetCache tries to load saved templates
// otherwise it constructs new empty shards
func GetCache(cacheFile string) MemCache {
	var (
		mem memCacheDisk
		err error
	)

	b, err := ioutil.ReadFile(cacheFile)
	if err == nil {
		err = json.Unmarshal(b, &mem)
		if err == nil && mem.ShardNo == shardNo {
			return mem.Cache
		}
	}

	m := make(MemCache, shardNo)
	for i := 0; i < shardNo; i++ {
		m[i] = &TemplatesShard{Templates: make(map[uint32]Data)}
	}

	return m
}

func (m MemCache) getShard(id uint16, addr net.IP, srcID uint32, srcPort uint16) (*TemplatesShard, uint32) {
	bSrcID := make([]byte, 4)
	btemplateID := make([]byte, 2)
	bSrcPort := make([]byte, 2)
	binary.BigEndian.PutUint32(bSrcID, srcID)
	binary.BigEndian.PutUint16(btemplateID, id)
	binary.BigEndian.PutUint16(bSrcPort, srcPort)

	var key []byte
	key = append(key, addr...)
	key = append(key, bSrcID...)
	key = append(key, btemplateID...)
	key = append(key, bSrcPort...)

	hash := fnv.New32()
	hash.Write(key)
	hSum32 := hash.Sum32()
	return m[uint(hSum32)%uint(shardNo)], hSum32
}

func (m MemCache) insert(id uint16, addr net.IP, srcPort uint16, tr TemplateRecord, srcID uint32) {
	shard, key := m.getShard(id, addr, srcID, srcPort)
	shard.Lock()
	defer shard.Unlock()
	shard.Templates[key] = Data{tr, time.Now().Unix()}
}

func (m MemCache) retrieve(id uint16, addr net.IP, srcPort uint16, srcID uint32) (TemplateRecord, bool) {
	shard, key := m.getShard(id, addr, srcID, srcPort)
	shard.RLock()
	defer shard.RUnlock()
	v, ok := shard.Templates[key]

	return v.Template, ok
}

// Fill a slice with all known set ids. This is inefficient and is only used for error reporting or debugging.
func (m MemCache) allSetIds() []int {
	num := 0
	for _, shard := range m {
		num += len(shard.Templates)
	}
	result := make([]int, 0, num)
	for _, shard := range m {
		shard.RLock()
		for _, set := range shard.Templates {
			result = append(result, int(set.Template.TemplateID))
		}
		shard.RUnlock()
	}
	sort.Ints(result)
	return result
}

// Dump saves the current templates to hard disk
func (m MemCache) Dump(cacheFile string) error {
	b, err := json.Marshal(
		memCacheDisk{
			m,
			shardNo,
		},
	)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(cacheFile, b, 0644)
	if err != nil {
		return err
	}

	return nil
}

// RefreshTemplates start periodic check to delete expired template
// based on specified expiration in seconds and dump the cache to file after
// deletion if a dump file is provided. It should be call in a goroutine.
func (m MemCache) RefreshTemplates(ctx context.Context, interval, expiration int64, dump string) {
	if interval < 1 {
		interval = 1
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, shard := range m {
				shard.removeExpiredTemplate(expiration)
			}
			if dump != "" {
				err := m.Dump(dump)
				if err != nil {
					log.Logger.Printf("Error in dumping memCache to file, %v", err)
				}
			}
		}
	}
}

func (t *TemplatesShard) removeExpiredTemplate(expiration int64) {
	t.Lock()
	defer t.Unlock()
	for key, template := range t.Templates {
		if time.Now().Unix()-template.Timestamp > expiration {
			delete(t.Templates, key)
		}
	}
}
