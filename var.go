package libIM

var LaneInput = "null"
var Pipeline = "."
var Trio = []string{"proband", "father", "mother"}
var threshold = 12
var throttle = make(chan bool, threshold)
