package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"grpc-proto-playground/greet/greetpb"
	"io"
	"log"
	"net"
	"strconv"
	"time"
)

type server struct{}

func (s *server) GreetWithDeadline(
	ctx context.Context, request *greetpb.GreetWithDeadlineRequest,
) (*greetpb.GreetWithDeadlineResponse, error) {
	//TODO implement me

	fmt.Printf("greet invoked, %v", request)

	for i := 0; i < 3; i++ {
		if ctx.Err() == context.Canceled {
			//client cancelled the request
			fmt.Println("client canceled the request")
			return nil, status.Error(codes.Canceled, "lu cancel bro")
		}
		time.Sleep(1 * time.Second)
	}

	greeting := request.GetGreeting()
	result := "Hello " + greeting
	res := &greetpb.GreetWithDeadlineResponse{
		Result: result,
	}
	return res, nil
}

func (s *server) GreetBiDirectionalStreamingExample(
	stream greetpb.
		GreetService_GreetBiDirectionalStreamingExampleServer,
) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			log.Fatalln("error")
			return err
		}

		firstName, lastName := req.GetGreeting().GetFirstName(), req.GetGreeting().GetLastName()
		result := "Hello " + firstName + " " + lastName + " !\n"
		err2 := stream.Send(
			&greetpb.GreetBiDirectionalStreamingExampleResponse{
				Result: result,
			},
		)
		if err2 != nil {
			log.Fatalf("errrrror")
			return err2
		}
	}
	return nil
}

func (s *server) GreetClientStreamExample(stream greetpb.GreetService_GreetClientStreamExampleServer) error {
	//TODO implement me

	var res string
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(
				&greetpb.GreetClientStreamExampleResponse{
					Result: res,
				},
			)
		}
		if err != nil {
			log.Fatalln("reiceiv fatal")
		} else {

			firstName, lastName := req.GetGreeting().GetFirstName(), req.GetGreeting().GetLastName()
			res += firstName + " " + lastName + "\n"
		}
	}
	return nil

}

func (s *server) GreetManyTimes(
	req *greetpb.GreetManyTimesRequest, stream greetpb.GreetService_GreetManyTimesServer,
) error {
	fmt.Printf("greet many times invoked, %v", req)

	firstName, lastName := req.GetGreeting().GetFirstName(), req.GetGreeting().GetLastName()

	for i := 0; i < 10; i++ {

		res := &greetpb.GreetManyTimesResponse{
			Result: "hello " + firstName + " " + lastName + " number: " + strconv.Itoa(i),
		}

		stream.Send(res)
		time.Sleep(time.Second)

	}
	return nil
}

func (*server) Greet(ctx context.Context, req *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	fmt.Printf("greet invoked, %v", req)
	firstName, lastName := req.GetGreeting().GetFirstName(), req.GetGreeting().GetLastName()
	result := "Hello " + firstName + " " + lastName
	res := &greetpb.GreetResponse{
		Result: result,
	}
	return res, nil
}

func main() {
	fmt.Println("hello world")
	//lis,err := net.Listen(network string, address string)
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	certFile := "ssl/server.crt"
	keyFile := "ssl/server.pem"
	creds, sslErr := credentials.NewServerTLSFromFile(certFile, keyFile)
	if sslErr != nil {
		log.Fatalf("eror: %v\n", sslErr)
		return
	}
	s := grpc.NewServer(grpc.Creds(creds))
	reflection.Register(s)
	greetpb.RegisterGreetServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to server %v", err)
	}

}
