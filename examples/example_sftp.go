package main

import (
	"fmt"
	"io/ioutil"

	"github.com/konart/go-universal-network-adapter/models"
	"github.com/konart/go-universal-network-adapter/services"
)

func sftpBrowseDirectory() {
	adapter := services.NewUniversalNetworkAdapter()

	destination := models.NewDestination("sftp://host.com:22/folder", &models.Credentials{
		User:     "Admin",
		Password: "Qwerty",
	}, nil)

	files, err := adapter.Browse(destination)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		fmt.Println("%v %v %v", file.Name, file.Path, file.Size)
	}
}

func sftpFileStat() {
	adapter := services.NewUniversalNetworkAdapter()

	destination := models.NewDestination("sftp://host.com:22/folder/file.json", &models.Credentials{
		User:     "Admin",
		Password: "Qwerty",
	}, nil)

	remoteFile, err := adapter.Stat(destination)
	if err != nil {
		panic(err)
	}

	fmt.Println("%v %v %v", remoteFile.Name, remoteFile.Path, remoteFile.Size)
}

func sftpDownloadLoginPass() {
	adapter := services.NewUniversalNetworkAdapter()

	remoteFile, err := models.NewRemoteFile(models.NewDestination("sftp://host.com:22/folder/file.json", &models.Credentials{
		User:     "Admin",
		Password: "Qwerty",
	}, nil))

	if err != nil {
		panic(err)
	}
	content, err := adapter.Download(remoteFile)
	if err != nil {
		panic(err)
	}

	fmt.Println(content.Name)
	fmt.Println(content.Path)
	buf, err := ioutil.ReadAll(content.Blob)
	fmt.Println(string(buf))
}

func sftpDownloadPrivateKey() {
	adapter := services.NewUniversalNetworkAdapter()
	remoteFile, err := models.NewRemoteFile(models.NewDestination("sftp://host.com:22/folder/file.json", &models.Credentials{
		RsaPrivateKey: `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAkTMV7QOcMKwJj8g9CLr5p09TP82lu8F/nfs3csK6Ft2vfev0
GJKAZAaIk3BjdQzOKeT7bpXdbofjF6NGEOgqG7IneKIuDNlt6Rr/pnQtwtFlq8iH
Qzj2kjkMN7AoHvcQ2O/FJD5I494E2Emn/CdreX1QSUIzZCfox4IQH/Iygwd28+zT
+IZgVLLyqB9xogYqRtY/aCRPFlXsVll5b9yYI5Po2Ztm3GKHR725PFC+gNZkqaL6
IO4qWJzZ+m6WWyFn+2cLs2AAz/FWDEeiGyrSbbkDxRG65OibuIexE26rl6FFxUtO
iHxe11WKHxvIMfTIXbwn/9E0ijbFkjQ9np2JMQIDAQABAoIBACyqSlReh+1w/n/8
WMoWujV9cV5vJExdeaRfYu8qd5QRHareqnfkmRor6FdyeyXGiqTYi3+5GiSmKHVd
BHCu+kUKyJmTShcpy2WpqHrLwkfrohr11KYZH8BbTCbVSYI8/jG9kCdWAZYW9eaS
wdoPRkBZDBl9A9ILdM/qcothNaiYaj4ex1BE8H8zOfE06kuS3wWuWD/rOKTKIj1L
2UlTAzY+VKATctcTqjwmbCBuTM9/74QlcuekyYuL5kb4Tp5vfO/hy71XgjCVpAjC
/OvKIb1DzJ4d1A6WaGiQ09bMeoE7j0+sUpwOZpXV1PBa328iZeFWCymSJZgFwae8
B8VAwwECgYEA649O0vuCus0l1NfYqUv4DBzXodceQE/z+jEDtMtc0g2YEKkTkvbJ
BFBXh6SUWGu6maRd56VFCUu7S6l7zmWpxVY7WixMh+tLTdzem1aPoy770PnYVgIq
oE1n+wdbT9an2Xm8RyQc4rN0DRki2JjlYYKQEnc2uN/Yo3NzgoqgEZ8CgYEAncyH
LebqcXbX+MeN82coZV6mdsg+OnUT1mt0XSk4GN0TRlSX12ucvue/FD+hAVX2a+IO
vCu3Ig/cDAt7Ck+zbWvkO0p9IQTamHWel6WiU7j61oc7Me80wBrAq8qmPLJ4zn0w
Ug3OyoMDFBtO4aeqXD5fiCEEwQWm8zf51EFkky8CgYEAo7hxArcIf6kCOFLlJZXF
izWosbWAMxbe7e+PMeN+WghUvo+lKSoZQMROcdpzuJj3kr3o/o2h0/os7UOY5zVF
sABlNCFB994T8sQmfDTHlJWdM/vS3sSWt9/U+2Z7kpwRAVhZAeEZqn8rk8b6ryxl
kUZbuFnfUNNUERa3G+4ZnuUCgYEAi+MuyyB0IVYLq72sN2TxyLuZsp9dmxEDHwhv
Rc8urcv+NFD1ssDxWcOz/s1RfA+qvoTOLz5JwOZyWjMrRj7Vf2EwGOe1+bmF17Yd
e64YM0Q/CkMj1OaLyulseF7T8+b7dYJBfdxDv/9YkVCMIzsxqUiaA+HRhxPtppuu
292EvX0CgYBtBjuCeEbNSMN8b9OHYHBexTqBkx8nDHTNW61R7cBo6gH30jAKmLSm
5uA1MemHJXEFtYm42g9fVai6ydWLSb44WISID1G5oF+/Nh/1j93aGSBBzFXq9ccH
MMpP/Pinco5UsoxgxQYBFD2HgGWt3FEuTG5p1ACmsDvl7MnXOJgmRQ==
-----END RSA PRIVATE KEY-----`,
	}, nil))

	if err != nil {
		panic(err)
	}
	content, err := adapter.Download(remoteFile)
	if err != nil {
		panic(err)
	}

	fmt.Println(content.Name)
	fmt.Println(content.Path)
	buf, err := ioutil.ReadAll(content.Blob)
	fmt.Println(string(buf))
}

