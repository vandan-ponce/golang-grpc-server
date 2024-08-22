package main

import (
	"context"
	"errors"
	"fmt"
	proto "grpc/protoc"
	"net"
	"strconv"	
	"io"
	"time"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	proto.UnimplementedExampleServer
}

func main() {

	listener, tcpErr := net.Listen("tcp", ":9000")
	if tcpErr != nil {
		panic(tcpErr)
	}
	srv := grpc.NewServer() // engine
	proto.RegisterExampleServer(srv, &server{})
	reflection.Register(srv)

	if e := srv.Serve(listener); e != nil {
		panic(e)
	}
}

func (s *server) ServerReply(c context.Context, req *proto.HelloRequest) (*proto.HelloResponse, error) {
	fmt.Println("recieve request from client", req.SomeString)
	fmt.Println("hello from server")
	return &proto.HelloResponse{}, errors.New("")
}

func (s *server) ClientServerReply(stream grpc.ClientStreamingServer[proto.HelloRequest, proto.HelloResponse]) (error){
	total :=0
	for {
		request , err := stream.Recv()
		if err == io.EOF{
			return stream.SendAndClose(&proto.HelloResponse{
				Reply: strconv.Itoa(total),
			})
		}
		if err != nil {
			return err
		}
		total++;
		fmt.Println(request);
	}
}

func (s *server) ServerStreamReply(req *proto.HelloRequest , stream grpc.ServerStreamingServer[proto.HelloResponse]) (error){
	fmt.Println(req.SomeString)
	time.Sleep(5 * time.Second)
	serverreply := []*proto.HelloResponse{
		{Reply: "Hii there"},
		{Reply: "How are you ?"},
		{Reply: "Welcome to Codewits"},
		{Reply: "This is a gRPC server"},
	}
	for _, msg := range serverreply {
		err := stream.Send(msg)
		if err != nil {
			return err
		}
	}
	return nil
}


func (s *server) BilateralStreamReply(stream grpc.BidiStreamingServer[proto.HelloRequest, proto.HelloResponse]) (error){
	for i := 0; i < 5; i++ {
		err := stream.Send(&proto.HelloResponse{Reply: "message " + strconv.Itoa(i) + " from server"})
		if err != nil {
			return errors.New("unable to send data from server")
		}
	}
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		fmt.Println(req.SomeString)
	}
	return nil
}