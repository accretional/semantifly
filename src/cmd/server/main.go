package main

import (
	"context"
	"log"
	"net"

	pb "accretional.com/semantifly/proto"
	"accretional.com/semantifly/subcommands"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedSemantiplyServer
}

func (s *server) Add(ctx context.Context, req *pb.AddRequest) (*pb.AddResponse, error) {
	args := subcommands.AddArgs{
		IndexPath:  req.IndexPath,
		DataType:   req.DataType,
		SourceType: req.SourceType,
		MakeCopy:   req.MakeCopy,
		DataURIs:   req.DataUris,
	}

	err := subcommands.Add(args)
	if err != nil {
		return &pb.AddResponse{Success: false, Message: err.Error()}, nil
	}
	return &pb.AddResponse{Success: true, Message: "Data added successfully"}, nil
}

func (s *server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	args := subcommands.DeleteArgs{
		IndexPath:  req.IndexPath,
		DeleteCopy: req.DeleteCopy,
		DataURIs:   req.DataUris,
	}

	err := subcommands.Delete(args)
	if err != nil {
		return &pb.DeleteResponse{Success: false, Message: err.Error()}, nil
	}
	return &pb.DeleteResponse{Success: true, Message: "Data deleted successfully"}, nil
}

func (s *server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	args := subcommands.GetArgs{
		IndexPath: req.IndexPath,
		Name:      req.Name,
	}

	content, err := subcommands.Get(args)
	if err != nil {
		return &pb.GetResponse{Success: false, Message: err.Error()}, nil
	}
	return &pb.GetResponse{Success: true, Content: content}, nil
}

func (s *server) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	args := subcommands.UpdateArgs{
		IndexPath:  req.IndexPath,
		Name:       req.Name,
		DataType:   req.DataType,
		SourceType: req.SourceType,
		UpdateCopy: req.UpdateCopy,
		DataURI:    req.DataUri,
	}

	err := subcommands.Update(args)
	if err != nil {
		return &pb.UpdateResponse{Success: false, Message: err.Error()}, nil
	}
	return &pb.UpdateResponse{Success: true, Message: "Data updated successfully"}, nil
}

func (s *server) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	args := subcommands.LexicalSearchArgs{
		IndexPath:  req.IndexPath,
		SearchTerm: req.SearchTerm,
		TopN:       int(req.TopN),
	}

	results, err := subcommands.LexicalSearch(args)
	if err != nil {
		return &pb.SearchResponse{Success: false, Message: err.Error()}, nil
	}

	pbResults := make([]*pb.SearchResult, len(results))
	for i, result := range results {
		pbResults[i] = &pb.SearchResult{
			Name:        result.FileName,
			Occurrences: float32(result.Occurrence),
		}
	}

	return &pb.SearchResponse{Success: true, Results: pbResults}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterSemantiplyServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
