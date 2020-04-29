package libIM

var LaneInput = "null"
var Pipeline = "."
var Trio = []string{"proband", "father", "mother"}
var threshold = 12
var Throttle = make(chan bool, threshold)
