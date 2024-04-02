package io_prometheus_client

type MetricType int32

const (
	// COUNTER must use the Metric field "counter".
	MetricType_COUNTER MetricType = 0
	// GAUGE must use the Metric field "gauge".
	MetricType_GAUGE MetricType = 1
	// SUMMARY must use the Metric field "summary".
	MetricType_SUMMARY MetricType = 2
	// UNTYPED must use the Metric field "untyped".
	MetricType_UNTYPED MetricType = 3
	// HISTOGRAM must use the Metric field "histogram".
	MetricType_HISTOGRAM MetricType = 4
	// GAUGE_HISTOGRAM must use the Metric field "histogram".
	MetricType_GAUGE_HISTOGRAM MetricType = 5
)

// Enum value maps for MetricType.
var (
	MetricType_name = map[int32]string{
		0: "COUNTER",
		1: "GAUGE",
		2: "SUMMARY",
		3: "UNTYPED",
		4: "HISTOGRAM",
		5: "GAUGE_HISTOGRAM",
	}
	MetricType_value = map[string]int32{
		"COUNTER":         0,
		"GAUGE":           1,
		"SUMMARY":         2,
		"UNTYPED":         3,
		"HISTOGRAM":       4,
		"GAUGE_HISTOGRAM": 5,
	}
)

func (x MetricType) Enum() *MetricType {
	p := new(MetricType)
	*p = x
	return p
}

func (x MetricType) String() string {
	return ""
}


// Deprecated: Do not use.
func (x *MetricType) UnmarshalJSON(b []byte) error {
	return nil
}

// Deprecated: Use MetricType.Descriptor instead.
func (MetricType) EnumDescriptor() ([]byte, []int) {
	return file_io_prometheus_client_metrics_proto_rawDescGZIP(), []int{0}
}

type LabelPair struct {
	Name  *string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Value *string `protobuf:"bytes,2,opt,name=value" json:"value,omitempty"`
}

func (x *LabelPair) Reset() {
}

func (x *LabelPair) String() string {
	return ""
}

func (*LabelPair) ProtoMessage() {}

// Deprecated: Use LabelPair.ProtoReflect.Descriptor instead.
func (*LabelPair) Descriptor() ([]byte, []int) {
	return []byte{}, []int{}
}

func (x *LabelPair) GetName() string {
	return ""
}

func (x *LabelPair) GetValue() string {
	return ""
}

type Gauge struct {
	Value *float64 `protobuf:"fixed64,1,opt,name=value" json:"value,omitempty"`
}

func (x *Gauge) Reset() {
}

func (x *Gauge) String() string {
	return ""
}

func (*Gauge) ProtoMessage() {}

