package grpcpool

import (
	"context"
	"strings"
	"sync"
	"time"

	"LittlePudding/modules/rpc/auth"
	rpc "LittlePudding/modules/rpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	backOffMaxDelay = 3 * time.Second
	dialTimeout     = 2 * time.Second
)

var (
	Pool = &GRPCPool{
		conns: make(map[string]*Client),
	}

	keepAliveParams = keepalive.ClientParameters{
		Time:                20 * time.Second,
		Timeout:             3 * time.Second,
		PermitWithoutStream: true,
	}
)

type Client struct {
	conn      *grpc.ClientConn
	rpcClient rpc.TaskServiceClient
}

type GRPCPool struct {
	// map key格式 ip:port
	conns map[string]*Client
	mu    sync.RWMutex
}

func (p *GRPCPool) Get(addr string) (rpc.TaskServiceClient, error) {
	p.mu.RLock()
	client, ok := p.conns[addr]
	p.mu.RUnlock()
	if ok {
		return client.rpcClient, nil
	}

	client, err := p.factory(addr)
	if err != nil {
		return nil, err
	}

	return client.rpcClient, nil
}

// 释放连接
func (p *GRPCPool) Release(addr string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	client, ok := p.conns[addr]
	if !ok {
		return
	}
	delete(p.conns, addr)
	client.conn.Close()
}

// 创建连接
func (p *GRPCPool) factory(addr string) (*Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	client, ok := p.conns[addr]
	if ok {
		return client, nil
	}
	opts := []grpc.DialOption{
		grpc.WithKeepaliveParams(keepAliveParams),
		grpc.WithBackoffMaxDelay(backOffMaxDelay),
	}

	server := strings.Split(addr, ":")
	certificate := auth.Certificate{
		ServerName: server[0],
		CertFile:   "cert.pem",
		KeyFile:    "key.pem",
	}

	transportCreds, err := certificate.GetTransportCredsForClient()
	if err != nil {
		return nil, err
	}
	opts = append(opts, grpc.WithTransportCredentials(transportCreds))

	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return nil, err
	}

	client = &Client{
		conn:      conn,
		rpcClient: rpc.NewTaskServiceClient(conn),
	}

	p.conns[addr] = client

	return client, nil
}
