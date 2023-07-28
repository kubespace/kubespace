package spacelet

import (
	"errors"
	"fmt"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/spacelet"
	"github.com/kubespace/kubespace/pkg/spacelet/pipeline_job"
	"github.com/kubespace/kubespace/pkg/third/httpclient"
	"github.com/kubespace/kubespace/pkg/utils"
)

// Client spacelet客户端调用接口
type Client interface {
	PipelineJobExecute(params *pipeline_job.JobRunParams) error
	PipelineJobStatus(params *pipeline_job.JobStatusParams) (*pipeline_job.StatusLog, error)
	PipelineJobCleanup(params *pipeline_job.JobCleanParams) error
	PipelineJobCancel(params *pipeline_job.JobCancelParams) error

	Exec(params *spacelet.ExecRequest) (*spacelet.ExecResponse, error)
}

func NewClient(spacelet *types.Spacelet) (Client, error) {
	httpcli, err := httpclient.NewHttpClient(fmt.Sprintf("http://%s:%d", spacelet.HostIp, spacelet.Port))
	if err != nil {
		return nil, err
	}
	return &client{
		httpclient: httpcli,
		spacelet:   spacelet,
	}, nil
}

type client struct {
	httpclient *httpclient.HttpClient
	spacelet   *types.Spacelet
}

// PipelineJobExecute 调用spacelet执行流水线构建任务
func (c *client) PipelineJobExecute(params *pipeline_job.JobRunParams) error {
	var resp utils.Response
	options := httpclient.RequestOptions{}
	options.WithHeader("token", c.spacelet.Token)
	_, err := c.httpclient.Post("/v1/pipeline_job/execute", params, &resp, options)
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return errors.New(resp.Msg)
	}
	return nil
}

// PipelineJobStatus 调用spaceelt流水线构建任务执行状态以及日志
func (c *client) PipelineJobStatus(params *pipeline_job.JobStatusParams) (*pipeline_job.StatusLog, error) {
	var resp utils.Response
	options := httpclient.RequestOptions{}
	options.WithHeader("token", c.spacelet.Token)
	_, err := c.httpclient.Get("/v1/pipeline_job/status", params, &resp, options)
	if err != nil {
		return nil, err
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf(resp.Msg)
	}
	var statusLog pipeline_job.StatusLog
	if err = utils.ConvertTypeByJson(resp.Data, &statusLog); err != nil {
		return nil, err
	}
	return &statusLog, nil
}

// PipelineJobCleanup 清理流水线构建任务
func (c *client) PipelineJobCleanup(params *pipeline_job.JobCleanParams) error {
	var resp utils.Response
	options := httpclient.RequestOptions{}
	options.WithHeader("token", c.spacelet.Token)
	_, err := c.httpclient.Put("/v1/pipeline_job/cleanup", params, &resp, options)
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return errors.New(resp.Msg)
	}
	return nil
}

// PipelineJobCancel 取消流水线构建任务
func (c *client) PipelineJobCancel(params *pipeline_job.JobCancelParams) error {
	var resp utils.Response
	options := httpclient.RequestOptions{}
	options.WithHeader("token", c.spacelet.Token)
	_, err := c.httpclient.Put("/v1/pipeline_job/cancel", params, &resp, options)
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return errors.New(resp.Msg)
	}
	return nil
}

func (c *client) Exec(req *spacelet.ExecRequest) (*spacelet.ExecResponse, error) {
	var resp utils.Response
	options := httpclient.RequestOptions{}
	options.WithHeader("token", c.spacelet.Token)
	_, err := c.httpclient.Post("/v1/exec", req, &resp, options)
	if err != nil {
		return nil, err
	}
	if !resp.IsSuccess() {
		return nil, errors.New(resp.Msg)
	}
	var execResp spacelet.ExecResponse
	if err = utils.ConvertTypeByJson(resp.Data, &execResp); err != nil {
		return nil, err
	}
	return &execResp, nil
}
