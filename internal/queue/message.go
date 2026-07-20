package queue

type EmailMessage struct {
	Type     string         `json:"type"`
	To       string         `json:"to"`
	Template string         `json:"template"`
	Data     map[string]any `json:"data"`
}

type DeployMessage struct {
	Type         string `json:"type"`
	DeploymentID string `json:"deployment_id"`
	ProjectID    string `json:"project_id"`
	UploadKey    string `json:"upload_key"`
}
