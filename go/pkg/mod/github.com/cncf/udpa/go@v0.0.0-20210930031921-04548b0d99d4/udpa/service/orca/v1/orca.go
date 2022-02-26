package udpa_service_orca_v1

import (
	xds "github.com/cncf/xds/go/udpa/service/orca/v1" // Note this is cncf/xds
)

type OrcaLoadReportRequest = xds.OrcaLoadReportRequest

// OpenRcaServiceClient is the client API for OpenRcaService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type OpenRcaServiceClient = xds.OpenRcaServiceClient

var NewOpenRcaServiceClient = xds.NewOpenRcaServiceClient
var RegisterOpenRcaServiceServer = xds.RegisterOpenRcaServiceServer

type OpenRcaService_StreamCoreMetricsServer = xds.OpenRcaService_StreamCoreMetricsServer
