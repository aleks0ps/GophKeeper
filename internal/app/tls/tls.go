// Package tls -- пакет используется для создания пары tls ключей для HTTPS сервера
package tls

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

// TLS -- содержит CSR и сгенерированые TLS CERT и PRIVATE KEY
type TLS struct {
	CSR     *x509.Certificate
	CertPEM bytes.Buffer
	KeyPEM  bytes.Buffer
}

// New -- генериует TLS сертификат и приватный ключ
func (t *TLS) New() {
	t.CSR = &x509.Certificate{
		// указываем уникальный номер сертификата
		SerialNumber: big.NewInt(12345),
		// заполняем базовую информацию о владельце сертификата
		Subject: pkix.Name{
			Organization: []string{"Yandex.Praktikum"},
			Country:      []string{"RU"},
		},
		IPAddresses: []net.IP{net.IPv4zero, net.IPv6zero},
		// сертификат верен, начиная со времени создания
		NotBefore: time.Now(),
		// время жизни сертификата — 10 лет
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		// устанавливаем использование ключа для цифровой подписи,
		// а также клиентской и серверной авторизации
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}
	// создаём новый приватный RSA-ключ длиной 4096 бит
	// обратите внимание, что для генерации ключа и сертификата
	// используется rand.Reader в качестве источника случайных данных
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatal(err)
	}

	// создаём сертификат x.509
	certBytes, err := x509.CreateCertificate(rand.Reader, t.CSR, t.CSR, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Fatal(err)
	}

	// кодируем сертификат и ключ в формате PEM, который
	// используется для хранения и обмена криптографическими ключами
	err = pem.Encode(&t.CertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		log.Fatal(err)
	}
	err = pem.Encode(&t.KeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		log.Fatal(err)
	}
}

// WriteCert -- записывает TLS сертификат в файл
func (t *TLS) WriteCert(fname string) error {
	if fname == "" {
		return errors.New("fname is empty")
	}
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	if _, err := w.Write(t.CertPEM.Bytes()); err != nil {
		return err
	}
	w.Flush()
	return nil
}

// WriteKey -- записывает TLS приватный ключ в файл
func (t *TLS) WriteKey(fname string) error {
	if fname == "" {
		return errors.New("fname is empty")
	}
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	_, _ = w.Write(t.KeyPEM.Bytes())
	w.Flush()
	return nil
}
