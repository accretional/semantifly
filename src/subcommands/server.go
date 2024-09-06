package subcommands

import (
	"bytes"
	"context"

	db "accretional.com/semantifly/database"
	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedSemantiflyServer
	serverIndexPath string
	dbContext       context.Context
	dbConn          *db.PgxIface
}

func SemantiflyNewServer(ctx context.Context, conn *db.PgxIface, serverIndexPath string) *Server {

	return &Server{
		serverIndexPath: serverIndexPath,
		dbContext:       ctx,
		dbConn:          conn,
	}
}

func (s *Server) Add(ctx context.Context, req *pb.AddRequest) (*pb.AddResponse, error) {

	var buf bytes.Buffer
	err := SubcommandAdd(s.dbContext, s.dbConn, req, s.serverIndexPath, &buf)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.AddResponse{ErrorMessage: buf.String()}, nil
}

func (s *Server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {

	var buf bytes.Buffer
	err := SubcommandDelete(s.dbContext, s.dbConn, req, s.serverIndexPath, &buf)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.DeleteResponse{ErrorMessage: buf.String()}, nil
}

func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {

	var buf bytes.Buffer
	content, contentMetadata, err := SubcommandGet(s.dbContext, s.dbConn, req, s.serverIndexPath, &buf)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.GetResponse{Content: &content, ReturnedMetadata: contentMetadata, ErrorMessage: buf.String()}, nil
}

func (s *Server) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {

	var buf bytes.Buffer
	err := SubcommandUpdate(s.dbContext, s.dbConn, req, s.serverIndexPath, &buf)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.UpdateResponse{ErrorMessage: buf.String()}, nil
}

func (s *Server) LexicalSearch(ctx context.Context, req *pb.LexicalSearchRequest) (*pb.LexicalSearchResponse, error) {
	var buf bytes.Buffer

	results, err := SubcommandLexicalSearch(req, s.serverIndexPath, &buf)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbResults := make([]*pb.LexicalSearchResult, len(results))
	for i, result := range results {
		pbResults[i] = &pb.LexicalSearchResult{
			Name:        result.FileName,
			Occurrences: int32(result.Occurrence),
		}
	}

	return &pb.LexicalSearchResponse{ErrorMessage: buf.String(), Results: pbResults}, nil
}
