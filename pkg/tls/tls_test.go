// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tls

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"
)

const (
	testRootCA = `-----BEGIN CERTIFICATE-----
MIIDCTCCAfGgAwIBAgIUUPwrcCIhH+xkH89MmTSGuuz5FsgwDQYJKoZIhvcNAQEL
BQAwFDESMBAGA1UEAwwJbG9jYWxob3N0MB4XDTIwMDcwMzAxMDAyNloXDTMwMDcw
MTAxMDAyNlowFDESMBAGA1UEAwwJbG9jYWxob3N0MIIBIjANBgkqhkiG9w0BAQEF
AAOCAQ8AMIIBCgKCAQEAxyYnyCLOHxFpKML+8aO8QwFPaovxZhjponyb5kE5WpVb
Xa2aOxYcevMAPm1QU5WkcT01HSKnDX4j9ymY/RMBMQIXRDgtAR2mec5ipxLQqzq1
z6+lj9Zt2rZ2VMIgzEXLHYjEwk1lDaCa4dMSVdOywqU+8L+tj/JRcMO4Qmr7uTTC
m9SUWWMsjWB4iGxiqX3PO9EATa5znogOZlREm15wuqsxdqljsxvXoq4G3iQFEq3H
imxrFYrDAWUvJqdUKBNTE9AW7j232EHM+0A0L3BZA69pKkr5AkwMTVBoBN4JDsbB
OrELNgzJlt1pCtIHCk6vIfL5d7ZRIZT8Fj7+k95zUQIDAQABo1MwUTAdBgNVHQ4E
FgQU9k7MjjFCQFoJQe++BDb2LNqAMiMwHwYDVR0jBBgwFoAU9k7MjjFCQFoJQe++
BDb2LNqAMiMwDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAmC3U
Ru+LjCsmMHhbNWviiOWqQ7DSgK/ePONHvDGEY9c8lSGsZS9v2kGOjAG9Fmimwq6u
x1amHCtzYEDLVQzlWxpikfHVncj5k20p8AikXyTlWUghuWD6ewSA8l6AENElFO9r
86p9KVNzIp4QQaeCMAs1coMnGOX8mQ8gbobaKDKKGezNq9/5or65dN22h+HGXa/h
URkqYpx2PY+dI/c8X8gSND7cBIKMGVKmAOf/TsJ1Mtzo97l00QVVCP0nrGG+lkkR
l3sef2sdeGpu0Iauj8ZqmA5EKZtDhkzZ/XjgnekXS2zRCEAimduQXlmO4K7GTgmM
BfS2y9tq7LYA3Xlm8w==
-----END CERTIFICATE-----`
	testServerCert = `-----BEGIN CERTIFICATE-----
MIICrzCCAZcCFHXlJNTtGrR0zQtuQwUAb4VieUy8MA0GCSqGSIb3DQEBCwUAMBQx
EjAQBgNVBAMMCWxvY2FsaG9zdDAeFw0yMDA3MDMwMTAwMjZaFw0zMDA3MDEwMTAw
MjZaMBQxEjAQBgNVBAMMCWxvY2FsaG9zdDCCASIwDQYJKoZIhvcNAQEBBQADggEP
ADCCAQoCggEBAO/i0I2KjJ9Iv47UQrWz5Qg00wzmcX1dXQTtApJPdAqtoBYVQI3y
VHSz+gCIVHIfkbLVTcs20lRb54Vrt/K15rHWI4HcHJMNborGMs8chVk8mmV5ttL7
YHLNrjC1HVK0Ym/Xtv6vGxYs3JCD+EBmDO8SwCqyIp6Xje/KLPNd9Ej93J6kWyjl
GJBF2hFP+GvTYz1mMqGOhj2GOWUuV4HvIEnM3Dfr575CgAXcKpCbNYV5gfQ+rUmd
nR2L1b6DBs6w4pim3XDC+faQnEvUVRYXWbeHGMoFPRNYFatKrcRCZ2nWhMy7LAz2
vjkcO6K7LJ7Kigvp0AG8eWY5xF6LjKMy9GkCAwEAATANBgkqhkiG9w0BAQsFAAOC
AQEAvVwxzJoDluCJwi03VM+sUBvyBRD+izSFHxJreSk2lJUELyyXIxgb9+502xNw
NiPhgnOeeeOntqmd/izA1BXFOUbs3mg36hgTgYrIVWtCewziFzT3DmpFASmO62tB
AizQU+bKcBWFhdzqflbH5OOUy7XQp7gXDvXXjomhtz8PHyiUH+yULLOW2LmtgOy5
xQQXKqFtE4WlSfksfkI9Zr7D6GHXf8/tIecr/1u5A2EE14qd7M9nxKHsJPKSkfis
HGUJBI39Lz0oEwKiR3SQ0M8hM87zQfJMVhZsCCZt7sg5OBmfyyGUJKg47mIEOtaD
4b2wWr7cshNNdT4o20eJ5+GKgg==
-----END CERTIFICATE-----`
	testServerKey = `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDv4tCNioyfSL+O
1EK1s+UINNMM5nF9XV0E7QKST3QKraAWFUCN8lR0s/oAiFRyH5Gy1U3LNtJUW+eF
a7fyteax1iOB3ByTDW6KxjLPHIVZPJplebbS+2Byza4wtR1StGJv17b+rxsWLNyQ
g/hAZgzvEsAqsiKel43vyizzXfRI/dyepFso5RiQRdoRT/hr02M9ZjKhjoY9hjll
LleB7yBJzNw36+e+QoAF3CqQmzWFeYH0Pq1JnZ0di9W+gwbOsOKYpt1wwvn2kJxL
1FUWF1m3hxjKBT0TWBWrSq3EQmdp1oTMuywM9r45HDuiuyyeyooL6dABvHlmOcRe
i4yjMvRpAgMBAAECggEAICJHA57apX+uQWjHouV9ObMxzlmUPFHkYwOIw6anCcMm
Xa7tHdanX5a4V5frj/oQn18Zc65dUHWbNvEuC/I3+/yOdxfQMBathyNzrPDKICER
IaTDS9HmoppQyi+IxQpq4UaJOeak/zx1M1qqR54j/2aInW+NWac9mTCBAvzD+37b
wJTO9phbPsHbtapXNB+Vi1Q15etC3Q69RAXeq7ggzfukqUoARykdh7ibz+oeqY5s
7GLjvUDutNSlNbBHst8DE+gI8AReUR5ijlIa4EmyWFwVf3fz/6ZyGK4JzuAIpz8x
SPPdcNb8hkI1CmjnJG0CI4+uFLZ3eoDKGGefDUJVcQKBgQD/74mPxOhWh5SAeZyJ
1XhnheNbZ+EoZshkKDOz3iG81mFI+GQZ5kqW18GsYm7YgazmKIwiZNOoFJNQ2d6F
lOV+lUGHu58wFFWCZABBB1IyC3K3ZyJyrJOv6fYTxMDlbtZSNqefTVO3AIgQ50KO
Vee5nywpnL1iIJ24YcSRK3rstQKBgQDv8j60UBcumF0QAWCIUpA9cndYSS+aiqxi
s+RAvUtgLoA7L3kPvdu4v9B/RfDq3GaezgGuM/CngRKWYTOzYJ2Uvipgd15zJWzj
PPyp/VDByKT2WeQtPOiuNyKMRBiSpu+AlhnlM3MqSz/Td1MhhsoR0K3a/RTluqbn
/2j0nlDtZQKBgGYZCKdxxb2/GF6oJOpXXlDt+GTea9PSLN7Hqth2JL9QRj711/j9
BYRnTbuXCOEV2gN26XBPCKBklOAMCdkALQzyPdIH5tknQ3VgrzlB0mhkaL8BRZy5
e7ERhDkepFTigl0JsJS8JEk8zQrxNnvRiT9jYBq5jFM477I+TkwbLE6BAoGATQwI
PnYJO5kJKP6INL3uOwzqbZrygjlNKVSCUpd2AKht3JNd2EQqraRKGtQmjAPBn/Se
bYWYHPFBOrBznYHEl+KIUPmDho3Z7Q3ERAFnURJQhhpPPod0X5ysp0WmblDUTwHj
xslCja4kDI8gTn+tmxLbAJqLf0j0F0LYPNQpQFECgYEAyV5NYPVc43KSmfMoD6Hf
7WbOLpTBcXyqIlJbET3Cf8QJI4B3ey75wMHf0EFl+/NtCptz1TLMYDzLEFlPjKPP
oIb/JeQDoU6kZsD0whxKr5aNJxShc9O5mx3wI9qOdQ++MSJ2ZJ//hHJvWrmkeVFP
7SAUSqatd1653lljAPEgKjc=
-----END PRIVATE KEY-----`
	testClientCert = `-----BEGIN CERTIFICATE-----
MIICrzCCAZcCFFQFLBfPrlMbwEqSTHjCCdkCeeS4MA0GCSqGSIb3DQEBCwUAMBQx
EjAQBgNVBAMMCWxvY2FsaG9zdDAeFw0yMDA3MDMwMTAwMjZaFw0zMDA3MDEwMTAw
MjZaMBQxEjAQBgNVBAMMCWxvY2FsaG9zdDCCASIwDQYJKoZIhvcNAQEBBQADggEP
ADCCAQoCggEBAK5HVhAUK1+52OWEUn26IsmOwuNekSkKTkVNYcu7jfpD/iO8Q0ZO
NLSVwVZD+E8plHPVangYSu8B3P6kakYXGxLt68kR1c2tggHSH7jDf1msiTDE9ZkV
rUsZRpOR3x9jnbezmy8cmOkcll2Zamh4zXJzy729ZETVw1ZABuiUPtS9K912tLI7
1xC9JLU8ErHf5OEQqVWee0P1b/vNOuNnw/DScjGe1aevJvolKDH9vRpTI5WkDrpL
Sz+bhRQ9MEOI6/iisCj5iO8wfsrb6L2CjTGwSkx3hqnTsbfShLy/N75YH3DCFPUh
66PYvAljBj7R7w2SVFDFygUUMEjgFh06MTECAwEAATANBgkqhkiG9w0BAQsFAAOC
AQEAt0deTpuQ3jY6xTDl2WcxGWrNwFK1NAL5/7cBSGK+HLLOscZ/Bj/emK+WyH1D
4Z3/2ycZaokIG40fCBv2CAqo8rq+36VoRKlZ+Y3PYBeJx24TfBf8sVjLLkpUNcUN
e2cvNlv1q/2sd1eDnhZ50WVFTtSaM6JAAXrWhNGC4M7Y4vY7/hwS/yfybJPFS0Ub
H2MyJyKLuu12AzSl1rknoEtJasYu49St8PHLimTEIGdBjGl07agH+tWbxQZGiF2G
p0JZ53puEJtGr6WIYhmvlNxBNJHHUgInPgKTkytNpCMGXlNqadRGr4cYOsBGzQv6
TcShQNVextmpXhj++g+moLf5ww==
-----END CERTIFICATE-----`
	testClientKey = `-----BEGIN PRIVATE KEY-----
MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQCuR1YQFCtfudjl
hFJ9uiLJjsLjXpEpCk5FTWHLu436Q/4jvENGTjS0lcFWQ/hPKZRz1Wp4GErvAdz+
pGpGFxsS7evJEdXNrYIB0h+4w39ZrIkwxPWZFa1LGUaTkd8fY523s5svHJjpHJZd
mWpoeM1yc8u9vWRE1cNWQAbolD7UvSvddrSyO9cQvSS1PBKx3+ThEKlVnntD9W/7
zTrjZ8Pw0nIxntWnryb6JSgx/b0aUyOVpA66S0s/m4UUPTBDiOv4orAo+YjvMH7K
2+i9go0xsEpMd4ap07G30oS8vze+WB9wwhT1Ieuj2LwJYwY+0e8NklRQxcoFFDBI
4BYdOjExAgMBAAECggEBAKQDlLYrFSrQr5RI9I1IWKbRyQ3MGNbD0SedjLT8vMBT
ruAYtEN9JFLzQPGbpBt0KTbeGYYObq8CVOX8+9scsakd6jHfrBQV/20RQDNVIQvW
uYIBSpWYde9gSTBmLtLOoEabLCepRSrVcZmC4UaSsd7NpWfaznuNpXkNZrMJmRwB
obgFGtmTsc6d9c/IXOx4r5ZxckcLM3R+1S8qIlcSSg67h+rCz2K7uyE2j12k+bUP
JMRSA6wLwpwRDPYvnddGJZT0LWoxBQQg0wOxj+NjdXFDpIXyvHktMXwIlX2e1dJs
kPcxFvWNUTraX5cb4aLokv3Hq7gjTi6P7ji/Dr4FKrECgYEA4meCr02afmSRm9UM
UjT3aHSUwwTRlxVqng4/fZG+AkBsgw0J35vohEVgc04mA8ssNeMFar9epijdtkWf
kGAzbo20TmbFnwEsl6aha5HjaouuBDTKhHmInMqPbMLUmY5+wqpq3Eg5n0JiF43V
ZVLncOmgMZYwjpIn5Wx9/VmyzG0CgYEAxQ94WDIioSEWMcI444MIMEEp16UqI01V
tJZDmTHqkYy2aaTmiZl8tGRt+NssuDNo4hYf0CnlRDWHPCJnEXeO3XJlIn82q0EV
PmYLA/aRBhjmEAtZ+AHoLVtLsh9YgkOt5jZYjqlx0qyqq0zW/QRlS2FtQBZUNOuk
9cx1yiEY9VUCgYEAk4dht/u4jV3ZKBM22SachRtaiI3OSUXyWJXuN1PN3ce/LdJE
OcptRCu1As3MpbIx19dcA0g6U8nTV1+c70ZQnVyHPoNniQoJ1bQGAYNqDlaAiUsh
IBvVsl2oAfYe9QTOgujrgykpKyblpnWMtV7FY2VJNqaoc263wp7kpj9Z7C0CgYBY
1B5lc0QU6fs9NtW6CMjaJ6Wa5YLXIvctTkbueYfJUGqxuHmLMTr7fgHsma/Q4Ku4
rkxs41XLp21sd/2J7CDkRbq9ECcuj0nqNMlmYfEBlJuwQ82sc/+a4np+so6NOcDb
80d8F3mNg/Py/9Ixf5fBoQtERkxzBn33ptC56q17eQKBgQCrS/sVj5ozYU6WnLum
PtLiKerO9GpelYJGSh5JsohHb86Yjv3ApVWQKf5cOtmonenmqT305qdGltIQs+by
5Eqjsi2bM/BH99QbuOlwM5FpYQ8JGm/xX2B37bApGry1q1/iwE5VTIfhA5UcLKeN
KlwHiQ3kjQ3NidX4vpHxD+7nYQ==
-----END PRIVATE KEY-----`
)

