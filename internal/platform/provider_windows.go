package platform

func New() Provider { return &unsupportedProvider{os: "windows"} }
