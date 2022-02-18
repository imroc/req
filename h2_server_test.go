package req

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"github.com/imroc/req/v3/internal/ascii"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/http/httpguts"
	"golang.org/x/net/http2/hpack"
)

// A list of the possible cipher suite ids. Taken from
// https://www.iana.org/assignments/tls-parameters/tls-parameters.txt

const (
	http2cipher_TLS_NULL_WITH_NULL_NULL               uint16 = 0x0000
	http2cipher_TLS_RSA_WITH_NULL_MD5                 uint16 = 0x0001
	http2cipher_TLS_RSA_WITH_NULL_SHA                 uint16 = 0x0002
	http2cipher_TLS_RSA_EXPORT_WITH_RC4_40_MD5        uint16 = 0x0003
	http2cipher_TLS_RSA_WITH_RC4_128_MD5              uint16 = 0x0004
	http2cipher_TLS_RSA_WITH_RC4_128_SHA              uint16 = 0x0005
	http2cipher_TLS_RSA_EXPORT_WITH_RC2_CBC_40_MD5    uint16 = 0x0006
	http2cipher_TLS_RSA_WITH_IDEA_CBC_SHA             uint16 = 0x0007
	http2cipher_TLS_RSA_EXPORT_WITH_DES40_CBC_SHA     uint16 = 0x0008
	http2cipher_TLS_RSA_WITH_DES_CBC_SHA              uint16 = 0x0009
	http2cipher_TLS_RSA_WITH_3DES_EDE_CBC_SHA         uint16 = 0x000A
	http2cipher_TLS_DH_DSS_EXPORT_WITH_DES40_CBC_SHA  uint16 = 0x000B
	http2cipher_TLS_DH_DSS_WITH_DES_CBC_SHA           uint16 = 0x000C
	http2cipher_TLS_DH_DSS_WITH_3DES_EDE_CBC_SHA      uint16 = 0x000D
	http2cipher_TLS_DH_RSA_EXPORT_WITH_DES40_CBC_SHA  uint16 = 0x000E
	http2cipher_TLS_DH_RSA_WITH_DES_CBC_SHA           uint16 = 0x000F
	http2cipher_TLS_DH_RSA_WITH_3DES_EDE_CBC_SHA      uint16 = 0x0010
	http2cipher_TLS_DHE_DSS_EXPORT_WITH_DES40_CBC_SHA uint16 = 0x0011
	http2cipher_TLS_DHE_DSS_WITH_DES_CBC_SHA          uint16 = 0x0012
	http2cipher_TLS_DHE_DSS_WITH_3DES_EDE_CBC_SHA     uint16 = 0x0013
	http2cipher_TLS_DHE_RSA_EXPORT_WITH_DES40_CBC_SHA uint16 = 0x0014
	http2cipher_TLS_DHE_RSA_WITH_DES_CBC_SHA          uint16 = 0x0015
	http2cipher_TLS_DHE_RSA_WITH_3DES_EDE_CBC_SHA     uint16 = 0x0016
	http2cipher_TLS_DH_anon_EXPORT_WITH_RC4_40_MD5    uint16 = 0x0017
	http2cipher_TLS_DH_anon_WITH_RC4_128_MD5          uint16 = 0x0018
	http2cipher_TLS_DH_anon_EXPORT_WITH_DES40_CBC_SHA uint16 = 0x0019
	http2cipher_TLS_DH_anon_WITH_DES_CBC_SHA          uint16 = 0x001A
	http2cipher_TLS_DH_anon_WITH_3DES_EDE_CBC_SHA     uint16 = 0x001B
	// Reserved uint16 =  0x001C-1D
	http2cipher_TLS_KRB5_WITH_DES_CBC_SHA             uint16 = 0x001E
	http2cipher_TLS_KRB5_WITH_3DES_EDE_CBC_SHA        uint16 = 0x001F
	http2cipher_TLS_KRB5_WITH_RC4_128_SHA             uint16 = 0x0020
	http2cipher_TLS_KRB5_WITH_IDEA_CBC_SHA            uint16 = 0x0021
	http2cipher_TLS_KRB5_WITH_DES_CBC_MD5             uint16 = 0x0022
	http2cipher_TLS_KRB5_WITH_3DES_EDE_CBC_MD5        uint16 = 0x0023
	http2cipher_TLS_KRB5_WITH_RC4_128_MD5             uint16 = 0x0024
	http2cipher_TLS_KRB5_WITH_IDEA_CBC_MD5            uint16 = 0x0025
	http2cipher_TLS_KRB5_EXPORT_WITH_DES_CBC_40_SHA   uint16 = 0x0026
	http2cipher_TLS_KRB5_EXPORT_WITH_RC2_CBC_40_SHA   uint16 = 0x0027
	http2cipher_TLS_KRB5_EXPORT_WITH_RC4_40_SHA       uint16 = 0x0028
	http2cipher_TLS_KRB5_EXPORT_WITH_DES_CBC_40_MD5   uint16 = 0x0029
	http2cipher_TLS_KRB5_EXPORT_WITH_RC2_CBC_40_MD5   uint16 = 0x002A
	http2cipher_TLS_KRB5_EXPORT_WITH_RC4_40_MD5       uint16 = 0x002B
	http2cipher_TLS_PSK_WITH_NULL_SHA                 uint16 = 0x002C
	http2cipher_TLS_DHE_PSK_WITH_NULL_SHA             uint16 = 0x002D
	http2cipher_TLS_RSA_PSK_WITH_NULL_SHA             uint16 = 0x002E
	http2cipher_TLS_RSA_WITH_AES_128_CBC_SHA          uint16 = 0x002F
	http2cipher_TLS_DH_DSS_WITH_AES_128_CBC_SHA       uint16 = 0x0030
	http2cipher_TLS_DH_RSA_WITH_AES_128_CBC_SHA       uint16 = 0x0031
	http2cipher_TLS_DHE_DSS_WITH_AES_128_CBC_SHA      uint16 = 0x0032
	http2cipher_TLS_DHE_RSA_WITH_AES_128_CBC_SHA      uint16 = 0x0033
	http2cipher_TLS_DH_anon_WITH_AES_128_CBC_SHA      uint16 = 0x0034
	http2cipher_TLS_RSA_WITH_AES_256_CBC_SHA          uint16 = 0x0035
	http2cipher_TLS_DH_DSS_WITH_AES_256_CBC_SHA       uint16 = 0x0036
	http2cipher_TLS_DH_RSA_WITH_AES_256_CBC_SHA       uint16 = 0x0037
	http2cipher_TLS_DHE_DSS_WITH_AES_256_CBC_SHA      uint16 = 0x0038
	http2cipher_TLS_DHE_RSA_WITH_AES_256_CBC_SHA      uint16 = 0x0039
	http2cipher_TLS_DH_anon_WITH_AES_256_CBC_SHA      uint16 = 0x003A
	http2cipher_TLS_RSA_WITH_NULL_SHA256              uint16 = 0x003B
	http2cipher_TLS_RSA_WITH_AES_128_CBC_SHA256       uint16 = 0x003C
	http2cipher_TLS_RSA_WITH_AES_256_CBC_SHA256       uint16 = 0x003D
	http2cipher_TLS_DH_DSS_WITH_AES_128_CBC_SHA256    uint16 = 0x003E
	http2cipher_TLS_DH_RSA_WITH_AES_128_CBC_SHA256    uint16 = 0x003F
	http2cipher_TLS_DHE_DSS_WITH_AES_128_CBC_SHA256   uint16 = 0x0040
	http2cipher_TLS_RSA_WITH_CAMELLIA_128_CBC_SHA     uint16 = 0x0041
	http2cipher_TLS_DH_DSS_WITH_CAMELLIA_128_CBC_SHA  uint16 = 0x0042
	http2cipher_TLS_DH_RSA_WITH_CAMELLIA_128_CBC_SHA  uint16 = 0x0043
	http2cipher_TLS_DHE_DSS_WITH_CAMELLIA_128_CBC_SHA uint16 = 0x0044
	http2cipher_TLS_DHE_RSA_WITH_CAMELLIA_128_CBC_SHA uint16 = 0x0045
	http2cipher_TLS_DH_anon_WITH_CAMELLIA_128_CBC_SHA uint16 = 0x0046
	// Reserved uint16 =  0x0047-4F
	// Reserved uint16 =  0x0050-58
	// Reserved uint16 =  0x0059-5C
	// Unassigned uint16 =  0x005D-5F
	// Reserved uint16 =  0x0060-66
	http2cipher_TLS_DHE_RSA_WITH_AES_128_CBC_SHA256 uint16 = 0x0067
	http2cipher_TLS_DH_DSS_WITH_AES_256_CBC_SHA256  uint16 = 0x0068
	http2cipher_TLS_DH_RSA_WITH_AES_256_CBC_SHA256  uint16 = 0x0069
	http2cipher_TLS_DHE_DSS_WITH_AES_256_CBC_SHA256 uint16 = 0x006A
	http2cipher_TLS_DHE_RSA_WITH_AES_256_CBC_SHA256 uint16 = 0x006B
	http2cipher_TLS_DH_anon_WITH_AES_128_CBC_SHA256 uint16 = 0x006C
	http2cipher_TLS_DH_anon_WITH_AES_256_CBC_SHA256 uint16 = 0x006D
	// Unassigned uint16 =  0x006E-83
	http2cipher_TLS_RSA_WITH_CAMELLIA_256_CBC_SHA        uint16 = 0x0084
	http2cipher_TLS_DH_DSS_WITH_CAMELLIA_256_CBC_SHA     uint16 = 0x0085
	http2cipher_TLS_DH_RSA_WITH_CAMELLIA_256_CBC_SHA     uint16 = 0x0086
	http2cipher_TLS_DHE_DSS_WITH_CAMELLIA_256_CBC_SHA    uint16 = 0x0087
	http2cipher_TLS_DHE_RSA_WITH_CAMELLIA_256_CBC_SHA    uint16 = 0x0088
	http2cipher_TLS_DH_anon_WITH_CAMELLIA_256_CBC_SHA    uint16 = 0x0089
	http2cipher_TLS_PSK_WITH_RC4_128_SHA                 uint16 = 0x008A
	http2cipher_TLS_PSK_WITH_3DES_EDE_CBC_SHA            uint16 = 0x008B
	http2cipher_TLS_PSK_WITH_AES_128_CBC_SHA             uint16 = 0x008C
	http2cipher_TLS_PSK_WITH_AES_256_CBC_SHA             uint16 = 0x008D
	http2cipher_TLS_DHE_PSK_WITH_RC4_128_SHA             uint16 = 0x008E
	http2cipher_TLS_DHE_PSK_WITH_3DES_EDE_CBC_SHA        uint16 = 0x008F
	http2cipher_TLS_DHE_PSK_WITH_AES_128_CBC_SHA         uint16 = 0x0090
	http2cipher_TLS_DHE_PSK_WITH_AES_256_CBC_SHA         uint16 = 0x0091
	http2cipher_TLS_RSA_PSK_WITH_RC4_128_SHA             uint16 = 0x0092
	http2cipher_TLS_RSA_PSK_WITH_3DES_EDE_CBC_SHA        uint16 = 0x0093
	http2cipher_TLS_RSA_PSK_WITH_AES_128_CBC_SHA         uint16 = 0x0094
	http2cipher_TLS_RSA_PSK_WITH_AES_256_CBC_SHA         uint16 = 0x0095
	http2cipher_TLS_RSA_WITH_SEED_CBC_SHA                uint16 = 0x0096
	http2cipher_TLS_DH_DSS_WITH_SEED_CBC_SHA             uint16 = 0x0097
	http2cipher_TLS_DH_RSA_WITH_SEED_CBC_SHA             uint16 = 0x0098
	http2cipher_TLS_DHE_DSS_WITH_SEED_CBC_SHA            uint16 = 0x0099
	http2cipher_TLS_DHE_RSA_WITH_SEED_CBC_SHA            uint16 = 0x009A
	http2cipher_TLS_DH_anon_WITH_SEED_CBC_SHA            uint16 = 0x009B
	http2cipher_TLS_RSA_WITH_AES_128_GCM_SHA256          uint16 = 0x009C
	http2cipher_TLS_RSA_WITH_AES_256_GCM_SHA384          uint16 = 0x009D
	http2cipher_TLS_DHE_RSA_WITH_AES_128_GCM_SHA256      uint16 = 0x009E
	http2cipher_TLS_DHE_RSA_WITH_AES_256_GCM_SHA384      uint16 = 0x009F
	http2cipher_TLS_DH_RSA_WITH_AES_128_GCM_SHA256       uint16 = 0x00A0
	http2cipher_TLS_DH_RSA_WITH_AES_256_GCM_SHA384       uint16 = 0x00A1
	http2cipher_TLS_DHE_DSS_WITH_AES_128_GCM_SHA256      uint16 = 0x00A2
	http2cipher_TLS_DHE_DSS_WITH_AES_256_GCM_SHA384      uint16 = 0x00A3
	http2cipher_TLS_DH_DSS_WITH_AES_128_GCM_SHA256       uint16 = 0x00A4
	http2cipher_TLS_DH_DSS_WITH_AES_256_GCM_SHA384       uint16 = 0x00A5
	http2cipher_TLS_DH_anon_WITH_AES_128_GCM_SHA256      uint16 = 0x00A6
	http2cipher_TLS_DH_anon_WITH_AES_256_GCM_SHA384      uint16 = 0x00A7
	http2cipher_TLS_PSK_WITH_AES_128_GCM_SHA256          uint16 = 0x00A8
	http2cipher_TLS_PSK_WITH_AES_256_GCM_SHA384          uint16 = 0x00A9
	http2cipher_TLS_DHE_PSK_WITH_AES_128_GCM_SHA256      uint16 = 0x00AA
	http2cipher_TLS_DHE_PSK_WITH_AES_256_GCM_SHA384      uint16 = 0x00AB
	http2cipher_TLS_RSA_PSK_WITH_AES_128_GCM_SHA256      uint16 = 0x00AC
	http2cipher_TLS_RSA_PSK_WITH_AES_256_GCM_SHA384      uint16 = 0x00AD
	http2cipher_TLS_PSK_WITH_AES_128_CBC_SHA256          uint16 = 0x00AE
	http2cipher_TLS_PSK_WITH_AES_256_CBC_SHA384          uint16 = 0x00AF
	http2cipher_TLS_PSK_WITH_NULL_SHA256                 uint16 = 0x00B0
	http2cipher_TLS_PSK_WITH_NULL_SHA384                 uint16 = 0x00B1
	http2cipher_TLS_DHE_PSK_WITH_AES_128_CBC_SHA256      uint16 = 0x00B2
	http2cipher_TLS_DHE_PSK_WITH_AES_256_CBC_SHA384      uint16 = 0x00B3
	http2cipher_TLS_DHE_PSK_WITH_NULL_SHA256             uint16 = 0x00B4
	http2cipher_TLS_DHE_PSK_WITH_NULL_SHA384             uint16 = 0x00B5
	http2cipher_TLS_RSA_PSK_WITH_AES_128_CBC_SHA256      uint16 = 0x00B6
	http2cipher_TLS_RSA_PSK_WITH_AES_256_CBC_SHA384      uint16 = 0x00B7
	http2cipher_TLS_RSA_PSK_WITH_NULL_SHA256             uint16 = 0x00B8
	http2cipher_TLS_RSA_PSK_WITH_NULL_SHA384             uint16 = 0x00B9
	http2cipher_TLS_RSA_WITH_CAMELLIA_128_CBC_SHA256     uint16 = 0x00BA
	http2cipher_TLS_DH_DSS_WITH_CAMELLIA_128_CBC_SHA256  uint16 = 0x00BB
	http2cipher_TLS_DH_RSA_WITH_CAMELLIA_128_CBC_SHA256  uint16 = 0x00BC
	http2cipher_TLS_DHE_DSS_WITH_CAMELLIA_128_CBC_SHA256 uint16 = 0x00BD
	http2cipher_TLS_DHE_RSA_WITH_CAMELLIA_128_CBC_SHA256 uint16 = 0x00BE
	http2cipher_TLS_DH_anon_WITH_CAMELLIA_128_CBC_SHA256 uint16 = 0x00BF
	http2cipher_TLS_RSA_WITH_CAMELLIA_256_CBC_SHA256     uint16 = 0x00C0
	http2cipher_TLS_DH_DSS_WITH_CAMELLIA_256_CBC_SHA256  uint16 = 0x00C1
	http2cipher_TLS_DH_RSA_WITH_CAMELLIA_256_CBC_SHA256  uint16 = 0x00C2
	http2cipher_TLS_DHE_DSS_WITH_CAMELLIA_256_CBC_SHA256 uint16 = 0x00C3
	http2cipher_TLS_DHE_RSA_WITH_CAMELLIA_256_CBC_SHA256 uint16 = 0x00C4
	http2cipher_TLS_DH_anon_WITH_CAMELLIA_256_CBC_SHA256 uint16 = 0x00C5
	// Unassigned uint16 =  0x00C6-FE
	http2cipher_TLS_EMPTY_RENEGOTIATION_INFO_SCSV uint16 = 0x00FF
	// Unassigned uint16 =  0x01-55,*
	http2cipher_TLS_FALLBACK_SCSV uint16 = 0x5600
	// Unassigned                                   uint16 = 0x5601 - 0xC000
	http2cipher_TLS_ECDH_ECDSA_WITH_NULL_SHA                 uint16 = 0xC001
	http2cipher_TLS_ECDH_ECDSA_WITH_RC4_128_SHA              uint16 = 0xC002
	http2cipher_TLS_ECDH_ECDSA_WITH_3DES_EDE_CBC_SHA         uint16 = 0xC003
	http2cipher_TLS_ECDH_ECDSA_WITH_AES_128_CBC_SHA          uint16 = 0xC004
	http2cipher_TLS_ECDH_ECDSA_WITH_AES_256_CBC_SHA          uint16 = 0xC005
	http2cipher_TLS_ECDHE_ECDSA_WITH_NULL_SHA                uint16 = 0xC006
	http2cipher_TLS_ECDHE_ECDSA_WITH_RC4_128_SHA             uint16 = 0xC007
	http2cipher_TLS_ECDHE_ECDSA_WITH_3DES_EDE_CBC_SHA        uint16 = 0xC008
	http2cipher_TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA         uint16 = 0xC009
	http2cipher_TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA         uint16 = 0xC00A
	http2cipher_TLS_ECDH_RSA_WITH_NULL_SHA                   uint16 = 0xC00B
	http2cipher_TLS_ECDH_RSA_WITH_RC4_128_SHA                uint16 = 0xC00C
	http2cipher_TLS_ECDH_RSA_WITH_3DES_EDE_CBC_SHA           uint16 = 0xC00D
	http2cipher_TLS_ECDH_RSA_WITH_AES_128_CBC_SHA            uint16 = 0xC00E
	http2cipher_TLS_ECDH_RSA_WITH_AES_256_CBC_SHA            uint16 = 0xC00F
	http2cipher_TLS_ECDHE_RSA_WITH_NULL_SHA                  uint16 = 0xC010
	http2cipher_TLS_ECDHE_RSA_WITH_RC4_128_SHA               uint16 = 0xC011
	http2cipher_TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA          uint16 = 0xC012
	http2cipher_TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA           uint16 = 0xC013
	http2cipher_TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA           uint16 = 0xC014
	http2cipher_TLS_ECDH_anon_WITH_NULL_SHA                  uint16 = 0xC015
	http2cipher_TLS_ECDH_anon_WITH_RC4_128_SHA               uint16 = 0xC016
	http2cipher_TLS_ECDH_anon_WITH_3DES_EDE_CBC_SHA          uint16 = 0xC017
	http2cipher_TLS_ECDH_anon_WITH_AES_128_CBC_SHA           uint16 = 0xC018
	http2cipher_TLS_ECDH_anon_WITH_AES_256_CBC_SHA           uint16 = 0xC019
	http2cipher_TLS_SRP_SHA_WITH_3DES_EDE_CBC_SHA            uint16 = 0xC01A
	http2cipher_TLS_SRP_SHA_RSA_WITH_3DES_EDE_CBC_SHA        uint16 = 0xC01B
	http2cipher_TLS_SRP_SHA_DSS_WITH_3DES_EDE_CBC_SHA        uint16 = 0xC01C
	http2cipher_TLS_SRP_SHA_WITH_AES_128_CBC_SHA             uint16 = 0xC01D
	http2cipher_TLS_SRP_SHA_RSA_WITH_AES_128_CBC_SHA         uint16 = 0xC01E
	http2cipher_TLS_SRP_SHA_DSS_WITH_AES_128_CBC_SHA         uint16 = 0xC01F
	http2cipher_TLS_SRP_SHA_WITH_AES_256_CBC_SHA             uint16 = 0xC020
	http2cipher_TLS_SRP_SHA_RSA_WITH_AES_256_CBC_SHA         uint16 = 0xC021
	http2cipher_TLS_SRP_SHA_DSS_WITH_AES_256_CBC_SHA         uint16 = 0xC022
	http2cipher_TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256      uint16 = 0xC023
	http2cipher_TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA384      uint16 = 0xC024
	http2cipher_TLS_ECDH_ECDSA_WITH_AES_128_CBC_SHA256       uint16 = 0xC025
	http2cipher_TLS_ECDH_ECDSA_WITH_AES_256_CBC_SHA384       uint16 = 0xC026
	http2cipher_TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256        uint16 = 0xC027
	http2cipher_TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA384        uint16 = 0xC028
	http2cipher_TLS_ECDH_RSA_WITH_AES_128_CBC_SHA256         uint16 = 0xC029
	http2cipher_TLS_ECDH_RSA_WITH_AES_256_CBC_SHA384         uint16 = 0xC02A
	http2cipher_TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256      uint16 = 0xC02B
	http2cipher_TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384      uint16 = 0xC02C
	http2cipher_TLS_ECDH_ECDSA_WITH_AES_128_GCM_SHA256       uint16 = 0xC02D
	http2cipher_TLS_ECDH_ECDSA_WITH_AES_256_GCM_SHA384       uint16 = 0xC02E
	http2cipher_TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256        uint16 = 0xC02F
	http2cipher_TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384        uint16 = 0xC030
	http2cipher_TLS_ECDH_RSA_WITH_AES_128_GCM_SHA256         uint16 = 0xC031
	http2cipher_TLS_ECDH_RSA_WITH_AES_256_GCM_SHA384         uint16 = 0xC032
	http2cipher_TLS_ECDHE_PSK_WITH_RC4_128_SHA               uint16 = 0xC033
	http2cipher_TLS_ECDHE_PSK_WITH_3DES_EDE_CBC_SHA          uint16 = 0xC034
	http2cipher_TLS_ECDHE_PSK_WITH_AES_128_CBC_SHA           uint16 = 0xC035
	http2cipher_TLS_ECDHE_PSK_WITH_AES_256_CBC_SHA           uint16 = 0xC036
	http2cipher_TLS_ECDHE_PSK_WITH_AES_128_CBC_SHA256        uint16 = 0xC037
	http2cipher_TLS_ECDHE_PSK_WITH_AES_256_CBC_SHA384        uint16 = 0xC038
	http2cipher_TLS_ECDHE_PSK_WITH_NULL_SHA                  uint16 = 0xC039
	http2cipher_TLS_ECDHE_PSK_WITH_NULL_SHA256               uint16 = 0xC03A
	http2cipher_TLS_ECDHE_PSK_WITH_NULL_SHA384               uint16 = 0xC03B
	http2cipher_TLS_RSA_WITH_ARIA_128_CBC_SHA256             uint16 = 0xC03C
	http2cipher_TLS_RSA_WITH_ARIA_256_CBC_SHA384             uint16 = 0xC03D
	http2cipher_TLS_DH_DSS_WITH_ARIA_128_CBC_SHA256          uint16 = 0xC03E
	http2cipher_TLS_DH_DSS_WITH_ARIA_256_CBC_SHA384          uint16 = 0xC03F
	http2cipher_TLS_DH_RSA_WITH_ARIA_128_CBC_SHA256          uint16 = 0xC040
	http2cipher_TLS_DH_RSA_WITH_ARIA_256_CBC_SHA384          uint16 = 0xC041
	http2cipher_TLS_DHE_DSS_WITH_ARIA_128_CBC_SHA256         uint16 = 0xC042
	http2cipher_TLS_DHE_DSS_WITH_ARIA_256_CBC_SHA384         uint16 = 0xC043
	http2cipher_TLS_DHE_RSA_WITH_ARIA_128_CBC_SHA256         uint16 = 0xC044
	http2cipher_TLS_DHE_RSA_WITH_ARIA_256_CBC_SHA384         uint16 = 0xC045
	http2cipher_TLS_DH_anon_WITH_ARIA_128_CBC_SHA256         uint16 = 0xC046
	http2cipher_TLS_DH_anon_WITH_ARIA_256_CBC_SHA384         uint16 = 0xC047
	http2cipher_TLS_ECDHE_ECDSA_WITH_ARIA_128_CBC_SHA256     uint16 = 0xC048
	http2cipher_TLS_ECDHE_ECDSA_WITH_ARIA_256_CBC_SHA384     uint16 = 0xC049
	http2cipher_TLS_ECDH_ECDSA_WITH_ARIA_128_CBC_SHA256      uint16 = 0xC04A
	http2cipher_TLS_ECDH_ECDSA_WITH_ARIA_256_CBC_SHA384      uint16 = 0xC04B
	http2cipher_TLS_ECDHE_RSA_WITH_ARIA_128_CBC_SHA256       uint16 = 0xC04C
	http2cipher_TLS_ECDHE_RSA_WITH_ARIA_256_CBC_SHA384       uint16 = 0xC04D
	http2cipher_TLS_ECDH_RSA_WITH_ARIA_128_CBC_SHA256        uint16 = 0xC04E
	http2cipher_TLS_ECDH_RSA_WITH_ARIA_256_CBC_SHA384        uint16 = 0xC04F
	http2cipher_TLS_RSA_WITH_ARIA_128_GCM_SHA256             uint16 = 0xC050
	http2cipher_TLS_RSA_WITH_ARIA_256_GCM_SHA384             uint16 = 0xC051
	http2cipher_TLS_DHE_RSA_WITH_ARIA_128_GCM_SHA256         uint16 = 0xC052
	http2cipher_TLS_DHE_RSA_WITH_ARIA_256_GCM_SHA384         uint16 = 0xC053
	http2cipher_TLS_DH_RSA_WITH_ARIA_128_GCM_SHA256          uint16 = 0xC054
	http2cipher_TLS_DH_RSA_WITH_ARIA_256_GCM_SHA384          uint16 = 0xC055
	http2cipher_TLS_DHE_DSS_WITH_ARIA_128_GCM_SHA256         uint16 = 0xC056
	http2cipher_TLS_DHE_DSS_WITH_ARIA_256_GCM_SHA384         uint16 = 0xC057
	http2cipher_TLS_DH_DSS_WITH_ARIA_128_GCM_SHA256          uint16 = 0xC058
	http2cipher_TLS_DH_DSS_WITH_ARIA_256_GCM_SHA384          uint16 = 0xC059
	http2cipher_TLS_DH_anon_WITH_ARIA_128_GCM_SHA256         uint16 = 0xC05A
	http2cipher_TLS_DH_anon_WITH_ARIA_256_GCM_SHA384         uint16 = 0xC05B
	http2cipher_TLS_ECDHE_ECDSA_WITH_ARIA_128_GCM_SHA256     uint16 = 0xC05C
	http2cipher_TLS_ECDHE_ECDSA_WITH_ARIA_256_GCM_SHA384     uint16 = 0xC05D
	http2cipher_TLS_ECDH_ECDSA_WITH_ARIA_128_GCM_SHA256      uint16 = 0xC05E
	http2cipher_TLS_ECDH_ECDSA_WITH_ARIA_256_GCM_SHA384      uint16 = 0xC05F
	http2cipher_TLS_ECDHE_RSA_WITH_ARIA_128_GCM_SHA256       uint16 = 0xC060
	http2cipher_TLS_ECDHE_RSA_WITH_ARIA_256_GCM_SHA384       uint16 = 0xC061
	http2cipher_TLS_ECDH_RSA_WITH_ARIA_128_GCM_SHA256        uint16 = 0xC062
	http2cipher_TLS_ECDH_RSA_WITH_ARIA_256_GCM_SHA384        uint16 = 0xC063
	http2cipher_TLS_PSK_WITH_ARIA_128_CBC_SHA256             uint16 = 0xC064
	http2cipher_TLS_PSK_WITH_ARIA_256_CBC_SHA384             uint16 = 0xC065
	http2cipher_TLS_DHE_PSK_WITH_ARIA_128_CBC_SHA256         uint16 = 0xC066
	http2cipher_TLS_DHE_PSK_WITH_ARIA_256_CBC_SHA384         uint16 = 0xC067
	http2cipher_TLS_RSA_PSK_WITH_ARIA_128_CBC_SHA256         uint16 = 0xC068
	http2cipher_TLS_RSA_PSK_WITH_ARIA_256_CBC_SHA384         uint16 = 0xC069
	http2cipher_TLS_PSK_WITH_ARIA_128_GCM_SHA256             uint16 = 0xC06A
	http2cipher_TLS_PSK_WITH_ARIA_256_GCM_SHA384             uint16 = 0xC06B
	http2cipher_TLS_DHE_PSK_WITH_ARIA_128_GCM_SHA256         uint16 = 0xC06C
	http2cipher_TLS_DHE_PSK_WITH_ARIA_256_GCM_SHA384         uint16 = 0xC06D
	http2cipher_TLS_RSA_PSK_WITH_ARIA_128_GCM_SHA256         uint16 = 0xC06E
	http2cipher_TLS_RSA_PSK_WITH_ARIA_256_GCM_SHA384         uint16 = 0xC06F
	http2cipher_TLS_ECDHE_PSK_WITH_ARIA_128_CBC_SHA256       uint16 = 0xC070
	http2cipher_TLS_ECDHE_PSK_WITH_ARIA_256_CBC_SHA384       uint16 = 0xC071
	http2cipher_TLS_ECDHE_ECDSA_WITH_CAMELLIA_128_CBC_SHA256 uint16 = 0xC072
	http2cipher_TLS_ECDHE_ECDSA_WITH_CAMELLIA_256_CBC_SHA384 uint16 = 0xC073
	http2cipher_TLS_ECDH_ECDSA_WITH_CAMELLIA_128_CBC_SHA256  uint16 = 0xC074
	http2cipher_TLS_ECDH_ECDSA_WITH_CAMELLIA_256_CBC_SHA384  uint16 = 0xC075
	http2cipher_TLS_ECDHE_RSA_WITH_CAMELLIA_128_CBC_SHA256   uint16 = 0xC076
	http2cipher_TLS_ECDHE_RSA_WITH_CAMELLIA_256_CBC_SHA384   uint16 = 0xC077
	http2cipher_TLS_ECDH_RSA_WITH_CAMELLIA_128_CBC_SHA256    uint16 = 0xC078
	http2cipher_TLS_ECDH_RSA_WITH_CAMELLIA_256_CBC_SHA384    uint16 = 0xC079
	http2cipher_TLS_RSA_WITH_CAMELLIA_128_GCM_SHA256         uint16 = 0xC07A
	http2cipher_TLS_RSA_WITH_CAMELLIA_256_GCM_SHA384         uint16 = 0xC07B
	http2cipher_TLS_DHE_RSA_WITH_CAMELLIA_128_GCM_SHA256     uint16 = 0xC07C
	http2cipher_TLS_DHE_RSA_WITH_CAMELLIA_256_GCM_SHA384     uint16 = 0xC07D
	http2cipher_TLS_DH_RSA_WITH_CAMELLIA_128_GCM_SHA256      uint16 = 0xC07E
	http2cipher_TLS_DH_RSA_WITH_CAMELLIA_256_GCM_SHA384      uint16 = 0xC07F
	http2cipher_TLS_DHE_DSS_WITH_CAMELLIA_128_GCM_SHA256     uint16 = 0xC080
	http2cipher_TLS_DHE_DSS_WITH_CAMELLIA_256_GCM_SHA384     uint16 = 0xC081
	http2cipher_TLS_DH_DSS_WITH_CAMELLIA_128_GCM_SHA256      uint16 = 0xC082
	http2cipher_TLS_DH_DSS_WITH_CAMELLIA_256_GCM_SHA384      uint16 = 0xC083
	http2cipher_TLS_DH_anon_WITH_CAMELLIA_128_GCM_SHA256     uint16 = 0xC084
	http2cipher_TLS_DH_anon_WITH_CAMELLIA_256_GCM_SHA384     uint16 = 0xC085
	http2cipher_TLS_ECDHE_ECDSA_WITH_CAMELLIA_128_GCM_SHA256 uint16 = 0xC086
	http2cipher_TLS_ECDHE_ECDSA_WITH_CAMELLIA_256_GCM_SHA384 uint16 = 0xC087
	http2cipher_TLS_ECDH_ECDSA_WITH_CAMELLIA_128_GCM_SHA256  uint16 = 0xC088
	http2cipher_TLS_ECDH_ECDSA_WITH_CAMELLIA_256_GCM_SHA384  uint16 = 0xC089
	http2cipher_TLS_ECDHE_RSA_WITH_CAMELLIA_128_GCM_SHA256   uint16 = 0xC08A
	http2cipher_TLS_ECDHE_RSA_WITH_CAMELLIA_256_GCM_SHA384   uint16 = 0xC08B
	http2cipher_TLS_ECDH_RSA_WITH_CAMELLIA_128_GCM_SHA256    uint16 = 0xC08C
	http2cipher_TLS_ECDH_RSA_WITH_CAMELLIA_256_GCM_SHA384    uint16 = 0xC08D
	http2cipher_TLS_PSK_WITH_CAMELLIA_128_GCM_SHA256         uint16 = 0xC08E
	http2cipher_TLS_PSK_WITH_CAMELLIA_256_GCM_SHA384         uint16 = 0xC08F
	http2cipher_TLS_DHE_PSK_WITH_CAMELLIA_128_GCM_SHA256     uint16 = 0xC090
	http2cipher_TLS_DHE_PSK_WITH_CAMELLIA_256_GCM_SHA384     uint16 = 0xC091
	http2cipher_TLS_RSA_PSK_WITH_CAMELLIA_128_GCM_SHA256     uint16 = 0xC092
	http2cipher_TLS_RSA_PSK_WITH_CAMELLIA_256_GCM_SHA384     uint16 = 0xC093
	http2cipher_TLS_PSK_WITH_CAMELLIA_128_CBC_SHA256         uint16 = 0xC094
	http2cipher_TLS_PSK_WITH_CAMELLIA_256_CBC_SHA384         uint16 = 0xC095
	http2cipher_TLS_DHE_PSK_WITH_CAMELLIA_128_CBC_SHA256     uint16 = 0xC096
	http2cipher_TLS_DHE_PSK_WITH_CAMELLIA_256_CBC_SHA384     uint16 = 0xC097
	http2cipher_TLS_RSA_PSK_WITH_CAMELLIA_128_CBC_SHA256     uint16 = 0xC098
	http2cipher_TLS_RSA_PSK_WITH_CAMELLIA_256_CBC_SHA384     uint16 = 0xC099
	http2cipher_TLS_ECDHE_PSK_WITH_CAMELLIA_128_CBC_SHA256   uint16 = 0xC09A
	http2cipher_TLS_ECDHE_PSK_WITH_CAMELLIA_256_CBC_SHA384   uint16 = 0xC09B
	http2cipher_TLS_RSA_WITH_AES_128_CCM                     uint16 = 0xC09C
	http2cipher_TLS_RSA_WITH_AES_256_CCM                     uint16 = 0xC09D
	http2cipher_TLS_DHE_RSA_WITH_AES_128_CCM                 uint16 = 0xC09E
	http2cipher_TLS_DHE_RSA_WITH_AES_256_CCM                 uint16 = 0xC09F
	http2cipher_TLS_RSA_WITH_AES_128_CCM_8                   uint16 = 0xC0A0
	http2cipher_TLS_RSA_WITH_AES_256_CCM_8                   uint16 = 0xC0A1
	http2cipher_TLS_DHE_RSA_WITH_AES_128_CCM_8               uint16 = 0xC0A2
	http2cipher_TLS_DHE_RSA_WITH_AES_256_CCM_8               uint16 = 0xC0A3
	http2cipher_TLS_PSK_WITH_AES_128_CCM                     uint16 = 0xC0A4
	http2cipher_TLS_PSK_WITH_AES_256_CCM                     uint16 = 0xC0A5
	http2cipher_TLS_DHE_PSK_WITH_AES_128_CCM                 uint16 = 0xC0A6
	http2cipher_TLS_DHE_PSK_WITH_AES_256_CCM                 uint16 = 0xC0A7
	http2cipher_TLS_PSK_WITH_AES_128_CCM_8                   uint16 = 0xC0A8
	http2cipher_TLS_PSK_WITH_AES_256_CCM_8                   uint16 = 0xC0A9
	http2cipher_TLS_PSK_DHE_WITH_AES_128_CCM_8               uint16 = 0xC0AA
	http2cipher_TLS_PSK_DHE_WITH_AES_256_CCM_8               uint16 = 0xC0AB
	http2cipher_TLS_ECDHE_ECDSA_WITH_AES_128_CCM             uint16 = 0xC0AC
	http2cipher_TLS_ECDHE_ECDSA_WITH_AES_256_CCM             uint16 = 0xC0AD
	http2cipher_TLS_ECDHE_ECDSA_WITH_AES_128_CCM_8           uint16 = 0xC0AE
	http2cipher_TLS_ECDHE_ECDSA_WITH_AES_256_CCM_8           uint16 = 0xC0AF
	// Unassigned uint16 =  0xC0B0-FF
	// Unassigned uint16 =  0xC1-CB,*
	// Unassigned uint16 =  0xCC00-A7
	http2cipher_TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256   uint16 = 0xCCA8
	http2cipher_TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256 uint16 = 0xCCA9
	http2cipher_TLS_DHE_RSA_WITH_CHACHA20_POLY1305_SHA256     uint16 = 0xCCAA
	http2cipher_TLS_PSK_WITH_CHACHA20_POLY1305_SHA256         uint16 = 0xCCAB
	http2cipher_TLS_ECDHE_PSK_WITH_CHACHA20_POLY1305_SHA256   uint16 = 0xCCAC
	http2cipher_TLS_DHE_PSK_WITH_CHACHA20_POLY1305_SHA256     uint16 = 0xCCAD
	http2cipher_TLS_RSA_PSK_WITH_CHACHA20_POLY1305_SHA256     uint16 = 0xCCAE
)

