package es_cli

import (
	"github.com/olivere/elastic"
	"os"
	"fmt"
	"context"
	logx "log"
)



type ESClient struct {
	host string
	client *elastic.Client
}

func NewESClient(host string) (*ESClient, error) {
	cli := &ESClient{host: host}
	err := cli.connect()
	if err != nil {
		return nil, err
	}
	return cli, nil
}

func (cli *ESClient) connect() error {
	errorLog := logx.New(os.Stdout, "APP", logx.LstdFlags)
	var err error
	cli.client, err = elastic.NewClient(elastic.SetErrorLog(errorLog), elastic.SetURL(cli.host))
	if err != nil {
		return err
	}
	info, code, err := cli.client.Ping(cli.host).Do(context.Background())
	if err != nil {
		return err
	}
	fmt.Printf("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)

	version, err := cli.client.ElasticsearchVersion(cli.host)
	if err != nil {
		return err
	}
	fmt.Printf("Elasticsearch version %s\n", version)
	return nil
}

func (cli *ESClient) Update(index, typ, id string, value string) error {
	res, err := cli.client.Index().Index(index).Type(typ).Id(id).BodyJson(value).Do(context.Background())
	fmt.Printf("Update log:%+v %+v\n", res, err)
	if err != nil {
		return err
	}
	return nil
}

func (cli *ESClient) Get(index, typ, id string) (map[string]interface{}, error) {
	res, err := cli.client.Get().Index(index).Type(typ).Id(id).Do(context.Background())
	if err != nil {
		return nil, err
	}
	return res.Fields, nil
}


/*
 Error 400 (Bad Request): Rejecting mapping update to [coins-bty] as the final mapping would have more than 1 type: [coins, 16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp] [type=illegal_argument_exception]
	index/type/id   type 在一个 index 里唯一 在新版本中可能去掉 type
*/


