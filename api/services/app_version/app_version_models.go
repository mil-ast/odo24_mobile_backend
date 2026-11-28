package app_version_service

type AppVersionModel struct {
	VersionCode int    `json:"version_code"`
	VersionName string `json:"version_name"`
}

type apkOutputMetadata struct {
	Elements []apkOutputMetadataElement `json:"elements"`
}

type apkOutputMetadataElement struct {
	Type        string `json:"type"`
	VersionCode int    `json:"versionCode"`
	VersionName string `json:"versionName"`
	OutputFile  string `json:"outputFile"`
}
