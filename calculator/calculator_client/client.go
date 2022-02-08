package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"grpc-proto-playground/calculator/calculatorpb"
	"io"
	"log"
)

func main() {
	fmt.Println("hello from calculator client")
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalln("could not connect: ", err)
	}
	defer cc.Close()
	c := calculatorpb.NewCalculatorServiceClient(cc)
	//doUnary(c)
	//doServerStream(c)
	//doClientStream(c)
	//doBiDiStreaming(c)
	calculateSquareRoot(c)
}

func doUnary(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("invoking unary RPC...")

	req := &calculatorpb.CalculatorRequest{
		Calculator: &calculatorpb.Calculator{
			FirstOperand:  10,
			SecondOperand: 0,
			Operator:      "-",
		},
	}

	res, err := c.Calculate(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling calculate rpc: %v", err)
	}
	log.Printf("response: %v", res.Result)
}

func doServerStream(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("invoking doServerStream RPC...")

	req := &calculatorpb.PrimeNumberRequest{
		Number: 102410241024,
	}

	stream, err := c.CalculatePrimeNumberDecomposition(context.Background(), req)

	if err != nil {
		log.Fatalf("error while callig prime number decom: %v \n", err)
	}

	for {

		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		fmt.Println(msg.GetFactor())
	}
}

func doClientStream(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("invoking doServerStream RPC...")

	requests := []*calculatorpb.CalculateAverageRequest{
		{
			Number: 1,
		},
		{
			Number: 2,
		},
		{
			Number: 3,
		},
		{
			Number: 4,
		},
	}

	stream, err := c.CalculateAverage_ClientStreaming(context.Background())

	if err != nil {
		log.Fatalln("asdjsal")
	}
	for _, req := range requests {
		err := stream.Send(req)
		if err != nil {
			return
		}
	}
	res, err2 := stream.CloseAndRecv()
	if err2 != nil {
		log.Fatalln("asdjsal")
	}
	fmt.Println(res)
}
func doBiDiStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("invoking doServerStream RPC...")
	requests := []*calculatorpb.FindMaxRequest{
		{
			Number: 3,
		},
		{
			Number: 2,
		},
		{
			Number: 5,
		},
		{
			Number: 1,
		},
		{
			Number: 10,
		},
		{
			Number: 2,
		},
		{
			Number: 20,
		},
	}

	stream, err := c.FindMaxBiDiStreaming(context.Background())

	if err != nil {
		return
	}

	waitChannel := make(chan struct{})
	go func() {
		for _, req := range requests {
			fmt.Println("sending")
			stream.Send(req)
			//uncomment sleep untuk bukti kalo ini async
			//time.Sleep(time.Second * 2)
		}
		stream.CloseSend()
	}()

	go func() {

		for {
			res, err := stream.Recv()
			fmt.Println("rec")
			if err == io.EOF {
				close(waitChannel)
			}
			if err != nil {
				return
			}
			if res != nil {
				fmt.Println(res)
			}
		}

	}()

	<-waitChannel
}

func calculateSquareRoot(c calculatorpb.CalculatorServiceClient) {

	req := &calculatorpb.SquareRootRequest{
		Number: -1024,
	}

	res, err := c.SquareRoot(context.Background(), req)
	if err != nil {
		respErr, ok := status.FromError(err)
		if ok {
			//actual error from grpc
			fmt.Println(respErr.Message())
			fmt.Println(respErr.Code())
			if respErr.Code() == codes.InvalidArgument {
				fmt.Println("no negative nunmber allowed")
			}

		} else {
			log.Fatalf("error something something %v", err)
		}

	}
	fmt.Printf("results: %v", res.GetResult())

}
