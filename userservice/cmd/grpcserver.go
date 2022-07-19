package cmd

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/situmorangbastian/skyros/userservice"
	grpcUser "github.com/situmorangbastian/skyros/userservice/grpc"
	grpcHandler "github.com/situmorangbastian/skyros/userservice/internal/grpc"
	"google.golang.org/grpc"
)

func GRPCServer(userService userservice.UserService) {
	port, err := strconv.Atoi(userservice.GetEnv("GRPC_SERVER_ADDRESS"))
	if err != nil {
		log.Fatal(errors.New("invalid grpc server port"))
	}

	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	grpcUserService := grpcHandler.NewUserGrpcServer(userService)
	grpcUser.RegisterUserServiceServer(grpcServer, grpcUserService)

	fmt.Println("GRPC Server Running on Port: ", port)
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatal(err)
	}
}
