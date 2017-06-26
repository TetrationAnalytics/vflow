//: ----------------------------------------------------------------------------
//: Copyright (C) 2017 Verizon.  All Rights Reserved.
//: All Rights Reserved
//:
//: file:    marshal_pb.go
//: details: encoding of each decoded netflow v9 data sets
//: author:  Tapan Patwardhan
//: date:    04/27/2017
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

package main

import (
	"bytes"
	"encoding/hex"
	"net"
	"reflect"
	"strconv"
	"strings"

	"github.com/TetrationAnalytics/vflow/netflow/v9"
	"github.com/golang/protobuf/proto"
	SensorProto "github.com/TetrationAnalytics/vflow/vflow/protos/sensor"
)

// DumpFlowRecord prints the incoming netflow record
func DumpFlowRecord(m *netflow9.Message) error {
	var (
		b   bytes.Buffer
		str string
	)
	// Dump agent info
	b.WriteString("Agent ID: ")
	b.WriteString(m.AgentID)
	b.WriteString(", Version: ")
	// Header
	b.WriteString(strconv.FormatInt(int64(m.Header.Version), 10))
	b.WriteByte('\n')
	for i := range m.DataSets {
		str = "\nFlow Record " + strconv.FormatInt(int64(i+1), 10) + ": \n"
		b.WriteString(str)
		for j := range m.DataSets[i] {
			str = "Elem ID: " + strconv.FormatInt(int64(m.DataSets[i][j].ID), 10)
			b.WriteString(str)
			str = " Type: " + reflect.TypeOf(m.DataSets[i][j].Value).Name()
			b.WriteString(str)
			b.WriteString(" Value: ")
			switch m.DataSets[i][j].Value.(type) {
			case uint:
				b.WriteString(strconv.FormatInt(int64(m.DataSets[i][j].Value.(uint)), 10))
			case uint8:
				b.WriteString(strconv.FormatInt(int64(m.DataSets[i][j].Value.(uint8)), 10))
			case uint16:
				b.WriteString(strconv.FormatInt(int64(m.DataSets[i][j].Value.(uint16)), 10))
			case uint32:
				b.WriteString(strconv.FormatInt(int64(m.DataSets[i][j].Value.(uint32)), 10))
			case uint64:
				b.WriteString(strconv.FormatInt(int64(m.DataSets[i][j].Value.(uint64)), 10))
			case int:
				b.WriteString(strconv.FormatInt(int64(m.DataSets[i][j].Value.(int)), 10))
			case int8:
				b.WriteString(strconv.FormatInt(int64(m.DataSets[i][j].Value.(int8)), 10))
			case int16:
				b.WriteString(strconv.FormatInt(int64(m.DataSets[i][j].Value.(int16)), 10))
			case int32:
				b.WriteString(strconv.FormatInt(int64(m.DataSets[i][j].Value.(int32)), 10))
			case int64:
				b.WriteString(strconv.FormatInt(int64(m.DataSets[i][j].Value.(int64)), 10))
			case float32:
				b.WriteString(strconv.FormatFloat(float64(m.DataSets[i][j].Value.(float32)), 'E', -1, 32))
			case float64:
				b.WriteString(strconv.FormatFloat(m.DataSets[i][j].Value.(float64), 'E', -1, 64))
			case string:
				b.WriteByte('"')
				b.WriteString(m.DataSets[i][j].Value.(string))
				b.WriteByte('"')
			case net.IP:
				b.WriteByte('"')
				b.WriteString(m.DataSets[i][j].Value.(net.IP).String())
				b.WriteByte('"')
			case net.HardwareAddr:
				b.WriteByte('"')
				b.WriteString(m.DataSets[i][j].Value.(net.HardwareAddr).String())
				b.WriteByte('"')
			case []uint8:
				b.WriteByte('"')
				b.WriteString("0x" + hex.EncodeToString(m.DataSets[i][j].Value.([]uint8)))
				b.WriteByte('"')
			default:
				b.WriteString("Invalid value")
			}
			b.WriteByte('\n')

		}
	}
	logger.Println(b.String())
	return nil
}

