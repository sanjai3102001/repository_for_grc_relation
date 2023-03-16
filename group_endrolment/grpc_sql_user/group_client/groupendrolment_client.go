package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/tech-with-moss/go-usermgmt-grpc/usermgmt"
	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"
)

func main() {

	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewUserManagementClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// var user_id = "sanjai"
	// var new_users1 = make(map[string]string)
	// new_users1["user_id"] = "sanjai"
	// var arr [5]int32
	// myarr := [5]int{1, 2, 3, 4, 5}

	var new_users = make(map[int]int)
	new_users[5] = 1
	new_users[4] = 2
	new_users[3] = 3
	new_users[2] = 4
	new_users[1] = 5
	// for service_id, name, owner_id, prefix, action := range new_users {
	// 	r, err := c.CreateNewUser(ctx, &pb.NewUser{ServiceId: int32(service_id), Name: string(name), OwnerId: int32(owner_id), Prefix: int32(prefix), Action: int32(action)})
	for group_id, user_id := range new_users {
		r, err := c.CreateNewUser(ctx, &pb.NewUser{GroupId: int32(group_id), UserId: int32(user_id)})
		if err != nil {
			log.Fatalf("could not create user: %v", err)
		}
		log.Printf(`service Details:
group_id: %d
user_id: %d`, r.GetGroupId(), r.GetUserId())

	}
	params := &pb.GetUsersParams{}
	r, err := c.GetUsers(ctx, params)
	if err != nil {
		log.Fatalf("could not create user: %v", err)
	}
	log.Print("\nUSER LIST: \n")
	fmt.Printf("r.GetRoles(): %v\n", r.GetUsers())
}
