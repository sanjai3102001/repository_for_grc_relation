package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/jackc/pgx/v4"
	pb "github.com/tech-with-moss/go-usermgmt-grpc/usermgmt"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

func NewUserManagementServer() *UserManagementServer {
	return &UserManagementServer{}
}

type UserManagementServer struct {
	conn                *pgx.Conn
	first_user_creation bool
	pb.UnimplementedUserManagementServer
}

func (server *UserManagementServer) Run() error {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterUserManagementServer(s, server)
	log.Printf("server listening at %v", lis.Addr())

	return s.Serve(lis)
}

// When user is added, read full userlist from file into
// userlist struct, then append new user and write new userlist back to file
func (server *UserManagementServer) CreateNewUser(ctx context.Context, in *pb.NewUser) (*pb.User, error) {

	createSql := `
	CREATE TABLE IF NOT EXIST groups (
		group_id UUID  PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		member INTEGER[] NOT NULL,
		admins INTEGER[] NOT NULL,
		service_id UUID,
		FOREIGN KEY (service_id) REFERENCES service(service_id)
	  );
	`
	_, err := server.conn.Exec(context.Background(), createSql)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Table creation failed: %v\n", err)
		os.Exit(1)
	}

	server.first_user_creation = false

	log.Printf("Received: %v %v %v %v %v", in.GetGroupId(), in.GetName(), in.GetMember(), in.GetAdmin(), in.GetServiceId())

	created_user := &pb.User{GroupId: in.GetGroupId(), Name: in.GetName(), Member: in.GetMember(), Admin: in.GetAdmin(), ServiceId: in.GetServiceId()}
	tx, err := server.conn.Begin(context.Background())
	if err != nil {
		log.Fatalf("conn.Begin failed: %v", err)
	}

	_, err = tx.Exec(context.Background(), "insert into group(group_id, name, member, admin, service_id) values ($1,$2,$3,$4,$5)",
		created_user.GroupId, created_user.Name, created_user.Member, created_user.Admin, created_user.ServiceId)
	if err != nil {
		log.Fatalf("tx.Exec failed: %v", err)
	}
	tx.Commit(context.Background())
	return created_user, nil

}

func (server *UserManagementServer) GetUsers(ctx context.Context, in *pb.GetUsersParams) (*pb.UsersList, error) {

	var users_list *pb.UsersList = &pb.UsersList{}
	rows, err := server.conn.Query(context.Background(), "select * from group")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		user := pb.User{}
		err = rows.Scan(&user.ServiceId, &user.Admin, &user.Member, &user.Name, &user.GroupId)
		if err != nil {
			return nil, err
		}
		users_list.Users = append(users_list.Users, &user)

	}
	return users_list, nil
}

func main() {
	database_url := "postgres://postgres:sanjai@localhost:5432/postgres"
	var user_mgmt_server *UserManagementServer = NewUserManagementServer()
	conn, err := pgx.Connect(context.Background(), database_url)
	if err != nil {
		log.Fatalf("Unable to establish connection: %v", err)
	}
	defer conn.Close(context.Background())
	user_mgmt_server.conn = conn
	user_mgmt_server.first_user_creation = true
	if err := user_mgmt_server.Run(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
