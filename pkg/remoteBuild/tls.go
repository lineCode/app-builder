package remoteBuild

import (
	"crypto/tls"
	"crypto/x509"

	"github.com/develar/app-builder/pkg/util"
)

//noinspection SpellCheckingInspection
const localCert = `-----BEGIN CERTIFICATE-----
MIIBiDCCAS+gAwIBAgIRAPHSzTRLcN2nElhQdaRP47IwCgYIKoZIzj0EAwIwJDEi
MCAGA1UEAxMZZWxlY3Ryb24uYnVpbGQubG9jYWwgcm9vdDAeFw0xNzExMTMxNzI4
NDFaFw0yNzExMTExNzI4NDFaMCQxIjAgBgNVBAMTGWVsZWN0cm9uLmJ1aWxkLmxv
Y2FsIHJvb3QwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAQVyduuCT2acuk2QH06
yal/b6O7eTTpOHk3Ucjc+ZZta2vC2+c1IKcSAwimKbTbK+nRxWWJl9ZYx9RTwbRf
QjD6o0IwQDAOBgNVHQ8BAf8EBAMCAQYwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4E
FgQUlm08vBe4CUNAOTQN5Z1RNTfJjjYwCgYIKoZIzj0EAwIDRwAwRAIgMXlT6YM8
4pQtnhUjijVMz+NlcYafS1CEbNBMaWhP87YCIGXUmu7ON9hRLanXzBNBlrtTQG+i
l/NT6REwZA64/lNy
-----END CERTIFICATE-----
`

//noinspection SpellCheckingInspection
const productionCert = `
-----BEGIN CERTIFICATE-----
MIIBfTCCASOgAwIBAgIRAIdieK1+3C4abgOvQ7pVVqAwCgYIKoZIzj0EAwIwHjEc
MBoGA1UEAxMTZWxlY3Ryb24uYnVpbGQgcm9vdDAeFw0xNzExMTMxNzI4NDFaFw0x
ODExMTMxNzI4NDFaMB4xHDAaBgNVBAMTE2VsZWN0cm9uLmJ1aWxkIHJvb3QwWTAT
BgcqhkjOPQIBBggqhkjOPQMBBwNCAAR+4b6twzizN/z27yvwrCV5kinGUrfo+W7n
L/l28ErscNe1BDSyh/IYrnMWb1rDMSLGhvkgI9Cfex1whNPHR101o0IwQDAOBgNV
HQ8BAf8EBAMCAQYwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQU6Dq8kK7tQlrt
zkIYrYiTZGpHEp0wCgYIKoZIzj0EAwIDSAAwRQIgKSfjAQbYlY/S1wMLUi84r8QN
hhMnUwsOmlDan0xPalICIQDLIAXAIyArVtH38a4aizvhH8YeXrxzpJh3U8RolBZF
SA==
-----END CERTIFICATE-----
`

func getTls() (*tls.Config) {
	caCertPool := x509.NewCertPool()
	pemCerts, serverName := getCaCerts()
	caCertPool.AppendCertsFromPEM(pemCerts)

	return &tls.Config{
		ServerName: serverName,
		RootCAs:    caCertPool,
	}
}

func getCaCerts() ([]byte, string) {
	isUseLocalCert := util.IsEnvTrue("USE_BUILD_SERVICE_LOCAL_CA")
	if isUseLocalCert {
		return []byte(localCert), "electron.build.local"
	} else {
		return []byte(productionCert), "electron.build"
	}
}
