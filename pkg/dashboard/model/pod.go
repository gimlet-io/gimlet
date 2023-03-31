package model

type Pod struct {
	ID         int64  `json:"-"  meddler:"id,pk"`
	Name       string `json:"name,omitempty"  meddler:"name"`
	Status     string `json:"status,omitempty"  meddler:"status"`
	StatusDesc string `json:"statusDesc,omitempty"  meddler:"status_desc"`
}

func (p *Pod) IsInErrorState() bool {
	return p.Status != "Running" && p.Status != "Pending" && p.Status != "Terminating" &&
		p.Status != "Succeeded" && p.Status != "Unknown" && p.Status != "ContainerCreating" &&
		p.Status != "PodInitializing"
}
