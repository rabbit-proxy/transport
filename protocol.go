package transport

import (
	"encoding/binary"
	"errors"
	"go.uber.org/zap"
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
	FunctionFlag byte // 功能位，用于标记是否是否启动一定功能	# todo 将 mux 功能集成到这个中
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
type TransProtocolType byte // 用于描述传输层的协议类型 todo 在下一个版本实现这个功能

const (
	AddressEnumTypeIpv4 = 0x01 << iota
	AddressEnumTypeIpv6
	AddressEnumTypeDomain
	AddressEnumTypeUnknown

	// todo 可以与上三个数据项目合并成为一个字节
	TransProtocolTypeTcp = 0x01 << iota
	TransProtocolTypeUdp
	TransProtocolTypeRaw
	TransProtocolTypeIcmp

	AddressTypeByteLength = 1
	AddressIpv4Length     = 4
	AddressIpv6Length     = 16
	AddressDomainLength   = 1
	PortLength            = 2
)

var (
	HostLengthLimitedErr  = errors.New("host length limited err")
	AddressTypeIllegalErr = errors.New("address type illegal")
)

type ConnectType struct {
	AddressType AddressEnumType
	Host        string
	Port        uint16
}

func (data *ConnectType) Marshal(buffer []byte) (int, error) {
	switch data.AddressType {
	case AddressEnumTypeIpv4:
		if len(buffer) < AddressTypeByteLength+AddressIpv4Length+PortLength {
			return 0, BufferLengthLimitErr
		}
		buffer[0] = byte(data.AddressType)
		if len(data.Host) < AddressIpv4Length {
			return 0, HostLengthLimitedErr
		}
		zap.L().Info("host", zap.String("host", data.Host))
		copy(buffer[AddressTypeByteLength:AddressTypeByteLength+AddressIpv4Length], net.ParseIP(data.Host).To4()[:AddressIpv4Length])
		binary.BigEndian.PutUint16(buffer[AddressTypeByteLength+AddressIpv4Length:], data.Port)
		return AddressTypeByteLength + AddressIpv4Length + PortLength, nil
	case AddressEnumTypeIpv6:
		if len(buffer) < (AddressTypeByteLength + AddressIpv6Length + PortLength) {
			return 0, BufferLengthLimitErr
		}
		buffer[0] = byte(data.AddressType)
		if len(data.Host) < AddressIpv6Length {
			return 0, HostLengthLimitedErr
		}
		zap.L().Info("host", zap.String("host", data.Host))
		copy(buffer[AddressTypeByteLength:AddressTypeByteLength+AddressIpv6Length], net.ParseIP(data.Host)[:AddressIpv6Length])
		binary.BigEndian.PutUint16(buffer[AddressTypeByteLength+AddressIpv6Length:], data.Port)
		return AddressTypeByteLength + AddressIpv6Length + PortLength, nil
	case AddressEnumTypeDomain:
		domainLen := len(data.Host)
		if len(buffer) < (AddressTypeByteLength + AddressDomainLength + domainLen + PortLength) {
			return 0, BufferLengthLimitErr
		}
		buffer[0] = byte(data.AddressType)
		buffer[AddressTypeByteLength] = byte(domainLen)
		copy(buffer[AddressTypeByteLength+AddressDomainLength:AddressTypeByteLength+AddressDomainLength+domainLen], data.Host)
		binary.BigEndian.PutUint16(buffer[AddressTypeByteLength+AddressDomainLength+domainLen:], data.Port)
		return AddressTypeByteLength + AddressDomainLength + domainLen + PortLength, nil
	default:
		return 0, AddressTypeIllegalErr
	}
}

func (data *ConnectType) UnMarshal(buffer []byte) (int, error) {
	if len(buffer) < AddressTypeByteLength {
		return 0, BufferLengthLimitErr
	}

	data.AddressType = AddressEnumType(buffer[0])
	switch data.AddressType {
	case AddressEnumTypeIpv4:
		if len(buffer) < (AddressTypeByteLength + AddressIpv4Length + PortLength) {
			return 0, BufferLengthLimitErr
		}

		data.Host = net.IP(buffer[AddressTypeByteLength : AddressTypeByteLength+AddressIpv4Length]).String()
		data.Port = binary.BigEndian.Uint16(buffer[AddressTypeByteLength+AddressIpv4Length : AddressTypeByteLength+AddressIpv4Length+PortLength])
		return AddressTypeByteLength + AddressIpv4Length + PortLength, nil
	case AddressEnumTypeIpv6:
		if len(buffer) < (AddressTypeByteLength + AddressIpv6Length + PortLength) {
			return 0, BufferLengthLimitErr
		}

		data.Host = net.IP(buffer[AddressTypeByteLength : AddressTypeByteLength+AddressIpv6Length]).String()
		data.Port = binary.BigEndian.Uint16(buffer[AddressTypeByteLength+AddressIpv6Length : AddressTypeByteLength+AddressIpv6Length+PortLength])
		return AddressTypeByteLength + AddressIpv6Length + PortLength, nil
	case AddressEnumTypeDomain:
		domainLen := int(buffer[AddressTypeByteLength])
		if len(buffer) < (AddressTypeByteLength + AddressDomainLength + domainLen + PortLength) {
			return 0, BufferLengthLimitErr
		}
		data.Host = string(buffer[AddressTypeByteLength+AddressDomainLength : AddressTypeByteLength+AddressDomainLength+domainLen])
		data.Port = binary.BigEndian.Uint16(buffer[AddressTypeByteLength+AddressDomainLength+domainLen : AddressTypeByteLength+AddressDomainLength+domainLen+PortLength])
		return AddressTypeByteLength + AddressDomainLength + domainLen + PortLength, nil
	default:
		return 0, errors.New("address type illegal")
	}
}
