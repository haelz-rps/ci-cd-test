package constants

type brokerTopics struct {
	CreateConnectorRequests              string
	CreateConnectorResponses             string
	CreateSourceRequests                 string
	CreateSourceResponses                string
	SuccessfulSyncNotificationsRequests  string
	SuccessfulSyncNotificationsResponses string
	GetSourceTablesRequests              string
	GetSourceTablesResponses             string
	UpdateSourceRequests                 string
	UpdateSourceResponses                string
	DeadLetter                           string
}

var BrokerTopics = brokerTopics{
	CreateConnectorRequests:              "create-connector-requests",
	CreateConnectorResponses:             "create-connector-responses",
	CreateSourceRequests:                 "create-source-requests",
	CreateSourceResponses:                "create-source-responses",
	SuccessfulSyncNotificationsRequests:  "successful-sync-notifications-requests",
	SuccessfulSyncNotificationsResponses: "successful-sync-notifications-responses",
	GetSourceTablesRequests:              "get-tables-requests",
	GetSourceTablesResponses:             "get-tables-responses",
	UpdateSourceRequests:                 "update-source-request",
	UpdateSourceResponses:                "update-source-responses",
	DeadLetter:                           "dead-letter",
}

type handlerNames struct {
	ConsumeAsyncResponses                   string
	ConsumeCreateSourceFromDefinitionStatus string
	ConsumeGetSourceTablesStatus            string
	CreateConnector                         string
	CreateSource                            string
	GetConnector                            string
	SuccessfulSyncNotification              string
	GetSourceTables                         string
	UpdateSource                            string
}

var HandlerNames = handlerNames{
	ConsumeAsyncResponses:                   "ConsumeAsyncResponses",
	ConsumeCreateSourceFromDefinitionStatus: "ConsumeCreateSourceFromDefinitionStatus",
	ConsumeGetSourceTablesStatus:            "ConsumeGetSourceTablesStatus",
	CreateConnector:                         "CreateConnector",
	CreateSource:                            "CreateSource",
	GetConnector:                            "GetConnector",
	SuccessfulSyncNotification:              "SuccessfulSyncNotification",
	GetSourceTables:                         "GetSourceTables",
	UpdateSource:                            "UpdateSource",
}
