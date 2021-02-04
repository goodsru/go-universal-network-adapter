package sftp

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/konart/go-universal-network-adapter/models"
	assertLib "github.com/stretchr/testify/require"
)

func Test_SftpDownloader_Unit(t *testing.T) {
	assert := assertLib.New(t)
	sftpDownloader := &SftpDownloader{}

	t.Run("SftpMocked_StatExistingFile_ReturnsExpectedSize", func(t *testing.T) {
		parsedDest, _ := models.ParseDestination(models.NewDestination("ftp://ftp.com/dir123/1.jpg", nil, nil))
		fakeSftp := &fakeSftpClient{}
		fakeSftp.On("Stat", "/dir123/1.jpg").Return(
			makeFakeFileInfo("1.jpg", false, 100, time.Now()), nil)

		info, err := sftpDownloader.stat(fakeSftp, parsedDest)

		assert.Nil(err, fmt.Sprintf("err == %v, expected - nil", err))
		assert.Equal(int64(100), info.Size)
		assert.Equal("1.jpg", info.Name)
		fakeSftp.AssertExpectations(t)
	})

	t.Run("SftpMocked_StatNonExistingFile_ReturnsError", func(t *testing.T) {
		parsedDest, _ := models.ParseDestination(models.NewDestination("ftp://ftp.com/dir123/1.jpg", nil, nil))
		fakeSftp := &fakeSftpClient{}
		fakeSftp.On("Stat", "/dir123/1.jpg").Return(nil, fmt.Errorf("no file"))

		info, err := sftpDownloader.stat(fakeSftp, parsedDest)

		assert.NotNil(err)
		assert.Nil(info)
		assert.Equal("no file", err.Error())

		fakeSftp.AssertExpectations(t)
	})

	t.Run("SftpMocked_BrowseExistingFolder_ReturnsListAndNoError", func(t *testing.T) {
		parsedDest, _ := models.ParseDestination(models.NewDestination("ftp://ftp.com/dir123", nil, nil))
		fakeSftp := &fakeSftpClient{}
		fakeSftp.On("ReadDir", "/dir123").Return([]os.FileInfo{
			makeFakeFileInfo("1.jpg", false, 100, time.Now()),
			makeFakeFileInfo("2.jpg", false, 200, time.Now()),
			makeFakeFileInfo("3.jpg", false, 300, time.Now()),
			makeFakeFileInfo("4.jpg", false, 400, time.Now()),
		}, nil)

		list, err := sftpDownloader.browse(fakeSftp, parsedDest)

		assert.Nil(err, fmt.Sprintf("err == %v, expected - nil", err))
		assert.Len(list, 4, fmt.Sprintf("found %v files, expected 4 files", len(list)))
		assert.Equal(list[0].Name, "1.jpg", fmt.Sprintf("Filename - %v, expected - %v", list[0].Name, "1.jpg"))
		assert.Equal(list[3].Name, "4.jpg", fmt.Sprintf("Filename - %v, expected - %v", list[0].Name, "4.jpg"))

		fakeSftp.AssertExpectations(t)
	})

	t.Run("SftpMocked_BrowseNonExistingFolder_ReturnsError", func(t *testing.T) {
		parsedDest, _ := models.ParseDestination(models.NewDestination("ftp://ftp.com/dir123", nil, nil))
		fakeSftp := &fakeSftpClient{}

		errText := "unknown directory"
		fakeSftp.On("ReadDir", "/dir123").Return(nil, fmt.Errorf(errText))

		list, err := sftpDownloader.browse(fakeSftp, parsedDest)

		assert.NotNil(err)
		assert.Nil(list)
		assert.Equal(errText, err.Error())

		fakeSftp.AssertExpectations(t)
	})

	t.Run("SftpMocked_Download_ReturnsCorrectFile", func(t *testing.T) {
		fileName := "test.txt"
		remoteFile, _ := models.NewRemoteFile(models.NewDestination("sftp://sftp.com:22/files/"+fileName, nil, nil))
		fakeSftp := &fakeSftpClient{}

		data := "hello world"
		fakeSftp.On("Open", "/files/"+fileName).Return(&fakeReadCloser{
			data: data,
		}, nil)

		result, err := sftpDownloader.download(fakeSftp, remoteFile)

		assert.Nil(err, fmt.Sprintf("err == %v, expected - nil", err))
		assert.Equal(fileName, result.Name, fmt.Sprintf("Filename - %v, expected - %v", result.Name, fileName))

		_, err = os.Stat(result.Path)
		assert.Nil(err)

		blobBytes, err := ioutil.ReadAll(result.Blob)
		assert.Equal(data, string(blobBytes))
		fakeSftp.AssertExpectations(t)
	})

	t.Run("SftpMocked_DownloadNonExistingFile_ReturnsError", func(t *testing.T) {
		fileName := "test.txt"
		remoteFile, _ := models.NewRemoteFile(models.NewDestination("sftp://sftp.com:22/files/"+fileName, nil, nil))
		fakeSftp := &fakeSftpClient{}

		errText := "incorrect file"
		fakeSftp.On("Open", "/files/"+fileName).Return(nil, fmt.Errorf(errText))

		result, err := sftpDownloader.download(fakeSftp, remoteFile)

		assert.NotNil(err)
		assert.Contains(err.Error(), errText)
		assert.Nil(result)

		fakeSftp.AssertExpectations(t)
	})
}

