package transport

import (
	"encoding/binary"
	"errors"
	"net"
)

type SerializableType interface {
	Marshal(buffer []byte) (int, error)
	UnMarshal(buffer []byte) (int, error)
}

const (
	RandomByteLength            = 1
	AccessTokenLengthByteLength = 1
)

var (
	BufferLengthLimitErr = errors.New("marshal AuthType buffer length limited")
)

type AuthType struct {
	AccessToken  string
	FunctionFlag byte // 功能位，用于标记是否是否启动一定功能	# todo 目前没有使用
}

func (data *AuthType) Marshal(buffer []byte) (int, error) {
	lengthNeed := RandomByteLength + AccessTokenLengthByteLength + len(data.AccessToken)
	if len(buffer) < lengthNeed {
		return 0, BufferLengthLimitErr
	}

	buffer[0] = getRandomBytes()

	buffer[RandomByteLength] = byte(len(data.AccessToken))

	copy(buffer[RandomByteLength+AccessTokenLengthByteLength:], data.AccessToken)
	return lengthNeed, nil
}

func (data *AuthType) UnMarshal(buffer []byte) (int, error) {
	minLengthNeed := RandomByteLength + AccessTokenLengthByteLength // 这里先忽略 AccessToken 长度
	if len(buffer) < minLengthNeed {
		return 0, BufferLengthLimitErr
	}

	tokenLen := int(buffer[RandomByteLength])
	if len(buffer) < (minLengthNeed + tokenLen) {
		return 0, BufferLengthLimitErr
	}

	data.AccessToken = string(buffer[minLengthNeed : minLengthNeed+tokenLen])
	return minLengthNeed + tokenLen, nil
}

type AddressEnumType byte
type TransProtocolType byte // 用于描述传输层的协议类型

const (
	AddressEnumTypeIpv4 = 0x01 + iota
	AddressEnumTypeIpv6
	AddressEnumTypeDomain
	AddressEnumTypeUnknown
)

const (
	TransProtocolTypeTcp = 0x01 + iota
	TransProtocolTypeUdp
)

const (
	TransProtocolTypeLength = 1
	AddressTypeByteLength   = 1
	AddressIpv4Length       = 4
	AddressIpv6Length       = 16
	AddressDomainLength     = 1
	PortLength              = 2
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

	if len(buffer) < (TransProtocolTypeLength) {
		return 0, BufferLengthLimitErr
	}
	buffer[0] = byte(data.ProtocolType)
	lengthUsed := TransProtocolTypeLength

	if checkTransProtocolType(data.ProtocolType) == false {
		return 0, TransProtocolIllegalErr
	}

	switch data.AddressType {
	case AddressEnumTypeIpv4:
		lengthNeed := TransProtocolTypeLength + AddressTypeByteLength + AddressIpv4Length + PortLength
		if len(buffer) < lengthNeed {
			return 0, BufferLengthLimitErr
		}

		buffer[lengthUsed] = byte(data.AddressType)
		lengthUsed += AddressTypeByteLength

		if len(data.Host) < AddressIpv4Length {
			return 0, HostLengthLimitedErr
		}
		copy(buffer[lengthUsed:lengthUsed+AddressIpv4Length], net.ParseIP(data.Host).To4()[:AddressIpv4Length])
		lengthUsed += AddressIpv4Length

		binary.BigEndian.PutUint16(buffer[lengthUsed:], data.Port)
		return lengthNeed, nil
	case AddressEnumTypeIpv6:
		lengthNeed := TransProtocolTypeLength + AddressTypeByteLength + AddressIpv6Length + PortLength
		if len(buffer) < lengthNeed {
			return 0, BufferLengthLimitErr
		}
		buffer[lengthUsed] = byte(data.AddressType)
		lengthUsed += AddressTypeByteLength

		if len(data.Host) < AddressIpv6Length {
			return 0, HostLengthLimitedErr
		}
		copy(buffer[lengthUsed:lengthUsed+AddressIpv6Length], net.ParseIP(data.Host)[:AddressIpv6Length])
		lengthUsed += AddressIpv6Length

		binary.BigEndian.PutUint16(buffer[lengthUsed:], data.Port)
		return lengthNeed, nil
	case AddressEnumTypeDomain:
		domainLen := len(data.Host)
		lengthNeed := TransProtocolTypeLength + AddressTypeByteLength + AddressDomainLength + domainLen + PortLength
		if len(buffer) < (lengthNeed) {
			return 0, BufferLengthLimitErr
		}

		buffer[lengthUsed] = byte(data.AddressType)
		lengthUsed += AddressTypeByteLength

		buffer[lengthUsed] = byte(domainLen)
		lengthUsed += AddressDomainLength

		copy(buffer[lengthUsed:lengthUsed+domainLen], data.Host)
		lengthUsed += domainLen

		binary.BigEndian.PutUint16(buffer[lengthUsed:], data.Port)
		return lengthNeed, nil
	default:
		return 0, AddressTypeIllegalErr
	}
}

func (data *ConnectType) UnMarshal(buffer []byte) (int, error) {
	if len(buffer) < TransProtocolTypeLength {
		return 0, BufferLengthLimitErr
	}

	data.ProtocolType = TransProtocolType(buffer[0])
	lengthUsed := TransProtocolTypeLength
	if len(buffer) < (lengthUsed + AddressTypeByteLength) {
		return 0, BufferLengthLimitErr
	}

	data.AddressType = AddressEnumType(buffer[lengthUsed])
	lengthUsed += AddressTypeByteLength

	switch data.AddressType {
	case AddressEnumTypeIpv4:
		if len(buffer) < (lengthUsed + AddressIpv4Length + PortLength) {
			return 0, BufferLengthLimitErr
		}

		data.Host = net.IP(buffer[lengthUsed : lengthUsed+AddressIpv4Length]).String()
		lengthUsed += AddressIpv4Length

		data.Port = binary.BigEndian.Uint16(buffer[lengthUsed : lengthUsed+PortLength])
		lengthUsed += PortLength
		return lengthUsed, nil
	case AddressEnumTypeIpv6:
		if len(buffer) < (lengthUsed + AddressIpv6Length + PortLength) {
			return 0, BufferLengthLimitErr
		}

		data.Host = net.IP(buffer[lengthUsed : lengthUsed+AddressIpv6Length]).String()
		lengthUsed += AddressIpv6Length

		data.Port = binary.BigEndian.Uint16(buffer[lengthUsed : lengthUsed+PortLength])
		lengthUsed += PortLength
		return lengthUsed, nil
	case AddressEnumTypeDomain:
		domainLen := int(buffer[lengthUsed])
		lengthUsed += AddressDomainLength

		if len(buffer) < (lengthUsed + domainLen + PortLength) {
			return 0, BufferLengthLimitErr
		}
		data.Host = string(buffer[lengthUsed : lengthUsed+domainLen])
		lengthUsed += domainLen

		data.Port = binary.BigEndian.Uint16(buffer[lengthUsed : lengthUsed+PortLength])
		lengthUsed += PortLength
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
