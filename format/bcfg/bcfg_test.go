package bcfg

import (
	"encoding/json"
	"fmt"
	"testing"
)

const src = `# Build Configuration

root = fc992d11d5052d03cd4eacdc92137e98
install = b6fee91f3f2b8a4810315f226b635f26 bd54980b8c5eb3db51870fc4167ae234
install-size = 17513 17031
download = 858daae448b7791f7fa37c3c4ddc5c5a 751d89018ecf6d5c2be676ccd47436d9
download-size = 36169994 32564559
encoding = b0d47080bcaedf7bc42408d4d30ddb03 b55711536e6b88b6a35736c7373b9736
encoding-size = 77133267 77098770
patch = 4dccd04051a1b2e7ed53946dc1fd8f40
patch-size = 802943
patch-config = 20f0593dd1b9fdaeaa7a808f83d48f1d
build-name = WOW-24974patch7.3.0_Retail
build-uid = wow
build-product = WoW
build-playbuild-installer = ngdptool_casc2
build-partial-priority = e1e6b5f0ad34f2b41e31a54d17e12af8:262144 d4e734c9b348968ab4cabfac2bfc2671:262144 65f799910a118ad61088faf0b92d5ba2:262144 730e23767fd7d381c0daa673059eeca7:262144 650277c63d68542d65b71495728dee1f:1048576 11fbfa9bab88eef295c04d91a7b2fb1f:262144 f3e2e50a087f9e94a478ee3b27faecc3:262144 14ad5d0809af4e31bb9a5ffdc889f437:1048576 39f7dd51abccd53697693ecc9fe3d46e:262144 5cc09f2a9a89f9a5341153d105c96edc:262144 4c03523e1e3fd7ccdde16b29bbc22a3e:262144 ea05c97c414b7d5e43f7812183f99ad9:262144 2485e3a1ab35d28f4b770775fbaf8db8:262144 890fe992fd5c0da7f32fea48070ead69:1048576`

func Test_Parse(t *testing.T) {
	header, e, _ := Parse([]byte(src))
	fmt.Println(header)
	dat, _ := json.MarshalIndent(e, "", "  ")
	fmt.Printf("%s\n", dat)
}
