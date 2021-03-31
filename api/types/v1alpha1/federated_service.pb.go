// Code generated by protoc-gen-go. DO NOT EDIT.
// source: federated_service.proto

package v1alpha1

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// FederatedService represents a federated service which can be discovered by
// a federated service mesh consumer.
type FederatedService struct {
	// A human readable name for the federated service.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// The description of the federated service.
	Description string `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	// Informative values for filtering purposes.
	Tags []string `protobuf:"bytes,3,rep,name=tags,proto3" json:"tags,omitempty"`
	// Informative KV pairs for filtering purposes.
	Labels map[string]string `protobuf:"bytes,4,rep,name=labels,proto3" json:"labels,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// REQUIRED. The FQDN (Fully Qualified Domain Name) to uniquely identify the
	// federated service.
	Fqdn string `protobuf:"bytes,5,opt,name=fqdn,proto3" json:"fqdn,omitempty"`
	// REQUIRED. The unique name that identifies the service within a federated
	// service mesh owner. The value of this field will be set as the SNI header
	// by the federated service mesh consumer. Each vendor may possibly have its
	// own SNI format, so this specification doesn't define a particular format
	// to use for this field.
	//
	// Following are some sample values for this field.
	// Example: foo.acme.com
	// Example: outbound_.8080_.v1_.foo.acme.com
	// Example: foo.acme.64a95d14-92f8-11e9-bc42-526af7764f64.com
	Id string `protobuf:"bytes,6,opt,name=id,proto3" json:"id,omitempty"`
	// REQUIRED. The SAN (Subject Alternative Name) to enable end-to-end security
	// for the federated service.
	San string `protobuf:"bytes,7,opt,name=san,proto3" json:"san,omitempty"`
	// REQUIRED. The protocols supported by the federated service.
	Protocols []string `protobuf:"bytes,8,rep,name=protocols,proto3" json:"protocols,omitempty"`
	// REQUIRED. The service mesh ingress endpoints that provide access to the
	// federated service.
	Endpoints            []*FederatedService_Endpoint `protobuf:"bytes,9,rep,name=endpoints,proto3" json:"endpoints,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                     `json:"-"`
	XXX_unrecognized     []byte                       `json:"-"`
	XXX_sizecache        int32                        `json:"-"`
}

func (m *FederatedService) Reset()         { *m = FederatedService{} }
func (m *FederatedService) String() string { return proto.CompactTextString(m) }
func (*FederatedService) ProtoMessage()    {}
func (*FederatedService) Descriptor() ([]byte, []int) {
	return fileDescriptor_0ca57f12e11e216c, []int{0}
}

func (m *FederatedService) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FederatedService.Unmarshal(m, b)
}
func (m *FederatedService) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FederatedService.Marshal(b, m, deterministic)
}
func (m *FederatedService) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FederatedService.Merge(m, src)
}
func (m *FederatedService) XXX_Size() int {
	return xxx_messageInfo_FederatedService.Size(m)
}
func (m *FederatedService) XXX_DiscardUnknown() {
	xxx_messageInfo_FederatedService.DiscardUnknown(m)
}

var xxx_messageInfo_FederatedService proto.InternalMessageInfo

func (m *FederatedService) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *FederatedService) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *FederatedService) GetTags() []string {
	if m != nil {
		return m.Tags
	}
	return nil
}

func (m *FederatedService) GetLabels() map[string]string {
	if m != nil {
		return m.Labels
	}
	return nil
}

func (m *FederatedService) GetFqdn() string {
	if m != nil {
		return m.Fqdn
	}
	return ""
}

func (m *FederatedService) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *FederatedService) GetSan() string {
	if m != nil {
		return m.San
	}
	return ""
}

func (m *FederatedService) GetProtocols() []string {
	if m != nil {
		return m.Protocols
	}
	return nil
}

func (m *FederatedService) GetEndpoints() []*FederatedService_Endpoint {
	if m != nil {
		return m.Endpoints
	}
	return nil
}

// Endpoint represents a service mesh ingress endpoint that provides access to
// a federated service over mTLS (mutual TLS).
type FederatedService_Endpoint struct {
	// REQUIRED. The service mesh ingress endpoint address (i.e. IP, CIDR, or a
	// domain name).
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	// REQUIRED. The port exposed by the service mesh ingress endpoint.
	Port                 uint32   `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FederatedService_Endpoint) Reset()         { *m = FederatedService_Endpoint{} }
func (m *FederatedService_Endpoint) String() string { return proto.CompactTextString(m) }
func (*FederatedService_Endpoint) ProtoMessage()    {}
func (*FederatedService_Endpoint) Descriptor() ([]byte, []int) {
	return fileDescriptor_0ca57f12e11e216c, []int{0, 1}
}

func (m *FederatedService_Endpoint) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FederatedService_Endpoint.Unmarshal(m, b)
}
func (m *FederatedService_Endpoint) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FederatedService_Endpoint.Marshal(b, m, deterministic)
}
func (m *FederatedService_Endpoint) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FederatedService_Endpoint.Merge(m, src)
}
func (m *FederatedService_Endpoint) XXX_Size() int {
	return xxx_messageInfo_FederatedService_Endpoint.Size(m)
}
func (m *FederatedService_Endpoint) XXX_DiscardUnknown() {
	xxx_messageInfo_FederatedService_Endpoint.DiscardUnknown(m)
}

