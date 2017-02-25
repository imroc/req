/*
 Usage:

import "github.com/imroc/req"

s, err :=
	req.Get("https://imroc.github.io").
		Param("ie", "UTF-8").
		Params(req.P{
			"category": "tech",
			"wd":       "go",
		}).
		Header("UserAgent", "custom-agent").
		String()
if err != nil {
	fmt.Println(err)
	return
}

*/
package req
