package req

import (
	"github.com/imroc/req/v3/http2"
	utls "github.com/refraction-networking/utls"
)

var (
	chromeHttp2Settings = []http2.Setting{
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

	chromePseudoHeaderOrder = []string{
		":method",
		":authority",
		":scheme",
		":path",
	}

	chromeHeaderOrder = []string{
		"host",
		"pragma",
		"cache-control",
		"sec-ch-ua",
		"sec-ch-ua-mobile",
		"sec-ch-ua-platform",
		"upgrade-insecure-requests",
		"user-agent",
		"accept",
		"sec-fetch-site",
		"sec-fetch-mode",
		"sec-fetch-user",
		"sec-fetch-dest",
		"referer",
		"accept-encoding",
		"accept-language",
		"cookie",
	}

	chromeHeaders = map[string]string{
		"pragma":                    "no-cache",
		"cache-control":             "no-cache",
		"sec-ch-ua":                 `"Not_A Brand";v="99", "Google Chrome";v="109", "Chromium";v="109"`,
		"sec-ch-ua-mobile":          "?0",
		"sec-ch-ua-platform":        `"macOS"`,
		"upgrade-insecure-requests": "1",
		"user-agent":                "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36",
		"accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"sec-fetch-site":            "none",
		"sec-fetch-mode":            "navigate",
		"sec-fetch-user":            "?1",
		"sec-fetch-dest":            "document",
		"accept-language":           "zh-CN,zh;q=0.9,en;q=0.8,zh-TW;q=0.7,it;q=0.6",
	}
	chromeHeaderPriority = http2.PriorityParam{
		StreamDep: 0,
		Exclusive: true,
		Weight:    255,
	}
)

// ImpersonateChrome impersonates Chrome browser (version 109).
func (c *Client) ImpersonateChrome() *Client {
	c.
		SetTLSFingerprint(utls.HelloChrome_106_Shuffle). // Chrome 106~109 shares the same tls fingerprint.
		SetHTTP2SettingsFrame(chromeHttp2Settings...).
		SetHTTP2ConnectionFlow(15663105).
		SetCommonPseudoHeaderOder(chromePseudoHeaderOrder...).
		SetCommonHeaderOrder(chromeHeaderOrder...).
		SetCommonHeaders(chromeHeaders).
		SetHTTP2HeaderPriority(chromeHeaderPriority)
	return c
}

var (
	firefoxHttp2Settings = []http2.Setting{
		{
			ID:  http2.SettingHeaderTableSize,
			Val: 65536,
		},
		{
			ID:  http2.SettingInitialWindowSize,
			Val: 131072,
		},
		{
			ID:  http2.SettingMaxFrameSize,
			Val: 16384,
		},
	}
	firefoxPriorityFrames = []http2.PriorityFrame{
		{
			StreamID: 3,
			PriorityParam: http2.PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    200,
			},
		},
		{
			StreamID: 5,
			PriorityParam: http2.PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    100,
			},
		},
		{
			StreamID: 7,
			PriorityParam: http2.PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    0,
			},
		},
		{
			StreamID: 9,
			PriorityParam: http2.PriorityParam{
				StreamDep: 7,
				Exclusive: false,
				Weight:    0,
			},
		},
		{
			StreamID: 11,
			PriorityParam: http2.PriorityParam{
				StreamDep: 3,
				Exclusive: false,
				Weight:    0,
			},
		},
		{
			StreamID: 13,
			PriorityParam: http2.PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    240,
			},
		},
	}
	firefoxPseudoHeaderOrder = []string{
		":method",
		":path",
		":authority",
		":scheme",
	}
	firefoxHeaderOrder = []string{
		"user-agent",
		"accept",
		"accept-language",
		"accept-encoding",
		"referer",
		"cookie",
		"upgrade-insecure-requests",
		"sec-fetch-dest",
		"sec-fetch-mode",
		"sec-fetch-site",
		"sec-fetch-user",
		"te",
	}
	firefoxHeaders = map[string]string{
		"user-agent":                "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:105.0) Gecko/20100101 Firefox/105.0",
		"accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		"accept-language":           "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2",
		"upgrade-insecure-requests": "1",
		"sec-fetch-dest":            "document",
		"sec-fetch-mode":            "navigate",
		"sec-fetch-site":            "same-origin",
		"sec-fetch-user":            "?1",
		//"te":                        "trailers",
	}
	firefoxHeaderPriority = http2.PriorityParam{
		StreamDep: 13,
		Exclusive: false,
		Weight:    41,
	}
)

// ImpersonateFirefox impersonates Firefox browser (version 105).
func (c *Client) ImpersonateFirefox() *Client {
	c.
		SetTLSFingerprint(utls.HelloFirefox_105).
		SetHTTP2SettingsFrame(firefoxHttp2Settings...).
		SetHTTP2ConnectionFlow(12517377).
		SetHTTP2PriorityFrames(firefoxPriorityFrames...).
		SetCommonPseudoHeaderOder(firefoxPseudoHeaderOrder...).
		SetCommonHeaderOrder(firefoxHeaderOrder...).
		SetCommonHeaders(firefoxHeaders).
		SetHTTP2HeaderPriority(firefoxHeaderPriority)
	return c
}

var (
	safariHttp2Settings = []http2.Setting{
		{
			ID:  http2.SettingInitialWindowSize,
			Val: 4194304,
		},
		{
			ID:  http2.SettingMaxConcurrentStreams,
			Val: 100,
		},
	}

	safariPseudoHeaderOrder = []string{
		":method",
		":scheme",
		":path",
		":authority",
	}

	safariHeaderOrder = []string{
		"accept",
		"sec-fetch-site",
		"cookie",
		"sec-fetch-dest",
		"accept-language",
		"sec-fetch-mode",
		"user-agent",
		"referer",
		"accept-encoding",
	}

	safariHeaders = map[string]string{
		"accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"sec-fetch-site":  "same-origin",
		"sec-fetch-dest":  "document",
		"accept-language": "zh-CN,zh-Hans;q=0.9",
		"sec-fetch-mode":  "navigate",
		"user-agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Safari/605.1.15",
	}

	safariHeaderPriority = http2.PriorityParam{
		StreamDep: 0,
		Exclusive: false,
		Weight:    254,
	}
)

// ImpersonateSafari impersonates Safari browser (version 16).
func (c *Client) ImpersonateSafari() *Client {
	c.
		SetTLSFingerprint(utls.HelloSafari_16_0).
		SetHTTP2SettingsFrame(safariHttp2Settings...).
		SetHTTP2ConnectionFlow(10485760).
		SetCommonPseudoHeaderOder(safariPseudoHeaderOrder...).
		SetCommonHeaderOrder(safariHeaderOrder...).
		SetCommonHeaders(safariHeaders).
		SetHTTP2HeaderPriority(safariHeaderPriority)
	return c
}
