package file

import (
	"context"
	"time"

	pb "core-service/proto/filepb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	cfg    FileClientConfig
	client pb.FileServiceClient
}

func NewFileClient(cfg FileClientConfig) (*Client, error) {
	conn, err := grpc.Dial(cfg.Addr, grpc.WithInsecure(), grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(10<<20),
		grpc.MaxCallSendMsgSize(10<<20),
	))
	if err != nil {
		return nil, err
	}
	return &Client{
		cfg:    cfg,
		client: pb.NewFileServiceClient(conn),
	}, nil
}

func (c *Client) getctx() (context.Context, context.CancelFunc) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancelFunc()
	md := metadata.New(map[string]string{
		"internal-key": c.cfg.InternalKey,
	})
	return metadata.NewOutgoingContext(ctx, md), cancelFunc
}

func (c *Client) GenerateUploadURL(filename, contentType string) (string, string, error) {
	ctx, cancel := c.getctx()
	defer cancel()
	res, err := c.client.GenerateUploadUrl(
		ctx,
		&pb.UploadRequest{
			Filename:    filename,
			ContentType: contentType,
		},
	)
	if err != nil {
		return "", "", err
	}
	return res.GetPresignedUrl(), res.GetFileId(), nil
}

func (c *Client) GenerateDownloadURL(fileID string) (string, error) {
	ctx, cancel := c.getctx()
	defer cancel()
	res, err := c.client.GenerateDownloadUrl(
		ctx,
		&pb.DownloadRequest{
			FileId: fileID,
		},
	)
	if err != nil {
		return "", err
	}
	return res.GetPresignedUrl(), nil
}
