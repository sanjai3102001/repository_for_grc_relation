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
	prefix := 2
	owner_id := 1
	action := 1
	var new_users = make(map[int]string)
	new_users[5] = "Sanjai"
	new_users[4] = "Dhanalakshmi"
	new_users[3] = "saravana"
	new_users[2] = "Adish"
	new_users[1] = "Haris"
	// for service_id, name, owner_id, prefix, action := range new_users {
	// 	r, err := c.CreateNewUser(ctx, &pb.NewUser{ServiceId: int32(service_id), Name: string(name), OwnerId: int32(owner_id), Prefix: int32(prefix), Action: int32(action)})
	for service_id, name := range new_users {
		r, err := c.CreateNewUser(ctx, &pb.NewUser{ServiceId: int32(service_id), Name: string(name), OwnerId: int32(owner_id), Prefix: int32(prefix), Action: int32(action)})
		if err != nil {
			log.Fatalf("could not create service: %v", err)
		}
		log.Printf(`service Details:
service_id: %d
name: %s
owner_id: %d
prefix: %d
action: %d`, r.GetServiceId(), r.GetName(), r.GetOwnerServiceId(), r.GetPrefix(), r.GetAction())

	}
	params := &pb.GetUsersParams{}
	r, err := c.GetUsers(ctx, params)
	if err != nil {
		log.Fatalf("could not create service: %v", err)
	}
	log.Print("\nUSER LIST: \n")
	fmt.Printf("r.GetRoles(): %v\n", r.GetUsers())
}