func Test_SftpDownloader_Integrational(t *testing.T) {
	dir, _ := os.Getwd()
	dirSplit := strings.Split(dir, ":")
	if len(dirSplit) == 2 {
		dir = dirSplit[1]
	}
	dir = filepath.ToSlash(dir)

	assert := assertLib.New(t)
	sftpDownloader := &SftpDownloader{}

	port := 39069
	login := "user"
	pass := "pass"
	isDebug := false

	t.Run("Sftp_StatFile_ReturnsCorrectSize", func(t *testing.T) {
		listener, _, _ := startTestSftpServer(port, login, pass, "", isDebug)
		defer listener.Close()

		parsedDest, _ := models.ParseDestination(models.NewDestination("sftp://user:pass@localhost:"+fmt.Sprintf("%v", port)+path.Join(dir, "test_files", "file1.txt"), nil, nil))
		info, err := sftpDownloader.Stat(parsedDest)

		assert.Nil(err, fmt.Sprintf("err == %v, expected - nil", err))
		assert.Equal(int64(6), info.Size)
		assert.Equal("file1.txt", info.Name)
	})

	t.Run("Sftp_BrowseCorrectFolder_ReturnsListAndNoError", func(t *testing.T) {
		listener, _, _ := startTestSftpServer(port, login, pass, "", isDebug)
		defer listener.Close()

		parsedDest, _ := models.ParseDestination(models.NewDestination("sftp://user:pass@localhost:"+fmt.Sprintf("%v", port)+path.Join(dir, "test_files"), nil, nil))
		list, err := sftpDownloader.Browse(parsedDest)

		assert.Nil(err, fmt.Sprintf("err == %v, expected - nil", err))
		assert.Equal(len(list), 2, "Files not found")
		fileNames := make([]string, 0)
		for _, name := range list {
			fileNames = append(fileNames, name.Name)
		}

		sort.Strings(fileNames)

		assert.Equal(fileNames[0], "file1.txt")
		assert.Equal(fileNames[1], "file2.txt")
	})

	t.Run("Sftp_BrowseBadCredentials_ReturnsAuthError", func(t *testing.T) {
		listener, _, _ := startTestSftpServer(port, login, pass, "", isDebug)
		defer listener.Close()

		parsedDest, _ := models.ParseDestination(models.NewDestination("sftp://user:badpass@localhost:"+fmt.Sprintf("%v", port)+path.Join(dir, "test_files"), nil, nil))
		list, err := sftpDownloader.Browse(parsedDest)

		assert.NotNil(err)
		assert.Contains(err.Error(), "unable to authenticate")
		assert.Len(list, 0)
	})

	t.Run("Sftp_BrowseOnWrongDir_ReturnsError", func(t *testing.T) {
		listener, _, _ := startTestSftpServer(port, login, pass, "", isDebug)
		defer listener.Close()

		parsedDest, _ := models.ParseDestination(models.NewDestination("sftp://user:pass@localhost:"+fmt.Sprintf("%v", port)+path.Join(dir, "bad_dir"), nil, nil))
		list, err := sftpDownloader.Browse(parsedDest)

		assert.NotNil(err)
		assert.Contains(err.Error(), "not exist")
		assert.Len(list, 0)
	})

	t.Run("Sftp_GetClientTimeout_ReturnsError", func(t *testing.T) {
		listener, _, _ := startTestSftpServer(port, login, pass, "", isDebug)
		defer listener.Close()

		timeout := time.Nanosecond
		parsedDest, _ := models.ParseDestination(models.NewDestination("sftp://user:pass@localhost:"+fmt.Sprintf("%v", port)+path.Join(dir, "test_files"), nil, &timeout))
		client, err := sftpDownloader.getClient(parsedDest)

		assert.Nil(client)
		assert.NotNil(err)
		assert.Contains(err.Error(), "timeout")
	})

	t.Run("Sftp_AuthUsingKey_ReturnsList", func(t *testing.T) {
		// https://8gwifi.org/sshfunctions.jsp
		listener, _, _ := startTestSftpServer(port, "", "",
			`ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCRMxXtA5wwrAmPyD0IuvmnT1M/zaW7wX+d+zdywroW3a996/QYkoBkBoiTcGN1DM4p5Ptuld1uh+MXo0YQ6Cobsid4oi4M2W3pGv+mdC3C0WWryIdDOPaSOQw3sCge9xDY78UkPkjj3gTYSaf8J2t5fVBJQjNkJ+jHghAf8jKDB3bz7NP4hmBUsvKoH3GiBipG1j9oJE8WVexWWXlv3Jgjk+jZm2bcYodHvbk8UL6A1mSpovog7ipYnNn6bpZbIWf7ZwuzYADP8VYMR6IbKtJtuQPFEbrk6Ju4h7ETbquXoUXFS06IfF7XVYofG8gx9MhdvCf/0TSKNsWSND2enYkx user@host`, isDebug)
		defer listener.Close()

		parsedDest, _ := models.ParseDestination(models.NewDestination("sftp://localhost:"+fmt.Sprintf("%v", port)+path.Join(dir, "test_files"),
			&models.Credentials{
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
		list, err := sftpDownloader.Browse(parsedDest)

		assert.Nil(err, fmt.Sprintf("err == %v, expected - nil", err))
		assert.Equal(len(list), 2, "Files not found")
	})

	t.Run("Sftp_AuthUsingKeyWithPassphrase_ReturnsList", func(t *testing.T) {

		listener, _, _ := startTestSftpServer(port, "", "",
			`ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCGxetNcnX+gQ0evHG42Dh1mqWC/FC8cjs45gHr80zrW/x2BL9mp/jDFrPCuA19mCO8hrhUU99XxzgDVGlwCUXbJa6AA4pLvaUJRB3cHxPk4iF8dRXdFNGk5FhAZavib/h8JnQKNwdzpKALWtw+z0ph6t+AyeAias5yRgPh3ziN1T4OH15xm944gOOyt0J9ysuYmO4wtwEG+uEjNMpQVGaa4BLRtcTwJqnZISWC2wdmaAL+jt8gSmoFrUxUbF01Nddr27Hd6XrqBQl6qk6Y1N16C9FUSTC/XEx0t23UXclNX2u/lRZ7BBqOOAYnzX+mMuQ8BnCv6ohvCZkY005dJK+L 
 user@host`, isDebug)
		defer listener.Close()

		parsedDest, _ := models.ParseDestination(models.NewDestination(
			"sftp://localhost:"+fmt.Sprintf("%v", port)+path.Join(dir, "test_files"),
			&models.Credentials{
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
		list, err := sftpDownloader.Browse(parsedDest)

		assert.Nil(err, fmt.Sprintf("err == %v, expected - nil", err))
		assert.Equal(len(list), 2, "Files not found")
	})

	t.Run("Sftp_AuthUsingUnauthorizedKey_ReturnsError", func(t *testing.T) {

		listener, _, _ := startTestSftpServer(port, "", "",
			`ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCK1Di4sVlk5O6Mjpg6YRfTDA7WlyRLJbUvXgB668ttImqKoTlbQUzbgSa9u9/LoEJxFcOelybppD3rUYoUhQp5v6x3ikA1oi0hKWbzL3pbLVKZ08NW1WQ7eb2UJMg2scCde6O8ELvmFodSB5NrhPdqghJDQ7q+gncQkzWs9mCDWPxN133+qpzouqV6Zc5r7uxPBpghMW6+pS+a69/EZFVyj6lCf+/zNzf4o/O4kUGejsc9e3zHomwxJnwCZkZwjFnKScsO9bQY0o4RJ7mWSQdeYfYQhgJMwJfTgmZgCx3lQqT/fY3ANvdKrGaXPcjXNNvogXnurKeg9lVIX1ij0Pan  user@host`, isDebug)
		defer listener.Close()

		parsedDest, _ := models.ParseDestination(models.NewDestination("sftp://localhost:"+fmt.Sprintf("%v", port)+path.Join(dir, "test_files"),
			&models.Credentials{
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
		list, err := sftpDownloader.Browse(parsedDest)

		assert.NotNil(err)
		assert.Contains(err.Error(), "unable to authenticate")
		assert.Len(list, 0)
	})

	t.Run("Sftp_AuthUsingKeyWithWrongPassphrase_ReturnsError", func(t *testing.T) {

		listener, _, _ := startTestSftpServer(port, "", "",
			`ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCGxetNcnX+gQ0evHG42Dh1mqWC/FC8cjs45gHr80zrW/x2BL9mp/jDFrPCuA19mCO8hrhUU99XxzgDVGlwCUXbJa6AA4pLvaUJRB3cHxPk4iF8dRXdFNGk5FhAZavib/h8JnQKNwdzpKALWtw+z0ph6t+AyeAias5yRgPh3ziN1T4OH15xm944gOOyt0J9ysuYmO4wtwEG+uEjNMpQVGaa4BLRtcTwJqnZISWC2wdmaAL+jt8gSmoFrUxUbF01Nddr27Hd6XrqBQl6qk6Y1N16C9FUSTC/XEx0t23UXclNX2u/lRZ7BBqOOAYnzX+mMuQ8BnCv6ohvCZkY005dJK+L 
 user@host`, isDebug)
		defer listener.Close()

		parsedDest, _ := models.ParseDestination(models.NewDestination("sftp://localhost:"+fmt.Sprintf("%v", port)+path.Join(dir, "test_files"),
			&models.Credentials{
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
				RsaPrivateKeyPassphrase: "wrong",
			}, nil))
		list, err := sftpDownloader.Browse(parsedDest)

		assert.NotNil(err)
		assert.Contains(err.Error(), "unable to authenticate")
		assert.Len(list, 0)
	})

	t.Run("Sftp_AuthUsingPasswordWhenKeyNeeded_ReturnsError", func(t *testing.T) {

		listener, _, _ := startTestSftpServer(port, "", "",
			`ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCRMxXtA5wwrAmPyD0IuvmnT1M/zaW7wX+d+zdywroW3a996/QYkoBkBoiTcGN1DM4p5Ptuld1uh+MXo0YQ6Cobsid4oi4M2W3pGv+mdC3C0WWryIdDOPaSOQw3sCge9xDY78UkPkjj3gTYSaf8J2t5fVBJQjNkJ+jHghAf8jKDB3bz7NP4hmBUsvKoH3GiBipG1j9oJE8WVexWWXlv3Jgjk+jZm2bcYodHvbk8UL6A1mSpovog7ipYnNn6bpZbIWf7ZwuzYADP8VYMR6IbKtJtuQPFEbrk6Ju4h7ETbquXoUXFS06IfF7XVYofG8gx9MhdvCf/0TSKNsWSND2enYkx user@host`, isDebug)
		defer listener.Close()

		parsedDest, _ := models.ParseDestination(models.NewDestination("sftp://user:pass@localhost:"+fmt.Sprintf("%v", port)+path.Join(dir, "test_files"),
			&models.Credentials{}, nil))
		list, err := sftpDownloader.Browse(parsedDest)

		assert.NotNil(err)
		assert.Contains(err.Error(), "unable to authenticate")
		assert.Len(list, 0)
	})

	t.Run("Sftp_DownloadExistingFile_ReturnsCorrectRemoteFile", func(t *testing.T) {

		listener, _, _ := startTestSftpServer(port, login, pass, "", isDebug)
		defer listener.Close()

		fileName := "file1.txt"
		remoteFile, _ := models.NewRemoteFile(models.NewDestination("sftp://user:pass@localhost:"+fmt.Sprintf("%v", port)+path.Join(dir, "test_files", fileName), nil, nil))

		result, err := sftpDownloader.Download(remoteFile)
		assert.Nil(err, fmt.Sprintf("err == %v, expected - nil", err))
		assert.Equal(fileName, result.Name, fmt.Sprintf("Filename - %v, expected - %v", result.Name, fileName))
		_, err = os.Stat(result.Path)
		assert.Nil(err)
		blobBytes, err := ioutil.ReadAll(result.Blob)
		assert.Equal("Privet", string(blobBytes))
	})

	t.Run("Sftp_DownloadsWithCredentialsNotInUrl_ReturnsCorrectRemoteFile", func(t *testing.T) {
		listener, _, _ := startTestSftpServer(port, login, pass, "", isDebug)
		defer listener.Close()

		fileName := "file1.txt"
		remoteFile, _ := models.NewRemoteFile(models.NewDestination(
			"sftp://"+fmt.Sprintf("localhost:%v", port)+path.Join(dir, "test_files", fileName),
			&models.Credentials{
				User:     `user`,
				Password: `pass`,
			},
			nil))

		result, err := sftpDownloader.Download(remoteFile)

		assert.Nil(err, fmt.Sprintf("err == %v, expected - nil", err))
		assert.Equal(fileName, result.Name, fmt.Sprintf("Filename - %v, expected - %v", result.Name, fileName))
		_, err = os.Stat(result.Path)
		assert.Nil(err)
		blobBytes, err := ioutil.ReadAll(result.Blob)
		assert.Equal("Privet", string(blobBytes))
	})

	t.Run("Sftp_PrefersCredentialsToCredentialsFromUrl", func(t *testing.T) {
		listener, _, _ := startTestSftpServer(port, login, pass, "", isDebug)
		defer listener.Close()

		fileName := "file1.txt"
		remoteFile, _ := models.NewRemoteFile(models.NewDestination(
			"sftp://"+fmt.Sprintf("bad:bad@localhost:%v", port)+path.Join(dir, "test_files", fileName),
			&models.Credentials{
				User:     `user`,
				Password: `pass`,
			}, nil))

		result, err := sftpDownloader.Download(remoteFile)

		assert.Nil(err, fmt.Sprintf("err == %v, expected - nil", err))
		assert.Equal(fileName, result.Name, fmt.Sprintf("Filename - %v, expected - %v", result.Name, fileName))
		_, err = os.Stat(result.Path)
		assert.Nil(err)
		blobBytes, err := ioutil.ReadAll(result.Blob)
		assert.Equal("Privet", string(blobBytes))
	})

	t.Run("Sftp_DownloadFolder_ReturnsError", func(t *testing.T) {

		listener, _, _ := startTestSftpServer(port, login, pass, "", isDebug)
		defer listener.Close()

		remoteFile, _ := models.NewRemoteFile(models.NewDestination("sftp://user:pass@localhost:"+fmt.Sprintf("%v", port)+path.Join(dir, "test_files"), nil, nil))

		file, err := sftpDownloader.Download(remoteFile)
		assert.NotNil(err)
		assert.Nil(file)
	})

}

func makeFakeFileInfo(name string, isDir bool, size int64, modTime time.Time) os.FileInfo {
	fileInfo := &fakeFileInfo{}

	fileInfo.On("Name").Return(name)
	fileInfo.On("IsDir").Return(isDir)
	fileInfo.On("Size").Return(size)
	fileInfo.On("ModTime").Return(modTime)

	return fileInfo
}
