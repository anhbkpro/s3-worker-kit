package s3infra

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func NewClient(
	ctx context.Context,
	region string,
	endpoint string,
) (*s3.Client, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
		config.WithRetryer(func() aws.Retryer {
			return retry.NewStandard(func(o *retry.StandardOptions) {
				o.MaxAttempts = 5
			})
		}),
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}

	if endpoint != "" {
		return s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = &endpoint
			o.UsePathStyle = true
		}), nil
	}

	return s3.NewFromConfig(cfg), nil
}
