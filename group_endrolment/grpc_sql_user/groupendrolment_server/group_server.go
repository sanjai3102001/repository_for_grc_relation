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
	CREATE TABLE group_enrollment (
		group_id UUID,
		user_id UUID,
		PRIMARY KEY (group_id, user_id),
		FOREIGN KEY (group_id) REFERENCES groups(group_id),
		FOREIGN KEY (user_id) REFERENCES users(user_id)
	  );
	`
	_, err := server.conn.Exec(context.Background(), createSql)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Table creation failed: %v\n", err)
		os.Exit(1)
	}

	server.first_user_creation = false

	log.Printf("Received: %v %v", in.GetGroupId(), in.GetUserId())

	created_user := &pb.User{GroupId: in.GetGroupId(), UserId: in.GetUserId()}
	tx, err := server.conn.Begin(context.Background())
	if err != nil {
		log.Fatalf("conn.Begin failed: %v", err)
	}

	_, err = tx.Exec(context.Background(), "insert into group_endrolment(group_id, user_id) values ($1,$2)",
		created_user.GroupId, created_user.UserId)
	if err != nil {
		log.Fatalf("tx.Exec failed: %v", err)
	}
	tx.Commit(context.Background())
	return created_user, nil

}

func (server *UserManagementServer) GetUsers(ctx context.Context, in *pb.GetUsersParams) (*pb.UsersList, error) {

	var users_list *pb.UsersList = &pb.UsersList{}
	rows, err := server.conn.Query(context.Background(), "select * from group_endrolment")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		user := pb.User{}
		err = rows.Scan(&user.UserId, &user.GroupId)
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