func Test_TLSConfig(t *testing.T) {
	testcases := []struct {
		name         string
		clientConfig func(rootCA, clientCert, clientKey string) *tls.Config
		serverConfig func(rootCA, serverCert, serverKey string) *tls.Config
	}{
		{
			name: "mTLS",
			clientConfig: func(rootCA, clientCert, clientKey string) *tls.Config {
				return PrepareClientConfig(rootCA, clientCert, clientKey, false)
			},
			serverConfig: func(rootCA, serverCert, serverKey string) *tls.Config {
				return PrepareServerConfig([]string{rootCA}, serverCert, serverKey)
			},
		},
		{
			name: "mTLS with reload",
			clientConfig: func(rootCA, clientCert, clientKey string) *tls.Config {
				return PrepareClientConfigWithPeriodicReload(rootCA, clientCert, clientKey, false, 0)
			},
			serverConfig: func(rootCA, serverCert, serverKey string) *tls.Config {
				return PrepareServerConfigWithPeriodicReload([]string{rootCA}, serverCert, serverKey, 0)
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			rootCAFile, err := ioutil.TempFile("", "rootCA.crt")
			if err != nil {
				t.Fatalf("error creating tmp root CA: %v", err)
			}
			defer os.Remove(rootCAFile.Name())
			if _, err := rootCAFile.Write([]byte(testRootCA)); err != nil {
				t.Fatalf("error writing to tmp root CA: %v", err)
			}

			clientCertFile, err := ioutil.TempFile("", "client.crt")
			if err != nil {
				t.Fatalf("error creating tmp client cert: %v", err)
			}
			defer os.Remove(clientCertFile.Name())
			if _, err := clientCertFile.Write([]byte(testClientCert)); err != nil {
				t.Fatalf("error writing tmp client cert: %v", err)
			}

			clientKeyFile, err := ioutil.TempFile("", "client.key")
			if err != nil {
				t.Fatalf("error creating tmp client key: %v", err)
			}
			defer os.Remove(clientKeyFile.Name())
			if _, err := clientKeyFile.Write([]byte(testClientKey)); err != nil {
				t.Fatalf("error writing to tmp client key: %v", err)
			}

			serverCertFile, err := ioutil.TempFile("", "server.crt")
			if err != nil {
				t.Fatalf("error creating tmp server cert: %v", err)
			}
			defer os.Remove(serverCertFile.Name())
			if _, err := serverCertFile.Write([]byte(testServerCert)); err != nil {
				t.Fatalf("error writing to tmp server cert: %v", err)
			}

			serverKeyFile, err := ioutil.TempFile("", "server.key")
			if err != nil {
				t.Fatalf("error creating tmp server key: %v", err)
			}
			defer os.Remove(serverKeyFile.Name())
			if _, err := serverKeyFile.Write([]byte(testServerKey)); err != nil {
				t.Fatalf("error writing to tmp server key: %v", err)
			}

			clientConfig := testcase.clientConfig(rootCAFile.Name(), clientCertFile.Name(), clientKeyFile.Name())
			serverConfig := testcase.serverConfig(rootCAFile.Name(), serverCertFile.Name(), serverKeyFile.Name())
			testClientServerHello(t, clientConfig, serverConfig)
		})
	}
}

