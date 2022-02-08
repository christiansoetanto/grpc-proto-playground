package main

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"grpc-proto-playground/calculator/calculatorpb"
	"io"
	"log"
	"math"
	"net"
)

type server struct{}

func (s server) SquareRoot(
	_ context.Context, request *calculatorpb.SquareRootRequest,
) (*calculatorpb.SquareRootResponse, error) {

	number := request.GetNumber()
	if number < 0 {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Negative value (%v) is not allowed", number))
	}

	return &calculatorpb.SquareRootResponse{
		Result: math.Sqrt(float64(number)),
	}, nil
}

func (s server) FindMaxBiDiStreaming(stream calculatorpb.CalculatorService_FindMaxBiDiStreamingServer) error {
	//TODO implement me
	currMax := int32(0)
	for {
		req, err := stream.Recv()
		fmt.Println("receiving")
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		num := req.GetNumber()

		if num > currMax {
			currMax = num
			err2 := stream.Send(
				&calculatorpb.FindMaxResponse{
					Max: currMax,
				},
			)
			println("sending")
			if err2 != nil {
				return err2
			}
		}

	}
}

func (s server) CalculateAverage_ClientStreaming(
	stream calculatorpb.
		CalculatorService_CalculateAverage_ClientStreamingServer,
) error {

	sum := float32(0)
	i := float32(0)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(
				&calculatorpb.CalculateAverageResponse{
					Average: sum / i,
				},
			)
		}
		if err != nil {
			log.Fatalln("something heppend")
		} else {
			sum += float32(req.GetNumber())
			i++
		}
	}

	return nil
}

func (s server) CalculatePrimeNumberDecomposition(
	request *calculatorpb.PrimeNumberRequest,
	stream calculatorpb.CalculatorService_CalculatePrimeNumberDecompositionServer,
) error {
	number := request.GetNumber()

	if number <= 1 {
		return errors.New("number cannot be less than 1")
	}
	var k int64 = 2
	for number > 1 {
		if number%k == 0 {
			err := stream.Send(
				&calculatorpb.PrimeNumberResult{
					Factor: k,
				},
			)
			if err != nil {
				return err
			}
			number /= k

		} else {
			k += 1
		}
	}

	return nil
}

func (s server) Calculate(
	_ context.Context, request *calculatorpb.CalculatorRequest,
) (*calculatorpb.CalculatorResponse, error) {
	calculator := request.Calculator
	firstOperand, secondOperand, operator := calculator.GetFirstOperand(), calculator.GetSecondOperand(), calculator.GetOperator()

	var result int32
	switch operator {
	case "+":
		result = firstOperand + secondOperand
	case "-":
		result = firstOperand - secondOperand
	case "*":
		result = firstOperand * secondOperand
	case ":", "/":
		if secondOperand == 0 {
			err := errors.New("division by zero")
			return nil, err
		}
		result = firstOperand / secondOperand
	default:
		return nil, errors.New("unrecognized operator")
	}

	fmt.Println(result)
	response := &calculatorpb.CalculatorResponse{
		Result: fmt.Sprint(result),
	}
	return response, nil

}

func main() {
	fmt.Println("hello world")
	//lis,err := net.Listen(network string, address string)
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to server %v", err)
	}
}
