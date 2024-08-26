package grpcclient

import (
	"context"
	"log"
	"time"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var client pb.SemantiflyClient
var conn *grpc.ClientConn

func Init() {
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Create a gRPC client using the grpc.NewClient function
	var err error
	conn, err = grpc.NewClient("localhost:50051", opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	client = pb.NewSemantiflyClient(conn)
}

func Close() {
	if conn != nil {
		conn.Close()
	}
}

func Add(args *pb.AddRequest) (*pb.AddResponse, error) {
	return client.Add(context.Background(), args)
}

func Delete(args *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	return client.Delete(context.Background(), args)
}

func Get(args *pb.GetRequest) (*pb.GetResponse, error) {
	return client.Get(context.Background(), args)
}

func Update(args *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	return client.Update(context.Background(), args)
}

func LexicalSearch(args *pb.LexicalSearchRequest) (*pb.LexicalSearchResponse, error) {
	return client.LexicalSearch(context.Background(), args)
}
