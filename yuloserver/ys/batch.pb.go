// Code generated by protoc-gen-go. DO NOT EDIT.
// source: batch.proto

package ys

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

type Point struct {
	Lon                  float64  `protobuf:"fixed64,1,opt,name=lon,proto3" json:"lon,omitempty"`
	Lat                  float64  `protobuf:"fixed64,2,opt,name=lat,proto3" json:"lat,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Point) Reset()         { *m = Point{} }
func (m *Point) String() string { return proto.CompactTextString(m) }
func (*Point) ProtoMessage()    {}
func (*Point) Descriptor() ([]byte, []int) {
	return fileDescriptor_905061dbf2994c5e, []int{0}
}

func (m *Point) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Point.Unmarshal(m, b)
}
func (m *Point) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Point.Marshal(b, m, deterministic)
}
func (m *Point) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Point.Merge(m, src)
}
func (m *Point) XXX_Size() int {
	return xxx_messageInfo_Point.Size(m)
}
func (m *Point) XXX_DiscardUnknown() {
	xxx_messageInfo_Point.DiscardUnknown(m)
}

var xxx_messageInfo_Point proto.InternalMessageInfo

func (m *Point) GetLon() float64 {
	if m != nil {
		return m.Lon
	}
	return 0
}

func (m *Point) GetLat() float64 {
	if m != nil {
		return m.Lat
	}
	return 0
}

type Observation struct {
	Datetime             int64    `protobuf:"varint,1,opt,name=datetime,proto3" json:"datetime,omitempty"`
	Location             *Point   `protobuf:"bytes,2,opt,name=location,proto3" json:"location,omitempty"`
	Speed                int64    `protobuf:"varint,3,opt,name=speed,proto3" json:"speed,omitempty"`
	Azimuth              int64    `protobuf:"varint,4,opt,name=azimuth,proto3" json:"azimuth,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Observation) Reset()         { *m = Observation{} }
func (m *Observation) String() string { return proto.CompactTextString(m) }
func (*Observation) ProtoMessage()    {}
func (*Observation) Descriptor() ([]byte, []int) {
	return fileDescriptor_905061dbf2994c5e, []int{1}
}

func (m *Observation) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Observation.Unmarshal(m, b)
}
func (m *Observation) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Observation.Marshal(b, m, deterministic)
}
func (m *Observation) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Observation.Merge(m, src)
}
func (m *Observation) XXX_Size() int {
	return xxx_messageInfo_Observation.Size(m)
}
func (m *Observation) XXX_DiscardUnknown() {
	xxx_messageInfo_Observation.DiscardUnknown(m)
}

var xxx_messageInfo_Observation proto.InternalMessageInfo

func (m *Observation) GetDatetime() int64 {
	if m != nil {
		return m.Datetime
	}
	return 0
}

func (m *Observation) GetLocation() *Point {
	if m != nil {
		return m.Location
	}
	return nil
}

func (m *Observation) GetSpeed() int64 {
	if m != nil {
		return m.Speed
	}
	return -1
}

func (m *Observation) GetAzimuth() int64 {
	if m != nil {
		return m.Azimuth
	}
	return 0
}

type Vehicle struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Firm                 string   `protobuf:"bytes,2,opt,name=firm,proto3" json:"firm,omitempty"`
	Type                 string   `protobuf:"bytes,3,opt,name=type,proto3" json:"type,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Vehicle) Reset()         { *m = Vehicle{} }
func (m *Vehicle) String() string { return proto.CompactTextString(m) }
func (*Vehicle) ProtoMessage()    {}
func (*Vehicle) Descriptor() ([]byte, []int) {
	return fileDescriptor_905061dbf2994c5e, []int{2}
}

func (m *Vehicle) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Vehicle.Unmarshal(m, b)
}
func (m *Vehicle) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Vehicle.Marshal(b, m, deterministic)
}
func (m *Vehicle) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Vehicle.Merge(m, src)
}
func (m *Vehicle) XXX_Size() int {
	return xxx_messageInfo_Vehicle.Size(m)
}
func (m *Vehicle) XXX_DiscardUnknown() {
	xxx_messageInfo_Vehicle.DiscardUnknown(m)
}

var xxx_messageInfo_Vehicle proto.InternalMessageInfo

func (m *Vehicle) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Vehicle) GetFirm() string {
	if m != nil {
		return m.Firm
	}
	return ""
}

func (m *Vehicle) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

type VehicleTrace struct {
	Vehicle              *Vehicle       `protobuf:"bytes,1,opt,name=vehicle,proto3" json:"vehicle,omitempty"`
	Observations         []*Observation `protobuf:"bytes,2,rep,name=observations,proto3" json:"observations,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *VehicleTrace) Reset()         { *m = VehicleTrace{} }
func (m *VehicleTrace) String() string { return proto.CompactTextString(m) }
func (*VehicleTrace) ProtoMessage()    {}
func (*VehicleTrace) Descriptor() ([]byte, []int) {
	return fileDescriptor_905061dbf2994c5e, []int{3}
}

func (m *VehicleTrace) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_VehicleTrace.Unmarshal(m, b)
}
func (m *VehicleTrace) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_VehicleTrace.Marshal(b, m, deterministic)
}
func (m *VehicleTrace) XXX_Merge(src proto.Message) {
	xxx_messageInfo_VehicleTrace.Merge(m, src)
}
func (m *VehicleTrace) XXX_Size() int {
	return xxx_messageInfo_VehicleTrace.Size(m)
}
func (m *VehicleTrace) XXX_DiscardUnknown() {
	xxx_messageInfo_VehicleTrace.DiscardUnknown(m)
}

var xxx_messageInfo_VehicleTrace proto.InternalMessageInfo