var xxx_messageInfo_FederatedService_Endpoint proto.InternalMessageInfo

func (m *FederatedService_Endpoint) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *FederatedService_Endpoint) GetPort() uint32 {
	if m != nil {
		return m.Port
	}
	return 0
}

func init() {
	proto.RegisterType((*FederatedService)(nil), "federation.types.v1alpha1.FederatedService")
	proto.RegisterMapType((map[string]string)(nil), "federation.types.v1alpha1.FederatedService.LabelsEntry")
	proto.RegisterType((*FederatedService_Endpoint)(nil), "federation.types.v1alpha1.FederatedService.Endpoint")
}

func init() { proto.RegisterFile("federated_service.proto", fileDescriptor_0ca57f12e11e216c) }

var fileDescriptor_0ca57f12e11e216c = []byte{
	// 311 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x90, 0x4f, 0x4b, 0x33, 0x31,
	0x10, 0xc6, 0xe9, 0x6e, 0xff, 0xed, 0x94, 0xb7, 0x94, 0xf0, 0x82, 0xb1, 0x78, 0x58, 0x3c, 0xf5,
	0xb4, 0x52, 0x15, 0xac, 0x1e, 0x85, 0x7a, 0x12, 0x84, 0xf5, 0xe6, 0x45, 0xd2, 0x66, 0xaa, 0x8b,
	0x6b, 0x12, 0x93, 0x58, 0xe8, 0x57, 0xf4, 0x53, 0x49, 0x66, 0xb3, 0xb4, 0x08, 0x1e, 0xbc, 0x3d,
	0x99, 0xe1, 0x37, 0xfb, 0xdb, 0x07, 0x8e, 0x36, 0x28, 0xd1, 0x0a, 0x8f, 0xf2, 0xd9, 0xa1, 0xdd,
	0x56, 0x6b, 0x2c, 0x8c, 0xd5, 0x5e, 0xb3, 0xe3, 0xb8, 0xa8, 0xb4, 0x2a, 0xfc, 0xce, 0xa0, 0x2b,
	0xb6, 0x73, 0x51, 0x9b, 0x57, 0x31, 0x3f, 0xfd, 0x4a, 0x61, 0x72, 0xd7, 0x62, 0x8f, 0x0d, 0xc5,
	0x18, 0x74, 0x95, 0x78, 0x47, 0xde, 0xc9, 0x3b, 0xb3, 0xac, 0xa4, 0xcc, 0x72, 0x18, 0x49, 0x74,
	0x6b, 0x5b, 0x99, 0x70, 0x86, 0x27, 0xb4, 0x3a, 0x1c, 0x05, 0xca, 0x8b, 0x17, 0xc7, 0xd3, 0x3c,
	0x0d, 0x54, 0xc8, 0xec, 0x01, 0xfa, 0xb5, 0x58, 0x61, 0xed, 0x78, 0x37, 0x4f, 0x67, 0xa3, 0xf3,
	0xab, 0xe2, 0x57, 0x95, 0xe2, 0xa7, 0x46, 0x71, 0x4f, 0xe4, 0x52, 0x79, 0xbb, 0x2b, 0xe3, 0x99,
	0xf0, 0x91, 0xcd, 0x87, 0x54, 0xbc, 0xd7, 0xa8, 0x85, 0xcc, 0xc6, 0x90, 0x54, 0x92, 0xf7, 0x69,
	0x92, 0x54, 0x92, 0x4d, 0x20, 0x75, 0x42, 0xf1, 0x01, 0x0d, 0x42, 0x64, 0x27, 0x90, 0x51, 0x13,
	0x6b, 0x5d, 0x3b, 0x3e, 0x24, 0xbf, 0xfd, 0x80, 0x95, 0x90, 0xa1, 0x92, 0x46, 0x57, 0xca, 0x3b,
	0x9e, 0x91, 0xe7, 0xe5, 0x5f, 0x3c, 0x97, 0x11, 0x2e, 0xf7, 0x67, 0xa6, 0xd7, 0x30, 0x3a, 0xd0,
	0x0f, 0x4a, 0x6f, 0xb8, 0x8b, 0x85, 0x86, 0xc8, 0xfe, 0x43, 0x6f, 0x2b, 0xea, 0x4f, 0x8c, 0x4d,
	0x36, 0x8f, 0x9b, 0x64, 0xd1, 0x99, 0x2e, 0x60, 0xd8, 0x5e, 0x64, 0x1c, 0x06, 0x42, 0x4a, 0x8b,
	0xce, 0x45, 0xb6, 0x7d, 0x86, 0x22, 0x8c, 0xb6, 0x9e, 0xf0, 0x7f, 0x25, 0xe5, 0xdb, 0xc9, 0xd3,
	0x98, 0x5c, 0xcf, 0x5a, 0xd7, 0x55, 0x9f, 0xfe, 0xf2, 0xe2, 0x3b, 0x00, 0x00, 0xff, 0xff, 0xae,
	0x67, 0x55, 0x1d, 0x1b, 0x02, 0x00, 0x00,
}
