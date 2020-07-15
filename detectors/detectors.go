package detectors

import "github.com/sirupsen/logrus"

//Resource a computational resource
type Resource struct {
	//Typ resource type. ex: ram, disk, cpu, fs, net
	Typ string
	//Name identification of the resource in the system. ex: /dev/fda021, cpu:23, process:prometheus[6456]
	Name string
	//PropertyName name of the resource property. ex: available-cpu, disk-writes-per-second
	PropertyName string
	//PropertyValue value of the resource property
	PropertyValue float64
}

//Issue detection results
type Issue struct {
	//Typ type of issue: bottleneck, risk, harm, lib-error
	Typ string
	//Name name of the issue. ex: low-available-cpu, disk-write-ceil
	Name string
	//Score how critical is this issue to the health of the system
	Score float64
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
	logrus.Debugf("Registering detector %v", d)
	DetectorFuncs = append(DetectorFuncs, d)
}

//Options performance analysis options
type Options struct {
	Loglevel                string
	LowMemPercRange         [2]float64
	LowCPUPercRange         [2]float64
	LowDiskPercRange        [2]float64
	LowFileHandlesPercRange [2]float64
}

//NewOptions create a new default options
func NewOptions() Options {
	return Options{
		LowMemPercRange:         [2]float64{0.70, 0.95},
		LowCPUPercRange:         [2]float64{0.70, 0.95},
		LowDiskPercRange:        [2]float64{0.70, 0.90},
		LowFileHandlesPercRange: [2]float64{0.70, 0.90},
	}
}

//calculates a score between 0-1. 0 is "no worry"; 1 is "IT BROKE!"
func score(value float64, criticityRange [2]float64) float64 {
	if value < criticityRange[0] {
		return 0
	}
	return (value - criticityRange[0]) / (criticityRange[1] - criticityRange[0])
}
