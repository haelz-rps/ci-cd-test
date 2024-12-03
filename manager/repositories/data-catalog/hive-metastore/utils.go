package hivemetastore

import (
	"fmt"

	"github.com/beltran/gohive/hive_metastore"

	dataextractor "github.com/ubix/dripper/manager/repositories/data-extractor"
)

const (
	bucketNameKey = "s3_bucket_name"
	bucketPathKey = "s3_bucket_path"
	pathFormatKey = "s3_path_format"
)

func getLocationUriFromConnector(connector *dataextractor.Connector) (*string, error) {
	locationUri := ""
	protocol := "s3a"
	if connector.DestinationConfiguration == nil {
		return nil, NewMetastoreError("Connector without destination configuration")
	} else if connector.DestinationConfiguration[bucketNameKey] == nil {
		return nil, NewMetastoreError("Connector without destination configuration bucket")
	} else if connector.DestinationConfiguration[bucketPathKey] == nil {
		return nil, NewMetastoreError("Connector without destination configuration bucket path")
	}
	bucket := connector.DestinationConfiguration[bucketNameKey].(string)
	bucketPath := connector.DestinationConfiguration[bucketPathKey].(string)
	dbName := connector.GetWarehouseDatabaseName()
	locationUri = fmt.Sprintf("%s://%s/%s/%s", protocol, bucket, bucketPath, dbName)
	return &locationUri, nil
}

const (
	parquetInputFormat  = "org.apache.hadoop.hive.ql.io.parquet.MapredParquetInputFormat"
	parquetOutputFormat = "org.apache.hadoop.hive.ql.io.parquet.MapredParquetOutputFormat"
	defaultTableOwner   = "dripper"
)

func getDefaultSerDeInfo() *hive_metastore.SerDeInfo {
	return &hive_metastore.SerDeInfo{
		SerializationLib: "org.apache.hadoop.hive.ql.io.parquet.serde.ParquetHiveSerDe",
	}
}

// TODO: get partitions key through regex on path format
func getDefaultPartitionKeys() []*hive_metastore.FieldSchema {
	datePartition := hive_metastore.FieldSchema{
		Name:    "ingestion_date",
		Type:    "string",
		Comment: "date the data was ingested",
	}

	return []*hive_metastore.FieldSchema{&datePartition}
}