// isBadCipher reports whether the cipher is blacklisted by the HTTP/2 spec.
// References:
// https://tools.ietf.org/html/rfc7540#appendix-A
// Reject cipher suites from Appendix A.
// "This list includes those cipher suites that do not
// offer an ephemeral key exchange and those that are
// based on the TLS null, stream or block cipher type"
func http2isBadCipher(cipher uint16) bool {
	switch cipher {
	case http2cipher_TLS_NULL_WITH_NULL_NULL,
		http2cipher_TLS_RSA_WITH_NULL_MD5,
		http2cipher_TLS_RSA_WITH_NULL_SHA,
		http2cipher_TLS_RSA_EXPORT_WITH_RC4_40_MD5,
		http2cipher_TLS_RSA_WITH_RC4_128_MD5,
		http2cipher_TLS_RSA_WITH_RC4_128_SHA,
		http2cipher_TLS_RSA_EXPORT_WITH_RC2_CBC_40_MD5,
		http2cipher_TLS_RSA_WITH_IDEA_CBC_SHA,
		http2cipher_TLS_RSA_EXPORT_WITH_DES40_CBC_SHA,
		http2cipher_TLS_RSA_WITH_DES_CBC_SHA,
		http2cipher_TLS_RSA_WITH_3DES_EDE_CBC_SHA,
		http2cipher_TLS_DH_DSS_EXPORT_WITH_DES40_CBC_SHA,
		http2cipher_TLS_DH_DSS_WITH_DES_CBC_SHA,
		http2cipher_TLS_DH_DSS_WITH_3DES_EDE_CBC_SHA,
		http2cipher_TLS_DH_RSA_EXPORT_WITH_DES40_CBC_SHA,
		http2cipher_TLS_DH_RSA_WITH_DES_CBC_SHA,
		http2cipher_TLS_DH_RSA_WITH_3DES_EDE_CBC_SHA,
		http2cipher_TLS_DHE_DSS_EXPORT_WITH_DES40_CBC_SHA,
		http2cipher_TLS_DHE_DSS_WITH_DES_CBC_SHA,
		http2cipher_TLS_DHE_DSS_WITH_3DES_EDE_CBC_SHA,
		http2cipher_TLS_DHE_RSA_EXPORT_WITH_DES40_CBC_SHA,
		http2cipher_TLS_DHE_RSA_WITH_DES_CBC_SHA,
		http2cipher_TLS_DHE_RSA_WITH_3DES_EDE_CBC_SHA,
		http2cipher_TLS_DH_anon_EXPORT_WITH_RC4_40_MD5,
		http2cipher_TLS_DH_anon_WITH_RC4_128_MD5,
		http2cipher_TLS_DH_anon_EXPORT_WITH_DES40_CBC_SHA,
		http2cipher_TLS_DH_anon_WITH_DES_CBC_SHA,
		http2cipher_TLS_DH_anon_WITH_3DES_EDE_CBC_SHA,
		http2cipher_TLS_KRB5_WITH_DES_CBC_SHA,
		http2cipher_TLS_KRB5_WITH_3DES_EDE_CBC_SHA,
		http2cipher_TLS_KRB5_WITH_RC4_128_SHA,
		http2cipher_TLS_KRB5_WITH_IDEA_CBC_SHA,
		http2cipher_TLS_KRB5_WITH_DES_CBC_MD5,
		http2cipher_TLS_KRB5_WITH_3DES_EDE_CBC_MD5,
		http2cipher_TLS_KRB5_WITH_RC4_128_MD5,
		http2cipher_TLS_KRB5_WITH_IDEA_CBC_MD5,
		http2cipher_TLS_KRB5_EXPORT_WITH_DES_CBC_40_SHA,
		http2cipher_TLS_KRB5_EXPORT_WITH_RC2_CBC_40_SHA,
		http2cipher_TLS_KRB5_EXPORT_WITH_RC4_40_SHA,
		http2cipher_TLS_KRB5_EXPORT_WITH_DES_CBC_40_MD5,
		http2cipher_TLS_KRB5_EXPORT_WITH_RC2_CBC_40_MD5,
		http2cipher_TLS_KRB5_EXPORT_WITH_RC4_40_MD5,
		http2cipher_TLS_PSK_WITH_NULL_SHA,
		http2cipher_TLS_DHE_PSK_WITH_NULL_SHA,
		http2cipher_TLS_RSA_PSK_WITH_NULL_SHA,
		http2cipher_TLS_RSA_WITH_AES_128_CBC_SHA,
		http2cipher_TLS_DH_DSS_WITH_AES_128_CBC_SHA,
		http2cipher_TLS_DH_RSA_WITH_AES_128_CBC_SHA,
		http2cipher_TLS_DHE_DSS_WITH_AES_128_CBC_SHA,
		http2cipher_TLS_DHE_RSA_WITH_AES_128_CBC_SHA,
		http2cipher_TLS_DH_anon_WITH_AES_128_CBC_SHA,
		http2cipher_TLS_RSA_WITH_AES_256_CBC_SHA,
		http2cipher_TLS_DH_DSS_WITH_AES_256_CBC_SHA,
		http2cipher_TLS_DH_RSA_WITH_AES_256_CBC_SHA,
		http2cipher_TLS_DHE_DSS_WITH_AES_256_CBC_SHA,
		http2cipher_TLS_DHE_RSA_WITH_AES_256_CBC_SHA,
		http2cipher_TLS_DH_anon_WITH_AES_256_CBC_SHA,
		http2cipher_TLS_RSA_WITH_NULL_SHA256,
		http2cipher_TLS_RSA_WITH_AES_128_CBC_SHA256,
		http2cipher_TLS_RSA_WITH_AES_256_CBC_SHA256,
		http2cipher_TLS_DH_DSS_WITH_AES_128_CBC_SHA256,
		http2cipher_TLS_DH_RSA_WITH_AES_128_CBC_SHA256,
		http2cipher_TLS_DHE_DSS_WITH_AES_128_CBC_SHA256,
		http2cipher_TLS_RSA_WITH_CAMELLIA_128_CBC_SHA,
		http2cipher_TLS_DH_DSS_WITH_CAMELLIA_128_CBC_SHA,
		http2cipher_TLS_DH_RSA_WITH_CAMELLIA_128_CBC_SHA,
		http2cipher_TLS_DHE_DSS_WITH_CAMELLIA_128_CBC_SHA,
		http2cipher_TLS_DHE_RSA_WITH_CAMELLIA_128_CBC_SHA,
		http2cipher_TLS_DH_anon_WITH_CAMELLIA_128_CBC_SHA,
		http2cipher_TLS_DHE_RSA_WITH_AES_128_CBC_SHA256,
		http2cipher_TLS_DH_DSS_WITH_AES_256_CBC_SHA256,
		http2cipher_TLS_DH_RSA_WITH_AES_256_CBC_SHA256,
		http2cipher_TLS_DHE_DSS_WITH_AES_256_CBC_SHA256,
		http2cipher_TLS_DHE_RSA_WITH_AES_256_CBC_SHA256,
		http2cipher_TLS_DH_anon_WITH_AES_128_CBC_SHA256,
		http2cipher_TLS_DH_anon_WITH_AES_256_CBC_SHA256,
		http2cipher_TLS_RSA_WITH_CAMELLIA_256_CBC_SHA,
		http2cipher_TLS_DH_DSS_WITH_CAMELLIA_256_CBC_SHA,
		http2cipher_TLS_DH_RSA_WITH_CAMELLIA_256_CBC_SHA,
		http2cipher_TLS_DHE_DSS_WITH_CAMELLIA_256_CBC_SHA,
		http2cipher_TLS_DHE_RSA_WITH_CAMELLIA_256_CBC_SHA,
		http2cipher_TLS_DH_anon_WITH_CAMELLIA_256_CBC_SHA,
		http2cipher_TLS_PSK_WITH_RC4_128_SHA,
		http2cipher_TLS_PSK_WITH_3DES_EDE_CBC_SHA,
		http2cipher_TLS_PSK_WITH_AES_128_CBC_SHA,
		http2cipher_TLS_PSK_WITH_AES_256_CBC_SHA,
		http2cipher_TLS_DHE_PSK_WITH_RC4_128_SHA,
		http2cipher_TLS_DHE_PSK_WITH_3DES_EDE_CBC_SHA,
		http2cipher_TLS_DHE_PSK_WITH_AES_128_CBC_SHA,
		http2cipher_TLS_DHE_PSK_WITH_AES_256_CBC_SHA,
		http2cipher_TLS_RSA_PSK_WITH_RC4_128_SHA,
		http2cipher_TLS_RSA_PSK_WITH_3DES_EDE_CBC_SHA,
		http2cipher_TLS_RSA_PSK_WITH_AES_128_CBC_SHA,
		http2cipher_TLS_RSA_PSK_WITH_AES_256_CBC_SHA,
		http2cipher_TLS_RSA_WITH_SEED_CBC_SHA,
		http2cipher_TLS_DH_DSS_WITH_SEED_CBC_SHA,
		http2cipher_TLS_DH_RSA_WITH_SEED_CBC_SHA,
		http2cipher_TLS_DHE_DSS_WITH_SEED_CBC_SHA,
		http2cipher_TLS_DHE_RSA_WITH_SEED_CBC_SHA,
		http2cipher_TLS_DH_anon_WITH_SEED_CBC_SHA,
		http2cipher_TLS_RSA_WITH_AES_128_GCM_SHA256,
		http2cipher_TLS_RSA_WITH_AES_256_GCM_SHA384,
		http2cipher_TLS_DH_RSA_WITH_AES_128_GCM_SHA256,
		http2cipher_TLS_DH_RSA_WITH_AES_256_GCM_SHA384,
		http2cipher_TLS_DH_DSS_WITH_AES_128_GCM_SHA256,
		http2cipher_TLS_DH_DSS_WITH_AES_256_GCM_SHA384,
		http2cipher_TLS_DH_anon_WITH_AES_128_GCM_SHA256,
		http2cipher_TLS_DH_anon_WITH_AES_256_GCM_SHA384,
		http2cipher_TLS_PSK_WITH_AES_128_GCM_SHA256,
		http2cipher_TLS_PSK_WITH_AES_256_GCM_SHA384,
		http2cipher_TLS_RSA_PSK_WITH_AES_128_GCM_SHA256,
		http2cipher_TLS_RSA_PSK_WITH_AES_256_GCM_SHA384,
		http2cipher_TLS_PSK_WITH_AES_128_CBC_SHA256,
		http2cipher_TLS_PSK_WITH_AES_256_CBC_SHA384,
		http2cipher_TLS_PSK_WITH_NULL_SHA256,
		http2cipher_TLS_PSK_WITH_NULL_SHA384,
		http2cipher_TLS_DHE_PSK_WITH_AES_128_CBC_SHA256,
		http2cipher_TLS_DHE_PSK_WITH_AES_256_CBC_SHA384,
		http2cipher_TLS_DHE_PSK_WITH_NULL_SHA256,
		http2cipher_TLS_DHE_PSK_WITH_NULL_SHA384,
		http2cipher_TLS_RSA_PSK_WITH_AES_128_CBC_SHA256,
		http2cipher_TLS_RSA_PSK_WITH_AES_256_CBC_SHA384,
		http2cipher_TLS_RSA_PSK_WITH_NULL_SHA256,
		http2cipher_TLS_RSA_PSK_WITH_NULL_SHA384,
		http2cipher_TLS_RSA_WITH_CAMELLIA_128_CBC_SHA256,
		http2cipher_TLS_DH_DSS_WITH_CAMELLIA_128_CBC_SHA256,
		http2cipher_TLS_DH_RSA_WITH_CAMELLIA_128_CBC_SHA256,
		http2cipher_TLS_DHE_DSS_WITH_CAMELLIA_128_CBC_SHA256,
		http2cipher_TLS_DHE_RSA_WITH_CAMELLIA_128_CBC_SHA256,
		http2cipher_TLS_DH_anon_WITH_CAMELLIA_128_CBC_SHA256,
		http2cipher_TLS_RSA_WITH_CAMELLIA_256_CBC_SHA256,
		http2cipher_TLS_DH_DSS_WITH_CAMELLIA_256_CBC_SHA256,
		http2cipher_TLS_DH_RSA_WITH_CAMELLIA_256_CBC_SHA256,
		http2cipher_TLS_DHE_DSS_WITH_CAMELLIA_256_CBC_SHA256,
		http2cipher_TLS_DHE_RSA_WITH_CAMELLIA_256_CBC_SHA256,
		http2cipher_TLS_DH_anon_WITH_CAMELLIA_256_CBC_SHA256,
		http2cipher_TLS_EMPTY_RENEGOTIATION_INFO_SCSV,
		http2cipher_TLS_ECDH_ECDSA_WITH_NULL_SHA,
		http2cipher_TLS_ECDH_ECDSA_WITH_RC4_128_SHA,
		http2cipher_TLS_ECDH_ECDSA_WITH_3DES_EDE_CBC_SHA,
		http2cipher_TLS_ECDH_ECDSA_WITH_AES_128_CBC_SHA,
		http2cipher_TLS_ECDH_ECDSA_WITH_AES_256_CBC_SHA,
		http2cipher_TLS_ECDHE_ECDSA_WITH_NULL_SHA,
		http2cipher_TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
		http2cipher_TLS_ECDHE_ECDSA_WITH_3DES_EDE_CBC_SHA,
		http2cipher_TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		http2cipher_TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		http2cipher_TLS_ECDH_RSA_WITH_NULL_SHA,
		http2cipher_TLS_ECDH_RSA_WITH_RC4_128_SHA,
		http2cipher_TLS_ECDH_RSA_WITH_3DES_EDE_CBC_SHA,
		http2cipher_TLS_ECDH_RSA_WITH_AES_128_CBC_SHA,
		http2cipher_TLS_ECDH_RSA_WITH_AES_256_CBC_SHA,
		http2cipher_TLS_ECDHE_RSA_WITH_NULL_SHA,
		http2cipher_TLS_ECDHE_RSA_WITH_RC4_128_SHA,
		http2cipher_TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
		http2cipher_TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		http2cipher_TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		http2cipher_TLS_ECDH_anon_WITH_NULL_SHA,
		http2cipher_TLS_ECDH_anon_WITH_RC4_128_SHA,
		http2cipher_TLS_ECDH_anon_WITH_3DES_EDE_CBC_SHA,
		http2cipher_TLS_ECDH_anon_WITH_AES_128_CBC_SHA,
		http2cipher_TLS_ECDH_anon_WITH_AES_256_CBC_SHA,
		http2cipher_TLS_SRP_SHA_WITH_3DES_EDE_CBC_SHA,
		http2cipher_TLS_SRP_SHA_RSA_WITH_3DES_EDE_CBC_SHA,
		http2cipher_TLS_SRP_SHA_DSS_WITH_3DES_EDE_CBC_SHA,
		http2cipher_TLS_SRP_SHA_WITH_AES_128_CBC_SHA,
		http2cipher_TLS_SRP_SHA_RSA_WITH_AES_128_CBC_SHA,
		http2cipher_TLS_SRP_SHA_DSS_WITH_AES_128_CBC_SHA,
		http2cipher_TLS_SRP_SHA_WITH_AES_256_CBC_SHA,
		http2cipher_TLS_SRP_SHA_RSA_WITH_AES_256_CBC_SHA,
		http2cipher_TLS_SRP_SHA_DSS_WITH_AES_256_CBC_SHA,
		http2cipher_TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
		http2cipher_TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA384,
		http2cipher_TLS_ECDH_ECDSA_WITH_AES_128_CBC_SHA256,
		http2cipher_TLS_ECDH_ECDSA_WITH_AES_256_CBC_SHA384,
		http2cipher_TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
		http2cipher_TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA384,
		http2cipher_TLS_ECDH_RSA_WITH_AES_128_CBC_SHA256,
		http2cipher_TLS_ECDH_RSA_WITH_AES_256_CBC_SHA384,
		http2cipher_TLS_ECDH_ECDSA_WITH_AES_128_GCM_SHA256,
		http2cipher_TLS_ECDH_ECDSA_WITH_AES_256_GCM_SHA384,
		http2cipher_TLS_ECDH_RSA_WITH_AES_128_GCM_SHA256,
		http2cipher_TLS_ECDH_RSA_WITH_AES_256_GCM_SHA384,
		http2cipher_TLS_ECDHE_PSK_WITH_RC4_128_SHA,
		http2cipher_TLS_ECDHE_PSK_WITH_3DES_EDE_CBC_SHA,
		http2cipher_TLS_ECDHE_PSK_WITH_AES_128_CBC_SHA,
		http2cipher_TLS_ECDHE_PSK_WITH_AES_256_CBC_SHA,
		http2cipher_TLS_ECDHE_PSK_WITH_AES_128_CBC_SHA256,
		http2cipher_TLS_ECDHE_PSK_WITH_AES_256_CBC_SHA384,
		http2cipher_TLS_ECDHE_PSK_WITH_NULL_SHA,
		http2cipher_TLS_ECDHE_PSK_WITH_NULL_SHA256,
		http2cipher_TLS_ECDHE_PSK_WITH_NULL_SHA384,
		http2cipher_TLS_RSA_WITH_ARIA_128_CBC_SHA256,
		http2cipher_TLS_RSA_WITH_ARIA_256_CBC_SHA384,
		http2cipher_TLS_DH_DSS_WITH_ARIA_128_CBC_SHA256,
		http2cipher_TLS_DH_DSS_WITH_ARIA_256_CBC_SHA384,
		http2cipher_TLS_DH_RSA_WITH_ARIA_128_CBC_SHA256,
		http2cipher_TLS_DH_RSA_WITH_ARIA_256_CBC_SHA384,
		http2cipher_TLS_DHE_DSS_WITH_ARIA_128_CBC_SHA256,
		http2cipher_TLS_DHE_DSS_WITH_ARIA_256_CBC_SHA384,
		http2cipher_TLS_DHE_RSA_WITH_ARIA_128_CBC_SHA256,
		http2cipher_TLS_DHE_RSA_WITH_ARIA_256_CBC_SHA384,
		http2cipher_TLS_DH_anon_WITH_ARIA_128_CBC_SHA256,
		http2cipher_TLS_DH_anon_WITH_ARIA_256_CBC_SHA384,
		http2cipher_TLS_ECDHE_ECDSA_WITH_ARIA_128_CBC_SHA256,
		http2cipher_TLS_ECDHE_ECDSA_WITH_ARIA_256_CBC_SHA384,
		http2cipher_TLS_ECDH_ECDSA_WITH_ARIA_128_CBC_SHA256,
		http2cipher_TLS_ECDH_ECDSA_WITH_ARIA_256_CBC_SHA384,
		http2cipher_TLS_ECDHE_RSA_WITH_ARIA_128_CBC_SHA256,
		http2cipher_TLS_ECDHE_RSA_WITH_ARIA_256_CBC_SHA384,
		http2cipher_TLS_ECDH_RSA_WITH_ARIA_128_CBC_SHA256,
		http2cipher_TLS_ECDH_RSA_WITH_ARIA_256_CBC_SHA384,
		http2cipher_TLS_RSA_WITH_ARIA_128_GCM_SHA256,
		http2cipher_TLS_RSA_WITH_ARIA_256_GCM_SHA384,
		http2cipher_TLS_DH_RSA_WITH_ARIA_128_GCM_SHA256,
		http2cipher_TLS_DH_RSA_WITH_ARIA_256_GCM_SHA384,
		http2cipher_TLS_DH_DSS_WITH_ARIA_128_GCM_SHA256,
		http2cipher_TLS_DH_DSS_WITH_ARIA_256_GCM_SHA384,
		http2cipher_TLS_DH_anon_WITH_ARIA_128_GCM_SHA256,
		http2cipher_TLS_DH_anon_WITH_ARIA_256_GCM_SHA384,
		http2cipher_TLS_ECDH_ECDSA_WITH_ARIA_128_GCM_SHA256,
		http2cipher_TLS_ECDH_ECDSA_WITH_ARIA_256_GCM_SHA384,
		http2cipher_TLS_ECDH_RSA_WITH_ARIA_128_GCM_SHA256,
		http2cipher_TLS_ECDH_RSA_WITH_ARIA_256_GCM_SHA384,
		http2cipher_TLS_PSK_WITH_ARIA_128_CBC_SHA256,
		http2cipher_TLS_PSK_WITH_ARIA_256_CBC_SHA384,
		http2cipher_TLS_DHE_PSK_WITH_ARIA_128_CBC_SHA256,
		http2cipher_TLS_DHE_PSK_WITH_ARIA_256_CBC_SHA384,
		http2cipher_TLS_RSA_PSK_WITH_ARIA_128_CBC_SHA256,
		http2cipher_TLS_RSA_PSK_WITH_ARIA_256_CBC_SHA384,
		http2cipher_TLS_PSK_WITH_ARIA_128_GCM_SHA256,
		http2cipher_TLS_PSK_WITH_ARIA_256_GCM_SHA384,
		http2cipher_TLS_RSA_PSK_WITH_ARIA_128_GCM_SHA256,
		http2cipher_TLS_RSA_PSK_WITH_ARIA_256_GCM_SHA384,
		http2cipher_TLS_ECDHE_PSK_WITH_ARIA_128_CBC_SHA256,
		http2cipher_TLS_ECDHE_PSK_WITH_ARIA_256_CBC_SHA384,
		http2cipher_TLS_ECDHE_ECDSA_WITH_CAMELLIA_128_CBC_SHA256,
		http2cipher_TLS_ECDHE_ECDSA_WITH_CAMELLIA_256_CBC_SHA384,
		http2cipher_TLS_ECDH_ECDSA_WITH_CAMELLIA_128_CBC_SHA256,
		http2cipher_TLS_ECDH_ECDSA_WITH_CAMELLIA_256_CBC_SHA384,
		http2cipher_TLS_ECDHE_RSA_WITH_CAMELLIA_128_CBC_SHA256,
		http2cipher_TLS_ECDHE_RSA_WITH_CAMELLIA_256_CBC_SHA384,
		http2cipher_TLS_ECDH_RSA_WITH_CAMELLIA_128_CBC_SHA256,
		http2cipher_TLS_ECDH_RSA_WITH_CAMELLIA_256_CBC_SHA384,
		http2cipher_TLS_RSA_WITH_CAMELLIA_128_GCM_SHA256,
		http2cipher_TLS_RSA_WITH_CAMELLIA_256_GCM_SHA384,
		http2cipher_TLS_DH_RSA_WITH_CAMELLIA_128_GCM_SHA256,
		http2cipher_TLS_DH_RSA_WITH_CAMELLIA_256_GCM_SHA384,
		http2cipher_TLS_DH_DSS_WITH_CAMELLIA_128_GCM_SHA256,
		http2cipher_TLS_DH_DSS_WITH_CAMELLIA_256_GCM_SHA384,
		http2cipher_TLS_DH_anon_WITH_CAMELLIA_128_GCM_SHA256,
		http2cipher_TLS_DH_anon_WITH_CAMELLIA_256_GCM_SHA384,
		http2cipher_TLS_ECDH_ECDSA_WITH_CAMELLIA_128_GCM_SHA256,
		http2cipher_TLS_ECDH_ECDSA_WITH_CAMELLIA_256_GCM_SHA384,
		http2cipher_TLS_ECDH_RSA_WITH_CAMELLIA_128_GCM_SHA256,
		http2cipher_TLS_ECDH_RSA_WITH_CAMELLIA_256_GCM_SHA384,
		http2cipher_TLS_PSK_WITH_CAMELLIA_128_GCM_SHA256,
		http2cipher_TLS_PSK_WITH_CAMELLIA_256_GCM_SHA384,
		http2cipher_TLS_RSA_PSK_WITH_CAMELLIA_128_GCM_SHA256,
		http2cipher_TLS_RSA_PSK_WITH_CAMELLIA_256_GCM_SHA384,
		http2cipher_TLS_PSK_WITH_CAMELLIA_128_CBC_SHA256,
		http2cipher_TLS_PSK_WITH_CAMELLIA_256_CBC_SHA384,
		http2cipher_TLS_DHE_PSK_WITH_CAMELLIA_128_CBC_SHA256,
		http2cipher_TLS_DHE_PSK_WITH_CAMELLIA_256_CBC_SHA384,
		http2cipher_TLS_RSA_PSK_WITH_CAMELLIA_128_CBC_SHA256,
		http2cipher_TLS_RSA_PSK_WITH_CAMELLIA_256_CBC_SHA384,
		http2cipher_TLS_ECDHE_PSK_WITH_CAMELLIA_128_CBC_SHA256,
		http2cipher_TLS_ECDHE_PSK_WITH_CAMELLIA_256_CBC_SHA384,
		http2cipher_TLS_RSA_WITH_AES_128_CCM,
		http2cipher_TLS_RSA_WITH_AES_256_CCM,
		http2cipher_TLS_RSA_WITH_AES_128_CCM_8,
		http2cipher_TLS_RSA_WITH_AES_256_CCM_8,
		http2cipher_TLS_PSK_WITH_AES_128_CCM,
		http2cipher_TLS_PSK_WITH_AES_256_CCM,
		http2cipher_TLS_PSK_WITH_AES_128_CCM_8,
		http2cipher_TLS_PSK_WITH_AES_256_CCM_8:
		return true
	default:
		return false
	}
}

