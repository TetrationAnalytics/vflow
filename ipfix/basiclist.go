package ipfix

import (
	"encoding/binary"
	"fmt"

	"github.com/VerizonDigital/vflow/reader"
)

/*
 * basicList Encoding with Enterprise Number
    0                   1                   2                   3
    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |    Semantic   |1|         Field ID            |   Element...  |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   | ...Length     |               Enterprise Number ...           |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |      ...      |              basicList Content ...            |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |                              ...                              |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/

// ParseBasicListWithEnterpriseNumber parse module name list
func ParseBasicListWithEnterpriseNumber(listData []byte) ([]interface{}, error) {
	x := []interface{}{}
	if len(listData) < 9 {
		return nil, fmt.Errorf("not enough data to decode a basiclist")
	}
	fieldID := binary.BigEndian.Uint16(listData[1:3]) & 0x7fff
	elementLength := binary.BigEndian.Uint16(listData[3:5])
	enterpriseNumber := binary.BigEndian.Uint32(listData[5:9])
	m, ok := InfoModel[ElementKey{enterpriseNumber, fieldID}]
	if !ok {
		return nil, fmt.Errorf("IPFIX enterprise id (%d), element key (%d) not exist", enterpriseNumber, fieldID)
	}
	r := reader.NewReader(listData[9:])
	for r.Len() > 0 {
		b, err := r.Read(int(elementLength))
		if err != nil {
			return nil, err
		}

		res := Interpret(&b, m.Type)
		x = append(x, res)
	}
	return x, nil
}
