package main

import (
	"context"
	proto "grpc/protoc"
	"net/http"
	"fmt"
	"io"
	"time"
	"strconv"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

var client proto.ExampleClient

func main() {
	// Connection to internal grpc server
	conn, err := grpc.Dial("localhost:9000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	client = proto.NewExampleClient(conn)
	// implement rest api
	r := gin.Default()
	r.GET("/unary/:message", clientConnectionServer)
	r.GET("/client", clientStreamingConnectionServer)
	r.GET("/server", ServerStreamingConnectionServer)
	r.GET("/bilateral", BilateralStreamReply)
	r.Run(":8000") // 8080

}

func clientConnectionServer(c *gin.Context) {
	variableName := c.Param("message")

	req := &proto.HelloRequest{SomeString: variableName}

	client.ServerReply(context.TODO(), req)
	c.JSON(http.StatusOK, gin.H{
		"message": "message sent suceesfully to server " + variableName,
	})
}

func clientStreamingConnectionServer(c *gin.Context) {

	req := []*proto.HelloRequest{
		{SomeString: "Request 1"},
		{SomeString: "Request 2"},
		{SomeString: "Request 3"},
		{SomeString: "Request 4"},
		{SomeString: "Request 5"},
		{SomeString: "Request 6"},
	}

	stream, err := client.ClientServerReply(context.TODO())
	if err != nil {
		fmt.Println("Something error")
		return
	}
	for _, re := range req {
		err = stream.Send(re)
		if err != nil {
			fmt.Println("request not fulfil")
			return
		}

	}
	response, err := stream.CloseAndRecv()
	if err != nil {
		fmt.Println("there is some error occure ", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message_count": response,
	})

}

func ServerStreamingConnectionServer(c *gin.Context) {
	stream, err := client.ServerStreamReply(context.TODO(), &proto.HelloRequest{SomeString: "Hii from the Client"})
	if err != nil {
		fmt.Println("Something error")
		return
	}
	count := 0
	for {
		message, err := stream.Recv()
		if err == io.EOF {
			break
		}
		fmt.Println("gServer messages:- ", message)
		time.Sleep(1 * time.Second)
		count++
	}
	c.JSON(http.StatusOK, gin.H{
		"message_count": count,
	})
}


func BilateralStreamReply(c *gin.Context) {
	stream, err := client.BilateralStreamReply(context.TODO())
	if err != nil {
		fmt.Println("Something error")
		return
	}
	send, recieve := 0, 0
	for i := 0; i < 10; i++ {
		err := stream.Send(&proto.HelloRequest{SomeString: "message " + strconv.Itoa(i) + " from client"})
		if err != nil {
			fmt.Println("unable to send data")
			return
		}
		send++

	}
	if err := stream.CloseSend(); err != nil {
		log.Println(err)
	}
	for {
		message, err := stream.Recv()
		if err == io.EOF {
			break
		}
		fmt.Println("Server message:- ", message)
		recieve++
	}
	c.JSON(http.StatusOK, gin.H{
		"message_sent":    send,
		"message_recieve": recieve,
	})
}