/*
 * Copyright 2022 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions a nd
 * limitations under the License.
 */

package prometheus

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	kclient "github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"
	"github.com/kitex-contrib/monitor-prometheus/kitex_gen/api"
	"github.com/kitex-contrib/monitor-prometheus/kitex_gen/api/echo"
	"github.com/stretchr/testify/assert"
)

// HelloImpl implements the last service interface defined in the IDL.
type helloImpl struct{}

// Echo implements the HelloImpl interface.
func (s *helloImpl) Echo(ctx context.Context, req *api.Request) (resp *api.Response, err error) {
	resp = &api.Response{Message: req.Message}
	return
}

// TestAllInOne all metrics in one registry
func TestAllInOne(t *testing.T) {
	SetGlobalAddrPath(":9092", "/metric")

	svr := echo.NewServer(new(helloImpl),
		server.WithTracer(NewGlobalServerTracer()),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "test-server"}),
		server.WithMetaHandler(transmeta.ServerTTHeaderHandler),
	)
	go func() {
		if err := svr.Run(); err != nil {
			klog.Fatal(err)
		}
	}()

	time.Sleep(time.Second) // wait server start

	req := &api.Request{Message: "my request"}

	client1, err := echo.NewClient("test-server",
		kclient.WithHostPorts("0.0.0.0:8888"), kclient.WithTracer(NewGlobalClientTracer()),
		kclient.WithTransportProtocol(transport.TTHeader),
		kclient.WithMetaHandler(transmeta.ClientTTHeaderHandler),
		kclient.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "test-client1"}),
	)
	if err != nil {
		klog.Fatal(err)
	}

	_, err = client1.Echo(context.Background(), req)

	assert.Nil(t, err)

	client2, err := echo.NewClient("test-server",
		kclient.WithHostPorts("0.0.0.0:8888"), kclient.WithTracer(NewGlobalClientTracer()),
		kclient.WithTransportProtocol(transport.TTHeader),
		kclient.WithMetaHandler(transmeta.ClientTTHeaderHandler),
		kclient.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "test-client2"}),
	)
	if err != nil {
		klog.Fatal(err)
	}

	_, err = client2.Echo(context.Background(), req)

	assert.Nil(t, err)

	promServerResp, err := http.Get("http://127.0.0.1:9092/metric")

	assert.Nil(t, err)
	assert.True(t, promServerResp.StatusCode == http.StatusOK)

	defer promServerResp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(promServerResp.Body)
	assert.True(t, err == nil)
	respStr := string(bodyBytes)

	assert.True(t, strings.Contains(respStr, "kitex_server_latency_us_count{callee=\"test-server\",caller=\"test-client1\",method=\"echo\",retry=\"0\",status=\"succeed\"} 1"))
	assert.True(t, strings.Contains(respStr, "kitex_server_latency_us_count{callee=\"test-server\",caller=\"test-client2\",method=\"echo\",retry=\"0\",status=\"succeed\"} 1"))
	assert.True(t, strings.Contains(respStr, "kitex_server_throughput{callee=\"test-server\",caller=\"test-client1\",method=\"echo\",retry=\"0\",status=\"succeed\"} 1"))
	assert.True(t, strings.Contains(respStr, "kitex_server_throughput{callee=\"test-server\",caller=\"test-client2\",method=\"echo\",retry=\"0\",status=\"succeed\"} 1"))
}
