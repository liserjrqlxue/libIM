package libIM

type Job struct {
	ComputingFlag string `json:"computingFlag"`
	Mem           int    `json:"mem"`
	Sh            string `json:"sh"`
}

func NewJob(sh string, mem int) Job {
	return Job{
		Mem:           mem,
		ComputingFlag: "cpu",
		Sh:            sh,
	}
}
