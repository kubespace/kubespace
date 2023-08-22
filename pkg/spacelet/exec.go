package spacelet

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/kubespace/kubespace/pkg/core/code"
	"github.com/kubespace/kubespace/pkg/utils"
	"net/http"
	"os/exec"
)

type ExecResponse struct {
	Status int
	Stderr string
	Stdin  string
	Stdout string
}

type ExecRequest struct {
	Command    string `json:"command"`
	Executable string `json:"executable"`
}

func (s *Server) Exec(c *gin.Context) {
	req := ExecRequest{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, &utils.Response{Code: code.ParamsError, Msg: err.Error()})
		return
	}
	if req.Command == "" {
		c.JSON(http.StatusOK, &utils.Response{Code: code.ParamsError, Msg: "command is required"})
		return
	}
	executable := req.Executable
	if executable == "" {
		executable = "/bin/sh"
	}

	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	stdin := bytes.NewBuffer(nil)

	cmd := exec.Command(executable, "-c", req.Command)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()

	res := ExecResponse{}
	if err != nil {
		res.Status = 1
	} else {
		res.Status = 0
	}
	res.Stdout = stdout.String()
	res.Stderr = stderr.String()
	c.JSON(http.StatusOK, &utils.Response{Code: code.Success, Data: res})
	return
}
