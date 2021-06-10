package interfaces

// Service interface defines the methods that every kind of interface, whether
// gRPC, REST, or whatever must be complaint with.
type Service interface {
	Start(operatorAddress, tradeAddress string) error
	Stop()
}
