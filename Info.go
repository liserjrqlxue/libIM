package libIM

type LaneInfo struct {
	LaneName string `json:"lane_code"`
	Fq1      string `json:"fastq1"`
	Fq2      string `json:"fastq2"`
}
type Info struct {
	Raw map[string]string

	SampleID     string
	Fq1          string
	Fq2          string
	ChipCode     string
	Type         string
	Gender       string
	ProductCode  string
	ProductType  string
	ProbandID    string
	RelationShip string
	HPO          string
	StandardTag  string
	StandardQC   string
	QChistory    string

	LaneInfos  []LaneInfo
	FamilyInfo map[string][]string
}

type FamilyInfo struct {
	ProbandID string
	FamilyMap map[string]string
}
