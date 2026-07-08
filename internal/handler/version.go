package handler

import (
	"net/http"

	"lmvpn/internal/version"

	"github.com/gin-gonic/gin"
)

type versionResponse struct {
	Version    string `json:"version"`
	Commit     string `json:"commit"`
	CommitTime string `json:"commit_time"`
	BuildTime  string `json:"build_time"`
}

func GetVersion(c *gin.Context) {
	c.JSON(http.StatusOK, versionResponse{
		Version:    version.Version,
		Commit:     version.Commit,
		CommitTime: version.CommitTime,
		BuildTime:  version.BuildTime,
	})
}