// ProtoBufMarshal encodes netflow v9 message to protobuf
func ProtoBufMarshal(m *netflow9.Message) ([]byte, error) {
	var (
		err      error
		ipProto  uint32
		v        int32
		srcPort  int32
		dstPort  int32
		srcAddr  net.IP
		dstAddr  net.IP
		validRec bool
		b        bytes.Buffer
		str      string
	)
	var flowInfoArray []*SensorProto.FlowInfo

	for i := range m.DataSets {
		str = "\nFlow Record " + strconv.FormatInt(int64(i+1), 10) + ": "
		b.WriteString(str)
		validRec = true
		for j := range m.DataSets[i] {
			if m.DataSets[i][j].ID == 4 || m.DataSets[i][j].ID == 7 || m.DataSets[i][j].ID == 11 ||
				m.DataSets[i][j].ID == 8 || m.DataSets[i][j].ID == 12 || m.DataSets[i][j].ID == 27 ||
				m.DataSets[i][j].ID == 28 {
				switch m.DataSets[i][j].ID {
				// Trasport Protocol
				case 4:
					if strings.Compare(reflect.TypeOf(m.DataSets[i][j].Value).Name(), "uint8") != 0 {
						b.WriteString(" Protocol invalid ")
						validRec = false
						break
					} else {
						ipProto = uint32(m.DataSets[i][j].Value.(uint8))
						str = " Protocol: " + strconv.FormatInt(int64(ipProto), 10)
						b.WriteString(str)
					}

				// Source Port
				case 7:
					if strings.Compare(reflect.TypeOf(m.DataSets[i][j].Value).Name(), "uint16") != 0 {
						b.WriteString(" SrcPort invalid ")
						validRec = false
						break
					} else {
						srcPort = int32(m.DataSets[i][j].Value.(uint16))
						str = " Src Port: " + strconv.FormatInt(int64(srcPort), 10)
						b.WriteString(str)
					}

				// Destintion Port
				case 11:
					if strings.Compare(reflect.TypeOf(m.DataSets[i][j].Value).Name(), "uint16") != 0 {
						b.WriteString(" DstPort invalid ")
						validRec = false
						break
					} else {
						dstPort = int32(m.DataSets[i][j].Value.(uint16))
						str = " Dst Port: " + strconv.FormatInt(int64(srcPort), 10)
						b.WriteString(str)
					}

				// Src IPv4 address
				case 8:
					if strings.Compare(reflect.TypeOf(m.DataSets[i][j].Value).Name(), "IP") != 0 {
						b.WriteString(" IPv4 src address invalid ")
					} else {
						v = SensorProto.NetworkCommunicationInfo_AddressType_value["IPV4"]
						srcAddr = m.DataSets[i][j].Value.(net.IP)
						b.WriteString(" Src Addr " + srcAddr.String())
					}

				// Dst IPv4 address
				case 12:
					if strings.Compare(reflect.TypeOf(m.DataSets[i][j].Value).Name(), "IP") != 0 {
						b.WriteString(" IPv4 dst address invalid ")
					} else {
						v = SensorProto.NetworkCommunicationInfo_AddressType_value["IPV4"]
						dstAddr = m.DataSets[i][j].Value.(net.IP)
						b.WriteString(" Dst Addr " + dstAddr.String())
					}

				// Src IPv6 address
				case 27:
					if strings.Compare(reflect.TypeOf(m.DataSets[i][j].Value).Name(), "IP") != 0 {
						b.WriteString(" IPv6 src address invalid ")
					} else {
						v = SensorProto.NetworkCommunicationInfo_AddressType_value["IPV6"]
						srcAddr = m.DataSets[i][j].Value.(net.IP)
						b.WriteString(" Src Addr " + srcAddr.String())
					}

				// Dst IPv6 address
				case 28:
					if strings.Compare(reflect.TypeOf(m.DataSets[i][j].Value).Name(), "IP") != 0 {
						b.WriteString(" IPv6 dst address invalid ")
					} else {
						v = SensorProto.NetworkCommunicationInfo_AddressType_value["IPV6"]
						dstAddr = m.DataSets[i][j].Value.(net.IP)
						b.WriteString(" Dst Addr " + dstAddr.String())
					}
					/*
						// Src MAC address
						case 56:
								srcAddr = m.DataSets[i][j].Value.(net.HardwareAddr).String()

						// Dst MAC address
						case 80:
								dstAddr = m.DataSets[i][j].Value.(net.HardwareAddr).String()
					*/
				}
			}
			if validRec != true {
				break
			}
		}

		if validRec {
			keyType := SensorProto.FlowKey_KeyType(v)

			// Input into Protobuf  FlowStartTime?
			flowKey := &SensorProto.FlowKey{
				Proto:      proto.Uint32(ipProto),
				KeyType:    &keyType,
				SrcAddress: srcAddr,
				SrcPort:    proto.Uint32(uint32(srcPort)),
				DstAddress: dstAddr,
				DstPort:    proto.Uint32(uint32(dstPort)),
			}

			sensorType := SensorProto.FlowInfo_SensorType(SensorProto.FlowInfo_SW_LIGHTWEIGHT)

			flowInfo := &SensorProto.FlowInfo{
				Key:                    flowKey,
				SensorType:             &sensorType,
				SourceSockInListenMode: proto.Bool(false),
			}

			flowInfoArray = append(flowInfoArray, flowInfo)

		}

	}

	// TODO: Add check if we are exceeding export limit

	netflowV9Flows := &SensorProto.FlowInfoFromSensor{
		SensorId: proto.String(m.AgentID),
		FlowInfo: flowInfoArray,
	}

	data, err := proto.Marshal(netflowV9Flows)
	if err != nil {
		return nil, err
	}

	return data, nil
}
