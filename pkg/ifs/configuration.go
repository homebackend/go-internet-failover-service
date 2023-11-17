package ifs

type Connection struct {
	Name   string `yaml:"name" validate:"required"`
	Rank   int    `yaml:"rank" validate:"required"`
	IP     string `yaml:"ip" validate:"required,ip_addr"`
	PeerIp string `yaml:"peerIp" validate:"required,ip_addr"`
	Mask   uint   `yaml:"mask" validate:"required,gte=1,lte=256"`
	GwIf   string `yaml:"gwIf" validate:"required"`
	GwIp   string `yaml:"gw" validate:"required,ip_addr"`
	Mark   uint   `yaml:"mark" validate:"required,gte=101,lte=200"`
}

type Configuration struct {
	MaxPacketLoss             uint          `yaml:"maxPacketLoss" default:"50" validate:"required,gte=1,lte=100"`
	MinPacketLoss             uint          `yaml:"minPacketLoss" default:"20" validate:"required,gte=1,lte=100"`
	MinSuccessivePacketsRecvd uint          `yaml:"minSuccessivePacketsRecvd" default:"20" validate:"required,gte=1,lte=100"`
	MaxSuccessivePacketsLost  uint          `yaml:"maxSuccessivePacketsLost" default:"10" validate:"required,gte=1,lte=100"`
	UseSudo                   bool          `yaml:"useSudo" default:"false"`
	CleanIfRequired           bool          `yaml:"cleanIfRequired" default:"true"`
	Ping                      string        `yaml:"ping" default:"1.1.1.1" validate:"required,ip_addr"`
	Connections               []*Connection `yaml:"connections" validate:"required,dive"`
}
