package buildcache

type (
	Metadata struct {
		TotalTasks int `json:"total_tasks"`
		Cached     int `json:"cached"`
	}
)
