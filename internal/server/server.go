package server

import (
	"log"
	"net/http"
)

func StartReportServer(reportPath string) {
	fs := http.FileServer(http.Dir(reportPath))
	http.Handle("/", fs) // 根路径提供静态资源

	go func() {
		log.Println("Serving report on http://localhost:8080/")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
}
