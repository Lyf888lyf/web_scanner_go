package internal

type APIEntry struct {
	URL        string `json:"url"`
	Method     string `json:"method"`
	PostData   string `json:"postData"`
	Status     int    `json:"status"`
	Mime       string `json:"mime"`
	Response   string `json:"response,omitempty"` // debug 模式下记录
	FromScript bool   `json:"fromScript"`         // 是否动态请求
}

type Resource struct {
	URL     string `json:"url"`
	Type    string `json:"type"`
	FromDOM bool   `json:"fromDom"`
	FromJS  bool   `json:"fromJs"`
	Status  int    `json:"status"`
	Mime    string `json:"mime"`
}

type ScanResult struct {
	URL       string     `json:"url"`
	ParentURL string     `json:"parent,omitempty"`
	Timestamp string     `json:"timestamp"`
	Title     string     `json:"title"`
	Status    int        `json:"status"`
	HTMLList  []string   `json:"htmlList"`
	JSList    []string   `json:"jsList"`
	CSSList   []string   `json:"cssList"`
	ImageList []string   `json:"imageList"`
	APIList   []APIEntry `json:"apiList"`
	OtherList []string   `json:"otherList"`

	Counts struct {
		HTML  int `json:"html"`
		JS    int `json:"js"`
		CSS   int `json:"css"`
		Image int `json:"image"`
		API   int `json:"api"`
		Other int `json:"other"`
	} `json:"counts"`
}
