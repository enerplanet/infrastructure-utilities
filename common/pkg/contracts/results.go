package contracts

// Result-processing statuses shared between the backend and the geoserver service.
const (
	ResultExtractionPending    = "pending"
	ResultExtractionProcessing = "processing"
	ResultExtractionCompleted  = "completed"
	ResultExtractionFailed     = "failed"

	ResultGeoserverPending    = "pending"
	ResultGeoserverProcessing = "processing"
	ResultGeoserverConfigured = "configured"
	ResultGeoserverFailed     = "failed"
)
