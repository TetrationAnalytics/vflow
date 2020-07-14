//: ----------------------------------------------------------------------------
//: Copyright (C) 2017 Verizon.  All Rights Reserved.
//: All Rights Reserved
//:
//: file:    memcache_test.go
//: details: memory template cache testing
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
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemCacheRetrieve(t *testing.T) {
	ip := net.ParseIP("127.0.0.1")
	port := uint16(0)
	mCache := GetCache("cache.file")
	d := NewDecoder(ip, port, tpl)
	d.Decode(mCache)
	v, ok := mCache.retrieve(256, ip, port, 33792)
	if !ok {
		t.Error("expected mCache retrieve status true, got", ok)
	}
	if v.TemplateID != 256 {
		t.Error("expected template id#:256, got", v.TemplateID)
	}
}

func TestMemCacheInsert(t *testing.T) {
	var tpl TemplateRecord
	ip := net.ParseIP("127.0.0.1")
	port := uint16(0)
	mCache := GetCache("cache.file")

	tpl.TemplateID = 310
	mCache.insert(310, ip, 0, tpl, 512)

	v, ok := mCache.retrieve(310, ip, port, 512)
	if !ok {
		t.Error("expected mCache retrieve status true, got", ok)
	}
	if v.TemplateID != 310 {
		t.Error("expected template id#:310, got", v.TemplateID)
	}
}

func TestMemCacheAllSetIds(t *testing.T) {
	var tpl TemplateRecord
	ip := net.ParseIP("127.0.0.1")
	port := uint16(0)
	mCache := GetCache("cache.file")

	tpl.TemplateID = 310
	mCache.insert(tpl.TemplateID, ip, port, tpl, 512)
	tpl.TemplateID = 410
	mCache.insert(tpl.TemplateID, ip, port, tpl, 512)
	tpl.TemplateID = 210
	mCache.insert(tpl.TemplateID, ip, port, tpl, 512)

	expected := []int{210, 310, 410}
	actual := mCache.allSetIds()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected set IDs %v, got %v", expected, actual)
	}
}

func TestTemplatesShard_removeExpiredTemplate(t *testing.T) {
	type fields struct {
		Templates map[uint32]Data
	}
	type args struct {
		expiration int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "not expired",
			fields: fields{
				Templates: map[uint32]Data{
					1: Data{
						Template:  TemplateRecord{},
						Timestamp: time.Now().Unix() - 5,
					},
				},
			},
			args: args{
				expiration: 10,
			},
		},
		{
			name: "expired",
			fields: fields{
				Templates: map[uint32]Data{
					1: Data{
						Template:  TemplateRecord{},
						Timestamp: time.Now().Unix() - 5,
					},
					2: Data{
						Template:  TemplateRecord{},
						Timestamp: time.Now().Unix() - 15,
					},
				},
			},
			args: args{
				expiration: 10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t1 *testing.T) {
			shard := &TemplatesShard{
				Templates: tt.fields.Templates,
			}
			shard.removeExpiredTemplate(tt.args.expiration)
			assert.Equal(t, 1, len(shard.Templates))
		})
	}
}

func TestMemCache_RemoveExpiredTemplates(t *testing.T) {
	var tpl TemplateRecord
	ip := net.ParseIP("127.0.0.1")
	port := uint16(0)
	mCache := GetCache("cache.file")

	tpl.TemplateID = 310
	mCache.insert(310, ip, port, tpl, 512)

	v, ok := mCache.retrieve(310, ip, port, 512)
	if !ok {
		t.Error("expected mCache retrieve status true, got", ok)
	}
	if v.TemplateID != 310 {
		t.Error("expected template id#:310, got", v.TemplateID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	removeCtx, removeCancel := context.WithCancel(ctx)
	stopped := make(chan bool)
	go func() {
		defer close(stopped)
		mCache.RemoveExpiredTemplates(removeCtx, 0, 1)
	}()
	time.Sleep(2 * time.Second)
	removeCancel()
	select {
	case <-stopped:
		v, ok = mCache.retrieve(310, ip, port, 512)
		if ok {
			t.Error("expected mCache retrieve status false, got", ok)
		}
	case <-ctx.Done():
		t.Error(ctx.Err())
	}
}
