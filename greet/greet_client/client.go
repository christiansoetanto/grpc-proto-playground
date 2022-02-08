package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"grpc-proto-playground/greet/greetpb"
	"io"
	"log"
	"time"
)

func main() {
	fmt.Println("hello from client")

	certFile := "ssl/ca.crt"
	creds, sslErr := credentials.NewClientTLSFromFile(certFile, "")
	if sslErr != nil {
		log.Fatalf("error bro %v\n", sslErr)
		return

	}

	opts := grpc.WithTransportCredentials(creds)
	cc, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalln("could not connect: ", err)

	}
	defer cc.Close()
	c := greetpb.NewGreetServiceClient(cc)
	doUnary(c)
	//doServerStreaming(c)
	//doClientStreaming(c)
	//doBiDiStreaming(c)

	//doUnaryWithDeadline(c, 5*time.Second) //should complete
	//doUnaryWithDeadline(c, 1*time.Second) //should timeouted
	//karena di server kita sleep selama 3s

}

func doUnary(c greetpb.GreetServiceClient) {
	fmt.Println("invoking unary RPC...")
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "firstname",
			LastName:  "lastname",
		},
	}
	res, err := c.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling greet rpc: %v", err)
	}
	log.Printf("response: %v", res.Result)
}

func doServerStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("invoking server streaming RPC")
	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "tayo",
			LastName:  "hehe",
		},
	}

	resStream, err := c.GreetManyTimes(context.Background(), req)
	if err != nil {
		fmt.Printf("error while server streaming %v", err)
	}

	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("error while reading stream %v", err)
		} else {
			log.Printf("response from greetmanytimes: %v", msg)
		}

	}

}
func doClientStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("invoking client streaming RPC")
	reqStream := []*greetpb.GreetClientStreamExampleRequest{
		{
			Greeting: &greetpb.Greeting{
				FirstName: "asd",
				LastName:  "asjdlk",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "asd",
				LastName:  "asjdlk",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "asd",
				LastName:  "asjdlk",
			},
		},
	}

	stream, err := c.GreetClientStreamExample(context.Background())
	if err != nil {
		fmt.Printf("error while server streaming %v", err)
	}

	for _, req := range reqStream {
		stream.Send(req)
	}

	msg, err := stream.CloseAndRecv()
	if err != nil {
		fmt.Printf("error while reading stream %v", err)
	} else {
		log.Printf("response from greetmanytimes: %v", msg)
	}

}

func doBiDiStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("invoking client streaming RPC")
	//open stream
	stream, err := c.GreetBiDirectionalStreamingExample(context.Background())

	if err != nil {
		return
	}
	waitChannel := make(chan struct{})
	reqStream := []*greetpb.GreetBiDirectionalStreamingExampleRequest{
		{
			Greeting: &greetpb.Greeting{
				FirstName: "asd",
				LastName:  "asjdlk",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "asd",
				LastName:  "asjdlk",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "asd",
				LastName:  "asjdlk",
			},
		},
	}

	//send a bunch of requests
	go func() {

		for _, req := range reqStream {
			fmt.Println("sending messeg")
			stream.Send(req)
			time.Sleep(time.Second)
		}
		stream.CloseSend()
	}()

	//receive a bunch of rqeuests

	go func() {

		for {
			res, err := stream.Recv()
			if err == io.EOF {
				close(waitChannel)
			}
			if err != nil {
				return
			}

			fmt.Println(res)
		}

	}()
	//block channel until closed
	<-waitChannel
}

func doUnaryWithDeadline(c greetpb.GreetServiceClient, seconds time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), seconds)
	defer cancel()
	resp, err := c.GreetWithDeadline(
		ctx, &greetpb.GreetWithDeadlineRequest{
			Greeting: "hello",
		},
	)
	if err != nil {
		respErr, ok := status.FromError(err)
		//ok ada error dari RPC
		if ok {
			message, code := respErr.Message(), respErr.Code()
			fmt.Printf("message: %v\n code: %v\n", message, code)
			if code == codes.DeadlineExceeded {
				fmt.Println("lewat deadline bro")
			} else {
				log.Fatalf("entahlah\n")
			}

		} else {
			log.Fatalf("error something something %v", err)
		}
		return
	}

	fmt.Println(resp.GetResult())
}
