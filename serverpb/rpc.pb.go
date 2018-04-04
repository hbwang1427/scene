// Code generated by protoc-gen-go. DO NOT EDIT.
// source: rpc.proto

/*
Package serverpb is a generated protocol buffer package.

It is generated from these files:
	rpc.proto

It has these top-level messages:
	AuthRequest
	AuthResponse
	GeoPosition
	SignSpot
	PhotoPredictRequest
	PhotoPredictResponse
*/
package serverpb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type SignSpot_SignSpotType int32

const (
	SignSpot_Museum SignSpot_SignSpotType = 0
)

var SignSpot_SignSpotType_name = map[int32]string{
	0: "Museum",
}
var SignSpot_SignSpotType_value = map[string]int32{
	"Museum": 0,
}

func (x SignSpot_SignSpotType) String() string {
	return proto.EnumName(SignSpot_SignSpotType_name, int32(x))
}
func (SignSpot_SignSpotType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{3, 0} }

type PhotoPredictRequest_PhotoType int32

const (
	PhotoPredictRequest_PNG PhotoPredictRequest_PhotoType = 0
	PhotoPredictRequest_JPG PhotoPredictRequest_PhotoType = 1
)

var PhotoPredictRequest_PhotoType_name = map[int32]string{
	0: "PNG",
	1: "JPG",
}
var PhotoPredictRequest_PhotoType_value = map[string]int32{
	"PNG": 0,
	"JPG": 1,
}

func (x PhotoPredictRequest_PhotoType) String() string {
	return proto.EnumName(PhotoPredictRequest_PhotoType_name, int32(x))
}
func (PhotoPredictRequest_PhotoType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor0, []int{4, 0}
}

type AuthRequest struct {
	Name     string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Password string `protobuf:"bytes,2,opt,name=password" json:"password,omitempty"`
	Token    string `protobuf:"bytes,3,opt,name=token" json:"token,omitempty"`
}

func (m *AuthRequest) Reset()                    { *m = AuthRequest{} }
func (m *AuthRequest) String() string            { return proto.CompactTextString(m) }
func (*AuthRequest) ProtoMessage()               {}
func (*AuthRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *AuthRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *AuthRequest) GetPassword() string {
	if m != nil {
		return m.Password
	}
	return ""
}

func (m *AuthRequest) GetToken() string {
	if m != nil {
		return m.Token
	}
	return ""
}

type AuthResponse struct {
	RequireLogin bool   `protobuf:"varint,1,opt,name=require_login,json=requireLogin" json:"require_login,omitempty"`
	Token        string `protobuf:"bytes,2,opt,name=token" json:"token,omitempty"`
	Msg          string `protobuf:"bytes,3,opt,name=msg" json:"msg,omitempty"`
}

func (m *AuthResponse) Reset()                    { *m = AuthResponse{} }
func (m *AuthResponse) String() string            { return proto.CompactTextString(m) }
func (*AuthResponse) ProtoMessage()               {}
func (*AuthResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *AuthResponse) GetRequireLogin() bool {
	if m != nil {
		return m.RequireLogin
	}
	return false
}

func (m *AuthResponse) GetToken() string {
	if m != nil {
		return m.Token
	}
	return ""
}

func (m *AuthResponse) GetMsg() string {
	if m != nil {
		return m.Msg
	}
	return ""
}

type GeoPosition struct {
	Latitude  float64 `protobuf:"fixed64,1,opt,name=latitude" json:"latitude,omitempty"`
	Longitude float64 `protobuf:"fixed64,2,opt,name=longitude" json:"longitude,omitempty"`
}

func (m *GeoPosition) Reset()                    { *m = GeoPosition{} }
func (m *GeoPosition) String() string            { return proto.CompactTextString(m) }
func (*GeoPosition) ProtoMessage()               {}
func (*GeoPosition) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *GeoPosition) GetLatitude() float64 {
	if m != nil {
		return m.Latitude
	}
	return 0
}

func (m *GeoPosition) GetLongitude() float64 {
	if m != nil {
		return m.Longitude
	}
	return 0
}