const (
	http2prefaceTimeout         = 10 * time.Second
	http2firstSettingsTimeout   = 2 * time.Second // should be in-flight with preface anyway
	http2handlerChunkWriteSize  = 4 << 10
	http2defaultMaxStreams      = 250 // TODO: make this 100 as the GFE seems to?
	http2maxQueuedControlFrames = 10000
)

var (
	http2errClientDisconnected = errors.New("client disconnected")
	http2errClosedBody         = errors.New("body closed by handler")
	http2errHandlerComplete    = errors.New("http2: request body closed due to handler exiting")
	http2errStreamClosed       = errors.New("http2: stream closed")
)

var http2responseWriterStatePool = sync.Pool{
	New: func() interface{} {
		rws := &http2responseWriterState{}
		rws.bw = bufio.NewWriterSize(http2chunkWriter{rws}, http2handlerChunkWriteSize)
		return rws
	},
}

// Test hooks.
var (
	http2testHookOnConn        func()
	http2testHookGetServerConn func(*http2serverConn)
	http2testHookOnPanicMu     *sync.Mutex // nil except in tests
	http2testHookOnPanic       func(sc *http2serverConn, panicVal interface{}) (rePanic bool)
)

// Server is an HTTP/2 server.
type http2Server struct {
	// MaxHandlers limits the number of http.Handler ServeHTTP goroutines
	// which may run at a time over all connections.
	// Negative or zero no limit.
	// TODO: implement
	MaxHandlers int

	// MaxConcurrentStreams optionally specifies the number of
	// concurrent streams that each client may have open at a
	// time. This is unrelated to the number of http.Handler goroutines
	// which may be active globally, which is MaxHandlers.
	// If zero, MaxConcurrentStreams defaults to at least 100, per
	// the HTTP/2 spec's recommendations.
	MaxConcurrentStreams uint32

	// MaxReadFrameSize optionally specifies the largest frame
	// this server is willing to read. A valid value is between
	// 16k and 16M, inclusive. If zero or otherwise invalid, a
	// default value is used.
	MaxReadFrameSize uint32

	// PermitProhibitedCipherSuites, if true, permits the use of
	// cipher suites prohibited by the HTTP/2 spec.
	PermitProhibitedCipherSuites bool

	// IdleTimeout specifies how long until idle clients should be
	// closed with a GOAWAY frame. PING frames are not considered
	// activity for the purposes of IdleTimeout.
	IdleTimeout time.Duration

	// MaxUploadBufferPerConnection is the size of the initial flow
	// control window for each connections. The HTTP/2 spec does not
	// allow this to be smaller than 65535 or larger than 2^32-1.
	// If the value is outside this range, a default value will be
	// used instead.
	MaxUploadBufferPerConnection int32

	// MaxUploadBufferPerStream is the size of the initial flow control
	// window for each stream. The HTTP/2 spec does not allow this to
	// be larger than 2^32-1. If the value is zero or larger than the
	// maximum, a default value will be used instead.
	MaxUploadBufferPerStream int32

	// NewWriteScheduler constructs a write scheduler for a connection.
	// If nil, a default scheduler is chosen.
	NewWriteScheduler func() http2WriteScheduler

	// CountError, if non-nil, is called on HTTP/2 server errors.
	// It's intended to increment a metric for monitoring, such
	// as an expvar or Prometheus metric.
	// The errType consists of only ASCII word characters.
	CountError func(errType string)

	// Internal state. This is a pointer (rather than embedded directly)
	// so that we don't embed a Mutex in this struct, which will make the
	// struct non-copyable, which might break some callers.
	state *http2serverInternalState
}

func (s *http2Server) initialConnRecvWindowSize() int32 {
	if s.MaxUploadBufferPerConnection > http2initialWindowSize {
		return s.MaxUploadBufferPerConnection
	}
	return 1 << 20
}

func (s *http2Server) initialStreamRecvWindowSize() int32 {
	if s.MaxUploadBufferPerStream > 0 {
		return s.MaxUploadBufferPerStream
	}
	return 1 << 20
}

func (s *http2Server) maxReadFrameSize() uint32 {
	if v := s.MaxReadFrameSize; v >= http2minMaxFrameSize && v <= http2maxFrameSize {
		return v
	}
	return http2defaultMaxReadFrameSize
}

func (s *http2Server) maxConcurrentStreams() uint32 {
	if v := s.MaxConcurrentStreams; v > 0 {
		return v
	}
	return http2defaultMaxStreams
}

// maxQueuedControlFrames is the maximum number of control frames like
// SETTINGS, PING and RST_STREAM that will be queued for writing before
// the connection is closed to prevent memory exhaustion attacks.
func (s *http2Server) maxQueuedControlFrames() int {
	// TODO: if anybody asks, add a Server field, and remember to define the
	// behavior of negative values.
	return http2maxQueuedControlFrames
}

// ServeConn serves HTTP/2 requests on the provided connection and
// blocks until the connection is no longer readable.
//
// ServeConn starts speaking HTTP/2 assuming that c has not had any
// reads or writes. It writes its initial settings frame and expects
// to be able to read the preface and settings frame from the
// client. If c has a ConnectionState method like a *tls.Conn, the
// ConnectionState is used to verify the TLS ciphersuite and to set
// the Request.TLS field in Handlers.
//
// ServeConn does not support h2c by itself. Any h2c support must be
// implemented in terms of providing a suitably-behaving net.Conn.
//
// The opts parameter is optional. If nil, default values are used.
func (s *http2Server) ServeConn(c net.Conn, opts *http2ServeConnOpts) {
	baseCtx, cancel := http2serverConnBaseContext(c, opts)
	defer cancel()

	sc := &http2serverConn{
		srv:                         s,
		hs:                          opts.baseConfig(),
		conn:                        c,
		baseCtx:                     baseCtx,
		remoteAddrStr:               c.RemoteAddr().String(),
		bw:                          http2newBufferedWriter(c),
		handler:                     opts.handler(),
		streams:                     make(map[uint32]*http2stream),
		readFrameCh:                 make(chan http2readFrameResult),
		wantWriteFrameCh:            make(chan http2FrameWriteRequest, 8),
		serveMsgCh:                  make(chan interface{}, 8),
		wroteFrameCh:                make(chan http2frameWriteResult, 1), // buffered; one send in writeFrameAsync
		bodyReadCh:                  make(chan http2bodyReadMsg),         // buffering doesn't matter either way
		doneServing:                 make(chan struct{}),
		clientMaxStreams:            math.MaxUint32, // Section 6.5.2: "Initially, there is no limit to this value"
		advMaxStreams:               s.maxConcurrentStreams(),
		initialStreamSendWindowSize: http2initialWindowSize,
		maxFrameSize:                http2initialMaxFrameSize,
		headerTableSize:             http2initialHeaderTableSize,
		serveG:                      http2newGoroutineLock(),
		pushEnabled:                 true,
	}

	s.state.registerConn(sc)
	defer s.state.unregisterConn(sc)

	// The net/http package sets the write deadline from the
	// http.Server.WriteTimeout during the TLS handshake, but then
	// passes the connection off to us with the deadline already set.
	// Write deadlines are set per stream in serverConn.newStream.
	// Disarm the net.Conn write deadline here.
	if sc.hs.WriteTimeout != 0 {
		sc.conn.SetWriteDeadline(time.Time{})
	}

	if s.NewWriteScheduler != nil {
		sc.writeSched = s.NewWriteScheduler()
	} else {
		sc.writeSched = http2NewRandomWriteScheduler()
	}

	// These start at the RFC-specified defaults. If there is a higher
	// configured value for inflow, that will be updated when we send a
	// WINDOW_UPDATE shortly after sending SETTINGS.
	sc.flow.add(http2initialWindowSize)
	sc.inflow.add(http2initialWindowSize)
	sc.hpackEncoder = hpack.NewEncoder(&sc.headerWriteBuf)

	fr := http2NewFramer(sc.bw, c)
	fr.ReadMetaHeaders = hpack.NewDecoder(http2initialHeaderTableSize, nil)
	fr.MaxHeaderListSize = sc.maxHeaderListSize()
	fr.SetMaxReadFrameSize(s.maxReadFrameSize())
	sc.framer = fr

	if tc, ok := c.(http2connectionStater); ok {
		sc.tlsState = new(tls.ConnectionState)
		*sc.tlsState = tc.ConnectionState()
		// 9.2 Use of TLS Features
		// An implementation of HTTP/2 over TLS MUST use TLS
		// 1.2 or higher with the restrictions on feature set
		// and cipher suite described in this section. Due to
		// implementation limitations, it might not be
		// possible to fail TLS negotiation. An endpoint MUST
		// immediately terminate an HTTP/2 connection that
		// does not meet the TLS requirements described in
		// this section with a connection error (Section
		// 5.4.1) of type INADEQUATE_SECURITY.
		if sc.tlsState.Version < tls.VersionTLS12 {
			sc.rejectConn(http2ErrCodeInadequateSecurity, "TLS version too low")
			return
		}

		if sc.tlsState.ServerName == "" {
			// Client must use SNI, but we don't enforce that anymore,
			// since it was causing problems when connecting to bare IP
			// addresses during development.
			//
			// TODO: optionally enforce? Or enforce at the time we receive
			// a new request, and verify the ServerName matches the :authority?
			// But that precludes proxy situations, perhaps.
			//
			// So for now, do nothing here again.
		}

		if !s.PermitProhibitedCipherSuites && http2isBadCipher(sc.tlsState.CipherSuite) {
			// "Endpoints MAY choose to generate a connection error
			// (Section 5.4.1) of type INADEQUATE_SECURITY if one of
			// the prohibited cipher suites are negotiated."
			//
			// We choose that. In my opinion, the spec is weak
			// here. It also says both parties must support at least
			// TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256 so there's no
			// excuses here. If we really must, we could allow an
			// "AllowInsecureWeakCiphers" option on the server later.
			// Let's see how it plays out first.
			sc.rejectConn(http2ErrCodeInadequateSecurity, fmt.Sprintf("Prohibited TLS 1.2 Cipher Suite: %x", sc.tlsState.CipherSuite))
			return
		}
	}

	if hook := http2testHookGetServerConn; hook != nil {
		hook(sc)
	}
	sc.serve()
}

type http2serverInternalState struct {
	mu          sync.Mutex
	activeConns map[*http2serverConn]struct{}
}

func (s *http2serverInternalState) registerConn(sc *http2serverConn) {
	if s == nil {
		return // if the Server was used without calling ConfigureServer
	}
	s.mu.Lock()
	s.activeConns[sc] = struct{}{}
	s.mu.Unlock()
}

func (s *http2serverInternalState) unregisterConn(sc *http2serverConn) {
	if s == nil {
		return // if the Server was used without calling ConfigureServer
	}
	s.mu.Lock()
	delete(s.activeConns, sc)
	s.mu.Unlock()
}

func (s *http2serverInternalState) startGracefulShutdown() {
	if s == nil {
		return // if the Server was used without calling ConfigureServer
	}
	s.mu.Lock()
	for sc := range s.activeConns {
		sc.startGracefulShutdown()
	}
	s.mu.Unlock()
}

// ServeConnOpts are options for the Server.ServeConn method.
type http2ServeConnOpts struct {
	// Context is the base context to use.
	// If nil, context.Background is used.
	Context context.Context

	// BaseConfig optionally sets the base configuration
	// for values. If nil, defaults are used.
	BaseConfig *http.Server

	// Handler specifies which handler to use for processing
	// requests. If nil, BaseConfig.Handler is used. If BaseConfig
	// or BaseConfig.Handler is nil, http.DefaultServeMux is used.
	Handler http.Handler
}

func (o *http2ServeConnOpts) context() context.Context {
	if o != nil && o.Context != nil {
		return o.Context
	}
	return context.Background()
}

func (o *http2ServeConnOpts) baseConfig() *http.Server {
	if o != nil && o.BaseConfig != nil {
		return o.BaseConfig
	}
	return new(http.Server)
}

func (o *http2ServeConnOpts) handler() http.Handler {
	if o != nil {
		if o.Handler != nil {
			return o.Handler
		}
		if o.BaseConfig != nil && o.BaseConfig.Handler != nil {
			return o.BaseConfig.Handler
		}
	}
	return http.DefaultServeMux
}

func http2serverConnBaseContext(c net.Conn, opts *http2ServeConnOpts) (ctx context.Context, cancel func()) {
	ctx, cancel = context.WithCancel(opts.context())
	ctx = context.WithValue(ctx, http.LocalAddrContextKey, c.LocalAddr())
	if hs := opts.baseConfig(); hs != nil {
		ctx = context.WithValue(ctx, http.ServerContextKey, hs)
	}
	return
}

// bufferedWriter is a buffered writer that writes to w.
// Its buffered writer is lazily allocated as needed, to minimize
// idle memory usage with many connections.
type http2bufferedWriter struct {
	_  http2incomparable
	w  io.Writer     // immutable
	bw *bufio.Writer // non-nil when data is buffered
}

func http2newBufferedWriter(w io.Writer) *http2bufferedWriter {
	return &http2bufferedWriter{w: w}
}

func (w *http2bufferedWriter) Available() int {
	if w.bw == nil {
		return http2bufWriterPoolBufferSize
	}
	return w.bw.Available()
}

func (w *http2bufferedWriter) Write(p []byte) (n int, err error) {
	if w.bw == nil {
		bw := http2bufWriterPool.Get().(*bufio.Writer)
		bw.Reset(w.w)
		w.bw = bw
	}
	return w.bw.Write(p)
}

func (w *http2bufferedWriter) Flush() error {
	bw := w.bw
	if bw == nil {
		return nil
	}
	err := bw.Flush()
	bw.Reset(nil)
	http2bufWriterPool.Put(bw)
	w.bw = nil
	return err
}

func (sc *http2serverConn) rejectConn(err http2ErrCode, debug string) {
	sc.vlogf("http2: server rejecting conn: %v, %s", err, debug)
	// ignoring errors. hanging up anyway.
	sc.framer.WriteGoAway(0, err, []byte(debug))
	sc.bw.Flush()
	sc.conn.Close()
}

type http2serverConn struct {
	// Immutable:
	srv              *http2Server
	hs               *http.Server
	conn             net.Conn
	bw               *http2bufferedWriter // writing to conn
	handler          http.Handler
	baseCtx          context.Context
	framer           *http2Framer
	doneServing      chan struct{}               // closed when serverConn.serve ends
	readFrameCh      chan http2readFrameResult   // written by serverConn.readFrames
	wantWriteFrameCh chan http2FrameWriteRequest // from handlers -> serve
	wroteFrameCh     chan http2frameWriteResult  // from writeFrameAsync -> serve, tickles more frame writes
	bodyReadCh       chan http2bodyReadMsg       // from handlers -> serve
	serveMsgCh       chan interface{}            // misc messages & code to send to / run on the serve loop
	flow             http2flow                   // conn-wide (not stream-specific) outbound flow control
	inflow           http2flow                   // conn-wide inbound flow control
	tlsState         *tls.ConnectionState        // shared by all handlers, like net/http
	remoteAddrStr    string
	writeSched       http2WriteScheduler

	// Everything following is owned by the serve loop; use serveG.check():
	serveG                      http2goroutineLock // used to verify funcs are on serve()
	pushEnabled                 bool
	sawFirstSettings            bool // got the initial SETTINGS frame after the preface
	needToSendSettingsAck       bool
	unackedSettings             int    // how many SETTINGS have we sent without ACKs?
	queuedControlFrames         int    // control frames in the writeSched queue
	clientMaxStreams            uint32 // SETTINGS_MAX_CONCURRENT_STREAMS from client (our PUSH_PROMISE limit)
	advMaxStreams               uint32 // our SETTINGS_MAX_CONCURRENT_STREAMS advertised the client
	curClientStreams            uint32 // number of open streams initiated by the client
	curPushedStreams            uint32 // number of open streams initiated by server push
	maxClientStreamID           uint32 // max ever seen from client (odd), or 0 if there have been no client requests
	maxPushPromiseID            uint32 // ID of the last push promise (even), or 0 if there have been no pushes
	streams                     map[uint32]*http2stream
	initialStreamSendWindowSize int32
	maxFrameSize                int32
	headerTableSize             uint32
	peerMaxHeaderListSize       uint32            // zero means unknown (default)
	canonHeader                 map[string]string // http2-lower-case -> Go-Canonical-Case
	writingFrame                bool              // started writing a frame (on serve goroutine or separate)
	writingFrameAsync           bool              // started a frame on its own goroutine but haven't heard back on wroteFrameCh
	needsFrameFlush             bool              // last frame write wasn't a flush
	inGoAway                    bool              // we've started to or sent GOAWAY
	inFrameScheduleLoop         bool              // whether we're in the scheduleFrameWrite loop
	needToSendGoAway            bool              // we need to schedule a GOAWAY frame write
	goAwayCode                  http2ErrCode
	shutdownTimer               *time.Timer // nil until used
	idleTimer                   *time.Timer // nil if unused

	// Owned by the writeFrameAsync goroutine:
	headerWriteBuf bytes.Buffer
	hpackEncoder   *hpack.Encoder

	// Used by startGracefulShutdown.
	shutdownOnce sync.Once
}

func (sc *http2serverConn) maxHeaderListSize() uint32 {
	n := sc.hs.MaxHeaderBytes
	if n <= 0 {
		n = http.DefaultMaxHeaderBytes
	}
	// http2's count is in a slightly different unit and includes 32 bytes per pair.
	// So, take the net/http.Server value and pad it up a bit, assuming 10 headers.
	const perFieldOverhead = 32 // per http2 spec
	const typicalHeaders = 10   // conservative
	return uint32(n + typicalHeaders*perFieldOverhead)
}

func (sc *http2serverConn) curOpenStreams() uint32 {
	sc.serveG.check()
	return sc.curClientStreams + sc.curPushedStreams
}

// A closeWaiter is like a sync.WaitGroup but only goes 1 to 0 (open to closed).
type http2closeWaiter chan struct{}

// Init makes a closeWaiter usable.
// It exists because so a closeWaiter value can be placed inside a
// larger struct and have the Mutex and Cond's memory in the same
// allocation.
func (cw *http2closeWaiter) Init() {
	*cw = make(chan struct{})
}

// Close marks the closeWaiter as closed and unblocks any waiters.
func (cw http2closeWaiter) Close() {
	close(cw)
}

// Wait waits for the closeWaiter to become closed.
func (cw http2closeWaiter) Wait() {
	<-cw
}

// stream represents a stream. This is the minimal metadata needed by
// the serve goroutine. Most of the actual stream state is owned by
// the http.Handler's goroutine in the responseWriter. Because the
// responseWriter's responseWriterState is recycled at the end of a
// handler, this struct intentionally has no pointer to the
// *responseWriter{,State} itself, as the Handler ending nils out the
// responseWriter's state field.
type http2stream struct {
	// immutable:
	sc        *http2serverConn
	id        uint32
	body      *http2pipe       // non-nil if expecting DATA frames
	cw        http2closeWaiter // closed wait stream transitions to closed state
	ctx       context.Context
	cancelCtx func()

	// owned by serverConn's serve loop:
	bodyBytes        int64     // body bytes seen so far
	declBodyBytes    int64     // or -1 if undeclared
	flow             http2flow // limits writing from Handler to client
	inflow           http2flow // what the client is allowed to POST/etc to us
	state            http2streamState
	resetQueued      bool        // RST_STREAM queued for write; set by sc.resetStream
	gotTrailerHeader bool        // HEADER frame for trailers was seen
	wroteHeaders     bool        // whether we wrote headers (not status 100)
	writeDeadline    *time.Timer // nil if unused

	trailer    http.Header // accumulated trailers
	reqTrailer http.Header // handler's Request.Trailer
}

func (sc *http2serverConn) Framer() *http2Framer { return sc.framer }

func (sc *http2serverConn) CloseConn() error { return sc.conn.Close() }

func (sc *http2serverConn) Flush() error { return sc.bw.Flush() }

func (sc *http2serverConn) HeaderEncoder() (*hpack.Encoder, *bytes.Buffer) {
	return sc.hpackEncoder, &sc.headerWriteBuf
}

const (
	// SETTINGS_MAX_FRAME_SIZE default
	// http://http2.github.io/http2-spec/#rfc.section.6.5.2
	http2initialMaxFrameSize = 16384

	http2defaultMaxReadFrameSize = 1 << 20
)

type http2streamState int

// HTTP/2 stream states.
//
// See http://tools.ietf.org/html/rfc7540#section-5.1.
//
// For simplicity, the server code merges "reserved (local)" into
// "half-closed (remote)". This is one less state transition to track.
// The only downside is that we send PUSH_PROMISEs slightly less
// liberally than allowable. More discussion here:
// https://lists.w3.org/Archives/Public/ietf-http-wg/2016JulSep/0599.html
//
// "reserved (remote)" is omitted since the client code does not
// support server push.
const (
	http2stateIdle http2streamState = iota
	http2stateOpen
	http2stateHalfClosedLocal
	http2stateHalfClosedRemote
	http2stateClosed
)

var http2stateName = [...]string{
	http2stateIdle:             "Idle",
	http2stateOpen:             "Open",
	http2stateHalfClosedLocal:  "HalfClosedLocal",
	http2stateHalfClosedRemote: "HalfClosedRemote",
	http2stateClosed:           "Closed",
}

func (st http2streamState) String() string {
	return http2stateName[st]
}

func (sc *http2serverConn) state(streamID uint32) (http2streamState, *http2stream) {
	sc.serveG.check()
	// http://tools.ietf.org/html/rfc7540#section-5.1
	if st, ok := sc.streams[streamID]; ok {
		return st.state, st
	}
	// "The first use of a new stream identifier implicitly closes all
	// streams in the "idle" state that might have been initiated by
	// that peer with a lower-valued stream identifier. For example, if
	// a client sends a HEADERS frame on stream 7 without ever sending a
	// frame on stream 5, then stream 5 transitions to the "closed"
	// state when the first frame for stream 7 is sent or received."
	if streamID%2 == 1 {
		if streamID <= sc.maxClientStreamID {
			return http2stateClosed, nil
		}
	} else {
		if streamID <= sc.maxPushPromiseID {
			return http2stateClosed, nil
		}
	}
	return http2stateIdle, nil
}

// setConnState calls the net/http ConnState hook for this connection, if configured.
// Note that the net/http package does StateNew and StateClosed for us.
// There is currently no plan for StateHijacked or hijacking HTTP/2 connections.
func (sc *http2serverConn) setConnState(state http.ConnState) {
	if sc.hs.ConnState != nil {
		sc.hs.ConnState(sc.conn, state)
	}
}

func (sc *http2serverConn) vlogf(format string, args ...interface{}) {
	if http2VerboseLogs {
		sc.logf(format, args...)
	}
}

