package util

import "fmt"

func GetHttpServiceName(serviceName string) string {
	return fmt.Sprintf("%s.http", serviceName)
}

func GetGrpcServiceName(serviceName string) string {
	return fmt.Sprintf("%s.grpc", serviceName)
}
