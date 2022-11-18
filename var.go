package libIM

var LaneInput = "null"
var Pipeline = "."
var Trio = []string{"proband", "father", "mother"}
var Threshold = 12
var Throttle = make(chan bool, Threshold)
var Assembly = "hg19"