func (sc *http2serverConn) logf(format string, args ...interface{}) {
	if lg := sc.hs.ErrorLog; lg != nil {
		lg.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

// errno returns v's underlying uintptr, else 0.
//
// TODO: remove this helper function once http2 can use build
// tags. See comment in isClosedConnError.
func http2errno(v error) uintptr {
	if rv := reflect.ValueOf(v); rv.Kind() == reflect.Uintptr {
		return uintptr(rv.Uint())
	}
	return 0
}

// isClosedConnError reports whether err is an error from use of a closed
// network connection.
func http2isClosedConnError(err error) bool {
	if err == nil {
		return false
	}

	// TODO: remove this string search and be more like the Windows
	// case below. That might involve modifying the standard library
	// to return better error types.
	str := err.Error()
	if strings.Contains(str, "use of closed network connection") {
		return true
	}

	// TODO(bradfitz): x/tools/cmd/bundle doesn't really support
	// build tags, so I can't make an http2_windows.go file with
	// Windows-specific stuff. Fix that and move this, once we
	// have a way to bundle this into std's net/http somehow.
	if runtime.GOOS == "windows" {
		if oe, ok := err.(*net.OpError); ok && oe.Op == "read" {
			if se, ok := oe.Err.(*os.SyscallError); ok && se.Syscall == "wsarecv" {
				const WSAECONNABORTED = 10053
				const WSAECONNRESET = 10054
				if n := http2errno(se.Err); n == WSAECONNRESET || n == WSAECONNABORTED {
					return true
				}
			}
		}
	}
	return false
}

func (sc *http2serverConn) condlogf(err error, format string, args ...interface{}) {
	if err == nil {
		return
	}
	if err == io.EOF || err == io.ErrUnexpectedEOF || http2isClosedConnError(err) || err == http2errPrefaceTimeout {
		// Boring, expected errors.
		sc.vlogf(format, args...)
	} else {
		sc.logf(format, args...)
	}
}

func (sc *http2serverConn) canonicalHeader(v string) string {
	sc.serveG.check()
	http2buildCommonHeaderMapsOnce()
	cv, ok := http2commonCanonHeader[v]
	if ok {
		return cv
	}
	cv, ok = sc.canonHeader[v]
	if ok {
		return cv
	}
	if sc.canonHeader == nil {
		sc.canonHeader = make(map[string]string)
	}
	cv = http.CanonicalHeaderKey(v)
	// maxCachedCanonicalHeaders is an arbitrarily-chosen limit on the number of
	// entries in the canonHeader cache. This should be larger than the number
	// of unique, uncommon header keys likely to be sent by the peer, while not
	// so high as to permit unreasonable memory usage if the peer sends an unbounded
	// number of unique header keys.
	const maxCachedCanonicalHeaders = 32
	if len(sc.canonHeader) < maxCachedCanonicalHeaders {
		sc.canonHeader[v] = cv
	}
	return cv
}

type http2readFrameResult struct {
	f   http2Frame // valid until readMore is called
	err error

	// readMore should be called once the consumer no longer needs or
	// retains f. After readMore, f is invalid and more frames can be
	// read.
	readMore func()
}

// A gate lets two goroutines coordinate their activities.
type http2gate chan struct{}

func (g http2gate) Done() { g <- struct{}{} }

func (g http2gate) Wait() { <-g }

// readFrames is the loop that reads incoming frames.
// It takes care to only read one frame at a time, blocking until the
// consumer is done with the frame.
// It's run on its own goroutine.
func (sc *http2serverConn) readFrames() {
	gate := make(http2gate)
	gateDone := gate.Done
	for {
		f, err := sc.framer.ReadFrame()
		select {
		case sc.readFrameCh <- http2readFrameResult{f, err, gateDone}:
		case <-sc.doneServing:
			return
		}
		select {
		case <-gate:
		case <-sc.doneServing:
			return
		}
		if http2terminalReadFrameError(err) {
			return
		}
	}
}

// frameWriteResult is the message passed from writeFrameAsync to the serve goroutine.
type http2frameWriteResult struct {
	_   http2incomparable
	wr  http2FrameWriteRequest // what was written (or attempted)
	err error                  // result of the writeFrame call
}

// writeFrameAsync runs in its own goroutine and writes a single frame
// and then reports when it's done.
// At most one goroutine can be running writeFrameAsync at a time per
// serverConn.
func (sc *http2serverConn) writeFrameAsync(wr http2FrameWriteRequest) {
	err := wr.write.writeFrame(sc)
	sc.wroteFrameCh <- http2frameWriteResult{wr: wr, err: err}
}

func (sc *http2serverConn) closeAllStreamsOnConnClose() {
	sc.serveG.check()
	for _, st := range sc.streams {
		sc.closeStream(st, http2errClientDisconnected)
	}
}

func (sc *http2serverConn) stopShutdownTimer() {
	sc.serveG.check()
	if t := sc.shutdownTimer; t != nil {
		t.Stop()
	}
}

func (sc *http2serverConn) notePanic() {
	// Note: this is for serverConn.serve panicking, not http.Handler code.
	if http2testHookOnPanicMu != nil {
		http2testHookOnPanicMu.Lock()
		defer http2testHookOnPanicMu.Unlock()
	}
	if http2testHookOnPanic != nil {
		if e := recover(); e != nil {
			if http2testHookOnPanic(sc, e) {
				panic(e)
			}
		}
	}
}

func (sc *http2serverConn) serve() {
	sc.serveG.check()
	defer sc.notePanic()
	defer sc.conn.Close()
	defer sc.closeAllStreamsOnConnClose()
	defer sc.stopShutdownTimer()
	defer close(sc.doneServing) // unblocks handlers trying to send

	if http2VerboseLogs {
		sc.vlogf("http2: server connection from %v on %p", sc.conn.RemoteAddr(), sc.hs)
	}

	sc.writeFrame(http2FrameWriteRequest{
		write: http2writeSettings{
			{http2SettingMaxFrameSize, sc.srv.maxReadFrameSize()},
			{http2SettingMaxConcurrentStreams, sc.advMaxStreams},
			{http2SettingMaxHeaderListSize, sc.maxHeaderListSize()},
			{http2SettingInitialWindowSize, uint32(sc.srv.initialStreamRecvWindowSize())},
		},
	})
	sc.unackedSettings++

	// Each connection starts with initialWindowSize inflow tokens.
	// If a higher value is configured, we add more tokens.
	if diff := sc.srv.initialConnRecvWindowSize() - http2initialWindowSize; diff > 0 {
		sc.sendWindowUpdate(nil, int(diff))
	}

	if err := sc.readPreface(); err != nil {
		sc.condlogf(err, "http2: server: error reading preface from client %v: %v", sc.conn.RemoteAddr(), err)
		return
	}
	// Now that we've got the preface, get us out of the
	// "StateNew" state. We can't go directly to idle, though.
	// Active means we read some data and anticipate a request. We'll
	// do another Active when we get a HEADERS frame.
	sc.setConnState(http.StateActive)
	sc.setConnState(http.StateIdle)

	if sc.srv.IdleTimeout != 0 {
		sc.idleTimer = time.AfterFunc(sc.srv.IdleTimeout, sc.onIdleTimer)
		defer sc.idleTimer.Stop()
	}

	go sc.readFrames() // closed by defer sc.conn.Close above

	settingsTimer := time.AfterFunc(http2firstSettingsTimeout, sc.onSettingsTimer)
	defer settingsTimer.Stop()

	loopNum := 0
	for {
		loopNum++
		select {
		case wr := <-sc.wantWriteFrameCh:
			if se, ok := wr.write.(http2StreamError); ok {
				sc.resetStream(se)
				break
			}
			sc.writeFrame(wr)
		case res := <-sc.wroteFrameCh:
			sc.wroteFrame(res)
		case res := <-sc.readFrameCh:
			// Process any written frames before reading new frames from the client since a
			// written frame could have triggered a new stream to be started.
			if sc.writingFrameAsync {
				select {
				case wroteRes := <-sc.wroteFrameCh:
					sc.wroteFrame(wroteRes)
				default:
				}
			}
			if !sc.processFrameFromReader(res) {
				return
			}
			res.readMore()
			if settingsTimer != nil {
				settingsTimer.Stop()
				settingsTimer = nil
			}
		case m := <-sc.bodyReadCh:
			sc.noteBodyRead(m.st, m.n)
		case msg := <-sc.serveMsgCh:
			switch v := msg.(type) {
			case func(int):
				v(loopNum) // for testing
			case *http2serverMessage:
				switch v {
				case http2settingsTimerMsg:
					sc.logf("timeout waiting for SETTINGS frames from %v", sc.conn.RemoteAddr())
					return
				case http2idleTimerMsg:
					sc.vlogf("connection is idle")
					sc.goAway(http2ErrCodeNo)
				case http2shutdownTimerMsg:
					sc.vlogf("GOAWAY close timer fired; closing conn from %v", sc.conn.RemoteAddr())
					return
				case http2gracefulShutdownMsg:
					sc.startGracefulShutdownInternal()
				default:
					panic("unknown timer")
				}
			case *http2startPushRequest:
				sc.startPush(v)
			default:
				panic(fmt.Sprintf("unexpected type %T", v))
			}
		}

		// If the peer is causing us to generate a lot of control frames,
		// but not reading them from us, assume they are trying to make us
		// run out of memory.
		if sc.queuedControlFrames > sc.srv.maxQueuedControlFrames() {
			sc.vlogf("http2: too many control frames in send queue, closing connection")
			return
		}

		// Start the shutdown timer after sending a GOAWAY. When sending GOAWAY
		// with no error code (graceful shutdown), don't start the timer until
		// all open streams have been completed.
		sentGoAway := sc.inGoAway && !sc.needToSendGoAway && !sc.writingFrame
		gracefulShutdownComplete := sc.goAwayCode == http2ErrCodeNo && sc.curOpenStreams() == 0
		if sentGoAway && sc.shutdownTimer == nil && (sc.goAwayCode != http2ErrCodeNo || gracefulShutdownComplete) {
			sc.shutDownIn(http2goAwayTimeout)
		}
	}
}

func (sc *http2serverConn) awaitGracefulShutdown(sharedCh <-chan struct{}, privateCh chan struct{}) {
	select {
	case <-sc.doneServing:
	case <-sharedCh:
		close(privateCh)
	}
}

type http2serverMessage int

// Message values sent to serveMsgCh.
var (
	http2settingsTimerMsg    = new(http2serverMessage)
	http2idleTimerMsg        = new(http2serverMessage)
	http2shutdownTimerMsg    = new(http2serverMessage)
	http2gracefulShutdownMsg = new(http2serverMessage)
)

func (sc *http2serverConn) onSettingsTimer() { sc.sendServeMsg(http2settingsTimerMsg) }

func (sc *http2serverConn) onIdleTimer() { sc.sendServeMsg(http2idleTimerMsg) }

func (sc *http2serverConn) onShutdownTimer() { sc.sendServeMsg(http2shutdownTimerMsg) }

func (sc *http2serverConn) sendServeMsg(msg interface{}) {
	sc.serveG.checkNotOn() // NOT
	select {
	case sc.serveMsgCh <- msg:
	case <-sc.doneServing:
	}
}

var http2errPrefaceTimeout = errors.New("timeout waiting for client preface")

// readPreface reads the ClientPreface greeting from the peer or
// returns errPrefaceTimeout on timeout, or an error if the greeting
// is invalid.
func (sc *http2serverConn) readPreface() error {
	errc := make(chan error, 1)
	go func() {
		// Read the client preface
		buf := make([]byte, len(http2ClientPreface))
		if _, err := io.ReadFull(sc.conn, buf); err != nil {
			errc <- err
		} else if !bytes.Equal(buf, http2clientPreface) {
			errc <- fmt.Errorf("bogus greeting %q", buf)
		} else {
			errc <- nil
		}
	}()
	timer := time.NewTimer(http2prefaceTimeout) // TODO: configurable on *http2Server?
	defer timer.Stop()
	select {
	case <-timer.C:
		return http2errPrefaceTimeout
	case err := <-errc:
		if err == nil {
			if http2VerboseLogs {
				sc.vlogf("http2: server: client %v said hello", sc.conn.RemoteAddr())
			}
		}
		return err
	}
}

var http2errChanPool = sync.Pool{
	New: func() interface{} { return make(chan error, 1) },
}

var http2writeDataPool = sync.Pool{
	New: func() interface{} { return new(http2writeData) },
}

// writeDataFromHandler writes DATA response frames from a handler on
// the given stream.
func (sc *http2serverConn) writeDataFromHandler(stream *http2stream, data []byte, endStream bool) error {
	ch := http2errChanPool.Get().(chan error)
	writeArg := http2writeDataPool.Get().(*http2writeData)
	*writeArg = http2writeData{stream.id, data, endStream}
	err := sc.writeFrameFromHandler(http2FrameWriteRequest{
		write:  writeArg,
		stream: stream,
		done:   ch,
	})
	if err != nil {
		return err
	}
	var frameWriteDone bool // the frame write is done (successfully or not)
	select {
	case err = <-ch:
		frameWriteDone = true
	case <-sc.doneServing:
		return http2errClientDisconnected
	case <-stream.cw:
		// If both ch and stream.cw were ready (as might
		// happen on the final Write after an http.Handler
		// ends), prefer the write result. Otherwise this
		// might just be us successfully closing the stream.
		// The writeFrameAsync and serve goroutines guarantee
		// that the ch send will happen before the stream.cw
		// close.
		select {
		case err = <-ch:
			frameWriteDone = true
		default:
			return http2errStreamClosed
		}
	}
	http2errChanPool.Put(ch)
	if frameWriteDone {
		http2writeDataPool.Put(writeArg)
	}
	return err
}

// writeFrameFromHandler sends wr to sc.wantWriteFrameCh, but aborts
// if the connection has gone away.
//
// This must not be run from the serve goroutine itself, else it might
// deadlock writing to sc.wantWriteFrameCh (which is only mildly
// buffered and is read by serve itself). If you're on the serve
// goroutine, call writeFrame instead.
func (sc *http2serverConn) writeFrameFromHandler(wr http2FrameWriteRequest) error {
	sc.serveG.checkNotOn() // NOT
	select {
	case sc.wantWriteFrameCh <- wr:
		return nil
	case <-sc.doneServing:
		// Serve loop is gone.
		// Client has closed their connection to the server.
		return http2errClientDisconnected
	}
}

// writeFrame schedules a frame to write and sends it if there's nothing
// already being written.
//
// There is no pushback here (the serve goroutine never blocks). It's
// the http.Handlers that block, waiting for their previous frames to
// make it onto the wire
//
// If you're not on the serve goroutine, use writeFrameFromHandler instead.
func (sc *http2serverConn) writeFrame(wr http2FrameWriteRequest) {
	sc.serveG.check()

	// If true, wr will not be written and wr.done will not be signaled.
	var ignoreWrite bool

	// We are not allowed to write frames on closed streams. RFC 7540 Section
	// 5.1.1 says: "An endpoint MUST NOT send frames other than PRIORITY on
	// a closed stream." Our server never sends PRIORITY, so that exception
	// does not apply.
	//
	// The serverConn might close an open stream while the stream's handler
	// is still running. For example, the server might close a stream when it
	// receives bad data from the client. If this happens, the handler might
	// attempt to write a frame after the stream has been closed (since the
	// handler hasn't yet been notified of the close). In this case, we simply
	// ignore the frame. The handler will notice that the stream is closed when
	// it waits for the frame to be written.
	//
	// As an exception to this rule, we allow sending RST_STREAM after close.
	// This allows us to immediately reject new streams without tracking any
	// state for those streams (except for the queued RST_STREAM frame). This
	// may result in duplicate RST_STREAMs in some cases, but the client should
	// ignore those.
	if wr.StreamID() != 0 {
		_, isReset := wr.write.(http2StreamError)
		if state, _ := sc.state(wr.StreamID()); state == http2stateClosed && !isReset {
			ignoreWrite = true
		}
	}

	// Don't send a 100-continue response if we've already sent headers.
	// See golang.org/issue/14030.
	switch wr.write.(type) {
	case *http2writeResHeaders:
		wr.stream.wroteHeaders = true
	case http2write100ContinueHeadersFrame:
		if wr.stream.wroteHeaders {
			// We do not need to notify wr.done because this frame is
			// never written with wr.done != nil.
			if wr.done != nil {
				panic("wr.done != nil for write100ContinueHeadersFrame")
			}
			ignoreWrite = true
		}
	}

	if !ignoreWrite {
		if wr.isControl() {
			sc.queuedControlFrames++
			// For extra safety, detect wraparounds, which should not happen,
			// and pull the plug.
			if sc.queuedControlFrames < 0 {
				sc.conn.Close()
			}
		}
		sc.writeSched.Push(wr)
	}
	sc.scheduleFrameWrite()
}

// startFrameWrite starts a goroutine to write wr (in a separate
// goroutine since that might block on the network), and updates the
// serve goroutine's state about the world, updated from info in wr.
func (sc *http2serverConn) startFrameWrite(wr http2FrameWriteRequest) {
	sc.serveG.check()
	if sc.writingFrame {
		panic("internal error: can only be writing one frame at a time")
	}

	st := wr.stream
	if st != nil {
		switch st.state {
		case http2stateHalfClosedLocal:
			switch wr.write.(type) {
			case http2StreamError, http2handlerPanicRST, http2writeWindowUpdate:
				// RFC 7540 Section 5.1 allows sending RST_STREAM, PRIORITY, and WINDOW_UPDATE
				// in this state. (We never send PRIORITY from the server, so that is not checked.)
			default:
				panic(fmt.Sprintf("internal error: attempt to send frame on a half-closed-local stream: %v", wr))
			}
		case http2stateClosed:
			panic(fmt.Sprintf("internal error: attempt to send frame on a closed stream: %v", wr))
		}
	}
	if wpp, ok := wr.write.(*http2writePushPromise); ok {
		var err error
		wpp.promisedID, err = wpp.allocatePromisedID()
		if err != nil {
			sc.writingFrameAsync = false
			wr.replyToWriter(err)
			return
		}
	}

	sc.writingFrame = true
	sc.needsFrameFlush = true
	if wr.write.staysWithinBuffer(sc.bw.Available()) {
		sc.writingFrameAsync = false
		err := wr.write.writeFrame(sc)
		sc.wroteFrame(http2frameWriteResult{wr: wr, err: err})
	} else {
		sc.writingFrameAsync = true
		go sc.writeFrameAsync(wr)
	}
}

// errHandlerPanicked is the error given to any callers blocked in a read from
// Request.Body when the main goroutine panics. Since most handlers read in the
// main ServeHTTP goroutine, this will show up rarely.
var http2errHandlerPanicked = errors.New("http2: handler panicked")

// wroteFrame is called on the serve goroutine with the result of
// whatever happened on writeFrameAsync.
func (sc *http2serverConn) wroteFrame(res http2frameWriteResult) {
	sc.serveG.check()
	if !sc.writingFrame {
		panic("internal error: expected to be already writing a frame")
	}
	sc.writingFrame = false
	sc.writingFrameAsync = false

	wr := res.wr

	if http2writeEndsStream(wr.write) {
		st := wr.stream
		if st == nil {
			panic("internal error: expecting non-nil stream")
		}
		switch st.state {
		case http2stateOpen:
			// Here we would go to stateHalfClosedLocal in
			// theory, but since our handler is done and
			// the net/http package provides no mechanism
			// for closing a ResponseWriter while still
			// reading data (see possible TODO at top of
			// this file), we go into closed state here
			// anyway, after telling the peer we're
			// hanging up on them. We'll transition to
			// stateClosed after the RST_STREAM frame is
			// written.
			st.state = http2stateHalfClosedLocal
			// Section 8.1: a server MAY request that the client abort
			// transmission of a request without error by sending a
			// RST_STREAM with an error code of NO_ERROR after sending
			// a complete response.
			sc.resetStream(http2streamError(st.id, http2ErrCodeNo))
		case http2stateHalfClosedRemote:
			sc.closeStream(st, http2errHandlerComplete)
		}
	} else {
		switch v := wr.write.(type) {
		case http2StreamError:
			// st may be unknown if the RST_STREAM was generated to reject bad input.
			if st, ok := sc.streams[v.StreamID]; ok {
				sc.closeStream(st, v)
			}
		case http2handlerPanicRST:
			sc.closeStream(wr.stream, http2errHandlerPanicked)
		}
	}

	// Reply (if requested) to unblock the ServeHTTP goroutine.
	wr.replyToWriter(res.err)

	sc.scheduleFrameWrite()
}

// scheduleFrameWrite tickles the frame writing scheduler.
//
// If a frame is already being written, nothing happens. This will be called again
// when the frame is done being written.
//
// If a frame isn't being written and we need to send one, the best frame
// to send is selected by writeSched.
//
// If a frame isn't being written and there's nothing else to send, we
// flush the write buffer.
func (sc *http2serverConn) scheduleFrameWrite() {
	sc.serveG.check()
	if sc.writingFrame || sc.inFrameScheduleLoop {
		return
	}
	sc.inFrameScheduleLoop = true
	for !sc.writingFrameAsync {
		if sc.needToSendGoAway {
			sc.needToSendGoAway = false
			sc.startFrameWrite(http2FrameWriteRequest{
				write: &http2writeGoAway{
					maxStreamID: sc.maxClientStreamID,
					code:        sc.goAwayCode,
				},
			})
			continue
		}
		if sc.needToSendSettingsAck {
			sc.needToSendSettingsAck = false
			sc.startFrameWrite(http2FrameWriteRequest{write: http2writeSettingsAck{}})
			continue
		}
		if !sc.inGoAway || sc.goAwayCode == http2ErrCodeNo {
			if wr, ok := sc.writeSched.Pop(); ok {
				if wr.isControl() {
					sc.queuedControlFrames--
				}
				sc.startFrameWrite(wr)
				continue
			}
		}
		if sc.needsFrameFlush {
			sc.startFrameWrite(http2FrameWriteRequest{write: http2flushFrameWriter{}})
			sc.needsFrameFlush = false // after startFrameWrite, since it sets this true
			continue
		}
		break
	}
	sc.inFrameScheduleLoop = false
}

// startGracefulShutdown gracefully shuts down a connection. This
// sends GOAWAY with http2ErrCodeNo to tell the client we're gracefully
// shutting down. The connection isn't closed until all current
// streams are done.
//
// startGracefulShutdown returns immediately; it does not wait until
// the connection has shut down.
func (sc *http2serverConn) startGracefulShutdown() {
	sc.serveG.checkNotOn() // NOT
	sc.shutdownOnce.Do(func() { sc.sendServeMsg(http2gracefulShutdownMsg) })
}

// After sending GOAWAY with an error code (non-graceful shutdown), the
// connection will close after goAwayTimeout.
//
// If we close the connection immediately after sending GOAWAY, there may
// be unsent data in our kernel receive buffer, which will cause the kernel
// to send a TCP RST on close() instead of a FIN. This RST will abort the
// connection immediately, whether or not the client had received the GOAWAY.
//
// Ideally we should delay for at least 1 RTT + epsilon so the client has
// a chance to read the GOAWAY and stop sending messages. Measuring RTT
// is hard, so we approximate with 1 second. See golang.org/issue/18701.
//
// This is a var so it can be shorter in tests, where all requests uses the
// loopback interface making the expected RTT very small.
//
// TODO: configurable?
var http2goAwayTimeout = 1 * time.Second

func (sc *http2serverConn) startGracefulShutdownInternal() {
	sc.goAway(http2ErrCodeNo)
}

func (sc *http2serverConn) goAway(code http2ErrCode) {
	sc.serveG.check()
	if sc.inGoAway {
		return
	}
	sc.inGoAway = true
	sc.needToSendGoAway = true
	sc.goAwayCode = code
	sc.scheduleFrameWrite()
}

func (sc *http2serverConn) shutDownIn(d time.Duration) {
	sc.serveG.check()
	sc.shutdownTimer = time.AfterFunc(d, sc.onShutdownTimer)
}

func (sc *http2serverConn) resetStream(se http2StreamError) {
	sc.serveG.check()
	sc.writeFrame(http2FrameWriteRequest{write: se})
	if st, ok := sc.streams[se.StreamID]; ok {
		st.resetQueued = true
	}
}

// 6.9.1 The Flow Control Window
// "If a sender receives a WINDOW_UPDATE that causes a flow control
// window to exceed this maximum it MUST terminate either the stream
// or the connection, as appropriate. For streams, [...]; for the
// connection, a GOAWAY frame with a FLOW_CONTROL_ERROR code."
type http2goAwayFlowError struct{}

func (http2goAwayFlowError) Error() string { return "connection exceeded flow control window size" }

// processFrameFromReader processes the serve loop's read from readFrameCh from the
// frame-reading goroutine.
// processFrameFromReader returns whether the connection should be kept open.
func (sc *http2serverConn) processFrameFromReader(res http2readFrameResult) bool {
	sc.serveG.check()
	err := res.err
	if err != nil {
		if err == http2ErrFrameTooLarge {
			sc.goAway(http2ErrCodeFrameSize)
			return true // goAway will close the loop
		}
		clientGone := err == io.EOF || err == io.ErrUnexpectedEOF || http2isClosedConnError(err)
		if clientGone {
			// TODO: could we also get into this state if
			// the peer does a half close
			// (e.g. CloseWrite) because they're done
			// sending frames but they're still wanting
			// our open replies?  Investigate.
			// TODO: add CloseWrite to crypto/tls.Conn first
			// so we have a way to test this? I suppose
			// just for testing we could have a non-TLS mode.
			return false
		}
	} else {
		f := res.f
		if http2VerboseLogs {
			sc.vlogf("http2: server read frame %v", http2summarizeFrame(f))
		}
		err = sc.processFrame(f)
		if err == nil {
			return true
		}
	}

	switch ev := err.(type) {
	case http2StreamError:
		sc.resetStream(ev)
		return true
	case http2goAwayFlowError:
		sc.goAway(http2ErrCodeFlowControl)
		return true
	case http2ConnectionError:
		sc.logf("http2: server connection error from %v: %v", sc.conn.RemoteAddr(), ev)
		sc.goAway(http2ErrCode(ev))
		return true // goAway will handle shutdown
	default:
		if res.err != nil {
			sc.vlogf("http2: server closing client connection; error reading frame from client %s: %v", sc.conn.RemoteAddr(), err)
		} else {
			sc.logf("http2: server closing client connection: %v", err)
		}
		return false
	}
}

func (sc *http2serverConn) processFrame(f http2Frame) error {
	sc.serveG.check()

	// First frame received must be SETTINGS.
	if !sc.sawFirstSettings {
		if _, ok := f.(*http2SettingsFrame); !ok {
			return sc.countError("first_settings", http2ConnectionError(http2ErrCodeProtocol))
		}
		sc.sawFirstSettings = true
	}

	switch f := f.(type) {
	case *http2SettingsFrame:
		return sc.processSettings(f)
	case *http2MetaHeadersFrame:
		return sc.processHeaders(f)
	case *http2WindowUpdateFrame:
		return sc.processWindowUpdate(f)
	case *http2PingFrame:
		return sc.processPing(f)
	case *http2DataFrame:
		return sc.processData(f)
	case *http2RSTStreamFrame:
		return sc.processResetStream(f)
	case *http2PriorityFrame:
		return sc.processPriority(f)
	case *http2GoAwayFrame:
		return sc.processGoAway(f)
	case *http2PushPromiseFrame:
		// A client cannot push. Thus, servers MUST treat the receipt of a PUSH_PROMISE
		// frame as a connection error (Section 5.4.1) of type PROTOCOL_ERROR.
		return sc.countError("push_promise", http2ConnectionError(http2ErrCodeProtocol))
	default:
		sc.vlogf("http2: server ignoring frame: %v", f.Header())
		return nil
	}
}

func (sc *http2serverConn) processPing(f *http2PingFrame) error {
	sc.serveG.check()
	if f.IsAck() {
		// 6.7 PING: " An endpoint MUST NOT respond to PING frames
		// containing this flag."
		return nil
	}
	if f.StreamID != 0 {
		// "PING frames are not associated with any individual
		// stream. If a PING frame is received with a stream
		// identifier field value other than 0x0, the recipient MUST
		// respond with a connection error (Section 5.4.1) of type
		// PROTOCOL_ERROR."
		return sc.countError("ping_on_stream", http2ConnectionError(http2ErrCodeProtocol))
	}
	if sc.inGoAway && sc.goAwayCode != http2ErrCodeNo {
		return nil
	}
	sc.writeFrame(http2FrameWriteRequest{write: http2writePingAck{f}})
	return nil
}

func (sc *http2serverConn) processWindowUpdate(f *http2WindowUpdateFrame) error {
	sc.serveG.check()
	switch {
	case f.StreamID != 0: // stream-level flow control
		state, st := sc.state(f.StreamID)
		if state == http2stateIdle {
			// Section 5.1: "Receiving any frame other than HEADERS
			// or PRIORITY on a stream in this state MUST be
			// treated as a connection error (Section 5.4.1) of
			// type PROTOCOL_ERROR."
			return sc.countError("stream_idle", http2ConnectionError(http2ErrCodeProtocol))
		}
		if st == nil {
			// "WINDOW_UPDATE can be sent by a peer that has sent a
			// frame bearing the END_STREAM flag. This means that a
			// receiver could receive a WINDOW_UPDATE frame on a "half
			// closed (remote)" or "closed" stream. A receiver MUST
			// NOT treat this as an error, see Section 5.1."
			return nil
		}
		if !st.flow.add(int32(f.Increment)) {
			return sc.countError("bad_flow", http2streamError(f.StreamID, http2ErrCodeFlowControl))
		}
	default: // connection-level flow control
		if !sc.flow.add(int32(f.Increment)) {
			return http2goAwayFlowError{}
		}
	}
	sc.scheduleFrameWrite()
	return nil
}

func (sc *http2serverConn) processResetStream(f *http2RSTStreamFrame) error {
	sc.serveG.check()

	state, st := sc.state(f.StreamID)
	if state == http2stateIdle {
		// 6.4 "RST_STREAM frames MUST NOT be sent for a
		// stream in the "idle" state. If a RST_STREAM frame
		// identifying an idle stream is received, the
		// recipient MUST treat this as a connection error
		// (Section 5.4.1) of type PROTOCOL_ERROR.
		return sc.countError("reset_idle_stream", http2ConnectionError(http2ErrCodeProtocol))
	}
	if st != nil {
		st.cancelCtx()
		sc.closeStream(st, http2streamError(f.StreamID, f.ErrCode))
	}
	return nil
}

func (sc *http2serverConn) closeStream(st *http2stream, err error) {
	sc.serveG.check()
	if st.state == http2stateIdle || st.state == http2stateClosed {
		panic(fmt.Sprintf("invariant; can't close stream in state %v", st.state))
	}
	st.state = http2stateClosed
	if st.writeDeadline != nil {
		st.writeDeadline.Stop()
	}
	if st.isPushed() {
		sc.curPushedStreams--
	} else {
		sc.curClientStreams--
	}
	delete(sc.streams, st.id)
	if len(sc.streams) == 0 {
		sc.setConnState(http.StateIdle)
		if sc.srv.IdleTimeout != 0 {
			sc.idleTimer.Reset(sc.srv.IdleTimeout)
		}
		if http2h1ServerKeepAlivesDisabled(sc.hs) {
			sc.startGracefulShutdownInternal()
		}
	}
	if p := st.body; p != nil {
		// Return any buffered unread bytes worth of conn-level flow control.
		// See golang.org/issue/16481
		sc.sendWindowUpdate(nil, p.Len())

		p.CloseWithError(err)
	}
	st.cw.Close() // signals Handler's CloseNotifier, unblocks writes, etc
	sc.writeSched.CloseStream(st.id)
}

func (sc *http2serverConn) processSettings(f *http2SettingsFrame) error {
	sc.serveG.check()
	if f.IsAck() {
		sc.unackedSettings--
		if sc.unackedSettings < 0 {
			// Why is the peer ACKing settings we never sent?
			// The spec doesn't mention this case, but
			// hang up on them anyway.
			return sc.countError("ack_mystery", http2ConnectionError(http2ErrCodeProtocol))
		}
		return nil
	}
	if f.NumSettings() > 100 || f.HasDuplicates() {
		// This isn't actually in the spec, but hang up on
		// suspiciously large settings frames or those with
		// duplicate entries.
		return sc.countError("settings_big_or_dups", http2ConnectionError(http2ErrCodeProtocol))
	}
	if err := f.ForeachSetting(sc.processSetting); err != nil {
		return err
	}
	// TODO: judging by RFC 7540, Section 6.5.3 each SETTINGS frame should be
	// acknowledged individually, even if multiple are received before the ACK.
	sc.needToSendSettingsAck = true
	sc.scheduleFrameWrite()
	return nil
}

func (sc *http2serverConn) processSetting(s http2Setting) error {
	sc.serveG.check()
	if err := s.Valid(); err != nil {
		return err
	}
	if http2VerboseLogs {
		sc.vlogf("http2: server processing setting %v", s)
	}
	switch s.ID {
	case http2SettingHeaderTableSize:
		sc.headerTableSize = s.Val
		sc.hpackEncoder.SetMaxDynamicTableSize(s.Val)
	case http2SettingEnablePush:
		sc.pushEnabled = s.Val != 0
	case http2SettingMaxConcurrentStreams:
		sc.clientMaxStreams = s.Val
	case http2SettingInitialWindowSize:
		return sc.processSettingInitialWindowSize(s.Val)
	case http2SettingMaxFrameSize:
		sc.maxFrameSize = int32(s.Val) // the maximum valid s.Val is < 2^31
	case http2SettingMaxHeaderListSize:
		sc.peerMaxHeaderListSize = s.Val
	default:
		// Unknown setting: "An endpoint that receives a SETTINGS
		// frame with any unknown or unsupported identifier MUST
		// ignore that setting."
		if http2VerboseLogs {
			sc.vlogf("http2: server ignoring unknown setting %v", s)
		}
	}
	return nil
}

func (sc *http2serverConn) processSettingInitialWindowSize(val uint32) error {
	sc.serveG.check()
	// Note: val already validated to be within range by
	// processSetting's Valid call.

	// "A SETTINGS frame can alter the initial flow control window
	// size for all current streams. When the value of
	// SETTINGS_INITIAL_WINDOW_SIZE changes, a receiver MUST
	// adjust the size of all stream flow control windows that it
	// maintains by the difference between the new value and the
	// old value."
	old := sc.initialStreamSendWindowSize
	sc.initialStreamSendWindowSize = int32(val)
	growth := int32(val) - old // may be negative
	for _, st := range sc.streams {
		if !st.flow.add(growth) {
			// 6.9.2 Initial Flow Control Window Size
			// "An endpoint MUST treat a change to
			// SETTINGS_INITIAL_WINDOW_SIZE that causes any flow
			// control window to exceed the maximum size as a
			// connection error (Section 5.4.1) of type
			// FLOW_CONTROL_ERROR."
			return sc.countError("setting_win_size", http2ConnectionError(http2ErrCodeFlowControl))
		}
	}
	return nil
}

func (sc *http2serverConn) processData(f *http2DataFrame) error {
	sc.serveG.check()
	id := f.Header().StreamID
	if sc.inGoAway && (sc.goAwayCode != http2ErrCodeNo || id > sc.maxClientStreamID) {
		// Discard all DATA frames if the GOAWAY is due to an
		// error, or:
		//
		// Section 6.8: After sending a GOAWAY frame, the sender
		// can discard frames for streams initiated by the
		// receiver with identifiers higher than the identified
		// last stream.
		return nil
	}

	data := f.Data()
	state, st := sc.state(id)
	if id == 0 || state == http2stateIdle {
		// Section 6.1: "DATA frames MUST be associated with a
		// stream. If a DATA frame is received whose stream
		// identifier field is 0x0, the recipient MUST respond
		// with a connection error (Section 5.4.1) of type
		// PROTOCOL_ERROR."
		//
		// Section 5.1: "Receiving any frame other than HEADERS
		// or PRIORITY on a stream in this state MUST be
		// treated as a connection error (Section 5.4.1) of
		// type PROTOCOL_ERROR."
		return sc.countError("data_on_idle", http2ConnectionError(http2ErrCodeProtocol))
	}

	// "If a DATA frame is received whose stream is not in "open"
	// or "half closed (local)" state, the recipient MUST respond
	// with a stream error (Section 5.4.2) of type STREAM_CLOSED."
	if st == nil || state != http2stateOpen || st.gotTrailerHeader || st.resetQueued {
		// This includes sending a RST_STREAM if the stream is
		// in stateHalfClosedLocal (which currently means that
		// the http.Handler returned, so it's done reading &
		// done writing). Try to stop the client from sending
		// more DATA.

		// But still enforce their connection-level flow control,
		// and return any flow control bytes since we're not going
		// to consume them.
		if sc.inflow.available() < int32(f.Length) {
			return sc.countError("data_flow", http2streamError(id, http2ErrCodeFlowControl))
		}
		// Deduct the flow control from inflow, since we're
		// going to immediately add it back in
		// sendWindowUpdate, which also schedules sending the
		// frames.
		sc.inflow.take(int32(f.Length))
		sc.sendWindowUpdate(nil, int(f.Length)) // conn-level

		if st != nil && st.resetQueued {
			// Already have a stream error in flight. Don't send another.
			return nil
		}
		return sc.countError("closed", http2streamError(id, http2ErrCodeStreamClosed))
	}
	if st.body == nil {
		panic("internal error: should have a body in this state")
	}

	// Sender sending more than they'd declared?
	if st.declBodyBytes != -1 && st.bodyBytes+int64(len(data)) > st.declBodyBytes {
		st.body.CloseWithError(fmt.Errorf("sender tried to send more than declared Content-Length of %d bytes", st.declBodyBytes))
		// RFC 7540, sec 8.1.2.6: A request or response is also malformed if the
		// value of a content-length header field does not equal the sum of the
		// DATA frame payload lengths that form the body.
		return sc.countError("send_too_much", http2streamError(id, http2ErrCodeProtocol))
	}
	if f.Length > 0 {
		// Check whether the client has flow control quota.
		if st.inflow.available() < int32(f.Length) {
			return sc.countError("flow_on_data_length", http2streamError(id, http2ErrCodeFlowControl))
		}
		st.inflow.take(int32(f.Length))

		if len(data) > 0 {
			wrote, err := st.body.Write(data)
			if err != nil {
				sc.sendWindowUpdate(nil, int(f.Length)-wrote)
				return sc.countError("body_write_err", http2streamError(id, http2ErrCodeStreamClosed))
			}
			if wrote != len(data) {
				panic("internal error: bad Writer")
			}
			st.bodyBytes += int64(len(data))
		}

		// Return any padded flow control now, since we won't
		// refund it later on body reads.
		if pad := int32(f.Length) - int32(len(data)); pad > 0 {
			sc.sendWindowUpdate32(nil, pad)
			sc.sendWindowUpdate32(st, pad)
		}
	}
	if f.StreamEnded() {
		st.endStream()
	}
	return nil
}

func (sc *http2serverConn) processGoAway(f *http2GoAwayFrame) error {
	sc.serveG.check()
	if f.ErrCode != http2ErrCodeNo {
		sc.logf("http2: received GOAWAY %+v, starting graceful shutdown", f)
	} else {
		sc.vlogf("http2: received GOAWAY %+v, starting graceful shutdown", f)
	}
	sc.startGracefulShutdownInternal()
	// http://tools.ietf.org/html/rfc7540#section-6.8
	// We should not create any new streams, which means we should disable push.
	sc.pushEnabled = false
	return nil
}

// isPushed reports whether the stream is server-initiated.
func (st *http2stream) isPushed() bool {
	return st.id%2 == 0
}

// endStream closes a Request.Body's pipe. It is called when a DATA
// frame says a request body is over (or after trailers).
func (st *http2stream) endStream() {
	sc := st.sc
	sc.serveG.check()

	if st.declBodyBytes != -1 && st.declBodyBytes != st.bodyBytes {
		st.body.CloseWithError(fmt.Errorf("request declared a Content-Length of %d but only wrote %d bytes",
			st.declBodyBytes, st.bodyBytes))
	} else {
		st.body.closeWithErrorAndCode(io.EOF, st.copyTrailersToHandlerRequest)
		st.body.CloseWithError(io.EOF)
	}
	st.state = http2stateHalfClosedRemote
}

// copyTrailersToHandlerRequest is run in the Handler's goroutine in
// its Request.Body.Read just before it gets io.EOF.
func (st *http2stream) copyTrailersToHandlerRequest() {
	for k, vv := range st.trailer {
		if _, ok := st.reqTrailer[k]; ok {
			// Only copy it over it was pre-declared.
			st.reqTrailer[k] = vv
		}
	}
}

// onWriteTimeout is run on its own goroutine (from time.AfterFunc)
// when the stream's WriteTimeout has fired.
func (st *http2stream) onWriteTimeout() {
	st.sc.writeFrameFromHandler(http2FrameWriteRequest{write: http2streamError(st.id, http2ErrCodeInternal)})
}

func (sc *http2serverConn) processHeaders(f *http2MetaHeadersFrame) error {
	sc.serveG.check()
	id := f.StreamID
	if sc.inGoAway {
		// Ignore.
		return nil
	}
	// http://tools.ietf.org/html/rfc7540#section-5.1.1
	// Streams initiated by a client MUST use odd-numbered stream
	// identifiers. [...] An endpoint that receives an unexpected
	// stream identifier MUST respond with a connection error
	// (Section 5.4.1) of type PROTOCOL_ERROR.
	if id%2 != 1 {
		return sc.countError("headers_even", http2ConnectionError(http2ErrCodeProtocol))
	}
	// A HEADERS frame can be used to create a new stream or
	// send a trailer for an open one. If we already have a stream
	// open, let it process its own HEADERS frame (trailers at this
	// point, if it's valid).
	if st := sc.streams[f.StreamID]; st != nil {
		if st.resetQueued {
			// We're sending RST_STREAM to close the stream, so don't bother
			// processing this frame.
			return nil
		}
		// RFC 7540, sec 5.1: If an endpoint receives additional frames, other than
		// WINDOW_UPDATE, PRIORITY, or RST_STREAM, for a stream that is in
		// this state, it MUST respond with a stream error (Section 5.4.2) of
		// type STREAM_CLOSED.
		if st.state == http2stateHalfClosedRemote {
			return sc.countError("headers_half_closed", http2streamError(id, http2ErrCodeStreamClosed))
		}
		return st.processTrailerHeaders(f)
	}

	// [...] The identifier of a newly established stream MUST be
	// numerically greater than all streams that the initiating
	// endpoint has opened or reserved. [...]  An endpoint that
	// receives an unexpected stream identifier MUST respond with
	// a connection error (Section 5.4.1) of type PROTOCOL_ERROR.
	if id <= sc.maxClientStreamID {
		return sc.countError("stream_went_down", http2ConnectionError(http2ErrCodeProtocol))
	}
	sc.maxClientStreamID = id

	if sc.idleTimer != nil {
		sc.idleTimer.Stop()
	}

	// http://tools.ietf.org/html/rfc7540#section-5.1.2
	// [...] Endpoints MUST NOT exceed the limit set by their peer. An
	// endpoint that receives a HEADERS frame that causes their
	// advertised concurrent stream limit to be exceeded MUST treat
	// this as a stream error (Section 5.4.2) of type PROTOCOL_ERROR
	// or REFUSED_STREAM.
	if sc.curClientStreams+1 > sc.advMaxStreams {
		if sc.unackedSettings == 0 {
			// They should know better.
			return sc.countError("over_max_streams", http2streamError(id, http2ErrCodeProtocol))
		}
		// Assume it's a network race, where they just haven't
		// received our last SETTINGS update. But actually
		// this can't happen yet, because we don't yet provide
		// a way for users to adjust server parameters at
		// runtime.
		return sc.countError("over_max_streams_race", http2streamError(id, http2ErrCodeRefusedStream))
	}

	initialState := http2stateOpen
	if f.StreamEnded() {
		initialState = http2stateHalfClosedRemote
	}
	st := sc.newStream(id, 0, initialState)

	if f.HasPriority() {
		if err := sc.checkPriority(f.StreamID, f.Priority); err != nil {
			return err
		}
		sc.writeSched.AdjustStream(st.id, f.Priority)
	}

	rw, req, err := sc.newWriterAndRequest(st, f)
	if err != nil {
		return err
	}
	st.reqTrailer = req.Trailer
	if st.reqTrailer != nil {
		st.trailer = make(http.Header)
	}
	st.body = req.Body.(*http2requestBody).pipe // may be nil
	st.declBodyBytes = req.ContentLength

	handler := sc.handler.ServeHTTP
	if f.Truncated {
		// Their header list was too long. Send a 431 error.
		handler = http2handleHeaderListTooLong
	} else if err := http2checkValidHTTP2RequestHeaders(req.Header); err != nil {
		handler = http2new400Handler(err)
	}

	// The net/http package sets the read deadline from the
	// http.Server.ReadTimeout during the TLS handshake, but then
	// passes the connection off to us with the deadline already
	// set. Disarm it here after the request headers are read,
	// similar to how the http1 server works. Here it's
	// technically more like the http1 Server's ReadHeaderTimeout
	// (in Go 1.8), though. That's a more sane option anyway.
	if sc.hs.ReadTimeout != 0 {
		sc.conn.SetReadDeadline(time.Time{})
	}

	go sc.runHandler(rw, req, handler)
	return nil
}

func (st *http2stream) processTrailerHeaders(f *http2MetaHeadersFrame) error {
	sc := st.sc
	sc.serveG.check()
	if st.gotTrailerHeader {
		return sc.countError("dup_trailers", http2ConnectionError(http2ErrCodeProtocol))
	}
	st.gotTrailerHeader = true
	if !f.StreamEnded() {
		return sc.countError("trailers_not_ended", http2streamError(st.id, http2ErrCodeProtocol))
	}

	if len(f.PseudoFields()) > 0 {
		return sc.countError("trailers_pseudo", http2streamError(st.id, http2ErrCodeProtocol))
	}
	if st.trailer != nil {
		for _, hf := range f.RegularFields() {
			key := sc.canonicalHeader(hf.Name)
			if !httpguts.ValidTrailerHeader(key) {
				// TODO: send more details to the peer somehow. But http2 has
				// no way to send debug data at a stream level. Discuss with
				// HTTP folk.
				return sc.countError("trailers_bogus", http2streamError(st.id, http2ErrCodeProtocol))
			}
			st.trailer[key] = append(st.trailer[key], hf.Value)
		}
	}
	st.endStream()
	return nil
}

func (sc *http2serverConn) checkPriority(streamID uint32, p http2PriorityParam) error {
	if streamID == p.StreamDep {
		// Section 5.3.1: "A stream cannot depend on itself. An endpoint MUST treat
		// this as a stream error (Section 5.4.2) of type PROTOCOL_ERROR."
		// Section 5.3.3 says that a stream can depend on one of its dependencies,
		// so it's only self-dependencies that are forbidden.
		return sc.countError("priority", http2streamError(streamID, http2ErrCodeProtocol))
	}
	return nil
}

func (sc *http2serverConn) processPriority(f *http2PriorityFrame) error {
	if sc.inGoAway {
		return nil
	}
	if err := sc.checkPriority(f.StreamID, f.http2PriorityParam); err != nil {
		return err
	}
	sc.writeSched.AdjustStream(f.StreamID, f.http2PriorityParam)
	return nil
}

func (sc *http2serverConn) newStream(id, pusherID uint32, state http2streamState) *http2stream {
	sc.serveG.check()
	if id == 0 {
		panic("internal error: cannot create stream with id 0")
	}

	ctx, cancelCtx := context.WithCancel(sc.baseCtx)
	st := &http2stream{
		sc:        sc,
		id:        id,
		state:     state,
		ctx:       ctx,
		cancelCtx: cancelCtx,
	}
	st.cw.Init()
	st.flow.conn = &sc.flow // link to conn-level counter
	st.flow.add(sc.initialStreamSendWindowSize)
	st.inflow.conn = &sc.inflow // link to conn-level counter
	st.inflow.add(sc.srv.initialStreamRecvWindowSize())
	if sc.hs.WriteTimeout != 0 {
		st.writeDeadline = time.AfterFunc(sc.hs.WriteTimeout, st.onWriteTimeout)
	}

	sc.streams[id] = st
	sc.writeSched.OpenStream(st.id, http2OpenStreamOptions{PusherID: pusherID})
	if st.isPushed() {
		sc.curPushedStreams++
	} else {
		sc.curClientStreams++
	}
	if sc.curOpenStreams() == 1 {
		sc.setConnState(http.StateActive)
	}

	return st
}

func (sc *http2serverConn) newWriterAndRequest(st *http2stream, f *http2MetaHeadersFrame) (*http2responseWriter, *http.Request, error) {
	sc.serveG.check()

	rp := http2requestParam{
		method:    f.PseudoValue("method"),
		scheme:    f.PseudoValue("scheme"),
		authority: f.PseudoValue("authority"),
		path:      f.PseudoValue("path"),
	}

	isConnect := rp.method == "CONNECT"
	if isConnect {
		if rp.path != "" || rp.scheme != "" || rp.authority == "" {
			return nil, nil, sc.countError("bad_connect", http2streamError(f.StreamID, http2ErrCodeProtocol))
		}
	} else if rp.method == "" || rp.path == "" || (rp.scheme != "https" && rp.scheme != "http") {
		// See 8.1.2.6 Malformed Requests and Responses:
		//
		// Malformed requests or responses that are detected
		// MUST be treated as a stream error (Section 5.4.2)
		// of type PROTOCOL_ERROR."
		//
		// 8.1.2.3 Request Pseudo-Header Fields
		// "All HTTP/2 requests MUST include exactly one valid
		// value for the :method, :scheme, and :path
		// pseudo-header fields"
		return nil, nil, sc.countError("bad_path_method", http2streamError(f.StreamID, http2ErrCodeProtocol))
	}

	bodyOpen := !f.StreamEnded()
	if rp.method == "HEAD" && bodyOpen {
		// HEAD requests can't have bodies
		return nil, nil, sc.countError("head_body", http2streamError(f.StreamID, http2ErrCodeProtocol))
	}

	rp.header = make(http.Header)
	for _, hf := range f.RegularFields() {
		rp.header.Add(sc.canonicalHeader(hf.Name), hf.Value)
	}
	if rp.authority == "" {
		rp.authority = rp.header.Get("Host")
	}

	rw, req, err := sc.newWriterAndRequestNoBody(st, rp)
	if err != nil {
		return nil, nil, err
	}
	if bodyOpen {
		if vv, ok := rp.header["Content-Length"]; ok {
			if cl, err := strconv.ParseUint(vv[0], 10, 63); err == nil {
				req.ContentLength = int64(cl)
			} else {
				req.ContentLength = 0
			}
		} else {
			req.ContentLength = -1
		}
		req.Body.(*http2requestBody).pipe = &http2pipe{
			b: &http2dataBuffer{expected: req.ContentLength},
		}
	}
	return rw, req, nil
}

type http2requestParam struct {
	method                  string
	scheme, authority, path string
	header                  http.Header
}

func (sc *http2serverConn) newWriterAndRequestNoBody(st *http2stream, rp http2requestParam) (*http2responseWriter, *http.Request, error) {
	sc.serveG.check()

	var tlsState *tls.ConnectionState // nil if not scheme https
	if rp.scheme == "https" {
		tlsState = sc.tlsState
	}

	needsContinue := rp.header.Get("Expect") == "100-continue"
	if needsContinue {
		rp.header.Del("Expect")
	}
	// Merge Cookie headers into one "; "-delimited value.
	if cookies := rp.header["Cookie"]; len(cookies) > 1 {
		rp.header.Set("Cookie", strings.Join(cookies, "; "))
	}

	// Setup Trailers
	var trailer http.Header
	for _, v := range rp.header["Trailer"] {
		for _, key := range strings.Split(v, ",") {
			key = http.CanonicalHeaderKey(textproto.TrimString(key))
			switch key {
			case "Transfer-Encoding", "Trailer", "Content-Length":
				// Bogus. (copy of http1 rules)
				// Ignore.
			default:
				if trailer == nil {
					trailer = make(http.Header)
				}
				trailer[key] = nil
			}
		}
	}
	delete(rp.header, "Trailer")

	var url_ *url.URL
	var requestURI string
	if rp.method == "CONNECT" {
		url_ = &url.URL{Host: rp.authority}
		requestURI = rp.authority // mimic HTTP/1 server behavior
	} else {
		var err error
		url_, err = url.ParseRequestURI(rp.path)
		if err != nil {
			return nil, nil, sc.countError("bad_path", http2streamError(st.id, http2ErrCodeProtocol))
		}
		requestURI = rp.path
	}

	body := &http2requestBody{
		conn:          sc,
		stream:        st,
		needsContinue: needsContinue,
	}
	req := &http.Request{
		Method:     rp.method,
		URL:        url_,
		RemoteAddr: sc.remoteAddrStr,
		Header:     rp.header,
		RequestURI: requestURI,
		Proto:      "HTTP/2.0",
		ProtoMajor: 2,
		ProtoMinor: 0,
		TLS:        tlsState,
		Host:       rp.authority,
		Body:       body,
		Trailer:    trailer,
	}
	req = req.WithContext(st.ctx)

	rws := http2responseWriterStatePool.Get().(*http2responseWriterState)
	bwSave := rws.bw
	*rws = http2responseWriterState{} // zero all the fields
	rws.conn = sc
	rws.bw = bwSave
	rws.bw.Reset(http2chunkWriter{rws})
	rws.stream = st
	rws.req = req
	rws.body = body

	rw := &http2responseWriter{rws: rws}
	return rw, req, nil
}

// Run on its own goroutine.
func (sc *http2serverConn) runHandler(rw *http2responseWriter, req *http.Request, handler func(http.ResponseWriter, *http.Request)) {
	didPanic := true
	defer func() {
		rw.rws.stream.cancelCtx()
		if didPanic {
			e := recover()
			sc.writeFrameFromHandler(http2FrameWriteRequest{
				write:  http2handlerPanicRST{rw.rws.stream.id},
				stream: rw.rws.stream,
			})
			// Same as net/http:
			if e != nil && e != http.ErrAbortHandler {
				const size = 64 << 10
				buf := make([]byte, size)
				buf = buf[:runtime.Stack(buf, false)]
				sc.logf("http2: panic serving %v: %v\n%s", sc.conn.RemoteAddr(), e, buf)
			}
			return
		}
		rw.handlerDone()
	}()
	handler(rw, req)
	didPanic = false
}

func http2handleHeaderListTooLong(w http.ResponseWriter, r *http.Request) {
	// 10.5.1 Limits on Header Block Size:
	// .. "A server that receives a larger header block than it is
	// willing to handle can send an HTTP 431 (Request Header Fields Too
	// Large) status code"
	const statusRequestHeaderFieldsTooLarge = 431 // only in Go 1.6+
	w.WriteHeader(statusRequestHeaderFieldsTooLarge)
	io.WriteString(w, "<h1>HTTP Error 431</h1><p>Request Header Field(s) Too Large</p>")
}

// called from handler goroutines.
// h may be nil.
func (sc *http2serverConn) writeHeaders(st *http2stream, headerData *http2writeResHeaders) error {
	sc.serveG.checkNotOn() // NOT on
	var errc chan error
	if headerData.h != nil {
		// If there's a header map (which we don't own), so we have to block on
		// waiting for this frame to be written, so an http.Flush mid-handler
		// writes out the correct value of keys, before a handler later potentially
		// mutates it.
		errc = http2errChanPool.Get().(chan error)
	}
	if err := sc.writeFrameFromHandler(http2FrameWriteRequest{
		write:  headerData,
		stream: st,
		done:   errc,
	}); err != nil {
		return err
	}
	if errc != nil {
		select {
		case err := <-errc:
			http2errChanPool.Put(errc)
			return err
		case <-sc.doneServing:
			return http2errClientDisconnected
		case <-st.cw:
			return http2errStreamClosed
		}
	}
	return nil
}

// called from handler goroutines.
func (sc *http2serverConn) write100ContinueHeaders(st *http2stream) {
	sc.writeFrameFromHandler(http2FrameWriteRequest{
		write:  http2write100ContinueHeadersFrame{st.id},
		stream: st,
	})
}

// A bodyReadMsg tells the server loop that the http.Handler read n
// bytes of the DATA from the client on the given stream.
type http2bodyReadMsg struct {
	st *http2stream
	n  int
}

// called from handler goroutines.
// Notes that the handler for the given stream ID read n bytes of its body
// and schedules flow control tokens to be sent.
func (sc *http2serverConn) noteBodyReadFromHandler(st *http2stream, n int, err error) {
	sc.serveG.checkNotOn() // NOT on
	if n > 0 {
		select {
		case sc.bodyReadCh <- http2bodyReadMsg{st, n}:
		case <-sc.doneServing:
		}
	}
}

func (sc *http2serverConn) noteBodyRead(st *http2stream, n int) {
	sc.serveG.check()
	sc.sendWindowUpdate(nil, n) // conn-level
	if st.state != http2stateHalfClosedRemote && st.state != http2stateClosed {
		// Don't send this WINDOW_UPDATE if the stream is closed
		// remotely.
		sc.sendWindowUpdate(st, n)
	}
}

// st may be nil for conn-level
func (sc *http2serverConn) sendWindowUpdate(st *http2stream, n int) {
	sc.serveG.check()
	// "The legal range for the increment to the flow control
	// window is 1 to 2^31-1 (2,147,483,647) octets."
	// A Go Read call on 64-bit machines could in theory read
	// a larger Read than this. Very unlikely, but we handle it here
	// rather than elsewhere for now.
	const maxUint31 = 1<<31 - 1
	for n >= maxUint31 {
		sc.sendWindowUpdate32(st, maxUint31)
		n -= maxUint31
	}
	sc.sendWindowUpdate32(st, int32(n))
}

// st may be nil for conn-level
func (sc *http2serverConn) sendWindowUpdate32(st *http2stream, n int32) {
	sc.serveG.check()
	if n == 0 {
		return
	}
	if n < 0 {
		panic("negative update")
	}
	var streamID uint32
	if st != nil {
		streamID = st.id
	}
	sc.writeFrame(http2FrameWriteRequest{
		write:  http2writeWindowUpdate{streamID: streamID, n: uint32(n)},
		stream: st,
	})
	var ok bool
	if st == nil {
		ok = sc.inflow.add(n)
	} else {
		ok = st.inflow.add(n)
	}
	if !ok {
		panic("internal error; sent too many window updates without decrements?")
	}
}

// requestBody is the Handler's Request.Body type.
// Read and Close may be called concurrently.
type http2requestBody struct {
	_             http2incomparable
	stream        *http2stream
	conn          *http2serverConn
	closed        bool       // for use by Close only
	sawEOF        bool       // for use by Read only
	pipe          *http2pipe // non-nil if we have a HTTP entity message body
	needsContinue bool       // need to send a 100-continue
}

func (b *http2requestBody) Close() error {
	if b.pipe != nil && !b.closed {
		b.pipe.BreakWithError(http2errClosedBody)
	}
	b.closed = true
	return nil
}

func (b *http2requestBody) Read(p []byte) (n int, err error) {
	if b.needsContinue {
		b.needsContinue = false
		b.conn.write100ContinueHeaders(b.stream)
	}
	if b.pipe == nil || b.sawEOF {
		return 0, io.EOF
	}
	n, err = b.pipe.Read(p)
	if err == io.EOF {
		b.sawEOF = true
	}
	if b.conn == nil && http2inTests {
		return
	}
	b.conn.noteBodyReadFromHandler(b.stream, n, err)
	return
}

// responseWriter is the http.ResponseWriter implementation. It's
// intentionally small (1 pointer wide) to minimize garbage. The
// responseWriterState pointer inside is zeroed at the end of a
// request (in handlerDone) and calls on the responseWriter thereafter
// simply crash (caller's mistake), but the much larger responseWriterState
// and buffers are reused between multiple requests.
type http2responseWriter struct {
	rws *http2responseWriterState
}

// from pkg io
type http2stringWriter interface {
	WriteString(s string) (n int, err error)
}

// Optional http.ResponseWriter interfaces implemented.
var (
	_ http.CloseNotifier = (*http2responseWriter)(nil)
	_ http.Flusher       = (*http2responseWriter)(nil)
	_ http2stringWriter  = (*http2responseWriter)(nil)
)

type http2responseWriterState struct {
	// immutable within a request:
	stream *http2stream
	req    *http.Request
	body   *http2requestBody // to close at end of request, if DATA frames didn't
	conn   *http2serverConn

	// TODO: adjust buffer writing sizes based on server config, frame size updates from peer, etc
	bw *bufio.Writer // writing to a chunkWriter{this *responseWriterState}

	// mutated by http.Handler goroutine:
	handlerHeader http.Header // nil until called
	snapHeader    http.Header // snapshot of handlerHeader at WriteHeader time
	trailers      []string    // set in writeChunk
	status        int         // status code passed to WriteHeader
	wroteHeader   bool        // WriteHeader called (explicitly or implicitly). Not necessarily sent to user yet.
	sentHeader    bool        // have we sent the header frame?
	handlerDone   bool        // handler has finished
	dirty         bool        // a Write failed; don't reuse this responseWriterState

	sentContentLen int64 // non-zero if handler set a Content-Length header
	wroteBytes     int64

	closeNotifierMu sync.Mutex // guards closeNotifierCh
	closeNotifierCh chan bool  // nil until first used
}

type http2chunkWriter struct{ rws *http2responseWriterState }

func (cw http2chunkWriter) Write(p []byte) (n int, err error) { return cw.rws.writeChunk(p) }

func (rws *http2responseWriterState) hasTrailers() bool { return len(rws.trailers) > 0 }

func (rws *http2responseWriterState) hasNonemptyTrailers() bool {
	for _, trailer := range rws.trailers {
		if _, ok := rws.handlerHeader[trailer]; ok {
			return true
		}
	}
	return false
}

// declareTrailer is called for each Trailer header when the
// response header is written. It notes that a header will need to be
// written in the trailers at the end of the response.
func (rws *http2responseWriterState) declareTrailer(k string) {
	k = http.CanonicalHeaderKey(k)
	if !httpguts.ValidTrailerHeader(k) {
		// Forbidden by RFC 7230, section 4.1.2.
		rws.conn.logf("ignoring invalid trailer %q", k)
		return
	}
	if !http2strSliceContains(rws.trailers, k) {
		rws.trailers = append(rws.trailers, k)
	}
}

// writeChunk writes chunks from the bufio.Writer. But because
// bufio.Writer may bypass its chunking, sometimes p may be
// arbitrarily large.
//
// writeChunk is also responsible (on the first chunk) for sending the
// HEADER response.
func (rws *http2responseWriterState) writeChunk(p []byte) (n int, err error) {
	if !rws.wroteHeader {
		rws.writeHeader(200)
	}

	isHeadResp := rws.req.Method == "HEAD"
	if !rws.sentHeader {
		rws.sentHeader = true
		var ctype, clen string
		if clen = rws.snapHeader.Get("Content-Length"); clen != "" {
			rws.snapHeader.Del("Content-Length")
			if cl, err := strconv.ParseUint(clen, 10, 63); err == nil {
				rws.sentContentLen = int64(cl)
			} else {
				clen = ""
			}
		}
		if clen == "" && rws.handlerDone && http2bodyAllowedForStatus(rws.status) && (len(p) > 0 || !isHeadResp) {
			clen = strconv.Itoa(len(p))
		}
		_, hasContentType := rws.snapHeader["Content-Type"]
		// If the Content-Encoding is non-blank, we shouldn't
		// sniff the body. See Issue golang.org/issue/31753.
		ce := rws.snapHeader.Get("Content-Encoding")
		hasCE := len(ce) > 0
		if !hasCE && !hasContentType && http2bodyAllowedForStatus(rws.status) && len(p) > 0 {
			ctype = http.DetectContentType(p)
		}
		var date string
		if _, ok := rws.snapHeader["Date"]; !ok {
			// TODO(bradfitz): be faster here, like net/http? measure.
			date = time.Now().UTC().Format(http.TimeFormat)
		}

		for _, v := range rws.snapHeader["Trailer"] {
			http2foreachHeaderElement(v, rws.declareTrailer)
		}

		// "Connection" headers aren't allowed in HTTP/2 (RFC 7540, 8.1.2.2),
		// but respect "Connection" == "close" to mean sending a GOAWAY and tearing
		// down the TCP connection when idle, like we do for HTTP/1.
		// TODO: remove more Connection-specific header fields here, in addition
		// to "Connection".
		if _, ok := rws.snapHeader["Connection"]; ok {
			v := rws.snapHeader.Get("Connection")
			delete(rws.snapHeader, "Connection")
			if v == "close" {
				rws.conn.startGracefulShutdown()
			}
		}

		endStream := (rws.handlerDone && !rws.hasTrailers() && len(p) == 0) || isHeadResp
		err = rws.conn.writeHeaders(rws.stream, &http2writeResHeaders{
			streamID:      rws.stream.id,
			httpResCode:   rws.status,
			h:             rws.snapHeader,
			endStream:     endStream,
			contentType:   ctype,
			contentLength: clen,
			date:          date,
		})
		if err != nil {
			rws.dirty = true
			return 0, err
		}
		if endStream {
			return 0, nil
		}
	}
	if isHeadResp {
		return len(p), nil
	}
	if len(p) == 0 && !rws.handlerDone {
		return 0, nil
	}

	if rws.handlerDone {
		rws.promoteUndeclaredTrailers()
	}

	// only send trailers if they have actually been defined by the
	// server handler.
	hasNonemptyTrailers := rws.hasNonemptyTrailers()
	endStream := rws.handlerDone && !hasNonemptyTrailers
	if len(p) > 0 || endStream {
		// only send a 0 byte DATA frame if we're ending the stream.
		if err := rws.conn.writeDataFromHandler(rws.stream, p, endStream); err != nil {
			rws.dirty = true
			return 0, err
		}
	}

	if rws.handlerDone && hasNonemptyTrailers {
		err = rws.conn.writeHeaders(rws.stream, &http2writeResHeaders{
			streamID:  rws.stream.id,
			h:         rws.handlerHeader,
			trailers:  rws.trailers,
			endStream: true,
		})
		if err != nil {
			rws.dirty = true
		}
		return len(p), err
	}
	return len(p), nil
}

// TrailerPrefix is a magic prefix for ResponseWriter.Header map keys
// that, if present, signals that the map entry is actually for
// the response trailers, and not the response headers. The prefix
// is stripped after the ServeHTTP call finishes and the values are
// sent in the trailers.
//
// This mechanism is intended only for trailers that are not known
// prior to the headers being written. If the set of trailers is fixed
// or known before the header is written, the normal Go trailers mechanism
// is preferred:
//    https://golang.org/pkg/net/http/#ResponseWriter
//    https://golang.org/pkg/net/http/#example_ResponseWriter_trailers
const http2TrailerPrefix = "Trailer:"

// promoteUndeclaredTrailers permits http.Handlers to set trailers
// after the header has already been flushed. Because the Go
// ResponseWriter interface has no way to set Trailers (only the
// Header), and because we didn't want to expand the ResponseWriter
// interface, and because nobody used trailers, and because RFC 7230
// says you SHOULD (but not must) predeclare any trailers in the
// header, the official ResponseWriter rules said trailers in Go must
// be predeclared, and then we reuse the same ResponseWriter.Header()
// map to mean both Headers and Trailers. When it's time to write the
// Trailers, we pick out the fields of Headers that were declared as
// trailers. That worked for a while, until we found the first major
// user of Trailers in the wild: gRPC (using them only over http2),
// and gRPC libraries permit setting trailers mid-stream without
// predeclaring them. So: change of plans. We still permit the old
// way, but we also permit this hack: if a Header() key begins with
// "Trailer:", the suffix of that key is a Trailer. Because ':' is an
// invalid token byte anyway, there is no ambiguity. (And it's already
// filtered out) It's mildly hacky, but not terrible.
//
// This method runs after the Handler is done and promotes any Header
// fields to be trailers.
func (rws *http2responseWriterState) promoteUndeclaredTrailers() {
	for k, vv := range rws.handlerHeader {
		if !strings.HasPrefix(k, http2TrailerPrefix) {
			continue
		}
		trailerKey := strings.TrimPrefix(k, http2TrailerPrefix)
		rws.declareTrailer(trailerKey)
		rws.handlerHeader[http.CanonicalHeaderKey(trailerKey)] = vv
	}

	if len(rws.trailers) > 1 {
		sorter := http2sorterPool.Get().(*http2sorter)
		sorter.SortStrings(rws.trailers)
		http2sorterPool.Put(sorter)
	}
}

func (w *http2responseWriter) Flush() {
	rws := w.rws
	if rws == nil {
		panic("Header called after Handler finished")
	}
	if rws.bw.Buffered() > 0 {
		if err := rws.bw.Flush(); err != nil {
			// Ignore the error. The frame writer already knows.
			return
		}
	} else {
		// The bufio.Writer won't call chunkWriter.Write
		// (writeChunk with zero bytes, so we have to do it
		// ourselves to force the HTTP response header and/or
		// final DATA frame (with END_STREAM) to be sent.
		rws.writeChunk(nil)
	}
}

func (w *http2responseWriter) CloseNotify() <-chan bool {
	rws := w.rws
	if rws == nil {
		panic("CloseNotify called after Handler finished")
	}
	rws.closeNotifierMu.Lock()
	ch := rws.closeNotifierCh
	if ch == nil {
		ch = make(chan bool, 1)
		rws.closeNotifierCh = ch
		cw := rws.stream.cw
		go func() {
			cw.Wait() // wait for close
			ch <- true
		}()
	}
	rws.closeNotifierMu.Unlock()
	return ch
}

func (w *http2responseWriter) Header() http.Header {
	rws := w.rws
	if rws == nil {
		panic("Header called after Handler finished")
	}
	if rws.handlerHeader == nil {
		rws.handlerHeader = make(http.Header)
	}
	return rws.handlerHeader
}

// checkWriteHeaderCode is a copy of net/http's checkWriteHeaderCode.
func http2checkWriteHeaderCode(code int) {
	// Issue 22880: require valid WriteHeader status codes.
	// For now we only enforce that it's three digits.
	// In the future we might block things over 599 (600 and above aren't defined
	// at http://httpwg.org/specs/rfc7231.html#status.codes)
	// and we might block under 200 (once we have more mature 1xx support).
	// But for now any three digits.
	//
	// We used to send "HTTP/1.1 000 0" on the wire in responses but there's
	// no equivalent bogus thing we can realistically send in HTTP/2,
	// so we'll consistently panic instead and help people find their bugs
	// early. (We can't return an error from WriteHeader even if we wanted to.)
	if code < 100 || code > 999 {
		panic(fmt.Sprintf("invalid WriteHeader code %v", code))
	}
}

func (w *http2responseWriter) WriteHeader(code int) {
	rws := w.rws
	if rws == nil {
		panic("WriteHeader called after Handler finished")
	}
	rws.writeHeader(code)
}

func (rws *http2responseWriterState) writeHeader(code int) {
	if !rws.wroteHeader {
		http2checkWriteHeaderCode(code)
		rws.wroteHeader = true
		rws.status = code
		if len(rws.handlerHeader) > 0 {
			rws.snapHeader = http2cloneHeader(rws.handlerHeader)
		}
	}
}

func http2cloneHeader(h http.Header) http.Header {
	h2 := make(http.Header, len(h))
	for k, vv := range h {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}
	return h2
}

// The Life Of A Write is like this:
//
// * Handler calls w.Write or w.WriteString ->
// * -> rws.bw (*bufio.Writer) ->
// * (Handler might call Flush)
// * -> chunkWriter{rws}
// * -> responseWriterState.writeChunk(p []byte)
// * -> responseWriterState.writeChunk (most of the magic; see comment there)
func (w *http2responseWriter) Write(p []byte) (n int, err error) {
	return w.write(len(p), p, "")
}

func (w *http2responseWriter) WriteString(s string) (n int, err error) {
	return w.write(len(s), nil, s)
}

// either dataB or dataS is non-zero.
func (w *http2responseWriter) write(lenData int, dataB []byte, dataS string) (n int, err error) {
	rws := w.rws
	if rws == nil {
		panic("Write called after Handler finished")
	}
	if !rws.wroteHeader {
		w.WriteHeader(200)
	}
	if !http2bodyAllowedForStatus(rws.status) {
		return 0, http.ErrBodyNotAllowed
	}
	rws.wroteBytes += int64(len(dataB)) + int64(len(dataS)) // only one can be set
	if rws.sentContentLen != 0 && rws.wroteBytes > rws.sentContentLen {
		// TODO: send a RST_STREAM
		return 0, errors.New("http2: handler wrote more than declared Content-Length")
	}

	if dataB != nil {
		return rws.bw.Write(dataB)
	} else {
		return rws.bw.WriteString(dataS)
	}
}

func (w *http2responseWriter) handlerDone() {
	rws := w.rws
	dirty := rws.dirty
	rws.handlerDone = true
	w.Flush()
	w.rws = nil
	if !dirty {
		// Only recycle the pool if all prior Write calls to
		// the serverConn goroutine completed successfully. If
		// they returned earlier due to resets from the peer
		// there might still be write goroutines outstanding
		// from the serverConn referencing the rws memory. See
		// issue 20704.
		http2responseWriterStatePool.Put(rws)
	}
}

// Push errors.
var (
	errRecursivePush    = errors.New("http2: recursive push not allowed")
	errPushLimitReached = errors.New("http2: push would exceed peer's SETTINGS_MAX_CONCURRENT_STREAMS")
)

var _ http.Pusher = (*http2responseWriter)(nil)

func (w *http2responseWriter) Push(target string, opts *http.PushOptions) error {
	st := w.rws.stream
	sc := st.sc
	sc.serveG.checkNotOn()

	// No recursive pushes: "PUSH_PROMISE frames MUST only be sent on a peer-initiated stream."
	// http://tools.ietf.org/html/rfc7540#section-6.6
	if st.isPushed() {
		return errRecursivePush
	}

	if opts == nil {
		opts = new(http.PushOptions)
	}

	// Default options.
	if opts.Method == "" {
		opts.Method = "GET"
	}
	if opts.Header == nil {
		opts.Header = http.Header{}
	}
	wantScheme := "http"
	if w.rws.req.TLS != nil {
		wantScheme = "https"
	}

	// Validate the request.
	u, err := url.Parse(target)
	if err != nil {
		return err
	}
	if u.Scheme == "" {
		if !strings.HasPrefix(target, "/") {
			return fmt.Errorf("target must be an absolute URL or an absolute path: %q", target)
		}
		u.Scheme = wantScheme
		u.Host = w.rws.req.Host
	} else {
		if u.Scheme != wantScheme {
			return fmt.Errorf("cannot push URL with scheme %q from request with scheme %q", u.Scheme, wantScheme)
		}
		if u.Host == "" {
			return errors.New("URL must have a host")
		}
	}
	for k := range opts.Header {
		if strings.HasPrefix(k, ":") {
			return fmt.Errorf("promised request headers cannot include pseudo header %q", k)
		}
		// These headers are meaningful only if the request has a body,
		// but PUSH_PROMISE requests cannot have a body.
		// http://tools.ietf.org/html/rfc7540#section-8.2
		// Also disallow Host, since the promised URL must be absolute.
		if ascii.EqualFold(k, "content-length") ||
			ascii.EqualFold(k, "content-encoding") ||
			ascii.EqualFold(k, "trailer") ||
			ascii.EqualFold(k, "te") ||
			ascii.EqualFold(k, "expect") ||
			ascii.EqualFold(k, "host") {
			return fmt.Errorf("promised request headers cannot include %q", k)
		}
	}
	if err := http2checkValidHTTP2RequestHeaders(opts.Header); err != nil {
		return err
	}

	// The RFC effectively limits promised requests to GET and HEAD:
	// "Promised requests MUST be cacheable [GET, HEAD, or POST], and MUST be safe [GET or HEAD]"
	// http://tools.ietf.org/html/rfc7540#section-8.2
	if opts.Method != "GET" && opts.Method != "HEAD" {
		return fmt.Errorf("method %q must be GET or HEAD", opts.Method)
	}

	msg := &http2startPushRequest{
		parent: st,
		method: opts.Method,
		url:    u,
		header: http2cloneHeader(opts.Header),
		done:   http2errChanPool.Get().(chan error),
	}

	select {
	case <-sc.doneServing:
		return http2errClientDisconnected
	case <-st.cw:
		return http2errStreamClosed
	case sc.serveMsgCh <- msg:
	}

	select {
	case <-sc.doneServing:
		return http2errClientDisconnected
	case <-st.cw:
		return http2errStreamClosed
	case err := <-msg.done:
		http2errChanPool.Put(msg.done)
		return err
	}
}

type http2startPushRequest struct {
	parent *http2stream
	method string
	url    *url.URL
	header http.Header
	done   chan error
}

func (sc *http2serverConn) startPush(msg *http2startPushRequest) {
	sc.serveG.check()

	// http://tools.ietf.org/html/rfc7540#section-6.6.
	// PUSH_PROMISE frames MUST only be sent on a peer-initiated stream that
	// is in either the "open" or "half-closed (remote)" state.
	if msg.parent.state != http2stateOpen && msg.parent.state != http2stateHalfClosedRemote {
		// responseWriter.Push checks that the stream is peer-initiated.
		msg.done <- http2errStreamClosed
		return
	}

	// http://tools.ietf.org/html/rfc7540#section-6.6.
	if !sc.pushEnabled {
		msg.done <- http.ErrNotSupported
		return
	}

	// PUSH_PROMISE frames must be sent in increasing order by stream ID, so
	// we allocate an ID for the promised stream lazily, when the PUSH_PROMISE
	// is written. Once the ID is allocated, we start the request handler.
	allocatePromisedID := func() (uint32, error) {
		sc.serveG.check()

		// Check this again, just in case. Technically, we might have received
		// an updated SETTINGS by the time we got around to writing this frame.
		if !sc.pushEnabled {
			return 0, http.ErrNotSupported
		}
		// http://tools.ietf.org/html/rfc7540#section-6.5.2.
		if sc.curPushedStreams+1 > sc.clientMaxStreams {
			return 0, errPushLimitReached
		}

		// http://tools.ietf.org/html/rfc7540#section-5.1.1.
		// Streams initiated by the server MUST use even-numbered identifiers.
		// A server that is unable to establish a new stream identifier can send a GOAWAY
		// frame so that the client is forced to open a new connection for new streams.
		if sc.maxPushPromiseID+2 >= 1<<31 {
			sc.startGracefulShutdownInternal()
			return 0, errPushLimitReached
		}
		sc.maxPushPromiseID += 2
		promisedID := sc.maxPushPromiseID

		// http://tools.ietf.org/html/rfc7540#section-8.2.
		// Strictly speaking, the new stream should start in "reserved (local)", then
		// transition to "half closed (remote)" after sending the initial HEADERS, but
		// we start in "half closed (remote)" for simplicity.
		// See further comments at the definition of stateHalfClosedRemote.
		promised := sc.newStream(promisedID, msg.parent.id, http2stateHalfClosedRemote)
		rw, req, err := sc.newWriterAndRequestNoBody(promised, http2requestParam{
			method:    msg.method,
			scheme:    msg.url.Scheme,
			authority: msg.url.Host,
			path:      msg.url.RequestURI(),
			header:    http2cloneHeader(msg.header), // clone since handler runs concurrently with writing the PUSH_PROMISE
		})
		if err != nil {
			// Should not happen, since we've already validated msg.url.
			panic(fmt.Sprintf("newWriterAndRequestNoBody(%+v): %v", msg.url, err))
		}

		go sc.runHandler(rw, req, sc.handler.ServeHTTP)
		return promisedID, nil
	}

	sc.writeFrame(http2FrameWriteRequest{
		write: &http2writePushPromise{
			streamID:           msg.parent.id,
			method:             msg.method,
			url:                msg.url,
			h:                  msg.header,
			allocatePromisedID: allocatePromisedID,
		},
		stream: msg.parent,
		done:   msg.done,
	})
}

// From http://httpwg.org/specs/rfc7540.html#rfc.section.8.1.2.2
var http2connHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Connection",
	"Transfer-Encoding",
	"Upgrade",
}

// checkValidHTTP2RequestHeaders checks whether h is a valid HTTP/2 request,
// per RFC 7540 Section 8.1.2.2.
// The returned error is reported to users.
func http2checkValidHTTP2RequestHeaders(h http.Header) error {
	for _, k := range http2connHeaders {
		if _, ok := h[k]; ok {
			return fmt.Errorf("request header %q is not valid in HTTP/2", k)
		}
	}
	te := h["Te"]
	if len(te) > 0 && (len(te) > 1 || (te[0] != "trailers" && te[0] != "")) {
		return errors.New(`request header "TE" may only be "trailers" in HTTP/2`)
	}
	return nil
}

func http2new400Handler(err error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// h1ServerKeepAlivesDisabled reports whether hs has its keep-alives
// disabled. See comments on h1ServerShutdownChan above for why
// the code is written this way.
func http2h1ServerKeepAlivesDisabled(hs *http.Server) bool {
	var x interface{} = hs
	type I interface {
		doKeepAlives() bool
	}
	if hs, ok := x.(I); ok {
		return !hs.doKeepAlives()
	}
	return false
}

func (sc *http2serverConn) countError(name string, err error) error {
	if sc == nil || sc.srv == nil {
		return err
	}
	f := sc.srv.CountError
	if f == nil {
		return err
	}
	var typ string
	var code http2ErrCode
	switch e := err.(type) {
	case http2ConnectionError:
		typ = "conn"
		code = http2ErrCode(e)
	case http2StreamError:
		typ = "stream"
		code = http2ErrCode(e.Code)
	default:
		return err
	}
	codeStr := http2errCodeName[code]
	if codeStr == "" {
		codeStr = strconv.Itoa(int(code))
	}
	f(fmt.Sprintf("%s_%s_%s", typ, codeStr, name))
	return err
}

// writeFramer is implemented by any type that is used to write frames.
type http2writeFramer interface {
	writeFrame(http2writeContext) error

	// staysWithinBuffer reports whether this writer promises that
	// it will only write less than or equal to size bytes, and it
	// won't Flush the write context.
	staysWithinBuffer(size int) bool
}

// writeContext is the interface needed by the various frame writer
// types below. All the writeFrame methods below are scheduled via the
// frame writing scheduler (see writeScheduler in writesched.go).
//
// This interface is implemented by *http2serverConn.
//
// TODO: decide whether to a) use this in the client code (which didn't
// end up using this yet, because it has a simpler design, not
// currently implementing priorities), or b) delete this and
// make the server code a bit more concrete.
type http2writeContext interface {
	Framer() *http2Framer
	Flush() error
	CloseConn() error
	// HeaderEncoder returns an HPACK encoder that writes to the
	// returned buffer.
	HeaderEncoder() (*hpack.Encoder, *bytes.Buffer)
}

// writeEndsStream reports whether w writes a frame that will transition
// the stream to a half-closed local state. This returns false for RST_STREAM,
// which closes the entire stream (not just the local half).
func http2writeEndsStream(w http2writeFramer) bool {
	switch v := w.(type) {
	case *http2writeData:
		return v.endStream
	case *http2writeResHeaders:
		return v.endStream
	case nil:
		// This can only happen if the caller reuses w after it's
		// been intentionally nil'ed out to prevent use. Keep this
		// here to catch future refactoring breaking it.
		panic("writeEndsStream called on nil writeFramer")
	}
	return false
}

type http2flushFrameWriter struct{}

func (http2flushFrameWriter) writeFrame(ctx http2writeContext) error {
	return ctx.Flush()
}

func (http2flushFrameWriter) staysWithinBuffer(max int) bool { return false }

type http2writeSettings []http2Setting

func (s http2writeSettings) staysWithinBuffer(max int) bool {
	const settingSize = 6 // uint16 + uint32
	return http2frameHeaderLen+settingSize*len(s) <= max

}

func (s http2writeSettings) writeFrame(ctx http2writeContext) error {
	return ctx.Framer().WriteSettings([]http2Setting(s)...)
}

type http2writeGoAway struct {
	maxStreamID uint32
	code        http2ErrCode
}

func (p *http2writeGoAway) writeFrame(ctx http2writeContext) error {
	err := ctx.Framer().WriteGoAway(p.maxStreamID, p.code, nil)
	ctx.Flush() // ignore error: we're hanging up on them anyway
	return err
}

func (*http2writeGoAway) staysWithinBuffer(max int) bool { return false } // flushes

type http2writeData struct {
	streamID  uint32
	p         []byte
	endStream bool
}

func (w *http2writeData) String() string {
	return fmt.Sprintf("writeData(stream=%d, p=%d, endStream=%v)", w.streamID, len(w.p), w.endStream)
}

func (w *http2writeData) writeFrame(ctx http2writeContext) error {
	return ctx.Framer().WriteData(w.streamID, w.endStream, w.p)
}

func (w *http2writeData) staysWithinBuffer(max int) bool {
	return http2frameHeaderLen+len(w.p) <= max
}

// handlerPanicRST is the message sent from handler goroutines when
// the handler panics.
type http2handlerPanicRST struct {
	StreamID uint32
}

func (hp http2handlerPanicRST) writeFrame(ctx http2writeContext) error {
	return ctx.Framer().WriteRSTStream(hp.StreamID, http2ErrCodeInternal)
}

func (hp http2handlerPanicRST) staysWithinBuffer(max int) bool { return http2frameHeaderLen+4 <= max }

func (se http2StreamError) writeFrame(ctx http2writeContext) error {
	return ctx.Framer().WriteRSTStream(se.StreamID, se.Code)
}

func (se http2StreamError) staysWithinBuffer(max int) bool { return http2frameHeaderLen+4 <= max }

type http2writePingAck struct{ pf *http2PingFrame }

func (w http2writePingAck) writeFrame(ctx http2writeContext) error {
	return ctx.Framer().WritePing(true, w.pf.Data)
}

func (w http2writePingAck) staysWithinBuffer(max int) bool {
	return http2frameHeaderLen+len(w.pf.Data) <= max
}

type http2writeSettingsAck struct{}

func (http2writeSettingsAck) writeFrame(ctx http2writeContext) error {
	return ctx.Framer().WriteSettingsAck()
}

func (http2writeSettingsAck) staysWithinBuffer(max int) bool { return http2frameHeaderLen <= max }

// splitHeaderBlock splits headerBlock into fragments so that each fragment fits
// in a single frame, then calls fn for each fragment. firstFrag/lastFrag are true
// for the first/last fragment, respectively.
func http2splitHeaderBlock(ctx http2writeContext, headerBlock []byte, fn func(ctx http2writeContext, frag []byte, firstFrag, lastFrag bool) error) error {
	// For now we're lazy and just pick the minimum MAX_FRAME_SIZE
	// that all peers must support (16KB). Later we could care
	// more and send larger frames if the peer advertised it, but
	// there's little point. Most headers are small anyway (so we
	// generally won't have CONTINUATION frames), and extra frames
	// only waste 9 bytes anyway.
	const http2maxFrameSize = 16384

	first := true
	for len(headerBlock) > 0 {
		frag := headerBlock
		if len(frag) > http2maxFrameSize {
			frag = frag[:http2maxFrameSize]
		}
		headerBlock = headerBlock[len(frag):]
		if err := fn(ctx, frag, first, len(headerBlock) == 0); err != nil {
			return err
		}
		first = false
	}
	return nil
}

// writeResHeaders is a request to write a HEADERS and 0+ CONTINUATION frames
// for HTTP response headers or trailers from a server handler.
type http2writeResHeaders struct {
	streamID    uint32
	httpResCode int         // 0 means no ":status" line
	h           http.Header // may be nil
	trailers    []string    // if non-nil, which keys of h to write. nil means all.
	endStream   bool

	date          string
	contentType   string
	contentLength string
}

func http2encKV(enc *hpack.Encoder, k, v string) {
	if http2VerboseLogs {
		log.Printf("http2: server encoding header %q = %q", k, v)
	}
	enc.WriteField(hpack.HeaderField{Name: k, Value: v})
}

func (w *http2writeResHeaders) staysWithinBuffer(max int) bool {
	// TODO: this is a common one. It'd be nice to return true
	// here and get into the fast path if we could be clever and
	// calculate the size fast enough, or at least a conservative
	// upper bound that usually fires. (Maybe if w.h and
	// w.trailers are nil, so we don't need to enumerate it.)
	// Otherwise I'm afraid that just calculating the length to
	// answer this question would be slower than the ~2µs benefit.
	return false
}

func (w *http2writeResHeaders) writeFrame(ctx http2writeContext) error {
	enc, buf := ctx.HeaderEncoder()
	buf.Reset()

	if w.httpResCode != 0 {
		http2encKV(enc, ":status", http2httpCodeString(w.httpResCode))
	}

	http2encodeHeaders(enc, w.h, w.trailers)

	if w.contentType != "" {
		http2encKV(enc, "content-type", w.contentType)
	}
	if w.contentLength != "" {
		http2encKV(enc, "content-length", w.contentLength)
	}
	if w.date != "" {
		http2encKV(enc, "date", w.date)
	}

	headerBlock := buf.Bytes()
	if len(headerBlock) == 0 && w.trailers == nil {
		panic("unexpected empty hpack")
	}

	return http2splitHeaderBlock(ctx, headerBlock, w.writeHeaderBlock)
}

func (w *http2writeResHeaders) writeHeaderBlock(ctx http2writeContext, frag []byte, firstFrag, lastFrag bool) error {
	if firstFrag {
		return ctx.Framer().WriteHeaders(http2HeadersFrameParam{
			StreamID:      w.streamID,
			BlockFragment: frag,
			EndStream:     w.endStream,
			EndHeaders:    lastFrag,
		})
	}
	return ctx.Framer().WriteContinuation(w.streamID, lastFrag, frag)
}

// writePushPromise is a request to write a PUSH_PROMISE and 0+ CONTINUATION frames.
type http2writePushPromise struct {
	streamID uint32   // pusher stream
	method   string   // for :method
	url      *url.URL // for :scheme, :authority, :path
	h        http.Header

	// Creates an ID for a pushed stream. This runs on serveG just before
	// the frame is written. The returned ID is copied to promisedID.
	allocatePromisedID func() (uint32, error)
	promisedID         uint32
}

func (w *http2writePushPromise) staysWithinBuffer(max int) bool {
	// TODO: see writeResHeaders.staysWithinBuffer
	return false
}

func (w *http2writePushPromise) writeFrame(ctx http2writeContext) error {
	enc, buf := ctx.HeaderEncoder()
	buf.Reset()

	http2encKV(enc, ":method", w.method)
	http2encKV(enc, ":scheme", w.url.Scheme)
	http2encKV(enc, ":authority", w.url.Host)
	http2encKV(enc, ":path", w.url.RequestURI())
	http2encodeHeaders(enc, w.h, nil)

	headerBlock := buf.Bytes()
	if len(headerBlock) == 0 {
		panic("unexpected empty hpack")
	}

	return http2splitHeaderBlock(ctx, headerBlock, w.writeHeaderBlock)
}

func (w *http2writePushPromise) writeHeaderBlock(ctx http2writeContext, frag []byte, firstFrag, lastFrag bool) error {
	if firstFrag {
		return ctx.Framer().WritePushPromise(http2PushPromiseParam{
			StreamID:      w.streamID,
			PromiseID:     w.promisedID,
			BlockFragment: frag,
			EndHeaders:    lastFrag,
		})
	} else {
		return ctx.Framer().WriteContinuation(w.streamID, lastFrag, frag)
	}
}

type http2write100ContinueHeadersFrame struct {
	streamID uint32
}

func (w http2write100ContinueHeadersFrame) writeFrame(ctx http2writeContext) error {
	enc, buf := ctx.HeaderEncoder()
	buf.Reset()
	http2encKV(enc, ":status", "100")
	return ctx.Framer().WriteHeaders(http2HeadersFrameParam{
		StreamID:      w.streamID,
		BlockFragment: buf.Bytes(),
		EndStream:     false,
		EndHeaders:    true,
	})
}

func (w http2write100ContinueHeadersFrame) staysWithinBuffer(max int) bool {
	// Sloppy but conservative:
	return 9+2*(len(":status")+len("100")) <= max
}

type http2writeWindowUpdate struct {
	streamID uint32 // or 0 for conn-level
	n        uint32
}

func (wu http2writeWindowUpdate) staysWithinBuffer(max int) bool { return http2frameHeaderLen+4 <= max }

func (wu http2writeWindowUpdate) writeFrame(ctx http2writeContext) error {
	return ctx.Framer().WriteWindowUpdate(wu.streamID, wu.n)
}

// encodeHeaders encodes an http.Header. If keys is not nil, then (k, h[k])
// is encoded only if k is in keys.
func http2encodeHeaders(enc *hpack.Encoder, h http.Header, keys []string) {
	if keys == nil {
		sorter := http2sorterPool.Get().(*http2sorter)
		// Using defer here, since the returned keys from the
		// sorter.Keys method is only valid until the sorter
		// is returned:
		defer http2sorterPool.Put(sorter)
		keys = sorter.Keys(h)
	}
	for _, k := range keys {
		vv := h[k]
		k, ascii := http2lowerHeader(k)
		if !ascii {
			// Skip writing invalid headers. Per RFC 7540, Section 8.1.2, header
			// field names have to be ASCII characters (just as in HTTP/1.x).
			continue
		}
		if !http2validWireHeaderFieldName(k) {
			// Skip it as backup paranoia. Per
			// golang.org/issue/14048, these should
			// already be rejected at a higher level.
			continue
		}
		isTE := k == "transfer-encoding"
		for _, v := range vv {
			if !httpguts.ValidHeaderFieldValue(v) {
				// TODO: return an error? golang.org/issue/14048
				// For now just omit it.
				continue
			}
			// TODO: more of "8.1.2.2 Connection-Specific Header Fields"
			if isTE && v != "trailers" {
				continue
			}
			http2encKV(enc, k, v)
		}
	}
}

// WriteScheduler is the interface implemented by HTTP/2 write schedulers.
// Methods are never called concurrently.
type http2WriteScheduler interface {
	// OpenStream opens a new stream in the write scheduler.
	// It is illegal to call this with streamID=0 or with a streamID that is
	// already open -- the call may panic.
	OpenStream(streamID uint32, options http2OpenStreamOptions)

	// CloseStream closes a stream in the write scheduler. Any frames queued on
	// this stream should be discarded. It is illegal to call this on a stream
	// that is not open -- the call may panic.
	CloseStream(streamID uint32)

	// AdjustStream adjusts the priority of the given stream. This may be called
	// on a stream that has not yet been opened or has been closed. Note that
	// RFC 7540 allows PRIORITY frames to be sent on streams in any state. See:
	// https://tools.ietf.org/html/rfc7540#section-5.1
	AdjustStream(streamID uint32, priority http2PriorityParam)

	// Push queues a frame in the scheduler. In most cases, this will not be
	// called with wr.StreamID()!=0 unless that stream is currently open. The one
	// exception is RST_STREAM frames, which may be sent on idle or closed streams.
	Push(wr http2FrameWriteRequest)

	// Pop dequeues the next frame to write. Returns false if no frames can
	// be written. Frames with a given wr.StreamID() are Pop'd in the same
	// order they are Push'd, except RST_STREAM frames. No frames should be
	// discarded except by CloseStream.
	Pop() (wr http2FrameWriteRequest, ok bool)
}

// OpenStreamOptions specifies extra options for WriteScheduler.OpenStream.
type http2OpenStreamOptions struct {
	// PusherID is zero if the stream was initiated by the client. Otherwise,
	// PusherID names the stream that pushed the newly opened stream.
	PusherID uint32
}

// FrameWriteRequest is a request to write a frame.
type http2FrameWriteRequest struct {
	// write is the interface value that does the writing, once the
	// WriteScheduler has selected this frame to write. The write
	// functions are all defined in write.go.
	write http2writeFramer

	// stream is the stream on which this frame will be written.
	// nil for non-stream frames like PING and SETTINGS.
	// nil for RST_STREAM streams, which use the StreamError.StreamID field instead.
	stream *http2stream

	// done, if non-nil, must be a buffered channel with space for
	// 1 message and is sent the return value from write (or an
	// earlier error) when the frame has been written.
	done chan error
}

// StreamID returns the id of the stream this frame will be written to.
// 0 is used for non-stream frames such as PING and SETTINGS.
func (wr http2FrameWriteRequest) StreamID() uint32 {
	if wr.stream == nil {
		if se, ok := wr.write.(http2StreamError); ok {
			// (*http2serverConn).resetStream doesn't set
			// stream because it doesn't necessarily have
			// one. So special case this type of write
			// message.
			return se.StreamID
		}
		return 0
	}
	return wr.stream.id
}

// isControl reports whether wr is a control frame for MaxQueuedControlFrames
// purposes. That includes non-stream frames and RST_STREAM frames.
func (wr http2FrameWriteRequest) isControl() bool {
	return wr.stream == nil
}

// DataSize returns the number of flow control bytes that must be consumed
// to write this entire frame. This is 0 for non-DATA frames.
func (wr http2FrameWriteRequest) DataSize() int {
	if wd, ok := wr.write.(*http2writeData); ok {
		return len(wd.p)
	}
	return 0
}

// Consume consumes min(n, available) bytes from this frame, where available
// is the number of flow control bytes available on the stream. Consume returns
// 0, 1, or 2 frames, where the integer return value gives the number of frames
// returned.
//
// If flow control prevents consuming any bytes, this returns (_, _, 0). If
// the entire frame was consumed, this returns (wr, _, 1). Otherwise, this
// returns (consumed, rest, 2), where 'consumed' contains the consumed bytes and
// 'rest' contains the remaining bytes. The consumed bytes are deducted from the
// underlying stream's flow control budget.
func (wr http2FrameWriteRequest) Consume(n int32) (http2FrameWriteRequest, http2FrameWriteRequest, int) {
	var empty http2FrameWriteRequest

	// Non-DATA frames are always consumed whole.
	wd, ok := wr.write.(*http2writeData)
	if !ok || len(wd.p) == 0 {
		return wr, empty, 1
	}

	// Might need to split after applying limits.
	allowed := wr.stream.flow.available()
	if n < allowed {
		allowed = n
	}
	if wr.stream.sc.maxFrameSize < allowed {
		allowed = wr.stream.sc.maxFrameSize
	}
	if allowed <= 0 {
		return empty, empty, 0
	}
	if len(wd.p) > int(allowed) {
		wr.stream.flow.take(allowed)
		consumed := http2FrameWriteRequest{
			stream: wr.stream,
			write: &http2writeData{
				streamID: wd.streamID,
				p:        wd.p[:allowed],
				// Even if the original had endStream set, there
				// are bytes remaining because len(wd.p) > allowed,
				// so we know endStream is false.
				endStream: false,
			},
			// Our caller is blocking on the final DATA frame, not
			// this intermediate frame, so no need to wait.
			done: nil,
		}
		rest := http2FrameWriteRequest{
			stream: wr.stream,
			write: &http2writeData{
				streamID:  wd.streamID,
				p:         wd.p[allowed:],
				endStream: wd.endStream,
			},
			done: wr.done,
		}
		return consumed, rest, 2
	}

	// The frame is consumed whole.
	// NB: This cast cannot overflow because allowed is <= math.MaxInt32.
	wr.stream.flow.take(int32(len(wd.p)))
	return wr, empty, 1
}

// String is for debugging only.
func (wr http2FrameWriteRequest) String() string {
	var des string
	if s, ok := wr.write.(fmt.Stringer); ok {
		des = s.String()
	} else {
		des = fmt.Sprintf("%T", wr.write)
	}
	return fmt.Sprintf("[FrameWriteRequest stream=%d, ch=%v, writer=%v]", wr.StreamID(), wr.done != nil, des)
}

// replyToWriter sends err to wr.done and panics if the send must block
// This does nothing if wr.done is nil.
func (wr *http2FrameWriteRequest) replyToWriter(err error) {
	if wr.done == nil {
		return
	}
	select {
	case wr.done <- err:
	default:
		panic(fmt.Sprintf("unbuffered done channel passed in for type %T", wr.write))
	}
	wr.write = nil // prevent use (assume it's tainted after wr.done send)
}

// writeQueue is used by implementations of WriteScheduler.
type http2writeQueue struct {
	s []http2FrameWriteRequest
}

func (q *http2writeQueue) empty() bool { return len(q.s) == 0 }

func (q *http2writeQueue) push(wr http2FrameWriteRequest) {
	q.s = append(q.s, wr)
}

func (q *http2writeQueue) shift() http2FrameWriteRequest {
	if len(q.s) == 0 {
		panic("invalid use of queue")
	}
	wr := q.s[0]
	// TODO: less copy-happy queue.
	copy(q.s, q.s[1:])
	q.s[len(q.s)-1] = http2FrameWriteRequest{}
	q.s = q.s[:len(q.s)-1]
	return wr
}

// consume consumes up to n bytes from q.s[0]. If the frame is
// entirely consumed, it is removed from the queue. If the frame
// is partially consumed, the frame is kept with the consumed
// bytes removed. Returns true iff any bytes were consumed.
func (q *http2writeQueue) consume(n int32) (http2FrameWriteRequest, bool) {
	if len(q.s) == 0 {
		return http2FrameWriteRequest{}, false
	}
	consumed, rest, numresult := q.s[0].Consume(n)
	switch numresult {
	case 0:
		return http2FrameWriteRequest{}, false
	case 1:
		q.shift()
	case 2:
		q.s[0] = rest
	}
	return consumed, true
}

type http2writeQueuePool []*http2writeQueue

// put inserts an unused writeQueue into the pool.

// put inserts an unused writeQueue into the pool.
func (p *http2writeQueuePool) put(q *http2writeQueue) {
	for i := range q.s {
		q.s[i] = http2FrameWriteRequest{}
	}
	q.s = q.s[:0]
	*p = append(*p, q)
}

// get returns an empty writeQueue.
func (p *http2writeQueuePool) get() *http2writeQueue {
	ln := len(*p)
	if ln == 0 {
		return new(http2writeQueue)
	}
	x := ln - 1
	q := (*p)[x]
	(*p)[x] = nil
	*p = (*p)[:x]
	return q
}

// RFC 7540, Section 5.3.5: the default weight is 16.
const http2priorityDefaultWeight = 15 // 16 = 15 + 1

// PriorityWriteSchedulerConfig configures a priorityWriteScheduler.
type http2PriorityWriteSchedulerConfig struct {
	// MaxClosedNodesInTree controls the maximum number of closed streams to
	// retain in the priority tree. http2Setting this to zero saves a small amount
	// of memory at the cost of performance.
	//
	// See RFC 7540, Section 5.3.4:
	//   "It is possible for a stream to become closed while prioritization
	//   information ... is in transit. ... This potentially creates suboptimal
	//   prioritization, since the stream could be given a priority that is
	//   different from what is intended. To avoid these problems, an endpoint
	//   SHOULD retain stream prioritization state for a period after streams
	//   become closed. The longer state is retained, the lower the chance that
	//   streams are assigned incorrect or default priority values."
	MaxClosedNodesInTree int

	// MaxIdleNodesInTree controls the maximum number of idle streams to
	// retain in the priority tree. http2Setting this to zero saves a small amount
	// of memory at the cost of performance.
	//
	// See RFC 7540, Section 5.3.4:
	//   Similarly, streams that are in the "idle" state can be assigned
	//   priority or become a parent of other streams. This allows for the
	//   creation of a grouping node in the dependency tree, which enables
	//   more flexible expressions of priority. Idle streams begin with a
	//   default priority (Section 5.3.5).
	MaxIdleNodesInTree int

	// ThrottleOutOfOrderWrites enables write throttling to help ensure that
	// data is delivered in priority order. This works around a race where
	// stream B depends on stream A and both streams are about to call Write
	// to queue DATA frames. If B wins the race, a naive scheduler would eagerly
	// write as much data from B as possible, but this is suboptimal because A
	// is a higher-priority stream. With throttling enabled, we write a small
	// amount of data from B to minimize the amount of bandwidth that B can
	// steal from A.
	ThrottleOutOfOrderWrites bool
}

// NewPriorityWriteScheduler constructs a WriteScheduler that schedules
// frames by following HTTP/2 priorities as described in RFC 7540 Section 5.3.
// If cfg is nil, default options are used.
func http2NewPriorityWriteScheduler(cfg *http2PriorityWriteSchedulerConfig) http2WriteScheduler {
	if cfg == nil {
		// For justification of these defaults, see:
		// https://docs.google.com/document/d/1oLhNg1skaWD4_DtaoCxdSRN5erEXrH-KnLrMwEpOtFY
		cfg = &http2PriorityWriteSchedulerConfig{
			MaxClosedNodesInTree:     10,
			MaxIdleNodesInTree:       10,
			ThrottleOutOfOrderWrites: false,
		}
	}

	ws := &http2priorityWriteScheduler{
		nodes:                make(map[uint32]*http2priorityNode),
		maxClosedNodesInTree: cfg.MaxClosedNodesInTree,
		maxIdleNodesInTree:   cfg.MaxIdleNodesInTree,
		enableWriteThrottle:  cfg.ThrottleOutOfOrderWrites,
	}
	ws.nodes[0] = &ws.root
	if cfg.ThrottleOutOfOrderWrites {
		ws.writeThrottleLimit = 1024
	} else {
		ws.writeThrottleLimit = math.MaxInt32
	}
	return ws
}

type http2priorityNodeState int

const (
	http2priorityNodeOpen http2priorityNodeState = iota
	http2priorityNodeClosed
	http2priorityNodeIdle
)

// priorityNode is a node in an HTTP/2 priority tree.
// Each node is associated with a single stream ID.
// See RFC 7540, Section 5.3.
type http2priorityNode struct {
	q            http2writeQueue        // queue of pending frames to write
	id           uint32                 // id of the stream, or 0 for the root of the tree
	weight       uint8                  // the actual weight is weight+1, so the value is in [1,256]
	state        http2priorityNodeState // open | closed | idle
	bytes        int64                  // number of bytes written by this node, or 0 if closed
	subtreeBytes int64                  // sum(node.bytes) of all nodes in this subtree

	// These links form the priority tree.
	parent     *http2priorityNode
	kids       *http2priorityNode // start of the kids list
	prev, next *http2priorityNode // doubly-linked list of siblings
}

func (n *http2priorityNode) setParent(parent *http2priorityNode) {
	if n == parent {
		panic("setParent to self")
	}
	if n.parent == parent {
		return
	}
	// Unlink from current parent.
	if parent := n.parent; parent != nil {
		if n.prev == nil {
			parent.kids = n.next
		} else {
			n.prev.next = n.next
		}
		if n.next != nil {
			n.next.prev = n.prev
		}
	}
	// Link to new parent.
	// If parent=nil, remove n from the tree.
	// Always insert at the head of parent.kids (this is assumed by walkReadyInOrder).
	n.parent = parent
	if parent == nil {
		n.next = nil
		n.prev = nil
	} else {
		n.next = parent.kids
		n.prev = nil
		if n.next != nil {
			n.next.prev = n
		}
		parent.kids = n
	}
}

func (n *http2priorityNode) addBytes(b int64) {
	n.bytes += b
	for ; n != nil; n = n.parent {
		n.subtreeBytes += b
	}
}

// walkReadyInOrder iterates over the tree in priority order, calling f for each node
// with a non-empty write queue. When f returns true, this function returns true and the
// walk halts. tmp is used as scratch space for sorting.
//
// f(n, openParent) takes two arguments: the node to visit, n, and a bool that is true
// if any ancestor p of n is still open (ignoring the root node).
func (n *http2priorityNode) walkReadyInOrder(openParent bool, tmp *[]*http2priorityNode, f func(*http2priorityNode, bool) bool) bool {
	if !n.q.empty() && f(n, openParent) {
		return true
	}
	if n.kids == nil {
		return false
	}

	// Don't consider the root "open" when updating openParent since
	// we can't send data frames on the root stream (only control frames).
	if n.id != 0 {
		openParent = openParent || (n.state == http2priorityNodeOpen)
	}

	// Common case: only one kid or all kids have the same weight.
	// Some clients don't use weights; other clients (like web browsers)
	// use mostly-linear priority trees.
	w := n.kids.weight
	needSort := false
	for k := n.kids.next; k != nil; k = k.next {
		if k.weight != w {
			needSort = true
			break
		}
	}
	if !needSort {
		for k := n.kids; k != nil; k = k.next {
			if k.walkReadyInOrder(openParent, tmp, f) {
				return true
			}
		}
		return false
	}

	// Uncommon case: sort the child nodes. We remove the kids from the parent,
	// then re-insert after sorting so we can reuse tmp for future sort calls.
	*tmp = (*tmp)[:0]
	for n.kids != nil {
		*tmp = append(*tmp, n.kids)
		n.kids.setParent(nil)
	}
	sort.Sort(http2sortPriorityNodeSiblings(*tmp))
	for i := len(*tmp) - 1; i >= 0; i-- {
		(*tmp)[i].setParent(n) // setParent inserts at the head of n.kids
	}
	for k := n.kids; k != nil; k = k.next {
		if k.walkReadyInOrder(openParent, tmp, f) {
			return true
		}
	}
	return false
}

type http2sortPriorityNodeSiblings []*http2priorityNode

func (z http2sortPriorityNodeSiblings) Len() int { return len(z) }

func (z http2sortPriorityNodeSiblings) Swap(i, k int) { z[i], z[k] = z[k], z[i] }

func (z http2sortPriorityNodeSiblings) Less(i, k int) bool {
	// Prefer the subtree that has sent fewer bytes relative to its weight.
	// See sections 5.3.2 and 5.3.4.
	wi, bi := float64(z[i].weight+1), float64(z[i].subtreeBytes)
	wk, bk := float64(z[k].weight+1), float64(z[k].subtreeBytes)
	if bi == 0 && bk == 0 {
		return wi >= wk
	}
	if bk == 0 {
		return false
	}
	return bi/bk <= wi/wk
}

type http2priorityWriteScheduler struct {
	// root is the root of the priority tree, where root.id = 0.
	// The root queues control frames that are not associated with any stream.
	root http2priorityNode

	// nodes maps stream ids to priority tree nodes.
	nodes map[uint32]*http2priorityNode

	// maxID is the maximum stream id in nodes.
	maxID uint32

	// lists of nodes that have been closed or are idle, but are kept in
	// the tree for improved prioritization. When the lengths exceed either
	// maxClosedNodesInTree or maxIdleNodesInTree, old nodes are discarded.
	closedNodes, idleNodes []*http2priorityNode

	// From the config.
	maxClosedNodesInTree int
	maxIdleNodesInTree   int
	writeThrottleLimit   int32
	enableWriteThrottle  bool

	// tmp is scratch space for priorityNode.walkReadyInOrder to reduce allocations.
	tmp []*http2priorityNode

	// pool of empty queues for reuse.
	queuePool http2writeQueuePool
}

func (ws *http2priorityWriteScheduler) OpenStream(streamID uint32, options http2OpenStreamOptions) {
	// The stream may be currently idle but cannot be opened or closed.
	if curr := ws.nodes[streamID]; curr != nil {
		if curr.state != http2priorityNodeIdle {
			panic(fmt.Sprintf("stream %d already opened", streamID))
		}
		curr.state = http2priorityNodeOpen
		return
	}

	// RFC 7540, Section 5.3.5:
	//  "All streams are initially assigned a non-exclusive dependency on stream 0x0.
	//  Pushed streams initially depend on their associated stream. In both cases,
	//  streams are assigned a default weight of 16."
	parent := ws.nodes[options.PusherID]
	if parent == nil {
		parent = &ws.root
	}
	n := &http2priorityNode{
		q:      *ws.queuePool.get(),
		id:     streamID,
		weight: http2priorityDefaultWeight,
		state:  http2priorityNodeOpen,
	}
	n.setParent(parent)
	ws.nodes[streamID] = n
	if streamID > ws.maxID {
		ws.maxID = streamID
	}
}

func (ws *http2priorityWriteScheduler) CloseStream(streamID uint32) {
	if streamID == 0 {
		panic("violation of WriteScheduler interface: cannot close stream 0")
	}
	if ws.nodes[streamID] == nil {
		panic(fmt.Sprintf("violation of WriteScheduler interface: unknown stream %d", streamID))
	}
	if ws.nodes[streamID].state != http2priorityNodeOpen {
		panic(fmt.Sprintf("violation of WriteScheduler interface: stream %d already closed", streamID))
	}

	n := ws.nodes[streamID]
	n.state = http2priorityNodeClosed
	n.addBytes(-n.bytes)

	q := n.q
	ws.queuePool.put(&q)
	n.q.s = nil
	if ws.maxClosedNodesInTree > 0 {
		ws.addClosedOrIdleNode(&ws.closedNodes, ws.maxClosedNodesInTree, n)
	} else {
		ws.removeNode(n)
	}
}

func (ws *http2priorityWriteScheduler) AdjustStream(streamID uint32, priority http2PriorityParam) {
	if streamID == 0 {
		panic("adjustPriority on root")
	}

	// If streamID does not exist, there are two cases:
	// - A closed stream that has been removed (this will have ID <= maxID)
	// - An idle stream that is being used for "grouping" (this will have ID > maxID)
	n := ws.nodes[streamID]
	if n == nil {
		if streamID <= ws.maxID || ws.maxIdleNodesInTree == 0 {
			return
		}
		ws.maxID = streamID
		n = &http2priorityNode{
			q:      *ws.queuePool.get(),
			id:     streamID,
			weight: http2priorityDefaultWeight,
			state:  http2priorityNodeIdle,
		}
		n.setParent(&ws.root)
		ws.nodes[streamID] = n
		ws.addClosedOrIdleNode(&ws.idleNodes, ws.maxIdleNodesInTree, n)
	}

	// Section 5.3.1: A dependency on a stream that is not currently in the tree
	// results in that stream being given a default priority (Section 5.3.5).
	parent := ws.nodes[priority.StreamDep]
	if parent == nil {
		n.setParent(&ws.root)
		n.weight = http2priorityDefaultWeight
		return
	}

	// Ignore if the client tries to make a node its own parent.
	if n == parent {
		return
	}

	// Section 5.3.3:
	//   "If a stream is made dependent on one of its own dependencies, the
	//   formerly dependent stream is first moved to be dependent on the
	//   reprioritized stream's previous parent. The moved dependency retains
	//   its weight."
	//
	// That is: if parent depends on n, move parent to depend on n.parent.
	for x := parent.parent; x != nil; x = x.parent {
		if x == n {
			parent.setParent(n.parent)
			break
		}
	}

	// Section 5.3.3: The exclusive flag causes the stream to become the sole
	// dependency of its parent stream, causing other dependencies to become
	// dependent on the exclusive stream.
	if priority.Exclusive {
		k := parent.kids
		for k != nil {
			next := k.next
			if k != n {
				k.setParent(n)
			}
			k = next
		}
	}

	n.setParent(parent)
	n.weight = priority.Weight
}

func (ws *http2priorityWriteScheduler) Push(wr http2FrameWriteRequest) {
	var n *http2priorityNode
	if id := wr.StreamID(); id == 0 {
		n = &ws.root
	} else {
		n = ws.nodes[id]
		if n == nil {
			// id is an idle or closed stream. wr should not be a HEADERS or
			// DATA frame. However, wr can be a RST_STREAM. In this case, we
			// push wr onto the root, rather than creating a new priorityNode,
			// since RST_STREAM is tiny and the stream's priority is unknown
			// anyway. See issue #17919.
			if wr.DataSize() > 0 {
				panic("add DATA on non-open stream")
			}
			n = &ws.root
		}
	}
	n.q.push(wr)
}

func (ws *http2priorityWriteScheduler) Pop() (wr http2FrameWriteRequest, ok bool) {
	ws.root.walkReadyInOrder(false, &ws.tmp, func(n *http2priorityNode, openParent bool) bool {
		limit := int32(math.MaxInt32)
		if openParent {
			limit = ws.writeThrottleLimit
		}
		wr, ok = n.q.consume(limit)
		if !ok {
			return false
		}
		n.addBytes(int64(wr.DataSize()))
		// If B depends on A and B continuously has data available but A
		// does not, gradually increase the throttling limit to allow B to
		// steal more and more bandwidth from A.
		if openParent {
			ws.writeThrottleLimit += 1024
			if ws.writeThrottleLimit < 0 {
				ws.writeThrottleLimit = math.MaxInt32
			}
		} else if ws.enableWriteThrottle {
			ws.writeThrottleLimit = 1024
		}
		return true
	})
	return wr, ok
}

func (ws *http2priorityWriteScheduler) addClosedOrIdleNode(list *[]*http2priorityNode, maxSize int, n *http2priorityNode) {
	if maxSize == 0 {
		return
	}
	if len(*list) == maxSize {
		// Remove the oldest node, then shift left.
		ws.removeNode((*list)[0])
		x := (*list)[1:]
		copy(*list, x)
		*list = (*list)[:len(x)]
	}
	*list = append(*list, n)
}

func (ws *http2priorityWriteScheduler) removeNode(n *http2priorityNode) {
	for k := n.kids; k != nil; k = k.next {
		k.setParent(n.parent)
	}
	n.setParent(nil)
	delete(ws.nodes, n.id)
}

// NewRandomWriteScheduler constructs a WriteScheduler that ignores HTTP/2
// priorities. Control frames like SETTINGS and PING are written before DATA
// frames, but if no control frames are queued and multiple streams have queued
// HEADERS or DATA frames, Pop selects a ready stream arbitrarily.
func http2NewRandomWriteScheduler() http2WriteScheduler {
	return &http2randomWriteScheduler{sq: make(map[uint32]*http2writeQueue)}
}

type http2randomWriteScheduler struct {
	// zero are frames not associated with a specific stream.
	zero http2writeQueue

	// sq contains the stream-specific queues, keyed by stream ID.
	// When a stream is idle, closed, or emptied, it's deleted
	// from the map.
	sq map[uint32]*http2writeQueue

	// pool of empty queues for reuse.
	queuePool http2writeQueuePool
}

func (ws *http2randomWriteScheduler) OpenStream(streamID uint32, options http2OpenStreamOptions) {
	// no-op: idle streams are not tracked
}

func (ws *http2randomWriteScheduler) CloseStream(streamID uint32) {
	q, ok := ws.sq[streamID]
	if !ok {
		return
	}
	delete(ws.sq, streamID)
	ws.queuePool.put(q)
}

func (ws *http2randomWriteScheduler) AdjustStream(streamID uint32, priority http2PriorityParam) {
	// no-op: priorities are ignored
}

func (ws *http2randomWriteScheduler) Push(wr http2FrameWriteRequest) {
	if wr.isControl() {
		ws.zero.push(wr)
		return
	}
	id := wr.StreamID()
	q, ok := ws.sq[id]
	if !ok {
		q = ws.queuePool.get()
		ws.sq[id] = q
	}
	q.push(wr)
}

func (ws *http2randomWriteScheduler) Pop() (http2FrameWriteRequest, bool) {
	// Control and RST_STREAM frames first.
	if !ws.zero.empty() {
		return ws.zero.shift(), true
	}
	// Iterate over all non-idle streams until finding one that can be consumed.
	for streamID, q := range ws.sq {
		if wr, ok := q.consume(math.MaxInt32); ok {
			if q.empty() {
				delete(ws.sq, streamID)
				ws.queuePool.put(q)
			}
			return wr, true
		}
	}
	return http2FrameWriteRequest{}, false
}

var stderrVerbose = flag.Bool("stderr_verbose", false, "Mirror verbosity to stderr, unbuffered")

func stderrv() io.Writer {
	if *stderrVerbose {
		return os.Stderr
	}

	return ioutil.Discard
}

type safeBuffer struct {
	b bytes.Buffer
	m sync.Mutex
}

func (sb *safeBuffer) Write(d []byte) (int, error) {
	sb.m.Lock()
	defer sb.m.Unlock()
	return sb.b.Write(d)
}

func (sb *safeBuffer) Bytes() []byte {
	sb.m.Lock()
	defer sb.m.Unlock()
	return sb.b.Bytes()
}

func (sb *safeBuffer) Len() int {
	sb.m.Lock()
	defer sb.m.Unlock()
	return sb.b.Len()
}

type serverTester struct {
	cc             net.Conn // client conn
	t              testing.TB
	ts             *httptest.Server
	fr             *http2Framer
	serverLogBuf   safeBuffer // logger for httptest.Server
	logFilter      []string   // substrings to filter out
	scMu           sync.Mutex // guards sc
	sc             *http2serverConn
	hpackDec       *hpack.Decoder
	decodedHeaders [][2]string

	// If http2debug!=2, then we capture Frame debug logs that will be written
	// to t.Log after a test fails. The read and write logs use separate locks
	// and buffers so we don't accidentally introduce synchronization between
	// the read and write goroutines, which may hide data races.
	frameReadLogMu   sync.Mutex
	frameReadLogBuf  bytes.Buffer
	frameWriteLogMu  sync.Mutex
	frameWriteLogBuf bytes.Buffer

	// writing headers:
	headerBuf bytes.Buffer
	hpackEnc  *hpack.Encoder
}

func (st *serverTester) onHeaderField(f hpack.HeaderField) {
	if f.Name == "date" {
		return
	}
	st.decodedHeaders = append(st.decodedHeaders, [2]string{f.Name, f.Value})
}

func (st *serverTester) decodeHeader(headerBlock []byte) (pairs [][2]string) {
	st.decodedHeaders = nil
	if _, err := st.hpackDec.Write(headerBlock); err != nil {
		st.t.Fatalf("hpack decoding error: %v", err)
	}
	if err := st.hpackDec.Close(); err != nil {
		st.t.Fatalf("hpack decoding error: %v", err)
	}
	return st.decodedHeaders
}

func init() {
	http2testHookOnPanicMu = new(sync.Mutex)
	http2goAwayTimeout = 25 * time.Millisecond
}

func resetHooks() {
	http2testHookOnPanicMu.Lock()
	http2testHookOnPanic = nil
	http2testHookOnPanicMu.Unlock()
}

// ConfigureServer adds HTTP/2 support to a net/http Server.
//
// The configuration conf may be nil.
//
// ConfigureServer must be called before s begins serving.
func http2ConfigureServer(s *http.Server, conf *http2Server) error {
	if s == nil {
		panic("nil *http.Server")
	}
	if conf == nil {
		conf = new(http2Server)
	}
	conf.state = &http2serverInternalState{activeConns: make(map[*http2serverConn]struct{})}
	if h1, h2 := s, conf; h2.IdleTimeout == 0 {
		if h1.IdleTimeout != 0 {
			h2.IdleTimeout = h1.IdleTimeout
		} else {
			h2.IdleTimeout = h1.ReadTimeout
		}
	}
	s.RegisterOnShutdown(conf.state.startGracefulShutdown)

	if s.TLSConfig == nil {
		s.TLSConfig = new(tls.Config)
	} else if s.TLSConfig.CipherSuites != nil && s.TLSConfig.MinVersion < tls.VersionTLS13 {
		// If they already provided a TLS 1.0–1.2 CipherSuite list, return an
		// error if it is missing ECDHE_RSA_WITH_AES_128_GCM_SHA256 or
		// ECDHE_ECDSA_WITH_AES_128_GCM_SHA256.
		haveRequired := false
		for _, cs := range s.TLSConfig.CipherSuites {
			switch cs {
			case tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				// Alternative MTI cipher to not discourage ECDSA-only servers.
				// See http://golang.org/cl/30721 for further information.
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:
				haveRequired = true
			}
		}
		if !haveRequired {
			return fmt.Errorf("http2: TLSConfig.CipherSuites is missing an HTTP/2-required AES_128_GCM_SHA256 cipher (need at least one of TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256 or TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256)")
		}
	}

	// Note: not setting MinVersion to tls.VersionTLS12,
	// as we don't want to interfere with HTTP/1.1 traffic
	// on the user's server. We enforce TLS 1.2 later once
	// we accept a connection. Ideally this should be done
	// during next-proto selection, but using TLS <1.2 with
	// HTTP/2 is still the client's bug.

	s.TLSConfig.PreferServerCipherSuites = true

	if !http2strSliceContains(s.TLSConfig.NextProtos, http2NextProtoTLS) {
		s.TLSConfig.NextProtos = append(s.TLSConfig.NextProtos, http2NextProtoTLS)
	}
	if !http2strSliceContains(s.TLSConfig.NextProtos, "http/1.1") {
		s.TLSConfig.NextProtos = append(s.TLSConfig.NextProtos, "http/1.1")
	}

	if s.TLSNextProto == nil {
		s.TLSNextProto = map[string]func(*http.Server, *tls.Conn, http.Handler){}
	}
	protoHandler := func(hs *http.Server, c *tls.Conn, h http.Handler) {
		if http2testHookOnConn != nil {
			http2testHookOnConn()
		}
		// The TLSNextProto interface predates contexts, so
		// the net/http package passes down its per-connection
		// base context via an exported but unadvertised
		// method on the Handler. This is for internal
		// net/http<=>http2 use only.
		var ctx context.Context
		type baseContexter interface {
			BaseContext() context.Context
		}
		if bc, ok := h.(baseContexter); ok {
			ctx = bc.BaseContext()
		}
		conf.ServeConn(c, &http2ServeConnOpts{
			Context:    ctx,
			Handler:    h,
			BaseConfig: hs,
		})
	}
	s.TLSNextProto[http2NextProtoTLS] = protoHandler
	return nil
}

type twriter struct {
	t  testing.TB
	st *serverTester // optional
}

func (w twriter) Write(p []byte) (n int, err error) {
	if w.st != nil {
		ps := string(p)
		for _, phrase := range w.st.logFilter {
			if strings.Contains(ps, phrase) {
				return len(p), nil // no logging
			}
		}
	}
	w.t.Logf("%s", p)
	return len(p), nil
}

type serverTesterOpt string

var optOnlyServer = serverTesterOpt("only_server")
var optQuiet = serverTesterOpt("quiet_logging")
var optFramerReuseFrames = serverTesterOpt("frame_reuse_frames")

func newServerTester(t testing.TB, handler http.HandlerFunc, opts ...interface{}) *serverTester {
	resetHooks()

	ts := httptest.NewUnstartedServer(handler)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{http2NextProtoTLS},
	}

	var onlyServer, quiet, framerReuseFrames bool
	h2server := new(http2Server)
	for _, opt := range opts {
		switch v := opt.(type) {
		case func(*tls.Config):
			v(tlsConfig)
		case func(*httptest.Server):
			v(ts)
		case func(*http2Server):
			v(h2server)
		case serverTesterOpt:
			switch v {
			case optOnlyServer:
				onlyServer = true
			case optQuiet:
				quiet = true
			case optFramerReuseFrames:
				framerReuseFrames = true
			}
		case func(net.Conn, http.ConnState):
			ts.Config.ConnState = v
		default:
			t.Fatalf("unknown newServerTester option type %T", v)
		}
	}

	http2ConfigureServer(ts.Config, h2server)

	st := &serverTester{
		t:  t,
		ts: ts,
	}
	st.hpackEnc = hpack.NewEncoder(&st.headerBuf)
	st.hpackDec = hpack.NewDecoder(http2initialHeaderTableSize, st.onHeaderField)

	ts.TLS = ts.Config.TLSConfig // the httptest.Server has its own copy of this TLS config
	if quiet {
		ts.Config.ErrorLog = log.New(ioutil.Discard, "", 0)
	} else {
		ts.Config.ErrorLog = log.New(io.MultiWriter(stderrv(), twriter{t: t, st: st}, &st.serverLogBuf), "", log.LstdFlags)
	}
	ts.StartTLS()

	if http2VerboseLogs {
		t.Logf("Running test server at: %s", ts.URL)
	}
	http2testHookGetServerConn = func(v *http2serverConn) {
		st.scMu.Lock()
		defer st.scMu.Unlock()
		st.sc = v
	}
	log.SetOutput(io.MultiWriter(stderrv(), twriter{t: t, st: st}))
	if !onlyServer {
		cc, err := tls.Dial("tcp", ts.Listener.Addr().String(), tlsConfig)
		if err != nil {
			t.Fatal(err)
		}
		st.cc = cc
		st.fr = http2NewFramer(cc, cc)
		if framerReuseFrames {
			st.fr.SetReuseFrames()
		}
		if !http2logFrameReads && !http2logFrameWrites {
			st.fr.debugReadLoggerf = func(m string, v ...interface{}) {
				m = time.Now().Format("2006-01-02 15:04:05.999999999 ") + strings.TrimPrefix(m, "http2: ") + "\n"
				st.frameReadLogMu.Lock()
				fmt.Fprintf(&st.frameReadLogBuf, m, v...)
				st.frameReadLogMu.Unlock()
			}
			st.fr.debugWriteLoggerf = func(m string, v ...interface{}) {
				m = time.Now().Format("2006-01-02 15:04:05.999999999 ") + strings.TrimPrefix(m, "http2: ") + "\n"
				st.frameWriteLogMu.Lock()
				fmt.Fprintf(&st.frameWriteLogBuf, m, v...)
				st.frameWriteLogMu.Unlock()
			}
			st.fr.logReads = true
			st.fr.logWrites = true
		}
	}
	return st
}

func (st *serverTester) closeConn() {
	st.scMu.Lock()
	defer st.scMu.Unlock()
	st.sc.conn.Close()
}

func (st *serverTester) addLogFilter(phrase string) {
	st.logFilter = append(st.logFilter, phrase)
}

func (st *serverTester) stream(id uint32) *http2stream {
	ch := make(chan *http2stream, 1)
	st.sc.serveMsgCh <- func(int) {
		ch <- st.sc.streams[id]
	}
	return <-ch
}

func (st *serverTester) http2streamState(id uint32) http2streamState {
	ch := make(chan http2streamState, 1)
	st.sc.serveMsgCh <- func(int) {
		state, _ := st.sc.state(id)
		ch <- state
	}
	return <-ch
}

// loopNum reports how many times this conn's select loop has gone around.
func (st *serverTester) loopNum() int {
	lastc := make(chan int, 1)
	st.sc.serveMsgCh <- func(loopNum int) {
		lastc <- loopNum
	}
	return <-lastc
}

// awaitIdle heuristically awaits for the server conn's select loop to be idle.
// The heuristic is that the server connection's serve loop must schedule
// 50 times in a row without any channel sends or receives occurring.
func (st *serverTester) awaitIdle() {
	remain := 50
	last := st.loopNum()
	for remain > 0 {
		n := st.loopNum()
		if n == last+1 {
			remain--
		} else {
			remain = 50
		}
		last = n
	}
}

func (st *serverTester) Close() {
	if st.t.Failed() {
		st.frameReadLogMu.Lock()
		if st.frameReadLogBuf.Len() > 0 {
			st.t.Logf("Framer read log:\n%s", st.frameReadLogBuf.String())
		}
		st.frameReadLogMu.Unlock()

		st.frameWriteLogMu.Lock()
		if st.frameWriteLogBuf.Len() > 0 {
			st.t.Logf("Framer write log:\n%s", st.frameWriteLogBuf.String())
		}
		st.frameWriteLogMu.Unlock()

		// If we failed already (and are likely in a Fatal,
		// unwindowing), force close the connection, so the
		// httptest.Server doesn't wait forever for the conn
		// to close.
		if st.cc != nil {
			st.cc.Close()
		}
	}
	st.ts.Close()
	if st.cc != nil {
		st.cc.Close()
	}
	log.SetOutput(os.Stderr)
}

// greet initiates the client's HTTP/2 connection into a state where
// frames may be sent.
func (st *serverTester) greet() {
	st.greetAndCheckSettings(func(http2Setting) error { return nil })
}

func (st *serverTester) greetAndCheckSettings(checkSetting func(s http2Setting) error) {
	st.writePreface()
	st.writeInitialSettings()
	st.wantSettings().ForeachSetting(checkSetting)
	st.writeSettingsAck()

	// The initial WINDOW_UPDATE and SETTINGS ACK can come in any order.
	var gotSettingsAck bool
	var gotWindowUpdate bool

	for i := 0; i < 2; i++ {
		f, err := st.readFrame()
		if err != nil {
			st.t.Fatal(err)
		}
		switch f := f.(type) {
		case *http2SettingsFrame:
			if !f.Header().Flags.Has(http2FlagSettingsAck) {
				st.t.Fatal("Settings Frame didn't have ACK set")
			}
			gotSettingsAck = true

		case *http2WindowUpdateFrame:
			if f.http2FrameHeader.StreamID != 0 {
				st.t.Fatalf("WindowUpdate StreamID = %d; want 0", f.http2FrameHeader.StreamID)
			}
			incr := uint32((&http2Server{}).initialConnRecvWindowSize() - http2initialWindowSize)
			if f.Increment != incr {
				st.t.Fatalf("WindowUpdate increment = %d; want %d", f.Increment, incr)
			}
			gotWindowUpdate = true

		default:
			st.t.Fatalf("Wanting a settings ACK or window update, received a %T", f)
		}
	}

	if !gotSettingsAck {
		st.t.Fatalf("Didn't get a settings ACK")
	}
	if !gotWindowUpdate {
		st.t.Fatalf("Didn't get a window update")
	}
}

func (st *serverTester) writePreface() {
	n, err := st.cc.Write(http2clientPreface)
	if err != nil {
		st.t.Fatalf("Error writing client preface: %v", err)
	}
	if n != len(http2clientPreface) {
		st.t.Fatalf("Writing client preface, wrote %d bytes; want %d", n, len(http2clientPreface))
	}
}

func (st *serverTester) writeInitialSettings() {
	if err := st.fr.WriteSettings(); err != nil {
		st.t.Fatalf("Error writing initial SETTINGS frame from client to server: %v", err)
	}
}

func (st *serverTester) writeSettingsAck() {
	if err := st.fr.WriteSettingsAck(); err != nil {
		st.t.Fatalf("Error writing ACK of server's SETTINGS: %v", err)
	}
}

func (st *serverTester) writeHeaders(p http2HeadersFrameParam) {
	if err := st.fr.WriteHeaders(p); err != nil {
		st.t.Fatalf("Error writing HEADERS: %v", err)
	}
}

func (st *serverTester) writePriority(id uint32, p http2PriorityParam) {
	if err := st.fr.WritePriority(id, p); err != nil {
		st.t.Fatalf("Error writing PRIORITY: %v", err)
	}
}

func (st *serverTester) encodeHeaderField(k, v string) {
	err := st.hpackEnc.WriteField(hpack.HeaderField{Name: k, Value: v})
	if err != nil {
		st.t.Fatalf("HPACK encoding error for %q/%q: %v", k, v, err)
	}
}

// encodeHeaderRaw is the magic-free version of encodeHeader.
// It takes 0 or more (k, v) pairs and encodes them.
func (st *serverTester) encodeHeaderRaw(headers ...string) []byte {
	if len(headers)%2 == 1 {
		panic("odd number of kv args")
	}
	st.headerBuf.Reset()
	for len(headers) > 0 {
		k, v := headers[0], headers[1]
		st.encodeHeaderField(k, v)
		headers = headers[2:]
	}
	return st.headerBuf.Bytes()
}

// encodeHeader encodes headers and returns their HPACK bytes. headers
// must contain an even number of key/value pairs. There may be
// multiple pairs for keys (e.g. "cookie").  The :method, :path, and
// :scheme headers default to GET, / and https. The :authority header
// defaults to st.ts.Listener.Addr().
func (st *serverTester) encodeHeader(headers ...string) []byte {
	if len(headers)%2 == 1 {
		panic("odd number of kv args")
	}

	st.headerBuf.Reset()
	defaultAuthority := st.ts.Listener.Addr().String()

	if len(headers) == 0 {
		// Fast path, mostly for benchmarks, so test code doesn't pollute
		// profiles when we're looking to improve server allocations.
		st.encodeHeaderField(":method", "GET")
		st.encodeHeaderField(":scheme", "https")
		st.encodeHeaderField(":authority", defaultAuthority)
		st.encodeHeaderField(":path", "/")
		return st.headerBuf.Bytes()
	}

	if len(headers) == 2 && headers[0] == ":method" {
		// Another fast path for benchmarks.
		st.encodeHeaderField(":method", headers[1])
		st.encodeHeaderField(":scheme", "https")
		st.encodeHeaderField(":authority", defaultAuthority)
		st.encodeHeaderField(":path", "/")
		return st.headerBuf.Bytes()
	}

	pseudoCount := map[string]int{}
	keys := []string{":method", ":scheme", ":authority", ":path"}
	vals := map[string][]string{
		":method":    {"GET"},
		":scheme":    {"https"},
		":authority": {defaultAuthority},
		":path":      {"/"},
	}
	for len(headers) > 0 {
		k, v := headers[0], headers[1]
		headers = headers[2:]
		if _, ok := vals[k]; !ok {
			keys = append(keys, k)
		}
		if strings.HasPrefix(k, ":") {
			pseudoCount[k]++
			if pseudoCount[k] == 1 {
				vals[k] = []string{v}
			} else {
				// Allows testing of invalid headers w/ dup pseudo fields.
				vals[k] = append(vals[k], v)
			}
		} else {
			vals[k] = append(vals[k], v)
		}
	}
	for _, k := range keys {
		for _, v := range vals[k] {
			st.encodeHeaderField(k, v)
		}
	}
	return st.headerBuf.Bytes()
}

// bodylessReq1 writes a HEADERS frames with StreamID 1 and EndStream and EndHeaders set.
func (st *serverTester) bodylessReq1(headers ...string) {
	st.writeHeaders(http2HeadersFrameParam{
		StreamID:      1, // clients send odd numbers
		BlockFragment: st.encodeHeader(headers...),
		EndStream:     true,
		EndHeaders:    true,
	})
}

func (st *serverTester) writeData(streamID uint32, endStream bool, data []byte) {
	if err := st.fr.WriteData(streamID, endStream, data); err != nil {
		st.t.Fatalf("Error writing DATA: %v", err)
	}
}

func (st *serverTester) writeDataPadded(streamID uint32, endStream bool, data, pad []byte) {
	if err := st.fr.WriteDataPadded(streamID, endStream, data, pad); err != nil {
		st.t.Fatalf("Error writing DATA: %v", err)
	}
}

func (st *serverTester) readFrame() (http2Frame, error) {
	return st.fr.ReadFrame()
}

func (st *serverTester) wantHeaders() *http2HeadersFrame {
	f, err := st.readFrame()
	if err != nil {
		st.t.Fatalf("Error while expecting a HEADERS frame: %v", err)
	}
	hf, ok := f.(*http2HeadersFrame)
	if !ok {
		st.t.Fatalf("got a %T; want *http2HeadersFrame", f)
	}
	return hf
}

func (st *serverTester) wantContinuation() *http2ContinuationFrame {
	f, err := st.readFrame()
	if err != nil {
		st.t.Fatalf("Error while expecting a CONTINUATION frame: %v", err)
	}
	cf, ok := f.(*http2ContinuationFrame)
	if !ok {
		st.t.Fatalf("got a %T; want *http2ContinuationFrame", f)
	}
	return cf
}

func (st *serverTester) wantData() *http2DataFrame {
	f, err := st.readFrame()
	if err != nil {
		st.t.Fatalf("Error while expecting a DATA frame: %v", err)
	}
	df, ok := f.(*http2DataFrame)
	if !ok {
		st.t.Fatalf("got a %T; want *http2DataFrame", f)
	}
	return df
}

func (st *serverTester) wantSettings() *http2SettingsFrame {
	f, err := st.readFrame()
	if err != nil {
		st.t.Fatalf("Error while expecting a SETTINGS frame: %v", err)
	}
	sf, ok := f.(*http2SettingsFrame)
	if !ok {
		st.t.Fatalf("got a %T; want *http2SettingsFrame", f)
	}
	return sf
}

func (st *serverTester) wantPing() *http2PingFrame {
	f, err := st.readFrame()
	if err != nil {
		st.t.Fatalf("Error while expecting a PING frame: %v", err)
	}
	pf, ok := f.(*http2PingFrame)
	if !ok {
		st.t.Fatalf("got a %T; want *http2PingFrame", f)
	}
	return pf
}

func (st *serverTester) wantGoAway() *http2GoAwayFrame {
	f, err := st.readFrame()
	if err != nil {
		st.t.Fatalf("Error while expecting a GOAWAY frame: %v", err)
	}
	gf, ok := f.(*http2GoAwayFrame)
	if !ok {
		st.t.Fatalf("got a %T; want *http2GoAwayFrame", f)
	}
	return gf
}

func (st *serverTester) wantRSTStream(streamID uint32, errCode http2ErrCode) {
	f, err := st.readFrame()
	if err != nil {
		st.t.Fatalf("Error while expecting an RSTStream frame: %v", err)
	}
	rs, ok := f.(*http2RSTStreamFrame)
	if !ok {
		st.t.Fatalf("got a %T; want *http2RSTStreamFrame", f)
	}
	if rs.http2FrameHeader.StreamID != streamID {
		st.t.Fatalf("RSTStream StreamID = %d; want %d", rs.http2FrameHeader.StreamID, streamID)
	}
	if rs.ErrCode != errCode {
		st.t.Fatalf("RSTStream http2ErrCode = %d (%s); want %d (%s)", rs.ErrCode, rs.ErrCode, errCode, errCode)
	}
}

func (st *serverTester) wantWindowUpdate(streamID, incr uint32) {
	f, err := st.readFrame()
	if err != nil {
		st.t.Fatalf("Error while expecting a WINDOW_UPDATE frame: %v", err)
	}
	wu, ok := f.(*http2WindowUpdateFrame)
	if !ok {
		st.t.Fatalf("got a %T; want *http2WindowUpdateFrame", f)
	}
	if wu.http2FrameHeader.StreamID != streamID {
		st.t.Fatalf("WindowUpdate StreamID = %d; want %d", wu.http2FrameHeader.StreamID, streamID)
	}
	if wu.Increment != incr {
		st.t.Fatalf("WindowUpdate increment = %d; want %d", wu.Increment, incr)
	}
}

func (st *serverTester) wantSettingsAck() {
	f, err := st.readFrame()
	if err != nil {
		st.t.Fatal(err)
	}
	sf, ok := f.(*http2SettingsFrame)
	if !ok {
		st.t.Fatalf("Wanting a settings ACK, received a %T", f)
	}
	if !sf.Header().Flags.Has(http2FlagSettingsAck) {
		st.t.Fatal("Settings Frame didn't have ACK set")
	}
}

func (st *serverTester) wantPushPromise() *http2PushPromiseFrame {
	f, err := st.readFrame()
	if err != nil {
		st.t.Fatal(err)
	}
	ppf, ok := f.(*http2PushPromiseFrame)
	if !ok {
		st.t.Fatalf("Wanted PushPromise, received %T", ppf)
	}
	return ppf
}

type specCoverage struct {
	coverage map[specPart]bool
	d        *xml.Decoder
}

func joinSection(sec []int) string {
	s := fmt.Sprintf("%d", sec[0])
	for _, n := range sec[1:] {
		s = fmt.Sprintf("%s.%d", s, n)
	}
	return s
}

func (sc specCoverage) readSection(sec []int) {
	var (
		buf = new(bytes.Buffer)
		sub = 0
	)
	for {
		tk, err := sc.d.Token()
		if err != nil {
			if err == io.EOF {
				return
			}
			panic(err)
		}
		switch v := tk.(type) {
		case xml.StartElement:
			if skipElement(v) {
				if err := sc.d.Skip(); err != nil {
					panic(err)
				}
				if v.Name.Local == "section" {
					sub++
				}
				break
			}
			switch v.Name.Local {
			case "section":
				sub++
				sc.readSection(append(sec, sub))
			case "xref":
				buf.Write(sc.readXRef(v))
			}
		case xml.CharData:
			if len(sec) == 0 {
				break
			}
			buf.Write(v)
		case xml.EndElement:
			if v.Name.Local == "section" {
				sc.addSentences(joinSection(sec), buf.String())
				return
			}
		}
	}
}

func attrSig(se xml.StartElement) string {
	var names []string
	for _, attr := range se.Attr {
		if attr.Name.Local == "fmt" {
			names = append(names, "fmt-"+attr.Value)
		} else {
			names = append(names, attr.Name.Local)
		}
	}
	sort.Strings(names)
	return strings.Join(names, ",")
}

func attrValue(se xml.StartElement, attr string) string {
	for _, a := range se.Attr {
		if a.Name.Local == attr {
			return a.Value
		}
	}
	panic("unknown attribute " + attr)
}

func (sc specCoverage) readXRef(se xml.StartElement) []byte {
	var b []byte
	for {
		tk, err := sc.d.Token()
		if err != nil {
			panic(err)
		}
		switch v := tk.(type) {
		case xml.CharData:
			if b != nil {
				panic("unexpected CharData")
			}
			b = []byte(string(v))
		case xml.EndElement:
			if v.Name.Local != "xref" {
				panic("expected </xref>")
			}
			if b != nil {
				return b
			}
			sig := attrSig(se)
			switch sig {
			case "target":
				return []byte(fmt.Sprintf("[%s]", attrValue(se, "target")))
			case "fmt-of,rel,target", "fmt-,,rel,target":
				return []byte(fmt.Sprintf("[%s, %s]", attrValue(se, "target"), attrValue(se, "rel")))
			case "fmt-of,sec,target", "fmt-,,sec,target":
				return []byte(fmt.Sprintf("[section %s of %s]", attrValue(se, "sec"), attrValue(se, "target")))
			case "fmt-of,rel,sec,target":
				return []byte(fmt.Sprintf("[section %s of %s, %s]", attrValue(se, "sec"), attrValue(se, "target"), attrValue(se, "rel")))
			default:
				panic(fmt.Sprintf("unknown attribute signature %q in %#v", sig, fmt.Sprintf("%#v", se)))
			}
		default:
			panic(fmt.Sprintf("unexpected tag %q", v))
		}
	}
}

var skipAnchor = map[string]bool{
	"intro":    true,
	"Overview": true,
}

var skipTitle = map[string]bool{
	"Acknowledgements":            true,
	"Change Log":                  true,
	"Document Organization":       true,
	"Conventions and Terminology": true,
}

func skipElement(s xml.StartElement) bool {
	switch s.Name.Local {
	case "artwork":
		return true
	case "section":
		for _, attr := range s.Attr {
			switch attr.Name.Local {
			case "anchor":
				if skipAnchor[attr.Value] || strings.HasPrefix(attr.Value, "changes.since.") {
					return true
				}
			case "title":
				if skipTitle[attr.Value] {
					return true
				}
			}
		}
	}
	return false
}

type specPart struct {
	section  string
	sentence string
}

func (ss specPart) Less(oo specPart) bool {
	atoi := func(s string) int {
		n, err := strconv.Atoi(s)
		if err != nil {
			panic(err)
		}
		return n
	}
	a := strings.Split(ss.section, ".")
	b := strings.Split(oo.section, ".")
	for len(a) > 0 {
		if len(b) == 0 {
			return false
		}
		x, y := atoi(a[0]), atoi(b[0])
		if x == y {
			a, b = a[1:], b[1:]
			continue
		}
		return x < y
	}
	if len(b) > 0 {
		return true
	}
	return false
}

type bySpecSection []specPart

func (a bySpecSection) Len() int           { return len(a) }
func (a bySpecSection) Less(i, j int) bool { return a[i].Less(a[j]) }
func (a bySpecSection) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func readSpecCov(r io.Reader) specCoverage {
	sc := specCoverage{
		coverage: map[specPart]bool{},
		d:        xml.NewDecoder(r)}
	sc.readSection(nil)
	return sc
}

var whitespaceRx = regexp.MustCompile(`\s+`)

func parseSentences(sens string) []string {
	sens = strings.TrimSpace(sens)
	if sens == "" {
		return nil
	}
	ss := strings.Split(whitespaceRx.ReplaceAllString(sens, " "), ". ")
	for i, s := range ss {
		s = strings.TrimSpace(s)
		if !strings.HasSuffix(s, ".") {
			s += "."
		}
		ss[i] = s
	}
	return ss
}

func (sc specCoverage) addSentences(sec string, sentence string) {
	for _, s := range parseSentences(sentence) {
		sc.coverage[specPart{sec, s}] = false
	}
}

func (sc specCoverage) cover(sec string, sentence string) {
	for _, s := range parseSentences(sentence) {
		p := specPart{sec, s}
		if _, ok := sc.coverage[p]; !ok {
			panic(fmt.Sprintf("Not found in spec: %q, %q", sec, s))
		}
		sc.coverage[specPart{sec, s}] = true
	}

}

var coverSpec = flag.Bool("coverspec", false, "Run spec coverage tests")

// The global map of sentence coverage for the http2 spec.
var defaultSpecCoverage specCoverage

var loadSpecOnce sync.Once

func loadSpec() {
	if f, err := os.Open("testdata/draft-ietf-httpbis-http2.xml"); err != nil {
		panic(err)
	} else {
		defaultSpecCoverage = readSpecCov(f)
		f.Close()
	}
}

// covers marks all sentences for section sec in defaultSpecCoverage. Sentences not
// "covered" will be included in report outputted by TestSpecCoverage.
func covers(sec, sentences string) {
	loadSpecOnce.Do(loadSpec)
	defaultSpecCoverage.cover(sec, sentences)
}
