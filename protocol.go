package transport

import (
	"encoding/binary"
	"errors"
	"net"
)

var (
	ErrBufferLengthLimit = errors.New("marshal AuthType buffer length limited")
)

type SerializableType interface {
	Marshal(buffer []byte) (int, error)
	UnMarshal(buffer []byte) (int, error)
}

const (
	randomByteLength            = 1
	accessTokenLengthByteLength = 1
	functionTagLength           = 1
)

type functionFlagType uint8
const (
	FunctionTagMux functionFlagType = 0x01 << iota
)

type AuthType struct {
	Token    string
	FuncFlag functionFlagType // 功能位，用于标记是否是否启动一定功能
}

func (auth *AuthType) SetFuncTag(flag functionFlagType, enabled bool) {
	if enabled {
		auth.FuncFlag |= flag
	} else {
		auth.FuncFlag &= ^flag
	}
}

func (auth *AuthType) IsFuncEnabled(flag functionFlagType) bool {
	return (auth.FuncFlag & flag) > 0
}

func (auth *AuthType) Marshal(buffer []byte) (int, error) {
	lengthNeed := randomByteLength + accessTokenLengthByteLength + len(auth.Token) + functionTagLength
	if len(buffer) < lengthNeed {
		return 0, ErrBufferLengthLimit
	}

	used := 0

	buffer[used] = getRandomBytes()
	used += randomByteLength

	buffer[used] = byte(len(auth.Token))
	used += accessTokenLengthByteLength

	tokenLength := len(auth.Token)
	copy(buffer[used:used+tokenLength], auth.Token)
	used += tokenLength

	buffer[used] = byte(auth.FuncFlag)
	used += functionTagLength

	return lengthNeed, nil
}

func (auth *AuthType) UnMarshal(buffer []byte) (int, error) {
	minLengthNeed := randomByteLength + accessTokenLengthByteLength // 这里先忽略 Token 长度
	if len(buffer) < minLengthNeed {
		return 0, ErrBufferLengthLimit
	}

	tokenLen := int(buffer[randomByteLength])
	if len(buffer) < (minLengthNeed + tokenLen + functionTagLength) {
		return 0, ErrBufferLengthLimit
	}
	used := 2
	auth.Token = string(buffer[used : used+tokenLen])
	used += tokenLen

	auth.FuncFlag = functionFlagType(buffer[used])
	used += functionTagLength

	return used, nil
}

type AddressEnumType byte
type TransProtocolType byte // 用于描述传输层的协议类型

const (
	AddressEnumTypeIpv4 AddressEnumType = 0x01 + iota
	AddressEnumTypeIpv6
	AddressEnumTypeDomain
	AddressEnumTypeUnknown
)

const (
	TransProtocolTypeTcp TransProtocolType = 0x01 + iota
	TransProtocolTypeUdp
)

func (proto TransProtocolType) String() string {
	switch proto {
	case TransProtocolTypeTcp:
		return "tcp"
	case TransProtocolTypeUdp:
		return "udp"
	default:
		return "unknown"
	}
}

const (
	transProtocolTypeLength = 1
	addressTypeByteLength = 1
	addressIpv4Length   = 4
	addressIpv6Length   = 16
	addressDomainLength = 1
	portLength          = 2
)

var (
	HostLengthLimitedErr    = errors.New("host length limited err")
	AddressTypeIllegalErr   = errors.New("address type illegal")
	TransProtocolIllegalErr = errors.New("protocol type illegal")
)

type ConnectType struct {
	ProtocolType TransProtocolType
	AddressType  AddressEnumType
	Host         string
	Port         uint16
}