func sftpDownloadPrivateKeyPassphrase() {
	adapter := services.NewUniversalNetworkAdapter()
	remoteFile, err := models.NewRemoteFile(models.NewDestination("sftp://host.com:22/folder/file.json", &models.Credentials{
		RsaPrivateKey: `-----BEGIN RSA PRIVATE KEY-----
Proc-Type: 4,ENCRYPTED
DEK-Info: DES-EDE3-CBC,52C55E17E529A9C7

EXL2J93Domm3UDW3043P53AcsG8RPRp/kVUQaeBWP7QJGgiS1qV+cqsBNMk/LsHB
iJxqLD7T7WzePygoBru6q3bpKrQg9x7dHGSqEemLhXJAEjFSjW3mm9BUpb7nHPya
wrYzgZZV1lnkoN3OT0W4Lt8DTDpp8fvzK+4DGKV9C8DhH90/xc3qSlPvSmMd7cr/
eWUq+55Os3dvQaiIgf7/nWw/jvMd2EhSX80SSChNXWdXMe7a5s/osbmW/3Asgg1y
aeicWRlbTTUEskqgr+vMTB04U7+CMb0Tj28aVH37DPeQeA5BF9+2GufxNID1Dgo7
xPjCjmasVANTLU+8Y/+cEGA8Dkm6ToXKYTJ1E134rFqU9rlicOzFlqCAy5neuaUi
Iy3jWuJWl3uVIptFHRwEdKXlQiQDoz7AcoFcZBmawAuW7TAjmTDDB6L3XD8PPbBT
1KKVirebtNuDHK8iZ+CJDI0CT7VRRzRbWkofQPNEDASW7SVvTOjOckVT7RBjUcal
O398dd9phcwWIBIEtYkKqKe1fudf5HK9eWRobgWtpVy7SV/t9EQxiZvsMzbBWu/P
B/Qf+h45EWCNol3q/lcPL64SLfyNAUJvEIU7C/6s6YA2P5nV3l0w8C5QD6PDlRLn
vdbV3BvcVbRjZp8laW5zruL0vxjKyqEjzidXh2NiWFQlB+lUYrxUexFKvAY1MTW5
bfSg4tbDomB5OvnSN8Kl3/sIRTnsFJM/ZENLcdMlubguOhxfFhogOFkWYasvXerx
ba3n5a6Ha4K7bszmsuk2FyzR64uxnJxEiIN7O7qgm4TdQzlSZVQw851NwqD3albz
fJUh+BLWqlNckvsPSktV71Ir8kQV0AfX1P7t4EC+lvSGlWm0hi9ThRXrVes+1L8i
2ewjGvW+6aGnLK7623m71Rw1SouyFhg7v1dG/vO69k42H18FhhyqQ7TT8oxRbHch
9aDQYAk9ArixACRPE7a907b8BPDtHCuPhKDE2Icra6JU48OAYrNIQVxf8T0Eu1td
pdNn2wQVxhVYSoaymFRkWit+B6VEx4N0NzuJPWY1id85a0nurOjlqI3zddC4lmX+
E7oJLCxXP6lQBRd5LqM/+pQ8TsKyYGv3GFtdZzex/JOp2FwvE05GBPiIfO6eYdTn
0Jve4tG8q5OrNn9P9SDo+kvOXFg962/IPLE5UatXFX/mTkXd7Pz7veS4QVDb7Pvv
ZEYhl4LQb27t+6eJdS5XftrK7ezOx3joPqC04Kwhe8wjjlzIUUQI0gaWwmSGBRw+
NDuqu0fZB78Ma1bczm1LsP3otYKz7o9A2UT+Zlkp4NaNxBduUhl40KlCC5Yvttv9
8k4q/b8mCU4q+bVU964kJSCFeJ6vdH552Lnf1OfLPg/vSxdiozJtvwbPD/C0zhko
gxQmsimuxuyLvcdLja2Y6rHejonO61VsOrPIwbCeC4YGkLGKbBToU74uee2l07My
+/ubXC+wK8owHDRcAHxcNMSaGclSEIaiRM8Qr8NSJWaTfzwqiKOKi6RWW+xGIBIm
AGBu2h7VnRlNlUpBP+rMHOTKxeZeKcbx//noY3AV8rA0yBmG75ImxQ==
-----END RSA PRIVATE KEY-----
`,
		RsaPrivateKeyPassphrase: "pass",
	}, nil))

	if err != nil {
		panic(err)
	}
	content, err := adapter.Download(remoteFile)
	if err != nil {
		panic(err)
	}

	fmt.Println(content.Name)
	fmt.Println(content.Path)
	buf, err := ioutil.ReadAll(content.Blob)
	fmt.Println(string(buf))
}

func main() {
	sftpBrowseDirectory()
	sftpFileStat()
	sftpDownloadLoginPass()
	sftpDownloadPrivateKey()
	sftpDownloadPrivateKeyPassphrase()
}