func (m *VehicleTrace) GetVehicle() *Vehicle {
	if m != nil {
		return m.Vehicle
	}
	return nil
}

func (m *VehicleTrace) GetObservations() []*Observation {
	if m != nil {
		return m.Observations
	}
	return nil
}

type Batch struct {
	Traces               []*VehicleTrace `protobuf:"bytes,1,rep,name=traces,proto3" json:"traces,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *Batch) Reset()         { *m = Batch{} }
func (m *Batch) String() string { return proto.CompactTextString(m) }
func (*Batch) ProtoMessage()    {}
func (*Batch) Descriptor() ([]byte, []int) {
	return fileDescriptor_905061dbf2994c5e, []int{4}
}

func (m *Batch) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Batch.Unmarshal(m, b)
}
func (m *Batch) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Batch.Marshal(b, m, deterministic)
}
func (m *Batch) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Batch.Merge(m, src)
}
func (m *Batch) XXX_Size() int {
	return xxx_messageInfo_Batch.Size(m)
}
func (m *Batch) XXX_DiscardUnknown() {
	xxx_messageInfo_Batch.DiscardUnknown(m)
}

var xxx_messageInfo_Batch proto.InternalMessageInfo

func (m *Batch) GetTraces() []*VehicleTrace {
	if m != nil {
		return m.Traces
	}
	return nil
}

func init() {
	proto.RegisterType((*Point)(nil), "yulo.Point")
	proto.RegisterType((*Observation)(nil), "yulo.Observation")
	proto.RegisterType((*Vehicle)(nil), "yulo.Vehicle")
	proto.RegisterType((*VehicleTrace)(nil), "yulo.VehicleTrace")
	proto.RegisterType((*Batch)(nil), "yulo.Batch")
}

func init() { proto.RegisterFile("batch.proto", fileDescriptor_905061dbf2994c5e) }

var fileDescriptor_905061dbf2994c5e = []byte{
	// 288 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x51, 0xcb, 0x4e, 0xc3, 0x30,
	0x10, 0x54, 0x1e, 0x6d, 0x9a, 0x75, 0x41, 0xb0, 0xe2, 0x60, 0x71, 0xaa, 0x72, 0xa1, 0x02, 0x29,
	0x87, 0x54, 0x7c, 0x00, 0xfd, 0x01, 0x90, 0x85, 0x38, 0x70, 0x73, 0x12, 0xa3, 0x58, 0x4a, 0xe2,
	0x28, 0x71, 0x2b, 0x85, 0x23, 0x5f, 0x8e, 0xbc, 0x49, 0x0b, 0xbd, 0xed, 0xcc, 0xce, 0xce, 0xac,
	0xd7, 0xc0, 0x72, 0x69, 0x8b, 0x2a, 0xed, 0x7a, 0x63, 0x0d, 0x86, 0xe3, 0xa1, 0x36, 0xc9, 0x13,
	0x2c, 0xde, 0x8c, 0x6e, 0x2d, 0xde, 0x40, 0x50, 0x9b, 0x96, 0x7b, 0x1b, 0x6f, 0xeb, 0x09, 0x57,
	0x12, 0x23, 0x2d, 0xf7, 0x67, 0x46, 0xda, 0xe4, 0xc7, 0x03, 0xf6, 0x9a, 0x0f, 0xaa, 0x3f, 0x4a,
	0xab, 0x4d, 0x8b, 0xf7, 0xb0, 0x2a, 0xa5, 0x55, 0x56, 0x37, 0x8a, 0x06, 0x03, 0x71, 0xc6, 0xf8,
	0x00, 0xab, 0xda, 0x14, 0xa4, 0x23, 0x0b, 0x96, 0xb1, 0xd4, 0x25, 0xa6, 0x14, 0x27, 0xce, 0x4d,
	0xbc, 0x83, 0xc5, 0xd0, 0x29, 0x55, 0xf2, 0x80, 0x1c, 0x26, 0x80, 0x1c, 0x22, 0xf9, 0xad, 0x9b,
	0x83, 0xad, 0x78, 0x48, 0xfc, 0x09, 0x26, 0x2f, 0x10, 0x7d, 0xa8, 0x4a, 0x17, 0xb5, 0xc2, 0x6b,
	0xf0, 0x75, 0x49, 0xc9, 0xb1, 0xf0, 0x75, 0x89, 0x08, 0xe1, 0x97, 0xee, 0x1b, 0xca, 0x8b, 0x05,
	0xd5, 0x8e, 0xb3, 0x63, 0xa7, 0xc8, 0x3d, 0x16, 0x54, 0x27, 0x2d, 0xac, 0x67, 0x8b, 0xf7, 0x5e,
	0x16, 0x6e, 0xd7, 0xe8, 0x38, 0x61, 0x32, 0x63, 0xd9, 0xd5, 0xb4, 0xea, 0x2c, 0x12, 0xa7, 0x2e,
	0x3e, 0xc3, 0xda, 0xfc, 0xbd, 0x7f, 0xe0, 0xfe, 0x26, 0xd8, 0xb2, 0xec, 0x76, 0x52, 0xff, 0xbb,
	0x8c, 0xb8, 0x90, 0x25, 0x3b, 0x58, 0xec, 0xdd, 0xe5, 0xf1, 0x11, 0x96, 0xd6, 0x25, 0x0e, 0xdc,
	0xa3, 0x49, 0xbc, 0xc8, 0xa1, 0x65, 0xc4, 0xac, 0xd8, 0x87, 0x9f, 0xfe, 0x38, 0xe4, 0x4b, 0xfa,
	0xac, 0xdd, 0x6f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x57, 0x1c, 0x91, 0x01, 0xbb, 0x01, 0x00, 0x00,
}
