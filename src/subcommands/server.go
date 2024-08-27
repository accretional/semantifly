package subcommands

import (
	"bytes"
	"context"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
)

type Server struct {
	pb.UnimplementedSemantiflyServer
}

func (s *Server) Add(ctx context.Context, req *pb.AddRequest) (*pb.AddResponse, error) {
	var buf bytes.Buffer
	err := SubcommandAdd(req, &buf)
	if err != nil {
		return &pb.AddResponse{Success: false, Message: err.Error()}, nil
	}
	return &pb.AddResponse{Success: true, Message: buf.String()}, nil
}

func (s *Server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	var buf bytes.Buffer
	err := SubcommandDelete(req, &buf)
	if err != nil {
		return &pb.DeleteResponse{Success: false, Message: err.Error()}, nil
	}
	return &pb.DeleteResponse{Success: true, Message: buf.String()}, nil
}

func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	var buf bytes.Buffer

	content, err := SubcommandGet(req, &buf)
	if err != nil {
		return &pb.GetResponse{Success: false, Message: err.Error()}, nil
	}
	return &pb.GetResponse{Success: true, Content: content, Message: buf.String()}, nil
}

func (s *Server) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	var buf bytes.Buffer

	err := SubcommandUpdate(req, &buf)
	if err != nil {
		return &pb.UpdateResponse{Success: false, Message: err.Error()}, nil
	}
	return &pb.UpdateResponse{Success: true, Message: buf.String()}, nil
}

func (s *Server) LexicalSearch(ctx context.Context, req *pb.LexicalSearchRequest) (*pb.LexicalSearchResponse, error) {
	var buf bytes.Buffer

	results, err := SubcommandLexicalSearch(req, &buf)
	if err != nil {
		return &pb.LexicalSearchResponse{Success: false, Message: err.Error()}, nil
	}

	pbResults := make([]*pb.LexicalSearchResult, len(results))
	for i, result := range results {
		pbResults[i] = &pb.LexicalSearchResult{
			Name:        result.FileName,
			Occurrences: float32(result.Occurrence),
		}
	}

	return &pb.LexicalSearchResponse{Success: true, Message: buf.String(), Results: pbResults}, nil
}
