package main

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"
)

/**
* json是给json编解码使用的
* jsonschema是给JSON Schema 生成库
**/
type CalculatorArgs struct {
	Operation string  `json:"operation" jsonschema:"description=The operation to perform,enum=add,enum=subtract,enum=multiply,enum=divide,enum=power,required"`
	A         float64 `json:"a" jsonschema:"description=First number,required"`
	B         float64 `json:"b" jsonschema:"description=Second number,required"`
}

type CalculatorResult struct {
	Operation string  `json:"operation"`
	A         float64 `json:"a"`
	B         float64 `json:"b"`
	Result    float64 `json:"result"`
}

func calculator(ctx context.Context, args CalculatorArgs) (CalculatorResult, error) {
	var result float64

	switch strings.ToLower(args.Operation) {
	case "add":
		result = args.A + args.B
	case "subtract":
		result = args.A - args.B
	case "multiply":
		result = args.A * args.B
	case "divide":
		if args.B == 0 {
			return CalculatorResult{}, fmt.Errorf("除数不能为 0")
		}
		result = args.A / args.B
	case "power":
		result = math.Pow(args.A, args.B)
	default:
		return CalculatorResult{}, fmt.Errorf("不支持的 operation: %s", args.Operation)
	}

	return CalculatorResult{
		Operation: args.Operation,
		A:         args.A,
		B:         args.B,
		Result:    result,
	}, nil
}

type CurrentTimeArgs struct {
	Timezone string `json:"timezone" jsonschema:"description=Timezone name such as UTC or Local"`
}

type CurrentTimeResult struct {
	Timezone string `json:"timezone"`
	Date     string `json:"date"`
	Time     string `json:"time"`
	Weekday  string `json:"weekday"`
}

func currentTime(ctx context.Context, args CurrentTimeArgs) (CurrentTimeResult, error) {
	now := time.Now()

	timezone := strings.TrimSpace(args.Timezone)
	if timezone == "" {
		timezone = "Local"
	}

	var t time.Time

	switch strings.ToUpper(timezone) {
	case "UTC":
		t = now.UTC()
	case "LOCAL":
		t = now
	default:
		t = now
		timezone = "Local"
	}

	return CurrentTimeResult{
		Timezone: timezone,
		Date:     t.Format("2006-01-02"),
		Time:     t.Format("15:04:05"),
		Weekday:  t.Weekday().String(),
	}, nil
}