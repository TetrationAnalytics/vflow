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
	"net"

	"github.com/VerizonDigital/vflow/netflow/v9"
	"github.com/golang/protobuf/proto"
	SensorProto "github.com/tetration/vflow/vflow/protos/sensor"
)

//var errUknownProtoBufMarshalDataType = errors.New("unknown data type for protobuf marshall")

// ProtoBufMarshal encodes netflow v9 message in flow_info protobuf format
func ProtoBufMarshal(m *netflow9.Message) ([]byte, error) {
	var (
		err     error
		ipProto uint32
		v       int32
		srcPort int32
		dstPort int32
		srcAddr net.IP
		dstAddr net.IP
	)
	var flowInfoArray []*SensorProto.FlowInfo

	for i := range m.DataSets {

		for j := range m.DataSets[i] {
			switch m.DataSets[i][j].ID {
			// Trasport Protocol
			case 4:
				ipProto = uint32(m.DataSets[i][j].Value.(uint8))

			// Source Port
			case 7:
				srcPort = int32(m.DataSets[i][j].Value.(uint16))

			// Destintion Port
			case 11:
				dstPort = int32(m.DataSets[i][j].Value.(uint16))

			// Src IPv4 address
			case 8:
				v = SensorProto.NetworkCommunicationInfo_AddressType_value["IPV4"]
				srcAddr = m.DataSets[i][j].Value.(net.IP)

			// Dst IPv4 address
			case 12:
				v = SensorProto.NetworkCommunicationInfo_AddressType_value["IPV4"]
				dstAddr = m.DataSets[i][j].Value.(net.IP)

			// Src IPv6 address
			case 27:
				v = SensorProto.NetworkCommunicationInfo_AddressType_value["IPV6"]
				srcAddr = m.DataSets[i][j].Value.(net.IP)

			// Dst IPv6 address
			case 28:
				v = SensorProto.NetworkCommunicationInfo_AddressType_value["IPV6"]
				dstAddr = m.DataSets[i][j].Value.(net.IP)
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
