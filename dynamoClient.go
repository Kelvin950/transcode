package main

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoClient struct {
	Table  string
	Client *dynamodb.Client
}

type Ec2TaskState struct {
	State      string    `dynamodbav:"state"`
	StartedAt  time.Time `dynamodbav:"started_at"`
	FinishedAt time.Time `dynamodbav:"finished_at"`
	ErrMsg     string    `dynamodbav:"err_msg"`
	TaskID     string    `dynamodbav:"task_id"`
	Ec2Id      string    `dynamodbav:"ec2_id"`
}

func NewDynamoClient(cfg aws.Config, tablename string) *DynamoClient {

	client := dynamodb.NewFromConfig(cfg)

	return &DynamoClient{
		Table:  tablename,
		Client: client,
	}
}

func (d *DynamoClient) PutITem(item Ec2TaskState) error {
	av, err := attributevalue.MarshalMap(item)

	if err != nil {
		return err
	}

	_, err = d.Client.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName: aws.String(d.Table), Item: av,
	})
	return err
}
