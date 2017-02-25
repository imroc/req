req
==============
req is a super light weight and super easy-to-use  http request library.

## Useage
``` sh
go get github.com/imroc/req
```

##### Simple GET
``` go
import (
	"fmt"
	"github.com/imroc/req"
)
func main() {
	s, err := req.Get("https://imroc.github.io").String()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(s)
}
```

##### More Control
``` go
import (
	"fmt"
	"github.com/imroc/req"
)
func main() {
	data, err :=
		req.Get("https://www.baidu.com").
			Param("ie", "UTF-8"). // single param
			Params(req.P{         // multiple params
				"f":      "8",
				"rsv_bp": "1",
				"wd":     "go",
			}).
			Header("Accept-Encoding", "gzip"). // set header
			Bytes()

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%s", data)
}
```

##### Post
``` go
import (
	"fmt"
	"github.com/imroc/req"
)
func main() {
	var data Data
	err :=
		req.Post("http://blabla.com").
			Body("abcdefg").
			ToJson(&data)

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%+v", data)
}
```
