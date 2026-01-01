package bootstrap

import (
    "bytes"
    "crypto/tls"
    "crypto/x509"
    "encoding/json"
    "errors"
    "net/http"
)

type Response struct {
    Data struct {
        Cert string `json:"cert"`
        Key  string `json:"key"`
    } `json:"data"`
    Error interface{} `json:"error"`
}

func ExchangeToken(baseURL, path, token string, caPEM []byte) (string, string, error) {
    if token == "" {
        return "", "", errors.New("missing bootstrap token")
    }
    pool := x509.NewCertPool()
    if len(caPEM) > 0 {
        _ = pool.AppendCertsFromPEM(caPEM)
    }
    cli := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}}}
    body := map[string]string{"token": token}
    b, _ := json.Marshal(body)
    req, err := http.NewRequest("POST", baseURL+"/"+path, bytes.NewReader(b))
    if err != nil {
        return "", "", err
    }
    req.Header.Set("Content-Type", "application/json")
    resp, err := cli.Do(req)
    if err != nil {
        return "", "", err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return "", "", errors.New("bootstrap rejected")
    }
    var r Response
    if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
        return "", "", err
    }
    if r.Error != nil || r.Data.Cert == "" || r.Data.Key == "" {
        return "", "", errors.New("bootstrap failed")
    }
    return r.Data.Cert, r.Data.Key, nil
}

