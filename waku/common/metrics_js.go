// +build js

package common

type EmptyCounter struct {
}

func (e *EmptyCounter) Inc(){

}
func (e *EmptyCounter) Add(_ float64) {
}

func (e *EmptyCounter) WithLabelValues(_ string) *EmptyCounter {
	return e
}

func (e *EmptyCounter) Observe(_ float64) {

}

var (
	EnvelopesReceivedCounter = EmptyCounter{}
	EnvelopesValidatedCounter = EmptyCounter{}
	EnvelopesRejectedCounter = EmptyCounter{}
	EnvelopesCacheFailedCounter = EmptyCounter{}
	EnvelopesCachedCounter = EmptyCounter{}
	EnvelopesSizeMeter = EmptyCounter{}
	RateLimitsProcessed = EmptyCounter{}
	RateLimitsExceeded = EmptyCounter{}
	BridgeSent = EmptyCounter{}
	BridgeReceivedSucceed = EmptyCounter{}
	BridgeReceivedFailed = EmptyCounter{}
)