func (data *ConnectType) Marshal(buffer []byte) (int, error) {
	if checkTransProtocolType(data.ProtocolType) == false {
		return 0, TransProtocolIllegalErr
	}

	if len(buffer) < (transProtocolTypeLength) {
		return 0, ErrBufferLengthLimit
	}
	buffer[0] = byte(data.ProtocolType)
	lengthUsed := transProtocolTypeLength

	if checkTransProtocolType(data.ProtocolType) == false {
		return 0, TransProtocolIllegalErr
	}

	switch data.AddressType {
	case AddressEnumTypeIpv4:
		lengthNeed := transProtocolTypeLength + addressTypeByteLength + addressIpv4Length + portLength
		if len(buffer) < lengthNeed {
			return 0, ErrBufferLengthLimit
		}

		buffer[lengthUsed] = byte(data.AddressType)
		lengthUsed += addressTypeByteLength

		if len(data.Host) < addressIpv4Length {
			return 0, HostLengthLimitedErr
		}
		copy(buffer[lengthUsed:lengthUsed+addressIpv4Length], net.ParseIP(data.Host).To4()[:addressIpv4Length])
		lengthUsed += addressIpv4Length

		binary.BigEndian.PutUint16(buffer[lengthUsed:], data.Port)
		return lengthNeed, nil
	case AddressEnumTypeIpv6:
		lengthNeed := transProtocolTypeLength + addressTypeByteLength + addressIpv6Length + portLength
		if len(buffer) < lengthNeed {
			return 0, ErrBufferLengthLimit
		}
		buffer[lengthUsed] = byte(data.AddressType)
		lengthUsed += addressTypeByteLength

		if len(data.Host) < addressIpv6Length {
			return 0, HostLengthLimitedErr
		}
		copy(buffer[lengthUsed:lengthUsed+addressIpv6Length], net.ParseIP(data.Host)[:addressIpv6Length])
		lengthUsed += addressIpv6Length

		binary.BigEndian.PutUint16(buffer[lengthUsed:], data.Port)
		return lengthNeed, nil
	case AddressEnumTypeDomain:
		domainLen := len(data.Host)
		lengthNeed := transProtocolTypeLength + addressTypeByteLength + addressDomainLength + domainLen + portLength
		if len(buffer) < (lengthNeed) {
			return 0, ErrBufferLengthLimit
		}

		buffer[lengthUsed] = byte(data.AddressType)
		lengthUsed += addressTypeByteLength

		buffer[lengthUsed] = byte(domainLen)
		lengthUsed += addressDomainLength

		copy(buffer[lengthUsed:lengthUsed+domainLen], data.Host)
		lengthUsed += domainLen

		binary.BigEndian.PutUint16(buffer[lengthUsed:], data.Port)
		return lengthNeed, nil
	default:
		return 0, AddressTypeIllegalErr
	}
}

func (data *ConnectType) UnMarshal(buffer []byte) (int, error) {
	if len(buffer) < transProtocolTypeLength {
		return 0, ErrBufferLengthLimit
	}

	data.ProtocolType = TransProtocolType(buffer[0])
	lengthUsed := transProtocolTypeLength
	if len(buffer) < (lengthUsed + addressTypeByteLength) {
		return 0, ErrBufferLengthLimit
	}

	data.AddressType = AddressEnumType(buffer[lengthUsed])
	lengthUsed += addressTypeByteLength

	switch data.AddressType {
	case AddressEnumTypeIpv4:
		if len(buffer) < (lengthUsed + addressIpv4Length + portLength) {
			return 0, ErrBufferLengthLimit
		}

		data.Host = net.IP(buffer[lengthUsed : lengthUsed+addressIpv4Length]).String()
		lengthUsed += addressIpv4Length

		data.Port = binary.BigEndian.Uint16(buffer[lengthUsed : lengthUsed+portLength])
		lengthUsed += portLength
		return lengthUsed, nil
	case AddressEnumTypeIpv6:
		if len(buffer) < (lengthUsed + addressIpv6Length + portLength) {
			return 0, ErrBufferLengthLimit
		}

		data.Host = net.IP(buffer[lengthUsed : lengthUsed+addressIpv6Length]).String()
		lengthUsed += addressIpv6Length

		data.Port = binary.BigEndian.Uint16(buffer[lengthUsed : lengthUsed+portLength])
		lengthUsed += portLength
		return lengthUsed, nil
	case AddressEnumTypeDomain:
		domainLen := int(buffer[lengthUsed])
		lengthUsed += addressDomainLength

		if len(buffer) < (lengthUsed + domainLen + portLength) {
			return 0, ErrBufferLengthLimit
		}
		data.Host = string(buffer[lengthUsed : lengthUsed+domainLen])
		lengthUsed += domainLen

		data.Port = binary.BigEndian.Uint16(buffer[lengthUsed : lengthUsed+portLength])
		lengthUsed += portLength
		return lengthUsed, nil
	default:
		return 0, errors.New("address type illegal")
	}
}

func checkTransProtocolType(protocolType TransProtocolType) bool {
	if protocolType == TransProtocolTypeTcp || protocolType == TransProtocolTypeUdp {
		return true
	}
	return false
}
