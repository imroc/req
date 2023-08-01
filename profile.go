package req

import "github.com/imroc/req/v3/pkg/http2"

type ClientProfile func(c *Client)

var http2SettingsChrome = []http2.Setting{
	{
		ID:  http2.SettingHeaderTableSize,
		Val: 65536,
	},
	{
		ID:  http2.SettingEnablePush,
		Val: 0,
	},
	{
		ID:  http2.SettingMaxConcurrentStreams,
		Val: 1000,
	},
	{
		ID:  http2.SettingInitialWindowSize,
		Val: 6291456,
	},
	{
		ID:  http2.SettingMaxHeaderListSize,
		Val: 262144,
	},
}

var chromePseudoHeaderOrder = []string{
	":method",
	":authority",
	":scheme",
	":path",
}

var ClientProfile_Chrome ClientProfile = func(c *Client) {
	c.SetTLSFingerprintChrome().
		SetCommonPseudoHeaderOder(chromePseudoHeaderOrder...).
		SetHTTP2SettingsFrame(http2SettingsChrome...)
}
