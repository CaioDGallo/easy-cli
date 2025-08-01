package types

type Client struct {
	Name                string
	SanitizedClientName string
	Domain              string
	DatabaseHost        string
	DatabaseUser        string
	BackendBranch       string
	FrontendBranch      string
	SMTPInfo            SMTPInfo
	FrontendInfo        FrontendInfo
	BackendInfo         BackendInfo
}

type BackendInfo struct {
	URL              string
	DatabasePassword string
}

type FrontendInfo struct {
	URL string
}

type SMTPInfo struct {
	Server          string
	Port            string
	Username        string
	Password        string
	DoNotReplyName  string
	DoNotReplyEmail string
	DevEmail        string
}
