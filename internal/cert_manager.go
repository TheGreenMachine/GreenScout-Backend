package internal

// Utilities for managing user certificates

import (
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Gets the certificate of a given user. If one does not exist in the certificate db,mgenerate a new one by creating and hashing a uuid and insert it into the db.
func GetCertificate(username string, role string) string {
	var certificate string
	result := userDB.QueryRow("select certificate from users where username = ?", username)
	scanErr := result.Scan(&certificate)
	if scanErr != nil && !strings.Contains(scanErr.Error(), "NULL") {
		LogError(scanErr, "Problem scanning response to sql query SELECT certificate FROM users WHERE username = ? with args: "+username)
	}

	_, authenticated := VerifyCertificate(certificate)
	if certificate == "" || !authenticated {
		// Update user db
		newCertRaw := uuid.New()                                                       // Using uuid.new because it's a good source of randomness
		newCert, genErr := bcrypt.GenerateFromPassword([]byte(newCertRaw.String()), 6) // Pass through bcrypt to get a better format in my opinion

		if genErr != nil {
			LogError(genErr, "Problem generating new certificate from uuid "+newCertRaw.String())
		}

		_, err := userDB.Exec("update users set certificate = ? where username = ?", string(newCert), username)

		if err != nil {
			LogErrorf(err, "Problem executing sql query UPDATE users SET certificate = ? WHERE username = ? with args: %v, %v", newCert, username)
		}

		certificate = string(newCert)

		// Update certificate db
		_, execErr := authDB.Exec("insert into certs values(?,?,?)", string(newCert), role, username)
		if execErr != nil {
			LogErrorf(err, "Problem executing sql query INSERT INTO certs VALUES (?,?,?) with args: %v, %v, %v", newCert, role, username)
		}

	}

	return certificate
}

// Verifies the existence of a certificate
func VerifyCertificate(certificate string) (string, bool) {
	var certificateRole string
	result := authDB.QueryRow("select role from certs where certificate = ?", certificate)
	scanErr := result.Scan(&certificateRole)

	if scanErr != nil {
		return "none", false
	}

	return certificateRole, true
}
