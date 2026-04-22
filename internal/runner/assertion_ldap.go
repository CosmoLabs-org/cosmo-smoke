package runner

import (
	"fmt"
	"net"
	"time"

	"github.com/CosmoLabs-org/cosmo-smoke/internal/schema"
)

// CheckLDAPBind sends a BindRequest to an LDAP server and verifies the response.
func CheckLDAPBind(check *schema.LDAPCheck) AssertionResult {
	host := check.Host
	port := check.Port
	if port == 0 {
		if check.UseTLS {
			port = 636
		} else {
			port = 389
		}
	}
	addr := fmt.Sprintf("%s:%d", host, port)
	timeout := check.Timeout.Duration
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	proto := "tcp"
	conn, err := net.DialTimeout(proto, addr, timeout)
	if err != nil {
		return AssertionResult{Type: "ldap_bind", Expected: addr, Actual: err.Error(), Passed: false}
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(timeout))

	// LDAP BindRequest (ASN.1 BER encoded)
	// Currently supports anonymous bind only. PasswordEnv is accepted in schema
	// for forward compatibility but not yet wired into the BER payload.

	bindDN := check.BindDN

	// Authentication: simple [0] empty (anonymous)
	authChoice := []byte{0x80, 0x00} // context-0 (simple), length 0

	// BindRequest name
	nameBytes := []byte(bindDN)
	nameTLV := append([]byte{0x04, byte(len(nameBytes))}, nameBytes...)

	// BindRequest version (integer 3)
	versionTLV := []byte{0x02, 0x01, 0x03} // integer, length 1, value 3

	// BindRequest SEQUENCE body
	bindBody := append(append(versionTLV, nameTLV...), authChoice...)
	bindRequest := append([]byte{0x60, byte(len(bindBody))}, bindBody...) // APPLICATION 0 (0x60 = context + constructed)

	// messageID = 1
	msgIDTLV := []byte{0x02, 0x01, 0x01} // integer, length 1, value 1

	// LDAPMessage SEQUENCE
	msgBody := append(msgIDTLV, bindRequest...)
	if len(msgBody) > 127 {
		return AssertionResult{Type: "ldap_bind", Expected: addr, Actual: "message too long for simple BER encoding", Passed: false}
	}
	ldapMsg := append([]byte{0x30, byte(len(msgBody))}, msgBody...)

	if _, err := conn.Write(ldapMsg); err != nil {
		return AssertionResult{Type: "ldap_bind", Expected: addr, Actual: "write error: " + err.Error(), Passed: false}
	}

	// Read response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil || n < 8 {
		return AssertionResult{Type: "ldap_bind", Expected: addr, Actual: fmt.Sprintf("read error: %v (n=%d)", err, n), Passed: false}
	}

	// Parse response: SEQUENCE { messageID, bindResponse [APPLICATION 1] ... }
	// Check tag at offset after messageID
	// 0x30 (SEQUENCE), length, 0x02 (INTEGER), length=1, messageID, 0x61 (APPLICATION 1 = bindResponse)
	if buf[0] != 0x30 {
		return AssertionResult{Type: "ldap_bind", Expected: addr, Actual: fmt.Sprintf("expected SEQUENCE tag 0x30, got 0x%02x", buf[0]), Passed: false}
	}

	// Find bindResponse tag
	// Skip SEQUENCE tag(1) + length(1) + INTEGER tag(1) + length(1) + value(1) = 5 bytes
	if n > 5 && buf[5] == 0x61 {
		// bindResponse [APPLICATION 1] SEQUENCE { resultCode, matchedDN, diagnosticMessage }
		// Parse the inner SEQUENCE
		if n > 8 {
			// Skip: 0x61, length, 0x0a (ENUMERATED), length=1, resultCode
			resultCode := -1
			for i := 6; i < n-2; i++ {
				if buf[i] == 0x0a && buf[i+1] == 0x01 {
					resultCode = int(buf[i+2])
					break
				}
			}
			if resultCode == 0 { // success
				return AssertionResult{Type: "ldap_bind", Expected: addr, Actual: "bind success", Passed: true}
			}
			return AssertionResult{Type: "ldap_bind", Expected: addr, Actual: fmt.Sprintf("bind result code: %d", resultCode), Passed: false}
		}
	}

	return AssertionResult{Type: "ldap_bind", Expected: addr, Actual: fmt.Sprintf("unexpected response (n=%d)", n), Passed: false}
}
