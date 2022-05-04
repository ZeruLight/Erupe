// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.12.2
// source: google/cloud/orchestration/airflow/service/v1beta1/image_versions.proto

package service

import (
	context "context"
	reflect "reflect"
	sync "sync"

	_ "google.golang.org/genproto/googleapis/api/annotations"
	date "google.golang.org/genproto/googleapis/type/date"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// List ImageVersions in a project and location.
type ListImageVersionsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// List ImageVersions in the given project and location, in the form:
	// "projects/{projectId}/locations/{locationId}"
	Parent string `protobuf:"bytes,1,opt,name=parent,proto3" json:"parent,omitempty"`
	// The maximum number of image_versions to return.
	PageSize int32 `protobuf:"varint,2,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
	// The next_page_token value returned from a previous List request, if any.
	PageToken string `protobuf:"bytes,3,opt,name=page_token,json=pageToken,proto3" json:"page_token,omitempty"`
	// Whether or not image versions from old releases should be included.
	IncludePastReleases bool `protobuf:"varint,4,opt,name=include_past_releases,json=includePastReleases,proto3" json:"include_past_releases,omitempty"`
}

func (x *ListImageVersionsRequest) Reset() {
	*x = ListImageVersionsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListImageVersionsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListImageVersionsRequest) ProtoMessage() {}

func (x *ListImageVersionsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListImageVersionsRequest.ProtoReflect.Descriptor instead.
func (*ListImageVersionsRequest) Descriptor() ([]byte, []int) {
	return file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_rawDescGZIP(), []int{0}
}

func (x *ListImageVersionsRequest) GetParent() string {
	if x != nil {
		return x.Parent
	}
	return ""
}

func (x *ListImageVersionsRequest) GetPageSize() int32 {
	if x != nil {
		return x.PageSize
	}
	return 0
}

func (x *ListImageVersionsRequest) GetPageToken() string {
	if x != nil {
		return x.PageToken
	}
	return ""
}

func (x *ListImageVersionsRequest) GetIncludePastReleases() bool {
	if x != nil {
		return x.IncludePastReleases
	}
	return false
}

// The ImageVersions in a project and location.
type ListImageVersionsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The list of supported ImageVersions in a location.
	ImageVersions []*ImageVersion `protobuf:"bytes,1,rep,name=image_versions,json=imageVersions,proto3" json:"image_versions,omitempty"`
	// The page token used to query for the next page if one exists.
	NextPageToken string `protobuf:"bytes,2,opt,name=next_page_token,json=nextPageToken,proto3" json:"next_page_token,omitempty"`
}

func (x *ListImageVersionsResponse) Reset() {
	*x = ListImageVersionsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListImageVersionsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListImageVersionsResponse) ProtoMessage() {}

func (x *ListImageVersionsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListImageVersionsResponse.ProtoReflect.Descriptor instead.
func (*ListImageVersionsResponse) Descriptor() ([]byte, []int) {
	return file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_rawDescGZIP(), []int{1}
}

func (x *ListImageVersionsResponse) GetImageVersions() []*ImageVersion {
	if x != nil {
		return x.ImageVersions
	}
	return nil
}

func (x *ListImageVersionsResponse) GetNextPageToken() string {
	if x != nil {
		return x.NextPageToken
	}
	return ""
}

// Image Version information
type ImageVersion struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The string identifier of the ImageVersion, in the form:
	// "composer-x.y.z-airflow-a.b(.c)"
	ImageVersionId string `protobuf:"bytes,1,opt,name=image_version_id,json=imageVersionId,proto3" json:"image_version_id,omitempty"`
	// Whether this is the default ImageVersion used by Composer during
	// environment creation if no input ImageVersion is specified.
	IsDefault bool `protobuf:"varint,2,opt,name=is_default,json=isDefault,proto3" json:"is_default,omitempty"`
	// supported python versions
	SupportedPythonVersions []string `protobuf:"bytes,3,rep,name=supported_python_versions,json=supportedPythonVersions,proto3" json:"supported_python_versions,omitempty"`
	// The date of the version release.
	ReleaseDate *date.Date `protobuf:"bytes,4,opt,name=release_date,json=releaseDate,proto3" json:"release_date,omitempty"`
	// Whether it is impossible to create an environment with the image version.
	CreationDisabled bool `protobuf:"varint,5,opt,name=creation_disabled,json=creationDisabled,proto3" json:"creation_disabled,omitempty"`
	// Whether it is impossible to upgrade an environment running with the image
	// version.
	UpgradeDisabled bool `protobuf:"varint,6,opt,name=upgrade_disabled,json=upgradeDisabled,proto3" json:"upgrade_disabled,omitempty"`
}

func (x *ImageVersion) Reset() {
	*x = ImageVersion{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ImageVersion) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ImageVersion) ProtoMessage() {}

func (x *ImageVersion) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ImageVersion.ProtoReflect.Descriptor instead.
func (*ImageVersion) Descriptor() ([]byte, []int) {
	return file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_rawDescGZIP(), []int{2}
}

func (x *ImageVersion) GetImageVersionId() string {
	if x != nil {
		return x.ImageVersionId
	}
	return ""
}

func (x *ImageVersion) GetIsDefault() bool {
	if x != nil {
		return x.IsDefault
	}
	return false
}

func (x *ImageVersion) GetSupportedPythonVersions() []string {
	if x != nil {
		return x.SupportedPythonVersions
	}
	return nil
}

func (x *ImageVersion) GetReleaseDate() *date.Date {
	if x != nil {
		return x.ReleaseDate
	}
	return nil
}

func (x *ImageVersion) GetCreationDisabled() bool {
	if x != nil {
		return x.CreationDisabled
	}
	return false
}

func (x *ImageVersion) GetUpgradeDisabled() bool {
	if x != nil {
		return x.UpgradeDisabled
	}
	return false
}

var File_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto protoreflect.FileDescriptor

var file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_rawDesc = []byte{
	0x0a, 0x47, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2f, 0x6f,
	0x72, 0x63, 0x68, 0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x61, 0x69, 0x72,
	0x66, 0x6c, 0x6f, 0x77, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x76, 0x31, 0x62,
	0x65, 0x74, 0x61, 0x31, 0x2f, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x32, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6f, 0x72, 0x63, 0x68, 0x65, 0x73, 0x74, 0x72,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x61, 0x69, 0x72, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x1a, 0x1c, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x17, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x16, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x74, 0x79, 0x70,
	0x65, 0x2f, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xa2, 0x01, 0x0a,
	0x18, 0x4c, 0x69, 0x73, 0x74, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x61, 0x72,
	0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x61, 0x72, 0x65, 0x6e,
	0x74, 0x12, 0x1b, 0x0a, 0x09, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x70, 0x61, 0x67, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x1d,
	0x0a, 0x0a, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x09, 0x70, 0x61, 0x67, 0x65, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x32, 0x0a,
	0x15, 0x69, 0x6e, 0x63, 0x6c, 0x75, 0x64, 0x65, 0x5f, 0x70, 0x61, 0x73, 0x74, 0x5f, 0x72, 0x65,
	0x6c, 0x65, 0x61, 0x73, 0x65, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x13, 0x69, 0x6e,
	0x63, 0x6c, 0x75, 0x64, 0x65, 0x50, 0x61, 0x73, 0x74, 0x52, 0x65, 0x6c, 0x65, 0x61, 0x73, 0x65,
	0x73, 0x22, 0xac, 0x01, 0x0a, 0x19, 0x4c, 0x69, 0x73, 0x74, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x56,
	0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x67, 0x0a, 0x0e, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x40, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6f, 0x72, 0x63, 0x68, 0x65, 0x73, 0x74, 0x72, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x61, 0x69, 0x72, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x49, 0x6d, 0x61,
	0x67, 0x65, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x0d, 0x69, 0x6d, 0x61, 0x67, 0x65,
	0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x26, 0x0a, 0x0f, 0x6e, 0x65, 0x78, 0x74,
	0x5f, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0d, 0x6e, 0x65, 0x78, 0x74, 0x50, 0x61, 0x67, 0x65, 0x54, 0x6f, 0x6b, 0x65, 0x6e,
	0x22, 0xa1, 0x02, 0x0a, 0x0c, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x12, 0x28, 0x0a, 0x10, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x69, 0x6d, 0x61,
	0x67, 0x65, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x12, 0x1d, 0x0a, 0x0a, 0x69,
	0x73, 0x5f, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x09, 0x69, 0x73, 0x44, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x12, 0x3a, 0x0a, 0x19, 0x73, 0x75,
	0x70, 0x70, 0x6f, 0x72, 0x74, 0x65, 0x64, 0x5f, 0x70, 0x79, 0x74, 0x68, 0x6f, 0x6e, 0x5f, 0x76,
	0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x17, 0x73,
	0x75, 0x70, 0x70, 0x6f, 0x72, 0x74, 0x65, 0x64, 0x50, 0x79, 0x74, 0x68, 0x6f, 0x6e, 0x56, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x34, 0x0a, 0x0c, 0x72, 0x65, 0x6c, 0x65, 0x61, 0x73,
	0x65, 0x5f, 0x64, 0x61, 0x74, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x2e, 0x44, 0x61, 0x74, 0x65, 0x52,
	0x0b, 0x72, 0x65, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x44, 0x61, 0x74, 0x65, 0x12, 0x2b, 0x0a, 0x11,
	0x63, 0x72, 0x65, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65,
	0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x10, 0x63, 0x72, 0x65, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x44, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x12, 0x29, 0x0a, 0x10, 0x75, 0x70, 0x67,
	0x72, 0x61, 0x64, 0x65, 0x5f, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x08, 0x52, 0x0f, 0x75, 0x70, 0x67, 0x72, 0x61, 0x64, 0x65, 0x44, 0x69, 0x73, 0x61,
	0x62, 0x6c, 0x65, 0x64, 0x32, 0xd8, 0x02, 0x0a, 0x0d, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x56, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0xf9, 0x01, 0x0a, 0x11, 0x4c, 0x69, 0x73, 0x74, 0x49,
	0x6d, 0x61, 0x67, 0x65, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x4c, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6f, 0x72, 0x63, 0x68,
	0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x61, 0x69, 0x72, 0x66, 0x6c, 0x6f,
	0x77, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61,
	0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x56, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x4d, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6f, 0x72, 0x63, 0x68, 0x65, 0x73,
	0x74, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x61, 0x69, 0x72, 0x66, 0x6c, 0x6f, 0x77, 0x2e,
	0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e,
	0x4c, 0x69, 0x73, 0x74, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x47, 0x82, 0xd3, 0xe4, 0x93, 0x02,
	0x38, 0x12, 0x36, 0x2f, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2f, 0x7b, 0x70, 0x61, 0x72,
	0x65, 0x6e, 0x74, 0x3d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x2f, 0x2a, 0x2f, 0x6c,
	0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x2a, 0x7d, 0x2f, 0x69, 0x6d, 0x61, 0x67,
	0x65, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0xda, 0x41, 0x06, 0x70, 0x61, 0x72, 0x65,
	0x6e, 0x74, 0x1a, 0x4b, 0xca, 0x41, 0x17, 0x63, 0x6f, 0x6d, 0x70, 0x6f, 0x73, 0x65, 0x72, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f, 0x6d, 0xd2, 0x41,
	0x2e, 0x68, 0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f, 0x77, 0x77, 0x77, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x75, 0x74, 0x68,
	0x2f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2d, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x42,
	0x95, 0x01, 0x0a, 0x36, 0x63, 0x6f, 0x6d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63,
	0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6f, 0x72, 0x63, 0x68, 0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x2e, 0x61, 0x69, 0x72, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x50, 0x01, 0x5a, 0x59, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x67, 0x6f, 0x6c, 0x61, 0x6e, 0x67, 0x2e, 0x6f, 0x72, 0x67, 0x2f,
	0x67, 0x65, 0x6e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x61,
	0x70, 0x69, 0x73, 0x2f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2f, 0x6f, 0x72, 0x63, 0x68, 0x65, 0x73,
	0x74, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x61, 0x69, 0x72, 0x66, 0x6c, 0x6f, 0x77, 0x2f,
	0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x3b,
	0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_rawDescOnce sync.Once
	file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_rawDescData = file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_rawDesc
)

func file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_rawDescGZIP() []byte {
	file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_rawDescOnce.Do(func() {
		file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_rawDescData = protoimpl.X.CompressGZIP(file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_rawDescData)
	})
	return file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_rawDescData
}

var file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_goTypes = []interface{}{
	(*ListImageVersionsRequest)(nil),  // 0: google.cloud.orchestration.airflow.service.v1beta1.ListImageVersionsRequest
	(*ListImageVersionsResponse)(nil), // 1: google.cloud.orchestration.airflow.service.v1beta1.ListImageVersionsResponse
	(*ImageVersion)(nil),              // 2: google.cloud.orchestration.airflow.service.v1beta1.ImageVersion
	(*date.Date)(nil),                 // 3: google.type.Date
}
var file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_depIdxs = []int32{
	2, // 0: google.cloud.orchestration.airflow.service.v1beta1.ListImageVersionsResponse.image_versions:type_name -> google.cloud.orchestration.airflow.service.v1beta1.ImageVersion
	3, // 1: google.cloud.orchestration.airflow.service.v1beta1.ImageVersion.release_date:type_name -> google.type.Date
	0, // 2: google.cloud.orchestration.airflow.service.v1beta1.ImageVersions.ListImageVersions:input_type -> google.cloud.orchestration.airflow.service.v1beta1.ListImageVersionsRequest
	1, // 3: google.cloud.orchestration.airflow.service.v1beta1.ImageVersions.ListImageVersions:output_type -> google.cloud.orchestration.airflow.service.v1beta1.ListImageVersionsResponse
	3, // [3:4] is the sub-list for method output_type
	2, // [2:3] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_init() }
func file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_init() {
	if File_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListImageVersionsRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListImageVersionsResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ImageVersion); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_goTypes,
		DependencyIndexes: file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_depIdxs,
		MessageInfos:      file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_msgTypes,
	}.Build()
	File_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto = out.File
	file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_rawDesc = nil
	file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_goTypes = nil
	file_google_cloud_orchestration_airflow_service_v1beta1_image_versions_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// ImageVersionsClient is the client API for ImageVersions service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ImageVersionsClient interface {
	// List ImageVersions for provided location.
	ListImageVersions(ctx context.Context, in *ListImageVersionsRequest, opts ...grpc.CallOption) (*ListImageVersionsResponse, error)
}

type imageVersionsClient struct {
	cc grpc.ClientConnInterface
}

func NewImageVersionsClient(cc grpc.ClientConnInterface) ImageVersionsClient {
	return &imageVersionsClient{cc}
}

func (c *imageVersionsClient) ListImageVersions(ctx context.Context, in *ListImageVersionsRequest, opts ...grpc.CallOption) (*ListImageVersionsResponse, error) {
	out := new(ListImageVersionsResponse)
	err := c.cc.Invoke(ctx, "/google.cloud.orchestration.airflow.service.v1beta1.ImageVersions/ListImageVersions", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ImageVersionsServer is the server API for ImageVersions service.
type ImageVersionsServer interface {
	// List ImageVersions for provided location.
	ListImageVersions(context.Context, *ListImageVersionsRequest) (*ListImageVersionsResponse, error)
}

// UnimplementedImageVersionsServer can be embedded to have forward compatible implementations.
type UnimplementedImageVersionsServer struct {
}

func (*UnimplementedImageVersionsServer) ListImageVersions(context.Context, *ListImageVersionsRequest) (*ListImageVersionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListImageVersions not implemented")
}

func RegisterImageVersionsServer(s *grpc.Server, srv ImageVersionsServer) {
	s.RegisterService(&_ImageVersions_serviceDesc, srv)
}

func _ImageVersions_ListImageVersions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListImageVersionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ImageVersionsServer).ListImageVersions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.cloud.orchestration.airflow.service.v1beta1.ImageVersions/ListImageVersions",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ImageVersionsServer).ListImageVersions(ctx, req.(*ListImageVersionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _ImageVersions_serviceDesc = grpc.ServiceDesc{
	ServiceName: "google.cloud.orchestration.airflow.service.v1beta1.ImageVersions",
	HandlerType: (*ImageVersionsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListImageVersions",
			Handler:    _ImageVersions_ListImageVersions_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "google/cloud/orchestration/airflow/service/v1beta1/image_versions.proto",
}