func testClientServerHello(t *testing.T, clientConfig *tls.Config, serverConfig *tls.Config) {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello")
	})
	server := &http.Server{
		Addr:      ":8443",
		Handler:   mux,
		TLSConfig: serverConfig,
	}

	go func() {
		_ = server.ListenAndServeTLS("", "")
	}()
	defer func() {
		_ = server.Shutdown(context.TODO())
	}()

	// short delay to ensure server started
	time.Sleep(100 * time.Millisecond)

	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: clientConfig,
	}}
	_, err := client.Get("https://localhost:8443/hello")
	if err != nil {
		t.Errorf("client could not connect to server: %v", err)
	}
}

func testClientServerHelloError(t *testing.T, clientConfig *tls.Config, serverConfig *tls.Config) {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello")
	})
	server := &http.Server{
		Addr:      ":8443",
		Handler:   mux,
		TLSConfig: serverConfig,
	}

	go func() {
		_ = server.ListenAndServeTLS("", "")
	}()
	defer func() {
		_ = server.Shutdown(context.TODO())
	}()

	// short delay to ensure server started
	time.Sleep(100 * time.Millisecond)

	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: clientConfig,
	}}
	_, err := client.Get("https://localhost:8443/hello")
	if err == nil {
		t.Error("client connection to server should have errored but got nil")
	}
}
