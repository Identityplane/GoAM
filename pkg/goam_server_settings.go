package pkg

type GoamServerSettings struct {
	ListenerHttp  string
	ListenerHTTPS string
	TlsCertFile   string
	TlsKeyFile    string
}

func NewGoamServerSettings() *GoamServerSettings {
	return &GoamServerSettings{}
}

func (settings *GoamServerSettings) WithTls(listener, certFile, keyFile string) *GoamServerSettings {

	settings.ListenerHTTPS = listener
	settings.TlsCertFile = certFile
	settings.TlsKeyFile = keyFile

	return settings
}

func (settings *GoamServerSettings) WithListener(listener string) *GoamServerSettings {
	settings.ListenerHttp = listener
	return settings
}
