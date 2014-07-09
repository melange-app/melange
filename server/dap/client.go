package dap

import ()

type Client struct {
}

func (c *Client) Register()   {}
func (c *Client) Unregister() {}

func (c *Client) DownloadMessages() {}
func (c *Client) PublishMessage()   {}
func (c *Client) UpdateMessage()    {}

func (c *Client) SetData() {}
func (c *Client) GetData() {}
