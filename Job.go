package libIM

type Job struct {
	ComputingFlag string `json:"computingFlag"`
	Mem           int    `json:"mem"`
	Sh            string `json:"sh"`
}

func NewJob(mem int) (job Job) {
	job = Job{}
	job.Mem = mem
	job.ComputingFlag = "cpu"
	return
}