type SignSpot struct {
	Id   uint64                `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Name string                `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	Type SignSpot_SignSpotType `protobuf:"varint,3,opt,name=type,enum=serverpb.SignSpot_SignSpotType" json:"type,omitempty"`
	Geo  *GeoPosition          `protobuf:"bytes,4,opt,name=geo" json:"geo,omitempty"`
}

func (m *SignSpot) Reset()                    { *m = SignSpot{} }
func (m *SignSpot) String() string            { return proto.CompactTextString(m) }
func (*SignSpot) ProtoMessage()               {}
func (*SignSpot) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *SignSpot) GetId() uint64 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *SignSpot) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *SignSpot) GetType() SignSpot_SignSpotType {
	if m != nil {
		return m.Type
	}
	return SignSpot_Museum
}

func (m *SignSpot) GetGeo() *GeoPosition {
	if m != nil {
		return m.Geo
	}
	return nil
}

type PhotoPredictRequest struct {
	Type          PhotoPredictRequest_PhotoType `protobuf:"varint,1,opt,name=type,enum=serverpb.PhotoPredictRequest_PhotoType" json:"type,omitempty"`
	Data          []byte                        `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	Base64Encoded bool                          `protobuf:"varint,3,opt,name=base64_encoded,json=base64Encoded" json:"base64_encoded,omitempty"`
	Geo           *GeoPosition                  `protobuf:"bytes,4,opt,name=geo" json:"geo,omitempty"`
	AcquireText   bool                          `protobuf:"varint,5,opt,name=acquire_text,json=acquireText" json:"acquire_text,omitempty"`
	AcquireAudio  bool                          `protobuf:"varint,6,opt,name=acquire_audio,json=acquireAudio" json:"acquire_audio,omitempty"`
	AcquireVideo  bool                          `protobuf:"varint,7,opt,name=acquire_video,json=acquireVideo" json:"acquire_video,omitempty"`
	MaxLimits     int32                         `protobuf:"varint,8,opt,name=max_limits,json=maxLimits" json:"max_limits,omitempty"`
	Language      string                        `protobuf:"bytes,9,opt,name=language" json:"language,omitempty"`
}

func (m *PhotoPredictRequest) Reset()                    { *m = PhotoPredictRequest{} }
func (m *PhotoPredictRequest) String() string            { return proto.CompactTextString(m) }
func (*PhotoPredictRequest) ProtoMessage()               {}
func (*PhotoPredictRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *PhotoPredictRequest) GetType() PhotoPredictRequest_PhotoType {
	if m != nil {
		return m.Type
	}
	return PhotoPredictRequest_PNG
}

func (m *PhotoPredictRequest) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *PhotoPredictRequest) GetBase64Encoded() bool {
	if m != nil {
		return m.Base64Encoded
	}
	return false
}

func (m *PhotoPredictRequest) GetGeo() *GeoPosition {
	if m != nil {
		return m.Geo
	}
	return nil
}

func (m *PhotoPredictRequest) GetAcquireText() bool {
	if m != nil {
		return m.AcquireText
	}
	return false
}

func (m *PhotoPredictRequest) GetAcquireAudio() bool {
	if m != nil {
		return m.AcquireAudio
	}
	return false
}

func (m *PhotoPredictRequest) GetAcquireVideo() bool {
	if m != nil {
		return m.AcquireVideo
	}
	return false
}

func (m *PhotoPredictRequest) GetMaxLimits() int32 {
	if m != nil {
		return m.MaxLimits
	}
	return 0
}

func (m *PhotoPredictRequest) GetLanguage() string {
	if m != nil {
		return m.Language
	}
	return ""
}

type PhotoPredictResponse struct {
	Results []*PhotoPredictResponse_Result `protobuf:"bytes,1,rep,name=results" json:"results,omitempty"`
}

func (m *PhotoPredictResponse) Reset()                    { *m = PhotoPredictResponse{} }
func (m *PhotoPredictResponse) String() string            { return proto.CompactTextString(m) }
func (*PhotoPredictResponse) ProtoMessage()               {}
func (*PhotoPredictResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *PhotoPredictResponse) GetResults() []*PhotoPredictResponse_Result {
	if m != nil {
		return m.Results
	}
	return nil
}

type PhotoPredictResponse_Result struct {
	Text        string `protobuf:"bytes,1,opt,name=text" json:"text,omitempty"`
	ImageUrl    string `protobuf:"bytes,2,opt,name=image_url,json=imageUrl" json:"image_url,omitempty"`
	AudioUrl    string `protobuf:"bytes,3,opt,name=audio_url,json=audioUrl" json:"audio_url,omitempty"`
	VideoUrl    string `protobuf:"bytes,4,opt,name=video_url,json=videoUrl" json:"video_url,omitempty"`
	ImageWidth  int32  `protobuf:"varint,5,opt,name=image_width,json=imageWidth" json:"image_width,omitempty"`
	ImageHeight int32  `protobuf:"varint,6,opt,name=image_height,json=imageHeight" json:"image_height,omitempty"`
	AudioSize   int32  `protobuf:"varint,7,opt,name=audio_size,json=audioSize" json:"audio_size,omitempty"`
	AudioLen    int32  `protobuf:"varint,8,opt,name=audio_len,json=audioLen" json:"audio_len,omitempty"`
}

func (m *PhotoPredictResponse_Result) Reset()                    { *m = PhotoPredictResponse_Result{} }
func (m *PhotoPredictResponse_Result) String() string            { return proto.CompactTextString(m) }
func (*PhotoPredictResponse_Result) ProtoMessage()               {}
func (*PhotoPredictResponse_Result) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5, 0} }

func (m *PhotoPredictResponse_Result) GetText() string {
	if m != nil {
		return m.Text
	}
	return ""
}

func (m *PhotoPredictResponse_Result) GetImageUrl() string {
	if m != nil {
		return m.ImageUrl
	}
	return ""
}

func (m *PhotoPredictResponse_Result) GetAudioUrl() string {
	if m != nil {
		return m.AudioUrl
	}
	return ""
}

func (m *PhotoPredictResponse_Result) GetVideoUrl() string {
	if m != nil {
		return m.VideoUrl
	}
	return ""
}

func (m *PhotoPredictResponse_Result) GetImageWidth() int32 {
	if m != nil {
		return m.ImageWidth
	}
	return 0
}

func (m *PhotoPredictResponse_Result) GetImageHeight() int32 {
	if m != nil {
		return m.ImageHeight
	}
	return 0
}

func (m *PhotoPredictResponse_Result) GetAudioSize() int32 {
	if m != nil {
		return m.AudioSize
	}
	return 0
}

func (m *PhotoPredictResponse_Result) GetAudioLen() int32 {
	if m != nil {
		return m.AudioLen
	}
	return 0
}

func init() {
	proto.RegisterType((*AuthRequest)(nil), "serverpb.AuthRequest")
	proto.RegisterType((*AuthResponse)(nil), "serverpb.AuthResponse")
	proto.RegisterType((*GeoPosition)(nil), "serverpb.GeoPosition")
	proto.RegisterType((*SignSpot)(nil), "serverpb.SignSpot")
	proto.RegisterType((*PhotoPredictRequest)(nil), "serverpb.PhotoPredictRequest")
	proto.RegisterType((*PhotoPredictResponse)(nil), "serverpb.PhotoPredictResponse")
	proto.RegisterType((*PhotoPredictResponse_Result)(nil), "serverpb.PhotoPredictResponse.Result")
	proto.RegisterEnum("serverpb.SignSpot_SignSpotType", SignSpot_SignSpotType_name, SignSpot_SignSpotType_value)
	proto.RegisterEnum("serverpb.PhotoPredictRequest_PhotoType", PhotoPredictRequest_PhotoType_name, PhotoPredictRequest_PhotoType_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Auth service

type AuthClient interface {
	Authenticate(ctx context.Context, in *AuthRequest, opts ...grpc.CallOption) (*AuthResponse, error)
}

type authClient struct {
	cc *grpc.ClientConn
}

func NewAuthClient(cc *grpc.ClientConn) AuthClient {
	return &authClient{cc}
}

func (c *authClient) Authenticate(ctx context.Context, in *AuthRequest, opts ...grpc.CallOption) (*AuthResponse, error) {
	out := new(AuthResponse)
	err := grpc.Invoke(ctx, "/serverpb.Auth/Authenticate", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Auth service

type AuthServer interface {
	Authenticate(context.Context, *AuthRequest) (*AuthResponse, error)
}

func RegisterAuthServer(s *grpc.Server, srv AuthServer) {
	s.RegisterService(&_Auth_serviceDesc, srv)
}

func _Auth_Authenticate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AuthRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServer).Authenticate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/serverpb.Auth/Authenticate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServer).Authenticate(ctx, req.(*AuthRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Auth_serviceDesc = grpc.ServiceDesc{
	ServiceName: "serverpb.Auth",
	HandlerType: (*AuthServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Authenticate",
			Handler:    _Auth_Authenticate_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rpc.proto",
}

// Client API for Predict service

type PredictClient interface {
	PredictPhoto(ctx context.Context, in *PhotoPredictRequest, opts ...grpc.CallOption) (*PhotoPredictResponse, error)
}

type predictClient struct {
	cc *grpc.ClientConn
}

func NewPredictClient(cc *grpc.ClientConn) PredictClient {
	return &predictClient{cc}
}

func (c *predictClient) PredictPhoto(ctx context.Context, in *PhotoPredictRequest, opts ...grpc.CallOption) (*PhotoPredictResponse, error) {
	out := new(PhotoPredictResponse)
	err := grpc.Invoke(ctx, "/serverpb.Predict/PredictPhoto", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Predict service

type PredictServer interface {
	PredictPhoto(context.Context, *PhotoPredictRequest) (*PhotoPredictResponse, error)
}

func RegisterPredictServer(s *grpc.Server, srv PredictServer) {
	s.RegisterService(&_Predict_serviceDesc, srv)
}

func _Predict_PredictPhoto_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PhotoPredictRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PredictServer).PredictPhoto(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/serverpb.Predict/PredictPhoto",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PredictServer).PredictPhoto(ctx, req.(*PhotoPredictRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Predict_serviceDesc = grpc.ServiceDesc{
	ServiceName: "serverpb.Predict",
	HandlerType: (*PredictServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PredictPhoto",
			Handler:    _Predict_PredictPhoto_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rpc.proto",
}

func init() { proto.RegisterFile("rpc.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 696 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x94, 0xcf, 0x6f, 0xd3, 0x4a,
	0x10, 0xc7, 0x63, 0xe7, 0xa7, 0x27, 0x69, 0x14, 0xed, 0xeb, 0x7b, 0xb2, 0xf2, 0x28, 0x0d, 0x46,
	0xa5, 0x39, 0xe5, 0x90, 0x22, 0x2e, 0x1c, 0xaa, 0x56, 0x42, 0x41, 0xa8, 0x40, 0xe4, 0x14, 0x90,
	0xe0, 0x10, 0xb9, 0xf1, 0xc8, 0x59, 0x61, 0x7b, 0x5d, 0x7b, 0xdd, 0xa6, 0xfd, 0x73, 0x38, 0xf1,
	0xcf, 0xf0, 0xd7, 0x70, 0x46, 0x42, 0x3b, 0xeb, 0x24, 0x2e, 0x82, 0x8a, 0xdb, 0xec, 0x67, 0x66,
	0x67, 0x67, 0xbf, 0x33, 0xbb, 0x60, 0xa5, 0xc9, 0x62, 0x94, 0xa4, 0x42, 0x0a, 0xd6, 0xca, 0x30,
	0xbd, 0xc2, 0x34, 0xb9, 0x70, 0x66, 0xd0, 0x3e, 0xc9, 0xe5, 0xd2, 0xc5, 0xcb, 0x1c, 0x33, 0xc9,
	0x18, 0xd4, 0x62, 0x2f, 0x42, 0xdb, 0x18, 0x18, 0x43, 0xcb, 0x25, 0x9b, 0xf5, 0xa1, 0x95, 0x78,
	0x59, 0x76, 0x2d, 0x52, 0xdf, 0x36, 0x89, 0x6f, 0xd6, 0x6c, 0x17, 0xea, 0x52, 0x7c, 0xc6, 0xd8,
	0xae, 0x92, 0x43, 0x2f, 0x9c, 0x4f, 0xd0, 0xd1, 0x49, 0xb3, 0x44, 0xc4, 0x19, 0xb2, 0xc7, 0xb0,
	0x93, 0xe2, 0x65, 0xce, 0x53, 0x9c, 0x87, 0x22, 0xe0, 0x31, 0xa5, 0x6f, 0xb9, 0x9d, 0x02, 0x9e,
	0x29, 0xb6, 0x4d, 0x65, 0x96, 0x52, 0xb1, 0x1e, 0x54, 0xa3, 0x2c, 0x28, 0xd2, 0x2b, 0xd3, 0x99,
	0x40, 0x7b, 0x82, 0x62, 0x2a, 0x32, 0x2e, 0xb9, 0x88, 0x55, 0x75, 0xa1, 0x27, 0xb9, 0xcc, 0x7d,
	0x5d, 0xb5, 0xe1, 0x6e, 0xd6, 0xec, 0x01, 0x58, 0xa1, 0x88, 0x03, 0xed, 0x34, 0xc9, 0xb9, 0x05,
	0xce, 0x57, 0x03, 0x5a, 0x33, 0x1e, 0xc4, 0xb3, 0x44, 0x48, 0xd6, 0x05, 0x93, 0xfb, 0x94, 0xa0,
	0xe6, 0x9a, 0xdc, 0xdf, 0x08, 0x61, 0x96, 0x84, 0x38, 0x82, 0x9a, 0xbc, 0x49, 0x90, 0x8a, 0xe9,
	0x8e, 0xf7, 0x47, 0x6b, 0x11, 0x47, 0xeb, 0x2c, 0x1b, 0xe3, 0xfc, 0x26, 0x41, 0x97, 0x82, 0xd9,
	0x21, 0x54, 0x03, 0x14, 0x76, 0x6d, 0x60, 0x0c, 0xdb, 0xe3, 0x7f, 0xb7, 0x7b, 0x4a, 0x77, 0x70,
	0x55, 0x84, 0xd3, 0x87, 0x4e, 0x79, 0x3b, 0x03, 0x68, 0xbc, 0xce, 0x33, 0xcc, 0xa3, 0x5e, 0xc5,
	0xf9, 0x61, 0xc2, 0x3f, 0xd3, 0xa5, 0x90, 0x62, 0x9a, 0xa2, 0xcf, 0x17, 0x72, 0xdd, 0xae, 0xe7,
	0x45, 0x45, 0x06, 0x55, 0x74, 0xb8, 0xcd, 0xfe, 0x9b, 0x60, 0xcd, 0x4a, 0x95, 0x31, 0xa8, 0xf9,
	0x9e, 0xf4, 0xe8, 0x8a, 0x1d, 0x97, 0x6c, 0x76, 0x00, 0xdd, 0x0b, 0x2f, 0xc3, 0x67, 0x4f, 0xe7,
	0x18, 0x2f, 0x84, 0x8f, 0x3e, 0x5d, 0xb6, 0xe5, 0xee, 0x68, 0xfa, 0x42, 0xc3, 0xbf, 0xbe, 0x14,
	0x7b, 0x04, 0x1d, 0x6f, 0xa1, 0x3b, 0x2f, 0x71, 0x25, 0xed, 0x3a, 0x65, 0x6b, 0x17, 0xec, 0x1c,
	0x57, 0x52, 0x0d, 0xc7, 0x3a, 0xc4, 0xcb, 0x7d, 0x2e, 0xec, 0x86, 0x1e, 0x8e, 0x02, 0x9e, 0x28,
	0x56, 0x0e, 0xba, 0xe2, 0x3e, 0x0a, 0xbb, 0x79, 0x27, 0xe8, 0xbd, 0x62, 0x6c, 0x0f, 0x20, 0xf2,
	0x56, 0xf3, 0x90, 0x47, 0x5c, 0x66, 0x76, 0x6b, 0x60, 0x0c, 0xeb, 0xae, 0x15, 0x79, 0xab, 0x33,
	0x02, 0x7a, 0x52, 0xe2, 0x20, 0xf7, 0x02, 0xb4, 0x2d, 0x3d, 0xc7, 0xeb, 0xb5, 0xb3, 0x07, 0xd6,
	0x46, 0x1e, 0xd6, 0x84, 0xea, 0xf4, 0xcd, 0xa4, 0x57, 0x51, 0xc6, 0xab, 0xe9, 0xa4, 0x67, 0x38,
	0xdf, 0x4c, 0xd8, 0xbd, 0x2b, 0x69, 0x31, 0xd9, 0xc7, 0xd0, 0x4c, 0x31, 0xcb, 0x43, 0x99, 0xd9,
	0xc6, 0xa0, 0x3a, 0x6c, 0x8f, 0x0f, 0xfe, 0xd4, 0x03, 0xbd, 0x61, 0xe4, 0x52, 0xb4, 0xbb, 0xde,
	0xd5, 0xff, 0x6e, 0x40, 0x43, 0x33, 0xd5, 0x0f, 0xd2, 0xa8, 0x78, 0x7b, 0xca, 0x66, 0xff, 0x83,
	0xc5, 0x23, 0x2f, 0xc0, 0x79, 0x9e, 0x86, 0xeb, 0xc7, 0x47, 0xe0, 0x5d, 0x1a, 0x2a, 0x27, 0x29,
	0x46, 0x4e, 0xfd, 0x42, 0x5a, 0x04, 0x0a, 0x27, 0x29, 0x45, 0xce, 0x9a, 0x76, 0x12, 0x50, 0xce,
	0x7d, 0x68, 0xeb, 0xb4, 0xd7, 0xdc, 0x97, 0x4b, 0xea, 0x4a, 0xdd, 0x05, 0x42, 0x1f, 0x14, 0x51,
	0x7d, 0xd3, 0x01, 0x4b, 0xe4, 0xc1, 0x52, 0x52, 0x4f, 0xea, 0xae, 0xde, 0xf4, 0x92, 0x90, 0x52,
	0x5b, 0x9f, 0x9e, 0xf1, 0x5b, 0xa4, 0x7e, 0xd4, 0x5d, 0x5d, 0xcf, 0x8c, 0xdf, 0xe2, 0xb6, 0xb8,
	0x10, 0xe3, 0xa2, 0x17, 0xba, 0xb8, 0x33, 0x8c, 0xc7, 0x13, 0xa8, 0xa9, 0x0f, 0x82, 0x1d, 0xeb,
	0x8f, 0x02, 0x63, 0xc9, 0x17, 0x9e, 0x44, 0x56, 0x1a, 0xa5, 0xd2, 0xaf, 0xd4, 0xff, 0xef, 0x57,
	0xac, 0xc5, 0x74, 0x2a, 0xe3, 0x8f, 0xd0, 0x2c, 0x14, 0x66, 0x6f, 0xa1, 0x53, 0x98, 0x24, 0x3c,
	0xdb, 0xbb, 0xf7, 0x35, 0xf4, 0x1f, 0xde, 0xdf, 0x28, 0xa7, 0x72, 0xfa, 0x04, 0xba, 0x0b, 0x11,
	0x8d, 0x3c, 0x2e, 0x45, 0x9e, 0x8e, 0xd2, 0x64, 0x71, 0x6a, 0xa9, 0xd3, 0xa7, 0xea, 0x07, 0x9d,
	0x1a, 0x5f, 0xcc, 0x86, 0xf6, 0x5c, 0x34, 0xe8, 0x4f, 0x3d, 0xfa, 0x19, 0x00, 0x00, 0xff, 0xff,
	0x2e, 0x5b, 0x9d, 0xfe, 0x60, 0x05, 0x00, 0x00,
}
