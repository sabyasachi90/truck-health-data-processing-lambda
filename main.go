package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/location"
	"github.com/aws/aws-sdk-go-v2/service/location/types"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type TruckData struct {
	Id   string
	Fuel float64
	Lat  float64
	Lon  float64
	RPM  float64
}

func Handler(ctx context.Context, request TruckData) {
	client := influxdb2.NewClientWithOptions("http://172.27.4.241:8086",
		"",
		influxdb2.DefaultOptions().SetBatchSize(1))
	writeAPI := client.WriteAPIBlocking("auto1", "hackathon")
	cTime := time.Now()
	p := influxdb2.NewPoint("health",
		map[string]string{"id": request.Id},
		map[string]interface{}{"fuel": request.Fuel, "rpm": request.RPM, "lat": request.Lat, "lng": request.Lon},
		cTime)
	error := writeAPI.WritePoint(context.Background(), p)
	if error != nil {
		fmt.Print(error)
	}
	client.Close()
	updateTracker(request, cTime)
}

func updateTracker(request TruckData, cTime time.Time) {
	cfg, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion("eu-west-1"))
	locationClient := location.NewFromConfig(cfg)

	//Device updates

	locationUpdate := types.DevicePositionUpdate{
		DeviceId:   aws.String(request.Id),
		Position:   []float64{request.Lat, request.Lon},
		SampleTime: aws.Time(cTime),
	}
	locationUpdateInput := &location.BatchUpdateDevicePositionInput{
		TrackerName: aws.String("TruckTracker"),
		Updates:     []types.DevicePositionUpdate{locationUpdate},
	}

	_, err := locationClient.BatchUpdateDevicePosition(context.TODO(), locationUpdateInput)
	if err != nil {
		fmt.Printf("Error occured during tracker update " + err.Error())
	}
}

func main() {
	lambda.Start(Handler)
}
