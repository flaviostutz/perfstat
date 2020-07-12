package detectors

import "fmt"

//Resource a computational resource
type Resource struct {
	//Typ resource type. ex: ram, disk, cpu, fs, net
	Typ string
	//Name identification of the resource in the system. ex: /dev/fda021, cpu:23, process:prometheus[6456]
	Name string
	//PropertyName name of the resource property. ex: available-cpu, disk-writes-per-second
	PropertyName string
	//PropertyValue value of the resource property
	PropertyValue float32
}

//Issue detection results
type Issue struct {
	//Typ type of issue: bottleneck, risk, harm, lib-error
	Typ string
	//Name name of the issue. ex: low-available-cpu, disk-write-ceil
	Name string
	//Res resource directly related to the issue
	Res Resource
	//Related related resources to the issue (ex: for low CPU, place top 3 CPU processes)
	Related []Resource
	//Message text indicating the issue details
	Message string
}

//DetectorFunc function that is called for detecting issues on the system
type DetectorFunc func(*Options) []Issue

//DetectorFuncs detector functions
var DetectorFuncs = make([]DetectorFunc, 0)

//RegisterDetector register a new function to be called for detecting issues on the system
func RegisterDetector(d DetectorFunc) {
	fmt.Printf("Registering detector %v", d)
	DetectorFuncs = append(DetectorFuncs, d)
}

//Options performance analysis options
type Options struct {
	Loglevel                      string
	LowMemCriticalPercent         float64
	LowMemWarningPercent          float64
	LowCPUCriticalPercent         float64
	LowCPUWarningPercent          float64
	LowDiskCriticalPercent        float64
	LowDiskWarningPercent         float64
	LowFileHandlesCriticalPercent float64
	LowFileHandlesWarningPercent  float64
}

//NewOptions create a new default options
func NewOptions() Options {
	return Options{
		LowMemCriticalPercent:         0.95,
		LowMemWarningPercent:          0.80,
		LowCPUCriticalPercent:         0.95,
		LowCPUWarningPercent:          0.80,
		LowDiskCriticalPercent:        0.90,
		LowDiskWarningPercent:         0.70,
		LowFileHandlesCriticalPercent: 0.90,
		LowFileHandlesWarningPercent:  0.80,
	}
}
