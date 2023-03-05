package spacelet

import (
	"errors"
	"fmt"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/spacelet/pipeline_job"
	"github.com/kubespace/kubespace/pkg/spacelet/pipeline_job/plugins"
	"github.com/kubespace/kubespace/pkg/utils"
)

type Client interface {
	PipelineJobExecute(params *pipeline_job.JobRunParams) error
	PipelineJobStatus(params *pipeline_job.JobStatusParams) (*plugins.StatusLog, error)
	PipelineJobCleanup(params *pipeline_job.JobCleanParams) error
}

func NewClient(spacelet *types.Spacelet) (Client, error) {
	httpcli, err := utils.NewHttpClient(fmt.Sprintf("http://%s:%d", spacelet.HostIp, spacelet.Port))
	if err != nil {
		return nil, err
	}
	return &client{
		httpclient: httpcli,
		spacelet:   spacelet,
	}, nil
}

type client struct {
	httpclient *utils.HttpClient
	spacelet   *types.Spacelet
}

func (c *client) PipelineJobExecute(params *pipeline_job.JobRunParams) error {
	var resp utils.Response
	options := utils.RequestOptions{}
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

func (c *client) PipelineJobStatus(params *pipeline_job.JobStatusParams) (*plugins.StatusLog, error) {
	var resp utils.Response
	options := utils.RequestOptions{}
	options.WithHeader("token", c.spacelet.Token)
	_, err := c.httpclient.Get("/v1/pipeline_job/status", params, &resp, options)
	if err != nil {
		return nil, err
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf(resp.Msg)
	}
	var statusLog plugins.StatusLog
	if err = utils.ConvertTypeByJson(resp.Data, &statusLog); err != nil {
		return nil, err
	}
	return &statusLog, nil
}

func (c *client) PipelineJobCleanup(params *pipeline_job.JobCleanParams) error {
	var resp utils.Response
	options := utils.RequestOptions{}
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
