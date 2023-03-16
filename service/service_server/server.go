package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/jackc/pgx/v4"
	// pb "github.com/sanjai3102001/grpc_service/tree/main/grpc_sql_user_server/servicemgmt"

	// pb "github.com/sanjai3102001/gitrepo_test_1"
	pb "github.com/sanjai3102001"

	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

func NewserviceManagementServer() *serviceManagementServer {
	return &serviceManagementServer{}
}

type serviceManagementServer struct {
	conn                   *pgx.Conn
	first_service_creation bool
	pb.UnimplementedserviceManagementServer
}

func (server *serviceManagementServer) Run() error {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterserviceManagementServer(s, server)
	log.Printf("server listening at %v", lis.Addr())

	return s.Serve(lis)
}

// When service is added, read full servicelist from file into
// servicelist struct, then append new service and write new servicelist back to file
func (server *serviceManagementServer) CreateNewservice(ctx context.Context, in *pb.Newservice) (*pb.service, error) {

	createSql := `
	CREATE TABLE IF NOT EXISTS service (
	  service_id UUID  PRIMARY KEY,
	  name VARCHAR(255) NOT NULL,
	  owner UUID,
	  prefix VARCHAR(255) NOT NULL,
	  action VARCHAR(255) NOT NULL
	);
	`
	_, err := server.conn.Exec(context.Background(), createSql)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Table creation failed: %v\n", err)
		os.Exit(1)
	}

	server.first_service_creation = false

	log.Printf("Received: %v", in.GetserviceId())

	created_service := &pb.service{serviceId: in.GetserviceId(), Email: in.GetEmail()}
	tx, err := server.conn.Begin(context.Background())
	if err != nil {
		log.Fatalf("conn.Begin failed: %v", err)
	}

	_, err = tx.Exec(context.Background(), "insert into services(service_id, email) values ($1,$2)",
		created_service.serviceId, created_service.Email)
	if err != nil {
		log.Fatalf("tx.Exec failed: %v", err)
	}
	tx.Commit(context.Background())
	return created_service, nil

}

func (server *serviceManagementServer) Getservices(ctx context.Context, in *pb.GetservicesParams) (*pb.servicesList, error) {

	var services_list *pb.servicesList = &pb.servicesList{}
	rows, err := server.conn.Query(context.Background(), "select * from services")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		service := pb.service{}
		err = rows.Scan(&service.IsActive, &service.serviceId, &service.Email)
		if err != nil {
			return nil, err
		}
		services_list.services = append(services_list.services, &service)

	}
	return services_list, nil
}

func main() {
	database_url := "postgres://postgres:sanjai@localhost:5432/postgres"
	var service_mgmt_server *serviceManagementServer = NewserviceManagementServer()
	conn, err := pgx.Connect(context.Background(), database_url)
	if err != nil {
		log.Fatalf("Unable to establish connection: %v", err)
	}
	defer conn.Close(context.Background())
	service_mgmt_server.conn = conn
	service_mgmt_server.first_service_creation = true
	if err := service_mgmt_server.Run(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
